// src/validator/layout.go
package validator

import (
	"encoding/xml"
	"errors"
	"fmt"
	"sort"
	"strings"

	"holoplan-cli/src/shared"
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

func CheckLayout(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return errors.New("❌ layout check aborted: input XML is empty or blank")
	}

	sanitized, err := shared.SanitizeXML(raw)
	if err != nil {
		return fmt.Errorf("❌ failed sanitizing XML: %w", err)
	}

	var model mxGraphModel
	decoder := xml.NewDecoder(strings.NewReader(sanitized))
	if err := decoder.Decode(&model); err != nil {
		return fmt.Errorf("❌ XML parsing failed after sanitize: %w", err)
	}

	var renderables []mxCell
	for _, cell := range model.Cells {
		if cell.Vertex == "1" {
			renderables = append(renderables, cell)
		}
	}

	fmt.Printf("🔎 Validating %d visible elements\n", len(renderables))

	if err := checkCollisions(renderables); err != nil {
		return err
	}
	if err := checkVerticalFlow(renderables); err != nil {
		return err
	}
	if err := checkSemanticZones(renderables); err != nil {
		return err
	}

	return nil
}

// ──────────────────────────────────────────────
// 📐 RULE 1: Collision Detection
// ──────────────────────────────────────────────

func checkCollisions(cells []mxCell) error {
	for i, a := range cells {
		debugLog("🔍 [%s] (x=%.1f, y=%.1f, w=%.1f, h=%.1f)\n",
			a.ID, a.Geometry.X, a.Geometry.Y, a.Geometry.Width, a.Geometry.Height)

		for j := i + 1; j < len(cells); j++ {
			b := cells[j]
			if boxesOverlap(a, b) {
				return fmt.Errorf("🚫 layout collision: %s overlaps with %s", a.ID, b.ID)
			}
		}
	}
	fmt.Println("✅ No collisions detected")
	return nil
}

func boxesOverlap(a, b mxCell) bool {
	ax1, ay1 := a.Geometry.X, a.Geometry.Y
	ax2, ay2 := ax1+a.Geometry.Width, ay1+a.Geometry.Height

	bx1, by1 := b.Geometry.X, b.Geometry.Y
	bx2, by2 := bx1+b.Geometry.Width, by1+b.Geometry.Height

	return ax1 < bx2 && ax2 > bx1 && ay1 < by2 && ay2 > by1
}

// ──────────────────────────────────────────────
// 📏 RULE 2: Vertical Flow
// ──────────────────────────────────────────────

func checkVerticalFlow(cells []mxCell) error {
	sorted := make([]mxCell, len(cells))
	copy(sorted, cells)

	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Geometry.Y < sorted[j].Geometry.Y
	})

	for i := 0; i < len(cells); i++ {
		expected := sorted[i].ID
		actual := cells[i].ID
		if expected != actual {
			debugLog("🔀 Flow mismatch at position %d: expected %s but got %s\n", i, expected, actual)
			return fmt.Errorf("↕️ vertical flow error: element %s appears before %s", actual, expected)
		}
	}

	debugLog("✅ Top-down flow validated: %d elements in order\n", len(cells))
	return nil
}
