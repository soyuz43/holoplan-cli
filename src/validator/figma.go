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

// CollisionNode holds only the fields needed for collision detection.
type CollisionNode struct {
	ID                  string
	Name                string
	Type                string
	AbsoluteBoundingBox *types.FigmaBoundingBox
	Visible             *bool
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

	// P2: Explicit root node validation with clear error messages
	if doc.Document.Type != "FRAME" {
		return fmt.Errorf("figma root node type must be FRAME, got %q", doc.Document.Type)
	}
	if doc.Document.ID != "0:1" {
		return fmt.Errorf("figma root node ID must be \"0:1\", got %q", doc.Document.ID)
	}
	if doc.Document.AbsoluteBoundingBox == nil {
		return fmt.Errorf("figma root node must have absoluteBoundingBox")
	}
	if doc.Document.Visible == nil || !*doc.Document.Visible {
		return fmt.Errorf("figma root node must be visible")
	}

	seenIDs := make(map[string]bool)
	ancestry := make(map[string]string) // nodeID -> parentID
	var collisionNodes []CollisionNode

	// Combined validation and collection in a single traversal
	if err := validateAndCollect(&doc.Document, "", seenIDs, &collisionNodes, &ancestry); err != nil {
		return err
	}

	if err := checkFigmaCollisions(collisionNodes, ancestry); err != nil {
		return err
	}

	return nil
}

// validateAndCollect recursively validates schema and collects collision nodes in one pass.
func validateAndCollect(node *types.FigmaNode, parentID string, seenIDs map[string]bool, collisionNodes *[]CollisionNode, ancestry *map[string]string) error {
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

	// Collect for collision detection (only needed fields, no Children slice)
	*collisionNodes = append(*collisionNodes, CollisionNode{
		ID:                  node.ID,
		Name:                node.Name,
		Type:                node.Type,
		AbsoluteBoundingBox: node.AbsoluteBoundingBox,
		Visible:             node.Visible,
	})
	(*ancestry)[node.ID] = parentID

	for i := range node.Children {
		if err := validateAndCollect(&node.Children[i], node.ID, seenIDs, collisionNodes, ancestry); err != nil {
			return err
		}
	}
	return nil
}

// checkFigmaCollisions detects overlapping sibling bounding boxes.
// Parent-child containment is not treated as a collision.
func checkFigmaCollisions(nodes []CollisionNode, ancestry map[string]string) error {
	for i := 0; i < len(nodes); i++ {
		a := nodes[i].AbsoluteBoundingBox
		for j := i + 1; j < len(nodes); j++ {
			// Skip parent-child pairs (containment is expected).
			if isDescendant(nodes[i].ID, nodes[j].ID, ancestry) || isDescendant(nodes[j].ID, nodes[i].ID, ancestry) {
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

// isDescendant reports whether descendant is a child (at any depth) of ancestor in the tree.
func isDescendant(descendantID, ancestorID string, ancestry map[string]string) bool {
	current := descendantID
	for {
		parent, ok := ancestry[current]
		if !ok || parent == "" {
			return false // reached root without finding ancestor
		}
		if parent == ancestorID {
			return true
		}
		current = parent
	}
}

func boxesOverlapFigma(a, b *types.FigmaBoundingBox) bool {
	return a.X < b.X+b.Width && a.X+a.Width > b.X &&
		a.Y < b.Y+b.Height && a.Y+a.Height > b.Y
}
