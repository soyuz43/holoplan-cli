package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"holoplan-cli/src/types"
)

func Audit(story types.UserStory, xml string) types.Critique {
	prompt := buildAuditPrompt(story, xml)

	response, err := callOllama(prompt)
	if err != nil {
		return types.Critique{Issues: []string{"LLM call failed: " + err.Error()}}
	}

	issues := extractIssues(response)
	return types.Critique{Issues: issues}
}

func buildAuditPrompt(story types.UserStory, xml string) string {
	return fmt.Sprintf(`
You are a critical UI reviewer.

Below is a user story:
---
%s
---

And here is the current layout XML:
---
%s
---

List specific issues or mismatches. Use "- " for each issue.
If layout looks valid, say: "No issues."
`, story.Narrative, xml)
}

func extractIssues(text string) []string {
	lines := strings.Split(text, "\n")
	var issues []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- ") {
			issues = append(issues, strings.TrimPrefix(line, "- "))
		}
	}
	if len(issues) == 0 && strings.Contains(text, "No issues") {
		return nil
	}
	return issues
}

func callOllama(prompt string) (string, error) {
	body := map[string]string{
		"model":  "llama3.1:8b",
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
