package merge

import (
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func validateWorkflowMergerReferences(input workflowstate.MergerInput) error {
	refs := input.EvidenceRefs()
	if len(input.IntentRefs) == 0 || len(input.TargetChangeRefs) == 0 || len(refs) > workflowstate.PageLimit {
		return workflowstate.ErrConflict
	}

	seen := map[string]string{}

	for _, ref := range refs {
		if err := workflowstate.CheckEvidenceRef(input.InvocationID, ref); err != nil {
			return err
		}

		if hash, present := seen[ref.Path]; present && hash != ref.Hash {
			return workflowstate.ErrConflict
		}

		seen[ref.Path] = ref.Hash
	}

	return nil
}
