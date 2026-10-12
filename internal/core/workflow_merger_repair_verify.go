package core

import (
	"context"
	"fmt"
	"slices"
)

func (reader workflowIntegrationReader) expectedIntegrationTree(ctx context.Context, proof WorkflowIntegrationProof, original, parent, source string) (string, error) {
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
		return "", fmt.Errorf("%w: conflicting tree requires declared host repair", ErrWorkflowConflict)
	}

	return tree, nil
}

func (reader workflowIntegrationReader) verifyMergerTreeRepair(ctx context.Context, repair WorkflowIntegrationTreeRepair, original, parent, source, tree string, conflicts []string) error {
	if err := validateWorkflowRepairPosition(repair, original, parent, source, tree); err != nil {
		return err
	}

	data, err := reader.rawOutput(ctx, "diff", "--name-only", "-z", "--no-renames", "--no-ext-diff", "--no-textconv", tree, repair.ResolvedTree, "--")
	if err != nil {
		return err
	}

	paths, err := workflowMergerChangedPaths(data)
	if err != nil {
		return err
	}

	declared := slices.Clone(repair.Proposal.ChangedPaths)
	slices.Sort(declared)

	if !slices.Equal(paths, declared) {
		return fmt.Errorf("%w: actual repair diff differs from declared paths", ErrWorkflowConflict)
	}

	for _, path := range paths {
		if !workflowMergerPathAllowed(repair.Input, path) {
			return ErrWorkflowConflict
		}
	}

	for _, path := range conflicts {
		if !slices.Contains(paths, path) {
			return fmt.Errorf("%w: conflicting path has no resolution change", ErrWorkflowConflict)
		}
	}

	return reader.verifyMergerConflictKinds(ctx, repair, conflicts)
}

func (reader workflowIntegrationReader) verifyMergerConflictKinds(ctx context.Context, repair WorkflowIntegrationTreeRepair, paths []string) error {
	for _, path := range paths {
		index := slices.IndexFunc(repair.Input.Conflicts, func(c WorkflowMergerConflict) bool { return c.Path == path })
		if index < 0 {
			return ErrWorkflowConflict
		}

		for _, object := range []string{repair.OriginalParent, repair.PreparedParent, repair.Source} {
			kind, err := reader.mergerFileKind(ctx, object, path)
			if err != nil {
				return err
			}

			if kind != "text" && repair.Input.Conflicts[index].Kind != kind {
				return fmt.Errorf("%w: conflict type was misrepresented", ErrWorkflowConflict)
			}
		}
	}

	return nil
}
