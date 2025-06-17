// src/shared/xml.go
package shared

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/beevik/etree"
)

// ExtractXMLFrom trims LLM commentary and returns just the <mxGraphModel> XML.
func ExtractXMLFrom(response string) string {
	// Remove markdown code blocks and whitespace
	cleaned := strings.ReplaceAll(response, "```xml", "")
	cleaned = strings.ReplaceAll(cleaned, "```", "")
	cleaned = strings.TrimSpace(cleaned)

	// Remove <think> blocks completely
	cleaned = regexp.MustCompile(`(?s)<think>.*?</think>`).ReplaceAllString(cleaned, "")
	cleaned = strings.TrimSpace(cleaned)

	// Look for complete <mxGraphModel>
	re := regexp.MustCompile(`(?s)<mxGraphModel>.*?</mxGraphModel>`)
	match := re.FindString(cleaned)
	if match != "" {
		return match
	}

	// Fallback: attempt to slice from start of <mxGraphModel>
	start := strings.Index(cleaned, "<mxGraphModel")
	if start != -1 {
		return cleaned[start:]
	}

	return ""
}

// SanitizeXML ensures all <mxGeometry> elements have required attributes.
// Only fills in missing ones using safe defaults.
func SanitizeXML(raw string) (string, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(raw); err != nil {
		return "", fmt.Errorf("❌ etree failed to parse input XML: %w", err)
	}

	// Define required attributes and their defaults
	required := []string{"x", "y", "width", "height", "as"}
	defaults := map[string]string{
		"x":      "0",
		"y":      "0",
		"width":  "100",
		"height": "50",
		"as":     "geometry",
	}

	// Fill missing attrs on each <mxGeometry>
	for _, geo := range doc.FindElements("//mxGeometry") {
		for _, key := range required {
			if geo.SelectAttr(key) == nil {
				geo.CreateAttr(key, defaults[key])
				fmt.Printf("🧪 Filled missing '%s' with default '%s'\n", key, defaults[key])
			}
		}
	}

	return doc.WriteToString()
}
