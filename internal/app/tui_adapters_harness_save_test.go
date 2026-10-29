package app

import (
	"errors"
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	"forge.adrianpk.com/chr/internal/tui"
)

type tuiHarnessLostAck struct {
	inner    tuiHarnessSelectionExecutor
	dropped  bool
	requests []tui.HarnessSelectionRequest
}

func (e *tuiHarnessLostAck) InspectHarnessSelections() (tui.HarnessSelectionView, error) {
	return e.inner.InspectHarnessSelections()
}

func (e *tuiHarnessLostAck) SaveHarnessSelections(request tui.HarnessSelectionRequest) (tui.HarnessSelectionView, error) {
	e.requests = append(e.requests, request)

	view, err := e.inner.SaveHarnessSelections(request)
	if err == nil && !e.dropped {
		e.dropped = true

		return tui.HarnessSelectionView{}, errors.New("lost acknowledgement after committed change")
	}

	return view, err
}

// Keyboard saves reach the real atomic owner. Dropping one successful response
// preserves the exact retry while old work keeps its binding and new work binds
// the changed worker. The same live model clears either role without a restart.
func TestTUIHarnessSelectionSave(t *testing.T) {
	operations, record, request := serviceWorkflowTicketFixture(t)
	marker := prepareWebHarnessProfiles(t, operations)
	before := claimWebHarnessInvocation(t, operations, request)
	executor := &tuiHarnessLostAck{inner: newTUIHarnessSelectionExecutor(operations)}
	options := newTUIOptions(tuiOptionsData{projectContext: operations.projectContext, operations: operations})
	options.HarnessExecutor = executor

	var model tea.Model = tui.NewModel(options)

	model = tuiHarnessStep(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p"), Alt: true})
	model = tuiHarnessStep(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	model = tuiHarnessCycle(t, model, 3)
	model = tuiHarnessStep(t, model, tea.KeyMsg{Type: tea.KeyTab})
	model = tuiHarnessCycle(t, model, 4)
	model = tuiHarnessStep(t, model, tea.KeyMsg{Type: tea.KeyEnter})

	saved := webHarnessSnapshot(t, operations)
	if saved.Worker != "beta" || saved.Generation != "pi-writing" || saved.ChangedBy != "test-operator" {
		t.Fatal("keyboard save did not persist both roles and the local actor", saved)
	}

	model = tuiHarnessRetrySave(t, model, operations, executor, saved)

	webHarnessCheckProcess(t, operations, saved)
	webHarnessCheckRetained(t, operations, before)
	webHarnessCheckNewInvocation(t, operations, record, request)
	tuiHarnessClear(t, model, operations)

	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("changing defaults launched a driver")
	}
}

func tuiHarnessRetrySave(t *testing.T, model tea.Model, operations service, executor *tuiHarnessLostAck, saved harnesscore.SelectionSnapshot) tea.Model {
	t.Helper()

	visible := strings.Join(strings.Fields(model.View()), " ")
	if !strings.Contains(visible, "Worker choice: beta") || !strings.Contains(visible, "Generation choice: pi-writing") || !strings.Contains(visible, "Could not confirm save") {
		t.Fatal("lost response discarded the displayed choices", visible)
	}

	model = tuiHarnessStep(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	if len(executor.requests) != 2 || executor.requests[0].RequestID != executor.requests[1].RequestID || executor.requests[0].Revision != executor.requests[1].Revision || webHarnessSnapshot(t, operations) != saved {
		t.Fatal("retry replaced the exact request or published twice")
	}

	return model
}

func tuiHarnessClear(t *testing.T, model tea.Model, operations service) {
	t.Helper()

	model = tuiHarnessStep(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})

	model = tuiHarnessStep(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	if snapshot := webHarnessSnapshot(t, operations); snapshot.Worker != "beta" || snapshot.Generation != "" {
		t.Fatal("generation clear changed the worker default")
	}

	model = tuiHarnessStep(t, model, tea.KeyMsg{Type: tea.KeyShiftTab})
	model = tuiHarnessStep(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	tuiHarnessStep(t, model, tea.KeyMsg{Type: tea.KeyEnter})

	if snapshot := webHarnessSnapshot(t, operations); snapshot.Worker != "" || snapshot.Generation != "" {
		t.Fatal("worker clear did not restore legacy dispatch")
	}
}

// The application adapter derives actor authority locally and refuses foreign
// project bindings and unprepared profiles without mutating configuration.
func TestTUIHarnessSelectionRejected(t *testing.T) {
	operations := newTestService(t, newTestWorkspace(t))
	prepareWebHarnessProfiles(t, operations)

	executor := newTUIHarnessSelectionExecutor(operations)

	view, err := executor.InspectHarnessSelections()
	if err != nil {
		t.Fatal(err)
	}

	before := webHarnessSnapshot(t, operations)
	for _, test := range []struct{ name, project, profile, actor string }{
		{"foreign project", "another-project", "beta", "test-operator"},
		{"unprepared profile", view.Snapshot.ProjectRef, "unavailable", "test-operator"},
		{"missing actor", view.Snapshot.ProjectRef, "beta", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(workspacecore.ActiveOperatorIDEnv, test.actor)

			_, err := executor.SaveHarnessSelections(tui.HarnessSelectionRequest{ProjectRef: test.project, Revision: view.Snapshot.Revision, RequestID: "rejected-" + strings.ReplaceAll(test.name, " ", "-"), Worker: &test.profile})
			if err == nil || webHarnessSnapshot(t, operations) != before {
				t.Fatal("rejected local selection changed configuration")
			}
		})
	}
}

func tuiHarnessCycle(t *testing.T, model tea.Model, count int) tea.Model {
	t.Helper()

	for range count {
		model = tuiHarnessStep(t, model, tea.KeyMsg{Type: tea.KeyRight})
	}

	return model
}
