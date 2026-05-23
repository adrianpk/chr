package chctui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type CommandAction string

const (
	CommandActionNavigate           CommandAction = "navigate"
	CommandActionRefresh            CommandAction = "refresh"
	CommandActionHelp               CommandAction = "help"
	CommandActionOpenSelectedSource CommandAction = "open_selected_source"
	CommandActionPlanArtifact       CommandAction = "plan_artifact"
	CommandActionPlanTicket         CommandAction = "plan_ticket"
	CommandActionPlanWork           CommandAction = "plan_work"
)

type Command struct {
	ID          string
	Title       string
	Description string
	Action      CommandAction
	View        View
	Artifact    ArtifactActionID
	Ticket      TicketActionID
	Status      string
	Work        WorkActionID
}

type CommandPaletteState struct {
	Open     bool
	Filter   string
	Selected int
	Commands []Command
}

func NewCommandPaletteState(commands []Command) CommandPaletteState {
	return CommandPaletteState{
		Commands: append([]Command(nil), commands...),
	}
}

func DefaultCommandRegistry() []Command {
	commands := []Command{
		{ID: "refresh", Title: "Refresh", Description: "Reload workspace files into the runtime graph.", Action: CommandActionRefresh},
		{ID: "help", Title: "Help", Description: "Navigate to the Help view.", Action: CommandActionHelp, View: ViewHelp},
		{ID: "open-source", Title: "Open selected source path", Description: "Show the selected source path for external opening.", Action: CommandActionOpenSelectedSource},
		{ID: "create-ticket", Title: "Create ticket", Description: "Create a default task ticket after confirmation.", Action: CommandActionPlanArtifact, Artifact: ArtifactActionCreateTicket},
		{ID: "create-authoring-session", Title: "Create authoring session", Description: "Create a draft authoring session after confirmation.", Action: CommandActionPlanArtifact, Artifact: ArtifactActionCreateAuthoringSession},
		{ID: "create-context-quiz", Title: "Create context-aware quiz", Description: "Plan a context-aware quiz creation action.", Action: CommandActionPlanArtifact, Artifact: ArtifactActionCreateContextQuiz},
		{ID: "create-draft", Title: "Create draft", Description: "Plan a safe draft creation action.", Action: CommandActionPlanArtifact, Artifact: ArtifactActionCreateDraft},
		{ID: "ticket-add-comment", Title: "Add ticket comment", Description: "Add a standard comment to the selected ticket after confirmation.", Action: CommandActionPlanTicket, Ticket: TicketActionAddComment},
		{ID: "ticket-status-ready", Title: "Move ticket to ready", Description: "Move the selected ticket to ready after confirmation.", Action: CommandActionPlanTicket, Ticket: TicketActionMoveStatus, Status: "ready"},
		{ID: "ticket-status-in-progress", Title: "Move ticket to in_progress", Description: "Move the selected ticket to in_progress after confirmation.", Action: CommandActionPlanTicket, Ticket: TicketActionMoveStatus, Status: "in_progress"},
		{ID: "ticket-status-reviewing", Title: "Move ticket to reviewing", Description: "Move the selected ticket to reviewing after confirmation.", Action: CommandActionPlanTicket, Ticket: TicketActionMoveStatus, Status: "reviewing"},
		{ID: "ticket-checklist-toggle", Title: "Toggle ticket checklist item", Description: "Validate checklist toggle availability for the selected ticket.", Action: CommandActionPlanTicket, Ticket: TicketActionToggleChecklist},
		{ID: "ticket-checklist-add", Title: "Add ticket checklist item", Description: "Validate checklist creation availability for the selected ticket.", Action: CommandActionPlanTicket, Ticket: TicketActionAddChecklistItem},
		{ID: "work-start-ticket", Title: "Start selected ticket", Description: "Run the agent harness for the selected ticket after confirmation.", Action: CommandActionPlanWork, Work: WorkActionStartTicket},
		{ID: "work-start-ready-loop", Title: "Start ready ticket loop", Description: "Run one ready ticket through the agent loop after confirmation.", Action: CommandActionPlanWork, Work: WorkActionStartReadyLoop},
		{ID: "work-request-stop", Title: "Request agent stop", Description: "Request daemon stop after confirmation.", Action: CommandActionPlanWork, Work: WorkActionRequestStop},
	}

	for _, view := range defaultViews {
		commands = append(commands, Command{
			ID:          "view:" + strings.ToLower(string(view)),
			Title:       "Go to " + string(view),
			Description: "Navigate to the " + string(view) + " view.",
			Action:      CommandActionNavigate,
			View:        view,
		})
	}

	return commands
}

