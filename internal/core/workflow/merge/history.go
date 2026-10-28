package merge

import (
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ValidateWorkflowRepairHistory checks the retained repair observations without executing Git or accepting a candidate.
func ValidateWorkflowRepairHistory(proof workflowstate.IntegrationProof) error {
	for _, repair := range proof.Provenance.Repairs {
		index := slices.IndexFunc(proof.History, func(entry workflowstate.IntegrationCommit) bool { return entry.Source == repair.Source })
		if index < 0 || proof.History[index].Parent != repair.PreparedParent {
			return workflowstate.ErrConflict
		}

		original := proof.OriginCommit
		if index > 0 {
			original = proof.History[index-1].Source
		}

		if repair.OriginalParent != original {
			return workflowstate.ErrConflict
		}

		state := WorkflowRepairHistory{ReplayIndex: index, History: proof.History[:index]}
		if err := CheckWorkflowRepairPredecessors(state, repair.Input); err != nil {
			return err
		}
	}

	return nil
}

// CheckWorkflowRepairPredecessors checks the retained repair observations without executing Git or accepting a candidate.
func CheckWorkflowRepairPredecessors(state WorkflowRepairHistory, input workflowstate.MergerInput) error {
	if input.Method != "rebase" {
		return nil
	}

	checkpoint := input.HistoryCheckpoint
	if checkpoint == nil || checkpoint.ReplayIndex != state.ReplayIndex || len(checkpoint.Predecessors) != len(state.History) {
		return workflowstate.ErrConflict
	}

	for i, entry := range state.History {
		if checkpoint.Predecessors[i].Source != entry.Source || checkpoint.Predecessors[i].Commit != entry.Commit {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

// WorkflowRepairHistory checks the retained repair observations without executing Git or accepting a candidate.
type WorkflowRepairHistory struct {
	ReplayIndex int
	History     []workflowstate.IntegrationCommit
}
