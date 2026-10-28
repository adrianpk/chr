package workflow

import (
	"context"
	"fmt"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (reader workflowIntegrationReader) expectedIntegrationTree(ctx context.Context, proof IntegrationProof, original, parent, source string) (string, error) {
	if proof.Provenance == nil || len(proof.Provenance.Repairs) == 0 {
		return reader.preparedTree(ctx, original, parent, source)
	}

	if err := validateWorkflowTreeRepairs(proof); err != nil {
		return "", err
	}

	baseline, err := reader.mergerTree(ctx, original, parent, source)
	if err != nil {
		return "", err
	}

	tree, repaired := baseline.tree, false

	for _, repair := range proof.Provenance.Repairs {
		if repair.Source != source {
			continue
		}

		if err := reader.verifyMergerTreeRepair(ctx, repair, original, parent, source, tree, baseline.conflicts); err != nil {
			return "", err
		}

		tree, repaired = repair.ResolvedTree, true
	}

	if len(baseline.conflicts) != 0 && !repaired {
		return "", fmt.Errorf("%w: conflicting tree requires declared host repair", workflowstate.ErrConflict)
	}

	return tree, nil
}

func (reader workflowIntegrationReader) verifyMergerTreeRepair(ctx context.Context, repair IntegrationTreeRepair, original, parent, source, tree string, conflicts []string) error {
	if err := validateWorkflowRepairPosition(repair, original, parent, source, tree); err != nil {
		return err
	}

	if err := reader.checkMergerObservedPaths(ctx, repair.Input, repair.Proposal, tree, repair.ResolvedTree, true); err != nil {
		return err
	}

	paths, err := reader.mergerDiffPaths(ctx, tree, repair.ResolvedTree)
	if err != nil {
		return err
	}

	for _, path := range conflicts {
		if !slices.Contains(paths, path) {
			return fmt.Errorf("%w: conflicting path has no resolution change", workflowstate.ErrConflict)
		}
	}

	if err := reader.checkMergerChangedKinds(ctx, repair.Input, paths, repair.OriginalParent, repair.PreparedParent, repair.Source, repair.ResolvedTree); err != nil {
		return err
	}

	if err := reader.verifyMergerRegenerationTrees(ctx, repair); err != nil {
		return err
	}

	return reader.verifyMergerConflictKinds(ctx, repair, conflicts)
}

func (reader workflowIntegrationReader) verifyMergerConflictKinds(ctx context.Context, repair IntegrationTreeRepair, paths []string) error {
	for _, path := range paths {
		index := slices.IndexFunc(repair.Input.Conflicts, func(c MergerConflict) bool { return c.Path == path })
		if index < 0 {
			return workflowstate.ErrConflict
		}

		for _, object := range []string{repair.OriginalParent, repair.PreparedParent, repair.Source} {
			kind, err := reader.mergerFileKind(ctx, object, path)
			if err != nil {
				return err
			}

			if kind != "text" && repair.Input.Conflicts[index].Kind != kind {
				return fmt.Errorf("%w: conflict type was misrepresented", workflowstate.ErrConflict)
			}
		}
	}

	return nil
}
