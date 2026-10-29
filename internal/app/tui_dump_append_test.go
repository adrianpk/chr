package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/tui"
)

// The reachable append action writes only its captured case dump, preserving
// the other case and submitting exactly the user's text under its revision.
func TestTUIDumpAppend(t *testing.T) {
	operations := navigationTerminalFixture(t)

	cases, err := authoringcore.NewCaseMetadataService(operations.projectContext.ProjectDir).LoadAll()
	if err != nil {
		t.Fatal(err)
	}

	executor := newTUIDumpActionExecutor(operations)

	listed, err := executor.ExecuteDumpAction(tui.DumpActionRequest{Action: tui.DumpActionRefresh})
	if err != nil {
		t.Fatal(err)
	}

	before := findDumpViewByTitle(t, mustListDumps(t, operations), "Alpha navigation notes")
	model := tui.NewModel(tui.Options{Cases: cases, Dumps: listed.Dumps, DumpExecutor: executor})
	runGuidedTUIKey(t, model, "down", false)
	runGuidedTUIKey(t, model, "f2", false)
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Append distinct user text.")})
	model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

	after := mustListDumps(t, operations)
	beta := findDumpViewByTitle(t, after, "Beta navigation notes")

	alpha := findDumpViewByTitle(t, after, "Alpha navigation notes")
	if strings.Count(beta.Content, "Append distinct user text.") != 1 || alpha.Revision != before.Revision {
		t.Fatal("append was lost, duplicated or redirected", model.View())
	}
}

func mustListDumps(t *testing.T, operations service) []DumpView {
	t.Helper()

	listed, err := operations.ListDumps()
	if err != nil {
		t.Fatal(err)
	}

	return listed.Dumps
}
