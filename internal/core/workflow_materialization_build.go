package core

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) buildMaterialization(ctx context.Context, opts WorkflowMaterializationOptions, body []byte, number int) (WorkflowMaterializationResult, error) {
	plan := generationcore.OrderedWorkflowPlan(opts.Plan)
	now := time.Now().UTC()

	set, err := s.materializationEnvelope("plan", plan.Title, number, now)
	if err != nil {
		return WorkflowMaterializationResult{}, err
	}

	return s.buildMaterializationForSet(ctx, opts, plan, body, number+1, now, set)
}

func (s *WorkflowStore) buildMaterializationForSet(ctx context.Context, opts WorkflowMaterializationOptions, plan generationcore.GeneratedWorkflowPlan, body []byte, number int, now time.Time, set WorkflowRecord) (WorkflowMaterializationResult, error) {
	set.Metadata["materialization_id"] = json.RawMessage(fmt.Sprintf("%q", plan.MaterializationID))
	set.Body = body
	set.Set = &workflowstate.WorkflowSet{Schema: WorkflowSchema, UseCase: plan.UseCase, UseCaseOrder: plan.UseCaseOrder, Status: "ready", PlanRef: filepath.ToSlash(filepath.Join("plan", set.ID, "artifact.md")), PlanHash: plan.PlanHash, ClosingValidation: plan.ClosingValidation}

	tasks, keys, err := s.buildMaterializationTasks(ctx, opts, set, plan, number, now)
	if err != nil {
		return WorkflowMaterializationResult{}, err
	}

	for _, slice := range plan.Slices {
		descriptor, err := materializationSlice(slice, set, tasks, keys)
		if err != nil {
			return WorkflowMaterializationResult{}, err
		}

		set.Set.Slices = append(set.Set.Slices, descriptor)
	}

	return WorkflowMaterializationResult{Set: set, Tasks: tasks}, nil
}

func (s *WorkflowStore) materializationEnvelope(kind, title string, number int, now time.Time) (WorkflowRecord, error) {
	id := fmt.Sprintf("%05d-%s", number, opscore.ArtifactSlugOrDefault(title))
	opts := opscore.ArtifactOptions{Layout: s.layout, Type: kind, Title: title, Producer: "chr"}

	metadata, err := workflowJSONFields(opscore.NewArtifactMetadata(opts, id, now))
	if err != nil {
		return WorkflowRecord{}, err
	}

	return WorkflowRecord{WorkflowIdentity: WorkflowIdentity{ID: id, ProjectID: s.projectID, Namespace: s.layout.Namespace(), Revision: 1, CreatedAt: now, UpdatedAt: now}, Kind: kind, Path: kind + "/" + id + "/metadata.json", Metadata: metadata}, nil
}

func (s *WorkflowStore) buildMaterializationTasks(ctx context.Context, opts WorkflowMaterializationOptions, set WorkflowRecord, plan generationcore.GeneratedWorkflowPlan, number int, now time.Time) ([]WorkflowRecord, map[string]string, error) {
	keys := map[string]string{}
	tasks := []WorkflowRecord{}

	for _, slice := range plan.Slices {
		for _, input := range slice.Tasks {
			key, err := opts.ReserveTicketKey(ctx)
			if err != nil {
				return nil, nil, err
			}

			keys[input.ID] = key

			record, err := s.materializationTask(input, slice, set, key, number, now)
			if err != nil {
				return nil, nil, err
			}

			number++

			tasks = append(tasks, record)
		}
	}

	for i := range tasks {
		tasks[i].Task.DependsOn = resolveMaterializationRefs(tasks[i].Task.DependsOn, keys)
	}

	return tasks, keys, nil
}

func (s *WorkflowStore) materializationTask(input generationcore.GeneratedWorkflowTask, slice generationcore.GeneratedWorkflowSlice, set WorkflowRecord, key string, number int, now time.Time) (WorkflowRecord, error) {
	record, err := s.materializationEnvelope("ticket", input.Title, number, now)
	if err != nil {
		return record, err
	}

	record.Task = &WorkflowTask{Schema: WorkflowSchema, Key: key, Status: TicketStatusReadyForDev, UseCase: set.Set.UseCase, SetID: set.ID, SliceID: slice.ID, Order: input.Order, Phase: "pending", DependsOn: input.DependsOn}

	contract, err := json.Marshal(input)
	if err != nil {
		return record, err
	}

	record.Task.ContractHash = contenthash.Sum(append([]byte(set.ID+"\x00"+slice.ID+"\x00"), contract...))
	record.Domain = map[string]json.RawMessage{"plan_task_id": json.RawMessage(fmt.Sprintf("%q", input.ID)), "execution_contract": contract}
	body, err := NewTicketRenderer().RenderTicket(Ticket{Key: key, Title: input.Title, Synopsis: input.Outcome, Status: TicketStatusReadyForDev, Priority: input.Priority, Severity: input.Severity, Type: input.Type, Tags: input.Tags, UseCase: set.Set.UseCase, ProjectArea: "plan", SourceKind: "plan_tracking", SourceRef: set.Set.PlanRef, Slice: slice.Label, Body: input.Body, CreatedAt: now, UpdatedAt: now})
	record.Body = []byte(body)

	return record, err
}

func resolveMaterializationRefs(refs []workflowstate.Ref, keys map[string]string) []workflowstate.Ref {
	resolved := append([]workflowstate.Ref(nil), refs...)
	for i, ref := range resolved {
		if ref.Kind == "ticket" && keys[ref.ID] != "" {
			resolved[i].ID = keys[ref.ID]
		}
	}

	return resolved
}

func materializationSlice(input generationcore.GeneratedWorkflowSlice, set WorkflowRecord, tasks []WorkflowRecord, keys map[string]string) (workflowstate.WorkflowSlice, error) {
	// Membership is a digest of ordered ticket keys and contracts, never a second
	// editable task list on the set descriptor.
	contract, err := json.Marshal(struct {
		Contract   string
		Validation workflowstate.Validation
	}{input.Contract, input.Validation})
	if err != nil {
		return workflowstate.WorkflowSlice{}, err
	}

	return workflowstate.WorkflowSlice{ID: input.ID, Order: input.Order, Title: input.Title, ContractRef: filepath.ToSlash(filepath.Join("plan", set.ID, "extension", "chr", "slice-"+input.ID+".json")), ContractHash: contenthash.Sum(contract), TaskInputHash: WorkflowTaskInputHash(tasks, input.ID), DependsOn: resolveMaterializationRefs(input.DependsOn, keys), Phase: "pending"}, nil
}

// WorkflowTaskInputHash detects membership or contract changes in a slice.
func WorkflowTaskInputHash(records []WorkflowRecord, sliceID string) string {
	tasks := []WorkflowTask{}

	for _, record := range records {
		if record.Task != nil && record.Task.SliceID == sliceID {
			tasks = append(tasks, *record.Task)
		}
	}

	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Order < tasks[j].Order })

	values := []string{}
	for _, task := range tasks {
		values = append(values, fmt.Sprintf("%d:%q:%q", task.Order, task.Key, task.ContractHash))
	}

	return contenthash.Sum([]byte(strings.Join(values, "\n")))
}
