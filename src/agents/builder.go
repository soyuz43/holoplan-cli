// src/agents/builder.go
package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"holoplan-cli/src/shared"
	"holoplan-cli/src/types"
)

// Build takes a ViewLayout and generates Draw.io XML layout via LLM.
// If anything fails, it returns an empty string and logs the reason.
func Build(view types.ViewLayout) string {
	template, err := os.ReadFile("src/prompts/builder_prompt.txt")
	if err != nil {
		fmt.Printf("⚠️ Failed to read builder prompt template: %v\n", err)
		return ""
	}

	prompt := strings.ReplaceAll(string(template), "{{view_name}}", view.Name)
	prompt = strings.ReplaceAll(prompt, "{{view_type}}", view.Type)
	prompt = strings.ReplaceAll(prompt, "{{components}}", strings.Join(view.Components, ", "))

	// 🔍 Print the final prompt before sending it to the LLM
	// fmt.Printf("📤 DEBUG Prompt for view '%s':\n%s\n", view.Name, prompt)

	response, err := callOllamaForLayout(prompt)
	if err != nil {
		fmt.Printf("⚠️ Builder LLM call failed for view '%s': %v\n", view.Name, err)
		return ""
	}

	// fmt.Printf("🧪 DEBUG: Raw LLM response for '%s':\n%s\n", view.Name, response)

	if strings.TrimSpace(response) == "" {
		fmt.Printf("⚠️ Empty response from LLM for view '%s'. Skipping.\n", view.Name)
		return ""
	}

	xml := shared.ExtractXMLFrom(response)
	if strings.TrimSpace(xml) == "" {
		fmt.Printf("⚠️ No valid XML could be extracted for view '%s'. Skipping.\n", view.Name)
		return ""
	}

	return xml
}

// callOllamaForLayout streams a completion from Ollama and returns the full text.
func callOllamaForLayout(prompt string) (string, error) {
	body := map[string]interface{}{
		"model":  "qwen2.5-coder:7b-instruct-q6_K",
		"prompt": prompt,
		"options": map[string]interface{}{
			"temperature": 0.0,
			"seed":        42, // optional
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
