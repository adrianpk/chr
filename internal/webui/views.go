package webui

import "context"

// DataProvider loads read-only Web UI views from the application layer.
type DataProvider interface {
	ProjectView(context.Context) (ProjectView, error)
	BoardView(context.Context, BoardRequest) (BoardView, error)
	TicketDetailView(context.Context, TicketDetailRequest) (TicketDetailView, error)
	DashView(context.Context) (DashView, error)
	CatalogView(context.Context, string) (ArtifactCatalogView, error)
	ArtifactDetailView(context.Context, ArtifactDetailRequest) (ArtifactDetailView, error)
	WriteAction(context.Context, WriteActionRequest) (WriteActionResponse, error)
}

type staticDataProvider struct {
	shell ShellData
}

// PageData contains one rendered page and the shared shell state.
type PageData struct {
	Shell      ShellData
	View       string
	Title      string
	Revision   string
	RefreshURL string
	Project    ProjectView
	Board      BoardView
	Ticket     TicketDetailView
	Dash       DashView
	Catalog    ArtifactCatalogView
	Detail     ArtifactDetailView
	Write      WriteActionResponse
	CaseModal  bool
	DumpModal  bool
	Workflow   WorkflowView
}

func newStaticDataProvider(shell ShellData) staticDataProvider {
	return staticDataProvider{shell: shell}
}

func (p staticDataProvider) ProjectView(context.Context) (ProjectView, error) {
	return ProjectView{
		ProjectID:  p.shell.ProjectID,
		ProjectDir: p.shell.ProjectDir,
		Revision:   "static-project",
	}, nil
}

func (p staticDataProvider) BoardView(context.Context, BoardRequest) (BoardView, error) {
	return BoardView{ProjectID: p.shell.ProjectID, Revision: "static-board"}, nil
}

func (staticDataProvider) TicketDetailView(_ context.Context, request TicketDetailRequest) (TicketDetailView, error) {
	return TicketDetailView{
		Key:      request.Key,
		Title:    request.Key,
		Revision: "static-ticket-detail",
	}, nil
}

func (p staticDataProvider) DashView(context.Context) (DashView, error) {
	return DashView{Revision: "static-dash"}, nil
}

func (staticDataProvider) CatalogView(_ context.Context, family string) (ArtifactCatalogView, error) {
	return ArtifactCatalogView{
		Family:   family,
		Title:    CatalogTitle(family),
		Revision: "static-catalog",
	}, nil
}

func (staticDataProvider) ArtifactDetailView(_ context.Context, request ArtifactDetailRequest) (ArtifactDetailView, error) {
	return ArtifactDetailView{
		Family:   request.Family,
		ID:       request.ID,
		Title:    request.ID,
		Path:     request.ID,
		Revision: "static-detail",
	}, nil
}

func (staticDataProvider) WriteAction(_ context.Context, request WriteActionRequest) (WriteActionResponse, error) {
	return WriteActionResponse{
		Status:        WriteStatusBlocked,
		Summary:       "Write action is unavailable.",
		BlockedReason: request.Action + " is not supported by this provider.",
	}, nil
}

// ProjectView is the read-only project summary rendered by the Web UI.
type ProjectView struct {
	ProjectID  string
	ProjectDir string
	Revision   string
}

// BoardRequest selects one bounded page of canonical Board tickets.
type BoardRequest struct {
	Offset int
}

