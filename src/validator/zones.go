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
		id := strings.ToLower(c.ID)

		switch {
		case strings.Contains(id, "nav"):
			if yMid > 0.1*maxY {
				return fmt.Errorf("🧭 navbar (%s) should be near the top", c.ID)
			}
		case strings.Contains(id, "modal"):
			if yMid < 0.3*maxY || yMid > 0.7*maxY {
				return fmt.Errorf("🧭 modal (%s) should be centered", c.ID)
			}
		case strings.Contains(id, "foot"):
			if yMid < 0.9*maxY {
				return fmt.Errorf("🧭 footer (%s) should be at the bottom", c.ID)
			}
		}
	}

	debugLog("✅ Semantic zone checks passed")
	return nil
}
