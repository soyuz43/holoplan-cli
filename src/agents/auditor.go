// src\agents\auditor.go
package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"holoplan-cli/src/types"
)

// AuditResponse defines the expected JSON structure from the LLM
type AuditResponse struct {
	Issues []string `json:"issues"` // Array of issues or empty; "No issues" maps to empty array
}

func Audit(story types.UserStory, xml string) types.Critique {
	prompt := buildAuditPrompt(story, xml)

	// Log the full prompt for debugging
	log.Printf("📝 Audit Prompt:\n%s\n", prompt)

	response, err := callOllama(prompt)
	if err != nil {
		return types.Critique{Issues: []string{"LLM call failed: " + err.Error()}}
	}

	issues := extractIssues(response)
	return types.Critique{Issues: issues}
}

func buildAuditPrompt(story types.UserStory, xml string) string {
	return fmt.Sprintf(`
You are a critical UI reviewer. Your task is to compare the provided user story with the layout XML and identify specific mismatches or missing elements required by the user story.

**Instructions**:
- Respond with a JSON object containing an "issues" array.
- If there are issues, list them as strings in the format "<issue description>" (e.g., "Login button is not centered").
- Each issue must be specific, concise, and directly tied to the user story requirements not met by the XML.
- If the XML fully satisfies the user story, return an empty "issues" array: {"issues": []}.
- Do **not** include validation messages, element counts, collision checks, or any text outside the JSON structure.
- Ignore layout aesthetics unless explicitly mentioned in the user story.

**Examples**:
1. User Story: "As a user, I want a centered login button."
   XML: "<mxGraphModel><root><mxCell id='1' value='Button' x='0' y='0'/></root></mxGraphModel>"
   Response: {"issues": ["Login button is not centered"]}
2. User Story: "As a user, I want a list of items."
   XML: "<mxGraphModel><root><mxCell id='1' value='Item List' x='100' y='100' width='200' height='300'/></root></mxGraphModel>"
   Response: {"issues": ["Item List does not indicate multiple items"]}
3. User Story: "As a user, I want a search bar."
   XML: "<mxGraphModel><root><mxCell id='1' value='Search Bar' x='100' y='100' width='200' height='50'/></root></mxGraphModel>"
   Response: {"issues": []}

**User Story**:
---
%s
---

**Layout XML**:
---
%s
---

**Response Format**:
Return a JSON object: {"issues": ["<issue description>", ...]} or {"issues": []}
`, story.Narrative, xml)
}

func extractIssues(text string) []string {
	// Trim whitespace for consistent comparison
	trimmedText := strings.TrimSpace(text)
	log.Printf("🔍 Processing LLM response (trimmed):\n%s\n", trimmedText)

	// Parse JSON response
	var auditResp AuditResponse
	err := json.Unmarshal([]byte(trimmedText), &auditResp)
	if err != nil {
		log.Printf("🚨 Failed to parse JSON response: %v\nRaw response:\n%s\n", err, trimmedText)
		return []string{"Malformed LLM response: invalid JSON format"}
	}

	// Log parsed issues
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
		"stream": false,  // Explicitly disable streaming[](https://ollama.readthedocs.io/en/api/)
		"format": "json", // Enforce JSON output[](https://github.com/ollama/ollama/blob/main/docs/api.md?plain=1)
		"options": map[string]float64{
			"temperature": 0.0, // Lower temperature for deterministic output[](https://www.reddit.com/r/LocalLLaMA/comments/1d3x3m5/getting_llama3_to_produce_proper_json_through/)
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

	return output.Response, nil
}
