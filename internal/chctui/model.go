package chctui

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chc/internal/chc"
)

type View string

const (
	ViewProject View = "Project"
	ViewCases   View = "Cases"
	ViewDumps   View = "Dumps"
	ViewQuizzes View = "Quizzes"
	ViewSpecs   View = "Specs"
	ViewPlans   View = "Plans"
	ViewDash    View = "Dash"
	ViewHelp    View = "Help"

	// Legacy internal views kept for existing commands and non-shell rendering paths.
	ViewDashboard View = "Dashboard"
	ViewBoard     View = "Board"
	ViewTickets   View = "Tickets"
	ViewDocuments View = "Documents"
	ViewAuthoring View = "Authoring"
	ViewAgent     View = "Agent"
	ViewRuntime   View = "Runtime"
	ViewReports   View = "Reports"
)

var defaultViews = []View{
	ViewProject,
	ViewCases,
	ViewDumps,
	ViewQuizzes,
	ViewSpecs,
	ViewPlans,
	ViewBoard,
	ViewDash,
	ViewHelp,
}

var viewPlaceholders = map[View]string{
	ViewProject:   "Project context will appear here.",
	ViewCases:     "Use cases will appear here.",
	ViewDumps:     "Project and use-case dumps will appear here.",
	ViewQuizzes:   "Quiz administration will appear here.",
	ViewSpecs:     "Specs will appear here.",
	ViewPlans:     "Plans will appear here.",
	ViewDash:      "Agent activity will appear here.",
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
	ProjectDir           string
	ProjectID            string
	ProjectMetadata      chc.ProjectMetadata
	Cases                []chc.CaseMetadata
	Stdin                io.Reader
	Stdout               io.Writer
	Tickets              []chc.IndexedTicket
	Documents            []DocumentView
	Dumps                []DumpView
	Quizzes              []QuizView
	Specs                []SpecView
	Plans                []PlanView
	AuthoringSessions    []chc.AuthoringSession
	Observability        ObservabilityView
	WorkspaceGraph       chc.WorkspaceGraph
	WorkspaceIndex       chc.WorkspaceIndex
	SemanticDigestStates []chc.WorkspaceSemanticDigestState
	ActionExecutor       ArtifactActionExecutor
	DumpExecutor         DumpActionExecutor
	QuizExecutor         QuizActionExecutor
	SpecExecutor         SpecActionExecutor
	PlanExecutor         PlanActionExecutor
	TicketExecutor       TicketActionExecutor
	WorkExecutor         WorkActionExecutor
	AgentWorkExecutor    AgentWorkRefreshExecutor
	WorkspaceExecutor    WorkspaceRefreshExecutor
	ProjectExecutor      ProjectMetadataExecutor
	CaseExecutor         CaseMetadataExecutor
	SourceExecutor       SourceOpenExecutor
}

type Model struct {
	projectDir               string
	projectID                string
	projectMetadata          chc.ProjectMetadata
	selectedProjectField     int
	projectEdit              ProjectEditState
	projectDirty             bool
	cases                    []chc.CaseMetadata
	selectedCase             int
	selectedCaseField        int
	caseEdit                 ProjectEditState
	caseDirty                bool
	caseFilter               string
	caseFilterEditing        bool
	views                    []View
	selected                 int
	selectedPane             int
	maximizedPane            int
	verticalSplitPercent     int
	horizontalSplitPercent   int
	layoutDrag               layoutDrag
	width                    int
	height                   int
	tickets                  []TicketView
	selectedTicket           int
	agentWorkPriorityFilter  string
	agentWorkOrder           chc.TicketLoopOrder
	boardColumn              string
	previousBoardColumn      string
	ticketFilter             string
	filterEntryMode          bool
	documents                []DocumentView
	selectedDocument         int
	dumps                    []DumpView
	selectedDump             int
	dumpEdit                 DumpEditState
	quizzes                  []QuizView
	selectedQuiz             int
	quizAction               QuizActionState
	quizAnswer               QuizAnswerState
	quizBusy                 bool
	specs                    []SpecView
	selectedSpec             int
	plans                    []PlanView
	selectedPlan             int
	authoringSessions        []AuthoringSessionView
	selectedAuthoringSession int
	observability            ObservabilityView
	workspaceGraph           chc.WorkspaceGraph
	workspaceIndex           chc.WorkspaceIndex
	semanticDigestStates     []chc.WorkspaceSemanticDigestState
	commandPalette           CommandPaletteState
	artifactAction           ArtifactActionState
	actionExecutor           ArtifactActionExecutor
	dumpAction               DumpActionState
	dumpExecutor             DumpActionExecutor
	quizExecutor             QuizActionExecutor
	specAction               SpecActionState
	specBusy                 bool
	specExecutor             SpecActionExecutor
	planAction               PlanActionState
	planBusy                 bool
	planExecutor             PlanActionExecutor
	ticketAction             TicketActionState
	ticketExecutor           TicketActionExecutor
	workAction               WorkActionState
	workBusy                 bool
	workExecutor             WorkActionExecutor
	agentWorkRefreshExecutor AgentWorkRefreshExecutor
	agentWorkRefreshInFlight bool
	workspaceExecutor        WorkspaceRefreshExecutor
	projectExecutor          ProjectMetadataExecutor
	caseExecutor             CaseMetadataExecutor
	sourceExecutor           SourceOpenExecutor
	detailScroll             int
	statusMessage            string
}

