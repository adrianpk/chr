package workflow

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func telegramWorkAdmissionFixture(t *testing.T) (*Store, StartRequest, MaterializationResult) {
	t.Helper()

	store, hierarchy, request := workflowClaimFixture(t)
	request.Selection.Limit = workflowstate.PageLimit

	preview, err := store.Preview(t.Context(), request.Selection)
	if err != nil {
		t.Fatal(err)
	}

	policy, err := store.InspectWorkflowPolicy(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	request.Selection.SnapshotHash = preview.SnapshotHash
	request.Admission = &workflowstate.AdmissionGuard{SnapshotHash: preview.SnapshotHash, Selected: &workflowstate.SelectionGuard{Key: hierarchy.Tasks[0].Task.Key, Revision: hierarchy.Tasks[0].Revision}, Policy: policy}

	request.Admission.ExecutionPolicy, err = store.InspectWorkflowExecutionPolicy(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	return store, request, hierarchy
}

// TestTelegramWorkAdmissionReplay recovers the original native admission after
// source changes and rejects a replacement image under the same RequestKey.
func TestTelegramWorkAdmissionReplay(t *testing.T) {
	store, request, hierarchy := telegramWorkAdmissionFixture(t)
	first := workflowTestStart(t, store, request)
	path := filepath.Join(store.layout.Root(), hierarchy.Tasks[0].Path)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, append(data, ' '), 0o600); err != nil {
		t.Fatal(err)
	}

	check := func(context.Context) error {
		t.Error("replay repeated effect checks")

		return workflowstate.ErrConflict
	}

	again, err := NewWorkflowStore(store.layout, store.projectID).StartWorkflowGuarded(t.Context(), request, check)
	if err != nil || !again.Attached || again.Invocation.ID != first.Invocation.ID || len(again.Invocation.Attempts) != 0 || again.Invocation.Admission == nil {
		t.Fatal("admission replay lost original binding", again, err)
	}

	changed := request
	guard := *request.Admission
	guard.SnapshotHash = strings.Repeat("f", 64)

	changed.Admission, changed.Selection.SnapshotHash = &guard, guard.SnapshotHash
	if _, err := store.StartWorkflow(t.Context(), changed); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("changed snapshot reused original key", err)
	}
}

// TestTelegramWorkAdmissionGuards rejects stale policy, selected record and
// caller authority before an invocation or worker attempt is allocated.
func TestTelegramWorkAdmissionGuards(t *testing.T) {
	for _, test := range []string{"policy", "selected", "authority", "membership"} {
		t.Run(test, func(t *testing.T) {
			store, request, hierarchy := telegramWorkAdmissionFixture(t)

			check := telegramWorkChangeAdmission(t, store, &request, hierarchy, test)

			if _, err := store.StartWorkflowGuarded(t.Context(), request, check); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("changed admission accepted", err)
			}

			invocationPage, err := store.InspectWorkflowPage(t.Context(), 0, 10)

			if err != nil || len(invocationPage.Invocations) != 0 {
				t.Fatal("rejected admission retained an invocation", invocationPage.Invocations, err)
			}
		})
	}
}

func telegramWorkChangeAdmission(t *testing.T, store *Store, request *StartRequest, hierarchy MaterializationResult, change string) AdmissionCheck {
	t.Helper()

	switch change {
	case "policy":
		request.Admission.Policy.Revision = strings.Repeat("f", 64)
	case "selected":
		request.Admission.Selected.Revision++
	case "authority":
		return func(context.Context) error { return workflowstate.ErrConflict }
	case "membership":
		path := filepath.Join(store.layout.Root(), hierarchy.Tasks[1].Path)

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(path, append(data, ' '), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	return nil
}
