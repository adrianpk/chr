package app

import (
	"maps"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/app/operation"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/tui"
)

// Keyboard save uses real canonical writes: a concurrent update rejects the old
// revision, cancellation retains the draft, and only confirmed reload replaces it.
func TestTUIOperationDraft(t *testing.T) {
	project := newTestWorkspace(t)
	writeServiceWorkspaceQueryFixture(t, project)

	operations := newTestService(t, project)
	executor := newTUIDumpActionExecutor(operations)
	path := createDumpThroughTUIAdapter(t, executor)

	observed, err := operations.ReadDump(DumpReadRequest{RelativePath: path})
	if err != nil {
		t.Fatal(err)
	}

	listed, err := executor.ExecuteDumpAction(tui.DumpActionRequest{Action: tui.DumpActionRefresh})
	if err != nil {
		t.Fatal(err)
	}

	model := tui.NewModel(tui.Options{ProjectDir: project, Dumps: listed.Dumps, DumpExecutor: executor})
	model.Update(tea.WindowSizeMsg{Width: 120, Height: 44})
	model.Update(tea.KeyMsg{Type: tea.KeyF2})
	tuiHarnessStep(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	model.Update(tea.KeyMsg{Type: tea.KeyEnd})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" UNSENT_SENTINEL")})

	if _, err := operations.SaveDump(DumpSaveRequest{OperationID: "external-update", ExpectedRevision: observed.Revision, RelativePath: path, Content: observed.Content + "\nEXTERNAL_NEWER_BODY\n"}); err != nil {
		t.Fatal(err)
	}

	before := workflowCLIImage(t, filepath.Join(project, "ops"))

	model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

	if !strings.Contains(model.View(), "Operation: stale") {
		t.Fatal("stale write did not show Action feedback", model.View())
	}

	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !strings.Contains(model.View(), "UNSENT_SENTINEL") {
		t.Fatal("cancelled reload lost the draft", model.View())
	}

	if !maps.Equal(before, workflowCLIImage(t, filepath.Join(project, "ops"))) {
		t.Fatal("failed write or cancelled reload changed canonical documents")
	}

	assertTUIConfirmedReload(t, model)
	assertTUISavedDraft(t, model, operations, path)
}

func assertTUISavedDraft(t *testing.T, model *tui.Model, operations service, path string) {
	t.Helper()

	model.Update(tea.KeyMsg{Type: tea.KeyEnd})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ACCEPTED_SENTINEL")})
	model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o"), Alt: true})

	if !strings.Contains(model.View(), "Operation: ok") || !strings.Contains(model.View(), "Changed: "+path) {
		t.Fatal("successful write omitted its changed source", model.View())
	}

	saved, err := operations.ReadDump(DumpReadRequest{RelativePath: path})
	if err != nil || !strings.Contains(saved.Content, "ACCEPTED_SENTINEL") || strings.Contains(saved.Content, "UNSENT_SENTINEL") {
		t.Fatal("confirmed reload/save used the wrong document image", err)
	}
}

func assertTUIConfirmedReload(t *testing.T, model *tui.Model) {
	t.Helper()
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o"), Alt: true})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})

	if !strings.Contains(model.View(), "discard its unsaved text?") {
		t.Fatal("reload skipped draft-loss confirmation", model.View())
	}

	model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !strings.Contains(model.View(), "EXTERNAL_NEWER_BODY") || strings.Contains(model.View(), "UNSENT_SENTINEL") {
		t.Fatal("confirmed reload did not observe the concurrent writer", model.View())
	}
}

// Metadata writers retain their canonical hash checks. Their adapter outcomes
// classify the observed conflict and reload reads the actual new field values.
func TestTUIOperationMetadata(t *testing.T) {
	operations := newTestService(t, newTestWorkspace(t))

	initial, err := operations.SaveProjectMetadata(authoringcore.ProjectMetadata{Name: "Original", Description: "Original project description", Context: "Original project context"})
	if err != nil {
		t.Fatal(err)
	}

	concurrent := initial.Metadata

	concurrent.Name = "Concurrent name"
	if _, err := operations.SaveProjectMetadata(concurrent); err != nil {
		t.Fatal(err)
	}

	initial.Metadata.Name = "Unsent local name"
	executor := newTUIProjectMetadataExecutor(operations)

	rejected, err := executor.SaveProjectMetadata(initial.Metadata)
	if err == nil || rejected.Outcome.Status != operation.Stale || !rejected.Outcome.Reload {
		t.Fatal("metadata conflict lost its disposition", rejected.Outcome, err)
	}

	reloaded, err := executor.ReadProjectMetadata()
	if err != nil || reloaded.Metadata.Name != "Concurrent name" {
		t.Fatal("metadata reload missed the current source", err)
	}
}

// Case metadata uses the same guarded writer and reloads only its captured
// slug, preserving another case's independent local or canonical fields.
func TestTUIOperationCase(t *testing.T) {
	operations := newTestService(t, newTestWorkspace(t))

	initial, err := operations.SaveCaseMetadata(authoringcore.CaseMetadata{Name: "Original case", Description: "Original description"})
	if err != nil {
		t.Fatal(err)
	}

	concurrent := initial.Case

	concurrent.Name = "Concurrent case"
	if _, err := operations.SaveCaseMetadata(concurrent); err != nil {
		t.Fatal(err)
	}

	initial.Case.Name = "Unsent case sentinel"
	executor := newTUICaseMetadataExecutor(operations)

	rejected, err := executor.SaveCaseMetadata(initial.Case)
	if err == nil || rejected.Outcome.Status != operation.Stale || !rejected.Outcome.Reload {
		t.Fatal("case conflict lost its disposition", rejected.Outcome, err)
	}

	reloaded, err := executor.ReadCaseMetadata(initial.Case.Slug)
	if err != nil || reloaded.Case.Name != "Concurrent case" {
		t.Fatal("case reload missed its captured source", err)
	}
}
