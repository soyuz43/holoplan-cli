package validator

import (
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
)

var Debug = false

func debugLog(format string, args ...interface{}) {
	if Debug {
		fmt.Printf(format, args...)
	}
}

type mxCell struct {
	ID       string `xml:"id,attr"`
	Value    string `xml:"value,attr"`
	Style    string `xml:"style,attr"`
	Vertex   string `xml:"vertex,attr"`
	Parent   string `xml:"parent,attr"`
	Geometry struct {
		X      float64 `xml:"x,attr"`
		Y      float64 `xml:"y,attr"`
		Width  float64 `xml:"width,attr"`
		Height float64 `xml:"height,attr"`
	} `xml:"mxGeometry"`
}

type mxGraphModel struct {
	Cells []mxCell `xml:"root>mxCell"`
}

func CheckLayout(xmlInput string) error {
	var model mxGraphModel

	decoder := xml.NewDecoder(strings.NewReader(xmlInput))
	err := decoder.Decode(&model)
	if err != nil {
		return fmt.Errorf("XML parsing failed: %w", err)
	}

	// Filter only renderable elements
	var renderables []mxCell
	for _, cell := range model.Cells {
		if cell.Vertex == "1" {
			renderables = append(renderables, cell)
		}
	}

	fmt.Printf("�� Validating %d visible elements\n", len(renderables))

	// 🔍 Rule 1: Collision detection (bounding boxes)
	if err := checkCollisions(renderables); err != nil {
		return err
	}

	// 🔃 Rule 2: Top-down Y-order flow
	if err := checkVerticalFlow(renderables); err != nil {
		return err
	}

	// 🧭 Rule 3: Semantic zone conformance
	if err := checkSemanticZones(renderables); err != nil {
		return err
	}

	return nil
}

// ──────────────────────────────────────────────
// 🧱 RULE STUBS
// ──────────────────────────────────────────────

func checkCollisions(cells []mxCell) error {
	for i, a := range cells {
		debugLog("🔍 [%s] (x=%.1f, y=%.1f, w=%.1f, h=%.1f)\n",
			a.ID, a.Geometry.X, a.Geometry.Y, a.Geometry.Width, a.Geometry.Height)

		for j := i + 1; j < len(cells); j++ {
			b := cells[j]
			if boxesOverlap(a, b) {
				return fmt.Errorf("layout collision: %s overlaps with %s", a.ID, b.ID)
			}
		}
	}
	fmt.Println("✅ No collisions detected")
	return nil
}

// Defined outside of checkCollisions
func boxesOverlap(a, b mxCell) bool {
	ax1, ay1 := a.Geometry.X, a.Geometry.Y
	ax2, ay2 := ax1+a.Geometry.Width, ay1+a.Geometry.Height

	bx1, by1 := b.Geometry.X, b.Geometry.Y
	bx2, by2 := bx1+b.Geometry.Width, by1+b.Geometry.Height

	return ax1 < bx2 && ax2 > bx1 && ay1 < by2 && ay2 > by1
}

// Check Vertical Alignment: top-down Y-order flow
// This function checks if the elements are in top-down order based on their Y coordinates.

func checkVerticalFlow(cells []mxCell) error {
	// Copy to avoid modifying original slice
	sorted := make([]mxCell, len(cells))
	copy(sorted, cells)

	// Sort top-down: lowest Y first
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Geometry.Y < sorted[j].Geometry.Y
	})

	for i := 0; i < len(cells); i++ {
		expected := sorted[i].ID
		actual := cells[i].ID
		if expected != actual {
			debugLog("🔀 Flow mismatch at position %d: expected %s but got %s\n", i, expected, actual)
			return fmt.Errorf("vertical flow error: element %s appears before %s", actual, expected)
		}
	}

	debugLog("✅ Top-down flow validated: %d elements in order\n", len(cells))
	return nil
}

func checkSemanticZones(cells []mxCell) error {
	if len(cells) == 0 {
		return nil
	}

	// Step 1: Determine layout canvas height (max Y + height)
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
				return fmt.Errorf("navbar (%s) should be near the top", c.ID)
			}
		case strings.Contains(id, "modal"):
			if yMid < 0.3*maxY || yMid > 0.7*maxY {
				return fmt.Errorf("modal (%s) should be centered", c.ID)
			}
		case strings.Contains(id, "foot"):
			if yMid < 0.9*maxY {
				return fmt.Errorf("footer (%s) should be at the bottom", c.ID)
			}
		}
	}

	debugLog("✅ Semantic zone checks passed")
	return nil
}
