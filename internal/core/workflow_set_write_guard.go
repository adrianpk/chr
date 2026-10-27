package core

import (
	"fmt"
	"reflect"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func guardWorkflowSetChange(old, next WorkflowSet) error {
	if err := guardBoundWorkflowSetOrder(old, next); err != nil {
		return err
	}

	if old.Status == "completed" {
		return fmt.Errorf("completed set requires an explicit approved reopening")
	}

	if old.Status == "closing" || !reflect.DeepEqual(old.Closing, next.Closing) {
		return fmt.Errorf("set closing ownership requires a fenced operation")
	}

	if old.ClosureRef != next.ClosureRef || !reflect.DeepEqual(old.ClosureHistory, next.ClosureHistory) || !reflect.DeepEqual(old.Amendments, next.Amendments) {
		return fmt.Errorf("set closure and amendment evidence require an owned operation")
	}

	return nil
}

func guardBoundWorkflowSetOrder(old, next WorkflowSet) error {
	if old.PlanRevision == nil {
		return nil
	}

	if !reflect.DeepEqual(old.PlanRevision, next.PlanRevision) || old.PlanRef != next.PlanRef || old.PlanHash != next.PlanHash {
		return fmt.Errorf("%w: plan revision requires approved amendment", workflowstate.ErrConflict)
	}

	before, after := boundWorkflowSliceOrder(old), boundWorkflowSliceOrder(next)
	if !reflect.DeepEqual(before, after) {
		return fmt.Errorf("%w: ordered hierarchy requires approved amendment", workflowstate.ErrConflict)
	}

	return nil
}

func boundWorkflowSliceOrder(set WorkflowSet) []generationcore.PlanOrderSlice {
	result := make([]generationcore.PlanOrderSlice, 0, len(set.Slices))
	for _, slice := range set.Slices {
		result = append(result, generationcore.PlanOrderSlice{ID: slice.ID, Order: slice.Order, Title: slice.Title})
	}

	return generationcore.OrderedPlanSlices(result)
}
