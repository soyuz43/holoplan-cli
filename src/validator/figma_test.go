package validator

import (
	"fmt"
	"strings"
	"testing"
)

func figmaNode(id, name, nodeType string, x, y, w, h float64, visible bool) string {
	return fmt.Sprintf(`{
		"id": %q,
		"name": %q,
		"type": %q,
		"absoluteBoundingBox": {"x": %g, "y": %g, "width": %g, "height": %g},
		"visible": %t
	}`, id, name, nodeType, x, y, w, h, visible)
}

func figmaDoc(root string, children ...string) string {
	if len(children) > 0 {
		root = strings.Replace(root,
			`"visible": true`,
			`"visible": true, "children": [`+strings.Join(children, ",")+`]`,
			1)
	}
	return fmt.Sprintf(`{
		"schemaVersion": 0,
		"document": %s,
		"components": {},
		"textStyles": {}
	}`, root)
}

func TestCheckFigmaLayout(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError string
	}{
		{
			name: "valid minimal document with one child",
			input: figmaDoc(
				figmaNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true),
				figmaNode("0:2", "Submit Button", "TEXT", 100, 100, 200, 40, true),
			),
		},
		{
			name: "valid nested tree",
			input: figmaDoc(
				figmaNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true),
				fmt.Sprintf(`{"id":"0:2","name":"Group","type":"GROUP","absoluteBoundingBox":{"x":10,"y":10,"width":300,"height":200},"visible":true,"children":[%s]}`,
					figmaNode("0:3", "Label", "TEXT", 15, 15, 100, 20, true)),
			),
		},
		{
			name: "parent-child containment is not a collision",
			input: figmaDoc(
				figmaNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true),
				figmaNode("0:2", "Card", "RECTANGLE", 50, 50, 300, 200, true),
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
			input: figmaDoc(
				figmaNode("0:1", "Root Group", "GROUP", 0, 0, 800, 600, true),
			),
			expectError: `root node type must be FRAME`,
		},
		{
			name: "root ID is not 0:1",
			input: figmaDoc(
				figmaNode("9:9", "Root Frame", "FRAME", 0, 0, 800, 600, true),
			),
			expectError: `root node ID must be`,
		},
		{
			name: "missing absoluteBoundingBox on child",
			input: figmaDoc(
				figmaNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true),
				`{"id": "0:2", "name": "No Box", "type": "TEXT", "visible": true}`,
			),
			expectError: "missing absoluteBoundingBox",
		},
		{
			name: "missing visible field on child",
			input: figmaDoc(
				figmaNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true),
				`{"id": "0:2", "name": "Invisible?", "type": "TEXT", "absoluteBoundingBox": {"x": 10, "y": 10, "width": 100, "height": 20}}`,
			),
			expectError: "missing visible field",
		},
		{
			name: "visible false on child",
			input: figmaDoc(
				figmaNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true),
				`{"id": "0:2", "name": "Hidden", "type": "TEXT", "absoluteBoundingBox": {"x": 10, "y": 10, "width": 100, "height": 20}, "visible": false}`,
			),
			expectError: "has visible=false",
		},
		{
			name: "invalid node type BUTTON",
			input: figmaDoc(
				figmaNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true),
				figmaNode("0:2", "My Button", "BUTTON", 10, 10, 100, 30, true),
			),
			expectError: `invalid type "BUTTON"`,
		},
		{
			name: "duplicate IDs across subtrees",
			input: figmaDoc(
				figmaNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true),
				figmaNode("0:2", "Label A", "TEXT", 10, 10, 100, 20, true),
				figmaNode("0:2", "Label B", "TEXT", 200, 10, 100, 20, true),
			),
			expectError: "duplicate node id: 0:2",
		},
		{
			name: "empty id on child",
			input: figmaDoc(
				figmaNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true),
				figmaNode("", "Anonymous", "TEXT", 10, 10, 100, 20, true),
			),
			expectError: "empty id",
		},
		{
			name: "empty name on child",
			input: figmaDoc(
				figmaNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true),
				figmaNode("0:2", "", "TEXT", 10, 10, 100, 20, true),
			),
			expectError: "empty name",
		},
		{
			name: "sibling overlap detected",
			input: figmaDoc(
				figmaNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true),
				figmaNode("0:2", "Box A", "RECTANGLE", 10, 10, 100, 100, true),
				figmaNode("0:3", "Box B", "RECTANGLE", 50, 50, 100, 100, true),
			),
			expectError: "overlaps node",
		},
		{
			name: "adjacent boxes do not overlap",
			input: figmaDoc(
				figmaNode("0:1", "Root Frame", "FRAME", 0, 0, 800, 600, true),
				figmaNode("0:2", "Left Box", "RECTANGLE", 0, 0, 100, 50, true),
				figmaNode("0:3", "Right Box", "RECTANGLE", 100, 0, 100, 50, true),
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
				if !strings.Contains(err.Error(), tt.expectError) {
					t.Fatalf("error %q does not contain expected substring %q", err.Error(), tt.expectError)
				}
			}
		})
	}
}
