package app

import (
	"bytes"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

// TestServiceWorkCompat retains unadopted legacy selection and execution while
// exposing its authority. Adopted records remain unchanged and cannot enter it.
func TestServiceWorkCompat(t *testing.T) {
	operations, native, _, _ := serviceWorkNativeFixture(t)
	project := operations.projectContext.ProjectDir

	installFakeTicketStartHarnessExecutor(t, nil)
	writeRuntimeTemplate(t, project)
	writeCLIWorkspaceConfig(t, project, "work:\n  state_authority: tracker_compat\n  mode: no_git_enforcement\n")
	writeTicketMarkdownOnly(t, project, "PRJ-99-Legacy.md", testLoopTicket("PRJ-99", "Legacy task", "ready_for_dev", "high", "normal", "feature", "", "frontend", nil))

	checkServiceWorkCompatPreview(t, operations)

	result, err := operations.StartAgentWorkTicket("PRJ-99")
	if err != nil || result.Execution == nil || result.Execution.TicketKey != "PRJ-99" || result.Authority != core.WorkStateAuthorityTrackerCompat {
		t.Fatal("legacy work lost its execution route", err)
	}

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	retained, err := store.Read(t.Context(), native.Path)
	if err != nil || retained.Revision != native.Revision || retained.Task.Phase != "pending" {
		t.Fatal("legacy execution mutated adopted work", err)
	}
}

// TestServiceWorkTicketPreview reports exact scope without allocating a run or
// invoking either the native worker or legacy runtime setup.
func TestServiceWorkTicketPreview(t *testing.T) {
	operations, native, _, provider := serviceWorkNativeFixture(t)

	config, err := core.LoadWorkspaceConfig(operations.projectContext.ProjectDir)
	if err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	if err := printNativeTicketStartPreview(operations, config, native.Task.Key, &output); err != nil {
		t.Fatal(err)
	}

	for _, text := range []string{"dry_run: true", "authority: tickets", "scope: ticket"} {
		if !strings.Contains(output.String(), text) {
			t.Fatal("exact preview hid scope", output.String())
		}
	}

	if provider.creates != 0 {
		t.Fatal("preview created a PR")
	}

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	retained, err := store.Read(t.Context(), native.Path)
	if err != nil || retained.Revision != native.Revision || retained.Task.Phase != "pending" {
		t.Fatal("preview changed task authority", err)
	}
}

func checkServiceWorkCompatPreview(t *testing.T, operations service) {
	t.Helper()

	preview, err := operations.PreviewAllReadyAgentWork()
	if err != nil || preview.Authority != core.WorkStateAuthorityTrackerCompat || preview.EligibleCount != 1 || preview.SelectedQueue[0].Key != "PRJ-99" {
		t.Fatal("compatibility expanded its legacy scope", preview, err)
	}

	var output bytes.Buffer
	if err := printAgentWorkRunAllPreview(&output, operations.projectContext, AgentWorkSelectionOptions{}, preview); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output.String(), "authority: tracker_compat") || !strings.Contains(output.String(), "scope: unadopted_legacy") {
		t.Fatal("compatibility scope is hidden", output.String())
	}
}
