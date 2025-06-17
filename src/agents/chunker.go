// src/agents/chunker.go
package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"holoplan-cli/src/types"
	"io"
	"net/http"
	"regexp"
	"strings"
)

const ollamaURL = "http://localhost:11434/api/chat"

// Remove <think> tags and clean up LLM output
func extractCleanJSON(raw string) string {
	// 1. Remove all <think>...</think> blocks
	reThink := regexp.MustCompile(`(?s)<think>.*?</think>`)
	cleaned := reThink.ReplaceAllString(raw, "")

	// 2. Trim surrounding whitespace
	cleaned = strings.TrimSpace(cleaned)

	// 3. Extract first valid JSON object (naively)
	start := strings.Index(cleaned, "{")
	end := strings.LastIndex(cleaned, "}")
	if start == -1 || end == -1 || start > end {
		return cleaned // fallback to raw
	}
	jsonChunk := cleaned[start : end+1]

	// 4. Escape any literal newlines inside string fields
	jsonChunk = escapeLineBreaks(jsonChunk)

	return jsonChunk
}

// Naively escape unescaped newlines within double-quoted values
func escapeLineBreaks(input string) string {
	re := regexp.MustCompile(`"([^"\\]*(?:\\.[^"\\]*)*)"`)

	return re.ReplaceAllStringFunc(input, func(match string) string {
		unescaped := match[1 : len(match)-1] // remove quotes
		escaped := strings.ReplaceAll(unescaped, "\n", `\n`)
		return `"` + escaped + `"`
	})
}

// Chunk takes a UserStory and extracts views using the LLM.
func Chunk(story types.UserStory) types.ViewPlan {
	payload := map[string]interface{}{
		"model":       "qwen2.5-coder:7b-instruct-q6_K",
		"stream":      false,
		"temperature": 0,
		"seed":        42,
		"messages": []map[string]string{
			{
				"role": "system",
				"content": `You are the StoryChunker.

				Given a user story, return a JSON object with:

				- views: array of {name, type, components}
				- reasoning: short explanation of how you decided on the views

			Each view should include a reasonable set of UI components based on the story. Components should be descriptive nouns or short phrases like "Dog Image", "Breed", "Age", "Adopt Button".

			Respond ONLY with a JSON object. Do not include explanations, markdown, or extra formatting.`,
			},
			{
				"role":    "user",
				"content": fmt.Sprintf("User story:\n%s", story.Narrative),
			},
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
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal Ollama API response: %w", err))
	}

	cleaned := extractCleanJSON(parsed.Message.Content)

	// Debug Line: print the extracted JSON from the chunker
	// fmt.Println("\n🔎 DEBUG: Cleaned JSON extracted from LLM response:")
	fmt.Println(cleaned)

	var plan types.ViewPlan
	err = json.Unmarshal([]byte(cleaned), &plan)
	if err != nil {
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
