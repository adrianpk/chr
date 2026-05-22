package chctui

import (
	"context"
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type View string

const (
	ViewDashboard View = "Dashboard"
	ViewBoard     View = "Board"
	ViewTickets   View = "Tickets"
	ViewDocuments View = "Documents"
	ViewAuthoring View = "Authoring"
	ViewAgent     View = "Agent"
	ViewRuntime   View = "Runtime"
	ViewReports   View = "Reports"
	ViewHelp      View = "Help"
)

var defaultViews = []View{
	ViewDashboard,
	ViewBoard,
	ViewTickets,
	ViewDocuments,
	ViewAuthoring,
	ViewAgent,
	ViewRuntime,
	ViewReports,
	ViewHelp,
}

var viewPlaceholders = map[View]string{
	ViewDashboard: "Project overview will appear here.",
	ViewBoard:     "Ticket board will appear here.",
	ViewTickets:   "Ticket list and detail will appear here.",
	ViewDocuments: "Project documents will appear here.",
	ViewAuthoring: "Authoring sessions and quizzes will appear here.",
	ViewAgent:     "Agent work state will appear here.",
	ViewRuntime:   "Runtime readiness and execution evidence will appear here.",
	ViewReports:   "Reports and handoff material will appear here.",
	ViewHelp:      "Command palette and key help will appear here.",
}

type Options struct {
	ProjectDir string
	ProjectID  string
	Stdin      io.Reader
	Stdout     io.Writer
}

type Model struct {
	projectDir string
	projectID  string
	views      []View
	selected   int
	width      int
	height     int
}

func NewModel(opts Options) Model {
	return Model{
		projectDir: opts.ProjectDir,
		projectID:  opts.ProjectID,
		views:      append([]View(nil), defaultViews...),
	}
}

func Run(ctx context.Context, opts Options) error {
	programOptions := []tea.ProgramOption{
		tea.WithContext(ctx),
		tea.WithOutput(opts.Stdout),
	}

	if opts.Stdin != nil {
		programOptions = append(programOptions, tea.WithInput(opts.Stdin))
	}

	program := tea.NewProgram(NewModel(opts), programOptions...)

	_, err := program.Run()
	if err != nil {
		return fmt.Errorf("run tui: %w", err)
	}

	return nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typedMsg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typedMsg.Width
		m.height = typedMsg.Height

		return m, nil
	case tea.KeyMsg:
		return m.updateKey(typedMsg)
	default:
		return m, nil
	}
}

func (m Model) View() string {
	var output strings.Builder

	output.WriteString("ChrysoCoord\n\n")
	output.WriteString("Project\n")
	output.WriteString(fmt.Sprintf("  ID:  %s\n", m.projectID))
	output.WriteString(fmt.Sprintf("  Dir: %s\n\n", m.projectDir))
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
	output.WriteString("\n\n")
	output.WriteString("Keys: up/down or j/k navigate, tab cycles, q quits\n")

	return output.String()
}

func (m Model) SelectedView() View {
	return m.views[m.selected]
}

func (m Model) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		m.selectPreviousView()
	case "down", "j", "tab":
		m.selectNextView()
	}

	return m, nil
}

func (m *Model) selectPreviousView() {
	if m.selected == 0 {
		m.selected = len(m.views) - 1

		return
	}

	m.selected--
}

func (m *Model) selectNextView() {
	m.selected++
	if m.selected >= len(m.views) {
		m.selected = 0
	}
}

func (m Model) renderSelectedView() string {
	selectedView := m.SelectedView()

	placeholder, ok := viewPlaceholders[selectedView]
	if !ok {
		return "Unknown\n  No view is selected."
	}

	return fmt.Sprintf("%s\n  %s", selectedView, placeholder)
}
