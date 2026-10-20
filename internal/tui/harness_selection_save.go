package tui

import (
	"errors"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/core"
)

// HarnessSelectionRequest carries exact choices; the application resolves the
// trusted local actor. Nil leaves a role unchanged and empty clears it.
type HarnessSelectionRequest struct {
	ProjectRef, Revision, RequestID string
	Worker, Generation              *string
}

type harnessSelectionSavedMsg struct {
	Session uint64
	View    HarnessSelectionView
	Err     error
}

func (m *Model) saveHarnessPicker() tea.Cmd {
	state := &m.harnessSelection
	if state.View.Snapshot.Revision == "" || state.View.RequestID == "" {
		state.Message = "Reload current defaults before saving."

		return nil
	}

	worker, generation := harnessChangeChoice(state.Worker), harnessChangeChoice(state.Generation)
	if worker == nil && generation == nil {
		state.Message = "Choose a change for at least one role."

		return nil
	}

	request := HarnessSelectionRequest{ProjectRef: state.View.Snapshot.ProjectRef, Revision: state.View.Snapshot.Revision, RequestID: state.View.RequestID, Worker: worker, Generation: generation}
	state.Busy, state.Saving = true, true
	state.Message = "Saving defaults…"
	session, executor := state.Session, m.harnessExecutor

	return func() tea.Msg {
		view, err := executor.SaveHarnessSelections(request)

		return harnessSelectionSavedMsg{Session: session, View: view, Err: err}
	}
}

func (m *Model) completeHarnessSave(msg harnessSelectionSavedMsg) {
	state := &m.harnessSelection
	if !state.Active || !state.Saving || state.Session != msg.Session {
		return
	}

	state.Busy, state.Saving, state.Scroll = false, false, 0
	if msg.Err != nil {
		state.Message = "Could not confirm save. Choices retained; check readiness, retry unchanged or reload."
		if errors.Is(msg.Err, core.ErrWorkflowConflict) {
			state.Message = "Defaults or request identity changed. Choices retained; r reloads before saving."
		}

		return
	}

	state.View = msg.View
	state.Worker, state.Generation = harnessKeepProfile, harnessKeepProfile
	state.Message = "Defaults saved for new operations. Retained work keeps its binding."
}

func harnessChangeChoice(value string) *string {
	if value == harnessKeepProfile {
		return nil
	}

	return &value
}
