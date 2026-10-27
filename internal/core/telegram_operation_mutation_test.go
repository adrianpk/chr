package core

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func telegramOperationMutationRequest() WorkflowMutationRequest {
	return WorkflowMutationRequest{Binding: workflowstate.MutationBinding{RequestID: "opaque-request-1", Operator: "maintainer"}, Title: "Retained task", Record: workflowTestTask("CHR-1")}
}

// TestTelegramOperationOwnerReplay retains the original result after reopening and later edits.
func TestTelegramOperationOwnerReplay(t *testing.T) {
	owner := workflowTestStore(t)
	request := telegramOperationMutationRequest()

	first, err := owner.ApplyRecordMutation(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	actual, err := owner.Read(t.Context(), first.Record.Path)
	if err != nil || actual.SourceHash != first.Record.SourceHash {
		t.Fatal("receipt does not identify exact written image", actual, first, err)
	}

	request.Binding.RequestID = "display-update"
	request.Record = actual
	request.Record.Metadata["title"] = json.RawMessage(`"Updated title"`)

	request.ExpectedRevision = actual.Revision
	if _, err := owner.ApplyRecordMutation(t.Context(), request); err != nil {
		t.Fatal(err)
	}

	reopened := NewWorkflowStore(owner.layout, owner.projectID)

	again, err := reopened.ApplyRecordMutation(t.Context(), telegramOperationMutationRequest())
	if err != nil || !reflect.DeepEqual(again, first) {
		t.Fatal("replay did not return original result", again, first, err)
	}

	checkTelegramOperationOriginalReceipt(t, reopened, first)
}

// TestTelegramOperationOwnerCrash recovers a write whose receipt acknowledgement was lost.
func TestTelegramOperationOwnerCrash(t *testing.T) {
	for index, name := range []string{"metadata", "body", "coordinator", "receipt"} {
		t.Run(name, func(t *testing.T) {
			owner := workflowTestStore(t)
			owner.afterReplace = func(current int) error {
				if current == index {
					return errors.New("injected lost acknowledgement")
				}

				return nil
			}
			request := telegramOperationMutationRequest()

			partial, err := owner.ApplyRecordMutation(t.Context(), request)
			if err == nil {
				t.Fatal("failure injection did not fire", index)
			}

			reopened := NewWorkflowStore(owner.layout, owner.projectID)

			result, err := reopened.ApplyRecordMutation(t.Context(), request)
			if err != nil || result.Record.ID != partial.Record.ID || result.Record.SourceHash != partial.Record.SourceHash {
				t.Fatal("recovery changed reserved identity", result, partial, err)
			}

			checkTelegramOperationOriginalReceipt(t, reopened, result)
		})
	}
}

// TestTelegramOperationOwnerConflict preserves a divergent image and its unfinished journal.
func TestTelegramOperationOwnerConflict(t *testing.T) {
	owner := workflowTestStore(t)
	owner.afterReplace = func(int) error { return errors.New("injected crash") }

	partial, err := owner.ApplyRecordMutation(t.Context(), telegramOperationMutationRequest())
	if err == nil {
		t.Fatal("failure injection did not fire")
	}

	path := filepath.Join(owner.layout.Root(), partial.Record.Path)
	if err := os.WriteFile(path, []byte("divergent image"), 0o644); err != nil {
		t.Fatal(err)
	}

	reopened := NewWorkflowStore(owner.layout, owner.projectID)
	if _, err := reopened.ApplyRecordMutation(t.Context(), telegramOperationMutationRequest()); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("divergent journal accepted", err)
	}

	data, err := os.ReadFile(path)
	if err != nil || string(data) != "divergent image" {
		t.Fatal("divergent image overwritten", err)
	}

	if _, err := os.Stat(filepath.Join(owner.layout.Root(), workflowstate.JournalPath)); err != nil {
		t.Fatal("unfinished journal lost", err)
	}
}

// TestTelegramOperationOwnerArguments refuses altered arguments, operators and observed revisions.
func TestTelegramOperationOwnerArguments(t *testing.T) {
	changes := []struct {
		name   string
		change func(*WorkflowMutationRequest)
	}{
		{"body", func(r *WorkflowMutationRequest) { r.Record.Body = []byte("changed") }},
		{"title", func(r *WorkflowMutationRequest) { r.Title = "Changed" }},
		{"operator", func(r *WorkflowMutationRequest) { r.Binding.Operator = "another" }},
		{"revision", func(r *WorkflowMutationRequest) { r.ExpectedRevision = 2 }},
	}
	for _, test := range changes {
		t.Run(test.name, func(t *testing.T) {
			owner := workflowTestStore(t)

			request := telegramOperationMutationRequest()
			if _, err := owner.ApplyRecordMutation(t.Context(), request); err != nil {
				t.Fatal(err)
			}

			test.change(&request)

			if _, err := owner.ApplyRecordMutation(t.Context(), request); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("changed payload reused identity", err)
			}
		})
	}
}

