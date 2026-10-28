package workflow

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
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowArtifactAllocation struct {
	Number int
	Time   time.Time
}

type workflowMaterializationSetInput struct {
	Plan       generationcore.GeneratedWorkflowPlan
	Body       []byte
	Set        workflowstate.Record
	Allocation workflowArtifactAllocation
}

type workflowMaterializationTasks struct {
	Records []workflowstate.Record
	Keys    map[string]string
}

func (s *Store) buildMaterialization(ctx context.Context, opts MaterializationOptions, body []byte, number int) (MaterializationResult, error) {
	plan := generationcore.OrderedWorkflowPlan(opts.Plan)
	now := time.Now().UTC()

	set, err := s.materializationEnvelope("plan", plan.Title, number, now)
	if err != nil {
		return MaterializationResult{}, err
	}

	return s.buildMaterializationForSet(ctx, opts, workflowMaterializationSetInput{Plan: plan, Body: body, Set: set, Allocation: workflowArtifactAllocation{Number: number + 1, Time: now}})
}

func (s *Store) buildMaterializationForSet(ctx context.Context, opts MaterializationOptions, input workflowMaterializationSetInput) (MaterializationResult, error) {
	plan, body, set := input.Plan, input.Body, input.Set
	set.Metadata["materialization_id"] = json.RawMessage(fmt.Sprintf("%q", plan.MaterializationID))
	set.Body = body
	set.Set = &workflowstate.WorkflowSet{Schema: workflowstate.WorkflowSchema, UseCase: plan.UseCase, UseCaseOrder: plan.UseCaseOrder, Status: "ready", PlanRef: filepath.ToSlash(filepath.Join("plan", set.ID, "artifact.md")), PlanHash: plan.PlanHash, ClosingValidation: plan.ClosingValidation}

	tasks, err := s.buildMaterializationTasks(ctx, opts, set, plan, input.Allocation)
	if err != nil {
		return MaterializationResult{}, err
	}

	for _, slice := range plan.Slices {
		descriptor, err := materializationSlice(slice, set, tasks.Records, tasks.Keys)
		if err != nil {
			return MaterializationResult{}, err
		}

		set.Set.Slices = append(set.Set.Slices, descriptor)
	}

	return MaterializationResult{Set: set, Tasks: tasks.Records}, nil
}

func (s *Store) materializationEnvelope(kind, title string, number int, now time.Time) (workflowstate.Record, error) {
	id := fmt.Sprintf("%05d-%s", number, opscore.ArtifactSlugOrDefault(title))
	opts := opscore.ArtifactOptions{Layout: s.layout, Type: kind, Title: title, Producer: "chr"}

	metadata, err := workflowJSONFields(opscore.NewArtifactMetadata(opts, id, now))
	if err != nil {
		return workflowstate.Record{}, err
	}

	return workflowstate.Record{Identity: workflowstate.Identity{ID: id, ProjectID: s.projectID, Namespace: s.layout.Namespace(), Revision: 1, CreatedAt: now, UpdatedAt: now}, Kind: kind, Path: kind + "/" + id + "/metadata.json", Metadata: metadata}, nil
}

func (s *Store) buildMaterializationTasks(ctx context.Context, opts MaterializationOptions, set workflowstate.Record, plan generationcore.GeneratedWorkflowPlan, allocation workflowArtifactAllocation) (workflowMaterializationTasks, error) {
	keys := map[string]string{}
	tasks := []workflowstate.Record{}

	for _, slice := range plan.Slices {
		for _, input := range slice.Tasks {
			key, err := opts.ReserveTicketKey(ctx)
			if err != nil {
				return workflowMaterializationTasks{}, err
			}

			keys[input.ID] = key

			record, err := s.materializationTask(input, slice, set, key, allocation)
			if err != nil {
				return workflowMaterializationTasks{}, err
			}

			allocation.Number++

			tasks = append(tasks, record)
		}
	}

	for i := range tasks {
		tasks[i].Task.DependsOn = resolveMaterializationRefs(tasks[i].Task.DependsOn, keys)
	}

	return workflowMaterializationTasks{Records: tasks, Keys: keys}, nil
}

func (s *Store) materializationTask(input generationcore.GeneratedWorkflowTask, slice generationcore.GeneratedWorkflowSlice, set workflowstate.Record, key string, allocation workflowArtifactAllocation) (workflowstate.Record, error) {
	now := allocation.Time

	record, err := s.materializationEnvelope("ticket", input.Title, allocation.Number, now)
	if err != nil {
		return record, err
	}

	record.Task = &workflowstate.Task{Schema: workflowstate.WorkflowSchema, Key: key, Status: string(ticketcore.StatusReadyForDev), UseCase: set.Set.UseCase, SetID: set.ID, SliceID: slice.ID, Order: input.Order, Phase: "pending", DependsOn: input.DependsOn}

	contract, err := json.Marshal(input)
	if err != nil {
		return record, err
	}

	record.Task.ContractHash = contenthash.Sum(append([]byte(set.ID+"\x00"+slice.ID+"\x00"), contract...))
	record.Domain = map[string]json.RawMessage{"plan_task_id": json.RawMessage(fmt.Sprintf("%q", input.ID)), "execution_contract": contract}
	body, err := ticketcore.NewRenderer().Render(ticketcore.Ticket{Key: key, Title: input.Title, Synopsis: input.Outcome, Status: ticketcore.StatusReadyForDev, Priority: input.Priority, Severity: input.Severity, Type: input.Type, Tags: input.Tags, UseCase: set.Set.UseCase, ProjectArea: "plan", SourceKind: "plan_tracking", SourceRef: set.Set.PlanRef, Slice: slice.Label, Body: input.Body, CreatedAt: now, UpdatedAt: now})
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

func materializationSlice(input generationcore.GeneratedWorkflowSlice, set workflowstate.Record, tasks []workflowstate.Record, keys map[string]string) (workflowstate.WorkflowSlice, error) {
	// Membership is a digest of ordered ticket keys and contracts, never a second
	// editable task list on the set descriptor.
	contract, err := json.Marshal(struct {
		Contract   string
		Validation workflowstate.Validation
	}{input.Contract, input.Validation})
	if err != nil {
		return workflowstate.WorkflowSlice{}, err
	}

	return workflowstate.WorkflowSlice{ID: input.ID, Order: input.Order, Title: input.Title, ContractRef: filepath.ToSlash(filepath.Join("plan", set.ID, "extension", "chr", "slice-"+input.ID+".json")), ContractHash: contenthash.Sum(contract), TaskInputHash: TaskInputHash(tasks, input.ID), DependsOn: resolveMaterializationRefs(input.DependsOn, keys), Phase: "pending"}, nil
}

// TaskInputHash detects membership or contract changes in a slice.
func TaskInputHash(records []workflowstate.Record, sliceID string) string {
	tasks := []workflowstate.Task{}

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
