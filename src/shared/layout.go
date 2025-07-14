// src/shared/layout.go
package shared

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/beevik/etree"
)

type box struct {
	Cell        *etree.Element
	X, Y        int
	Width       int
	Height      int
	Bottom, Top int
	Right, Left int
}

// ResolveOverlaps detects and fixes 2D bounding box overlaps by adjusting y-values
func ResolveOverlaps(xml string, margin int) (string, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xml); err != nil {
		return "", fmt.Errorf("failed to parse XML: %w", err)
	}

	var boxes []box

	for _, cell := range doc.FindElements("//mxCell") {
		if cell.SelectAttrValue("vertex", "") != "1" {
			continue
		}
		geom := cell.FindElement("mxGeometry")
		if geom == nil {
			continue
		}

		x := atoi(geom.SelectAttrValue("x", "0"))
		y := atoi(geom.SelectAttrValue("y", "0"))
		w := atoi(geom.SelectAttrValue("width", "0"))
		h := atoi(geom.SelectAttrValue("height", "0"))

		boxes = append(boxes, box{
			Cell:   cell,
			X:      x,
			Y:      y,
			Width:  w,
			Height: h,
			Left:   x,
			Right:  x + w,
			Top:    y,
			Bottom: y + h,
		})
	}

	// Sort by top Y value
	sort.SliceStable(boxes, func(i, j int) bool {
		return boxes[i].Top < boxes[j].Top
	})

	for i := 0; i < len(boxes); i++ {
		curr := &boxes[i]
		for j := 0; j < i; j++ {
			prev := &boxes[j]
			if isOverlapping(*prev, *curr) {
				// Adjust Y to be below previous
				newY := prev.Bottom + margin
				geom := curr.Cell.FindElement("mxGeometry")
				geom.RemoveAttr("y")
				geom.CreateAttr("y", strconv.Itoa(newY))

				// Update box geometry
				curr.Top = newY
				curr.Bottom = newY + curr.Height
			}
		}
	}

	out, err := doc.WriteToString()
	if err != nil {
		return "", fmt.Errorf("failed to serialize XML: %w", err)
	}
	return stripXMLDecl(out), nil
}

// Overlap test with bounding boxes
func isOverlapping(a, b box) bool {
	xOverlap := a.Left < b.Right && b.Left < a.Right
	yOverlap := a.Top < b.Bottom && b.Top < a.Bottom
	return xOverlap && yOverlap
}

// --- Helpers ---

func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func stripXMLDecl(xml string) string {
	return strings.TrimPrefix(xml, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
}