// BoardView is the read-only board projection rendered by the Web UI.
type BoardView struct {
	LoadedCount     int
	NextURL         string
	PreviousURL     string
	ProjectID       string
	TotalCount      int
	NormalCount     int
	AttentionCount  int
	Columns         []BoardColumnView
	Attention       []BoardAttentionView
	NormalColumnIDs []string
	Revision        string
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
	Revision          string
	Assignee          string
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

// TicketDetailRequest identifies one ticket detail request.
type TicketDetailRequest struct {
	Key string
	New bool
}

// TicketDetailView is the modal projection for one ticket.
type TicketDetailView struct {
	Groups           []TicketFieldGroupView
	CommentRequestID string
	PreviousRevision string
	New              bool
	Description      string
	Fields           map[string]string
	ReadOnly         map[string]string
	ReadOnlyReason   string
	Phase            string
	RequestID        string
	Cases            []TicketCaseView
	NativeWorkflow   bool
	WorkflowEnabled  bool
	Key              string
	Title            string
	Synopsis         string
	Priority         string
	Status           string
	Type             string
	ProjectArea      string
	UseCase          string
	Slice            string
	RelativePath     string
	Comments         []TicketCommentView
	Missing          bool
	Revision         string
}

// TicketFieldGroupView groups descriptive metadata in the shared editor.
type TicketFieldGroupView struct {
	Title  string
	Fields []TicketFieldView
}

// TicketFieldView describes an input and its persistence owner's constraints.
type TicketFieldView struct {
	Name, Label, Value, ReadOnly string
	Multiline                    bool
	Options                      []TicketOptionView
}

// TicketOptionView is a named selection with a canonical stored value.
type TicketOptionView struct {
	Value, Label string
}

// TicketCaseView is a named association offered by the shared case owner.
type TicketCaseView struct {
	Slug string
	Name string
}

// TicketCommentView is one rendered ticket comment.
type TicketCommentView struct {
	Author    string
	CreatedAt string
	Body      string
}

// DashView is the read-only dashboard projection rendered by the Web UI.
type DashView struct {
	Readiness       WorkReadinessView
	Queue           AgentQueueView
	Current         CurrentWorkView
	RecentEvents    []AgentEventView
	LatestLoop      LoopRunView
	LatestExecution ExecutionView
	Revision        string
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

// ArtifactDetailRequest identifies one read-only artifact detail request.
type ArtifactDetailRequest struct {
	Family string
	ID     string
}

// ArtifactCatalogView is the Web UI projection for one artifact family.
type ArtifactCatalogView struct {
	Family       string
	Title        string
	EmptyMessage string
	Items        []ArtifactCardView
	Revision     string
}

// ArtifactCardView is one catalog item card.
type ArtifactCardView struct {
	Family       string
	ID           string
	Title        string
	DisplayTitle string
	Status       string
	Scope        string
	Meta         string
	Path         string
	SourcePath   string
	UpdatedAt    string
	Preview      string
	DetailURL    string
	ContentState string
	Revision     string
}

// ArtifactDetailView is the read-only detail projection for one artifact.
type ArtifactDetailView struct {
	SaveRequestID, AppendRequestID string
	WorkflowEnabled                bool
	Family                         string
	ID                             string
	Title                          string
	Description                    string
	Context                        string
	Status                         string
	Scope                          string
	Path                           string
	SourcePath                     string
	UpdatedAt                      string
	Content                        string
	Missing                        bool
	Revision                       string
}

// WriteActionRequest describes one confirmed Web UI write action.
type WriteActionRequest struct {
	Action   string
	Family   string
	ID       string
	Revision string
	Confirm  bool
	Surface  string
	Fields   map[string]string
}

const (
	// WriteStatusOK marks a completed write action.
	WriteStatusOK = "ok"
	// WriteStatusValidation marks a user-correctable validation failure.
	WriteStatusValidation = "validation_failed"
	// WriteStatusStale marks a rejected stale write.
	WriteStatusStale = "stale"
	// WriteStatusBlocked marks an unsupported or unconfirmed operation.
	WriteStatusBlocked = "blocked"
)

// WriteActionResponse is the stable write result envelope returned by Web UI actions.
type WriteActionResponse struct {
	TicketRevision string
	Workflow       *WorkflowRunView
	Status         string
	Summary        string
	ChangedFiles   []string
	Validation     []string
	BlockedReason  string
	RedirectURL    string
	ReloadURL      string
	DiscardURL     string
	Operation      OperationStateView
}

// OperationStateView summarizes a Web UI-triggered operation result.
type OperationStateView struct {
	ID     string
	Kind   string
	Status string
	Detail string
}

// CatalogTitle returns the display title for an artifact family.
func CatalogTitle(family string) string {
	switch family {
	case "cases":
		return "Cases"
	case "dumps":
		return "Dumps"
	case "quizzes":
		return "Quizzes"
	case "specs":
		return "Specs"
	case "plans":
		return "Plans"
	default:
		return "Artifacts"
	}
}
