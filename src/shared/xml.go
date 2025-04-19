package shared

import "strings"

// ExtractXMLFrom trims off any LLM commentary and returns just the XML.
func ExtractXMLFrom(response string) string {
	start := strings.Index(response, "<")
	if start == -1 {
		return response
	}
	return response[start:]
}
