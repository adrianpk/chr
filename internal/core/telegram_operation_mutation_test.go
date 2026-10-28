package core

import (
	"errors"
	"testing"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func telegramOperationMutationRequest() workflowcore.MutationRequest {
	return workflowcore.MutationRequest{Binding: workflowstate.MutationBinding{RequestID: "opaque-request-1", Operator: "maintainer"}, Title: "Retained task", Record: workflowTestTask("CHR-1")}
}

// TestTelegramOperationOwnerArguments refuses altered arguments, operators and observed revisions.
func TestTelegramOperationOwnerArguments(t *testing.T) {
	changes := []struct {
		name   string
		change func(*workflowcore.MutationRequest)
	}{
		{"body", func(r *workflowcore.MutationRequest) { r.Record.Body = []byte("changed") }},
		{"title", func(r *workflowcore.MutationRequest) { r.Title = "Changed" }},
		{"operator", func(r *workflowcore.MutationRequest) { r.Binding.Operator = "another" }},
		{"revision", func(r *workflowcore.MutationRequest) { r.ExpectedRevision = 2 }},
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
