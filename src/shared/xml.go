// src/shared/xml.go
package shared

import (
	"fmt"
	"regexp"
	"strconv"
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

// fixUnquotedAttributes ensures all attribute values are quoted, e.g., width=180 -> width="180"
// But preserves the internal structure of style attributes which contain key=value pairs
func fixUnquotedAttributes(xml string) string {
	fmt.Println("🛠️ Fixing unquoted XML attributes")

	// Step 1: Extract and temporarily replace style attributes to protect them
	styleRe := regexp.MustCompile(`style="[^"]*"`)
	styles := styleRe.FindAllString(xml, -1)
	placeholder := "@@STYLE_PLACEHOLDER@@"

	// Replace all style attributes with placeholders
	xml = styleRe.ReplaceAllString(xml, placeholder)

	// Step 2: Apply the original fixing logic to everything else
	attrRe := regexp.MustCompile(`\b([a-zA-Z_:]+)=([^\s"'=<>` + "`" + `]+)`)
	count := 0

	xml = attrRe.ReplaceAllStringFunc(xml, func(attr string) string {
		parts := strings.SplitN(attr, "=", 2)
		if len(parts) != 2 {
			return attr
		}
		key, val := parts[0], parts[1]
		val = strings.TrimSpace(val)
		if strings.HasPrefix(val, `"`) || strings.HasPrefix(val, `'`) {
			return attr
		}
		count++
		return fmt.Sprintf(`%s="%s"`, key, val)
	})

	// Step 3: Restore the original style attributes
	for _, style := range styles {
		xml = strings.Replace(xml, placeholder, style, 1)
	}

	fmt.Printf("🔧 Fixed %d unquoted attribute(s)\n", count)
	return xml
}

// escapeInvalidEntities replaces standalone & with &amp;, excluding valid XML entities
func escapeInvalidEntities(xml string) string {
	// Replace all & with &amp;
	xml = strings.ReplaceAll(xml, "&", "&amp;")

	// Restore valid XML entities
	xml = strings.ReplaceAll(xml, "&amp;lt;", "&lt;")
	xml = strings.ReplaceAll(xml, "&amp;gt;", "&gt;")
	xml = strings.ReplaceAll(xml, "&amp;quot;", "&quot;")
	xml = strings.ReplaceAll(xml, "&amp;apos;", "&apos;")
	xml = strings.ReplaceAll(xml, "&amp;amp;", "&amp;")

	return xml
}

// SanitizeXML ensures all <mxGeometry> elements have required attributes.
// Also wraps unquoted attribute values and escapes invalid ampersands.
func SanitizeXML(raw string) (string, error) {
	raw = ForceQuoteAllAttributes(raw)
	fmt.Println("🧼 [SanitizeXML] ForceQuoteAllAttributes applied")
	raw = fixUnquotedAttributes(raw)
	raw = fixHalfQuotedAttributes(raw)
	raw = escapeInvalidEntities(raw)

	doc := etree.NewDocument()
	// fmt.Println("📤 Sanitized XML preview:")
	// fmt.Println(raw)
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

// OffsetCellIDs modifies all mxCell id and parent attributes to avoid collisions.
func OffsetCellIDs(raw string, offset int) (string, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(raw); err != nil {
		return "", fmt.Errorf("failed to parse XML for ID offsetting: %w", err)
	}

	for _, cell := range doc.FindElements("//mxCell") {
		if idAttr := cell.SelectAttr("id"); idAttr != nil {
			if idInt, err := strconv.Atoi(idAttr.Value); err == nil {
				idAttr.Value = strconv.Itoa(idInt + offset)
			}
		}
		if parentAttr := cell.SelectAttr("parent"); parentAttr != nil {
			if parentInt, err := strconv.Atoi(parentAttr.Value); err == nil {
				parentAttr.Value = strconv.Itoa(parentInt + offset)
			}
		}
	}

	return doc.WriteToString()
}

// DetectEscapedFillColors scans XML and reports all fillColor attributes that are improperly escaped.
func DetectEscapedFillColors(raw string) ([]string, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(raw); err != nil {
		return nil, fmt.Errorf("failed to parse XML for fillColor check: %w", err)
	}

	var offenders []string
	pattern := regexp.MustCompile(`fillColor=&quot;#[0-9a-fA-F]{6}&quot;`)

	for _, cell := range doc.FindElements("//mxCell") {
		if styleAttr := cell.SelectAttr("style"); styleAttr != nil {
			if pattern.MatchString(styleAttr.Value) {
				offenders = append(offenders, styleAttr.Value)
			}
		}
	}

	return offenders, nil
}

// ForceQuoteAllAttributes is a last-resort fix to quote any attr that looks like key=value
func ForceQuoteAllAttributes(xml string) string {
	re := regexp.MustCompile(`(<\w+[^>]*?)\s+([a-zA-Z_:]+)=([^\s"'/>]+)`)
	for {
		newXML := re.ReplaceAllString(xml, `$1 $2="$3"`)
		if newXML == xml {
			break
		}
		xml = newXML
	}
	return xml
}

// fixHalfQuotedAttributes detects values starting with a quote but missing the end quote
func fixHalfQuotedAttributes(xml string) string {
	re := regexp.MustCompile(`\b([a-zA-Z_:]+)="([^"]*?)(\s+[a-zA-Z_:]+=)`)
	count := 0

	xml = re.ReplaceAllStringFunc(xml, func(match string) string {
		// Add missing closing quote before next attribute
		count++
		return regexp.MustCompile(`="([^"]*?)(\s+[a-zA-Z_:]+=)`).ReplaceAllString(match, `="$1"$2`)
	})

	if count > 0 {
		fmt.Printf("🩹 Fixed %d half-quoted attribute(s)\n", count)
	}
	return xml
}
