package generation

import (
	"fmt"
	"strings"
	"time"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// PlanSourceBinding retains source identities and revisions without copying
// source bodies into the plan's metadata.
type PlanSourceBinding struct{ Kind, Reference, Revision string }

// GeneratedPlan holds staged model output and caller-owned publication guards.
type GeneratedPlan struct {
	Operation             *PlanPublicationOperation `json:"-"`
	RequireOrderMap       bool                      `json:"-"`
	TrackerReplaceHash    string                    `json:"-"`
	Order                 []PlanOrderSlice          `json:"order,omitempty"`
	CreateOnly            bool                      `json:"-"`
	ExpectedRevision      string                    `json:"-"`
	SourceObservationHash string                    `json:"-"`
	SourceInputs          []PlanSourceBinding       `json:"-"`
	ArchiveAt             time.Time                 `json:"-"`
	Scope                 quizcore.Scope
	Title                 string
	Body                  string
	SourcePaths           []string
	SourceRequestID       string
	SourceResponseID      string
	CreatedAt             time.Time
}

// ValidateGeneratedPlan validates the corresponding package boundary value.
func ValidateGeneratedPlan(plan GeneratedPlan) error {
	if strings.TrimSpace(plan.Title) == "" {
		return fmt.Errorf("generated plan title is required")
	}

	if strings.TrimSpace(plan.Body) == "" {
		return fmt.Errorf("generated plan body is required")
	}

	err := quizcore.ValidateScope(plan.Scope)
	if err != nil {
		return err
	}

	if plan.Operation != nil || plan.RequireOrderMap || len(plan.Order) > 0 {
		if err := ValidatePlannedOrder(plan.Order); err != nil {
			return err
		}
	}

	if plan.CreatedAt.IsZero() {
		return fmt.Errorf("generated plan created time is required")
	}

	return nil
}

// ValidateGeneratedPlanModelContent validates the corresponding package boundary value.
func ValidateGeneratedPlanModelContent(plan GeneratedPlan) error {
	if strings.TrimSpace(plan.Title) == "" {
		return fmt.Errorf("generated plan title is required")
	}

	if strings.TrimSpace(plan.Body) == "" {
		return fmt.Errorf("generated plan body is required")
	}

	return quizcore.ValidateScope(plan.Scope)
}

// HasContent executes the corresponding package operation.
func (p GeneratedPlan) HasContent() bool {
	return strings.TrimSpace(p.Title) != "" || strings.TrimSpace(p.Body) != ""
}
