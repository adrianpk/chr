package workflow

import (
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// MergerRepairState is a retained private Git checkpoint. A conflicted
// tree is available without publishing or altering the source checkout.
type MergerRepairState struct {
	SchemaVersion  int                                   `json:"schema_version"`
	Query          PreparationQuery                      `json:"query"`
	Preparation    workflowstate.IntegrationPreparation  `json:"preparation"`
	ReplayIndex    int                                   `json:"replay_index"`
	OriginalParent string                                `json:"original_parent"`
	PreparedParent string                                `json:"prepared_parent"`
	CurrentSource  string                                `json:"current_source"`
	Tree           string                                `json:"tree"`
	Conflicts      []string                              `json:"conflicts"`
	History        []workflowstate.IntegrationCommit     `json:"history"`
	Repairs        []workflowstate.IntegrationTreeRepair `json:"repairs,omitempty"`
	Result         *workflowstate.IntegrationProof       `json:"result,omitempty"`
}

// MergerRepairPlan selects an original replay owner or an explicit
// integration fix. Ambiguity produces decisions, not implicit permission.
type MergerRepairPlan struct {
	InputHash      string                         `json:"input_hash"`
	Outcome        string                         `json:"outcome"`
	OriginalCommit string                         `json:"original_commit,omitempty"`
	IntegrationFix *workflowstate.ReviewFinding   `json:"integration_fix,omitempty"`
	Decisions      []workflowstate.MergerDecision `json:"decisions,omitempty"`
}
