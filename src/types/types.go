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
	return len(c.Issues) > 0
}

// UserStory defines a single user story
type UserStory struct {
	ID               string   `yaml:"id"`
	Title            string   `yaml:"title"`
	Narrative        string   `yaml:"narrative"`
	SharedComponents []string `yaml:"shared_components,omitempty"`
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
