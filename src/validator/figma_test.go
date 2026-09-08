package validator

import (
	"encoding/json"
	"testing"

	"holoplan-cli/src/types"
)

func makeNode(id, name, nodeType string, x, y, w, h float64, visible bool, children ...types.FigmaNode) types.FigmaNode {
	visiblePtr := visible
	return types.FigmaNode{
		ID:                  id,
		Name:                name,
		Type:                nodeType,
		AbsoluteBoundingBox: &types.FigmaBoundingBox{X: x, Y: y, Width: w, Height: h},
		Visible:             &visiblePtr,
		Children:            children,
	}
}

func makeDoc(root types.FigmaNode) string {
	doc := types.FigmaDocument{
		SchemaVersion: 0,
		Document:      root,
		Components:    json.RawMessage("{}"),
		Styles:        json.RawMessage("{}"),
	}
	data, _ := json.Marshal(doc)
	return string(data)
}

func TestCheckFigmaLayout(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError string
	}{
		{
			name: "valid minimal document with one child",
			input: makeDoc(
				makeNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true,
					makeNode("0:2", "Submit Button", "TEXT", 100, 100, 200, 40, true),
				),
			),
		},
		{
			name: "valid nested tree",
			input: makeDoc(
				makeNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true,
					makeNode("0:2", "Group", "GROUP", 10, 10, 300, 200, true,
						makeNode("0:3", "Label", "TEXT", 15, 15, 100, 20, true),
					),
				),
			),
		},
		{
			name: "parent-child containment is not a collision",
			input: makeDoc(
				makeNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true,
					makeNode("0:2", "Card", "RECTANGLE", 50, 50, 300, 200, true),
				),
			),
		},
		{
			name:        "malformed JSON",
			input:       `{invalid`,
			expectError: "figma JSON parsing failed",
		},
		{
			name:        "empty input",
			input:       "",
			expectError: "input JSON is empty or blank",
		},
		{
			name: "root type is not FRAME",
			input: makeDoc(
				makeNode("0:1", "Root Group", "GROUP", 0, 0, 800, 600, true),
			),
			expectError: `root node type must be FRAME`,
		},
		{
			name: "root ID is not 0:1",
			input: makeDoc(
				makeNode("9:9", "Root Frame", "FRAME", 0, 0, 800, 600, true),
			),
			expectError: `root node ID must be`,
		},
		// P2: Explicit root validation tests
		{
			name: "root missing absoluteBoundingBox",
			input: makeDoc(
				types.FigmaNode{
					ID:     "0:1",
					Name:   "Root Frame",
					Type:   "FRAME",
					Visible: func() *bool { v := true; return &v }(),
				},
			),
			expectError: "root node must have absoluteBoundingBox",
		},
		{
			name: "root missing visible field",
			input: makeDoc(
				types.FigmaNode{
					ID:                  "0:1",
					Name:                "Root Frame",
					Type:                "FRAME",
					AbsoluteBoundingBox: &types.FigmaBoundingBox{X: 0, Y: 0, Width: 800, Height: 600},
				},
			),
			expectError: "root node must be visible",
		},
		{
			name: "root visible false",
			input: makeDoc(
				types.FigmaNode{
					ID:                  "0:1",
					Name:                "Root Frame",
					Type:                "FRAME",
					AbsoluteBoundingBox: &types.FigmaBoundingBox{X: 0, Y: 0, Width: 800, Height: 600},
					Visible:             func() *bool { v := false; return &v }(),
				},
			),
			expectError: "root node must be visible",
		},
		{
			name: "missing absoluteBoundingBox on child",
			input: makeDoc(
				makeNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true,
					types.FigmaNode{
						ID:     "0:2",
						Name:   "No Box",
						Type:   "TEXT",
						Visible: func() *bool { v := true; return &v }(),
					},
				),
			),
			expectError: "missing absoluteBoundingBox",
		},
		{
			name: "missing visible field on child",
			input: makeDoc(
				makeNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true,
					types.FigmaNode{
						ID:                  "0:2",
						Name:                "Invisible?",
						Type:                "TEXT",
						AbsoluteBoundingBox: &types.FigmaBoundingBox{X: 10, Y: 10, Width: 100, Height: 20},
					},
				),
			),
			expectError: "missing visible field",
		},
		{
			name: "visible false on child",
			input: makeDoc(
				makeNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true,
					types.FigmaNode{
						ID:                  "0:2",
						Name:                "Hidden",
						Type:                "TEXT",
						AbsoluteBoundingBox: &types.FigmaBoundingBox{X: 10, Y: 10, Width: 100, Height: 20},
						Visible:             func() *bool { v := false; return &v }(),
					},
				),
			),
			expectError: "has visible=false",
		},
		{
			name: "invalid node type BUTTON",
			input: makeDoc(
				makeNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true,
					makeNode("0:2", "My Button", "BUTTON", 10, 10, 100, 30, true),
				),
			),
			expectError: `invalid type "BUTTON"`,
		},
		{
			name: "duplicate IDs across subtrees",
			input: makeDoc(
				makeNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true,
					makeNode("0:2", "Label A", "TEXT", 10, 10, 100, 20, true),
					makeNode("0:2", "Label B", "TEXT", 200, 10, 100, 20, true),
				),
			),
			expectError: "duplicate node id: 0:2",
		},
		{
			name: "empty id on child",
			input: makeDoc(
				makeNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true,
					makeNode("", "Anonymous", "TEXT", 10, 10, 100, 20, true),
				),
			),
			expectError: "empty id",
		},
		{
			name: "empty name on child",
			input: makeDoc(
				makeNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true,
					makeNode("0:2", "", "TEXT", 10, 10, 100, 20, true),
				),
			),
			expectError: "empty name",
		},
		{
			name: "sibling overlap detected",
			input: makeDoc(
				makeNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true,
					makeNode("0:2", "Box A", "RECTANGLE", 10, 10, 100, 100, true),
					makeNode("0:3", "Box B", "RECTANGLE", 50, 50, 100, 100, true),
				),
			),
			expectError: "overlaps node",
		},
		{
			name: "adjacent boxes do not overlap",
			input: makeDoc(
				makeNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true,
					makeNode("0:2", "Left Box", "RECTANGLE", 0, 0, 100, 50, true),
					makeNode("0:3", "Right Box", "RECTANGLE", 100, 0, 100, 50, true),
				),
			),
		},
		// P1: New test cases for fixed collision detection
		{
			name: "child at negative coordinates (overflows parent) is not a collision",
			input: makeDoc(
				makeNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true,
					makeNode("0:2", "Child", "RECTANGLE", -50, -50, 200, 200, true),
				),
			),
		},
		{
			name: "child extending beyond parent bounds is not a collision",
			input: makeDoc(
				makeNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true,
					makeNode("0:2", "Wide Child", "RECTANGLE", 700, 100, 300, 200, true),
				),
			),
		},
		{
			name: "unrelated nodes with geometric containment but no tree relationship is collision",
			input: makeDoc(
				makeNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true,
					makeNode("0:2", "Large Box", "RECTANGLE", 10, 10, 400, 400, true),
					makeNode("0:3", "Small Box", "RECTANGLE", 50, 50, 100, 100, true),
				),
			),
			expectError: "overlaps node",
		},
		{
			name: "deep nesting: grandparent-parent-child all containment checked",
			input: makeDoc(
				makeNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true,
					makeNode("0:2", "Parent", "GROUP", 100, 100, 300, 300, true,
						makeNode("0:3", "Child", "RECTANGLE", 150, 150, 100, 100, true),
					),
				),
			),
		},
		{
			name: "overlapping cousins (different parents) detected as collision",
			input: makeDoc(
				makeNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true,
					makeNode("0:2", "Parent A", "GROUP", 10, 10, 200, 200, true,
						makeNode("0:3", "Child A", "RECTANGLE", 20, 20, 150, 150, true),
					),
					makeNode("0:4", "Parent B", "GROUP", 100, 100, 200, 200, true,
						makeNode("0:5", "Child B", "RECTANGLE", 120, 120, 150, 150, true),
					),
				),
			),
			expectError: "overlaps node",
		},
		{
			name: "non-overlapping cousins (different parents) no collision",
			input: makeDoc(
				makeNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true,
					makeNode("0:2", "Parent A", "GROUP", 10, 10, 100, 100, true,
						makeNode("0:3", "Child A", "RECTANGLE", 20, 20, 50, 50, true),
					),
					makeNode("0:4", "Parent B", "GROUP", 200, 200, 100, 100, true,
						makeNode("0:5", "Child B", "RECTANGLE", 220, 220, 50, 50, true),
					),
				),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckFigmaLayout(tt.input)
			if tt.expectError == "" {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !contains(err.Error(), tt.expectError) {
					t.Fatalf("error %q does not contain expected substring %q", err.Error(), tt.expectError)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
