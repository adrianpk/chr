package generation

import (
	"fmt"
	"strings"
)

const (
	// AcceptanceFeatureStatusGenerated defines a package boundary constant.
	AcceptanceFeatureStatusGenerated = "generated"
	// AcceptanceFeatureStatusNeedsMoreInfo defines a package boundary constant.
	AcceptanceFeatureStatusNeedsMoreInfo = "needs_more_info"
	// AcceptanceFeatureStatusRejected defines a package boundary constant.
	AcceptanceFeatureStatusRejected = "rejected"
	// AcceptanceFeatureStatusNotApplicable defines a package boundary constant.
	AcceptanceFeatureStatusNotApplicable = "not_applicable"
)

// AcceptanceFeatureGeneration describes optional model-generated feature content.
type AcceptanceFeatureGeneration struct {
	Status               string
	Body                 string
	RejectionReason      string
	ClarificationPrompts []AcceptanceFeaturePrompt
}

// AcceptanceFeaturePrompt describes one default-oriented feature clarification prompt.
type AcceptanceFeaturePrompt struct {
	ID            string `json:"id"`
	Question      string `json:"question"`
	DefaultAnswer string `json:"default_answer"`
	Reason        string `json:"reason"`
	Answer        string `json:"answer,omitempty"`
}

// HasContent reports whether generated feature content is present.
func (g AcceptanceFeatureGeneration) HasContent() bool {
	return strings.TrimSpace(g.Status) != "" ||
		strings.TrimSpace(g.Body) != "" ||
		strings.TrimSpace(g.RejectionReason) != ""
}

// ValidateAcceptanceFeatureGeneration validates the corresponding package boundary value.
func ValidateAcceptanceFeatureGeneration(feature AcceptanceFeatureGeneration) (AcceptanceFeatureGeneration, error) {
	normalized := AcceptanceFeatureGeneration{
		Status:               strings.TrimSpace(feature.Status),
		Body:                 strings.TrimSpace(feature.Body),
		RejectionReason:      strings.TrimSpace(feature.RejectionReason),
		ClarificationPrompts: NormalizeAcceptanceFeaturePrompts(feature.ClarificationPrompts),
	}

	switch normalized.Status {
	case AcceptanceFeatureStatusGenerated:
		if normalized.Body == "" {
			return AcceptanceFeatureGeneration{}, fmt.Errorf("acceptance feature generated status requires body")
		}

		normalized.ClarificationPrompts = nil
	case AcceptanceFeatureStatusNeedsMoreInfo:
		if normalized.RejectionReason == "" {
			return AcceptanceFeatureGeneration{}, fmt.Errorf("acceptance feature needs_more_info status requires reason")
		}

		if len(normalized.ClarificationPrompts) == 0 {
			return AcceptanceFeatureGeneration{}, fmt.Errorf("acceptance feature needs_more_info status requires clarification prompts")
		}

		normalized.Body = ""
	case AcceptanceFeatureStatusRejected:
		if normalized.RejectionReason == "" {
			return AcceptanceFeatureGeneration{}, fmt.Errorf("acceptance feature rejected status requires reason")
		}

		normalized.Body = ""
		normalized.ClarificationPrompts = nil
	case AcceptanceFeatureStatusNotApplicable:
		normalized.Body = ""
		normalized.ClarificationPrompts = nil
	default:
		return AcceptanceFeatureGeneration{}, fmt.Errorf("acceptance feature status is invalid: %s", normalized.Status)
	}

	return normalized, nil
}

// NormalizeAcceptanceFeaturePrompts normalizes the shared optional clarification prompts.
func NormalizeAcceptanceFeaturePrompts(prompts []AcceptanceFeaturePrompt) []AcceptanceFeaturePrompt {
	normalized := make([]AcceptanceFeaturePrompt, 0, len(prompts))
	seen := map[string]bool{}

	for _, prompt := range prompts {
		cleaned := AcceptanceFeaturePrompt{
			ID:            strings.TrimSpace(prompt.ID),
			Question:      strings.TrimSpace(prompt.Question),
			DefaultAnswer: strings.TrimSpace(prompt.DefaultAnswer),
			Reason:        strings.TrimSpace(prompt.Reason),
			Answer:        strings.TrimSpace(prompt.Answer),
		}

		if cleaned.ID == "" || cleaned.Question == "" || cleaned.DefaultAnswer == "" {
			continue
		}

		if seen[cleaned.ID] {
			continue
		}

		seen[cleaned.ID] = true
		normalized = append(normalized, cleaned)
	}

	return normalized
}
