package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Omitted fields preserve the historical v1 request digest. Loading and replay
// keep checkpoint bytes and policy intact despite new explicit live selections.
func TestHarnessBindingLegacy(t *testing.T) {
	store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	started := workflowTestStart(t, store, request)
	path := filepath.Join(store.layout.Root(), workflowInvocationPath(started.Invocation.ID))

	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	writeWorkspaceConfig(t, store.layout.ProjectDir(), "harnesses:\n  worker: pi-unavailable\n")

	request.ResolveHarness = func(context.Context, WorkspaceConfig) (WorkflowHarnessResolution, error) {
		t.Fatal("v1 retry resolved a new profile")

		return WorkflowHarnessResolution{}, nil
	}

	current, err := store.ReadWorkflow(t.Context(), started.Invocation.ID)
	if err != nil || current.Schema != "chr-invocation/v1" || current.HarnessBinding != nil {
		t.Fatal("legacy read migrated the checkpoint", err)
	}

	replay := workflowTestStart(t, store, request)

	retained, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(retained, original) || !replay.Attached || replay.Invocation.Policy != started.Invocation.Policy {
		t.Fatal("legacy observation or replay changed its bytes", err)
	}

	checkHarnessBindingLegacyHash(t, current)
	workflowTestExecute(t, store, current.ID, "worker")
}

func checkHarnessBindingLegacyHash(t *testing.T, run WorkflowInvocation) {
	t.Helper()
	// Field order and tags are the pre-binding WorkflowStartRequest contract.
	legacy := struct {
		Admission           *WorkflowAdmissionGuard `json:"admission_guard,omitempty"`
		IncludeLaterTickets *bool                   `json:"include_later_tickets_override,omitempty"`
		RequestKey          string
		Selection           WorkflowPreviewRequest
		Policy              WorkflowExecutionPolicy
		ReviewRestriction   ReviewWorkflowMode `json:"review_restriction,omitempty"`
	}{Admission: run.Admission, Selection: run.Selection, Policy: run.Policy, ReviewRestriction: run.ReviewRestriction}

	data, err := json.Marshal(legacy)
	if err != nil || contenthash.Sum(data) != run.RequestHash {
		t.Fatal("v1 request hash changed", err)
	}
}

// Compatibility entrypoints restore the retained binding before hashing a new
// attachment key. Missing resolvers fail closed only for new explicit admissions.
func TestHarnessBindingEntry(t *testing.T) {
	store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	writeWorkspaceConfig(t, store.layout.ProjectDir(), "harnesses:\n  worker: alpha\n")

	if _, err := store.StartWorkflow(t.Context(), request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("explicit profile silently used legacy execution", err)
	}

	request.ResolveHarness = harnessBindingTestResolver(t, workflowTaskCommand(t, "worker"))
	started := workflowTestStart(t, store, request)
	request.RequestKey = "compatibility-entry"

	found, matched, err := store.FindWorkflowEntry(t.Context(), request)
	if err != nil || !matched || found.ID != started.Invocation.ID || !sameHarnessBinding(found.HarnessBinding, started.Invocation.HarnessBinding) {
		t.Fatal("entry lost frozen worker identity", err)
	}
}
