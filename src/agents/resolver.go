// src/agents/resolver.go
package agents

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"holoplan-cli/src/shared"
	"holoplan-cli/src/types"
)

//go:embed prompts/resolver_prompt.txt
var resolverPrompt string

// Resolve uses an LLM to repair layout XML based on critique feedback and user story
func Resolve(xml string, critique types.Critique, story types.UserStory) string {
	prompt := buildCorrectionPrompt(xml, critique.Issues, story)
	// DEBUG: Uncomment this line to see the prompt sent to the resolver
	// log.Printf("📝 Resolver Prompt:\n%s\n", prompt)

	response, err := callOllamaForCorrection(prompt)
	if err != nil {
		log.Printf("❌ Resolver failed: %v", err)
		return xml
	}

	extractedXML := shared.ExtractXMLFrom(response)
	if extractedXML == "" {
		log.Printf("🚨 Resolver returned invalid or empty XML:\n%s\n", response)
		return xml
	}

	sanitizedXML, err := shared.SanitizeXML(extractedXML)
	if err != nil {
		log.Printf("🚨 Sanitization failed: %v\nRaw XML:\n%s\n", err, extractedXML)
		return xml
	}

	// log.Printf("✅ Sanitized Corrected XML:\n%s\n", sanitizedXML)
	return sanitizedXML
}

// buildCorrectionPrompt fills the embedded resolver prompt template with values
func buildCorrectionPrompt(xml string, issues []string, story types.UserStory) string {
	prompt := resolverPrompt
	prompt = strings.ReplaceAll(prompt, "{{issues}}", formatList(issues))
	prompt = strings.ReplaceAll(prompt, "{{story}}", story.Narrative)
	prompt = strings.ReplaceAll(prompt, "{{xml}}", xml)
	return prompt
}

// formatList formats critique issues as a markdown-like bullet list
func formatList(items []string) string {
	var out strings.Builder
	for _, issue := range items {
		out.WriteString("- " + issue + "\n")
	}
	return out.String()
}

// callOllamaForCorrection sends the filled prompt to Ollama and returns the raw XML
func callOllamaForCorrection(prompt string) (string, error) {
	body := map[string]interface{}{
		"model":  "qwen2.5-coder:7b-instruct-q6_K",
		"prompt": prompt,
		"stream": false,
		"options": map[string]float64{
			"temperature": 0.0,
			"seed":        42,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return "", fmt.Errorf("HTTP error: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var output struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
	}
	if err := json.Unmarshal(bodyBytes, &output); err != nil {
		return "", fmt.Errorf("failed to parse top-level JSON: %w", err)
	}

	if !output.Done {
		return "", fmt.Errorf("incomplete LLM response: %#v", output)
	}

	rawXML := output.Response
	if strings.TrimSpace(rawXML) == "" {
		return "", fmt.Errorf("empty XML returned")
	}

	return rawXML, nil
}
