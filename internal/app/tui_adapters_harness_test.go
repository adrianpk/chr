package app

import (
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/tui"
)

// Real application assembly reads private prepared profiles and a claimed
// invocation through keyboard events, without launching or rebinding work.
func TestTUIHarnessSelectionRead(t *testing.T) {
	operations, _, request := serviceWorkflowTicketFixture(t)
	marker := prepareWebHarnessProfiles(t, operations)
	before := claimWebHarnessInvocation(t, operations, request)
	options := newTUIOptions(tuiOptionsData{projectContext: operations.projectContext, operations: operations})

	var model tea.Model = tui.NewModel(options)

	model = tuiHarnessStep(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	for _, expected := range []string{"Worker default: alpha", "model-a", "Generation default: Legacy dispatch", before.ID, "running", "Retained worker bindings", "remote auth not tested"} {
		if !strings.Contains(strings.Join(strings.Fields(model.View()), " "), expected) {
			t.Fatal("picker omitted shared selection or retained binding", expected, model.View())
		}
	}

	for range 3 {
		model = tuiHarnessStep(t, model, tea.KeyMsg{Type: tea.KeyRight})
	}

	if !strings.Contains(model.View(), "> Worker choice: beta") {
		t.Fatal("keyboard did not preview the named worker choice")
	}

	model = tuiHarnessStep(t, model, tea.KeyMsg{Type: tea.KeyEsc})
	if strings.Contains(model.View(), "Harness defaults") {
		t.Fatal("cancel did not close the picker")
	}

	webHarnessCheckRetained(t, operations, before)

	if webHarnessSnapshot(t, operations).Worker != "alpha" {
		t.Fatal("preview or cancel saved a choice")
	}

	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("inspection launched a driver")
	}
}

func tuiHarnessStep(t *testing.T, model tea.Model, message tea.Msg) tea.Model {
	t.Helper()

	updated, command := model.Update(message)
	for command != nil {
		updated, command = updated.Update(command())
	}

	return updated
}
