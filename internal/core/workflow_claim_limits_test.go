package core

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

// TestWorkflowClaimExpiry keeps serial capacity after its owner loses result authority.
func TestWorkflowClaimExpiry(t *testing.T) {
	store, _, request := workflowClaimFixture(t)
	request.Policy.WorkerTimeoutSeconds = 1
	run := workflowTestStart(t, store, request)
	claim := workflowTestClaim(t, store, run.Invocation.ID)

	deadline := time.NewTimer(time.Second)
	defer deadline.Stop()

	<-deadline.C

	if err := store.CheckpointWorkflow(t.Context(), claim); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatalf("expired owner applied checkpoint: %v", err)
	}

	if _, err := AcquireAgentOperationLock(t.Context(), AgentOperationLockOptions{ProjectDir: store.layout.ProjectDir(), Operation: AgentOperationDevelopment}); !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatalf("expired lease freed a live reservation: %v", err)
	}
}

// TestWorkflowClaimBudget retains attempt consumption through repeated attachments.
func TestWorkflowClaimBudget(t *testing.T) {
	store, _, request := workflowClaimFixture(t)
	request.Policy.MaxWorkerAttempts = 1
	run := workflowTestStart(t, store, request)

	claim := workflowTestClaim(t, store, run.Invocation.ID)
	if err := store.CheckpointWorkflow(t.Context(), claim); err != nil {
		t.Fatal(err)
	}

	checkWorkflowUnownedStatusWrite(t, store, run.Invocation.Members[0].Admission.Path)

	attached := workflowTestStart(t, store, request)
	if !attached.Attached || len(attached.Invocation.Attempts) != 1 {
		t.Fatal("resume reset consumed attempt budget")
	}

	if _, err := store.ClaimWorkflow(t.Context(), run.Invocation.ID); err == nil || !strings.Contains(err.Error(), "attempt budget exhausted") {
		t.Fatalf("exhausted budget admitted a worker: %v", err)
	}

	legacy, err := AcquireAgentOperationLock(t.Context(), AgentOperationLockOptions{ProjectDir: store.layout.ProjectDir(), Operation: AgentOperationDevelopment})
	if err != nil {
		t.Fatalf("rejected claim leaked serial capacity: %v", err)
	}

	if err := legacy.Release(); err != nil {
		t.Fatal(err)
	}
}

func checkWorkflowUnownedStatusWrite(t *testing.T, store *WorkflowStore, path string) {
	t.Helper()

	record, err := store.Read(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}

	record.Task.AttemptID, record.Task.Phase, record.Task.Status = "rogue", "running", TicketStatusInProgress
	if _, err := store.Write(t.Context(), record, record.Revision); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatalf("generic writer created an unfenced claim: %v", err)
	}
}

// TestWorkflowOperationRelease rejects a stale lock image before deleting its replacement.
func TestWorkflowOperationRelease(t *testing.T) {
	store := workflowTestStore(t)

	lock, err := AcquireAgentOperationLock(t.Context(), AgentOperationLockOptions{ProjectDir: store.layout.ProjectDir(), Operation: AgentOperationDevelopment})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if lock.reservation != nil {
			if err := lock.reservation.Close(); err != nil {
				t.Error(err)
			}
		}
	})

	replacement := []byte("{\"owner\":\"replacement\"}\n")
	if err := os.WriteFile(lock.path, replacement, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := lock.Release(); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatalf("stale release accepted: %v", err)
	}

	data, err := os.ReadFile(lock.path)
	if err != nil || string(data) != string(replacement) {
		t.Fatal("stale release changed the replacement image")
	}
}

// TestWorkflowStartBounds rejects invalid policy and incomplete explicit execution input.
func TestWorkflowStartBounds(t *testing.T) {
	cases := []struct {
		name   string
		change func(*WorkflowStartRequest)
	}{
		{"empty key", func(r *WorkflowStartRequest) { r.RequestKey = "" }},
		{"control key", func(r *WorkflowStartRequest) { r.RequestKey = "bad\nkey" }},
		{"missing identity", func(r *WorkflowStartRequest) { r.Selection.Selector.Identity = "" }},
		{"missing target", func(r *WorkflowStartRequest) { r.Selection.Target.Repository = "" }},
		{"mutable target", func(r *WorkflowStartRequest) { r.Selection.Target.Commit = "dev" }},
		{"worker deadline", func(r *WorkflowStartRequest) { r.Policy.WorkerTimeoutSeconds = 86401 }},
		{"attempt budget", func(r *WorkflowStartRequest) { r.Policy.MaxWorkerAttempts = 129 }},
		{"negative token limit", func(r *WorkflowStartRequest) { r.Policy.AgentTokenBudget = -1 }},
		{"negative merger limit", func(r *WorkflowStartRequest) { r.Policy.MaxMergerAttempts = -1 }},
		{"merger limit", func(r *WorkflowStartRequest) { r.Policy.MaxMergerAttempts = 11 }},
		{"negative drift limit", func(r *WorkflowStartRequest) { r.Policy.MaxDriftRevisions = -1 }},
		{"drift limit", func(r *WorkflowStartRequest) { r.Policy.MaxDriftRevisions = 11 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, _, request := workflowClaimFixture(t)
			tc.change(&request)

			if _, err := store.StartWorkflow(t.Context(), request); err == nil {
				t.Fatal("invalid execution request admitted")
			}
		})
	}
}

// TestWorkflowInvocationTerminal preserves an attached caller's retry identity after cancellation.
func TestWorkflowInvocationTerminal(t *testing.T) {
	store, hierarchy, request := workflowClaimFixture(t)
	first := workflowTestStart(t, store, request)
	request.RequestKey = "request-alias"

	attached := workflowTestStart(t, store, request)
	if !attached.Attached || attached.Invocation.ID != first.Invocation.ID {
		t.Fatal("equivalent caller did not attach")
	}

	markWorkflowInvocationCancelled(t, store, first.Invocation.ID)

	retry := workflowTestStart(t, store, request)
	if !retry.Attached || retry.Invocation.Status != "cancelled" || len(retry.Invocation.Attempts) != 0 {
		t.Fatal("terminal alias retry restarted work")
	}

	request.Selection.Selector.Identity = hierarchy.Tasks[1].Task.Key
	if _, err := store.StartWorkflow(t.Context(), request); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatalf("attached retry key broadened original input: %v", err)
	}
}

func markWorkflowInvocationCancelled(t *testing.T, store *WorkflowStore, id string) {
	t.Helper()

	if err := store.withLock(t.Context(), func() error {
		run, err := store.readWorkflowInvocation(id)
		if err != nil {
			return err
		}

		run.Status = "cancelled"
		run.Revision++

		change, err := workflowInvocationChange(run)
		if err != nil {
			return err
		}

		return store.commitFiles([]workflowFileChange{change})
	}); err != nil {
		t.Fatal(err)
	}
}
