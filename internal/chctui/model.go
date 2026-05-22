package chctui

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chc/internal/chc"
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
	ProjectDir        string
	ProjectID         string
	Stdin             io.Reader
	Stdout            io.Writer
	Tickets           []chc.IndexedTicket
	Documents         []DocumentView
	AuthoringSessions []chc.AuthoringSession
	Observability     ObservabilityView
	ActionExecutor    ArtifactActionExecutor
	TicketExecutor    TicketActionExecutor
	WorkExecutor      WorkActionExecutor
}

type Model struct {
	projectDir               string
	projectID                string
	views                    []View
	selected                 int
	width                    int
	height                   int
	tickets                  []TicketView
	selectedTicket           int
	ticketFilter             string
	filterEntryMode          bool
	documents                []DocumentView
	selectedDocument         int
	authoringSessions        []AuthoringSessionView
	selectedAuthoringSession int
	observability            ObservabilityView
	commandPalette           CommandPaletteState
	artifactAction           ArtifactActionState
	actionExecutor           ArtifactActionExecutor
	ticketAction             TicketActionState
	ticketExecutor           TicketActionExecutor
	workAction               WorkActionState
	workExecutor             WorkActionExecutor
	statusMessage            string
}

func NewModel(opts Options) Model {
	return Model{
		projectDir:        opts.ProjectDir,
		projectID:         opts.ProjectID,
		views:             append([]View(nil), defaultViews...),
		tickets:           NewTicketViews(opts.Tickets),
		documents:         append([]DocumentView(nil), opts.Documents...),
		authoringSessions: NewAuthoringSessionViews(opts.AuthoringSessions),
		observability:     opts.Observability,
		commandPalette:    NewCommandPaletteState(DefaultCommandRegistry()),
		actionExecutor:    opts.ActionExecutor,
		ticketExecutor:    opts.TicketExecutor,
		workExecutor:      opts.WorkExecutor,
	}
}

func Run(ctx context.Context, opts Options) error {
	programOptions := []tea.ProgramOption{
		tea.WithContext(ctx),
		tea.WithOutput(opts.Stdout),
		tea.WithAltScreen(),
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
	if m.width > 0 && m.height > 0 {
		return m.renderShell()
	}

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
	output.WriteString(m.renderStatusMessage())
	output.WriteString(m.renderArtifactAction())
	output.WriteString(m.renderTicketAction())
	output.WriteString(m.renderWorkAction())
	output.WriteString(m.renderCommandPalette())
	output.WriteString("\n\n")
	output.WriteString(m.renderKeys())
	output.WriteString("\n")

	return output.String()
}

func (m Model) SelectedView() View {
	return m.views[m.selected]
}

func (m Model) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.artifactAction.Pending {
		return m.updateArtifactActionKey(msg)
	}

	if m.ticketAction.Pending {
		return m.updateTicketActionKey(msg)
	}

	if m.workAction.Pending {
		return m.updateWorkActionKey(msg)
	}

	if m.commandPalette.Open {
		return m.updateCommandPaletteKey(msg)
	}

	if m.filterEntryMode {
		return m.updateFilterKey(msg)
	}

	return m.updateNavigationKey(msg)
}

func (m Model) updateNavigationKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "left", "shift+tab", "h":
		m.selectPreviousView()
	case "right", "tab", "l":
		m.selectNextView()
	case "up", "k":
		m.selectPreviousRecord()
	case "down", "j":
		m.selectNextRecord()
	case ":":
		m.openCommandPalette()
	case "/":
		m.filterEntryMode = true
	}

	return m, nil
}

func (m *Model) selectPreviousRecord() {
	switch m.SelectedView() {
	case ViewBoard, ViewTickets:
		m.selectPreviousTicket()
	case ViewDocuments:
		m.selectPreviousDocument()
	case ViewAuthoring:
		m.selectPreviousAuthoringSession()
	}
}

