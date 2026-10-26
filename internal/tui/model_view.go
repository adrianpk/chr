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

//nolint:cyclop // Key helper text is centralized by selected view.
func (m *Model) renderKeys() string {
	if keys, ok := m.activeInteractionKeys(); ok {
		return keys
	}

	switch {
	case m.SelectedView() == ViewBoard:
		return "Keys: F1-F9 or click changes tab, tab changes pane, left/right changes Board column, alt+a toggles attention, alt+e edits source, up/down selects or scrolls, / filters, q quits"
	case m.ticketKeysActive():
		return "Keys: F1-F9 or click changes tab, tab changes pane, alt+e edits source, up/down selects or scrolls, / filters, q quits"
	case m.documentKeysActive():
		return "Keys: F1-F9 or click changes tab, tab changes pane, alt+e edits source, up/down selects or scrolls, q quits"
	case m.SelectedView() == ViewProject:
		return "Keys: F1-F9 or click changes tab, tab changes pane, up/down selects field, enter edits, s saves, r refreshes, q quits"
	case m.SelectedView() == ViewCases:
		return "Keys: F1-F9 or click changes tab, tab changes pane, up/down selects focused pane, enter opens or edits, g guided use cases, s saves, r refreshes, q quits"
	case m.SelectedView() == ViewDumps:
		return "Keys: F1-F9 or click changes tab, tab changes pane, up/down selects or scrolls, enter edits, alt+e external, c creates, r refreshes, q quits"
	case m.SelectedView() == ViewQuizzes:
		return "Keys: F1-F9 or click changes tab, tab changes pane, alt+e edits source, up/down selects or scrolls, s starts, enter answers, r refreshes, q quits"
	case m.SelectedView() == ViewSpecs:
		return "Keys: F1-F9 or click changes tab, tab changes pane, alt+e edits source, up/down selects or scrolls, g generates spec, r refreshes, q quits"
	case m.SelectedView() == ViewPlans:
		return "Keys: F1-F9 or click changes tab, tab changes pane, alt+e edits source, up/down selects or scrolls, g generates missing plan, G regenerates, r refreshes, q quits"
	case m.SelectedView() == ViewDash:
		return "Keys: F1-F9 or click changes tab, tab changes pane, a runs all ready tickets, enter starts selected ticket, x repairs readiness, s requests stop, q quits"
	case m.SelectedView() == ViewAuthoring:
		return "Keys: F1-F9 or click changes tab, tab changes pane, alt+e edits source, up/down selects or scrolls, q quits"
	default:
		return "Keys: F1-F9 or click changes tab, tab changes pane, q quits"
	}
}

func (m *Model) ticketKeysActive() bool {
	return m.SelectedView() == ViewBoard || m.SelectedView() == ViewTickets
}

func (m *Model) documentKeysActive() bool {
	switch m.SelectedView() {
	case ViewDocuments:
		return true
	default:
		return false
	}
}

//nolint:cyclop // Active interaction key hints mirror modal precedence.
func (m *Model) activeInteractionKeys() (string, bool) {
	switch {
	case m.artifactAction.Pending:
		return "Keys: enter confirms action, esc cancels action", true
	case m.dumpAction.Pending:
		return "Keys: enter confirms dump action, esc cancels dump action", true
	case m.ticketAction.Pending:
		return "Keys: enter confirms ticket action, esc cancels ticket action", true
	case m.specAction.Pending:
		return "Keys: enter confirms spec action, esc cancels spec action", true
	case m.specBusy:
		return "Keys: waiting for spec action to finish", true
	case m.planAction.Pending:
		return "Keys: enter confirms plan action, esc cancels plan action", true
	case m.planBusy:
		return "Keys: waiting for plan action to finish", true
	case m.workAction.Pending:
		return "Keys: enter confirms work action, esc cancels work action", true
	case m.commandPalette.Open:
		return "Keys: up/down selects command, enter runs command, esc closes palette", true
	case m.quizBusy:
		return "Keys: waiting for quiz action to finish", true
	case m.quizAnswer.Active:
		return "Keys: enter submits answer, shift+enter inserts line, alt+s skips, alt+d done, esc cancels answer", true
	case m.filterEntryMode:
		return "Keys: enter applies filter, esc cancels filter, ctrl+c quits", true
	default:
		return "", false
	}
}

func (m *Model) renderStatusMessage() string {
	if m.statusMessage == "" {
		return ""
	}

	return "\n\nStatus\n  " + m.statusMessage
}
