// src\runner\pipeline.go
package runner

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"holoplan-cli/src/agents"
	"holoplan-cli/src/shared"
	"holoplan-cli/src/types"
	"holoplan-cli/src/validator"

	"github.com/beevik/etree"
	"gopkg.in/yaml.v3"
)

const MaxCorrections = 3

func RunPipeline(yamlPath string) error {
	stories, err := loadStories(yamlPath)
	if err != nil {
		return fmt.Errorf("failed to load stories: %w", err)
	}

	for _, story := range stories {
		fmt.Printf("🔍 Processing Story: %s\n", story.ID)

		viewPlan, ok := safeChunk(story)
		if !ok {
			log.Printf("⚠️ Failed to chunk story: %s — skipping\n", story.ID)
			continue
		}

		for _, view := range viewPlan.Views {
			fmt.Printf("⚙️  Generating view: %s\n", view.Name)

			xml, ok := safeBuild(view)
			if !ok {
				log.Printf("⚠️ Failed to build layout for view: %s\n", view.Name)
				continue
			}

			critique, ok := safeAudit(story, xml)
			if !ok {
				log.Printf("⚠️ Failed to audit layout for view: %s\n", view.Name)
				continue
			}

			attempt := 0
			for critique.HasIssues() && attempt < MaxCorrections {
				fmt.Printf("🔁 Correction attempt %d for %s\n", attempt+1, view.Name)

				xml, ok = safeResolve(xml, critique)
				if !ok {
					log.Printf("⚠️ Resolve failed at attempt %d — aborting view\n", attempt+1)
					break
				}

				critique, ok = safeAudit(story, xml)
				if !ok {
					log.Printf("⚠️ Audit failed during retry — aborting view\n")
					break
				}
				attempt++
			}
			xml = shared.ForceQuoteAllAttributes(xml)
			if err := validator.CheckLayout(xml); err != nil {
				log.Printf("❌ Layout validation failed: %v", err)
			} else {
				fmt.Println("✅ Spatial layout passed")
			}

			err = saveOutput(story.ID, view.Name, xml, critique)
			if err != nil {
				log.Printf("⚠️ Failed to save output: %v", err)
			}
		}
	}

	if err := mergeDrawio("output/final.drawio"); err != nil {
		return fmt.Errorf("failed to merge drawio files: %w", err)
	}

	return nil
}

func loadStories(path string) ([]types.UserStory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var stories []types.UserStory
	err = yaml.Unmarshal(data, &stories)
	return stories, err
}

func safeChunk(story types.UserStory) (types.ViewPlan, bool) {
	defer recoverLLM("Chunk")
	return agents.Chunk(story), true
}

func safeBuild(view types.ViewLayout) (string, bool) {
	defer recoverLLM("Build")
	return agents.Build(view), true
}

func safeAudit(story types.UserStory, xml string) (types.Critique, bool) {
	defer recoverLLM("Audit")
	return agents.Audit(story, xml), true
}

func safeResolve(xml string, critique types.Critique) (string, bool) {
	defer recoverLLM("Resolve")
	return agents.Resolve(xml, critique), true
}

func recoverLLM(agent string) {
	if r := recover(); r != nil {
		log.Printf("🔥 Panic recovered in %s agent: %v\n", agent, r)
	}
}

func saveOutput(storyID, viewName string, xml string, critique types.Critique) error {
	if err := os.MkdirAll("output", os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Combine storyID and viewName, then sanitize
	base := sanitize(fmt.Sprintf("%s_%s", storyID, viewName))

	// Use simple filepath.Join instead of dedupeFilename since we now have unique names
	xmlPath := filepath.Join("output", base+".drawio")
	if err := os.WriteFile(xmlPath, []byte(xml), 0644); err != nil {
		return fmt.Errorf("failed to write XML: %w", err)
	}

	if critique.HasIssues() {
		report := "Critique Issues:\n"
		for _, issue := range critique.Issues {
			report += "- " + issue + "\n"
		}
		critiquePath := filepath.Join("output", base+".critique.txt")
		if err := os.WriteFile(critiquePath, []byte(report), 0644); err != nil {
			return fmt.Errorf("failed to write critique report: %w", err)
		}
	}

	return nil
}

func sanitize(name string) string {
	// Keep hyphens and underscores, just replace spaces and convert to lowercase
	name = strings.ReplaceAll(strings.TrimSpace(name), " ", "_")
	return strings.ToLower(name)
}

// mergeDrawio builds a valid <mxfile> with one <diagram> per input .drawio file.
func mergeDrawio(outputPath string) error {
	files, err := filepath.Glob("output/*.drawio")
	if err != nil {
		return fmt.Errorf("failed to scan output files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no .drawio files found in output directory")
	}

	finalDoc := etree.NewDocument()
	finalDoc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
	mxfile := finalDoc.CreateElement("mxfile")

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}

		// Basic validation
		offenders, err := shared.DetectEscapedFillColors(string(content))
		if err != nil {
			return fmt.Errorf("[x] Error detecting escaped colors in %s: %w", file, err)
		}
		if len(offenders) > 0 {
			return fmt.Errorf("[x] Escaped fillColor values detected in %s: %v", file, offenders)
		}

		// Parse the input file
		subDoc := etree.NewDocument()
		if err := subDoc.ReadFromString(string(content)); err != nil {
			return fmt.Errorf("failed to parse XML in %s: %w", file, err)
		}

		// Find the <mxGraphModel> element
		model := subDoc.FindElement("//mxGraphModel")
		if model == nil {
			return fmt.Errorf("no <mxGraphModel> found in %s", file)
		}

		// Create a <diagram> element and append the <mxGraphModel> into it
		diagram := mxfile.CreateElement("diagram")

		// Clean up the filename for the diagram name
		baseName := filepath.Base(file)
		diagramName := strings.TrimSuffix(baseName, filepath.Ext(baseName))
		diagram.CreateAttr("name", diagramName)

		// Add the model to the diagram
		diagram.AddChild(model.Copy())
	}

	// Serialize to string with proper formatting
	finalDoc.Indent(2)
	finalXML, err := finalDoc.WriteToString()
	if err != nil {
		return fmt.Errorf("failed to serialize final <mxfile>: %w", err)
	}

	// Validate the final XML structure
	validateDoc := etree.NewDocument()
	if err := validateDoc.ReadFromString(finalXML); err != nil {
		return fmt.Errorf("🚨 final merged <mxfile> is malformed: %w", err)
	}

	return os.WriteFile(outputPath, []byte(finalXML), 0644)
}
