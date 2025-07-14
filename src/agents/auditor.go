// src/agents/auditor.go
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

	"holoplan-cli/src/types"
)

//go:embed prompts/auditor_prompt.txt
var auditorPrompt string

// AuditResponse defines the expected JSON structure from the LLM
type AuditResponse struct {
	Issues []string `json:"issues"`
}

func Audit(story types.UserStory, xml string) types.Critique {
	prompt := buildAuditPrompt(story, xml)

	// DEBUG: Uncomment this line to see the prompt sent to the auditor
	// log.Printf("📝 DEBUG: Audit Prompt:\n%s\n", prompt)

	response, err := callOllama(prompt)
	if err != nil {
		return types.Critique{Issues: []string{"LLM call failed: " + err.Error()}}
	}

	issues := extractIssues(response)
	return types.Critique{Issues: issues}
}

func buildAuditPrompt(story types.UserStory, xml string) string {
	prompt := auditorPrompt
	prompt = strings.ReplaceAll(prompt, "{{story}}", story.Narrative)
	prompt = strings.ReplaceAll(prompt, "{{xml}}", xml)
	return prompt
}

func extractIssues(text string) []string {
	trimmedText := strings.TrimSpace(text)
	log.Printf("🔍 Processing LLM response (trimmed):\n%s\n", trimmedText)

	var auditResp AuditResponse
	err := json.Unmarshal([]byte(trimmedText), &auditResp)
	if err != nil {
		log.Printf("🚨 Failed to parse JSON response: %v\nRaw response:\n%s\n", err, trimmedText)
		return []string{"Malformed LLM response: invalid JSON format"}
	}

	// Normalize and drop false positives
	if len(auditResp.Issues) == 1 && strings.ToLower(strings.TrimSpace(auditResp.Issues[0])) == "no issues" {
		log.Printf("✅ LLM response indicates no issues")
		return nil
	}

	if len(auditResp.Issues) == 0 {
		log.Printf("✅ LLM response has no issues (empty issues array)")
		return nil
	}

	log.Printf("📌 Found %d actionable issues: %v", len(auditResp.Issues), auditResp.Issues)
	return auditResp.Issues
}

func callOllama(prompt string) (string, error) {
	body := map[string]interface{}{
		"model":  "qwen2.5-coder:7b-instruct-q6_K",
		"prompt": prompt,
		"stream": false,
		"format": "json",
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
		return "", err
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
	err = json.Unmarshal(bodyBytes, &output)
	if err != nil {
		return "", fmt.Errorf("failed to decode JSON response: %w", err)
	}

	if !output.Done {
		log.Printf("🚨 Response not marked as done: %v", output)
		return "", fmt.Errorf("incomplete LLM response")
	}

	log.Printf("📤 LLM Response:\n%s\n", output.Response)
	return output.Response, nil
}
