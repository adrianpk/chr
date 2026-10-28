package workflow

import (
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// MergerRequirement binds required source or target intent to evidence.
type MergerRequirement struct {
	ID            string                    `json:"id"`
	Origin        string                    `json:"origin"`
	Evidence      workflowstate.EvidenceRef `json:"evidence"`
	ConflictsWith []string                  `json:"conflicts_with,omitempty"`
}

// MergerConflict describes a single actionable conflict. Regeneration
// is permitted only through the repository-declared command, never a role edit.
type MergerConflict struct {
	ID                  string                    `json:"id"`
	Path                string                    `json:"path"`
	Kind                string                    `json:"kind"`
	Evidence            workflowstate.EvidenceRef `json:"evidence"`
	RequirementIDs      []string                  `json:"requirement_ids"`
	RegenerationCommand []string                  `json:"regeneration_command,omitempty"`
	ValidationCommand   []string                  `json:"validation_command,omitempty"`
}

// MergerCheckpoint names the original replay step and every predecessor.
// A nil checkpoint means combined-tree repair, not a collapsed rebase history.
type MergerCheckpoint struct {
	ReplayIndex    int                       `json:"replay_index"`
	OriginalCommit string                    `json:"original_commit"`
	PreparedParent string                    `json:"prepared_parent"`
	Predecessors   []IntegrationReplay       `json:"predecessors"`
	Evidence       workflowstate.EvidenceRef `json:"evidence"`
}

// MergerInput is an immutable, credential-free contract for one role
// call. The opaque workspace identity is resolved by the host, never the agent.
type MergerInput struct {
	MergerOperator
	MergerIdentity
	MergerGit
	MergerContext
	MergerScope
	MergerValidation
	MergerLimits
}

// MergerOperator groups the retained operator fields without changing their wire order.
type MergerOperator struct {
	OperatorDecisionRef *workflowstate.EvidenceRef        `json:"operator_decision_ref,omitempty"`
	OperatorAnswers     []workflowstate.IntegrationAnswer `json:"operator_answers,omitempty"`
}

// MergerIdentity groups the retained identity fields without changing their wire order.
type MergerIdentity struct {
	ProtocolVersion   int               `json:"protocol_version"`
	ID                string            `json:"id"`
	InvocationID      string            `json:"invocation_id"`
	EntryID           string            `json:"entry_id"`
	Unit              workflowstate.Ref `json:"unit"`
	CandidateRevision uint64            `json:"candidate_revision"`
	CandidateHash     string            `json:"candidate_hash"`
	ClaimGeneration   uint64            `json:"claim_generation"`
	InputHash         string            `json:"input_hash"`
	ScopeHash         string            `json:"scope_hash"`
	Reason            string            `json:"reason"`
}

// MergerGit groups the retained git fields without changing their wire order.
type MergerGit struct {
	Repository       string `json:"repository"`
	SourceRef        string `json:"source_ref"`
	TargetRef        string `json:"target_ref"`
	Source           string `json:"source"`
	Base             string `json:"base"`
	Method           string `json:"method"`
	OwnedWorkspaceID string `json:"owned_workspace_id"`
}

// MergerContext groups the retained context fields without changing their wire order.
type MergerContext struct {
	SourceManifestRef   workflowstate.EvidenceRef   `json:"source_manifest_ref"`
	IntentRefs          []workflowstate.EvidenceRef `json:"intent_refs"`
	TargetChangeRefs    []workflowstate.EvidenceRef `json:"target_change_refs"`
	FindingRefs         []workflowstate.EvidenceRef `json:"finding_refs"`
	Requirements        []MergerRequirement         `json:"requirements"`
	Conflicts           []MergerConflict            `json:"conflicts"`
	HistoryCheckpoint   *MergerCheckpoint           `json:"history_checkpoint,omitempty"`
	RepairCheckpointRef *workflowstate.EvidenceRef  `json:"repair_checkpoint_ref,omitempty"`
}

// MergerScope groups the retained scope fields without changing their wire order.
type MergerScope struct {
	AllowedPaths   []string `json:"allowed_paths"`
	ProtectedPaths []string `json:"protected_paths"`
}

// MergerValidation groups the retained validation fields without changing their wire order.
type MergerValidation struct {
	ValidationContractRef workflowstate.EvidenceRef `json:"validation_contract_ref"`
	ValidationCommands    [][]string                `json:"validation_commands"`
}

// MergerLimits groups the retained limits fields without changing their wire order.
type MergerLimits struct {
	CapacityReservationID string          `json:"capacity_reservation_id"`
	BudgetReservationID   string          `json:"budget_reservation_id"`
	RemainingAttempts     int             `json:"remaining_attempts"`
	WorkerTimeoutSeconds  int             `json:"worker_timeout_seconds"`
	AdmittedAt            time.Time       `json:"admitted_at"`
	Deadline              time.Time       `json:"deadline"`
	TokenAllowance        *TokenAllowance `json:"token_allowance,omitempty"`
}

// MergerResolution explains one conflict and the requirements retained.
// Local validation observations do not authorize the repaired candidate.
type MergerResolution struct {
	ConflictID          string   `json:"conflict_id"`
	Path                string   `json:"path"`
	Rationale           string   `json:"rationale"`
	RequirementIDs      []string `json:"requirement_ids"`
	RegenerationCommand []string `json:"regeneration_command,omitempty"`
}

// MergerDecision requests a bounded product or ownership choice.
// Its request is tied to InputHash; elapsed time cannot answer it.
type MergerDecision struct {
	ID             string   `json:"id"`
	Issue          string   `json:"issue"`
	RequirementIDs []string `json:"requirement_ids"`
	Options        []string `json:"options"`
	Authority      string   `json:"authority"`
}

// MergerUsage is supplementary whole-call accounting from the harness.
// Nil usage remains unknown and cannot release a finite token reservation.
type MergerUsage struct {
	Input              int64  `json:"input"`
	Output             int64  `json:"output"`
	CapabilityRevision string `json:"capability_revision"`
	WholeCall          bool   `json:"whole_call"`
}

// MergerResult is an untrusted proposal. It contains no approval,
// receipt, method override, publication grant or process-control authority.
type MergerResult struct {
	ProtocolVersion       int                         `json:"protocol_version"`
	InputHash             string                      `json:"input_hash"`
	Outcome               string                      `json:"outcome"`
	ChangedPaths          []string                    `json:"changed_paths"`
	ResolutionNotes       []MergerResolution          `json:"resolution_notes"`
	PreservedRequirements []string                    `json:"preserved_requirements"`
	LocalCheckRefs        []workflowstate.EvidenceRef `json:"local_check_refs"`
	DecisionRequests      []MergerDecision            `json:"decision_requests"`
	Failure               string                      `json:"failure,omitempty"`
	Usage                 *MergerUsage                `json:"usage,omitempty"`
}
