package tui

import (
	"crypto/rand"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/app/operation"
)

// DumpDraftEditorExecutor prepares a private draft editor, not a source writer.
type DumpDraftEditorExecutor interface {
	OpenDumpDraftEditor(string) (DumpDraftEditor, error)
}

// DumpDraftEditor joins the external process result to its owned draft cleanup.
type DumpDraftEditor struct {
	Outcome operation.Result
	Command *exec.Cmd
	Finish  func(error) (DumpActionResult, error)
}

type dumpDraftEditorFinishedMsg struct {
	Binding operationBinding
	Result  DumpActionResult
	Err     error
}

func (m *Model) openDumpDraftEditor() tea.Cmd {
	executor, ok := m.dumpExecutor.(DumpDraftEditorExecutor)
	if !ok {
		m.statusMessage = "Draft editor is unavailable."

		return nil
	}

	session, err := executor.OpenDumpDraftEditor(m.dumpEdit.Input.Value())
	m.recordOperation(session.Outcome)

	if err != nil {
		m.statusMessage = "External edit failed: " + err.Error()

		return nil
	}

	binding := m.beginOperation("dump:external")
	m.statusMessage = "External editor opened for the unsaved body."

	return tea.ExecProcess(session.Command, func(runErr error) tea.Msg {
		result, finishErr := session.Finish(runErr)

		return dumpDraftEditorFinishedMsg{Binding: binding, Result: result, Err: finishErr}
	})
}

func (m *Model) completeDumpDraftEditor(message dumpDraftEditorFinishedMsg) {
	owned, apply := m.finishOperation(message.Binding)
	if !owned || !apply {
		return
	}

	m.recordOperation(message.Result.Outcome)

	if message.Err != nil {
		m.statusMessage = "External edit failed: " + message.Err.Error()

		return
	}

	if message.Result.Content != m.dumpEdit.Input.Value() {
		m.dumpEdit.Input.SetValue(message.Result.Content)

		m.dumpEdit.OperationID = rand.Text()
	}

	m.statusMessage = "External edits retained; Ctrl+S saves the dump."
}
