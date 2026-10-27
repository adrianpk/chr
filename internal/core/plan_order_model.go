package core

import (
	"strings"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// PlannedOrderFromDetails retains structured model identities without parsing
// generated Markdown as authority for task identity or execution.
func PlannedOrderFromDetails(details []GeneratedPlanSliceDetail) []PlanOrderSlice {
	slices := make([]PlanOrderSlice, 0, len(details))
	for index, detail := range details {
		slice := PlanOrderSlice{ID: strings.ReplaceAll(strings.ToLower(detail.SliceID), " ", "-"), Order: index + 1, Title: detail.ShortName}
		for taskIndex, task := range detail.Tasks {
			slice.Tasks = append(slice.Tasks, PlanOrderTask{ID: "task-" + strings.ToLower(task.ID), Order: taskIndex + 1, Title: task.Summary})
		}

		slices = append(slices, slice)
	}

	return slices
}

func planOrderJSONSchema() map[string]any {
	task := map[string]any{"type": "object", "additionalProperties": false, "required": []string{"id", "task_order", "title"}, "properties": map[string]any{"id": map[string]any{"type": "string"}, "task_order": map[string]any{"type": "integer", "minimum": 1}, "title": map[string]any{"type": "string"}}}
	slice := map[string]any{"type": "object", "additionalProperties": false, "required": []string{"id", "slice_order", "title", "tasks"}, "properties": map[string]any{"id": map[string]any{"type": "string"}, "slice_order": map[string]any{"type": "integer", "minimum": 1}, "title": map[string]any{"type": "string"}, "tasks": map[string]any{"type": "array", "minItems": 1, "maxItems": workflowstate.WorkflowPageLimit, "items": task}}}

	return map[string]any{"type": "array", "minItems": 1, "maxItems": workflowstate.WorkflowPageLimit, "items": slice}
}
