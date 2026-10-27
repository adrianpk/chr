package core

import (
	"fmt"
	"strings"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
)

// AssembleGeneratedPlanBody executes the corresponding package operation.
func AssembleGeneratedPlanBody(
	outline generationcore.GeneratedPlanOutline,
	details []generationcore.GeneratedPlanSliceDetail,
	requestIDs []string,
	responseIDs []string,
) (string, error) {
	if err := generationcore.ValidateGeneratedPlanOutline(outline); err != nil {
		return "", err
	}

	detailsBySliceID := map[string]generationcore.GeneratedPlanSliceDetail{}

	for _, detail := range details {
		if err := generationcore.ValidateGeneratedPlanSliceDetailAgainstOutline(detail, outline, detail.SliceID); err != nil {
			return "", err
		}

		detailsBySliceID[detail.SliceID] = detail
	}

	for _, slice := range outline.Slices {
		if _, ok := detailsBySliceID[slice.ID]; !ok {
			return "", fmt.Errorf("plan slice detail is missing for %s", slice.ID)
		}
	}

	var output strings.Builder

	writePlanOverview(&output, outline)
	writePlanSliceDecomposition(&output, outline, detailsBySliceID)
	writePlanOrdering(&output, outline)
	writePlanDependencyAnalysis(&output, outline)
	writePlanExecutionStrategy(&output, outline)
	writePlanValidationCriteria(&output, outline, detailsBySliceID)
	writePlanModelProvenance(&output, requestIDs, responseIDs)

	return strings.TrimSpace(output.String()), nil
}

func writePlanOverview(output *strings.Builder, outline generationcore.GeneratedPlanOutline) {
	output.WriteString("## High-level implementation strategy\n\n")
	output.WriteString(strings.TrimSpace(outline.Summary))
	output.WriteString("\n\n")

	writePlanStringList(output, "## Assumptions and Open Questions", append(
		append([]string{}, outline.Assumptions...),
		outline.OpenQuestions...,
	))
	writePlanStringList(output, "## Risks", outline.Risks)
}

func writePlanSliceDecomposition(
	output *strings.Builder,
	outline generationcore.GeneratedPlanOutline,
	detailsBySliceID map[string]generationcore.GeneratedPlanSliceDetail,
) {
	output.WriteString("## Slice decomposition\n\n")

	for _, slice := range outline.Slices {
		detail := detailsBySliceID[slice.ID]

		output.WriteString("### " + slice.ID + ": " + slice.ShortName + "\n\n")
		output.WriteString("Objective: " + strings.TrimSpace(slice.Objective) + "\n\n")
		output.WriteString("Rationale: " + strings.TrimSpace(slice.Rationale) + "\n\n")
		output.WriteString("Implementation boundary: " + strings.TrimSpace(slice.ImplementationBoundary) + "\n\n")
		output.WriteString("Complexity: " + strings.TrimSpace(slice.Complexity) + "\n\n")

		writePlanStringList(output, "Implementation notes", detail.ImplementationNotes)
		writePlanTasks(output, detail.Tasks)
		writePlanStringList(output, "Acceptance criteria", detail.AcceptanceCriteria)
		writePlanStringList(output, "Non-goals", detail.NonGoals)
	}
}

func writePlanOrdering(output *strings.Builder, outline generationcore.GeneratedPlanOutline) {
	output.WriteString("## Slice ordering\n\n")

	for index, slice := range outline.Slices {
		fmt.Fprintf(output, "%d. %s: %s\n", index+1, slice.ID, slice.ShortName)
	}

	output.WriteString("\n")
}

func writePlanDependencyAnalysis(output *strings.Builder, outline generationcore.GeneratedPlanOutline) {
	output.WriteString("## Slice dependency analysis\n\n")

	if len(outline.Dependencies) == 0 {
		output.WriteString("No cross-slice dependencies recorded.\n\n")

		return
	}

	for _, dependency := range outline.Dependencies {
		output.WriteString("- ")
		output.WriteString(dependency.FromSliceID)
		output.WriteString(" depends on ")
		output.WriteString(dependency.ToSliceID)
		output.WriteString(": ")
		output.WriteString(dependency.Reason)
		output.WriteString("\n")
	}

	output.WriteString("\n")
}

func writePlanExecutionStrategy(output *strings.Builder, outline generationcore.GeneratedPlanOutline) {
	output.WriteString("## Branch strategy\n\n")

	for _, slice := range outline.Slices {
		output.WriteString("- " + slice.ID + ": `" + slice.BranchName + "`\n")
	}

	output.WriteString("\n## Pull request strategy\n\n")

	for _, slice := range outline.Slices {
		output.WriteString("- " + slice.ID + ": `" + slice.PRTitle + "`\n")
	}

	output.WriteString("\n## Merge sequencing\n\n")
	output.WriteString("Merge slices in the listed order unless a later dependency analysis update proves they are independent.\n\n")
	output.WriteString("## Review requirements\n\n")
	output.WriteString("Each slice should be reviewed independently against its implementation boundary, validation expectations, and report path.\n\n")
}

func writePlanValidationCriteria(
	output *strings.Builder,
	outline generationcore.GeneratedPlanOutline,
	detailsBySliceID map[string]generationcore.GeneratedPlanSliceDetail,
) {
	output.WriteString("## Completion validation criteria\n\n")
	output.WriteString(strings.TrimSpace(outline.ValidationStrategy))
	output.WriteString("\n\n")

	for _, slice := range outline.Slices {
		detail := detailsBySliceID[slice.ID]

		writePlanStringList(output, slice.ID+" validation expectations", detail.ValidationExpectations)
	}
}

func writePlanModelProvenance(output *strings.Builder, requestIDs, responseIDs []string) {
	output.WriteString("## Model call provenance\n\n")
	writePlanStringList(output, "Request IDs", requestIDs)
	writePlanStringList(output, "Response IDs", responseIDs)
}

func writePlanStringList(output *strings.Builder, title string, values []string) {
	output.WriteString(title + "\n\n")

	if len(values) == 0 {
		output.WriteString("None recorded.\n\n")

		return
	}

	for _, value := range values {
		output.WriteString("- ")
		output.WriteString(strings.TrimSpace(value))
		output.WriteString("\n")
	}

	output.WriteString("\n")
}

func writePlanTasks(output *strings.Builder, tasks []generationcore.GeneratedPlanSliceTask) {
	output.WriteString("Tasks inside slice\n\n")

	for _, task := range tasks {
		output.WriteString("- " + task.ID + ": " + task.Summary)

		if strings.TrimSpace(task.CommitHint) != "" {
			output.WriteString(" (" + task.CommitHint + ")")
		}

		output.WriteString("\n")
	}

	output.WriteString("\n")
}
