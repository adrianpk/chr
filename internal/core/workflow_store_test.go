package core

import (
	"os/exec"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
)

func workflowTestStore(t *testing.T) *workflowcore.Store {
	t.Helper()

	if testing.Short() {
		t.Skip("real Git workflow integration runs in the nightly suite")
	}

	dir := t.TempDir()
	if output, err := exec.Command("git", "init", "-q", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: dir, Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	return workflowcore.NewWorkflowStore(layout, "project-1")
}
