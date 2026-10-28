package workflow

import (
	"slices"
	"strings"

	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func validateWorkflowTreeRepairs(proof workflowstate.IntegrationProof) error {
	if proof.Provenance == nil {
		return nil
	}

	if len(proof.Provenance.Repairs) > workflowstate.PageLimit {
		return workflowstate.ErrConflict
	}

	for _, repair := range proof.Provenance.Repairs {
		if err := validateWorkflowTreeRepair(proof, repair); err != nil {
			return err
		}
	}

	return nil
}

func validateWorkflowTreeRepair(proof workflowstate.IntegrationProof, repair workflowstate.IntegrationTreeRepair) error {
	if repair.CompletedAt.IsZero() || validateWorkflowMergerInput(repair.Input, repair.CompletedAt) != nil {
		return workflowstate.ErrConflict
	}

	input := repair.Input
	if !workflowMergerRepairIdentity(input, proof) {
		return workflowstate.ErrConflict
	}

	if repair.Proposal.Outcome != "resolved" || validateWorkflowMergerResult(repair.Proposal, input) != nil {
		return workflowstate.ErrConflict
	}

	for _, object := range []string{repair.Source, repair.OriginalParent, repair.PreparedParent, repair.BaselineTree, repair.ResolvedTree} {
		if !workflowCommitID(object) || len(object) != len(input.Source) {
			return workflowstate.ErrConflict
		}
	}

	if repair.BaselineTree == repair.ResolvedTree {
		return workflowstate.ErrConflict
	}

	return validateWorkflowRepairRegeneration(repair)
}

func validateWorkflowRepairRegeneration(repair workflowstate.IntegrationTreeRepair) error {
	seen := map[string]bool{}

	tree := repair.ProposalTree
	if tree == "" {
		tree = repair.BaselineTree
	}

	if !workflowCommitID(tree) {
		return workflowstate.ErrConflict
	}

	for _, evidence := range repair.Regeneration {
		if err := validateWorkflowRegenerationStep(repair.Input, evidence, seen, tree); err != nil {
			return err
		}

		seen[evidence.ConflictID] = true
		tree = evidence.OutputTree
	}

	if len(repair.Regeneration) > 0 && tree != repair.ResolvedTree {
		return workflowstate.ErrConflict
	}

	for _, conflict := range repair.Input.Conflicts {
		if workflowMergerProcedureRequired(conflict) && !seen[conflict.ID] {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func validWorkflowRegenerationEvidence(conflict workflowstate.MergerConflict, evidence workflowstate.MergerRegenerationEvidence, tree string) bool {
	if !slices.Equal(conflict.RegenerationCommand, evidence.Command.Command) || !slices.Equal(conflict.ValidationCommand, evidence.Validation.Command) {
		return false
	}

	return evidence.Command.ExitCode == 0 && evidence.Validation.ExitCode == 0 && workflowCommitID(evidence.InputTree) && workflowCommitID(evidence.OutputTree) && evidence.OutputTree == tree && !evidence.Command.Truncated && !evidence.Validation.Truncated
}

func validateWorkflowRepairPosition(repair workflowstate.IntegrationTreeRepair, original, parent, source, tree string) error {
	if repair.OriginalParent != original || repair.PreparedParent != parent || repair.Source != source || repair.BaselineTree != tree {
		return workflowstate.ErrConflict
	}

	if repair.Input.Method == "rebase" {
		checkpoint := repair.Input.HistoryCheckpoint
		if checkpoint == nil || checkpoint.OriginalCommit != source || checkpoint.PreparedParent != parent {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func workflowMergerChangedPaths(data string) ([]string, error) {
	paths := strings.Split(strings.TrimSuffix(data, "\x00"), "\x00")
	if data == "" {
		paths = nil
	}

	if err := workflowmerge.ValidateWorkflowMergerPaths(paths); err != nil {
		return nil, err
	}

	slices.Sort(paths)

	return paths, nil
}

func validateWorkflowRegenerationStep(input workflowstate.MergerInput, evidence workflowstate.MergerRegenerationEvidence, seen map[string]bool, tree string) error {
	index := slices.IndexFunc(input.Conflicts, func(c workflowstate.MergerConflict) bool { return c.ID == evidence.ConflictID })
	if index < 0 || seen[evidence.ConflictID] {
		return workflowstate.ErrConflict
	}

	conflict := input.Conflicts[index]
	if !workflowMergerProcedureRequired(conflict) || !validWorkflowRegenerationEvidence(conflict, evidence, evidence.OutputTree) || evidence.InputTree != tree {
		return workflowstate.ErrConflict
	}

	return nil
}
