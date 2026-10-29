package tui

import (
	"context"
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// View identifies one TUI screen.
type View string

// Primary view identifiers define the shell tabs shown by default.
const (
	ViewProject View = "Project"
	ViewCases   View = "Cases"
	ViewDumps   View = "Dumps"
	ViewQuizzes View = "Quizzes"
	ViewSpecs   View = "Specs"
	ViewPlans   View = "Plans"
	ViewDash    View = "Dash"
	ViewHelp    View = "Help"
)

// Legacy view identifiers are kept for existing commands and non-shell
// rendering paths.
const (
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

// Options configures the initial TUI model and its service adapters.
type Options struct {
	ProjectDir           string
	ProjectID            string
	ProjectMetadata      authoringcore.ProjectMetadata
	Cases                []authoringcore.CaseMetadata
	Stdin                io.Reader
	Stdout               io.Writer
	Tickets              []ticketcore.Indexed
	Documents            []DocumentView
	Dumps                []DumpView
	Quizzes              []QuizView
	Specs                []SpecView
	Plans                []PlanView
	AuthoringSessions    []authoringcore.Session
	Observability        ObservabilityView
	WorkReadiness        WorkReadinessView
	WorkspaceGraph       core.WorkspaceGraph
	WorkspaceIndex       core.WorkspaceIndex
	SemanticDigestStates []core.WorkspaceSemanticDigestState
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
	HarnessExecutor      HarnessSelectionExecutor
	GuidedExecutor       GuidedExecutor
}

// Model is the Bubble Tea model for the chr TUI.
type Model struct {
	projectState
	caseState
	navigationState
	layoutState
	ticketState
	documentState
	dumpState
	quizState
	specState
	planState
	authoringState
	workspaceState
	interactionState
	workState
	backgroundWorkState
	harnessState
	guidedModelState
}

// NewModel initializes a terminal UI model from loaded workspace state.
//
//nolint:funlen // Model construction keeps option mapping in one place.
func NewModel(opts Options) *Model {
	workspaceIndex := opts.WorkspaceIndex
	if len(workspaceIndex.NodesByID) == 0 {
		workspaceIndex = core.NewWorkspaceIndexFromGraph(opts.WorkspaceGraph)
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
		projectMetadata.SourcePath = authoringcore.ProjectMetadataPath(opts.ProjectDir)
	}

	tickets := NewTicketViews(opts.Tickets)
	boardColumn := defaultBoardColumn(tickets)

	model := &Model{
		projectState: projectState{
			projectDir:      opts.ProjectDir,
			projectID:       opts.ProjectID,
			projectMetadata: projectMetadata,
			projectExecutor: opts.ProjectExecutor,
		},
		caseState: caseState{
			cases:        casesFromWorkspaceIndex(workspaceIndex, opts.WorkspaceGraph, opts.Cases),
			caseExecutor: opts.CaseExecutor,
		},
		navigationState: navigationState{
			views: append([]View(nil), defaultViews...),
		},
		layoutState: layoutState{
			maximizedPane:          -1,
			verticalSplitPercent:   defaultPanelSplitPercent,
			horizontalSplitPercent: defaultPanelSplitPercent,
		},
		ticketState: ticketState{
			tickets:             tickets,
			boardColumn:         boardColumn,
			previousBoardColumn: boardColumn,
			ticketExecutor:      opts.TicketExecutor,
		},
		documentState: documentState{
			documents: append([]DocumentView(nil), opts.Documents...),
		},
		dumpState: dumpState{
			dumps:        normalizeDumpViews(opts.Dumps),
			dumpExecutor: opts.DumpExecutor,
		},
		quizState: quizState{
			quizzes:      normalizeQuizViews(opts.Quizzes),
			quizExecutor: opts.QuizExecutor,
		},
		specState: specState{
			specs:        normalizeSpecViews(opts.Specs),
			specExecutor: opts.SpecExecutor,
		},
		planState: planState{
			plans:        normalizePlanViews(opts.Plans),
			planExecutor: opts.PlanExecutor,
		},
		authoringState: authoringState{
			authoringSessions: NewAuthoringSessionViews(opts.AuthoringSessions),
		},
		workspaceState: workspaceState{
			workspaceGraph:       opts.WorkspaceGraph,
			workspaceIndex:       workspaceIndex,
			semanticDigestStates: append([]core.WorkspaceSemanticDigestState(nil), opts.SemanticDigestStates...),
			workspaceExecutor:    opts.WorkspaceExecutor,
		},
		interactionState: interactionState{
			commandPalette: NewCommandPaletteState(DefaultCommandRegistry()),
			actionExecutor: opts.ActionExecutor,
			sourceExecutor: opts.SourceExecutor,
		},
		workState: workState{
			observability:  opts.Observability,
			workReadiness:  opts.WorkReadiness,
			agentWorkOrder: workflowstate.TicketOrderCreatedAt,
			workExecutor:   opts.WorkExecutor,
		},
		backgroundWorkState: backgroundWorkState{
			agentWorkRefreshExecutor: opts.AgentWorkExecutor,
		},
		harnessState: harnessState{
			harnessExecutor: opts.HarnessExecutor,
		},
		guidedModelState: guidedModelState{
			guidedExecutor: opts.GuidedExecutor,
		},
	}
	model.reconcileDashSelection()

	return model
}

// Run starts the terminal UI event loop.
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

var _ tea.Model = (*Model)(nil)

// Init satisfies tea.Model and returns the initial command.
func (m *Model) Init() tea.Cmd {
	return agentWorkRefreshTickCmd()
}

// Update applies a Bubble Tea message to the event-loop owner.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if outcome := m.updateSystemMessage(msg); outcome.Handled {
		return m, outcome.Command
	}

	if typedMsg, ok := msg.(tea.KeyMsg); ok {
		return m, m.updateKey(typedMsg)
	}

	return m, nil
}
