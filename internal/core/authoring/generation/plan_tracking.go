package generation

import (
	"fmt"
	"strings"
	"time"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// GeneratedPlanTracking describes a package boundary value.
type GeneratedPlanTracking struct {
	Workflow *GeneratedWorkflowPlan

	Scope       quizcore.Scope
	PlanPath    string
	Title       string
	Body        string
	Tickets     []GeneratedPlanTrackingTicket
	SourcePaths []string
	CreatedAt   time.Time
}

// GeneratedPlanTrackingTicket describes a package boundary value.
type GeneratedPlanTrackingTicket struct {
	Key          string
	Path         string
	SliceID      string
	Title        string
	Synopsis     string
	Summary      string
	Body         string
	Priority     string
	Severity     string
	Type         string
	Tags         []string
	Dependencies []string
}

// ValidateGeneratedPlanTracking validates the corresponding package boundary value.
func ValidateGeneratedPlanTracking(tracking GeneratedPlanTracking) error {
	if tracking.Workflow != nil {
		return ValidateGeneratedWorkflowPlan(*tracking.Workflow)
	}

	if strings.TrimSpace(tracking.Title) == "" {
		return fmt.Errorf("plan tracking title is required")
	}

	if strings.TrimSpace(tracking.Body) == "" {
		return fmt.Errorf("plan tracking body is required")
	}

	if strings.TrimSpace(tracking.PlanPath) == "" {
		return fmt.Errorf("plan tracking plan path is required")
	}

	err := quizcore.ValidateScope(tracking.Scope)
	if err != nil {
		return err
	}

	if tracking.CreatedAt.IsZero() {
		return fmt.Errorf("plan tracking created time is required")
	}

	if len(tracking.Tickets) == 0 {
		return fmt.Errorf("plan tracking tickets are required")
	}

	for index, ticket := range tracking.Tickets {
		err = ValidateGeneratedPlanTrackingTicket(ticket)
		if err != nil {
			return fmt.Errorf("plan tracking ticket %d: %w", index+1, err)
		}
	}

	return nil
}

// ValidateGeneratedPlanTrackingTicket validates the corresponding package boundary value.
func ValidateGeneratedPlanTrackingTicket(ticket GeneratedPlanTrackingTicket) error {
	if strings.TrimSpace(ticket.Key) == "" {
		return fmt.Errorf("ticket key is required")
	}

	if strings.TrimSpace(ticket.SliceID) == "" {
		return fmt.Errorf("slice id is required")
	}

	if strings.TrimSpace(ticket.Title) == "" {
		return fmt.Errorf("ticket title is required")
	}

	if strings.TrimSpace(ticket.Synopsis) == "" {
		return fmt.Errorf("ticket synopsis is required")
	}

	return nil
}

// ValidateGeneratedPlanTrackingModelContent validates the corresponding package boundary value.
func ValidateGeneratedPlanTrackingModelContent(tracking GeneratedPlanTracking) error {
	if tracking.Workflow != nil {
		return ValidateGeneratedWorkflowPlan(*tracking.Workflow)
	}

	if strings.TrimSpace(tracking.Title) == "" {
		return fmt.Errorf("plan tracking title is required")
	}

	if strings.TrimSpace(tracking.Body) == "" {
		return fmt.Errorf("plan tracking body is required")
	}

	if len(tracking.Tickets) == 0 {
		return fmt.Errorf("plan tracking tickets are required")
	}

	for index, ticket := range tracking.Tickets {
		if err := ValidateGeneratedPlanTrackingTicketModelContent(ticket); err != nil {
			return fmt.Errorf("plan tracking ticket %d: %w", index+1, err)
		}
	}

	return nil
}

// ValidateGeneratedPlanTrackingTicketModelContent validates the corresponding package boundary value.
func ValidateGeneratedPlanTrackingTicketModelContent(ticket GeneratedPlanTrackingTicket) error {
	if strings.TrimSpace(ticket.Key) != "" {
		return fmt.Errorf("key must be empty before materialization")
	}

	if strings.TrimSpace(ticket.Path) != "" {
		return fmt.Errorf("path must be empty before materialization")
	}

	if strings.TrimSpace(ticket.SliceID) == "" {
		return fmt.Errorf("slice id is required")
	}

	if strings.TrimSpace(ticket.Title) == "" {
		return fmt.Errorf("title is required")
	}

	if strings.TrimSpace(ticket.Synopsis) == "" {
		return fmt.Errorf("synopsis is required")
	}

	if strings.TrimSpace(ticket.Summary) == "" {
		return fmt.Errorf("summary is required")
	}

	if strings.TrimSpace(ticket.Body) == "" {
		return fmt.Errorf("body is required")
	}

	return nil
}

// HasContent executes the corresponding package operation.
func (t GeneratedPlanTracking) HasContent() bool {
	return t.Workflow != nil || strings.TrimSpace(t.Title) != "" || strings.TrimSpace(t.Body) != "" || len(t.Tickets) > 0
}
