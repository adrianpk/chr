package core

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"
)

type workflowAmendmentBuild struct {
	result WorkflowMaterializationResult
	old    WorkflowRecord
	tasks  map[string]WorkflowRecord
	keys   map[string]string
	files  []workflowFileChange
	number int
	root   string
}

func (s *WorkflowStore) buildWorkflowAmendment(ctx context.Context, set WorkflowRecord, request WorkflowSetAmendmentRequest, body []byte) (WorkflowMaterializationResult, []workflowFileChange, error) {
	builder, err := s.workflowAmendmentBuilder(ctx, set)
	if err != nil {
		return builder.result, nil, err
	}

	plan := orderedWorkflowPlan(request.Plan)
	if err := s.buildWorkflowAmendmentTasks(ctx, &builder, plan, request.ReserveTicketKey); err != nil {
		return builder.result, nil, err
	}

	if err := s.buildWorkflowAmendmentSlices(&builder, plan); err != nil {
		return builder.result, nil, err
	}

	if err := s.finishWorkflowAmendment(ctx, &builder, request, body); err != nil {
		return builder.result, nil, err
	}

	return builder.result, builder.files, nil
}

func (s *WorkflowStore) workflowAmendmentBuilder(ctx context.Context, set WorkflowRecord) (workflowAmendmentBuild, error) {
	builder := workflowAmendmentBuild{old: set, tasks: map[string]WorkflowRecord{}, keys: map[string]string{}, number: nextOpsArtifactNumber(s.layout.Root()), root: filepath.ToSlash(filepath.Join("plan", set.ID, "extension", "chr", fmt.Sprintf("amendment-%04d", set.Revision+1)))}

	data, err := json.Marshal(set.Set)
	if err != nil {
		return builder, err
	}

	// Retain the observed set separately; changing its slice cannot mutate the
	// prior image used by the guards and immutable amendment receipt.
	builder.result.Set = set

	builder.result.Set.Set = &WorkflowSet{}
	if err := json.Unmarshal(data, builder.result.Set.Set); err != nil {
		return builder, err
	}

	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return builder, err
	}

	for _, task := range snapshot.Records {
		if task.Task == nil || task.Task.SetID != set.ID {
			continue
		}

		id := ""
		if err := json.Unmarshal(task.Domain["plan_task_id"], &id); err != nil || id == "" {
			return builder, fmt.Errorf("%w: task has no stable plan identity", ErrWorkflowConflict)
		}

		builder.tasks[id], builder.keys[id] = task, task.Task.Key
	}

	return builder, nil
}

func (s *WorkflowStore) buildWorkflowAmendmentTasks(ctx context.Context, builder *workflowAmendmentBuild, plan GeneratedWorkflowPlan, allocate func(context.Context) (string, error)) error {
	seen := map[string]bool{}

	for _, slice := range plan.Slices {
		for _, input := range slice.Tasks {
			task, err := s.buildWorkflowAmendedTask(ctx, builder, slice, input, allocate)
			if err != nil {
				return err
			}

			seen[input.ID] = true

			builder.result.Tasks = append(builder.result.Tasks, task)
		}
	}

	if len(seen) != len(builder.keys) {
		return fmt.Errorf("%w: amendment removed an existing task", ErrWorkflowConflict)
	}

	for i := range builder.result.Tasks {
		task := &builder.result.Tasks[i]
		if task.SourceHash == "" {
			task.Task.DependsOn = resolveMaterializationRefs(task.Task.DependsOn, builder.keys)
		}
	}

	return nil
}

func (s *WorkflowStore) buildWorkflowAmendedTask(ctx context.Context, builder *workflowAmendmentBuild, slice GeneratedWorkflowSlice, input GeneratedWorkflowTask, allocate func(context.Context) (string, error)) (WorkflowRecord, error) {
	if task, exists := builder.tasks[input.ID]; exists {
		return task, checkWorkflowAmendedTask(task, input, slice)
	}

	return s.newWorkflowAmendmentTask(ctx, builder, slice, input, allocate)
}

func checkWorkflowAmendedTask(task WorkflowRecord, input GeneratedWorkflowTask, slice GeneratedWorkflowSlice) error {
	var previous GeneratedWorkflowTask
	if err := json.Unmarshal(task.Domain["execution_contract"], &previous); err != nil {
		return err
	}

	data, err := json.Marshal(input)
	old, oldErr := json.Marshal(previous)

	if err != nil || oldErr != nil || task.Task.SliceID != slice.ID || task.Task.Order != input.Order || string(data) != string(old) {
		return fmt.Errorf("%w: amendment reordered, reassociated or rewrote retained task %s", ErrWorkflowConflict, task.Task.Key)
	}

	return nil
}

func (s *WorkflowStore) newWorkflowAmendmentTask(ctx context.Context, builder *workflowAmendmentBuild, slice GeneratedWorkflowSlice, input GeneratedWorkflowTask, allocate func(context.Context) (string, error)) (WorkflowRecord, error) {
	if err := checkWorkflowAmendmentAddition(*builder, slice, input); err != nil {
		return WorkflowRecord{}, err
	}

	key, err := allocate(ctx)
	if err != nil {
		return WorkflowRecord{}, err
	}

	task, err := s.materializationTask(input, slice, builder.result.Set, key, builder.number, time.Now().UTC())
	if err != nil {
		return task, err
	}

	builder.number++
	builder.keys[input.ID] = key

	return task, nil
}
