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

	"github.com/beevik/etree"
)

//go:embed prompts/resolver_prompt.txt
var resolverPrompt string

// Resolve uses an LLM to repair layout XML based on critique feedback and user story
func Resolve(xml string, critique types.Critique, story types.UserStory) string {
	prompt := buildCorrectionPrompt(xml, critique.Issues, story)

	log.Printf("📝 Resolver Prompt:\n%s\n", prompt)

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

	doc := etree.NewDocument()
	if err := doc.ReadFromString(extractedXML); err != nil {
		log.Printf("🚨 Resolver returned malformed XML: %v\nRaw XML:\n%s\n", err, extractedXML)
		return xml
	}

	return extractedXML
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

// callOllamaForCorrection sends the filled prompt to the LLM backend
func callOllamaForCorrection(prompt string) (string, error) {
	body := map[string]interface{}{
		"model":  "qwen2.5-coder:7b-instruct-q6_K",
		"prompt": prompt,
		"stream": false,
		"format": "json",
		"options": map[string]float64{
			"temperature": 0.0,
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

	if !output.Done {
		log.Printf("🚨 Response not marked as done: %v", output)
		return "", fmt.Errorf("incomplete LLM response")
	}

	log.Printf("📤 LLM Response:\n%s\n", output.Response)

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
