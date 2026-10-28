package merge

import (
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// CheckHistory checks each repair against the retained ordered replay history.
func CheckHistory(proof workflowstate.IntegrationProof) error {
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

		state := RepairHistory{ReplayIndex: index, History: proof.History[:index]}
		if err := CheckPredecessors(state, repair.Input); err != nil {
			return err
		}
	}

	return nil
}

// CheckPredecessors checks the original replay index and every retained predecessor.
func CheckPredecessors(state RepairHistory, input workflowstate.MergerInput) error {
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

// RepairHistory checks the retained repair observations without executing Git or accepting a candidate.
// RepairHistory contains only the retained replay index and predecessor commits.
type RepairHistory struct {
	ReplayIndex int
	History     []workflowstate.IntegrationCommit
}
