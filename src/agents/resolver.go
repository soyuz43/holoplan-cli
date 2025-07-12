// src/agents/resolver.go
package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"holoplan-cli/src/shared"
	"holoplan-cli/src/types"
)

// Resolve uses an LLM to repair layout XML based on critique feedback
func Resolve(xml string, critique types.Critique) string {
	prompt := buildCorrectionPrompt(xml, critique.Issues)

	// Log the full prompt for debugging
	log.Printf("📝 Resolver Prompt:\n%s\n", prompt)

	response, err := callOllamaForCorrection(prompt)
	if err != nil {
		log.Printf("❌ Resolver failed: %v", err)
		return xml // fallback: return unmodified
	}

	// Validate that the response contains XML
	extractedXML := shared.ExtractXMLFrom(response)
	if extractedXML == "" {
		log.Printf("🚨 Resolver returned invalid or empty XML:\n%s\n", response)
		return xml // fallback: return unmodified
	}

	return extractedXML
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
	body := map[string]interface{}{
		"model":  "qwen2.5-coder:7b-instruct-q6_K",
		"prompt": prompt,
		"stream": false,  // Explicitly disable streaming
		"format": "json", // Enforce JSON output
		"options": map[string]float64{
			"temperature": 0.0, // Lower temperature for deterministic output
		},
	}
	b, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Log raw HTTP response for debugging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	log.Printf("📥 Raw Ollama Response:\n%s\n", string(bodyBytes))

	var output struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
	}
	err = json.Unmarshal(bodyBytes, &output)
	if err != nil {
		return "", fmt.Errorf("failed to decode JSON response: %w", err)
	}

	// Validate response
	if !output.Done {
		log.Printf("🚨 Response not marked as done: %v", output)
		return "", fmt.Errorf("incomplete LLM response")
	}

	// Log the parsed LLM response
	log.Printf("📤 LLM Response:\n%s\n", output.Response)

	// Parse the response as JSON to extract the XML
	var responseJSON struct {
		XML string `json:"xml"`
	}
	err = json.Unmarshal([]byte(output.Response), &responseJSON)
	if err != nil {
		log.Printf("🚨 Failed to parse JSON response: %v\nRaw response:\n%s\n", err, output.Response)
		return "", fmt.Errorf("malformed LLM response: invalid JSON format")
	}

	if responseJSON.XML == "" {
		log.Printf("🚨 LLM response contains empty XML:\n%s\n", output.Response)
		return "", fmt.Errorf("malformed LLM response: no XML provided")
	}

	return responseJSON.XML, nil
}