type layoutDrag int

const (
	layoutDragNone layoutDrag = iota
	layoutDragVertical
	layoutDragHorizontal
)

const (
	defaultPanelSplitPercent = 50
	minPanelSplitPercent     = 25
	maxPanelSplitPercent     = 75
	boardAttentionColumn     = "attention"
)

//nolint:funlen // Model construction keeps option mapping in one place.
func NewModel(opts Options) Model {
	workspaceIndex := opts.WorkspaceIndex
	if len(workspaceIndex.NodesByID) == 0 {
		workspaceIndex = chc.NewWorkspaceIndexFromGraph(opts.WorkspaceGraph)
	}

	projectMetadata := opts.ProjectMetadata
	if opts.WorkspaceGraph.Project != nil {
		projectMetadata.Name = opts.WorkspaceGraph.Project.Name
		projectMetadata.Description = opts.WorkspaceGraph.Project.Description
		projectMetadata.SourcePath = opts.WorkspaceGraph.Project.Path
		projectMetadata.SourceHash = opts.WorkspaceGraph.Project.SourceHash
		projectMetadata.Exists = true
	}

	if strings.TrimSpace(projectMetadata.Name) == "" {
		projectMetadata.Name = projectDisplayName(opts.ProjectDir)
	}

	if strings.TrimSpace(projectMetadata.Description) == "" {
		projectMetadata.Description = "Not set."
	}

	if strings.TrimSpace(projectMetadata.SourcePath) == "" && strings.TrimSpace(opts.ProjectDir) != "" {
		projectMetadata.SourcePath = chc.ProjectMetadataPath(opts.ProjectDir)
	}

	tickets := NewTicketViews(opts.Tickets)
	boardColumn := defaultBoardColumn(tickets)

	return Model{
		projectDir:               opts.ProjectDir,
		projectID:                opts.ProjectID,
		projectMetadata:          projectMetadata,
		cases:                    casesFromWorkspaceIndex(workspaceIndex, opts.WorkspaceGraph, opts.Cases),
		views:                    append([]View(nil), defaultViews...),
		maximizedPane:            -1,
		verticalSplitPercent:     defaultPanelSplitPercent,
		horizontalSplitPercent:   defaultPanelSplitPercent,
		tickets:                  tickets,
		agentWorkOrder:           chc.TicketLoopOrderCreatedAt,
		boardColumn:              boardColumn,
		previousBoardColumn:      boardColumn,
		documents:                append([]DocumentView(nil), opts.Documents...),
		dumps:                    normalizeDumpViews(opts.Dumps),
		quizzes:                  normalizeQuizViews(opts.Quizzes),
		specs:                    normalizeSpecViews(opts.Specs),
		plans:                    normalizePlanViews(opts.Plans),
		authoringSessions:        NewAuthoringSessionViews(opts.AuthoringSessions),
		observability:            opts.Observability,
		workspaceGraph:           opts.WorkspaceGraph,
		workspaceIndex:           workspaceIndex,
		semanticDigestStates:     append([]chc.WorkspaceSemanticDigestState(nil), opts.SemanticDigestStates...),
		commandPalette:           NewCommandPaletteState(DefaultCommandRegistry()),
		actionExecutor:           opts.ActionExecutor,
		dumpExecutor:             opts.DumpExecutor,
		quizExecutor:             opts.QuizExecutor,
		specExecutor:             opts.SpecExecutor,
		planExecutor:             opts.PlanExecutor,
		ticketExecutor:           opts.TicketExecutor,
		workExecutor:             opts.WorkExecutor,
		agentWorkRefreshExecutor: opts.AgentWorkExecutor,
		workspaceExecutor:        opts.WorkspaceExecutor,
		projectExecutor:          opts.ProjectExecutor,
		caseExecutor:             opts.CaseExecutor,
		sourceExecutor:           opts.SourceExecutor,
	}
}

func Run(ctx context.Context, opts Options) error {
	programOptions := []tea.ProgramOption{
		tea.WithContext(ctx),
		tea.WithOutput(opts.Stdout),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
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
	return agentWorkRefreshTickCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if updated, cmd, ok := m.updateSystemMessage(msg); ok {
		return updated, cmd
	}

	if typedMsg, ok := msg.(tea.KeyMsg); ok {
		return m.updateKey(typedMsg)
	}

	return m, nil
}

func (m Model) updateSystemMessage(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	if updated, cmd, ok := m.updateAgentWorkRefreshMessage(msg); ok {
		return updated, cmd, true
	}

	if updated, cmd, ok := m.updateActionFinishedMessage(msg); ok {
		return updated, cmd, true
	}

	switch typedMsg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typedMsg.Width
		m.height = typedMsg.Height

		return m, nil, true
	case tea.MouseMsg:
		updated, cmd := m.updateMouse(typedMsg)

		return updated, cmd, true
	case dumpEditorFinishedMsg:
		m.completeDumpEdit(typedMsg)

		return m, nil, true
	case sourceEditorFinishedMsg:
		m.completeSourceEdit(typedMsg)

		return m, nil, true
	default:
		return m, nil, false
	}
}

