package agents

import (
    "holoplan-cli/src/shared"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"holoplan-cli/src/types"
)

// Build takes a ViewLayout and generates Draw.io XML layout via LLM
func Build(view types.ViewLayout) string {
	// Load prompt template from file
	template, err := os.ReadFile("src/prompts/builder_prompt.txt")
	if err != nil {
		fmt.Println("⚠️ Failed to read builder prompt:", err)
		return ""
	}

	// Fill in template placeholders
	prompt := strings.ReplaceAll(string(template), "{{view_name}}", view.Name)
	prompt = strings.ReplaceAll(prompt, "{{view_type}}", view.Type)
	prompt = strings.ReplaceAll(prompt, "{{components}}", strings.Join(view.Components, ", "))

	// LLM call to generate layout
	response, err := callOllamaForLayout(prompt)
	if err != nil {
		fmt.Println("⚠️ Builder LLM call failed:", err)
		return ""
	}

	return shared.ExtractXMLFrom(response)
}

// Sends the layout prompt to Ollama Hermes model
func callOllamaForLayout(prompt string) (string, error) {
	body := map[string]string{
		"model":  "huihui_ai/Hermes-3-Llama-3.2-abliterated:3b-q8_0",
		"prompt": prompt,
	}
	b, _ := json.Marshal(body)

	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var output struct {
		Response string `json:"response"`
	}
	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		return "", err
	}
	return output.Response, nil
}

// Extracts XML portion from model output
	start := strings.Index(response, "<")
	if start == -1 {
		return response
	}
	return response[start:]
}
