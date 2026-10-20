package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"forge.adrianpk.com/chr/internal/core"
)

// A submitted choice is frozen until acknowledgement. Rejection keeps the
// original revision/request and visible selection without an automatic rebase.
func TestTUIHarnessSelectionSave(t *testing.T) {
	for _, failure := range []struct {
		name string
		err  error
	}{{"stale", core.ErrWorkflowConflict}, {"lost acknowledgement", errors.New("private transport details")}} {
		t.Run(failure.name, func(t *testing.T) {
			executor := harnessSaveFixture()
			executor.saveErr = failure.err
			model := NewModel(Options{HarnessExecutor: executor})
			model.harnessSelection = harnessSelectionState{Active: true, View: executor.view, Worker: "beta", Generation: "", Session: 1}
			updated, command := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

			model = requireUpdatedModel(t, updated)
			if command == nil {
				t.Fatal("save did not dispatch")
			}

			model = harnessSaveInFlightKeys(t, model)

			updated, _ = model.Update(command())
			model = requireUpdatedModel(t, updated)

			harnessAssertSavedRequest(t, executor)

			if model.harnessSelection.Saving || model.harnessSelection.Worker != "beta" || model.harnessSelection.Generation != "" || model.harnessSelection.View.RequestID != executor.request.RequestID || strings.Contains(model.View(), "private transport") {
				t.Fatal("rejection lost choices or exposed a private failure")
			}
		})
	}
}

func harnessSaveInFlightKeys(t *testing.T, model Model) Model {
	t.Helper()

	for _, key := range []tea.KeyType{tea.KeyEnter, tea.KeyEsc, tea.KeyRight} {
		updated, duplicate := model.Update(tea.KeyMsg{Type: key})

		model = requireUpdatedModel(t, updated)
		if duplicate != nil || !model.harnessSelection.Saving || model.harnessSelection.Worker != "beta" {
			t.Fatal("in-flight request was edited or dispatched twice")
		}
	}

	return model
}

func harnessAssertSavedRequest(t *testing.T, executor *harnessSelectionFixture) {
	t.Helper()

	r := executor.request
	if r.RequestID != "exact-request" || r.Revision != executor.view.Snapshot.Revision || r.ProjectRef != "project-1" || r.Worker == nil || *r.Worker != "beta" || r.Generation == nil || *r.Generation != "" {
		t.Fatal("save replaced the displayed request binding", r)
	}
}

// Both selected roles and their controls remain visible in the existing compact
// shell. Longer profile names wrap, and page keys expose the retained bindings.
func TestTUIHarnessSelectionLayout(t *testing.T) {
	for _, size := range []struct {
		name          string
		width, height int
	}{{"wide", 120, 40}, {"compact", 60, 24}} {
		t.Run(size.name, func(t *testing.T) {
			executor := harnessSaveFixture()
			executor.view.Profiles[1].Model = strings.Repeat("long-model", 12)
			executor.view.Retained = []HarnessRetainedView{{ID: "original-worker", Status: "running", Binding: &core.HarnessBinding{Profile: "alpha", Harness: "codex", Model: "model-a"}}}
			model := NewModel(Options{HarnessExecutor: executor})
			model.width, model.height = size.width, size.height
			model.harnessSelection = harnessSelectionState{Active: true, View: executor.view, Worker: "beta", Generation: "", Message: "Choices retained; reload before saving."}

			view := model.View()
			harnessAssertFrame(t, view, size.width)

			for range 40 {
				updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyPgDown})
				model = requireUpdatedModel(t, updated)
			}

			if !strings.Contains(model.View(), "original-worker") {
				t.Fatal("page keys could not expose retained bindings")
			}
		})
	}
}

func harnessAssertFrame(t *testing.T, view string, width int) {
	t.Helper()

	for _, line := range strings.Split(view, "\n") {
		if lipgloss.Width(line) > width {
			t.Fatal("picker overflowed terminal width")
		}
	}

	for _, expected := range []string{"Harness defaults", "Enter: save", "Worker choice: beta", "Generation choice: Restore legacy"} {
		if !strings.Contains(view, expected) {
			t.Fatal("compact picker hid a choice or control", expected, view)
		}
	}
}

func harnessSaveFixture() *harnessSelectionFixture {
	ready := core.HarnessReadiness{Supported: true, Executable: "available", Runtime: "native", Credentials: "cli_managed", RemoteAuthentication: "not_tested"}

	return &harnessSelectionFixture{view: HarnessSelectionView{
		RequestID: "exact-request", Snapshot: core.HarnessSelectionSnapshot{HarnessSelectionAudit: core.HarnessSelectionAudit{ProjectRef: "project-1"}, Revision: strings.Repeat("a", 64)},
		Worker: core.HarnessSelection{Profile: "alpha", Harness: "codex", Model: "model-a", Readiness: ready}, Generation: core.HarnessSelection{Readiness: ready},
		Profiles: []HarnessProfileView{{Name: "alpha", Harness: "codex", Model: "model-a", Worker: ready, Generation: ready}, {Name: "beta", Harness: "codex", Model: "model-b", Worker: ready, Generation: ready}},
	}}
}
