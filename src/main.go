// src\main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"holoplan-cli/src/agents"
	"holoplan-cli/src/types"
	"holoplan-cli/src/validator"

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
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}

		start := strings.Index(string(content), "<mxCell")
		end := strings.LastIndex(string(content), "</root>")
		if start == -1 || end == -1 {
			continue
		}

		inner := string(content[start:end])
		allCells = append(allCells, inner)
	}

	// Write Final Output
	final := fmt.Sprintf(`<mxGraphModel><root>%s</root></mxGraphModel>`, strings.Join(allCells, "\n"))
	return os.WriteFile(outputPath, []byte(final), 0644)
}
