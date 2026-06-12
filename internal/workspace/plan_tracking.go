package workspace

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GeneratedPlanTracking describes a package boundary value.
type GeneratedPlanTracking struct {
	Scope       QuizScope
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

// GeneratedPlanTrackingPaths describes a package boundary value.
type GeneratedPlanTrackingPaths struct {
	ScopeID      string
	RelativePath string
	Path         string
}

// GeneratedPlanTrackingLayout executes the corresponding package operation.
func GeneratedPlanTrackingLayout(projectDir string, scope QuizScope) (GeneratedPlanTrackingPaths, error) {
	err := ValidateQuizScope(scope)
	if err != nil {
		return GeneratedPlanTrackingPaths{}, err
	}

	scopeID, err := quizScopeID(scope)
	if err != nil {
		return GeneratedPlanTrackingPaths{}, err
	}

	relativePath := UserRelativePath(PlanDirectory, scopeID+"-tracking.md")

	return GeneratedPlanTrackingPaths{
		ScopeID:      scopeID,
		RelativePath: relativePath,
		Path:         filepath.Join(projectDir, filepath.FromSlash(relativePath)),
	}, nil
}

// WriteGeneratedPlanTracking executes the corresponding package operation.
func WriteGeneratedPlanTracking(
	ctx context.Context,
	projectDir string,
	tracking GeneratedPlanTracking,
) (GeneratedPlanTrackingPaths, error) {
	err := ctx.Err()
	if err != nil {
		return GeneratedPlanTrackingPaths{}, err
	}

	err = ValidateGeneratedPlanTracking(tracking)
	if err != nil {
		return GeneratedPlanTrackingPaths{}, err
	}

	paths, err := GeneratedPlanTrackingLayout(projectDir, tracking.Scope)
	if err != nil {
		return GeneratedPlanTrackingPaths{}, err
	}

	err = os.MkdirAll(filepath.Dir(paths.Path), 0o755)
	if err != nil {
		return GeneratedPlanTrackingPaths{}, fmt.Errorf("create plan tracking directory: %w", err)
	}

	err = os.WriteFile(paths.Path, []byte(RenderGeneratedPlanTracking(tracking)), 0o644)
	if err != nil {
		return GeneratedPlanTrackingPaths{}, fmt.Errorf("write plan tracking: %w", err)
	}

	return paths, nil
}

// RenderGeneratedPlanTracking renders generated plan tracking as Markdown.
//
//nolint:wsl_v5 // Markdown rendering keeps adjacent writes together.
func RenderGeneratedPlanTracking(tracking GeneratedPlanTracking) string {
	var output bytes.Buffer

	output.WriteString("---\n")
	output.WriteString("title: " + yamlString(tracking.Title) + "\n")
	output.WriteString("project_id: " + yamlString(tracking.Scope.ProjectID) + "\n")
	output.WriteString("case_slug: " + yamlString(tracking.Scope.CaseSlug) + "\n")
	output.WriteString("status: " + yamlString("tracking") + "\n")
	output.WriteString("plan_path: " + yamlString(filepath.ToSlash(tracking.PlanPath)) + "\n")
	output.WriteString("created_at: " + yamlTimestamp(tracking.CreatedAt) + "\n")
	output.WriteString("---\n\n")
	output.WriteString("# " + tracking.Title + "\n\n")
	output.WriteString(strings.TrimSpace(tracking.Body) + "\n\n")
	output.WriteString("## Tickets\n\n")

	for _, ticket := range tracking.Tickets {
		output.WriteString("- ")
		output.WriteString(ticket.SliceID)
		output.WriteString(": ")
		output.WriteString(ticket.Key)
		output.WriteString(" - ")
		output.WriteString(ticket.Title)
		if strings.TrimSpace(ticket.Path) != "" {
			output.WriteString(" (`")
			output.WriteString(filepath.ToSlash(ticket.Path))
			output.WriteString("`)")
		}
		output.WriteString("\n")
	}

	output.WriteString("\n## Provenance\n\n")
	output.WriteString("- plan: " + filepath.ToSlash(tracking.PlanPath) + "\n")
	for _, path := range tracking.SourcePaths {
		output.WriteString("- source: " + filepath.ToSlash(path) + "\n")
	}

	return output.String()
}

// ValidateGeneratedPlanTracking validates the corresponding package boundary value.
func ValidateGeneratedPlanTracking(tracking GeneratedPlanTracking) error {
	if strings.TrimSpace(tracking.Title) == "" {
		return fmt.Errorf("plan tracking title is required")
	}

	if strings.TrimSpace(tracking.Body) == "" {
		return fmt.Errorf("plan tracking body is required")
	}

	if strings.TrimSpace(tracking.PlanPath) == "" {
		return fmt.Errorf("plan tracking plan path is required")
	}

	err := ValidateQuizScope(tracking.Scope)
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
	return strings.TrimSpace(t.Title) != "" || strings.TrimSpace(t.Body) != "" || len(t.Tickets) > 0
}
