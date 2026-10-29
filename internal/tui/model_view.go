package tui

import (
	"fmt"
	"strings"
)

// View renders the current terminal screen.
func (m *Model) View() string {
	if m.guided.Active && m.width == 0 {
		return m.renderGuidedWorkspace(80, 40)
	}

	if m.width > 0 && m.height > 0 {
		return m.renderShell()
	}

	var output strings.Builder

	output.WriteString("Project\n")
	fmt.Fprintf(&output, "  Name: %s\n", projectDisplayName(m.projectDir))
	fmt.Fprintf(&output, "  Dir:  %s\n\n", m.projectDir)
	output.WriteString("Navigation\n")

	for index, view := range m.views {
		prefix := "  "
		if index == m.selected {
			prefix = "> "
		}

		output.WriteString(prefix)
		output.WriteString(string(view))
		output.WriteString("\n")
	}

	output.WriteString("\n")
	output.WriteString(m.renderSelectedView())
	output.WriteString(m.renderStatusMessage())

	if feedback := m.renderOperationFeedback(); feedback != "" {
		output.WriteString("\n" + feedback)
	}

	output.WriteString(m.renderArtifactAction())
	output.WriteString(m.renderTicketAction())
	output.WriteString(m.renderWorkAction())
	output.WriteString(m.renderQuizAction())
	output.WriteString(m.renderSpecAction())
	output.WriteString(m.renderPlanAction())

	if m.harnessSelection.Active {
		output.WriteString("\n\nHarness defaults\n")
		output.WriteString(strings.Join(m.harnessSelectionLines(76), "\n"))
	}

	output.WriteString(m.renderCommandPalette())
	output.WriteString("\n\n")
	output.WriteString(m.renderKeys())
	output.WriteString("\n")

	return output.String()
}

func (m *Model) renderSelectedView() string {
	if m.textForm.Active {
		return m.textForm.Title + "\n" + strings.Join(m.textFormLines(76, 30), "\n")
	}

	selectedView := m.SelectedView()

	if view, ok := m.renderProductSelectedView(selectedView); ok {
		return view
	}

	return m.renderLegacySelectedView(selectedView)
}

func (m *Model) renderProductSelectedView(selectedView View) (string, bool) {
	if selectedView == ViewProject {
		description := strings.TrimSpace(m.projectMetadata.Description)
		if description == "" {
			description = "Not set."
		}

		name := strings.TrimSpace(m.projectMetadata.Name)
		if name == "" {
			name = projectDisplayName(m.projectDir)
		}

		return fmt.Sprintf("Project\n  Name: %s\n  Description: %s", name, description), true
	}

	if selectedView == ViewCases {
		return "Cases\n  No use cases found.", true
	}

	if selectedView == ViewDumps {
		return m.renderDumps(), true
	}

	if selectedView == ViewQuizzes {
		return m.renderQuizzes(), true
	}

	if selectedView == ViewSpecs {
		return m.renderSpecs(), true
	}

	if selectedView == ViewPlans {
		return m.renderPlans(), true
	}

	if selectedView == ViewDash {
		return m.renderDash(), true
	}

	return "", false
}

func (m *Model) renderLegacySelectedView(selectedView View) string {
	if selectedView == ViewHelp {
		return m.renderHelp()
	}

	if selectedView == ViewBoard {
		return m.renderTicketBoard()
	}

	if selectedView == ViewTickets {
		return m.renderTicketList()
	}

	if selectedView == ViewDocuments {
		return m.renderDocuments()
	}

	if selectedView == ViewAuthoring {
		return m.renderAuthoring()
	}

	if selectedView == ViewAgent {
		return m.renderAgent()
	}

	if selectedView == ViewRuntime {
		return m.renderRuntime()
	}

	placeholder, ok := viewPlaceholders[selectedView]
	if !ok {
		return "Unknown\n  No view is selected."
	}

	return fmt.Sprintf("%s\n  %s", selectedView, placeholder)
}

func (m *Model) renderKeys() string {
	if m.draftConfirmation.Active {
		return "Keys: enter confirm  esc keep complete draft"
	}

	if m.feedback.Open {
		return m.operationFeedbackKeys()
	}

	if m.feedback.Result.Status != "" {
		return m.renderInteractionKeys() + " | alt+o result"
	}

	return m.renderInteractionKeys()
}

func (m *Model) renderInteractionKeys() string {
	if keys, ok := m.activeInteractionKeys(); ok {
		return keys
	}

	if keys, ok := viewKeyHints[m.SelectedView()]; ok {
		return keys + " | F1-F7 tabs  ? Help  alt+p Project  : palette  Ctrl+C exit"
	}

	return "Keys: F1-F7 tabs  ? Help  alt+p Project  : palette  Ctrl+C exit"
}

func (m *Model) activeInteractionKeys() (string, bool) {
	hints := []interactionKeyHint{
		{Active: m.textForm.Active, Text: "Keys: tab next field  shift+tab previous  ctrl+s submit  esc cancel"},
		{Active: m.projectEdit.Active || m.caseEdit.Active, Text: "Keys: enter accepts field  esc cancels  Ctrl+C exits"},
		{Active: m.caseFilterEditing, Text: "Keys: enter applies case filter  esc cancels  Ctrl+C exits"},
		{Active: m.dumpEdit.Active, Text: "Keys: ctrl+s saves dump  esc cancels  alt+e external  Ctrl+C exits"},
		{Active: m.workBusy, Text: "Keys: waiting for work action  Ctrl+C exits"},
		{Active: m.artifactAction.Pending, Text: "Keys: enter confirms action, esc cancels action"},
		{Active: m.dumpAction.Pending, Text: "Keys: enter confirms dump action, esc cancels dump action"},
		{Active: m.ticketAction.Pending, Text: "Keys: enter confirms ticket action, esc cancels ticket action"},
		{Active: m.specAction.Pending, Text: "Keys: enter confirms spec action, esc cancels spec action"},
		{Active: m.specBusy, Text: "Keys: waiting for spec action to finish"},
		{Active: m.planAction.Pending, Text: "Keys: enter confirms plan action, esc cancels plan action"},
		{Active: m.planBusy, Text: "Keys: waiting for plan action to finish"},
		{Active: m.workAction.Pending, Text: "Keys: enter confirms work action, esc cancels work action"},
		{Active: m.commandPalette.Open, Text: "Keys: up/down selects command, enter runs command, esc closes palette"},
		{Active: m.quizBusy, Text: "Keys: waiting for quiz action to finish"},
		{Active: m.quizAnswer.Active, Text: "Keys: enter answer  alt+s skip  alt+d done  alt+x cancel quiz  alt+m mode  PgUp/PgDn history  esc discard"},
		{Active: m.SelectedView() == ViewDash && m.dashWorkflowDetails, Text: "Keys: c continue  p pause  x cancel  u resume  a approve  R reconcile  g allowance  d decision  t/C links  esc queue"},
		{Active: m.filterEntryMode, Text: "Keys: enter applies filter, esc cancels filter, ctrl+c quits"},
	}
	for _, hint := range hints {
		if hint.Active {
			return hint.Text, true
		}
	}

	return "", false
}

func (m *Model) renderStatusMessage() string {
	if m.operationStatus() == "" {
		return ""
	}

	return "\n\nStatus\n  " + m.operationStatus()
}
