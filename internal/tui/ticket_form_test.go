package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// A many-field editor keeps each focused value and its read-only explanation
// visible at a small terminal height rather than clipping later metadata.
func TestTicketFormViewport(t *testing.T) {
	model := NewModel(Options{})

	editor := TicketEditorView{Key: "CHR-1", Phase: "pending", Slice: "Slice 1"}
	for index := range 16 {
		editor.Fields = append(editor.Fields, TicketEditorField{Name: fmt.Sprint(index), Label: fmt.Sprintf("Field %d", index), Value: "saved", ReadOnly: "The execution owner controls this field."})
	}

	model.openTicketForm(editor)

	for index := range 16 {
		model.focusTextForm(index)

		view := strings.Join(model.textFormLines(60, 10), "\n")
		if !strings.Contains(view, fmt.Sprintf("> Field %d:", index)) || !strings.Contains(view, "Read-only:") {
			t.Fatal(index, view)
		}

		model.updateTextFormKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("changed")})

		if model.textForm.Fields[index].value() != "saved" {
			t.Fatal("read-only field changed")
		}
	}

	model.focusTextForm(16)

	if !strings.Contains(strings.Join(model.textFormLines(60, 10), "\n"), "> [Save]") {
		t.Fatal("submit was clipped")
	}
}
