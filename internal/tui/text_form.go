package tui

import (
	"crypto/rand"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type textFormChoice struct{ Label, Value, Revision string }

type textFormField struct {
	Name, ReadOnly string
	Choices        []textFormChoice
	Choice         int
	Label          string
	Multiline      bool
	Input          textarea.Model
}

type textFormState struct {
	GuidedRequest      *GuidedRequest
	WorkRequest        WorkActionRequest
	TicketRequest      TicketActionRequest
	Details            []string
	DumpRequest        DumpActionRequest
	SpecRequest        SpecActionRequest
	Active             bool
	Title, SubmitLabel string
	Fields             []textFormField
	Focus              int
	Submit             func(*Model) tea.Cmd
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
		if command, handled := m.enterTextForm(); handled {
			return command
		}
	}

	return m.updateTextFormInput(message)
}

func (m *Model) enterTextForm() (tea.Cmd, bool) {
	if m.textForm.Focus == len(m.textForm.Fields) {
		return m.textForm.Submit(m), true
	}

	if !m.textForm.Fields[m.textForm.Focus].Multiline {
		m.focusTextForm(m.textForm.Focus + 1)

		return nil, true
	}

	return nil, false
}

func (m *Model) updateTextFormInput(message tea.KeyMsg) tea.Cmd {
	if m.textForm.Focus == len(m.textForm.Fields) {
		return nil
	}

	field := &m.textForm.Fields[m.textForm.Focus]
	if field.ReadOnly != "" {
		m.statusMessage = field.ReadOnly
		if field.Multiline {
			return updateReadOnlyInput(field, message)
		}

		return nil
	}

	if len(field.Choices) != 0 {
		return m.updateTextFormChoice(message)
	}

	before := field.Input.Value()
	input, command := field.Input.Update(message)

	field.Input = input
	if input.Value() != before {
		m.rotateTextFormIntent()
	}

	return command
}

func (m *Model) textFormHasInput() bool {
	for _, field := range m.textForm.Fields {
		if field.ReadOnly == "" && field.Input.Value() != "" {
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
	values := []string{m.textForm.Title, m.textForm.DumpRequest.Path, m.textForm.DumpRequest.Revision, m.textForm.DumpRequest.OperationID, m.textForm.SpecRequest.Path, m.textForm.SpecRequest.Revision, m.textForm.SpecRequest.OperationID, m.textForm.SpecRequest.Scope.CaseSlug, m.textForm.TicketRequest.Key, m.textForm.TicketRequest.Revision, m.textForm.TicketRequest.OperationID, m.textForm.TicketRequest.CaseSlug}
	for _, field := range m.textForm.Fields {
		values = append(values, field.value())
	}

	return strings.Join(values, "\x00")
}

func (m *Model) textFormLines(width, height int) []string {
	lines := []string{"Tab / Shift+Tab choose a field; Ctrl+S submits.", ""}
	lines = append(lines, m.textForm.Details...)
	focusStart, focusEnd := 0, 0

	for index, field := range m.textForm.Fields {
		start := len(lines)

		lines = append(lines, m.textFormFieldLines(field, index, width, height)...)
		if index == m.textForm.Focus {
			focusStart, focusEnd = start, len(lines)
		}
	}

	prefix := "  "
	if m.textForm.Focus == len(m.textForm.Fields) {
		prefix = "> "
		focusStart, focusEnd = len(lines)+1, len(lines)+2
	}

	lines = append(lines, "", prefix+"["+m.textForm.SubmitLabel+"]", "Esc cancels; Enter on the button submits.")
	if len(lines) <= height {
		return lines
	}

	start := max(0, focusEnd-height+2)
	start = min(start, focusStart)
	end := min(len(lines), start+max(1, height-1))

	return append([]string{"Tab / Shift+Tab scroll to each field; Ctrl+S submits."}, lines[start:end]...)
}

func (m *Model) textFormFieldLines(field textFormField, index, width, height int) []string {
	if field.ReadOnly != "" && (index != m.textForm.Focus || !field.Multiline) {
		prefix := "  "
		if index == m.textForm.Focus {
			prefix = "> "
		}

		return []string{prefix + field.Label + ": " + fitPlainLine(strings.ReplaceAll(field.value(), "\n", " / "), max(1, width-len(field.Label)-4)), "  Read-only: " + field.ReadOnly}
	}

	if len(field.Choices) != 0 {
		return []string{m.textFormChoiceLine(field, index)}
	}

	if index != m.textForm.Focus {
		return []string{field.Label + ": " + fitPlainLine(strings.ReplaceAll(field.Input.Value(), "\n", " / "), max(1, width-len(field.Label)-2))}
	}

	input := field.Input
	input.SetWidth(max(1, width-2))

	rows := 1
	if field.Multiline {
		rows = max(1, min(8, height-8))
	}

	input.SetHeight(rows)

	lines := append([]string{"> " + field.Label}, strings.Split(input.View(), "\n")...)
	if field.ReadOnly != "" {
		lines = append(lines, "Read-only: "+field.ReadOnly)
	}

	return lines
}

func (field textFormField) value() string {
	if len(field.Choices) != 0 {
		return field.Choices[field.Choice].Value
	}

	return field.Input.Value()
}

func (m *Model) rotateTextFormIntent() {
	if native := m.textForm.WorkRequest.Workflow; native != nil {
		copy := *native
		copy.RequestKey = rand.Text()
		m.textForm.WorkRequest.Workflow = &copy
	}

	if m.textForm.TicketRequest.Action != "" {
		m.textForm.TicketRequest.OperationID = rand.Text()
	}

	if m.textForm.DumpRequest.Action != "" {
		m.textForm.DumpRequest.OperationID = rand.Text()
	}

	if m.textForm.SpecRequest.Action != "" {
		m.textForm.SpecRequest.OperationID = rand.Text()
	}
}

func (m *Model) updateTextFormChoice(message tea.KeyMsg) tea.Cmd {
	field := &m.textForm.Fields[m.textForm.Focus]
	delta := 0

	switch message.String() {
	case "right", "down", " ":
		delta = 1
	case "left", "up":
		delta = -1
	}

	if delta != 0 {
		field.Choice = (field.Choice + delta + len(field.Choices)) % len(field.Choices)

		m.rotateTextFormIntent()
	}

	return nil
}

func (m *Model) textFormChoiceLine(field textFormField, index int) string {
	prefix := "  "
	if index == m.textForm.Focus {
		prefix = "> "
	}

	line := prefix + field.Label + ": " + field.Choices[field.Choice].Label
	if len(field.Choices) > 1 {
		line += " (left/right choose)"
	}

	return line
}

func updateReadOnlyInput(field *textFormField, message tea.KeyMsg) tea.Cmd {
	switch message.Type {
	case tea.KeyUp, tea.KeyDown, tea.KeyLeft, tea.KeyRight, tea.KeyHome, tea.KeyEnd, tea.KeyPgUp, tea.KeyPgDown:
		input, command := field.Input.Update(message)
		field.Input = input

		return command
	default:
		return nil
	}
}
