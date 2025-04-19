package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"holoplan-cli/src/types"
	"io/ioutil"
	"net/http"
)

const ollamaURL = "http://localhost:11434/api/chat"

func Chunk(story types.UserStory) types.ViewPlan {
	payload := map[string]interface{}{
		"model":       "deepcoder:14b-preview-q4_K_M",
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

	// Marshal payload
	data, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	resp, err := http.Post(ollamaURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	// Parse LLM output
	var parsed struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		panic(err)
	}

	// Extract JSON from message content
	var plan types.ViewPlan
	err = json.Unmarshal([]byte(parsed.Message.Content), &plan)
	if err != nil {
		fmt.Println("🛑 Could not parse LLM output:")
		fmt.Println(parsed.Message.Content)
		panic(err)
	}

	plan.StoryID = story.ID
	return plan
}
