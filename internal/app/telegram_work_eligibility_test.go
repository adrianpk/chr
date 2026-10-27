package app

import (
	"context"
	"maps"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestTelegramWorkDecisionWait uses a real retained merger question to prove
// automatic observation cannot consume a checkpoint or answer operator authority.
func TestTelegramWorkDecisionWait(t *testing.T) {
	operations, id := serviceIntegrationDecisionReady(t)

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil || status.Integration == nil || status.Integration.Decision == nil || status.Integration.Decision.State != "waiting" {
		t.Fatal("native merger wait unavailable", status, err)
	}

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)

	progress, err := operations.ResumeRetainedWork(t.Context(), id, status.Revision)
	if err != nil || progress.Attempted || progress.Steps != 0 {
		t.Fatal("merger decision was implicitly continued", progress, err)
	}

	if !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("observing an operator decision changed native bytes")
	}
}

// TestTelegramWorkCapacity observes an independently held repository slot.
// Waiting authorization retains its revision and checkpoint budget unchanged.
func TestTelegramWorkCapacity(t *testing.T) {
	f, operations, ref := telegramWorkAdmitted(t)
	telegramWorkHoldCapacity(t, operations, ref)

	before, err := operations.ReadWorkflow(t.Context(), ref.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	result := f.receiver.activateWork(t.Context(), ref)
	if result.Failure != nil || result.Progress.Attempted || result.Progress.Steps != 0 {
		t.Fatal("full capacity accepted a new checkpoint", result.Progress, result.Failure)
	}

	after, err := operations.ReadWorkflow(t.Context(), ref.Invocation.ID)
	if err != nil || after.Revision != before.Revision || after.Continuation != nil || len(after.Attempts) != 0 {
		t.Fatal("capacity observation consumed native history", after, err)
	}
}

func telegramWorkHoldCapacity(t *testing.T, operations service, ref telegramcore.WorkReference) {
	t.Helper()

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	snapshot, err := store.InspectWorkflowRecords(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	preview, err := telegramRetainedWorkPreview(ref.Request)
	if err != nil {
		t.Fatal(err)
	}

	for _, record := range snapshot.Records {
		if record.Task != nil && record.Task.SetID != "" {
			selection := workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "ticket", Identity: record.Task.Key}, Target: preview.Request.Target}

			run, err := operations.StartWorkflow(t.Context(), core.WorkflowStartRequest{StartInput: workflowstate.StartInput{
				RequestKey: "another-adapter",
				Selection:  selection,
			}})
			if err != nil {
				t.Fatal(err)
			}

			telegramWorkManualOwner(t, operations, run.Invocation.ID)

			return
		}
	}

	t.Fatal("fixture has no independent grouped work")
}

func telegramWorkManualOwner(t *testing.T, operations service, id string) {
	t.Helper()

	claim, err := operations.ClaimWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := operations.PrepareWorkflowManualTask(t.Context(), claim, nil); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := operations.CheckpointWorkflow(context.Background(), claim); err != nil {
			t.Error(err)
		}
	})
}
