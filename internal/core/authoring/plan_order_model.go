package authoring

import (
	"strings"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
)

// PlannedOrderFromDetails retains structured model identities without parsing
// generated Markdown as authority for task identity or execution.
func PlannedOrderFromDetails(details []generationcore.GeneratedPlanSliceDetail) []generationcore.PlanOrderSlice {
	slices := make([]generationcore.PlanOrderSlice, 0, len(details))
	for index, detail := range details {
		slice := generationcore.PlanOrderSlice{ID: strings.ReplaceAll(strings.ToLower(detail.SliceID), " ", "-"), Order: index + 1, Title: detail.ShortName}
		for taskIndex, task := range detail.Tasks {
			slice.Tasks = append(slice.Tasks, generationcore.PlanOrderTask{ID: "task-" + strings.ToLower(task.ID), Order: taskIndex + 1, Title: task.Summary})
		}

		slices = append(slices, slice)
	}

	return slices
}
