package runner

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

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

			if err := validator.CheckLayout(xml); err != nil {
				log.Printf("❌ Layout validation failed: %v", err)
			} else {
				fmt.Println("✅ Spatial layout passed")
			}

			err = saveOutput(view.Name, xml, critique)
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

func saveOutput(viewName string, xml string, critique types.Critique) error {
	if err := os.MkdirAll("output", os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	xmlPath := filepath.Join("output", viewName+".drawio")
	if err := os.WriteFile(xmlPath, []byte(xml), 0644); err != nil {
		return fmt.Errorf("failed to write XML: %w", err)
	}

	if critique.HasIssues() {
		report := "Critique Issues:\n"
		for _, issue := range critique.Issues {
			report += "- " + issue + "\n"
		}
		critiquePath := filepath.Join("output", viewName+".critique.txt")
		if err := os.WriteFile(critiquePath, []byte(report), 0644); err != nil {
			return fmt.Errorf("failed to write critique report: %w", err)
		}
	}

	return nil
}

// mergeDrawio combines all *.drawio files from the output directory into one valid mxGraphModel XML.
func mergeDrawio(outputPath string) error {
	files, err := filepath.Glob("output/*.drawio")
	if err != nil {
		return fmt.Errorf("failed to scan output files: %w", err)
	}

	var allCells []etree.Element
	for i, file := range files {
		offset := (i + 1) * 100

		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}

		doc := etree.NewDocument()
		if err := doc.ReadFromString(string(content)); err != nil {
			return fmt.Errorf("failed to parse XML in %s: %w", file, err)
		}

		// Check for escaped fillColor issues
		offenders, err := shared.DetectEscapedFillColors(string(content))
		if err != nil {
			return fmt.Errorf("[x] Error detecting escaped colors in %s: %w", file, err)
		}
		if len(offenders) > 0 {
			return fmt.Errorf("[x] Escaped fillColor values detected in %s: %v", file, offenders)
		}

		// Offset IDs to avoid collisions
		for _, cell := range doc.FindElements("//mxCell") {
			if idAttr := cell.SelectAttr("id"); idAttr != nil {
				if idInt, err := strconv.Atoi(idAttr.Value); err == nil {
					idAttr.Value = strconv.Itoa(idInt + offset)
				}
			}
			if parentAttr := cell.SelectAttr("parent"); parentAttr != nil {
				if parentInt, err := strconv.Atoi(parentAttr.Value); err == nil {
					parentAttr.Value = strconv.Itoa(parentInt + offset)
				}
			}

			// Ensure valid mxCell structure
			if len(cell.ChildElements()) == 0 {
				cell.SetText("") // makes etree serialize as self-closing
			}
		}

		// Collect <mxCell> elements from <root>
		root := doc.FindElement("//root")
		for _, cell := range root.ChildElements() {
			allCells = append(allCells, *cell.Copy())
		}
	}

	// Construct final document
	finalDoc := etree.NewDocument()
	finalDoc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
	graphModel := finalDoc.CreateElement("mxGraphModel")
	root := graphModel.CreateElement("root")
	for _, cell := range allCells {
		root.AddChild(&cell)
	}

	// Validate final structure before saving
	finalXML, err := finalDoc.WriteToString()
	if err != nil {
		return fmt.Errorf("failed to serialize final merged XML: %w", err)
	}

	validateDoc := etree.NewDocument()
	if err := validateDoc.ReadFromString(finalXML); err != nil {
		return fmt.Errorf("🚨 final merged XML is malformed: %w", err)
	}

	return os.WriteFile(outputPath, []byte(finalXML), 0644)
}