func (m Model) updateActionFinishedMessage(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	switch typedMsg := msg.(type) {
	case quizActionFinishedMsg:
		m.completeQuizAction(typedMsg)

		return m, nil, true
	case specActionFinishedMsg:
		m.completeSpecAction(typedMsg)

		return m, nil, true
	case planActionFinishedMsg:
		m.completePlanAction(typedMsg)

		return m, nil, true
	case workActionFinishedMsg:
		m.completeWorkAction(typedMsg)

		return m, nil, true
	default:
		return m, nil, false
	}
}

func (m Model) View() string {
	if m.width > 0 && m.height > 0 {
		return m.renderShell()
	}

	var output strings.Builder

	output.WriteString("Project\n")
	output.WriteString(fmt.Sprintf("  Name: %s\n", projectDisplayName(m.projectDir)))
	output.WriteString(fmt.Sprintf("  Dir:  %s\n\n", m.projectDir))
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
	if updated, cmd, handled := m.updateActiveModeKey(msg); handled {
		return updated, cmd
	}

	if cmd, handled := m.handleGlobalKey(msg); handled {
		return m, cmd
	}

	if m.SelectedView() == ViewProject {
		return m.updateProjectKey(msg)
	}

	if m.SelectedView() == ViewCases {
		return m.updateCasesKey(msg)
	}

	if m.SelectedView() == ViewDumps {
		return m.updateDumpsKey(msg)
	}

	if m.SelectedView() == ViewQuizzes {
		return m.updateQuizzesKey(msg)
	}

	if m.SelectedView() == ViewSpecs {
		return m.updateSpecsKey(msg)
	}

	if m.SelectedView() == ViewPlans {
		return m.updatePlansKey(msg)
	}

	return m.updateNavigationKey(msg)
}

//nolint:cyclop // Active mode dispatch is centralized to preserve modal ordering.
func (m Model) updateActiveModeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	switch {
	case m.artifactAction.Pending:
		updated, cmd := m.updateArtifactActionKey(msg)

		return updated, cmd, true
	case m.dumpAction.Pending:
		updated, cmd := m.updateDumpActionKey(msg)

		return updated, cmd, true
	case m.dumpEdit.Active:
		updated, cmd := m.updateDumpEditKey(msg)

		return updated, cmd, true
	case m.ticketAction.Pending:
		updated, cmd := m.updateTicketActionKey(msg)

		return updated, cmd, true
	case m.specAction.Pending:
		updated, cmd := m.updateSpecActionKey(msg)

		return updated, cmd, true
	case m.specBusy:
		return m, nil, true
	case m.planAction.Pending:
		updated, cmd := m.updatePlanActionKey(msg)

		return updated, cmd, true
	case m.planBusy:
		m.statusMessage = "Plan action already running."

		return m, nil, true
	case m.workBusy:
		m.statusMessage = "Work action already running."

		return m, nil, true
	case m.workAction.Pending:
		updated, cmd := m.updateWorkActionKey(msg)

		return updated, cmd, true
	case m.commandPalette.Open:
		updated, cmd := m.updateCommandPaletteKey(msg)

		return updated, cmd, true
	case m.filterEntryMode:
		updated, cmd := m.updateFilterKey(msg)

		return updated, cmd, true
	default:
		return m.updateActiveQuizModeKey(msg)
	}
}

func (m Model) updateActiveQuizModeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	switch {
	case m.quizAction.Pending:
		updated, cmd := m.updateQuizActionKey(msg)

		return updated, cmd, true
	case m.quizBusy:
		return m, nil, true
	case m.quizAnswer.Active:
		updated, cmd := m.updateQuizAnswerKey(msg)

		return updated, cmd, true
	default:
		return m, nil, false
	}
}

func (m *Model) handleGlobalKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "alt+m":
		m.toggleMaximizedPane()

		return nil, true
	case "alt+e":
		if m.SelectedView() != ViewDumps {
			return m.openSelectedSource(), true
		}

		m.planDumpAction(DumpActionExternalEdit)

		return m.confirmDumpAction(), true
	default:
		return nil, false
	}
}

func (m Model) updateMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	event := tea.MouseEvent(msg)

	if event.Action == tea.MouseActionRelease {
		m.layoutDrag = layoutDragNone

		return m, nil
	}

	if event.Action == tea.MouseActionMotion {
		return m.updateMouseMotion(event)
	}

	if event.Action != tea.MouseActionPress || event.Button != tea.MouseButtonLeft {
		return m, nil
	}

	width := maxInt(m.width, 60)
	height := maxInt(m.height, minShellHeight)

	if drag, ok := m.layoutDragAt(event.X, event.Y, width, height); ok {
		m.layoutDrag = drag
		m.applyLayoutDrag(event)

		return m, nil
	}

	button, ok := shellFooterButtonAt(event.X, event.Y, width, height)
	if !ok {
		if m.SelectedView() == ViewDumps {
			m.selectDumpPaneAt(event.X, width)
		}

		return m, nil
	}

	if button.View != "" {
		m.selectView(button.View)

		return m, nil
	}

	if button.Key == "F10" {
		return m, tea.Quit
	}

	return m, nil
}

