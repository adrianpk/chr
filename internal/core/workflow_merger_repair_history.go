package core

import "slices"

func validateWorkflowRepairHistory(proof WorkflowIntegrationProof) error {
	for _, repair := range proof.Provenance.Repairs {
		index := slices.IndexFunc(proof.History, func(entry WorkflowIntegrationCommit) bool { return entry.Source == repair.Source })
		if index < 0 || proof.History[index].Parent != repair.PreparedParent {
			return ErrWorkflowConflict
		}

		original := proof.OriginCommit
		if index > 0 {
			original = proof.History[index-1].Source
		}

		if repair.OriginalParent != original {
			return ErrWorkflowConflict
		}

		state := WorkflowMergerRepairState{ReplayIndex: index, History: proof.History[:index]}
		if err := checkWorkflowRepairPredecessors(state, repair.Input); err != nil {
			return err
		}
	}

	return nil
}
