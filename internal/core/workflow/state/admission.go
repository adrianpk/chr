package state

import (
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
)

// StartInput is the exact serializable native admission input retained by guided state.
// Host callbacks and commands belong to the caller, outside the durable value.
type StartInput struct {
	HarnessBinding      *harnesscore.Binding `json:"harness_binding,omitempty"`
	Admission           *AdmissionGuard      `json:"admission_guard,omitempty"`
	IncludeLaterTickets *bool                `json:"include_later_tickets_override,omitempty"`
	RequestKey          string
	Selection           PreviewRequest
	Policy              ExecutionPolicy
	ReviewRestriction   ReviewMode `json:"review_restriction,omitempty"`
}

// GuidedAdmission records authority provenance without extending native policy.
type GuidedAdmission struct {
	RunID, Actor, Policy, OperationDigest, CheckpointDigest, MaterializationID string
}

// ReviewMode defines a package boundary type.
type ReviewMode string

// TicketFilters describes a package boundary value.
type TicketFilters struct {
	Priorities   []string
	Severities   []string
	ProjectAreas []string
	Tags         []string
	Types        []string
	UseCases     []string
}

// TicketOrder defines a package boundary type.
type TicketOrder string

const (
	// TicketOrderKey defines a package boundary constant.
	TicketOrderKey TicketOrder = "key"
	// TicketOrderCreatedAt defines a package boundary constant.
	TicketOrderCreatedAt TicketOrder = "created_at"
	// TicketOrderPriority defines a package boundary constant.
	TicketOrderPriority TicketOrder = "priority"
	// TicketOrderSeverity defines a package boundary constant.
	TicketOrderSeverity TicketOrder = "severity"
	// TicketOrderPrioritySeverity defines a package boundary constant.
	TicketOrderPrioritySeverity TicketOrder = "priority_severity"
)

// AdmissionGuard retains exactly what a caller displayed before admission.
// It is part of the request identity; replay observes the original invocation.
type AdmissionGuard struct {
	Guided          *GuidedAdmission `json:"guided,omitempty"`
	ReadyTicket     *TicketAdmission `json:"ready_ticket,omitempty"`
	ExecutionPolicy ExecutionPolicy  `json:"execution_policy"`
	SnapshotHash    string           `json:"snapshot_hash"`
	Selected        *SelectionGuard  `json:"selected,omitempty"`
	Policy          ReviewPolicy     `json:"publication_policy"`
}

// ExecutionPolicy freezes finite execution limits at admission.
type ExecutionPolicy struct {
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

// Selector names an exact extent. Only loose work omits Identity.
type Selector struct {
	Extent   string
	Identity string
}

// PreviewRequest bounds a read-only selection of authoritative records.
type PreviewRequest struct {
	Selector         Selector
	Offset           int
	Limit            int
	SnapshotHash     string
	Target           PreviewTarget
	CumulativeBranch string
	CumulativeCommit string
	Filters          TicketFilters
	Statuses         []string
	Order            TicketOrder
	MinPriority      string
	MinSeverity      string
}

// PreviewTarget binds proof to an observed immutable canonical target.
type PreviewTarget struct {
	Repository string
	Commit     string
}

// ReviewPolicy identifies the current effective publication policy.
type ReviewPolicy struct {
	Mode     ReviewMode `json:"mode"`
	Method   string     `json:"method"`
	Source   string     `json:"source"`
	Revision string     `json:"revision"`
}

// SelectionGuard binds admission to the selected native ticket image.
// It is admission input, not a retained credential or an execution override.
type SelectionGuard struct {
	Key      string
	Revision uint64
}

// TicketAdmission binds the explicit transition of one independent
// backlog ticket to ready, together with its existing executable contract.
type TicketAdmission struct {
	Key          string     `json:"key"`
	Revision     uint64     `json:"revision"`
	ContractHash string     `json:"contract_hash"`
	Validation   Validation `json:"validation"`
}

// ReviewAgentic defines a package boundary constant.
const ReviewAgentic ReviewMode = "agentic_review"

// ReviewHuman defines a package boundary constant.
const ReviewHuman ReviewMode = "human_review"

// ReviewTrustedAutoMerge defines a package boundary constant.
const ReviewTrustedAutoMerge ReviewMode = "trusted_auto_merge"

// DefaultReviewMode defines a package boundary constant.
const DefaultReviewMode = ReviewHuman
