package app

import (
	"os/exec"
	"reflect"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestTelegramOperationService proves another adapter can reuse the shared mutation receipt without transport metadata.
func TestTelegramOperationService(t *testing.T) {
	project := newTestWorkspace(t)
	if output, err := exec.Command("git", "init", "-q", project).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}

	operations := newTestService(t, project)
	request := core.WorkflowMutationRequest{Binding: workflowstate.MutationBinding{RequestID: "caller-neutral", Operator: "maintainer"}, Title: "Shared request", Record: core.WorkflowRecord{Kind: "ticket", Body: []byte("# Task\n"), Task: &core.WorkflowTask{Schema: core.WorkflowSchema, Key: "CHR-1", Status: ticketcore.StatusBacklog, Phase: "pending", ContractHash: "contract"}}}

	first, err := operations.ApplyWorkflowRecordMutation(t.Context(), "default", request)
	if err != nil {
		t.Fatal(err)
	}

	reopened := newTestService(t, project)

	again, err := reopened.ApplyWorkflowRecordMutation(t.Context(), "default", request)
	if err != nil || !reflect.DeepEqual(again, first) {
		t.Fatal("shared caller replay changed result", again, first, err)
	}
}
