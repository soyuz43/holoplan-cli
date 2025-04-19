package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"holoplan-cli/src/agents" // LLM agents (Chunker, Builder, etc)
	"holoplan-cli/src/types"
	"holoplan-cli/src/validator" // Go-based spatial checks

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
	// 1. Load user stories (to be implemented)
	stories, err := LoadStories("examples/user_stories.yaml")
	if err != nil {
		log.Fatalf("Error loading stories: %v", err)
	}

	// 2. Process each user story
	for _, story := range stories {
		fmt.Printf("🔍 Processing Story: %s\n", story.ID)

		// LLM Agent 1: StoryChunker
		viewPlan := agents.Chunk(story)

		// Generate + audit each view in the plan
		for _, view := range viewPlan.Views {
			fmt.Printf("⚙️  Generating view: %s\n", view.Name)

			xml := agents.Build(view)            // Builder
			critique := agents.Audit(story, xml) // Auditor

			// 3. Retry up to N times if audit fails
			attempt := 0
			for critique.HasIssues() && attempt < MaxCorrections {
				fmt.Printf("🔁 Correction attempt %d for %s\n", attempt+1, view.Name)
				xml = agents.Resolve(xml, critique)
				critique = agents.Audit(story, xml)
				attempt++
			}

			// 4. Validate layout geometry with Go
			if err := validator.CheckLayout(xml); err != nil {
				log.Printf("❌ Layout validation failed: %v", err)
			} else {
				fmt.Println("✅ Spatial layout passed")
			}

			// 5. Write per-view XML + audit report (to be implemented)
			err = SaveOutput(view.Name, xml, critique)
			if err != nil {
				log.Printf("⚠️ Failed to save output: %v", err)
			}
		}
	}

	// 6. Merge all XMLs into final file (to be implemented)
	err = MergeDrawio("output/final.drawio.xml")
	if err != nil {
		log.Fatalf("Merge failed: %v", err)
	}
}
func SaveOutput(viewName string, xml string, critique types.Critique) error {
	// Ensure output directory exists
	if err := os.MkdirAll("output", os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 1. Save XML
	xmlPath := filepath.Join("output", viewName+".drawio.xml")
	if err := os.WriteFile(xmlPath, []byte(xml), 0644); err != nil {
		return fmt.Errorf("failed to write XML: %w", err)
	}

	// 2. Save critique if any
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
	files, err := filepath.Glob("output/*.drawio.xml")
	if err != nil {
		return fmt.Errorf("failed to scan output files: %w", err)
	}

	var allCells []string

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}

		// Extract <mxCell> blocks from each XML
		start := strings.Index(string(content), "<mxCell")
		end := strings.LastIndex(string(content), "</root>")
		if start == -1 || end == -1 {
			continue
		}

		// Extract all cell elements (naive)
		inner := string(content[start:end])
		allCells = append(allCells, inner)
	}

	// Wrap in outer structure
	final := fmt.Sprintf(`<mxGraphModel><root>%s</root></mxGraphModel>`, strings.Join(allCells, "\n"))

	// Write final output
	return os.WriteFile(outputPath, []byte(final), 0644)
}
