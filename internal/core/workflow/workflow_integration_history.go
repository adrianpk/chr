package workflow

import (
	"context"
	"fmt"
	"strings"

	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (reader workflowIntegrationReader) verifyMethod(ctx context.Context, proof *workflowstate.IntegrationProof) error {
	if proof.Method == "rebase" {
		return reader.verifyRebase(ctx, proof)
	}

	parents, err := reader.parents(ctx, proof.ResultCommit)
	if err != nil {
		return err
	}

	if !workflowintegrate.WorkflowIntegrationParentsMatch(*proof, parents) {
		return fmt.Errorf("canonical result does not match the configured integration method")
	}

	proof.BaseCommit = parents[0]
	if err := reader.verifyBase(ctx, *proof); err != nil {
		return err
	}

	proof.PreparedTree, err = reader.expectedIntegrationTree(ctx, *proof, proof.OriginCommit, proof.BaseCommit, proof.SourceCommit)
	if err != nil {
		return err
	}

	proof.VerifiedTree, err = reader.tree(ctx, proof.ResultCommit)
	proof.History = []workflowstate.IntegrationCommit{{Source: proof.SourceCommit, Commit: proof.ResultCommit, Tree: proof.VerifiedTree, Parent: proof.BaseCommit}}

	if err != nil {
		return err
	}

	return reader.verifyCombinedProvenance(ctx, proof)
}

func (reader workflowIntegrationReader) verifyBase(ctx context.Context, proof workflowstate.IntegrationProof) error {
	if _, err := reader.Output(ctx, "merge-base", "--is-ancestor", proof.OriginCommit, proof.BaseCommit); err != nil {
		return fmt.Errorf("integration base does not contain the frozen target: %w", err)
	}

	return nil
}

func (reader workflowIntegrationReader) verifyRebase(ctx context.Context, proof *workflowstate.IntegrationProof) error {
	sources, err := reader.linearSource(ctx, *proof)
	if err != nil {
		return err
	}

	rebaseResults, err := reader.rebaseResults(ctx, *proof, len(sources))

	results, base := rebaseResults.Commits, rebaseResults.Base

	if err != nil {
		return err
	}

	proof.BaseCommit = base
	if err := reader.verifyBase(ctx, *proof); err != nil {
		return err
	}

	previousSource := proof.OriginCommit
	for i, source := range sources {
		tree, err := reader.expectedIntegrationTree(ctx, *proof, previousSource, base, source)
		if err != nil {
			return err
		}

		actual, err := reader.tree(ctx, results[i])
		if err != nil || tree != actual {
			return fmt.Errorf("rebase ordered commit tree differs from the prepared source change")
		}

		proof.History = append(proof.History, workflowstate.IntegrationCommit{Source: source, Commit: results[i], Tree: tree, Parent: base})
		proof.PreparedTree, proof.VerifiedTree = tree, actual
		previousSource, base = source, results[i]
	}

	return reader.verifyRebaseProvenance(ctx, proof)
}

func (reader workflowIntegrationReader) linearSource(ctx context.Context, proof workflowstate.IntegrationProof) ([]string, error) {
	line, err := reader.Output(ctx, "rev-list", "--reverse", "--first-parent", "--max-count=4097", proof.OriginCommit+".."+proof.SourceCommit)

	commits := strings.Fields(line)
	if err != nil || len(commits) == 0 || len(commits) > workflowstate.RecordLimit {
		return nil, fmt.Errorf("integration source history is empty or exceeds commit bounds")
	}

	previous := proof.OriginCommit

	for _, commit := range commits {
		parents, err := reader.parents(ctx, commit)
		if err != nil || len(parents) != 1 || parents[0] != previous {
			return nil, fmt.Errorf("configured rebase requires an ordered linear source")
		}

		previous = commit
	}

	return commits, nil
}
