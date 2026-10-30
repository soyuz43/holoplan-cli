// src/validator/zones.go
package validator

import (
	"fmt"
	"strings"
)

// ──────────────────────────────────────────────
// RULE: Semantic Zone Conformance
// ──────────────────────────────────────────────

func checkSemanticZones(cells []mxCell) error {
	if len(cells) == 0 {
		return nil
	}

	var maxY float64
	for _, c := range cells {
		yBottom := c.Geometry.Y + c.Geometry.Height
		if yBottom > maxY {
			maxY = yBottom
		}
	}

	for _, c := range cells {
		yMid := c.Geometry.Y + c.Geometry.Height/2
		label := semanticLabel(c)

		switch {
		case strings.Contains(label, "nav"):
			if yMid > 0.1*maxY {
				return fmt.Errorf("🧭 navbar (%s) should be near the top", c.ID)
			}
		case strings.Contains(label, "modal"):
			if yMid < 0.3*maxY || yMid > 0.7*maxY {
				return fmt.Errorf("🧭 modal (%s) should be centered", c.ID)
			}
		case strings.Contains(label, "foot"):
			if yMid < 0.9*maxY {
				return fmt.Errorf("🧭 footer (%s) should be at the bottom", c.ID)
			}
		}
	}

	debugLog("✅ Semantic zone checks passed")
	return nil
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
