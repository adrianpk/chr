package core

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowClaimFixture(t *testing.T) (*WorkflowStore, WorkflowMaterializationResult, WorkflowStartRequest) {
	t.Helper()

	store, result := workflowSelectionFixture(t)
	commit, _ := workflowSelectionCommit(t, store)
	request := WorkflowStartRequest{StartInput: workflowstate.StartInput{
		RequestKey: "request-1",
		Selection:  workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "ticket", Identity: result.Tasks[0].Task.Key}, Target: workflowstate.PreviewTarget{Repository: "https://forge.example/test/repo", Commit: commit}},
	}}

	return store, result, request
}

func workflowTestStart(t *testing.T, store *WorkflowStore, request WorkflowStartRequest) WorkflowStartResult {
	t.Helper()

	result, err := store.StartWorkflow(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	return result
}

func workflowTestClaim(t *testing.T, store *WorkflowStore, id string) *WorkflowClaim {
	t.Helper()

	claim, err := store.ClaimWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := claim.releaseReservation(); err != nil {
			t.Error(err)
		}
	})

	return claim
}

// TestWorkflowInvocationRetry protects frozen membership and token-free attachments.
func TestWorkflowInvocationRetry(t *testing.T) {
	store, _, request := workflowClaimFixture(t)
	first := workflowTestStart(t, store, request)
	claim := workflowTestClaim(t, store, first.Invocation.ID)

	retry := workflowTestStart(t, store, request)
	checkWorkflowInvocationAttachment(t, first, retry, claim)

	if err := store.CheckpointWorkflow(t.Context(), claim); err != nil {
		t.Fatal(err)
	}

	paused := workflowTestStart(t, store, request)
	if !paused.Attached || paused.Invocation.Status != "paused" || len(paused.Invocation.Attempts) != 1 {
		t.Fatal("checkpoint retry allocated another invocation or lost history")
	}
}

func checkWorkflowInvocationAttachment(t *testing.T, first, retry WorkflowStartResult, claim *WorkflowClaim) {
	t.Helper()

	if first.Attached || !retry.Attached || retry.Invocation.ID != first.Invocation.ID || retry.Invocation.Members[0].Admission.Revision != first.Invocation.Members[0].Admission.Revision {
		t.Fatal("retry did not attach to immutable admission")
	}

	member := retry.Invocation.Members[0]
	if member.ExpectedRevision != member.Admission.Revision+1 || member.ExpectedSourceHash == member.Admission.SourceHash {
		t.Fatal("accepted claim did not advance expected image")
	}

	checkWorkflowAttachmentCredentials(t, retry, claim)
}

func checkWorkflowAttachmentCredentials(t *testing.T, retry WorkflowStartResult, claim *WorkflowClaim) {
	t.Helper()

	data, err := json.Marshal(retry)
	if err != nil || strings.Contains(string(data), claim.Token) || strings.Contains(string(data), "owner_token_hashes") {
		t.Fatal("public attachment exposed a credential")
	}
}

// TestWorkflowInvocationOverlap rejects exact-ticket, slice and frontier bypasses.
func TestWorkflowInvocationOverlap(t *testing.T) {
	cases := []struct{ name, first, next string }{
		{"ticket slice", "ticket", "slice"},
		{"slice ticket", "slice", "ticket"},
		{"set ticket", "slice_set", "ticket"},
		{"use case ticket", "use_case", "ticket"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, hierarchy, request := workflowClaimFixture(t)
			request.Selection.Selector = workflowClaimSelector(hierarchy, tc.first)
			first := workflowTestStart(t, store, request)
			request.RequestKey = "request-2"

			request.Selection.Selector = workflowClaimSelector(hierarchy, tc.next)
			if _, err := store.StartWorkflow(t.Context(), request); !errors.Is(err, ErrWorkflowOwned) {
				t.Fatalf("overlapping request accepted: %v", err)
			}

			retained, err := store.ReadWorkflow(t.Context(), first.Invocation.ID)
			if err != nil || retained.Revision != first.Invocation.Revision || len(retained.Attempts) != 0 {
				t.Fatal("conflict changed accepted invocation")
			}
		})
	}
}

func workflowClaimSelector(hierarchy WorkflowMaterializationResult, extent string) workflowstate.Selector {
	ids := map[string]string{"ticket": hierarchy.Tasks[0].Task.Key, "slice": hierarchy.Set.Set.Slices[0].ID, "slice_set": hierarchy.Set.ID, "use_case": hierarchy.Set.Set.UseCase}

	return workflowstate.Selector{Extent: extent, Identity: ids[extent]}
}

