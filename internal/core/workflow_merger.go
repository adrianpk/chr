package core

import "time"

// WorkflowMergerRequirement binds required source or target intent to evidence.
type WorkflowMergerRequirement struct {
	ID            string              `json:"id"`
	Origin        string              `json:"origin"`
	Evidence      WorkflowEvidenceRef `json:"evidence"`
	ConflictsWith []string            `json:"conflicts_with,omitempty"`
}

// WorkflowMergerConflict describes a single actionable conflict. Regeneration
// is permitted only through the repository-declared command, never a role edit.
type WorkflowMergerConflict struct {
	ID                  string              `json:"id"`
	Path                string              `json:"path"`
	Kind                string              `json:"kind"`
	Evidence            WorkflowEvidenceRef `json:"evidence"`
	RequirementIDs      []string            `json:"requirement_ids"`
	RegenerationCommand []string            `json:"regeneration_command,omitempty"`
	ValidationCommand   []string            `json:"validation_command,omitempty"`
}

// WorkflowMergerCheckpoint names the original replay step and every predecessor.
// A nil checkpoint means combined-tree repair, not a collapsed rebase history.
type WorkflowMergerCheckpoint struct {
	ReplayIndex    int                         `json:"replay_index"`
	OriginalCommit string                      `json:"original_commit"`
	PreparedParent string                      `json:"prepared_parent"`
	Predecessors   []WorkflowIntegrationReplay `json:"predecessors"`
	Evidence       WorkflowEvidenceRef         `json:"evidence"`
}

// WorkflowMergerInput is an immutable, credential-free contract for one role
// call. The opaque workspace identity is resolved by the host, never the agent.
type WorkflowMergerInput struct {
	ProtocolVersion       int                         `json:"protocol_version"`
	ID                    string                      `json:"id"`
	InvocationID          string                      `json:"invocation_id"`
	EntryID               string                      `json:"entry_id"`
	Unit                  WorkflowRef                 `json:"unit"`
	CandidateRevision     uint64                      `json:"candidate_revision"`
	CandidateHash         string                      `json:"candidate_hash"`
	ClaimGeneration       uint64                      `json:"claim_generation"`
	InputHash             string                      `json:"input_hash"`
	ScopeHash             string                      `json:"scope_hash"`
	Reason                string                      `json:"reason"`
	Repository            string                      `json:"repository"`
	SourceRef             string                      `json:"source_ref"`
	TargetRef             string                      `json:"target_ref"`
	Source                string                      `json:"source"`
	Base                  string                      `json:"base"`
	Method                string                      `json:"method"`
	OwnedWorkspaceID      string                      `json:"owned_workspace_id"`
	SourceManifestRef     WorkflowEvidenceRef         `json:"source_manifest_ref"`
	IntentRefs            []WorkflowEvidenceRef       `json:"intent_refs"`
	TargetChangeRefs      []WorkflowEvidenceRef       `json:"target_change_refs"`
	FindingRefs           []WorkflowEvidenceRef       `json:"finding_refs"`
	Requirements          []WorkflowMergerRequirement `json:"requirements"`
	Conflicts             []WorkflowMergerConflict    `json:"conflicts"`
	HistoryCheckpoint     *WorkflowMergerCheckpoint   `json:"history_checkpoint,omitempty"`
	AllowedPaths          []string                    `json:"allowed_paths"`
	ProtectedPaths        []string                    `json:"protected_paths"`
	ValidationContractRef WorkflowEvidenceRef         `json:"validation_contract_ref"`
	ValidationCommands    [][]string                  `json:"validation_commands"`
	CapacityReservationID string                      `json:"capacity_reservation_id"`
	BudgetReservationID   string                      `json:"budget_reservation_id"`
	RemainingAttempts     int                         `json:"remaining_attempts"`
	WorkerTimeoutSeconds  int                         `json:"worker_timeout_seconds"`
	AdmittedAt            time.Time                   `json:"admitted_at"`
	Deadline              time.Time                   `json:"deadline"`
	TokenAllowance        *WorkflowTokenAllowance     `json:"token_allowance,omitempty"`
}

// WorkflowMergerResolution explains one conflict and the requirements retained.
// Local validation observations do not authorize the repaired candidate.
type WorkflowMergerResolution struct {
	ConflictID          string   `json:"conflict_id"`
	Path                string   `json:"path"`
	Rationale           string   `json:"rationale"`
	RequirementIDs      []string `json:"requirement_ids"`
	RegenerationCommand []string `json:"regeneration_command,omitempty"`
}

// WorkflowMergerDecision requests a bounded product or ownership choice.
// Its request is tied to InputHash; elapsed time cannot answer it.
type WorkflowMergerDecision struct {
	ID             string   `json:"id"`
	Issue          string   `json:"issue"`
	RequirementIDs []string `json:"requirement_ids"`
	Options        []string `json:"options"`
	Authority      string   `json:"authority"`
}

// WorkflowMergerUsage is supplementary whole-call accounting from the harness.
// Nil usage remains unknown and cannot release a finite token reservation.
type WorkflowMergerUsage struct {
	Input              int64  `json:"input"`
	Output             int64  `json:"output"`
	CapabilityRevision string `json:"capability_revision"`
	WholeCall          bool   `json:"whole_call"`
}

// WorkflowMergerResult is an untrusted proposal. It contains no approval,
// receipt, method override, publication grant or process-control authority.
type WorkflowMergerResult struct {
	ProtocolVersion       int                        `json:"protocol_version"`
	InputHash             string                     `json:"input_hash"`
	Outcome               string                     `json:"outcome"`
	ChangedPaths          []string                   `json:"changed_paths"`
	ResolutionNotes       []WorkflowMergerResolution `json:"resolution_notes"`
	PreservedRequirements []string                   `json:"preserved_requirements"`
	LocalCheckRefs        []WorkflowEvidenceRef      `json:"local_check_refs"`
	DecisionRequests      []WorkflowMergerDecision   `json:"decision_requests"`
	Failure               string                     `json:"failure,omitempty"`
	Usage                 *WorkflowMergerUsage       `json:"usage,omitempty"`
}
