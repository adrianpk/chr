package tui

import (
	"slices"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// The same seven view identities own startup, function keys and footer clicks;
// auxiliary navigation never adds a second tab set or shifts a key target.
func TestShellTabs(t *testing.T) {
	model := NewModel(Options{})

	want := []View{ViewCases, ViewDumps, ViewQuizzes, ViewSpecs, ViewPlans, ViewBoard, ViewDash}
	if model.SelectedView() != ViewCases || !slices.Equal(model.views, want) {
		t.Fatal("startup differs from the seven-view shell", model.views)
	}

	model.width, model.height = 80, 30
	keys := []tea.KeyType{tea.KeyF1, tea.KeyF2, tea.KeyF3, tea.KeyF4, tea.KeyF5, tea.KeyF6, tea.KeyF7}
	buttons := shellFooterButtons()

	for index, view := range want {
		model.Update(tea.KeyMsg{Type: keys[index]})

		if model.SelectedView() != view || buttons[index].View != view {
			t.Fatal("function key or footer points at another view", index, view)
		}

		model.selectView(ViewProject)
		model.Update(tea.MouseMsg{X: index*10 + 5, Y: 29, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
		if model.SelectedView() != view {
			t.Fatal("footer click differs from its visible target", index)
		}
	}
}

// Auxiliary views must leave the canonical tab identities and keys unchanged.
func TestShellAuxiliaryTabs(t *testing.T) {
	model := NewModel(Options{})
	want := slices.Clone(defaultViews)

	for _, view := range []View{ViewProject, ViewHelp, ViewRuntime, ViewDocuments} {
		model.selectView(view)

		if model.SelectedView() != view || !slices.Equal(model.views, want) {
			t.Fatal("auxiliary view changed the shell tab identities", view)
		}
	}

	model.Update(tea.KeyMsg{Type: tea.KeyF1})

	if model.SelectedView() != ViewCases {
		t.Fatal("auxiliary navigation shifted the Cases key")
	}
}

// Help is a real contextual screen, returns to the same view and pane, and
// Project remains reachable without altering the seven normal tabs.
func TestShellHelp(t *testing.T) {
	model := NewModel(Options{})
	model.selectView(ViewSpecs)

	model.selectedPane = 1
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})

	if model.SelectedView() != ViewHelp || !strings.Contains(model.View(), "generate spec") {
		t.Fatal("Help did not expose the active view actions", model.View())
	}

	model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if model.SelectedView() != ViewSpecs || model.selectedPane != 1 {
		t.Fatal("Help lost its return view or pane")
	}

	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p"), Alt: true})

	if model.SelectedView() != ViewProject || len(model.views) != 7 {
		t.Fatal("Project shortcut changed the tab set")
	}
}

// Punctuation remains input while editing Project or Case fields and filters;
// Help must never consume text or close a modal to navigate.
func TestShellText(t *testing.T) {
	cases := []struct {
		name string
		view View
		edit func(*Model)
		read func(*Model) string
	}{
		{"project", ViewProject, func(m *Model) { m.projectEdit = newProjectEditState(0, "") }, func(m *Model) string { return m.projectEdit.Input.Value() }},
		{"case", ViewCases, func(m *Model) { m.caseEdit = newProjectEditState(0, "") }, func(m *Model) string { return m.caseEdit.Input.Value() }},
		{"case filter", ViewCases, func(m *Model) { m.caseFilterEditing = true }, func(m *Model) string { return m.caseFilter }},
		{"palette", ViewCases, func(m *Model) { m.openCommandPalette() }, func(m *Model) string { return m.commandPalette.Filter }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			model := NewModel(Options{})
			model.selectView(tc.view)
			tc.edit(model)
			model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})

			model.width, model.height = 80, 30
			model.Update(tea.KeyMsg{Type: tea.KeyF4})
			model.Update(tea.MouseMsg{X: 35, Y: 29, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})

			if model.SelectedView() != tc.view || tc.read(model) != "?" {
				t.Fatal("Help shortcut consumed editor text")
			}
		})
	}
}
