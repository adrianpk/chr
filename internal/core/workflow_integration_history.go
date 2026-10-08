package core

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

func (reader workflowIntegrationReader) verifyMethod(ctx context.Context, proof *WorkflowIntegrationProof) error {
	if proof.Method == "rebase" {
		return reader.verifyRebase(ctx, proof)
	}

	parents, err := reader.parents(ctx, proof.ResultCommit)
	if err != nil {
		return err
	}

	if !workflowIntegrationParentsMatch(*proof, parents) {
		return fmt.Errorf("canonical result does not match the configured integration method")
	}

	proof.BaseCommit = parents[0]
	if err := reader.verifyBase(ctx, *proof); err != nil {
		return err
	}

	proof.PreparedTree, err = reader.preparedTree(ctx, proof.OriginCommit, proof.BaseCommit, proof.SourceCommit)
	if err != nil {
		return err
	}

	proof.VerifiedTree, err = reader.tree(ctx, proof.ResultCommit)
	proof.History = []WorkflowIntegrationCommit{{Source: proof.SourceCommit, Commit: proof.ResultCommit, Tree: proof.VerifiedTree, Parent: proof.BaseCommit}}

	return err
}

func workflowIntegrationParentsMatch(proof WorkflowIntegrationProof, parents []string) bool {
	switch proof.Method {
	case "merge":
		return len(parents) == 2 && parents[1] == proof.SourceCommit
	case "squash":
		return len(parents) == 1
	default:
		return false
	}
}

func (reader workflowIntegrationReader) verifyBase(ctx context.Context, proof WorkflowIntegrationProof) error {
	if _, err := reader.output(ctx, "merge-base", "--is-ancestor", proof.OriginCommit, proof.BaseCommit); err != nil {
		return fmt.Errorf("integration base does not contain the frozen target: %w", err)
	}

	return nil
}

func validateWorkflowIntegrationHistory(proof WorkflowIntegrationProof) error {
	if proof.Method != "rebase" && len(proof.History) != 1 {
		return ErrWorkflowConflict
	}

	previous := proof.BaseCommit
	for _, entry := range proof.History {
		if entry.Parent != previous || !validWorkflowIntegrationCommit(entry) {
			return ErrWorkflowConflict
		}

		previous = entry.Commit
	}

	last := proof.History[len(proof.History)-1]
	if last.Source != proof.SourceCommit || last.Commit != proof.ResultCommit || last.Tree != proof.PreparedTree {
		return ErrWorkflowConflict
	}

	return nil
}

func validWorkflowIntegrationCommit(entry WorkflowIntegrationCommit) bool {
	for _, id := range []string{entry.Source, entry.Commit, entry.Tree, entry.Parent} {
		if !workflowCommitID(id) {
			return false
		}
	}

	return true
}

func (reader workflowIntegrationReader) verifyRebase(ctx context.Context, proof *WorkflowIntegrationProof) error {
	sources, err := reader.linearSource(ctx, *proof)
	if err != nil {
		return err
	}

	results := make([]string, len(sources))

	base := proof.ResultCommit
	for i := len(results) - 1; i >= 0; i-- {
		results[i] = base

		parents, err := reader.parents(ctx, base)
		if err != nil || len(parents) != 1 {
			return fmt.Errorf("rebase result requires the complete ordered linear history")
		}

		base = parents[0]
	}

	proof.BaseCommit = base
	if err := reader.verifyBase(ctx, *proof); err != nil {
		return err
	}

	previousSource := proof.OriginCommit
	for i, source := range sources {
		tree, err := reader.preparedTree(ctx, previousSource, base, source)
		if err != nil {
			return err
		}

		actual, err := reader.tree(ctx, results[i])
		if err != nil || tree != actual {
			return fmt.Errorf("rebase ordered commit tree differs from the prepared source change")
		}

		proof.History = append(proof.History, WorkflowIntegrationCommit{Source: source, Commit: results[i], Tree: tree, Parent: base})
		proof.PreparedTree, proof.VerifiedTree = tree, actual
		previousSource, base = source, results[i]
	}

	return nil
}

func (reader workflowIntegrationReader) linearSource(ctx context.Context, proof WorkflowIntegrationProof) ([]string, error) {
	line, err := reader.output(ctx, "rev-list", "--reverse", "--first-parent", proof.OriginCommit+".."+proof.SourceCommit)

	commits := strings.Fields(line)
	if err != nil || len(commits) == 0 || len(commits) > WorkflowRecordLimit {
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

func workflowIntegrationHistoryHash(proof WorkflowIntegrationProof) (string, error) {
	data, err := json.Marshal(struct {
		Method, Source, Base, Tree string
		History                    []WorkflowIntegrationCommit
	}{proof.Method, proof.SourceCommit, proof.BaseCommit, proof.PreparedTree, proof.History})

	return workflowHash(data), err
}

func validateWorkflowIntegrationProof(query WorkflowIntegrationQuery, proof WorkflowIntegrationProof) error {
	query.SourceDirectory = ""
	if proof.WorkflowIntegrationQuery != query || !slices.Contains([]string{"merge", "rebase", "squash"}, proof.Method) || proof.PreparedTree != proof.VerifiedTree || len(proof.History) == 0 || len(proof.History) > WorkflowRecordLimit {
		return ErrWorkflowConflict
	}

	for _, id := range []string{proof.BaseCommit, proof.PreparedTree, proof.VerifiedTree, proof.ObservedTarget} {
		if !workflowCommitID(id) {
			return ErrWorkflowConflict
		}
	}

	hash, err := workflowIntegrationHistoryHash(proof)
	if err != nil || hash != proof.HistoryHash {
		return ErrWorkflowConflict
	}

	return validateWorkflowIntegrationHistory(proof)
}
