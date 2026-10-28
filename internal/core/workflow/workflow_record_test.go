package workflow

import (
	"testing"
	"time"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowTestTask(id string) workflowstate.Record {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	return workflowstate.Record{
		Identity: workflowstate.Identity{ID: id, ProjectID: "project-1", Namespace: "default", Revision: 1, CreatedAt: now, UpdatedAt: now},
		Kind:     "ticket", Path: "ticket/" + id + "/metadata.json", Body: []byte("# Original body\n"),
		Task: &workflowstate.Task{Schema: workflowstate.WorkflowSchema, Key: id, Status: string(ticketcore.StatusReadyForDev), Phase: "pending", ContractHash: "contract-1"},
	}
}

func workflowTestSet() workflowstate.Record {
	r := workflowTestTask("set-1")
	r.Kind, r.Path, r.Task = "plan", "plan/set-1/metadata.json", nil
	r.Set = &workflowstate.WorkflowSet{Schema: workflowstate.WorkflowSchema, Status: "ready", PlanRef: "artifact.md", PlanHash: "plan-1", ClosingValidation: workflowstate.Validation{Required: true, Commands: [][]string{{"go", "test", "./..."}}, TimeoutSeconds: 60}, Slices: []workflowstate.WorkflowSlice{{ID: "slice-1", Order: 1, Title: "First", Phase: "pending", ContractRef: "artifact.md#first", ContractHash: "slice-contract", TaskInputHash: "inputs-1"}}}

	return r
}

// TestWorkflowRecord covers invalid lifecycle claims and partial associations.
func TestWorkflowRecord(t *testing.T) {
	cases := []struct {
		name   string
		change func(*workflowstate.Record)
	}{
		{"schema", func(r *workflowstate.Record) { r.Task.Schema = "chr-workflow/v2" }},
		{"partial", func(r *workflowstate.Record) { r.Task.SetID = "set-1" }},
		{"negative order", func(r *workflowstate.Record) { r.Task.Order = -1 }},
		{"unproven done", func(r *workflowstate.Record) {
			r.Task.Phase, r.Task.Status = "integrated", string(ticketcore.StatusDone)
		}},
		{"unowned running", func(r *workflowstate.Record) {
			r.Task.Phase, r.Task.Status = "running", string(ticketcore.StatusInProgress)
		}},
		{"unvalidated", func(r *workflowstate.Record) {
			r.Task.Phase, r.Task.Status = "validated", string(ticketcore.StatusInProgress)
		}},
		{"revision", func(r *workflowstate.Record) { r.Revision = 0 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := workflowTestTask("task-1")
			tc.change(&r)

			if err := validateWorkflowRecord(r); err == nil {
				t.Fatal("invalid record accepted")
			}
		})
	}

	if err := validateWorkflowRecord(workflowTestTask("independent")); err != nil {
		t.Fatal(err)
	}
}

// TestWorkflowHierarchy protects scoped ownership and ordering across representations.
func TestWorkflowHierarchy(t *testing.T) {
	cases := []struct {
		name   string
		change func(*workflowstate.Snapshot)
	}{
		{"duplicate key", func(s *workflowstate.Snapshot) { s.Records[2].Task.Key = "task-1" }},
		{"duplicate order", func(s *workflowstate.Snapshot) { s.Records[2].Task.Order = 1 }},
		{"foreign scope", func(s *workflowstate.Snapshot) { s.Records[2].ProjectID = "foreign" }},
		{"foreign owner", func(s *workflowstate.Snapshot) { s.Records[2].Task.SetID = "foreign" }},
		{"unresolved slice", func(s *workflowstate.Snapshot) { s.Records[2].Task.SliceID = "foreign" }},
		{"cycle", func(s *workflowstate.Snapshot) {
			s.Records[1].Task.DependsOn = []workflowstate.Ref{{Kind: "ticket", ID: "task-2"}}
			s.Records[2].Task.DependsOn = []workflowstate.Ref{{Kind: "ticket", ID: "task-1"}}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := workflowstate.Snapshot{Records: []workflowstate.Record{workflowTestSet(), workflowTestTask("task-1"), workflowTestTask("task-2")}}
			for i := 1; i < 3; i++ {
				s.Records[i].Task.SetID, s.Records[i].Task.SliceID, s.Records[i].Task.Order = "set-1", "slice-1", i
			}

			if err := validateWorkflowSnapshot(s); err != nil {
				t.Fatal(err)
			}

			tc.change(&s)

			if err := validateWorkflowSnapshot(s); err == nil {
				t.Fatal("invalid hierarchy accepted")
			}
		})
	}
}
