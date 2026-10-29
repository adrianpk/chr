package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type textFormField struct {
	Label     string
	Multiline bool
	Input     textarea.Model
}

type textFormState struct {
	Active             bool
	Title, SubmitLabel string
	Fields             []textFormField
	Focus              int
	Submit             feedbackRecovery
}

func newTextFormField(label string, multiline bool, limit int) textFormField {
	input := textarea.New()
	input.CharLimit, input.ShowLineNumbers, input.Prompt = limit, false, ""
	input.SetHeight(1)
	input.SetWidth(64)

	return textFormField{Label: label, Multiline: multiline, Input: input}
}

func (m *Model) openTextForm(form textFormState) {
	m.draftSequence++
	m.textForm = form
	m.textForm.Active = true
	m.focusTextForm(0)

	m.statusMessage = ""
}

func (m *Model) focusTextForm(index int) {
	form := &m.textForm

	form.Focus = (index + len(form.Fields) + 1) % (len(form.Fields) + 1)
	for i := range form.Fields {
		form.Fields[i].Input.Blur()

		if i == form.Focus {
			form.Fields[i].Input.Focus()
		}
	}
}

func (m *Model) updateTextFormKey(message tea.KeyMsg) tea.Cmd {
	switch message.Type {
	case tea.KeyEsc:
		if m.textFormHasInput() {
			return m.confirmDraftChange("Discard the unsubmitted form?", discardTextForm)
		}

		return discardTextForm(m)
	case tea.KeyTab:
		m.focusTextForm(m.textForm.Focus + 1)

		return nil
	case tea.KeyShiftTab:
		m.focusTextForm(m.textForm.Focus - 1)

		return nil
	case tea.KeyCtrlS:
		return m.textForm.Submit(m)
	case tea.KeyEnter:
		if m.textForm.Focus == len(m.textForm.Fields) {
			return m.textForm.Submit(m)
		}

		if !m.textForm.Fields[m.textForm.Focus].Multiline {
			m.focusTextForm(m.textForm.Focus + 1)

			return nil
		}
	}

	if m.textForm.Focus == len(m.textForm.Fields) {
		return nil
	}

	field := &m.textForm.Fields[m.textForm.Focus]
	input, command := field.Input.Update(message)
	field.Input = input

	return command
}

func (m *Model) textFormHasInput() bool {
	for _, field := range m.textForm.Fields {
		if field.Input.Value() != "" {
			return true
		}
	}

	return false
}

func discardTextForm(m *Model) tea.Cmd {
	m.textForm = textFormState{}
	m.feedback.Open = false
	m.statusMessage = "Form discarded."

	return nil
}

func (m *Model) textFormBinding() string {
	values := []string{m.textForm.Title}
	for _, field := range m.textForm.Fields {
		values = append(values, field.Input.Value())
	}

	return strings.Join(values, "\x00")
}

func (m *Model) textFormLines(width, height int) []string {
	lines := []string{"Tab / Shift+Tab choose a field; Ctrl+S submits.", ""}

	for index, field := range m.textForm.Fields {
		if index != m.textForm.Focus {
			lines = append(lines, field.Label+": "+fitPlainLine(strings.ReplaceAll(field.Input.Value(), "\n", " / "), max(1, width-len(field.Label)-2)))

			continue
		}

		input := field.Input
		input.SetWidth(max(1, width-2))

		rows := 1
		if field.Multiline {
			rows = max(1, min(8, height-len(m.textForm.Fields)-8))
		}

		input.SetHeight(rows)

		lines = append(lines, "> "+field.Label)
		lines = append(lines, strings.Split(input.View(), "\n")...)
	}

	prefix := "  "
	if m.textForm.Focus == len(m.textForm.Fields) {
		prefix = "> "
	}

	return append(lines, "", prefix+"["+m.textForm.SubmitLabel+"]", "Esc cancels; Enter on the button submits.")
}
