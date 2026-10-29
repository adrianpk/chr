package tui

import "strings"

func (m *Model) renderHelp() string {
	return strings.Join([]string{
		"Help: " + string(m.helpReturnView),
		"",
		m.helpViewKeys(),
		"",
		"F1 Cases  F2 Dumps  F3 Quizzes  F4 Specs  F5 Plans  F6 Board  F7 Dash",
		"F10 Exit  Ctrl+C exits during busy operations",
		"? Help  alt+p Project  : Command palette  tab Next pane  alt+m Maximize",
		"",
		"Project: edit metadata and context, select Harness defaults with h.",
		"Palette: Go to Runtime for runtime readiness and workspace index health.",
		"Palette: Go to Documents, Authoring or Agent for additional workspace views.",
		"",
		"Esc returns to " + string(m.helpReturnView) + ".",
	}, "\n")
}

func (m *Model) renderHelpWorkspace(width, height int) string {
	return m.renderDetailPanel("Help", strings.Split(m.renderHelp(), "\n"), width, height, true)
}

func (m *Model) helpViewKeys() string {
	if m.helpReturnView == ViewDash && m.dashWorkflowDetails {
		return "Workflow browser: up/down select; tab detail; [/] workflows; n/m members; r refresh.\nc continue; p pause; x cancel; u resume; a approve; R reconcile.\ng allowance; d decision; b scope preview; t ticket; C case; w/Esc queue."
	}

	return viewKeyHints[m.helpReturnView]
}
