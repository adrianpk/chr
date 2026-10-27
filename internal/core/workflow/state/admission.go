package state

// GuidedWorkflowAdmission records authority provenance without extending native policy.
type GuidedWorkflowAdmission struct {
	RunID, Actor, Policy, OperationDigest, CheckpointDigest, MaterializationID string
}

// ReviewWorkflowMode defines a package boundary type.
type ReviewWorkflowMode string

// TicketLoopFilters describes a package boundary value.
type TicketLoopFilters struct {
	Priorities   []string
	Severities   []string
	ProjectAreas []string
	Tags         []string
	Types        []string
	UseCases     []string
}

// TicketLoopOrder defines a package boundary type.
type TicketLoopOrder string

const (
	// TicketLoopOrderKey defines a package boundary constant.
	TicketLoopOrderKey TicketLoopOrder = "key"
	// TicketLoopOrderCreatedAt defines a package boundary constant.
	TicketLoopOrderCreatedAt TicketLoopOrder = "created_at"
	// TicketLoopOrderPriority defines a package boundary constant.
	TicketLoopOrderPriority TicketLoopOrder = "priority"
	// TicketLoopOrderSeverity defines a package boundary constant.
	TicketLoopOrderSeverity TicketLoopOrder = "severity"
	// TicketLoopOrderPrioritySeverity defines a package boundary constant.
	TicketLoopOrderPrioritySeverity TicketLoopOrder = "priority_severity"
)

// WorkflowAdmissionGuard retains exactly what a caller displayed before admission.
// It is part of the request identity; replay observes the original invocation.
type WorkflowAdmissionGuard struct {
	Guided          *GuidedWorkflowAdmission `json:"guided,omitempty"`
	ReadyTicket     *WorkflowTicketAdmission `json:"ready_ticket,omitempty"`
	ExecutionPolicy WorkflowExecutionPolicy  `json:"execution_policy"`
	SnapshotHash    string                   `json:"snapshot_hash"`
	Selected        *WorkflowSelectionGuard  `json:"selected,omitempty"`
	Policy          WorkflowReviewPolicy     `json:"publication_policy"`
}

// WorkflowExecutionPolicy freezes finite execution limits at admission.
type WorkflowExecutionPolicy struct {
	AgentTokenBudget         int64  `json:"agent_token_budget,omitempty"`
	ConfigRevision           string `json:"config_revision,omitempty"`
	ValidationTimeoutSeconds int    `json:"validation_timeout_seconds,omitempty"`
	IncludeLaterTickets      bool   `json:"include_later_tickets,omitempty"`
	WorkerTimeoutSeconds     int    `json:"worker_timeout_seconds"`
	MaxWorkerAttempts        int    `json:"max_worker_attempts"`
	MaxIntegrationAttempts   int    `json:"max_integration_attempts,omitempty"`
	MaxMergerAttempts        int    `json:"max_merger_attempts,omitempty"`
	MaxDriftRevisions        int    `json:"max_drift_revisions,omitempty"`
	CheckpointStepLimit      int    `json:"checkpoint_step_limit,omitempty"`
}

// WorkflowSelector names an exact extent. Only loose work omits Identity.
type WorkflowSelector struct {
	Extent   string
	Identity string
}

// WorkflowPreviewRequest bounds a read-only selection of authoritative records.
type WorkflowPreviewRequest struct {
	Selector         WorkflowSelector
	Offset           int
	Limit            int
	SnapshotHash     string
	Target           WorkflowPreviewTarget
	CumulativeBranch string
	CumulativeCommit string
	Filters          TicketLoopFilters
	Statuses         []string
	Order            TicketLoopOrder
	MinPriority      string
	MinSeverity      string
}

// WorkflowPreviewTarget binds proof to an observed immutable canonical target.
type WorkflowPreviewTarget struct {
	Repository string
	Commit     string
}

// WorkflowReviewPolicy identifies the current effective publication policy.
type WorkflowReviewPolicy struct {
	Mode     ReviewWorkflowMode `json:"mode"`
	Method   string             `json:"method"`
	Source   string             `json:"source"`
	Revision string             `json:"revision"`
}

// WorkflowSelectionGuard binds admission to the selected native ticket image.
// It is admission input, not a retained credential or an execution override.
type WorkflowSelectionGuard struct {
	Key      string
	Revision uint64
}

// WorkflowTicketAdmission binds the explicit transition of one independent
// backlog ticket to ready, together with its existing executable contract.
type WorkflowTicketAdmission struct {
	Key          string     `json:"key"`
	Revision     uint64     `json:"revision"`
	ContractHash string     `json:"contract_hash"`
	Validation   Validation `json:"validation"`
}

// ReviewWorkflowModeAgenticReview defines a package boundary constant.
const ReviewWorkflowModeAgenticReview ReviewWorkflowMode = "agentic_review"

// ReviewWorkflowModeHumanReview defines a package boundary constant.
const ReviewWorkflowModeHumanReview ReviewWorkflowMode = "human_review"

// ReviewWorkflowModeTrustedAutoMerge defines a package boundary constant.
const ReviewWorkflowModeTrustedAutoMerge ReviewWorkflowMode = "trusted_auto_merge"

// DefaultReviewWorkflowMode defines a package boundary constant.
const DefaultReviewWorkflowMode = ReviewWorkflowModeHumanReview
