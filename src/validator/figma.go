package validator

import (
	"encoding/json"
	"fmt"
	"strings"

	"holoplan-cli/src/types"
)

var validFigmaNodeTypes = map[string]bool{
	"FRAME":     true,
	"RECTANGLE": true,
	"TEXT":      true,
	"GROUP":     true,
	"COMPONENT": true,
}

// CheckFigmaLayout validates the structural schema and geometric layout
// of a generated Figma JSON document.
func CheckFigmaLayout(rawJSON string) error {
	if strings.TrimSpace(rawJSON) == "" {
		return fmt.Errorf("figma layout check aborted: input JSON is empty or blank")
	}

	var doc types.FigmaDocument
	if err := json.Unmarshal([]byte(rawJSON), &doc); err != nil {
		return fmt.Errorf("figma JSON parsing failed: %w", err)
	}

	if doc.Document.Type != "FRAME" {
		return fmt.Errorf("figma root node type must be FRAME, got %q", doc.Document.Type)
	}
	if doc.Document.ID != "0:1" {
		return fmt.Errorf("figma root node ID must be \"0:1\", got %q", doc.Document.ID)
	}

	seenIDs := make(map[string]bool)
	if err := validateNode(&doc.Document, nil, seenIDs); err != nil {
		return err
	}

	var boxes []types.FigmaNode
	collectBoxes(&doc.Document, &boxes)
	if err := checkFigmaCollisions(boxes); err != nil {
		return err
	}

	return nil
}

// validateNode recursively checks schema conformance for a node and its children.
// parent is used to skip parent-child containment from collision detection later;
// here it is unused but kept for future extension.
func validateNode(node *types.FigmaNode, parent *types.FigmaNode, seenIDs map[string]bool) error {
	if node.ID == "" {
		return fmt.Errorf("figma node has empty id (name: %q)", node.Name)
	}
	if seenIDs[node.ID] {
		return fmt.Errorf("figma duplicate node id: %s", node.ID)
	}
	seenIDs[node.ID] = true

	if node.Name == "" {
		return fmt.Errorf("figma node %s has empty name", node.ID)
	}
	if !validFigmaNodeTypes[node.Type] {
		return fmt.Errorf("figma node %s has invalid type %q", node.ID, node.Type)
	}
	if node.AbsoluteBoundingBox == nil {
		return fmt.Errorf("figma node %s (%s) missing absoluteBoundingBox", node.ID, node.Name)
	}
	if node.Visible == nil {
		return fmt.Errorf("figma node %s (%s) missing visible field", node.ID, node.Name)
	}
	if !*node.Visible {
		return fmt.Errorf("figma node %s (%s) has visible=false", node.ID, node.Name)
	}

	for i := range node.Children {
		if err := validateNode(&node.Children[i], node, seenIDs); err != nil {
			return err
		}
	}
	return nil
}

// collectBoxes gathers all nodes with a bounding box for collision detection.
func collectBoxes(node *types.FigmaNode, out *[]types.FigmaNode) {
	if node.AbsoluteBoundingBox != nil {
		*out = append(*out, *node)
	}
	for i := range node.Children {
		collectBoxes(&node.Children[i], out)
	}
}

// checkFigmaCollisions detects overlapping sibling bounding boxes.
// Parent-child containment is not treated as a collision.
func checkFigmaCollisions(nodes []types.FigmaNode) error {
	for i := 0; i < len(nodes); i++ {
		a := nodes[i].AbsoluteBoundingBox
		for j := i + 1; j < len(nodes); j++ {
			// Skip parent-child pairs (containment is expected).
			if isAncestor(nodes[i], nodes[j]) || isAncestor(nodes[j], nodes[i]) {
				continue
			}
			b := nodes[j].AbsoluteBoundingBox
			if boxesOverlapFigma(a, b) {
				return fmt.Errorf("figma layout collision: node %s (%s) overlaps node %s (%s)",
					nodes[i].ID, nodes[i].Name, nodes[j].ID, nodes[j].Name)
			}
		}
	}
	return nil
}

// isAncestor reports whether ancestor contains descendant in the tree by ID chain.
// Since we only have flat copies, we approximate by checking if one node's box fully
// contains the other AND one is a parent in the tree structure. However, we don't
// retain tree structure here, so we use full containment as a proxy for parent-child.
func isAncestor(a, b types.FigmaNode) bool {
	ab := a.AbsoluteBoundingBox
	bb := b.AbsoluteBoundingBox
	if ab == nil || bb == nil {
		return false
	}
	return ab.X <= bb.X && ab.Y <= bb.Y &&
		ab.X+ab.Width >= bb.X+bb.Width &&
		ab.Y+ab.Height >= bb.Y+bb.Height
}

func boxesOverlapFigma(a, b *types.FigmaBoundingBox) bool {
	return a.X < b.X+b.Width && a.X+a.Width > b.X &&
		a.Y < b.Y+b.Height && a.Y+a.Height > b.Y
}
