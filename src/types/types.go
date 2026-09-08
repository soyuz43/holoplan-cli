// src\types\types.go
package types

import (
	"encoding/json"
	"fmt"
)

// Critique holds validation or review issues
type Critique struct {
	Issues []string
}

func (c Critique) HasIssues() bool {
	return len(c.Issues) > 0 && !(len(c.Issues) == 1 && c.Issues[0] == "no issues")
}

// UserStory defines a single user story
type UserStory struct {
	ID                string   `yaml:"id"`
	Title             string   `yaml:"title"`
	Narrative         string   `yaml:"narrative"`
	InteractionOrigin string   `yaml:"interaction_origin,omitempty"` // e.g., "plant_detail"
	View              string   `yaml:"view,omitempty"`               // for single-view stories
	Views             []string `yaml:"views,omitempty"`              // for multi-view stories
	ResultingView     string   `yaml:"resulting_view,omitempty"`     // if action produces a new view
	SharedComponents  []string `yaml:"shared_components,omitempty"`  // persistent UI elements
}

// Components is a custom type that unmarshals from either a string array
// or an array of { "component": string } objects
type Components []string

func (c *Components) UnmarshalJSON(data []byte) error {
	// Try simple list of strings first
	var simple []string
	if err := json.Unmarshal(data, &simple); err == nil {
		*c = simple
		return nil
	}

	// Try list of maps with "component" key
	var kvList []map[string]string
	if err := json.Unmarshal(data, &kvList); err == nil {
		var extracted []string
		for _, kv := range kvList {
			if val, ok := kv["component"]; ok {
				extracted = append(extracted, val)
			}
		}
		*c = extracted
		return nil
	}

	return fmt.Errorf("components must be either an array of strings or an array of {component: string} objects")
}

// ViewPlan is the structured plan produced from a user story
type ViewPlan struct {
	StoryID   string       `json:"story_id"`
	Views     []ViewLayout `json:"views"`
	Reasoning string       `json:"reasoning,omitempty"` // optional LLM explanation
}

// ViewLayout defines a single visual component hierarchy
type ViewLayout struct {
	Name       string     `json:"name"`                 // e.g., "HomePage"
	Type       string     `json:"type"`                 // e.g., "primary", "modal"
	Narrative  string     `json:"narrative"`            // specific slice of the story
	Components Components `json:"components,omitempty"` // flexible parsing
}

// AuditReport captures violations from a visual audit
type AuditReport struct {
	ViewName           string   `json:"view"`
	MissingElements    []string `json:"missing_elements"`
	SemanticMismatches []string `json:"semantic_mismatches"`
	StyleViolations    []string `json:"style_violations"`
	Pass               bool     `json:"pass"`
}

func (a AuditReport) HasIssues() bool {
	return !a.Pass
}

// FigmaBoundingBox mirrors absoluteBoundingBox in the Figma JSON output.
type FigmaBoundingBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// FigmaNode is a single node in the Figma document tree.
type FigmaNode struct {
	ID                  string            `json:"id"`
	Name                string            `json:"name"`
	Type                string            `json:"type"`
	AbsoluteBoundingBox *FigmaBoundingBox `json:"absoluteBoundingBox"`
	Visible             *bool             `json:"visible"`
	Characters          string            `json:"characters,omitempty"`
	Children            []FigmaNode       `json:"children,omitempty"`
}

// FigmaDocument is the top-level structure produced by the Figma builder.
type FigmaDocument struct {
	SchemaVersion int             `json:"schemaVersion"`
	Document      FigmaNode       `json:"document"`
	Components    json.RawMessage `json:"components"`
	Styles        json.RawMessage `json:"styles"`
}
