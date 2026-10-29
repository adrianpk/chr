package tui

import "strings"

func (m *Model) renderHelp() string {
	return strings.Join([]string{
		"Help: " + string(m.helpReturnView),
		"",
		viewKeyHints[m.helpReturnView],
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
