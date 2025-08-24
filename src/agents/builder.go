// src/agents/builder.go
package agents

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"holoplan-cli/src/shared"
	"holoplan-cli/src/types"
)

//go:embed prompts/builder_prompt.txt
var builderPromptDrawio string

//go:embed prompts/builder_prompt_figma.txt
var builderPromptFigma string

// Build takes a ViewLayout and generates layout output (Draw.io XML or Figma JSON) via LLM.
// The `format` should be "drawio" or "figma".
// If anything fails, it returns an empty string and logs the reason.
func Build(view types.ViewLayout, story types.UserStory, format string) string {
	// Select prompt template based on format
	var promptTemplate string
	switch format {
	case "figma":
		promptTemplate = builderPromptFigma
	case "drawio", "":
		promptTemplate = builderPromptDrawio
	default:
		fmt.Printf("⚠️ Unknown format '%s', defaulting to 'drawio'\n", format)
		promptTemplate = builderPromptDrawio
	}

	// 🔧 Fill in template placeholders
	components := strings.Join(view.Components, ", ")
	prompt := strings.ReplaceAll(promptTemplate, "{{view_name}}", view.Name)
	prompt = strings.ReplaceAll(prompt, "{{view_type}}", view.Type)
	prompt = strings.ReplaceAll(prompt, "{{components}}", components)
	prompt = strings.ReplaceAll(prompt, "{{story_narrative}}", story.Narrative)

	// 📤 DEBUG: Uncomment to inspect prompt
	// fmt.Printf("📤 DEBUG Prompt for view '%s' (format=%s):\n%s\n", view.Name, format, prompt)

	response, err := callOllamaForLayout(prompt)
	if err != nil {
		fmt.Printf("⚠️ Builder LLM call failed for view '%s': %v\n", view.Name, err)
		return ""
	}

	if strings.TrimSpace(response) == "" {
		fmt.Printf("⚠️ Empty response from LLM for view '%s'. Skipping.\n", view.Name)
		return ""
	}

	// Extract output based on format
	var result string
	switch format {
	case "figma":
		// Extract clean JSON from LLM response
		result = extractFigmaJSON(response)
		if result == "" || !json.Valid([]byte(result)) {
			fmt.Printf("📥 Raw LLM response for Figma:\n%s\n", response) // Debug
			fmt.Printf("⚠️ Could not extract valid JSON for Figma view '%s'. Skipping.\n", view.Name)
			return ""
		}
	default:
		// For Draw.io, extract XML
		result = shared.ExtractXMLFrom(response)
		if strings.TrimSpace(result) == "" {
			fmt.Printf("⚠️ No valid XML could be extracted for view '%s'. Skipping.\n", view.Name)
			return ""
		}
	}

	return result
}

// extractCleanJSON removes markdown, think tags, and extracts valid JSON
func extractFigmaJSON(raw string) string {
	// Remove markdown code blocks
	cleaned := strings.ReplaceAll(raw, "```json", "")
	cleaned = strings.ReplaceAll(cleaned, "```", "")

	// Remove <think> tags and content
	reThink := regexp.MustCompile(`(?s)<think>.*?</think>`)
	cleaned = reThink.ReplaceAllString(cleaned, "")

	// Trim whitespace
	cleaned = strings.TrimSpace(cleaned)

	// Find JSON object boundaries
	start := strings.Index(cleaned, "{")
	end := strings.LastIndex(cleaned, "}")
	if start == -1 || end == -1 || start > end {
		return cleaned
	}

	jsonChunk := cleaned[start : end+1]
	return jsonChunk
}

// callOllamaForLayout streams a completion from Ollama and returns the full text.
func callOllamaForLayout(prompt string) (string, error) {
	body := map[string]interface{}{
		"model":  "qwen2.5-coder:7b-instruct-q6_K",
		"prompt": prompt,
		"options": map[string]interface{}{
			"temperature": 0.0,
			"seed":        42,
		},
	}
	b, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return "", fmt.Errorf("HTTP POST failed: %w", err)
	}
	defer resp.Body.Close()

	var fullResponse strings.Builder
	decoder := json.NewDecoder(resp.Body)

	for {
		var chunk struct {
			Response string `json:"response"`
			Done     bool   `json:"done"`
		}
		if err := decoder.Decode(&chunk); err != nil {
			fmt.Printf("🛑 Streaming decode failed: %v\n", err)
			break
		}

		fullResponse.WriteString(chunk.Response)
		if chunk.Done {
			break
		}
	}

	return fullResponse.String(), nil
}
