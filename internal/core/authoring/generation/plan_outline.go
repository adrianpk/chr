package generation

import (
	"fmt"
	"strings"
	"time"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// GeneratedPlanOutline describes a package boundary value.
type GeneratedPlanOutline struct {
	Scope              quizcore.Scope
	Title              string
	Status             string
	Summary            string
	Assumptions        []string
	OpenQuestions      []string
	Risks              []string
	ValidationStrategy string
	Slices             []GeneratedPlanOutlineSlice
	Dependencies       []GeneratedPlanOutlineDependency
	CreatedAt          time.Time
}

// GeneratedPlanOutlineSlice describes a package boundary value.
type GeneratedPlanOutlineSlice struct {
	ID                     string
	ShortName              string
	Objective              string
	Rationale              string
	Dependencies           []string
	BranchName             string
	PRTitle                string
	ReportPath             string
	Complexity             string
	ImplementationBoundary string
}

// GeneratedPlanOutlineDependency describes a package boundary value.
type GeneratedPlanOutlineDependency struct {
	FromSliceID string
	ToSliceID   string
	Reason      string
}

// HasContent executes the corresponding package operation.
func (o GeneratedPlanOutline) HasContent() bool {
	return strings.TrimSpace(o.Title) != "" ||
		strings.TrimSpace(o.Summary) != "" ||
		len(o.Slices) > 0
}

// Slice executes the corresponding package operation.
func (o GeneratedPlanOutline) Slice(sliceID string) (GeneratedPlanOutlineSlice, bool) {
	for _, slice := range o.Slices {
		if slice.ID == sliceID {
			return slice, true
		}
	}

	return GeneratedPlanOutlineSlice{}, false
}

// ValidateGeneratedPlanOutline validates the corresponding package boundary value.
func ValidateGeneratedPlanOutline(outline GeneratedPlanOutline) error {
	if err := quizcore.ValidateScope(outline.Scope); err != nil {
		return err
	}

	requiredFields := []struct {
		value string
		name  string
	}{
		{value: outline.Title, name: "title"},
		{value: outline.Summary, name: "summary"},
		{value: outline.ValidationStrategy, name: "validation strategy"},
	}

	for _, field := range requiredFields {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("plan outline %s is required", field.name)
		}
	}

	if strings.TrimSpace(outline.Status) != "planned" {
		return fmt.Errorf("plan outline status must be planned")
	}

	if len(outline.Slices) == 0 {
		return fmt.Errorf("plan outline slices are required")
	}

	if err := validateGeneratedPlanOutlineSlices(outline.Slices); err != nil {
		return err
	}

	return validateGeneratedPlanOutlineDependencies(outline.Dependencies)
}

func validateGeneratedPlanOutlineSlices(slices []GeneratedPlanOutlineSlice) error {
	seen := map[string]bool{}

	for index, slice := range slices {
		expectedID := fmt.Sprintf("Slice %d", index+1)
		if strings.TrimSpace(slice.ID) != expectedID {
			return fmt.Errorf("plan outline slice %d id = %q, want %q", index+1, slice.ID, expectedID)
		}

		if seen[slice.ID] {
			return fmt.Errorf("plan outline duplicate slice id: %s", slice.ID)
		}

		seen[slice.ID] = true

		if err := validateGeneratedPlanOutlineSlice(slice); err != nil {
			return fmt.Errorf("plan outline %s: %w", expectedID, err)
		}
	}

	return nil
}

func validateGeneratedPlanOutlineDependencies(dependencies []GeneratedPlanOutlineDependency) error {
	for index, dependency := range dependencies {
		if strings.TrimSpace(dependency.FromSliceID) == "" {
			return fmt.Errorf("plan outline dependency %d from slice id is required", index+1)
		}

		if strings.TrimSpace(dependency.ToSliceID) == "" {
			return fmt.Errorf("plan outline dependency %d to slice id is required", index+1)
		}

		if strings.TrimSpace(dependency.Reason) == "" {
			return fmt.Errorf("plan outline dependency %d reason is required", index+1)
		}
	}

	return nil
}

func validateGeneratedPlanOutlineSlice(slice GeneratedPlanOutlineSlice) error {
	requiredFields := []struct {
		value string
		name  string
	}{
		{value: slice.ShortName, name: "short name"},
		{value: slice.Objective, name: "objective"},
		{value: slice.Rationale, name: "rationale"},
		{value: slice.BranchName, name: "branch name"},
		{value: slice.PRTitle, name: "PR title"},
		{value: slice.ReportPath, name: "report path"},
		{value: slice.Complexity, name: "complexity"},
		{value: slice.ImplementationBoundary, name: "implementation boundary"},
	}

	for _, field := range requiredFields {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("%s is required", field.name)
		}
	}

	return nil
}
