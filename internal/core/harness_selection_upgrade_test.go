package core

import (
	"context"
	"os"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// A failed receipt rename retains the configuration owner's real journal. The
// workflow upgrade must wait until replay finishes that exact selection change.
func TestHarnessSelectionUpgradeFence(t *testing.T) {
	store, options := workflowUpgradeFixture(t, "")
	project := store.layout.ProjectDir()
	request := harnessSelectionTestRequest("before-upgrade", options.ExpectedRevision, "alpha")
	request.ProjectRef = store.projectID

	receipt := retainHarnessSelectionUpgradeIntent(t, project, request)
	if _, err := store.UpgradeWorkflowConfiguration(t.Context(), options); err == nil {
		t.Fatal("upgrade overwrote a pending selection intent")
	}

	if err := os.Remove(receipt); err != nil {
		t.Fatal(err)
	}

	accepted, err := workspaceconfig.ChangeHarnessSelections(t.Context(), project, harnessSelectionTestOptions(request))
	if err != nil {
		t.Fatal(err)
	}

	options.ExpectedRevision = accepted.Revision

	upgraded, err := store.UpgradeWorkflowConfiguration(t.Context(), options)
	if err != nil || upgraded.Current.Version != 1 {
		t.Fatal("upgrade did not resume after selection recovery", err)
	}

	config, err := workspaceconfig.Load(project)
	if err != nil || config.Harnesses.Worker != "alpha" || config.HarnessAudit.RequestID != request.Binding.RequestID {
		t.Fatal("upgrade lost selections or their audit", err)
	}
}

func retainHarnessSelectionUpgradeIntent(t *testing.T, project string, request workspaceconfig.HarnessSelectionChangeRequest) string {
	t.Helper()

	receipt := workspacecore.UserPath(project, "harness-selection-receipts/"+contenthash.Sum([]byte(request.Binding.RequestID))+".json")
	options := harnessSelectionTestOptions(request)
	// Preparation follows the initial receipt lookup. This obstruction therefore
	// fails receipt retention after the journal and configuration image exist.
	options.Validate = func(context.Context, workspaceconfig.Config) error {
		return os.MkdirAll(receipt, 0o700)
	}

	if _, err := workspaceconfig.ChangeHarnessSelections(t.Context(), project, options); err == nil {
		t.Fatal("receipt rename unexpectedly replaced the fixture directory")
	}

	if _, err := os.Stat(workspacecore.UserPath(project, "tmp/harness-selection-change.json")); err != nil {
		t.Fatal("failed publication did not retain its journal", err)
	}

	return receipt
}
