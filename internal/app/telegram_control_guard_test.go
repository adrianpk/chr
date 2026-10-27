package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestTelegramControlDecisionGuards rejects each independently stale displayed
// image and a last-boundary authority loss without changing native history.
func TestTelegramControlDecisionGuards(t *testing.T) {
	telegramControlParallel(t)

	operations, id := serviceIntegrationDecisionReady(t)

	before, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}

	original := workflowIntegrationAnswerRequest(before, "resolve")

	cases := []struct {
		name   string
		change func(*WorkflowControlRequest)
	}{
		{"invocation", func(r *WorkflowControlRequest) { r.ExpectedRevision++ }},
		{"entry", func(r *WorkflowControlRequest) { r.Integration.EntryID += "-new" }},
		{"entry revision", func(r *WorkflowControlRequest) { r.Integration.EntryRevision++ }},
		{"budget", func(r *WorkflowControlRequest) { r.Integration.BudgetRevision++ }},
		{"candidate", func(r *WorkflowControlRequest) { r.Integration.CandidateHash = strings.Repeat("0", 64) }},
		{"review", func(r *WorkflowControlRequest) { r.BindingHash = strings.Repeat("0", 64) }},
		{"question revision", func(r *WorkflowControlRequest) { r.DecisionRevision++ }},
		{"question hash", func(r *WorkflowControlRequest) { r.DecisionHash = strings.Repeat("0", 64) }},
		{"authority", func(r *WorkflowControlRequest) { r.DecisionAnswer.Authority = "different-authority" }},
		{"choice", func(r *WorkflowControlRequest) { r.DecisionAnswer.Choice = "undisplayed-choice" }},
		{"actor", func(r *WorkflowControlRequest) { r.DecisionAnswer.Actor = "spoofed-actor" }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := telegramControlClone(t, original, "guard-"+strings.ReplaceAll(test.name, " ", "-"))
			test.change(&request)
			telegramControlRejected(t, operations, request, before.Revision)
		})
	}

	request := telegramControlClone(t, original, "guard-final-authority")
	checks := 0
	request.AuthorizationCheck = func(context.Context) error {
		checks++
		if checks == 2 {
			return workflowstate.ErrConflict
		}

		return nil
	}
	telegramControlRejected(t, operations, request, before.Revision)

	if checks != 2 {
		t.Fatal("authority was not checked immediately before the native commit", checks)
	}

	request.RequestBinding = nil
	telegramControlRejected(t, operations, request, before.Revision)
}

func telegramControlClone(t *testing.T, source WorkflowControlRequest, id string) WorkflowControlRequest {
	t.Helper()

	data, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}

	var request WorkflowControlRequest
	if err := json.Unmarshal(data, &request); err != nil {
		t.Fatal(err)
	}

	request.RequestBinding = &workflowstate.MutationBinding{RequestID: id, Operator: "operator"}
	request.DecisionAnswer.ID = id

	return request
}

func telegramControlRejected(t *testing.T, operations service, request WorkflowControlRequest, revision uint64) {
	t.Helper()

	after, err := operations.ControlWorkExtent(t.Context(), request)
	if !errors.Is(err, workflowstate.ErrConflict) || after.Revision != revision || after.ControlReceipt != nil {
		t.Fatal("stale or unauthorized decision changed native state", after, err)
	}
}

// TestTelegramControlStaleReview keeps candidate and target movement from
// authorizing an older preview, and displays the currently retained native state.
func TestTelegramControlStaleReview(t *testing.T) {
	telegramControlParallel(t)

	f, operations, id := telegramControlDecisionFixture(t, "approve")

	provider, ok := operations.workflowPullRequests.(*serviceNativeReviewFixture)
	if !ok {
		t.Fatal("native review fixture unavailable")
	}

	original := provider.pr

	before, err := operations.ReadWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	for _, change := range []string{"candidate", "base"} {
		t.Run(change, func(t *testing.T) {
			token := telegramControlOffer(t, f, "/approve "+id)

			if change == "candidate" {
				provider.pr.HeadCommit = strings.Repeat("a", 40)
			} else {
				provider.pr.BaseCommit = strings.Repeat("b", 40)
			}

			text := f.send(t, "/answer "+token+" approve", 0).Text
			if !strings.Contains(text, "unavailable or stale") || !strings.Contains(text, id) {
				t.Fatal("stale review hid current native state", text)
			}

			after, err := operations.ReadWorkflow(t.Context(), id)
			if err != nil || after.Revision != before.Revision || after.Review.State != "waiting_approval" {
				t.Fatal("old preview approved changed input", after, err)
			}

			provider.pr = original
		})
	}
}
