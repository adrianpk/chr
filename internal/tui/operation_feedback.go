package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/app/operation"
)

type feedbackRecovery func(*Model) tea.Cmd

type operationFeedback struct {
	Context                   interactionContext
	Scroll                    int
	Result                    operation.Result
	Open                      bool
	Reload, Retry             feedbackRecovery
	OpenDocument, ViewTickets feedbackRecovery
}

type draftConfirmation struct {
	Active  bool
	Summary string
	Context interactionContext
	Apply   feedbackRecovery
}

func (m *Model) recordOperation(outcome operation.Result) {
	if outcome.Status == "" {
		return
	}

	m.feedback = operationFeedback{Result: outcome, Open: outcome.Status != operation.OK}
}

func (m *Model) renderOperationFeedback() string {
	if m.draftConfirmation.Active {
		return "Confirm draft change\n\n" + m.draftConfirmation.Summary + "\n\nEnter confirms; Esc keeps the complete draft."
	}

	if !m.feedback.Open {
		return ""
	}

	lines := m.operationFeedbackLines()
	start := min(m.feedback.Scroll, len(lines)-1)

	return strings.Join(lines[max(start, 0):], "\n")
}

func (m *Model) operationFeedbackLines() []string {
	result := m.feedback.Result

	lines := []string{"Operation: " + result.Status, result.Summary}
	if result.BlockedReason != "" {
		lines = append(lines, "", result.BlockedReason)
	}

	for _, finding := range result.Validation {
		lines = append(lines, "Finding: "+finding)
	}

	for _, path := range result.ChangedFiles {
		lines = append(lines, "Changed: "+path)
	}

	lines = append(lines, m.generationFeedbackActions()...)

	lines = append(lines, "", "Esc returns to the retained input; alt+o opens this result again.")

	return lines
}

func (m *Model) updateFeedbackKey(message tea.KeyMsg) (tea.Cmd, bool) {
	if m.guided.Active || m.harnessSelection.Active {
		return nil, false
	}

	if m.draftConfirmation.Active {
		return m.updateDraftConfirmation(message), true
	}

	if message.String() == "alt+o" && m.feedback.Result.Status != "" {
		m.feedback.Open = !m.feedback.Open

		return nil, true
	}

	if !m.feedback.Open {
		return nil, false
	}

	if m.scrollOperationFeedback(message.String()) {
		return nil, true
	}

	return m.runFeedbackKey(message.String()), true
}

func (m *Model) runFeedbackKey(key string) tea.Cmd {
	switch key {
	case "esc", "enter":
		m.feedback.Open = false
	case "r":
		return m.runFeedbackRecovery(m.feedback.Result.Reload, m.feedback.Reload)
	case "y":
		return m.runFeedbackRecovery(m.feedback.Result.Retry, m.feedback.Retry)
	case "o":
		return m.runFeedbackRecovery(true, m.feedback.OpenDocument)
	case "t":
		return m.runFeedbackRecovery(true, m.feedback.ViewTickets)
	}

	return nil
}

func (m *Model) runFeedbackRecovery(enabled bool, recover feedbackRecovery) tea.Cmd {
	if !enabled || recover == nil {
		return nil
	}

	if !m.feedbackCurrent() {
		m.statusMessage = "This result belongs to an earlier interaction. No recovery action was applied."

		return nil
	}

	return recover(m)
}

func (m *Model) confirmDraftChange(summary string, apply feedbackRecovery) tea.Cmd {
	m.draftConfirmation = draftConfirmation{Active: true, Summary: summary, Context: m.currentInteraction(), Apply: apply}

	return nil
}

func (m *Model) updateDraftConfirmation(message tea.KeyMsg) tea.Cmd {
	pending := m.draftConfirmation

	switch message.Type {
	case tea.KeyEsc:
		m.draftConfirmation = draftConfirmation{}
	case tea.KeyEnter:
		m.draftConfirmation = draftConfirmation{}
		if pending.Context != m.currentInteraction() {
			m.statusMessage = "The draft changed after confirmation opened. No input was discarded."

			return nil
		}

		return pending.Apply(m)
	}

	return nil
}

func (m *Model) feedbackCurrent() bool { return m.feedback.Context == m.currentInteraction() }

func (m *Model) operationFeedbackKeys() string {
	keys := "Keys: esc return  up/down scroll"

	if !m.feedbackCurrent() {
		return keys
	}

	if m.feedback.Reload != nil && m.feedback.Result.Reload {
		keys += "  r reload"
	}

	if m.feedback.Retry != nil && m.feedback.Result.Retry {
		keys += "  y retry original"
	}

	if m.feedback.OpenDocument != nil {
		keys += "  o open document"
	}

	if m.feedback.ViewTickets != nil {
		keys += "  t view tickets"
	}

	return keys
}

func (m *Model) scrollOperationFeedback(key string) bool {
	switch key {
	case "up":
		m.feedback.Scroll--
	case "down":
		m.feedback.Scroll++
	case "home":
		m.feedback.Scroll = 0
	case "end":
		m.feedback.Scroll = len(m.operationFeedbackLines()) - 4
	default:
		return false
	}

	m.feedback.Scroll = max(0, min(m.feedback.Scroll, len(m.operationFeedbackLines())-1))

	return true
}

func (m *Model) operationStatus() string {
	if m.statusMessage != "" {
		return m.statusMessage
	}

	return m.feedback.Result.Summary
}