// TestTelegramOperationOwnerConcurrent serializes a shared request across independent owners.
func TestTelegramOperationOwnerConcurrent(t *testing.T) {
	owner := workflowTestStore(t)
	second := NewWorkflowStore(owner.layout, owner.projectID)
	results := make(chan error, 2)

	for _, store := range []*WorkflowStore{owner, second} {
		go func() {
			_, err := store.ApplyRecordMutation(t.Context(), telegramOperationMutationRequest())
			results <- err
		}()
	}

	firstError, secondError := <-results, <-results
	if firstError != nil || secondError != nil {
		t.Fatal("concurrent replay failed", firstError, secondError)
	}

	snapshot, err := owner.Snapshot(t.Context())
	if err != nil || len(snapshot.Records) != 1 {
		t.Fatal("concurrent request duplicated record", snapshot, err)
	}
}

func checkTelegramOperationOriginalReceipt(t *testing.T, owner *WorkflowStore, first WorkflowMutationResult) {
	t.Helper()

	snapshot, err := owner.Snapshot(t.Context())
	if err != nil || len(snapshot.Records) != 1 {
		t.Fatal("replay duplicated artifact", snapshot, err)
	}

	receipt, err := os.ReadFile(filepath.Join(owner.layout.Root(), first.ReceiptRef))
	if err != nil || strings.Contains(string(receipt), "telegram") || strings.Contains(string(receipt), "chat_id") {
		t.Fatal("domain receipt acquired routing metadata", err)
	}
}

// TestTelegramOperationOwnerBounds rejects invalid bindings, oversized input and unsupported evidence writes before allocation.
func TestTelegramOperationOwnerBounds(t *testing.T) {
	tests := []struct {
		name   string
		change func(*WorkflowMutationRequest)
	}{
		{"empty identity", func(r *WorkflowMutationRequest) { r.Binding.RequestID = "" }},
		{"large identity", func(r *WorkflowMutationRequest) { r.Binding.RequestID = strings.Repeat("x", 129) }},
		{"missing operator", func(r *WorkflowMutationRequest) { r.Binding.Operator = "" }},
		{"control operator", func(r *WorkflowMutationRequest) { r.Binding.Operator = "one\ntwo" }},
		{"large arguments", func(r *WorkflowMutationRequest) { r.Record.Body = []byte(strings.Repeat("x", 64*1024)) }},
		{"missing title", func(r *WorkflowMutationRequest) { r.Title = "" }},
		{"receipt mutation", func(r *WorkflowMutationRequest) { r.Record.Kind = "receipt" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			owner := workflowTestStore(t)
			request := telegramOperationMutationRequest()
			test.change(&request)

			if _, err := owner.ApplyRecordMutation(t.Context(), request); err == nil {
				t.Fatal("invalid mutation accepted")
			}

			if _, err := os.Stat(filepath.Join(owner.layout.Root(), "ticket")); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("invalid mutation allocated artifact", err)
			}
		})
	}
}

// TestTelegramOperationOwnerTamper refuses a changed receipt instead of reconstructing a success from current files.
func TestTelegramOperationOwnerTamper(t *testing.T) {
	owner := workflowTestStore(t)
	request := telegramOperationMutationRequest()

	result, err := owner.ApplyRecordMutation(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(owner.layout.Root(), result.ReceiptRef)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	changed := strings.Replace(string(data), `"state":"confirmed"`, `"state":"unknown"`, 1)
	if err := os.WriteFile(path, []byte(changed), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := owner.ApplyRecordMutation(t.Context(), request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("changed receipt trusted", err)
	}
}

// TestTelegramOperationOwnerFailedArguments retains identity even when validation rejects the first effect.
func TestTelegramOperationOwnerFailedArguments(t *testing.T) {
	owner := workflowTestStore(t)
	request := telegramOperationMutationRequest()

	request.Title = ""
	if _, err := owner.ApplyRecordMutation(t.Context(), request); err == nil {
		t.Fatal("missing title accepted")
	}

	request.Title = "A different request"
	if _, err := owner.ApplyRecordMutation(t.Context(), request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("failed attempt allowed changed arguments", err)
	}

	snapshot, err := owner.Snapshot(t.Context())
	if err != nil || len(snapshot.Records) != 0 {
		t.Fatal("failed request created a record", snapshot, err)
	}
}

// TestTelegramOperationOwnerScope refuses foreign template scopes instead of retargeting them during allocation.
func TestTelegramOperationOwnerScope(t *testing.T) {
	for _, scope := range []string{"project", "namespace"} {
		t.Run(scope, func(t *testing.T) {
			owner := workflowTestStore(t)

			request := telegramOperationMutationRequest()
			if scope == "project" {
				request.Record.ProjectID = "foreign"
			} else {
				request.Record.Namespace = "foreign"
			}

			if _, err := owner.ApplyRecordMutation(t.Context(), request); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("foreign scope was retargeted", err)
			}

			snapshot, err := owner.Snapshot(t.Context())
			if err != nil || len(snapshot.Records) != 0 {
				t.Fatal("foreign request created a record", snapshot, err)
			}
		})
	}
}