// TestWorkflowClaimFence rejects counterfeit and superseded owners before writes.
func TestWorkflowClaimFence(t *testing.T) {
	store, _, request := workflowClaimFixture(t)
	run := workflowTestStart(t, store, request)
	first := workflowTestClaim(t, store, run.Invocation.ID)
	falseOwner := *first

	falseOwner.Token = "counterfeit"
	if err := store.CheckpointWorkflow(t.Context(), &falseOwner); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatalf("counterfeit checkpoint accepted: %v", err)
	}

	if err := store.CheckpointWorkflow(t.Context(), first); err != nil {
		t.Fatal(err)
	}

	second := workflowTestClaim(t, store, run.Invocation.ID)
	if second.Generation != first.Generation+1 || second.AttemptID == first.AttemptID {
		t.Fatal("replacement reused fencing identity")
	}

	if err := store.CheckpointWorkflow(t.Context(), first); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatalf("old owner checkpoint accepted: %v", err)
	}

	if err := store.CheckWorkflowClaim(t.Context(), second); err != nil {
		t.Fatalf("stale owner affected replacement: %v", err)
	}
}

// TestWorkflowClaimEdits rejects cooperative status writes and incompatible manual images.
func TestWorkflowClaimEdits(t *testing.T) {
	store, hierarchy, request := workflowClaimFixture(t)
	run := workflowTestStart(t, store, request)
	claim := workflowTestClaim(t, store, run.Invocation.ID)

	record, err := store.Read(t.Context(), hierarchy.Tasks[0].Path)
	if err != nil {
		t.Fatal(err)
	}

	record.Task.AttemptID, record.Task.Phase, record.Task.Status = "", "pending", ticketcore.StatusReadyForDev
	if _, err := store.Write(t.Context(), record, record.Revision); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatalf("status edit bypassed owner: %v", err)
	}

	path := filepath.Join(store.layout.Root(), hierarchy.Tasks[0].Path)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, append(data, ' '), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := store.CheckpointWorkflow(t.Context(), claim); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatalf("manual image overwritten by owner: %v", err)
	}
}

// TestWorkflowClaimFrozenInputs checks body, membership, set and request changes.
func TestWorkflowClaimFrozenInputs(t *testing.T) {
	cases := []struct {
		name   string
		change func(*testing.T, *WorkflowStore, WorkflowMaterializationResult)
	}{
		{"description", func(t *testing.T, s *WorkflowStore, r WorkflowMaterializationResult) {
			record := r.Tasks[0]
			record.Body = []byte("Changed instructions\n")
			workflowTestWrite(t, s, record)
		}},
		{"set contract", func(t *testing.T, s *WorkflowStore, r WorkflowMaterializationResult) {
			record := r.Set
			record.Set.Slices[0].ContractHash = "changed"
			workflowTestWrite(t, s, record)
		}},
		{"sibling contract", func(t *testing.T, s *WorkflowStore, r WorkflowMaterializationResult) {
			record := r.Tasks[1]
			record.Task.ContractHash = "changed"
			workflowTestWrite(t, s, record)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, hierarchy, request := workflowClaimFixture(t)
			run := workflowTestStart(t, store, request)
			tc.change(t, store, hierarchy)

			if _, err := store.ClaimWorkflow(t.Context(), run.Invocation.ID); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatalf("changed frozen input accepted: %v", err)
			}
		})
	}
}

// TestWorkflowClaimPrerequisite refuses exact work whose earlier task lacks evidence.
func TestWorkflowClaimPrerequisite(t *testing.T) {
	store, hierarchy, request := workflowClaimFixture(t)
	request.Selection.Selector.Identity = hierarchy.Tasks[1].Task.Key

	run := workflowTestStart(t, store, request)
	if _, err := store.ClaimWorkflow(t.Context(), run.Invocation.ID); err == nil {
		t.Fatal("unmet prerequisite admitted a worker")
	}

	request.Selection.Target.Commit = strings.Repeat("a", 40)
	if _, err := store.StartWorkflow(t.Context(), request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatalf("changed target reused request key: %v", err)
	}
}

// TestWorkflowClaimJournalGuard rejects an external image arriving after owner observation.
func TestWorkflowClaimJournalGuard(t *testing.T) {
	store, _, request := workflowClaimFixture(t)
	started := workflowTestStart(t, store, request)
	claim := workflowTestClaim(t, store, started.Invocation.ID)

	err := store.withLock(t.Context(), func() error {
		run, record, err := store.checkWorkflowClaim(t.Context(), claim)
		if err != nil {
			return err
		}

		path := filepath.Join(store.layout.Root(), record.Path)

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if err := os.WriteFile(path, append(data, ' '), 0o600); err != nil {
			return err
		}

		return store.commitWorkflowClaim(run, record)
	})
	if !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatalf("changed pre-journal image accepted: %v", err)
	}

	retained, err := store.ReadWorkflow(t.Context(), started.Invocation.ID)
	if err != nil || retained.Revision != started.Invocation.Revision+1 || retained.Status != "running" {
		t.Fatal("rejected image changed invocation state")
	}
}
