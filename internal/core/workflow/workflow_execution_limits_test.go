package workflow

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowTaskGateOrder stops an ordered shared-deadline gate at its first failure.
func TestWorkflowTaskGateOrder(t *testing.T) {
	gate := workflowstate.Validation{Required: true, Commands: [][]string{workflowTaskCommand(t, "gate"), workflowTaskCommand(t, "fail"), {"git", "status", "--porcelain"}}, TimeoutSeconds: 30}
	store, _, request := workflowExecutionGateFixture(t, gate)
	run := workflowTestStart(t, store, request)
	claim := workflowTestClaim(t, store, run.Invocation.ID)

	proof, err := ExecuteTask(t.Context(), store, claim, workflowTaskCommand(t, "worker"))
	if err == nil || proof.Successful || len(proof.Validation) != 2 || proof.Validation[0].ExitCode != 0 || proof.Validation[1].ExitCode != 7 {
		t.Fatal("gate commands were reordered or continued after failure")
	}
}

// TestWorkflowTaskWaiver retains the explicit contract reason without inventing checks.
func TestWorkflowTaskWaiver(t *testing.T) {
	gate := workflowstate.Validation{Required: false, WaiverReason: "Host Git inspection is sufficient for the fixture"}
	store, _, request := workflowExecutionGateFixture(t, gate)
	run := workflowTestStart(t, store, request)

	proof := workflowTestExecute(t, store, run.Invocation.ID, "worker")
	if !proof.Successful || len(proof.Validation) != 0 || proof.WaiverReason != gate.WaiverReason {
		t.Fatal("explicit waiver was not bound to task evidence")
	}
}

// TestWorkflowTaskArgvBounds rejects oversized commands before preparing a lane.
func TestWorkflowTaskArgvBounds(t *testing.T) {
	cases := []struct {
		name string
		argv []string
	}{
		{"empty", nil}, {"too many", strings.Fields(strings.Repeat("arg ", 65))}, {"too long", []string{strings.Repeat("x", 4097)}}, {"nul", []string{"git", "bad\x00argument"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
			run := workflowTestStart(t, store, request)

			claim := workflowTestClaim(t, store, run.Invocation.ID)
			if _, err := ExecuteTask(t.Context(), store, claim, tc.argv); err == nil {
				t.Fatal("invalid worker command accepted")
			}

			retained, err := store.ReadWorkflow(t.Context(), run.Invocation.ID)
			if err != nil || retained.Lane != nil || retained.Attempts[0].InputRef != "" {
				t.Fatal("invalid argv changed durable execution input")
			}
		})
	}
}

// TestWorkflowTaskReuseEdits rejects changed evidence, instructions and lane history.
func TestWorkflowTaskReuseEdits(t *testing.T) {
	cases := []struct {
		name   string
		change func(*testing.T, *Store, workflowstate.Invocation, workflowstate.Record)
	}{
		{"result", func(t *testing.T, s *Store, run workflowstate.Invocation, _ workflowstate.Record) {
			workflowProcessWrite(t, filepath.Join(s.layout.Root(), run.Attempts[0].ResultRef), "Changed\n")
		}},
		{"input", func(t *testing.T, s *Store, run workflowstate.Invocation, _ workflowstate.Record) {
			workflowProcessWrite(t, filepath.Join(s.layout.Root(), run.Attempts[0].InputRef), "Changed\n")
		}},
		{"instructions", func(t *testing.T, s *Store, _ workflowstate.Invocation, record workflowstate.Record) {
			record.Body = []byte("Changed instructions\n")
			workflowTestWrite(t, s, record)
		}},
		{"lane", func(t *testing.T, s *Store, run workflowstate.Invocation, _ workflowstate.Record) {
			workflowSelectionGit(t, s, "update-ref", run.Lane.Branch, run.Lane.BaseCommit)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, hierarchy, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
			first := workflowTestStart(t, store, request)
			workflowTestExecute(t, store, first.Invocation.ID, "worker")

			retained, err := store.ReadWorkflow(t.Context(), first.Invocation.ID)
			if err != nil {
				t.Fatal(err)
			}

			record, err := store.Read(t.Context(), hierarchy.Tasks[0].Path)
			if err != nil {
				t.Fatal(err)
			}

			tc.change(t, store, retained, record)

			request.RequestKey = "execution-2"
			request.Selection.Selector = workflowClaimSelector(hierarchy, "slice")

			next := workflowTestStart(t, store, request)
			if _, err := store.ClaimWorkflow(t.Context(), next.Invocation.ID); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatalf("changed retained input accepted: %v", err)
			}
		})
	}
}

// TestWorkflowTaskRecordEdit prevents a worker from overwriting an external ticket image.
func TestWorkflowTaskRecordEdit(t *testing.T) {
	store, hierarchy, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	run := workflowTestStart(t, store, request)
	claim := workflowTestClaim(t, store, run.Invocation.ID)
	path := filepath.Join(store.layout.Root(), filepath.Dir(hierarchy.Tasks[0].Path), "artifact.md")

	worker := []string{"sh", "-c", "printf 'Changed instructions\\n' >> \"$1\"", "worker", path}
	if _, err := ExecuteTask(t.Context(), store, claim, worker); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatalf("external input overwritten: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(data), "Changed instructions") {
		t.Fatal("external task body was overwritten")
	}
}

// TestWorkflowTaskReuseTarget cannot confirm a cursor against an unavailable target.
func TestWorkflowTaskReuseTarget(t *testing.T) {
	store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	first := workflowTestStart(t, store, request)
	workflowTestExecute(t, store, first.Invocation.ID, "worker")

	request.RequestKey = "execution-2"
	request.Selection.Target.Commit = strings.Repeat("a", 40)

	next := workflowTestStart(t, store, request)
	if _, err := store.ClaimWorkflow(t.Context(), next.Invocation.ID); err == nil {
		t.Fatal("unavailable target accepted")
	}

	retained, err := store.ReadWorkflow(t.Context(), next.Invocation.ID)
	if err != nil || retained.Cursor != 0 || retained.Status != workflowstate.InvocationPending {
		t.Fatal("reused evidence advanced an unavailable target")
	}
}
