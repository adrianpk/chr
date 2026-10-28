package workflow

import (
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func validateWorkflowRepairHistory(proof workflowstate.IntegrationProof) error {
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

		state := MergerRepairState{ReplayIndex: index, History: proof.History[:index]}
		if err := checkWorkflowRepairPredecessors(state, repair.Input); err != nil {
			return err
		}
	}

	return nil
}
