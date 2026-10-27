package core

import (
	"encoding/json"
	"fmt"
	"reflect"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func checkWorkflowAmendmentAddition(builder workflowAmendmentBuild, slice generationcore.GeneratedWorkflowSlice, input generationcore.GeneratedWorkflowTask) error {
	for _, descriptor := range builder.old.Set.Slices {
		if descriptor.Phase != "pending" && slice.Order <= descriptor.Order {
			return fmt.Errorf("%w: new work precedes an execution or integration frontier", workflowstate.ErrConflict)
		}

		if descriptor.ID != slice.ID {
			continue
		}

		for _, task := range builder.tasks {
			if task.Task.SliceID == slice.ID && (task.Task.Phase != "pending" || input.Order <= task.Task.Order) {
				return fmt.Errorf("%w: an existing slice accepts only later unstarted tasks", workflowstate.ErrConflict)
			}
		}
	}

	return nil
}

func (s *WorkflowStore) buildWorkflowAmendmentSlices(builder *workflowAmendmentBuild, plan generationcore.GeneratedWorkflowPlan) error {
	old := map[string]WorkflowSlice{}
	for _, descriptor := range builder.old.Set.Slices {
		old[descriptor.ID] = descriptor
	}

	builder.result.Set.Set.Slices = nil
	seen := 0

	for _, input := range plan.Slices {
		descriptor, retained, err := buildWorkflowAmendedSlice(builder, old, input)
		if err != nil {
			return err
		}

		seen += retained

		builder.result.Set.Set.Slices = append(builder.result.Set.Set.Slices, descriptor)
	}

	if seen != len(old) {
		return fmt.Errorf("%w: amendment removed an existing slice", workflowstate.ErrConflict)
	}

	return nil
}

func buildWorkflowAmendedSlice(builder *workflowAmendmentBuild, old map[string]WorkflowSlice, input generationcore.GeneratedWorkflowSlice) (WorkflowSlice, int, error) {
	descriptor, err := materializationSlice(input, builder.result.Set, builder.result.Tasks, builder.keys)
	if err != nil {
		return descriptor, 0, err
	}

	if previous, exists := old[input.ID]; exists {
		descriptor, err = workflowAmendedSlice(previous, descriptor)

		return descriptor, 1, err
	}

	err = workflowNewAmendmentSlice(builder, input, &descriptor)

	return descriptor, 0, err
}

func workflowAmendedSlice(previous, descriptor WorkflowSlice) (WorkflowSlice, error) {
	if previous.Order != descriptor.Order || previous.ContractHash != descriptor.ContractHash || !reflect.DeepEqual(previous.DependsOn, descriptor.DependsOn) {
		return descriptor, fmt.Errorf("%w: amendment changed an existing slice contract or order", workflowstate.ErrConflict)
	}

	previous.TaskInputHash = descriptor.TaskInputHash

	return previous, nil
}

func workflowNewAmendmentSlice(builder *workflowAmendmentBuild, input generationcore.GeneratedWorkflowSlice, descriptor *WorkflowSlice) error {
	data, err := json.Marshal(workflowSliceContract{Contract: input.Contract, Validation: input.Validation})
	if err != nil {
		return err
	}

	descriptor.ContractRef = builder.root + "/slice-" + input.ID + ".json"
	builder.files = append(builder.files, workflowstate.FileChange{Path: descriptor.ContractRef, After: data})

	return nil
}
