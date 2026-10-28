package workflow

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

func harnessSelectionTestRequest(id, revision, profile string) workspaceconfig.HarnessSelectionChangeRequest {
	return workspaceconfig.HarnessSelectionChangeRequest{Binding: workflowstate.MutationBinding{RequestID: id, Operator: "operator"}, ProjectRef: "fixture", ExpectedRevision: revision, Worker: harnessSelectionTestChoice(profile), Generation: harnessSelectionTestChoice(profile)}
}

func harnessSelectionTestOptions(request workspaceconfig.HarnessSelectionChangeRequest) workspaceconfig.HarnessSelectionChangeOptions {
	return workspaceconfig.HarnessSelectionChangeOptions{Request: request, Authorize: func(context.Context) error { return nil }, Validate: func(context.Context, workspaceconfig.Config) error { return nil }}
}

func writeHarnessSelectionTestFile(t *testing.T, path, body string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func harnessSelectionTestChoice(value string) *string { return &value }
