package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowAmendmentBuild struct {
	result MaterializationResult
	old    workflowstate.Record
	tasks  map[string]workflowstate.Record
	keys   map[string]string
	files  []workflowstate.FileChange
	number int
	root   string
}

type workflowAmendmentPreparation struct {
	Result MaterializationResult
	Files  []workflowstate.FileChange
}

// workflowAmendmentHost sequences bounded amendment construction under Store-owned allocation and mutation locks.
type workflowAmendmentHost struct{ store *Store }

func (host workflowAmendmentHost) build(ctx context.Context, set workflowstate.Record, request SetAmendmentRequest, body []byte) (workflowAmendmentPreparation, error) {
	builder, err := host.store.workflowAmendmentBuilder(ctx, set)
	if err != nil {
		return workflowAmendmentPreparation{Result: builder.result}, err
	}

	plan := generationcore.OrderedWorkflowPlan(request.Plan)
	if err := (workflowAmendmentHost{store: host.store}).tasks(ctx, &builder, plan, request.ReserveTicketKey); err != nil {
		return workflowAmendmentPreparation{Result: builder.result}, err
	}

	if err := buildWorkflowAmendmentSlices(&builder, plan); err != nil {
		return workflowAmendmentPreparation{Result: builder.result}, err
	}

	if err := host.store.finishWorkflowAmendment(ctx, &builder, request, body); err != nil {
		return workflowAmendmentPreparation{Result: builder.result}, err
	}

	return workflowAmendmentPreparation{Result: builder.result, Files: builder.files}, nil
}

func (s *Store) workflowAmendmentBuilder(ctx context.Context, set workflowstate.Record) (workflowAmendmentBuild, error) {
	builder := workflowAmendmentBuild{old: set, tasks: map[string]workflowstate.Record{}, keys: map[string]string{}, number: opscore.NextArtifactNumber(s.layout.Root()), root: filepath.ToSlash(filepath.Join("plan", set.ID, "extension", "chr", fmt.Sprintf("amendment-%04d", set.Revision+1)))}

	data, err := json.Marshal(set.Set)
	if err != nil {
		return builder, err
	}

	// Retain the observed set separately; changing its slice cannot mutate the
	// prior image used by the guards and immutable amendment receipt.
	builder.result.Set = set

	builder.result.Set.Set = &workflowstate.WorkflowSet{}
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
			return builder, fmt.Errorf("%w: task has no stable plan identity", workflowstate.ErrConflict)
		}

		builder.tasks[id], builder.keys[id] = task, task.Task.Key
	}

	return builder, nil
}

func (host workflowAmendmentHost) tasks(ctx context.Context, builder *workflowAmendmentBuild, plan generationcore.GeneratedWorkflowPlan, allocate func(context.Context) (string, error)) error {
	seen := map[string]bool{}

	for _, slice := range plan.Slices {
		for _, input := range slice.Tasks {
			task, err := (workflowAmendmentHost{store: host.store}).task(ctx, builder, slice, input, allocate)
			if err != nil {
				return err
			}

			seen[input.ID] = true

			builder.result.Tasks = append(builder.result.Tasks, task)
		}
	}

	if len(seen) != len(builder.keys) {
		return fmt.Errorf("%w: amendment removed an existing task", workflowstate.ErrConflict)
	}

	for i := range builder.result.Tasks {
		task := &builder.result.Tasks[i]
		if task.SourceHash == "" {
			task.Task.DependsOn = resolveMaterializationRefs(task.Task.DependsOn, builder.keys)
		}
	}

	return nil
}

func (host workflowAmendmentHost) task(ctx context.Context, builder *workflowAmendmentBuild, slice generationcore.GeneratedWorkflowSlice, input generationcore.GeneratedWorkflowTask, allocate func(context.Context) (string, error)) (workflowstate.Record, error) {
	if task, exists := builder.tasks[input.ID]; exists {
		return task, checkWorkflowAmendedTask(task, input, slice)
	}

	return (workflowAmendmentHost{store: host.store}).newTask(ctx, builder, slice, input, allocate)
}

func checkWorkflowAmendedTask(task workflowstate.Record, input generationcore.GeneratedWorkflowTask, slice generationcore.GeneratedWorkflowSlice) error {
	var previous generationcore.GeneratedWorkflowTask
	if err := json.Unmarshal(task.Domain["execution_contract"], &previous); err != nil {
		return err
	}

	data, err := json.Marshal(input)
	old, oldErr := json.Marshal(previous)

	if err != nil || oldErr != nil || task.Task.SliceID != slice.ID || task.Task.Order != input.Order || string(data) != string(old) {
		return fmt.Errorf("%w: amendment reordered, reassociated or rewrote retained task %s", workflowstate.ErrConflict, task.Task.Key)
	}

	return nil
}

func (host workflowAmendmentHost) newTask(ctx context.Context, builder *workflowAmendmentBuild, slice generationcore.GeneratedWorkflowSlice, input generationcore.GeneratedWorkflowTask, allocate func(context.Context) (string, error)) (workflowstate.Record, error) {
	if err := checkWorkflowAmendmentAddition(*builder, slice, input); err != nil {
		return workflowstate.Record{}, err
	}

	key, err := allocate(ctx)
	if err != nil {
		return workflowstate.Record{}, err
	}

	task, err := host.store.materializationTask(input, slice, builder.result.Set, key, workflowArtifactAllocation{Number: builder.number, Time: time.Now().UTC()})
	if err != nil {
		return task, err
	}

	builder.number++
	builder.keys[input.ID] = key

	return task, nil
}