func (m *Model) selectNextRecord() {
	switch m.SelectedView() {
	case ViewBoard, ViewTickets:
		m.selectNextTicket()
	case ViewDocuments:
		m.selectNextDocument()
	case ViewAuthoring:
		m.selectNextAuthoringSession()
	}
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

func (m Model) renderTicketBoard() string {
	filteredTickets := m.filteredTickets()
	if len(filteredTickets) == 0 {
		return "Board\n  No tickets match the current view."
	}

	var output strings.Builder

	output.WriteString("Board\n")
	output.WriteString(fmt.Sprintf("  Tickets: %d", len(filteredTickets)))

	if m.ticketFilter != "" {
		output.WriteString(fmt.Sprintf("  Filter: %q", m.ticketFilter))
	}

	output.WriteString("\n\n")

	for _, group := range groupTicketsByStatus(filteredTickets) {
		output.WriteString(group.Status)
		output.WriteString("\n")

		for _, ticket := range group.Tickets {
			prefix := "  "
			if ticket.Index == m.selectedTicket {
				prefix = "> "
			}

			output.WriteString(prefix)
			output.WriteString(renderTicketRow(ticket))
			output.WriteString("\n")
		}

		output.WriteString("\n")
	}

	output.WriteString(m.renderSelectedTicketDetail())

	return strings.TrimRight(output.String(), "\n")
}

func (m Model) renderTicketList() string {
	filteredTickets := m.filteredTickets()
	if len(filteredTickets) == 0 {
		return "Tickets\n  No tickets match the current view."
	}

	var output strings.Builder

	output.WriteString("Tickets\n")
	output.WriteString("  Key     Status      Priority  Type      Title\n")

	for _, ticket := range filteredTickets {
		prefix := "  "
		if ticket.Index == m.selectedTicket {
			prefix = "> "
		}

		output.WriteString(prefix)
		output.WriteString(renderTicketRow(ticket))
		output.WriteString("\n")
	}

	output.WriteString("\n")
	output.WriteString(m.renderSelectedTicketDetail())

	return strings.TrimRight(output.String(), "\n")
}

func renderTicketRow(ticket TicketView) string {
	return fmt.Sprintf(
		"%-7s %-11s %-9s %-9s %s",
		ticket.Key,
		ticket.Status,
		ticket.Priority,
		ticket.Type,
		ticket.Title,
	)
}

func (m Model) renderSelectedTicketDetail() string {
	ticket, ok := m.selectedTicketView()
	if !ok {
		return "Selected Ticket\n  No ticket selected."
	}

	var output strings.Builder

	output.WriteString("Selected Ticket\n")
	output.WriteString(fmt.Sprintf("  Key: %s\n", ticket.Key))
	output.WriteString(fmt.Sprintf("  Title: %s\n", ticket.Title))
	output.WriteString(fmt.Sprintf("  Status: %s\n", ticket.Status))
	output.WriteString(fmt.Sprintf("  Priority: %s\n", ticket.Priority))
	output.WriteString(fmt.Sprintf("  Severity: %s\n", ticket.Severity))
	output.WriteString(fmt.Sprintf("  Type: %s\n", ticket.Type))
	output.WriteString(fmt.Sprintf("  Use Case: %s\n", ticket.UseCase))
	output.WriteString(fmt.Sprintf("  Project Area: %s\n", ticket.ProjectArea))
	output.WriteString(fmt.Sprintf("  Checklist: %d done / %d total\n", ticket.ChecklistDone, ticket.ChecklistTotal))
	output.WriteString(fmt.Sprintf("  Comments: %d\n", ticket.CommentCount))
	output.WriteString(fmt.Sprintf("  Path: %s", ticket.RelativePath))

	return output.String()
}

func (m Model) renderKeys() string {
	if m.artifactAction.Pending {
		return "Keys: enter confirms action, esc cancels action"
	}

	if m.ticketAction.Pending {
		return "Keys: enter confirms ticket action, esc cancels ticket action"
	}

	if m.workAction.Pending {
		return "Keys: enter confirms work action, esc cancels work action"
	}

	if m.commandPalette.Open {
		return "Keys: up/down selects command, enter runs command, esc closes palette"
	}

	if m.filterEntryMode {
		return "Keys: enter applies filter, esc cancels filter, ctrl+c quits"
	}

	if m.SelectedView() == ViewBoard || m.SelectedView() == ViewTickets {
		return "Keys: up/down selects ticket, left/right or tab changes view, / filters, q quits"
	}

	if m.SelectedView() == ViewDocuments {
		return "Keys: up/down selects document, left/right or tab changes view, q quits"
	}

	if m.SelectedView() == ViewAuthoring {
		return "Keys: up/down selects session, left/right or tab changes view, q quits"
	}

	return "Keys: left/right or tab changes view, q quits"
}

func (m Model) renderStatusMessage() string {
	if m.statusMessage == "" {
		return ""
	}

	return "\n\nStatus\n  " + m.statusMessage
}

func (m Model) selectedTicketView() (TicketView, bool) {
	if len(m.tickets) == 0 {
		return TicketView{}, false
	}

	for _, ticket := range m.filteredTickets() {
		if ticket.Index == m.selectedTicket {
			return ticket, true
		}
	}

	return TicketView{}, false
}

func (m *Model) selectPreviousTicket() {
	if !m.ticketSelectionActive() {
		return
	}

	filteredTickets := m.filteredTickets()
	if len(filteredTickets) == 0 {
		return
	}

	position := m.selectedTicketPosition(filteredTickets)
	if position == 0 {
		m.selectedTicket = filteredTickets[len(filteredTickets)-1].Index

		return
	}

	m.selectedTicket = filteredTickets[position-1].Index
}

func (m *Model) selectNextTicket() {
	if !m.ticketSelectionActive() {
		return
	}

	filteredTickets := m.filteredTickets()
	if len(filteredTickets) == 0 {
		return
	}

	position := m.selectedTicketPosition(filteredTickets)
	if position >= len(filteredTickets)-1 {
		m.selectedTicket = filteredTickets[0].Index

		return
	}

	m.selectedTicket = filteredTickets[position+1].Index
}

func (m Model) ticketSelectionActive() bool {
	return m.SelectedView() == ViewBoard || m.SelectedView() == ViewTickets
}

func (m Model) selectedTicketPosition(tickets []TicketView) int {
	for index, ticket := range tickets {
		if ticket.Index == m.selectedTicket {
			return index
		}
	}

	return 0
}

func (m Model) filteredTickets() []TicketView {
	if m.ticketFilter == "" {
		return append([]TicketView(nil), m.tickets...)
	}

	filter := strings.ToLower(m.ticketFilter)
	tickets := []TicketView{}

	for _, ticket := range m.tickets {
		if ticket.Matches(filter) {
			tickets = append(tickets, ticket)
		}
	}

	return tickets
}

func (m Model) updateFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.filterEntryMode = false
	case tea.KeyEnter:
		m.filterEntryMode = false
		m.selectFirstFilteredTicket()
	case tea.KeyBackspace:
		m.ticketFilter = strings.TrimSuffix(m.ticketFilter, lastRune(m.ticketFilter))
	default:
		if msg.Type == tea.KeyRunes {
			m.ticketFilter += msg.String()
		}
	}

	return m, nil
}