func (m Model) renderCommandPalette() string {
	if !m.commandPalette.Open {
		return ""
	}

	var output strings.Builder

	output.WriteString("\n\nCommand Palette\n")
	output.WriteString(fmt.Sprintf("  Filter: %s\n", m.commandPalette.Filter))

	commands := m.filteredCommands()
	if len(commands) == 0 {
		output.WriteString("  No commands match.")

		return output.String()
	}

	for index, command := range commands {
		prefix := "  "
		if index == m.commandPalette.Selected {
			prefix = "> "
		}

		output.WriteString(prefix)
		output.WriteString(command.Title)
		output.WriteString(" - ")
		output.WriteString(command.Description)
		output.WriteString("\n")
	}

	return strings.TrimRight(output.String(), "\n")
}

func (m Model) updateCommandPaletteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.closeCommandPalette()
	case tea.KeyEnter:
		m.executeSelectedCommand()
	case tea.KeyBackspace:
		m.commandPalette.Filter = strings.TrimSuffix(m.commandPalette.Filter, lastRune(m.commandPalette.Filter))
		m.boundSelectedCommand()
	case tea.KeyUp:
		m.selectPreviousCommand()
	case tea.KeyDown:
		m.selectNextCommand()
	default:
		return m.updateCommandPaletteRunes(msg)
	}

	return m, nil
}

func (m Model) updateCommandPaletteRunes(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type != tea.KeyRunes {
		return m, nil
	}

	m.commandPalette.Filter += msg.String()
	m.commandPalette.Selected = 0

	return m, nil
}

func (m *Model) openCommandPalette() {
	m.commandPalette.Open = true
	m.commandPalette.Filter = ""
	m.commandPalette.Selected = 0
	m.statusMessage = ""
}

func (m *Model) closeCommandPalette() {
	m.commandPalette.Open = false
	m.commandPalette.Filter = ""
	m.commandPalette.Selected = 0
}

func (m *Model) selectPreviousCommand() {
	commands := m.filteredCommands()
	if len(commands) == 0 {
		return
	}

	if m.commandPalette.Selected == 0 {
		m.commandPalette.Selected = len(commands) - 1

		return
	}

	m.commandPalette.Selected--
}

func (m *Model) selectNextCommand() {
	commands := m.filteredCommands()
	if len(commands) == 0 {
		return
	}

	m.commandPalette.Selected++
	if m.commandPalette.Selected >= len(commands) {
		m.commandPalette.Selected = 0
	}
}

func (m *Model) boundSelectedCommand() {
	commands := m.filteredCommands()
	if len(commands) == 0 || m.commandPalette.Selected >= len(commands) {
		m.commandPalette.Selected = 0
	}
}

func (m *Model) executeSelectedCommand() {
	commands := m.filteredCommands()
	if len(commands) == 0 {
		m.statusMessage = "No command selected."
		m.closeCommandPalette()

		return
	}

	command := commands[m.commandPalette.Selected]
	m.closeCommandPalette()
	m.executeCommand(command)
}

func (m *Model) executeCommand(command Command) {
	switch command.Action {
	case CommandActionNavigate, CommandActionHelp:
		m.selectView(command.View)

		m.statusMessage = "Navigated to " + string(command.View) + "."
	case CommandActionRefresh:
		m.refreshWorkspace()
	case CommandActionOpenSelectedSource:
		m.statusMessage = m.selectedSourcePathMessage()
	case CommandActionPlanArtifact:
		m.planArtifactAction(command.Artifact)
	case CommandActionPlanTicket:
		m.planTicketAction(command.Ticket, command.Status)
	case CommandActionPlanWork:
		m.planWorkAction(command.Work)
	}
}

func (m *Model) selectView(view View) {
	for index, candidate := range m.views {
		if candidate == view {
			m.selected = index
			m.selectPrimaryPane()

			return
		}
	}
}

func (m Model) filteredCommands() []Command {
	filter := strings.ToLower(strings.TrimSpace(m.commandPalette.Filter))
	if filter == "" {
		return append([]Command(nil), m.commandPalette.Commands...)
	}

	commands := []Command{}

	for _, command := range m.commandPalette.Commands {
		if command.Matches(filter) {
			commands = append(commands, command)
		}
	}

	return commands
}

func (c Command) Matches(filter string) bool {
	return strings.Contains(strings.ToLower(c.ID), filter) ||
		strings.Contains(strings.ToLower(c.Title), filter) ||
		strings.Contains(strings.ToLower(c.Description), filter)
}

func (m Model) selectedSourcePathMessage() string {
	path := m.selectedSourcePath()
	if path == "" {
		return "No selected source path is available."
	}

	return "Selected source path: " + path
}

func (m Model) selectedSourcePath() string {
	switch m.SelectedView() {
	case ViewBoard, ViewTickets:
		ticket, ok := m.selectedTicketView()
		if ok {
			return ticket.RelativePath
		}
	case ViewDumps, ViewQuizzes, ViewSpecs, ViewPlans, ViewDocuments:
		document, ok := m.selectedDocumentViewFor(m.currentDocumentViews())
		if ok {
			return document.RelativePath
		}
	case ViewAuthoring:
		session, ok := m.selectedAuthoringSessionView()
		if ok {
			return session.SessionPath
		}
	}

	return ""
}
