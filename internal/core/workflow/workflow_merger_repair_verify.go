package workflow

import (
	"context"
	"fmt"
	"slices"

	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (reader workflowIntegrationReader) expectedIntegrationTree(ctx context.Context, proof workflowstate.IntegrationProof, original, parent, source string) (string, error) {
	if proof.Provenance == nil || len(proof.Provenance.Repairs) == 0 {
		return reader.preparedTree(ctx, original, parent, source)
	}

	if err := workflowmerge.CheckRepairs(proof); err != nil {
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

		if err := reader.verifyMergerTreeRepair(ctx, repair, workflowmerge.RepairPosition{OriginalParent: original, PreparedParent: parent, Source: source, BaselineTree: tree}, baseline.conflicts); err != nil {
			return "", err
		}

		tree, repaired = repair.ResolvedTree, true
	}

	if len(baseline.conflicts) != 0 && !repaired {
		return "", fmt.Errorf("%w: conflicting tree requires declared host repair", workflowstate.ErrConflict)
	}

	return tree, nil
}

func (reader workflowIntegrationReader) verifyMergerTreeRepair(ctx context.Context, repair workflowstate.IntegrationTreeRepair, position workflowmerge.RepairPosition, conflicts []string) error {
	tree := position.BaselineTree

	if err := workflowmerge.CheckPosition(repair, position); err != nil {
		return err
	}

	if err := reader.checkMergerObservedPaths(ctx, repair.Input, repair.Proposal, workflowMergerPathCheck{Before: tree, After: repair.ResolvedTree, Mode: workflowMergerExactPaths}); err != nil {
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

func (reader workflowIntegrationReader) verifyMergerConflictKinds(ctx context.Context, repair workflowstate.IntegrationTreeRepair, paths []string) error {
	for _, path := range paths {
		index := slices.IndexFunc(repair.Input.Conflicts, func(c workflowstate.MergerConflict) bool { return c.Path == path })
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
