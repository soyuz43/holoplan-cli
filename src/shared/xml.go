// src/shared/xml.go
package shared

import (
	"regexp"
	"strings"
)

// ExtractXMLFrom trims LLM commentary and returns just the <mxGraphModel> XML.
func ExtractXMLFrom(response string) string {
	// Clean up any markdown code blocks
	cleaned := strings.ReplaceAll(response, "```xml", "")
	cleaned = strings.ReplaceAll(cleaned, "```", "")
	cleaned = strings.TrimSpace(cleaned)

	// Remove <think> blocks completely
	cleaned = regexp.MustCompile(`(?s)<think>.*?</think>`).ReplaceAllString(cleaned, "")
	cleaned = strings.TrimSpace(cleaned)

	// Look for actual <mxGraphModel> content
	re := regexp.MustCompile(`(?s)<mxGraphModel>.*?</mxGraphModel>`)
	match := re.FindString(cleaned)
	if match != "" {
		return match
	}

	// Fallback: if XML starts somewhere, try returning from first XML tag
	start := strings.Index(cleaned, "<mxGraphModel")
	if start != -1 {
		return cleaned[start:]
	}

	return ""
}