func (m *Model) selectDumpPaneAt(x, width int) {
	if m.maximizedPane >= 0 {
		m.selectedPane = m.maximizedPane

		return
	}

	leftWidth, _ := m.splitWidths(width)
	if x <= leftWidth {
		m.selectedPane = 0

		return
	}

	m.selectedPane = 1
}

func (m Model) updateMouseMotion(event tea.MouseEvent) (tea.Model, tea.Cmd) {
	if m.layoutDrag != layoutDragNone {
		m.applyLayoutDrag(event)

		return m, nil
	}

	if event.Button != tea.MouseButtonLeft {
		return m, nil
	}

	width := maxInt(m.width, 60)
	height := maxInt(m.height, minShellHeight)

	if drag, ok := m.layoutDragAt(event.X, event.Y, width, height); ok {
		m.layoutDrag = drag
		m.applyLayoutDrag(event)
	}

	return m, nil
}

func (m Model) updateNavigationKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if view, ok := functionKeyView(msg); ok {
		m.selectView(view)

		return m, nil
	}

	if cmd, ok := m.handleNavigationCommandKey(key); ok {
		return m, cmd
	}

	if m.handleNavigationPageKey(msg.Type) {
		return m, nil
	}

	return m.handleNavigationActionKey(key)
}

func (m *Model) handleNavigationPageKey(keyType tea.KeyType) bool {
	switch keyType {
	case tea.KeyPgUp:
		m.scrollDetailPage(-1)
	case tea.KeyPgDown:
		m.scrollDetailPage(1)
	default:
		return false
	}

	return true
}

func (m Model) handleNavigationActionKey(key string) (tea.Model, tea.Cmd) {
	if m.handleDashActionKey(key) {
		return m, nil
	}

	if m.handleDetailScrollKey(key) {
		return m, nil
	}

	switch key {
	case "e":
		if m.selectedPane == 1 {
			return m, m.openSelectedSource()
		}
	case "up", "k":
		m.selectPreviousRecord()
	case "down", "j":
		m.selectNextRecord()
	case "esc":
		m.selectPrimaryPane()
	case ":":
		m.openCommandPalette()
	case "/":
		m.filterEntryMode = true
	}

	return m, nil
}

func (m *Model) handleDetailScrollKey(key string) bool {
	switch key {
	case "ctrl+u":
		m.scrollDetailPage(-1)
	case "ctrl+d":
		m.scrollDetailPage(1)
	default:
		return false
	}

	return true
}

func (m *Model) handleDashActionKey(key string) bool {
	if m.SelectedView() != ViewDash {
		return false
	}

	switch key {
	case "enter":
		m.planWorkAction(WorkActionStartTicket)
	case "a":
		m.planWorkAction(WorkActionRunAllReady)
	case "R":
		m.planWorkAction(WorkActionRunAllReady)
	case "p":
		m.cycleAgentWorkPriorityFilter()
	case "o":
		m.cycleAgentWorkOrder()
	case "s":
		m.planWorkAction(WorkActionRequestStop)
	default:
		return false
	}

	return true
}

func (m *Model) handleNavigationCommandKey(key string) (tea.Cmd, bool) {
	switch {
	case isQuitKey(key):
		return tea.Quit, true
	case key == "alt+m":
		m.toggleMaximizedPane()

		return nil, true
	case key == "ctrl+e":
		return m.openSelectedSource(), true
	case m.handlePaneNavigationKey(key):
		return nil, true
	case m.handleBoardColumnNavigationKey(key):
		return nil, true
	default:
		return nil, false
	}
}

func (m *Model) handleBoardColumnNavigationKey(key string) bool {
	switch key {
	case "alt+a":
		m.toggleBoardAttention()
	case "left", "h":
		m.selectPreviousBoardColumn()
	case "right", "l":
		m.selectNextBoardColumn()
	default:
		return false
	}

	return true
}

func (m *Model) toggleMaximizedPane() {
	if m.maximizedPane == m.selectedPane {
		m.maximizedPane = -1

		return
	}

	m.maximizedPane = m.selectedPane
}

func (m Model) layoutDragAt(x, y, width, height int) (layoutDrag, bool) {
	if m.maximizedPane >= 0 {
		return layoutDragNone, false
	}

	bodyTop, bodyHeight := m.shellBodyGeometry(width, height)
	if y < bodyTop || y >= bodyTop+bodyHeight {
		return layoutDragNone, false
	}

	leftWidth, _ := m.splitWidths(width)
	if nearInt(x, leftWidth, 4) {
		return layoutDragVertical, true
	}

	if !m.hasStackedRightPanels() || x < leftWidth {
		return layoutDragNone, false
	}

	topHeight, _ := m.splitHeights(bodyHeight)
	if nearInt(y-bodyTop, topHeight, 1) {
		return layoutDragHorizontal, true
	}

	return layoutDragNone, false
}