func (m *Model) selectFirstFilteredTicket() {
	tickets := m.filteredTickets()
	if len(tickets) == 0 {
		m.selectedTicket = 0

		return
	}

	m.selectedTicket = tickets[0].Index
}

func lastRune(value string) string {
	if value == "" {
		return ""
	}

	runes := []rune(value)

	return string(runes[len(runes)-1])
}

type TicketView struct {
	Index          int
	Key            string
	Title          string
	Status         string
	Priority       string
	Severity       string
	Type           string
	UseCase        string
	ProjectArea    string
	RelativePath   string
	ChecklistDone  int
	ChecklistTotal int
	CommentCount   int
}

type TicketStatusGroup struct {
	Status  string
	Tickets []TicketView
}

func NewTicketViews(tickets []chc.IndexedTicket) []TicketView {
	views := make([]TicketView, 0, len(tickets))

	for index, ticket := range tickets {
		view := TicketView{
			Index:        index,
			Key:          ticket.Key,
			Title:        ticket.Title,
			Status:       ticket.Status,
			Priority:     ticket.Priority,
			Severity:     ticket.Severity,
			Type:         ticket.Type,
			UseCase:      ticket.UseCase,
			ProjectArea:  ticket.ProjectArea,
			RelativePath: ticket.RelativePath,
		}

		view.ChecklistDone, view.ChecklistTotal, view.CommentCount = ticketMarkdownSummary(ticket.Path)

		views = append(views, view)
	}

	return views
}

func (t TicketView) Matches(filter string) bool {
	values := []string{
		t.Key,
		t.Title,
		t.Status,
		t.Priority,
		t.Severity,
		t.Type,
		t.UseCase,
		t.ProjectArea,
		t.RelativePath,
	}

	for _, value := range values {
		if strings.Contains(strings.ToLower(value), filter) {
			return true
		}
	}

	return false
}

func groupTicketsByStatus(tickets []TicketView) []TicketStatusGroup {
	grouped := map[string][]TicketView{}

	for _, ticket := range tickets {
		grouped[ticket.Status] = append(grouped[ticket.Status], ticket)
	}

	statuses := make([]string, 0, len(grouped))
	for status := range grouped {
		statuses = append(statuses, status)
	}

	sort.Strings(statuses)

	groups := make([]TicketStatusGroup, 0, len(statuses))
	for _, status := range statuses {
		groups = append(groups, TicketStatusGroup{
			Status:  status,
			Tickets: grouped[status],
		})
	}

	return groups
}

func ticketMarkdownSummary(path string) (int, int, int) {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, 0
	}

	checklistDone := 0
	checklistTotal := 0
	commentCount := 0
	inComments := false

	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "# ") {
			inComments = trimmed == "# Comments"
		}

		if strings.HasPrefix(trimmed, "- [") {
			checklistTotal++

			if strings.HasPrefix(trimmed, "- [x]") || strings.HasPrefix(trimmed, "- [X]") {
				checklistDone++
			}
		}

		if inComments && strings.HasPrefix(trimmed, "## ") {
			commentCount++
		}
	}

	return checklistDone, checklistTotal, commentCount
}
