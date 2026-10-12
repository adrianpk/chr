package core

import (
	"slices"
	"strings"
)

func validateWorkflowTreeRepairs(proof WorkflowIntegrationProof) error {
	if proof.Provenance == nil {
		return nil
	}

	if len(proof.Provenance.Repairs) > WorkflowPageLimit {
		return ErrWorkflowConflict
	}

	for _, repair := range proof.Provenance.Repairs {
		if err := validateWorkflowTreeRepair(proof, repair); err != nil {
			return err
		}
	}

	return nil
}

func validateWorkflowTreeRepair(proof WorkflowIntegrationProof, repair WorkflowIntegrationTreeRepair) error {
	if repair.CompletedAt.IsZero() || validateWorkflowMergerInput(repair.Input, repair.CompletedAt) != nil {
		return ErrWorkflowConflict
	}

	input := repair.Input
	if !workflowMergerRepairIdentity(input, proof) {
		return ErrWorkflowConflict
	}

	if repair.Proposal.Outcome != "resolved" || validateWorkflowMergerResult(repair.Proposal, input) != nil {
		return ErrWorkflowConflict
	}

	for _, object := range []string{repair.Source, repair.OriginalParent, repair.PreparedParent, repair.BaselineTree, repair.ResolvedTree} {
		if !workflowCommitID(object) || len(object) != len(input.Source) {
			return ErrWorkflowConflict
		}
	}

	if repair.BaselineTree == repair.ResolvedTree {
		return ErrWorkflowConflict
	}

	return validateWorkflowRepairRegeneration(repair)
}

func validateWorkflowRepairRegeneration(repair WorkflowIntegrationTreeRepair) error {
	seen := map[string]bool{}

	for _, evidence := range repair.Regeneration {
		index := slices.IndexFunc(repair.Input.Conflicts, func(c WorkflowMergerConflict) bool { return c.ID == evidence.ConflictID })
		if index < 0 || seen[evidence.ConflictID] {
			return ErrWorkflowConflict
		}

		conflict := repair.Input.Conflicts[index]
		if !workflowMergerProcedureRequired(conflict) || !validWorkflowRegenerationEvidence(conflict, evidence, repair.ResolvedTree) || evidence.InputTree != repair.BaselineTree {
			return ErrWorkflowConflict
		}

		seen[evidence.ConflictID] = true
	}

	for _, conflict := range repair.Input.Conflicts {
		if workflowMergerProcedureRequired(conflict) && !seen[conflict.ID] {
			return ErrWorkflowConflict
		}
	}

	return nil
}

func validWorkflowRegenerationEvidence(conflict WorkflowMergerConflict, evidence WorkflowMergerRegenerationEvidence, tree string) bool {
	if !slices.Equal(conflict.RegenerationCommand, evidence.Command.Command) || !slices.Equal(conflict.ValidationCommand, evidence.Validation.Command) {
		return false
	}

	return evidence.Command.ExitCode == 0 && evidence.Validation.ExitCode == 0 && workflowCommitID(evidence.InputTree) && evidence.OutputTree == tree && !evidence.Command.Truncated && !evidence.Validation.Truncated
}

func validateWorkflowRepairPosition(repair WorkflowIntegrationTreeRepair, original, parent, source, tree string) error {
	if repair.OriginalParent != original || repair.PreparedParent != parent || repair.Source != source || repair.BaselineTree != tree {
		return ErrWorkflowConflict
	}

	if repair.Input.Method == "rebase" {
		checkpoint := repair.Input.HistoryCheckpoint
		if checkpoint == nil || checkpoint.OriginalCommit != source || checkpoint.PreparedParent != parent {
			return ErrWorkflowConflict
		}
	}

	return nil
}

func workflowMergerChangedPaths(data string) ([]string, error) {
	paths := strings.Split(strings.TrimSuffix(data, "\x00"), "\x00")
	if data == "" {
		paths = nil
	}

	if err := validateWorkflowMergerPaths(paths); err != nil {
		return nil, err
	}

	slices.Sort(paths)

	return paths, nil
}
