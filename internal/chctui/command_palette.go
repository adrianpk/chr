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
)

type Command struct {
	ID          string
	Title       string
	Description string
	Action      CommandAction
	View        View
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
		{ID: "refresh", Title: "Refresh", Description: "Close the palette and keep the current read-only data visible.", Action: CommandActionRefresh},
		{ID: "help", Title: "Help", Description: "Navigate to the Help view.", Action: CommandActionHelp, View: ViewHelp},
		{ID: "open-source", Title: "Open selected source path", Description: "Show the selected source path for external opening.", Action: CommandActionOpenSelectedSource},
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
		m.statusMessage = "Refresh requested. Restart the TUI to reload read-only data."
	case CommandActionOpenSelectedSource:
		m.statusMessage = m.selectedSourcePathMessage()
	}
}

func (m *Model) selectView(view View) {
	for index, candidate := range m.views {
		if candidate == view {
			m.selected = index

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
	case ViewDocuments:
		document, ok := m.selectedDocumentView()
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
