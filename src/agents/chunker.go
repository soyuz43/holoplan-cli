// src\agents\chunker.go
package agents

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"holoplan-cli/src/types"
	"io"
	"net/http"
	"regexp"
	"strings"
)

//go:embed prompts/chunker_prompt.txt
var chunkerSystemPrompt string

const ollamaURL = "http://localhost:11434/api/chat"

// Remove <think> tags and clean up LLM output
func extractCleanJSON(raw string) string {
	reThink := regexp.MustCompile(`(?s)<think>.*?</think>`)
	cleaned := reThink.ReplaceAllString(raw, "")
	cleaned = strings.TrimSpace(cleaned)

	start := strings.Index(cleaned, "{")
	end := strings.LastIndex(cleaned, "}")
	if start == -1 || end == -1 || start > end {
		return cleaned
	}
	jsonChunk := cleaned[start : end+1]
	return escapeLineBreaks(jsonChunk)
}

// Naively escape unescaped newlines within double-quoted values
func escapeLineBreaks(input string) string {
	re := regexp.MustCompile(`"([^"\\]*(?:\\.[^"\\]*)*)"`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		unescaped := match[1 : len(match)-1]
		escaped := strings.ReplaceAll(unescaped, "\n", `\n`)
		return `"` + escaped + `"`
	})
}

// Chunk takes a UserStory and extracts views using the LLM
func Chunk(story types.UserStory) types.ViewPlan {
	sysPrompt := chunkerSystemPrompt

	userPrompt := fmt.Sprintf(`User Story:

- ID: %s
- Title: %s
- Narrative: %s
- View: %s
- Views: %v
- Interaction Origin: %s
- Resulting View: %s
- Shared Components: %v`,
		story.ID,
		story.Title,
		story.Narrative,
		story.View,
		story.Views,
		story.InteractionOrigin,
		story.ResultingView,
		story.SharedComponents,
	)

	payload := map[string]interface{}{
		"model":  "qwen2.5-coder:7b-instruct-q6_K",
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.0,
			"seed":        42,
		},
		"messages": []map[string]string{
			{"role": "system", "content": strings.TrimSpace(sysPrompt)},
			{"role": "user", "content": strings.TrimSpace(userPrompt)},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		panic(fmt.Errorf("failed to marshal request payload: %w", err))
	}

	resp, err := http.Post(ollamaURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		panic(fmt.Errorf("failed to call Ollama API: %w", err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Errorf("failed to read response body: %w", err))
	}

	var parsed struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		panic(fmt.Errorf("failed to unmarshal Ollama API response: %w", err))
	}

	cleaned := extractCleanJSON(parsed.Message.Content)

	// Uncomment this block for debugging
	/*
		fmt.Println("\n🔎 DEBUG: Cleaned JSON extracted from LLM response:")
		fmt.Println(cleaned)
	*/

	var plan types.ViewPlan
	if err := json.Unmarshal([]byte(cleaned), &plan); err != nil {
		fmt.Println("\n🛑 Failed to parse cleaned JSON:")
		fmt.Println("──── Original Output ────")
		fmt.Println(parsed.Message.Content)
		fmt.Println("──── Extracted JSON ────")
		fmt.Println(cleaned)
		panic(fmt.Errorf("JSON parse error: %w", err))
	}

	plan.StoryID = story.ID

	for _, v := range plan.Views {
		fmt.Printf("✅ Extracted view: %s (%s)\n", v.Name, v.Type)
	}

	return plan
}
