package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/tui"
)

// Project/case capture submits the entered title and body through the canonical
// owner; body editing keeps scope/title metadata outside the editable content.
func TestTUIDumpForm(t *testing.T) {
	for _, scope := range []string{"", "beta-navigation"} {
		t.Run(scope, func(t *testing.T) {
			operations := navigationTerminalFixture(t)
			operations.dumpCurator = &countingDumpCurator{}

			cases, err := authoringcore.NewCaseMetadataService(operations.projectContext.ProjectDir).LoadAll()
			if err != nil {
				t.Fatal(err)
			}

			executor := newTUIDumpActionExecutor(operations)
			model := tui.NewModel(tui.Options{ProjectDir: operations.projectContext.ProjectDir, Cases: cases, DumpExecutor: executor})
			model.Update(tea.WindowSizeMsg{Width: 120, Height: 44})
			runGuidedTUIKey(t, model, "f2", false)
			model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})

			key := tea.KeyLeft
			if scope != "" {
				key = tea.KeyRight
			}

			model.Update(tea.KeyMsg{Type: key})
			model.Update(tea.KeyMsg{Type: tea.KeyTab})
			model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Entered dump")})
			model.Update(tea.KeyMsg{Type: tea.KeyTab})
			model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Entered body sentinel.")})
			model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
			checkTUIDumpForm(t, operations, executor, model, scope)
		})
	}
}

func checkTUIDumpForm(t *testing.T, operations service, executor tuiDumpActionExecutor, model *tui.Model, scope string) {
	t.Helper()

	listed, err := operations.ListDumps()
	if err != nil {
		t.Fatal(err)
	}

	saved := findDumpViewByTitle(t, listed.Dumps, "Entered dump")
	if saved.Scope.CaseSlug != scope || !strings.Contains(saved.Content, "Entered body sentinel.") {
		t.Fatal(saved)
	}

	runGuidedTUIKey(t, model, "enter", true)

	if !strings.Contains(model.View(), "Title and scope are retained") {
		t.Fatal("captured document was not selected", model.View())
	}

	edited, err := executor.ExecuteDumpAction(tui.DumpActionRequest{Action: tui.DumpActionEdit, Path: saved.RelativePath})
	if err != nil || strings.Contains(edited.Content, "type:") {
		t.Fatal("metadata remained editable", edited, err)
	}

	model.Update(tea.KeyMsg{Type: tea.KeyEnd})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" EDITED_BODY")})
	model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

	after, err := operations.ReadDump(DumpReadRequest{RelativePath: saved.RelativePath})
	if err != nil || !strings.Contains(after.Content, "EDITED_BODY") {
		t.Fatal(after, err)
	}

	assertDumpFormScope(t, operations, scope)
}

func assertDumpFormScope(t *testing.T, operations service, scope string) {
	t.Helper()

	listed, err := operations.ListDumps()
	if err != nil {
		t.Fatal(err)
	}

	final := findDumpViewByTitle(t, listed.Dumps, "Entered dump")
	if final.Scope.CaseSlug != scope {
		t.Fatal("edit changed document scope", final)
	}
}
