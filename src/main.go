// src\main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"holoplan-cli/src/agents"
	"holoplan-cli/src/types"
	"holoplan-cli/src/validator"

	"github.com/beevik/etree"
	"gopkg.in/yaml.v3"
)

func LoadStories(path string) ([]types.UserStory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var stories []types.UserStory
	err = yaml.Unmarshal(data, &stories)
	return stories, err
}

// Max LLM fix attempts per view
const MaxCorrections = 3

func main() {
	storyPath := flag.String("stories", "", "Path to user stories YAML")
	flag.Parse()

	if *storyPath == "" {
		log.Fatalf("❌ Please provide a path to a YAML file using --stories")
	}

	stories, err := LoadStories(*storyPath)
	if err != nil {
		log.Fatalf("Error loading stories: %v", err)
	}

	for _, story := range stories {
		fmt.Printf("🔍 Processing Story: %s\n", story.ID)

		// LLM Agent 1: StoryChunker (safe)
		viewPlan, ok := safeChunk(story)
		if !ok {
			log.Printf("⚠️ Failed to chunk story: %s — skipping\n", story.ID)
			continue
		}

		for _, view := range viewPlan.Views {
			fmt.Printf("⚙️  Generating view: %s\n", view.Name)

			// Builder (safe)
			xml, ok := safeBuild(view)
			if !ok {
				log.Printf("⚠️ Failed to build layout for view: %s\n", view.Name)
				continue
			}

			// Auditor (safe)
			critique, ok := safeAudit(story, xml)
			if !ok {
				log.Printf("⚠️ Failed to audit layout for view: %s\n", view.Name)
				continue
			}

			// Retry corrections
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

			err = SaveOutput(view.Name, xml, critique)
			if err != nil {
				log.Printf("⚠️ Failed to save output: %v", err)
			}
		}
	}

	err = MergeDrawio("output/final.drawio")
	if err != nil {
		log.Fatalf("Merge failed: %v", err)
	}
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

func SaveOutput(viewName string, xml string, critique types.Critique) error {
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

func MergeDrawio(outputPath string) error {
	files, err := filepath.Glob("output/*.drawio")
	if err != nil {
		return fmt.Errorf("failed to scan output files: %w", err)
	}

	var allCells []string
	for i, file := range files {
		offset := (i + 1) * 100 // Each file gets a unique ID range

		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}

		// Use etree to parse and re-ID
		doc := etree.NewDocument()
		if err := doc.ReadFromString(string(content)); err != nil {
			return fmt.Errorf("failed to parse XML in %s: %w", file, err)
		}

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
		}

		root := doc.FindElement("//root")
		for _, cell := range root.ChildElements() {
			// Create a temporary document to serialize the element
			tempDoc := etree.NewDocument()
			tempDoc.SetRoot(cell.Copy())

			// Use WriteToString on the document
			cellXML, err := tempDoc.WriteToString()
			if err != nil {
				return fmt.Errorf("failed to serialize mxCell: %w", err)
			}

			// Extract just the element content (remove XML declaration)
			lines := strings.Split(cellXML, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && !strings.HasPrefix(line, "<?xml") {
					allCells = append(allCells, line)
					break
				}
			}
		}
	}

	// Write Final Output
	final := fmt.Sprintf(`<mxGraphModel><root>%s</root></mxGraphModel>`, strings.Join(allCells, "\n"))
	return os.WriteFile(outputPath, []byte(final), 0644)
}
