package core

import (
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowIntegrationTreeRepair records one host-verified transformation at an
// exact original source step. It supplements method proof, never approval.
type WorkflowIntegrationTreeRepair struct {
	Input          WorkflowMergerInput                  `json:"input"`
	Proposal       WorkflowMergerResult                 `json:"proposal"`
	CompletedAt    time.Time                            `json:"completed_at"`
	OriginalParent string                               `json:"original_parent"`
	PreparedParent string                               `json:"prepared_parent"`
	Source         string                               `json:"source"`
	BaselineTree   string                               `json:"baseline_tree"`
	ProposalTree   string                               `json:"proposal_tree,omitempty"`
	ResolvedTree   string                               `json:"resolved_tree"`
	Regeneration   []WorkflowMergerRegenerationEvidence `json:"regeneration,omitempty"`
}

// WorkflowMergerRegenerationEvidence is observed by the host after diff scope
// acceptance. Model-reported local checks cannot supply this evidence.
type WorkflowMergerRegenerationEvidence struct {
	ConflictID string                `json:"conflict_id"`
	Command    WorkflowCommandResult `json:"command"`
	Validation WorkflowCommandResult `json:"validation"`
	InputTree  string                `json:"input_tree"`
	OutputTree string                `json:"output_tree"`
}

// WorkflowMergerRepairState is a retained private Git checkpoint. A conflicted
// tree is available without publishing or altering the source checkout.
type WorkflowMergerRepairState struct {
	SchemaVersion  int                             `json:"schema_version"`
	Query          WorkflowPreparationQuery        `json:"query"`
	Preparation    WorkflowIntegrationPreparation  `json:"preparation"`
	ReplayIndex    int                             `json:"replay_index"`
	OriginalParent string                          `json:"original_parent"`
	PreparedParent string                          `json:"prepared_parent"`
	CurrentSource  string                          `json:"current_source"`
	Tree           string                          `json:"tree"`
	Conflicts      []string                        `json:"conflicts"`
	History        []WorkflowIntegrationCommit     `json:"history"`
	Repairs        []WorkflowIntegrationTreeRepair `json:"repairs,omitempty"`
	Result         *WorkflowIntegrationProof       `json:"result,omitempty"`
}

// WorkflowMergerRepairPlan selects an original replay owner or an explicit
// integration fix. Ambiguity produces decisions, not implicit permission.
type WorkflowMergerRepairPlan struct {
	InputHash      string                       `json:"input_hash"`
	Outcome        string                       `json:"outcome"`
	OriginalCommit string                       `json:"original_commit,omitempty"`
	IntegrationFix *workflowstate.ReviewFinding `json:"integration_fix,omitempty"`
	Decisions      []WorkflowMergerDecision     `json:"decisions,omitempty"`
}
