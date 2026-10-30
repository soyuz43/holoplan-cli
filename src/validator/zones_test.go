package validator

import (
	"strings"
	"testing"
)

func makeCell(id, value string, x, y, w, h float64) mxCell {
	return mxCell{
		ID:      id,
		Value:   value,
		Vertex:  "1",
		Parent:  "1",
		Geometry: struct {
			X      float64 `xml:"x,attr"`
			Y      float64 `xml:"y,attr"`
			Width  float64 `xml:"width,attr"`
			Height float64 `xml:"height,attr"`
		}{
			X:      x,
			Y:      y,
			Width:  w,
			Height: h,
		},
	}
}

func TestCheckSemanticZones(t *testing.T) {
	tests := []struct {
		name        string
		cells       []mxCell
		expectError string
	}{
		{
			name:   "navbar at top passes",
			cells: []mxCell{
				makeCell("3", "Navbar", 0, 0, 100, 30),
				makeCell("4", "Body", 0, 100, 100, 200),
			},
		},
		{
			name:        "navbar too low fails",
			cells: []mxCell{
				makeCell("3", "Navbar", 0, 500, 100, 30),
				makeCell("4", "Body", 0, 100, 100, 200),
			},
			expectError: "navbar",
		},
		{
			name:   "modal centered passes",
			cells: []mxCell{
				makeCell("3", "Modal", 0, 300, 100, 100),
				makeCell("4", "Body", 0, 0, 100, 600),
			},
		},
		{
			name:        "modal too high fails",
			cells: []mxCell{
				makeCell("3", "Modal", 0, 50, 100, 100),
				makeCell("4", "Body", 0, 0, 100, 600),
			},
			expectError: "modal",
		},
		{
			name:        "modal too low fails",
			cells: []mxCell{
				makeCell("3", "Modal", 0, 500, 100, 100),
				makeCell("4", "Body", 0, 0, 100, 600),
			},
			expectError: "modal",
		},
		{
			name:   "footer near bottom passes",
			cells: []mxCell{
				makeCell("3", "Body", 0, 0, 100, 100),
				makeCell("4", "Footer", 0, 850, 100, 100),
			},
		},
		{
			name:        "footer too high fails",
			cells: []mxCell{
				makeCell("3", "Body", 0, 0, 100, 100),
				makeCell("4", "Footer", 0, 300, 100, 100),
			},
			expectError: "footer",
		},
		{
			name:   "numeric id with semantic value fires zone rule",
			cells: []mxCell{
				makeCell("3", "Navbar", 0, 500, 100, 30),
				makeCell("4", "Body", 0, 100, 100, 200),
			},
			expectError: "navbar",
		},
		{
			name:   "non-zone elements ignored",
			cells: []mxCell{
				makeCell("3", "Submit Button", 0, 100, 100, 30),
				makeCell("4", "Body", 0, 200, 100, 200),
			},
		},
		{
			name:   "empty cell slice passes",
			cells:  []mxCell{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkSemanticZones(tt.cells)
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
