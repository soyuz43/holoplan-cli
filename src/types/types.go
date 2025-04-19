package types

type Critique struct {
	Issues []string
}

func (c Critique) HasIssues() bool {
	return len(c.Issues) > 0
}

type UserStory struct {
	ID        string `yaml:"id"`
	Title     string `yaml:"title"`
	Narrative string `yaml:"narrative"`
}

type ViewPlan struct {
	StoryID   string       `json:"story_id"`
	Views     []ViewLayout `json:"views"`
	Reasoning string       `json:"reasoning,omitempty"` // optional LLM explanation
}

type ViewLayout struct {
	Name       string   `json:"name"`                 // e.g., "HomePage"
	Type       string   `json:"type"`                 // "primary", "modal", etc.
	Components []string `json:"components,omitempty"` // optional list from chunking
}

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
