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

	log.Printf("📝 Audit Prompt:\n%s\n", prompt)

	response, err := callOllama(prompt)
	if err != nil {
		return types.Critique{Issues: []string{"LLM call failed: " + err.Error()}}
	}

	issues := extractIssues(response)
	return types.Critique{Issues: issues}
}

func buildAuditPrompt(story types.UserStory, xml string) string {
	// fill in the embedded template instead of fmt.Sprintf inline block
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

	if len(auditResp.Issues) == 0 {
		log.Printf("✅ LLM response has no issues (empty issues array)")
		return nil
	}

	log.Printf("📌 Found %d valid issues: %v", len(auditResp.Issues), auditResp.Issues)
	return auditResp.Issues
}

func callOllama(prompt string) (string, error) {
	body := map[string]interface{}{
		"model":  "llama3.1:8b",
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
	// log.Printf("📥 Raw Ollama Response:\n%s\n", string(bodyBytes))

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
