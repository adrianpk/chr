package webui

import "context"

// DataProvider loads read-only Web UI views from the application layer.
type DataProvider interface {
	ProjectView(context.Context) (ProjectView, error)
	BoardView(context.Context) (BoardView, error)
	DashView(context.Context) (DashView, error)
}

type staticDataProvider struct {
	shell ShellData
}

func newStaticDataProvider(shell ShellData) staticDataProvider {
	return staticDataProvider{shell: shell}
}

func (p staticDataProvider) ProjectView(context.Context) (ProjectView, error) {
	return ProjectView{
		ProjectID:      p.shell.ProjectID,
		ProjectDir:     p.shell.ProjectDir,
		URL:            p.shell.URL,
		Host:           p.shell.Host,
		Port:           p.shell.Port,
		RequireToken:   p.shell.RequireToken,
		RefreshSeconds: p.shell.RefreshSeconds,
		Paths: []ProjectPathView{
			{Label: "Project directory", Path: p.shell.ProjectDir},
		},
	}, nil
}

func (p staticDataProvider) BoardView(context.Context) (BoardView, error) {
	return BoardView{ProjectID: p.shell.ProjectID}, nil
}

func (p staticDataProvider) DashView(context.Context) (DashView, error) {
	return DashView{}, nil
}

// ProjectView is the read-only project summary rendered by the Web UI.
type ProjectView struct {
	ProjectID      string
	ProjectDir     string
	URL            string
	Host           string
	Port           int
	RequireToken   string
	RefreshSeconds int
	Paths          []ProjectPathView
}

// ProjectPathView describes one important project path.
type ProjectPathView struct {
	Label string
	Path  string
}

// BoardView is the read-only board projection rendered by the Web UI.
type BoardView struct {
	ProjectID       string
	TotalCount      int
	NormalCount     int
	AttentionCount  int
	Columns         []BoardColumnView
	Attention       []BoardAttentionView
	NormalColumnIDs []string
}

// BoardColumnView is one board column.
type BoardColumnView struct {
	ID      string
	Title   string
	Count   int
	Tickets []BoardTicketView
}

// BoardAttentionView is one attention ticket with its reasons.
type BoardAttentionView struct {
	Ticket  BoardTicketView
	Reasons []string
}

// BoardTicketView is the Web UI card projection of a ticket.
type BoardTicketView struct {
	Key               string
	Title             string
	Synopsis          string
	Status            string
	BoardColumn       string
	Priority          string
	Severity          string
	Type              string
	UseCase           string
	ProjectArea       string
	Slice             string
	Branch            string
	ReviewState       string
	ReviewRound       int
	ReviewResult      string
	ReviewCommentsURL string
	MergeState        string
	MergeResult       string
	MergeBlocker      string
	RelativePath      string
	SourceKind        string
	SourceRef         string
}

// DashView is the read-only dashboard projection rendered by the Web UI.
type DashView struct {
	Readiness       WorkReadinessView
	Queue           AgentQueueView
	Current         CurrentWorkView
	RecentEvents    []AgentEventView
	LatestLoop      LoopRunView
	LatestExecution ExecutionView
}

// AgentQueueView summarizes currently eligible work.
type AgentQueueView struct {
	EligibleCount int
	Tickets       []BoardTicketView
}

// CurrentWorkView summarizes the most recent active work signal.
type CurrentWorkView struct {
	Phase       string
	Message     string
	TicketKey   string
	ExecutionID string
	LoopID      string
}

// WorkReadinessView is the Web UI projection of work readiness state.
type WorkReadinessView struct {
	Status            string
	WorkspaceMode     string
	IntegrationBranch string
	CanonicalRemote   string
	CurrentBranch     string
	Diagnostics       []WorkReadinessDiagnosticView
}

// WorkReadinessDiagnosticView is one readiness diagnostic.
type WorkReadinessDiagnosticView struct {
	Code            string
	Severity        string
	Detail          string
	RepairAvailable bool
	RepairSafe      bool
	RepairDetail    string
}

// AgentEventView is one recent agent event row.
type AgentEventView struct {
	Time        string
	Level       string
	Phase       string
	TicketKey   string
	Message     string
	LoopID      string
	ExecutionID string
}

// LoopRunView summarizes the latest loop run.
type LoopRunView struct {
	ID             string
	Status         string
	UpdatedAt      string
	EligibleCount  int
	ProcessedCount int
	BlockedCount   int
	FailedCount    int
	FailureReason  string
}

// ExecutionView summarizes the latest execution.
type ExecutionView struct {
	ID            string
	TicketKey     string
	Status        string
	UpdatedAt     string
	FailureReason string
}
