package tui

import tea "github.com/charmbracelet/bubbletea"

type keyMode struct {
	Active  bool
	Handle  func(*Model, tea.KeyMsg) tea.Cmd
	Message string
}

var viewKeyHandlers = map[View]func(*Model, tea.KeyMsg) tea.Cmd{
	ViewProject: (*Model).updateProjectKey,
	ViewCases:   (*Model).updateCasesKey,
	ViewDumps:   (*Model).updateDumpsKey,
	ViewQuizzes: (*Model).updateQuizzesKey,
	ViewSpecs:   (*Model).updateSpecsKey,
	ViewPlans:   (*Model).updatePlansKey,
}

func (m *Model) updateModeKey(msg tea.KeyMsg, modes []keyMode) updateOutcome {
	for _, mode := range modes {
		if !mode.Active {
			continue
		}

		if mode.Message != "" {
			m.statusMessage = mode.Message
		}

		var command tea.Cmd
		if mode.Handle != nil {
			command = mode.Handle(m, msg)
		}

		return updateOutcome{Command: command, Handled: true}
	}

	return updateOutcome{}
}
