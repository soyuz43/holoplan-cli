// src/validator/zones.go
package validator

import (
	"fmt"
	"strings"
	"unicode"
)

// ──────────────────────────────────────────────
// RULE: Semantic Zone Conformance
// ──────────────────────────────────────────────

func checkSemanticZones(cells []mxCell) error {
	if len(cells) == 0 {
		return nil
	}

	// Determine the layout's own vertical extent. Zones are measured relative
	// to the layout top (min Y) and bottom (max Y) rather than absolute zero,
	// since the builder prompt begins layouts at y="100".
	var top, maxY float64
	for i, c := range cells {
		if i == 0 || c.Geometry.Y < top {
			top = c.Geometry.Y
		}
		yBottom := c.Geometry.Y + c.Geometry.Height
		if yBottom > maxY {
			maxY = yBottom
		}
	}

	for _, c := range cells {
		yMid := c.Geometry.Y + c.Geometry.Height/2
		label := semanticLabel(c)

		switch roleOf(label) {
		case "nav":
			// Navbars belong at the top of the layout. Measure from the
			// layout's own top so a topmost navbar in a layout that starts
			// lower down (e.g. y="100") is not flagged as misplaced.
			if yMid-top > 0.1*(maxY-top) {
				return fmt.Errorf("🧭 navbar (%s) should be near the top", c.ID)
			}
		case "modal":
			if yMid < 0.3*maxY || yMid > 0.7*maxY {
				return fmt.Errorf("🧭 modal (%s) should be centered", c.ID)
			}
		case "footer":
			if yMid < 0.9*maxY {
				return fmt.Errorf("🧭 footer (%s) should be at the bottom", c.ID)
			}
		}
	}

	debugLog("✅ Semantic zone checks passed")
	return nil
}

// roleOf returns the semantic zone role implied by a label, matching whole
// words only. Substring matching (e.g. "foot" inside "Football Scores")
// produced false positives, so we tokenize the label and compare each token
// against known role words.
func roleOf(label string) string {
	for _, tok := range tokenize(label) {
		switch tok {
		case "nav", "navbar", "navigation":
			return "nav"
		case "modal":
			return "modal"
		case "footer":
			return "footer"
		}
	}
	return ""
}

// tokenize splits a label into lowercase word tokens on spaces, hyphens,
// underscores, slashes, and dots.
func tokenize(label string) []string {
	return strings.FieldsFunc(label, func(r rune) bool {
		return unicode.IsSpace(r) || r == '-' || r == '_' || r == '/' || r == '.'
	})
}

// semanticLabel returns the human-readable semantic label for a cell, preferring
// the draw.io `value` attribute and falling back to the cell `id`. This ensures
// zone rules fire on the labels the builder actually emits (e.g. value="Navbar")
// rather than on numeric IDs that carry no semantic meaning.
func semanticLabel(c mxCell) string {
	if c.Value != "" {
		return strings.ToLower(c.Value)
	}
	return strings.ToLower(c.ID)
}
