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

// Remove <think>...</think> and extract JSON block only
func extractCleanJSON(raw string) string {
	// Remove all <think>...</think> blocks
	reThink := regexp.MustCompile(`(?s)<think>.*?</think>`)
	cleaned := reThink.ReplaceAllString(raw, "")

	// Trim whitespace
	cleaned = strings.TrimSpace(cleaned)

	// Try to extract first valid JSON block
	start := strings.Index(cleaned, "{")
	end := strings.LastIndex(cleaned, "}")
	if start == -1 || end == -1 || start > end {
		return cleaned // fallback: return raw
	}

	return cleaned[start : end+1]
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
		panic(fmt.Errorf("failed to read response: %w", err))
	}

	var parsed struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal Ollama response: %w", err))
	}

	cleaned := extractCleanJSON(parsed.Message.Content)

	var plan types.ViewPlan
	err = json.Unmarshal([]byte(cleaned), &plan)
	if err != nil {
		fmt.Println("🛑 Could not parse LLM output:")
		fmt.Println(parsed.Message.Content)
		panic(fmt.Errorf("JSON parse error: %w", err))
	}

	plan.StoryID = story.ID
	return plan
}
