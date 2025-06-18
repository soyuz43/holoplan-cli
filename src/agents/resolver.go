package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"holoplan-cli/src/shared"
	"holoplan-cli/src/types"
)

// Resolve uses an LLM to repair layout XML based on critique feedback
func Resolve(xml string, critique types.Critique) string {
	prompt := buildCorrectionPrompt(xml, critique.Issues)

	response, err := callOllamaForCorrection(prompt)
	if err != nil {
		fmt.Printf("❌ Resolver failed: %v\n", err)
		return xml // fallback: return unmodified
	}

	return shared.ExtractXMLFrom(response)
}

// Create a structured prompt for the Hermes model
func buildCorrectionPrompt(xml string, issues []string) string {
	return fmt.Sprintf(`
You are an expert UI layout assistant.

Your task is to revise a Draw.io layout XML based on these issues:

%s

Here is the original layout:
%s

Return only the corrected XML. Do not include explanations or extra text.
`, formatList(issues), xml)
}

// Formats the list of critique items as a bullet list
func formatList(items []string) string {
	var out strings.Builder
	for _, issue := range items {
		out.WriteString("- " + issue + "\n")
	}
	return out.String()
}

// Sends the correction prompt to Ollama
func callOllamaForCorrection(prompt string) (string, error) {
	body := map[string]string{
		"model":  "qwen2.5-coder:7b-instruct-q6_K",
		"prompt": prompt,
	}
	b, _ := json.Marshal(body)

	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var output struct {
		Response string `json:"response"`
	}
	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		return "", err
	}

	return output.Response, nil
}
