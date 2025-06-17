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
	// This is necessary because LLMs often insert `\n` directly, which breaks JSON
	jsonChunk = escapeLineBreaks(jsonChunk)

	return jsonChunk
}

// Naively escape unescaped newlines within double-quoted values
func escapeLineBreaks(input string) string {
	re := regexp.MustCompile(`"([^"\\]*(?:\\.[^"\\]*)*)"`)

	return re.ReplaceAllStringFunc(input, func(match string) string {
		// Match is a quoted string, e.g. `"some\nvalue"`
		unescaped := match[1 : len(match)-1] // remove quotes
		escaped := strings.ReplaceAll(unescaped, "\n", `\n`)
		return `"` + escaped + `"`
	})
}

func Chunk(story types.UserStory) types.ViewPlan {
	payload := map[string]interface{}{
		"model":       "huihui_ai/Hermes-3-Llama-3.2-abliterated:3b-q8_0",
		"stream":      false,
		"temperature": 0,
		"seed":        42,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are the StoryChunker. Given a user story, return a JSON object with:\n- views: array of {name, type}\n- reasoning: a short explanation of how you decided on the views.\nOnly respond with a JSON object.",
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

	var plan types.ViewPlan
	err = json.Unmarshal([]byte(cleaned), &plan)
	if err != nil {
		fmt.Println("🛑 Failed to parse cleaned JSON:")
		fmt.Println("──── Original Output ────")
		fmt.Println(parsed.Message.Content)
		fmt.Println("──── Extracted JSON ────")
		fmt.Println(cleaned)
		panic(fmt.Errorf("JSON parse error: %w", err))
	}

	plan.StoryID = story.ID
	return plan
}
