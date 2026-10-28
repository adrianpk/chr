package state

import (
	"time"
)

// IntegrationTreeRepair records one host-verified transformation at an
// exact original source step. It supplements method proof, never approval.
type IntegrationTreeRepair struct {
	Input          MergerInput                  `json:"input"`
	Proposal       MergerResult                 `json:"proposal"`
	CompletedAt    time.Time                    `json:"completed_at"`
	OriginalParent string                       `json:"original_parent"`
	PreparedParent string                       `json:"prepared_parent"`
	Source         string                       `json:"source"`
	BaselineTree   string                       `json:"baseline_tree"`
	ProposalTree   string                       `json:"proposal_tree,omitempty"`
	ResolvedTree   string                       `json:"resolved_tree"`
	Regeneration   []MergerRegenerationEvidence `json:"regeneration,omitempty"`
}

// MergerRegenerationEvidence is observed by the host after diff scope
// acceptance. Model-reported local checks cannot supply this evidence.
type MergerRegenerationEvidence struct {
	ConflictID string        `json:"conflict_id"`
	Command    CommandResult `json:"command"`
	Validation CommandResult `json:"validation"`
	InputTree  string        `json:"input_tree"`
	OutputTree string        `json:"output_tree"`
}

// MergerRepairPlan selects an original replay owner or an explicit
// integration fix. Ambiguity produces decisions, not implicit permission.
type MergerRepairPlan struct {
	InputHash      string           `json:"input_hash"`
	Outcome        string           `json:"outcome"`
	OriginalCommit string           `json:"original_commit,omitempty"`
	IntegrationFix *ReviewFinding   `json:"integration_fix,omitempty"`
	Decisions      []MergerDecision `json:"decisions,omitempty"`
}