func (m *Model) applyLayoutDrag(event tea.MouseEvent) {
	width := maxInt(m.width, 60)
	height := maxInt(m.height, minShellHeight)
	bodyTop, bodyHeight := m.shellBodyGeometry(width, height)

	switch m.layoutDrag {
	case layoutDragVertical:
		m.verticalSplitPercent = percentFromPosition(event.X, width)
	case layoutDragHorizontal:
		m.horizontalSplitPercent = percentFromPosition(event.Y-bodyTop, bodyHeight)
	}
}

func (m Model) shellBodyGeometry(width, height int) (int, int) {
	topbar := m.renderShellTopbar(width)
	footer := m.renderShellFooter(width)
	status := m.renderShellStatus(width)

	bodyTop := lineCount(topbar)
	bodyHeight := height - lineCount(topbar) - lineCount(status) - lineCount(footer)

	if bodyHeight < 1 {
		bodyHeight = 1
	}

	return bodyTop, bodyHeight
}

func (m Model) hasStackedRightPanels() bool {
	switch m.SelectedView() {
	case ViewProject, ViewDashboard, ViewCases:
		return true
	default:
		return false
	}
}

func isQuitKey(key string) bool {
	return key == "ctrl+c" || key == "q" || key == "f10"
}

func (m *Model) handlePaneNavigationKey(key string) bool {
	switch key {
	case "shift+tab":
		m.selectPreviousPane()
	case "tab":
		m.selectNextPane()
	default:
		return false
	}

	return true
}

func functionKeyView(msg tea.KeyMsg) (View, bool) {
	viewsByType := map[tea.KeyType]View{
		tea.KeyF1: ViewProject,
		tea.KeyF2: ViewCases,
		tea.KeyF3: ViewDumps,
		tea.KeyF4: ViewQuizzes,
		tea.KeyF5: ViewSpecs,
		tea.KeyF6: ViewPlans,
		tea.KeyF7: ViewBoard,
		tea.KeyF8: ViewDash,
		tea.KeyF9: ViewHelp,
	}

	if view, ok := viewsByType[msg.Type]; ok {
		return view, true
	}

	viewsByString := map[string]View{
		"f1": ViewProject,
		"f2": ViewCases,
		"f3": ViewDumps,
		"f4": ViewQuizzes,
		"f5": ViewSpecs,
		"f6": ViewPlans,
		"f7": ViewBoard,
		"f8": ViewDash,
		"f9": ViewHelp,
	}

	view, ok := viewsByString[msg.String()]

	return view, ok
}

func (m *Model) selectPreviousRecord() {
	if m.selectedPane != 0 {
		m.scrollDetail(-1)

		return
	}

	selectedPath := m.selectedSourcePath()
	switch m.SelectedView() {
	case ViewBoard, ViewTickets:
		m.selectPreviousTicket()
	case ViewDumps:
		m.selectPreviousDump()
	case ViewQuizzes:
		m.selectPreviousQuiz()
	case ViewSpecs, ViewPlans, ViewDocuments:
		if m.SelectedView() == ViewSpecs {
			m.selectPreviousSpec()
		} else if m.SelectedView() == ViewPlans {
			m.selectPreviousPlan()
		} else {
			m.selectPreviousDocument()
		}
	case ViewAuthoring:
		m.selectPreviousAuthoringSession()
	}

	m.resetDetailScrollIfSourceChanged(selectedPath)
}

func (m *Model) selectNextRecord() {
	if m.selectedPane != 0 {
		m.scrollDetail(1)

		return
	}

	selectedPath := m.selectedSourcePath()
	switch m.SelectedView() {
	case ViewBoard, ViewTickets:
		m.selectNextTicket()
	case ViewDumps:
		m.selectNextDump()
	case ViewQuizzes:
		m.selectNextQuiz()
	case ViewSpecs, ViewPlans, ViewDocuments:
		if m.SelectedView() == ViewSpecs {
			m.selectNextSpec()
		} else if m.SelectedView() == ViewPlans {
			m.selectNextPlan()
		} else {
			m.selectNextDocument()
		}
	case ViewAuthoring:
		m.selectNextAuthoringSession()
	}

	m.resetDetailScrollIfSourceChanged(selectedPath)
}

func (m *Model) selectNextPane() {
	count := m.paneCount()
	if count <= 1 {
		m.selectedPane = 0

		return
	}

	m.selectedPane++
	if m.selectedPane >= count {
		m.selectedPane = 0
	}
}

func (m *Model) selectPreviousPane() {
	count := m.paneCount()
	if count <= 1 {
		m.selectedPane = 0

		return
	}

	if m.selectedPane == 0 {
		m.selectedPane = count - 1

		return
	}

	m.selectedPane--
}

func (m *Model) selectPrimaryPane() {
	m.selectedPane = 0
	m.detailScroll = 0
}

func (m Model) paneCount() int {
	switch m.SelectedView() {
	case ViewProject, ViewCases, ViewDash:
		return 3
	case ViewBoard, ViewTickets:
		return 2
	case ViewDumps, ViewQuizzes, ViewSpecs, ViewPlans, ViewDocuments:
		return 2
	case ViewAuthoring:
		return 2
	case ViewAgent, ViewRuntime:
		return 2
	default:
		return 1
	}
}

func (m Model) renderSelectedView() string {
	selectedView := m.SelectedView()

	if view, ok := m.renderProductSelectedView(selectedView); ok {
		return view
	}

	return m.renderLegacySelectedView(selectedView)
}

