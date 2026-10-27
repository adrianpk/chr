package generation

import (
	"fmt"
	"strings"
	"time"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// GeneratedPlanSliceDetail describes a package boundary value.
type GeneratedPlanSliceDetail struct {
	Scope                  quizcore.Scope
	SliceID                string
	ShortName              string
	Objective              string
	ImplementationNotes    []string
	Tasks                  []GeneratedPlanSliceTask
	AcceptanceCriteria     []string
	TicketMetadata         GeneratedPlanSliceTicketMetadata
	BranchMetadata         GeneratedPlanSliceBranchMetadata
	PRExpectations         []string
	ValidationExpectations []string
	Dependencies           []string
	NonGoals               []string
	CreatedAt              time.Time
}

// GeneratedPlanSliceTask describes a package boundary value.
type GeneratedPlanSliceTask struct {
	ID         string
	Summary    string
	CommitHint string
}

// GeneratedPlanSliceTicketMetadata describes a package boundary value.
type GeneratedPlanSliceTicketMetadata struct {
	Title    string
	Synopsis string
	Priority string
	Severity string
	Type     string
	Tags     []string
}

// GeneratedPlanSliceBranchMetadata describes a package boundary value.
type GeneratedPlanSliceBranchMetadata struct {
	BranchName string
}

// HasContent executes the corresponding package operation.
func (d GeneratedPlanSliceDetail) HasContent() bool {
	return strings.TrimSpace(d.SliceID) != "" ||
		strings.TrimSpace(d.ShortName) != "" ||
		len(d.Tasks) > 0
}

// ValidateGeneratedPlanSliceDetailContent validates the corresponding package boundary value.
func ValidateGeneratedPlanSliceDetailContent(detail GeneratedPlanSliceDetail) error {
	if err := quizcore.ValidateScope(detail.Scope); err != nil {
		return err
	}

	requiredFields := []struct {
		value string
		name  string
	}{
		{value: detail.SliceID, name: "slice id"},
		{value: detail.ShortName, name: "short name"},
		{value: detail.Objective, name: "objective"},
	}

	for _, field := range requiredFields {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("plan slice detail %s is required", field.name)
		}
	}

	if len(detail.ImplementationNotes) == 0 {
		return fmt.Errorf("plan slice detail implementation notes are required")
	}

	if len(detail.Tasks) == 0 {
		return fmt.Errorf("plan slice detail tasks are required")
	}

	if len(detail.AcceptanceCriteria) == 0 {
		return fmt.Errorf("plan slice detail acceptance criteria are required")
	}

	if err := validateGeneratedPlanSliceDetailLists(detail); err != nil {
		return err
	}

	if err := validateGeneratedPlanSliceTicketMetadata(detail.TicketMetadata); err != nil {
		return err
	}

	if strings.TrimSpace(detail.BranchMetadata.BranchName) == "" {
		return fmt.Errorf("plan slice detail branch name is required")
	}

	return nil
}

func validateGeneratedPlanSliceDetailLists(detail GeneratedPlanSliceDetail) error {
	if err := validateGeneratedPlanSliceTasks(detail.Tasks); err != nil {
		return err
	}

	requiredLists := []struct {
		values []string
		name   string
	}{
		{values: detail.PRExpectations, name: "PR expectations"},
		{values: detail.ValidationExpectations, name: "validation expectations"},
	}

	for _, list := range requiredLists {
		if len(list.values) == 0 {
			return fmt.Errorf("plan slice detail %s are required", list.name)
		}
	}

	return nil
}

// ValidateGeneratedPlanSliceDetailAgainstOutline validates the corresponding package boundary value.
func ValidateGeneratedPlanSliceDetailAgainstOutline(detail GeneratedPlanSliceDetail, outline GeneratedPlanOutline, targetSliceID string) error {
	if err := ValidateGeneratedPlanSliceDetailContent(detail); err != nil {
		return err
	}

	target, ok := outline.Slice(targetSliceID)
	if !ok {
		return fmt.Errorf("plan outline slice is unknown: %s", targetSliceID)
	}

	if detail.SliceID != target.ID {
		return fmt.Errorf("plan slice detail id = %q, want %q", detail.SliceID, target.ID)
	}

	if detail.ShortName != target.ShortName {
		return fmt.Errorf("plan slice detail short name = %q, want %q", detail.ShortName, target.ShortName)
	}

	if detail.BranchMetadata.BranchName != target.BranchName {
		return fmt.Errorf("plan slice detail branch name = %q, want %q", detail.BranchMetadata.BranchName, target.BranchName)
	}

	return nil
}

func validateGeneratedPlanSliceTasks(tasks []GeneratedPlanSliceTask) error {
	for index, task := range tasks {
		if strings.TrimSpace(task.ID) == "" {
			return fmt.Errorf("plan slice detail task %d id is required", index+1)
		}

		if strings.TrimSpace(task.Summary) == "" {
			return fmt.Errorf("plan slice detail task %d summary is required", index+1)
		}
	}

	return nil
}

func validateGeneratedPlanSliceTicketMetadata(metadata GeneratedPlanSliceTicketMetadata) error {
	requiredFields := []struct {
		value string
		name  string
	}{
		{value: metadata.Title, name: "ticket title"},
		{value: metadata.Synopsis, name: "ticket synopsis"},
		{value: metadata.Priority, name: "ticket priority"},
		{value: metadata.Severity, name: "ticket severity"},
		{value: metadata.Type, name: "ticket type"},
	}

	for _, field := range requiredFields {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("plan slice detail %s is required", field.name)
		}
	}

	return nil
}
