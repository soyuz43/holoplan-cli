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

	response, err := callOllamaForLayout(prompt)
	if err != nil {
		fmt.Printf("⚠️ Builder LLM call failed for view '%s': %v\n", view.Name, err)
		return ""
	}

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

// callOllamaForLayout sends the layout prompt to Ollama Hermes model and returns the raw response text.
func callOllamaForLayout(prompt string) (string, error) {
	body := map[string]string{
		"model":  "huihui_ai/Hermes-3-Llama-3.2-abliterated:3b-q8_0",
		"prompt": prompt,
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

	var output struct {
		Response string `json:"response"`
	}

	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		return "", fmt.Errorf("failed to decode Ollama response: %w", err)
	}

	return output.Response, nil
}