func (m Model) renderProductSelectedView(selectedView View) (string, bool) {
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

func (m Model) renderLegacySelectedView(selectedView View) string {
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

	var output strings.Builder

	output.WriteString("Board\n")
	output.WriteString(m.renderBoardHeaderColumnLine())
	output.WriteString(fmt.Sprintf("  Tickets: %d", len(filteredTickets)))

	if m.ticketFilter != "" {
		output.WriteString(fmt.Sprintf("  Filter: %q", m.ticketFilter))
	}

	output.WriteString("\n\n")

	if len(filteredTickets) == 0 {
		output.WriteString("No tickets in this column.\n\n")
		output.WriteString(m.renderSelectedTicketDetail())

		return strings.TrimRight(output.String(), "\n")
	}

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

func (m Model) renderBoardHeaderColumnLine() string {
	if m.boardAttentionActive() {
		return fmt.Sprintf(
			"  Attention (%d)\n",
			len(m.attentionTickets()),
		)
	}

	return fmt.Sprintf(
		"  %s (%d/%d)\n",
		chc.BoardColumnTitle(m.activeBoardColumn()),
		m.activeBoardColumnPosition()+1,
		len(chc.NormalBoardColumns),
	)
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
	if len(ticket.AttentionReasons) > 0 {
		return fmt.Sprintf(
			"! %-7s %-11s %-9s %-9s %s",
			ticket.Key,
			ticket.Status,
			ticket.Priority,
			ticket.Type,
			ticket.Title,
		)
	}

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

	if ticket.Synopsis != "" {
		output.WriteString(fmt.Sprintf("  Synopsis: %s\n", ticket.Synopsis))
	}

	output.WriteString(fmt.Sprintf("  Status: %s\n", ticket.Status))
	output.WriteString(fmt.Sprintf("  Priority: %s\n", ticket.Priority))
	output.WriteString(fmt.Sprintf("  Severity: %s\n", ticket.Severity))
	output.WriteString(fmt.Sprintf("  Type: %s\n", ticket.Type))
	output.WriteString(fmt.Sprintf("  Use Case: %s\n", ticket.UseCase))
	output.WriteString(fmt.Sprintf("  Project Area: %s\n", ticket.ProjectArea))
	renderTicketSourceDetail(&output, ticket)
	renderTicketReviewDetail(&output, ticket)
	output.WriteString(fmt.Sprintf("  Checklist: %d done / %d total\n", ticket.ChecklistDone, ticket.ChecklistTotal))
	output.WriteString(fmt.Sprintf("  Comments: %d\n", ticket.CommentCount))
	output.WriteString(fmt.Sprintf("  Path: %s\n", ticketDisplayPath(ticket.RelativePath)))
	renderTicketBodySection(&output, ticket.BodyLines)

	if len(ticket.AttentionReasons) > 0 {
		output.WriteString("  Attention:\n")

		for _, reason := range ticket.AttentionReasons {
			output.WriteString(fmt.Sprintf("    - %s\n", reason))
		}
	}

	return output.String()
}

func renderTicketReviewDetail(output *strings.Builder, ticket TicketView) {
	if !hasTicketReviewDetail(ticket) {
		return
	}

	renderTicketReviewStringDetail(output, "Branch", ticket.Branch)
	renderTicketReviewStringDetail(output, "Review State", ticket.ReviewState)

	if ticket.ReviewRound > 0 {
		fmt.Fprintf(output, "  Review Round: %d\n", ticket.ReviewRound)
	}

	renderTicketReviewStringDetail(output, "Review Result", ticket.ReviewResult)
	renderTicketReviewStringDetail(output, "Review Comments", ticket.ReviewCommentsURL)
}

func hasTicketReviewDetail(ticket TicketView) bool {
	return ticket.Branch != "" ||
		ticket.ReviewState != "" ||
		ticket.ReviewResult != "" ||
		ticket.ReviewCommentsURL != "" ||
		ticket.ReviewRound > 0
}

func renderTicketReviewStringDetail(output *strings.Builder, label, value string) {
	if value == "" {
		return
	}

	fmt.Fprintf(output, "  %s: %s\n", label, value)
}

func renderTicketSourceDetail(output *strings.Builder, ticket TicketView) {
	if ticket.SourceKind == "" && ticket.SourceRef == "" && ticket.Slice == "" {
		return
	}

	if ticket.SourceKind != "" {
		fmt.Fprintf(output, "  Source: %s\n", ticket.SourceKind)
	}

	if ticket.SourceRef != "" {
		fmt.Fprintf(output, "  Source Ref: %s\n", ticket.SourceRef)
	}

	if ticket.Slice != "" {
		fmt.Fprintf(output, "  Slice: %s\n", ticket.Slice)
	}
}

func renderTicketBodySection(output *strings.Builder, lines []string) {
	if len(lines) == 0 {
		return
	}

	output.WriteString("  Body:\n")

	for _, line := range lines {
		fmt.Fprintf(output, "    %s\n", line)
	}
}

//nolint:cyclop // Key helper text is centralized by selected view.
func (m Model) renderKeys() string {
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
		return "Keys: F1-F9 or click changes tab, tab changes pane, up/down selects focused pane, enter opens or edits, s saves, r refreshes, q quits"
	case m.SelectedView() == ViewDumps:
		return "Keys: F1-F9 or click changes tab, tab changes pane, up/down selects or scrolls, enter edits, alt+e external, c creates, r refreshes, q quits"
	case m.SelectedView() == ViewQuizzes:
		return "Keys: F1-F9 or click changes tab, tab changes pane, alt+e edits source, up/down selects or scrolls, s starts, enter answers, r refreshes, q quits"
	case m.SelectedView() == ViewSpecs:
		return "Keys: F1-F9 or click changes tab, tab changes pane, alt+e edits source, up/down selects or scrolls, g generates spec, r refreshes, q quits"
	case m.SelectedView() == ViewPlans:
		return "Keys: F1-F9 or click changes tab, tab changes pane, alt+e edits source, up/down selects or scrolls, g generates missing plan, G regenerates, r refreshes, q quits"
	case m.SelectedView() == ViewDash:
		return "Keys: F1-F9 or click changes tab, tab changes pane, a runs all ready tickets, enter starts selected ticket, s requests stop, q quits"
	case m.SelectedView() == ViewAuthoring:
		return "Keys: F1-F9 or click changes tab, tab changes pane, alt+e edits source, up/down selects or scrolls, q quits"
	default:
		return "Keys: F1-F9 or click changes tab, tab changes pane, q quits"
	}
}

func (m Model) ticketKeysActive() bool {
	return m.SelectedView() == ViewBoard || m.SelectedView() == ViewTickets
}

func (m Model) documentKeysActive() bool {
	switch m.SelectedView() {
	case ViewDocuments:
		return true
	default:
		return false
	}
}

//nolint:cyclop // Active interaction key hints mirror modal precedence.
func (m Model) activeInteractionKeys() (string, bool) {
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
	tickets := m.tickets
	if m.SelectedView() == ViewBoard {
		if m.boardAttentionActive() {
			tickets = m.attentionTickets()
		} else {
			tickets = m.activeBoardColumnTickets()
		}
	}

	if m.ticketFilter == "" {
		return append([]TicketView(nil), tickets...)
	}

	filter := strings.ToLower(m.ticketFilter)
	filtered := []TicketView{}

	for _, ticket := range tickets {
		if ticket.Matches(filter) {
			filtered = append(filtered, ticket)
		}
	}

	return filtered
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

func (m Model) activeBoardColumn() string {
	if m.boardAttentionActive() {
		return m.previousNormalBoardColumn()
	}

	if chc.IsNormalBoardTicketStatus(m.boardColumn) {
		return m.boardColumn
	}

	return chc.TicketStatusBacklog
}

func (m Model) activeBoardColumnPosition() int {
	for index, column := range chc.NormalBoardColumns {
		if column == m.activeBoardColumn() {
			return index
		}
	}

	return 0
}

func (m Model) activeBoardColumnTickets() []TicketView {
	return m.ticketsForBoardColumn(m.activeBoardColumn())
}

func (m Model) ticketsForBoardColumn(column string) []TicketView {
	tickets := []TicketView{}

	for _, ticket := range m.tickets {
		if ticket.BoardColumn == column {
			tickets = append(tickets, ticket)
		}
	}

	return tickets
}

func (m Model) attentionTickets() []TicketView {
	tickets := []TicketView{}

	for _, ticket := range m.tickets {
		if len(ticket.AttentionReasons) > 0 {
			tickets = append(tickets, ticket)
		}
	}

	return tickets
}

func (m Model) boardAttentionActive() bool {
	return m.SelectedView() == ViewBoard && m.boardColumn == boardAttentionColumn
}

func (m *Model) toggleBoardAttention() {
	if !m.boardColumnNavigationActive() {
		return
	}

	if m.boardAttentionActive() {
		m.boardColumn = m.previousNormalBoardColumn()
	} else {
		m.previousBoardColumn = m.activeBoardColumn()
		m.boardColumn = boardAttentionColumn
	}

	m.selectFirstFilteredTicket()
}

func (m Model) previousNormalBoardColumn() string {
	if chc.IsNormalBoardTicketStatus(m.previousBoardColumn) {
		return m.previousBoardColumn
	}

	return chc.TicketStatusBacklog
}

func (m *Model) selectPreviousBoardColumn() {
	if !m.boardColumnNavigationActive() {
		return
	}

	if m.boardAttentionActive() {
		return
	}

	position := m.activeBoardColumnPosition()
	if position == 0 {
		position = len(chc.NormalBoardColumns) - 1
	} else {
		position--
	}

	m.selectBoardColumn(chc.NormalBoardColumns[position])
}

func (m *Model) selectNextBoardColumn() {
	if !m.boardColumnNavigationActive() {
		return
	}

	if m.boardAttentionActive() {
		return
	}

	position := m.activeBoardColumnPosition() + 1
	if position >= len(chc.NormalBoardColumns) {
		position = 0
	}

	m.selectBoardColumn(chc.NormalBoardColumns[position])
}

func (m Model) boardColumnNavigationActive() bool {
	return m.SelectedView() == ViewBoard && m.selectedPane == 0
}

func (m *Model) selectBoardColumn(column string) {
	if !chc.IsNormalBoardTicketStatus(column) {
		return
	}

	m.boardColumn = column
	m.previousBoardColumn = column
	m.selectFirstFilteredTicket()
}

func defaultBoardColumn(tickets []TicketView) string {
	for _, column := range chc.NormalBoardColumns {
		for _, ticket := range tickets {
			if ticket.BoardColumn == column {
				return column
			}
		}
	}

	return chc.TicketStatusBacklog
}

func lastRune(value string) string {
	if value == "" {
		return ""
	}

	runes := []rune(value)

	return string(runes[len(runes)-1])
}

type TicketView struct {
	Index             int
	Key               string
	Title             string
	Synopsis          string
	Status            string
	Priority          string
	Severity          string
	Type              string
	UseCase           string
	ProjectArea       string
	SourceKind        string
	SourceRef         string
	Slice             string
	Branch            string
	ReviewState       string
	ReviewRound       int
	ReviewResult      string
	ReviewCommentsURL string
	BoardColumn       string
	AttentionReasons  []string
	RelativePath      string
	ChecklistDone     int
	ChecklistTotal    int
	CommentCount      int
	BodyLines         []string
}

func NewTicketViews(tickets []chc.IndexedTicket) []TicketView {
	views := make([]TicketView, 0, len(tickets))

	for index, ticket := range tickets {
		boardColumn, ok := chc.BoardTicketStatusFor(ticket.Status)
		if ok && chc.IsNormalBoardTicketStatus(ticket.BoardColumn) {
			boardColumn = ticket.BoardColumn
		}

		view := TicketView{
			Index:             index,
			Key:               ticket.Key,
			Title:             ticket.Title,
			Synopsis:          ticket.Synopsis,
			Status:            ticket.Status,
			Priority:          ticket.Priority,
			Severity:          ticket.Severity,
			Type:              ticket.Type,
			UseCase:           ticket.UseCase,
			ProjectArea:       ticket.ProjectArea,
			SourceKind:        ticket.SourceKind,
			SourceRef:         ticket.SourceRef,
			Slice:             ticket.Slice,
			Branch:            ticket.Branch,
			ReviewState:       ticket.ReviewState,
			ReviewRound:       ticket.ReviewRound,
			ReviewResult:      ticket.ReviewResult,
			ReviewCommentsURL: ticket.ReviewCommentsURL,
			BoardColumn:       boardColumn,
			AttentionReasons:  chc.BoardAttentionReasons(ticket),
			RelativePath:      ticket.RelativePath,
		}

		detail := ticketMarkdownDetail(ticket.Path)
		view.ChecklistDone = detail.ChecklistDone
		view.ChecklistTotal = detail.ChecklistTotal
		view.CommentCount = detail.CommentCount
		view.BodyLines = detail.BodyLines

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
		t.Synopsis,
		t.SourceKind,
		t.SourceRef,
		t.Slice,
		t.Branch,
		t.ReviewState,
		t.ReviewResult,
		t.ReviewCommentsURL,
		t.RelativePath,
	}

	for _, value := range values {
		if strings.Contains(strings.ToLower(value), filter) {
			return true
		}
	}

	return false
}

type ticketMarkdownDetailView struct {
	ChecklistDone  int
	ChecklistTotal int
	CommentCount   int
	BodyLines      []string
}

type ticketMarkdownParser struct {
	detail            ticketMarkdownDetailView
	inFrontMatter     bool
	frontMatterClosed bool
	inComments        bool
}

func ticketMarkdownDetail(path string) ticketMarkdownDetailView {
	content, err := os.ReadFile(path)
	if err != nil {
		return ticketMarkdownDetailView{}
	}

	parser := ticketMarkdownParser{}

	for _, line := range strings.Split(string(content), "\n") {
		parser.readLine(line)
	}

	return parser.detail
}

func (p *ticketMarkdownParser) readLine(line string) {
	trimmed := strings.TrimSpace(line)
	if p.handleFrontMatter(trimmed) || p.inFrontMatter {
		return
	}

	p.detail.BodyLines = append(p.detail.BodyLines, line)

	if trimmed == "" {
		return
	}

	switch {
	case strings.HasPrefix(trimmed, "# "):
		p.readHeading(trimmed)
	case strings.HasPrefix(trimmed, "- ["):
		p.readChecklistItem(trimmed)
	case p.inComments && strings.HasPrefix(trimmed, "## "):
		p.detail.CommentCount++
	}
}

func (p *ticketMarkdownParser) handleFrontMatter(trimmed string) bool {
	if trimmed != "---" || p.frontMatterClosed {
		return false
	}

	p.inFrontMatter = !p.inFrontMatter
	if !p.inFrontMatter {
		p.frontMatterClosed = true
	}

	return true
}

func (p *ticketMarkdownParser) readHeading(trimmed string) {
	p.inComments = trimmed == "# Comments"
}

func (p *ticketMarkdownParser) readChecklistItem(trimmed string) {
	p.detail.ChecklistTotal++

	if strings.HasPrefix(trimmed, "- [x]") || strings.HasPrefix(trimmed, "- [X]") {
		p.detail.ChecklistDone++
	}
}
