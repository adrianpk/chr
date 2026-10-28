package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
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

	if !workflowIntegrationParentsMatch(*proof, parents) {
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

func workflowIntegrationParentsMatch(proof workflowstate.IntegrationProof, parents []string) bool {
	switch proof.Method {
	case "merge":
		return len(parents) == 2 && parents[1] == proof.SourceCommit
	case "squash":
		return len(parents) == 1
	default:
		return false
	}
}

func (reader workflowIntegrationReader) verifyBase(ctx context.Context, proof workflowstate.IntegrationProof) error {
	if _, err := reader.Output(ctx, "merge-base", "--is-ancestor", proof.OriginCommit, proof.BaseCommit); err != nil {
		return fmt.Errorf("integration base does not contain the frozen target: %w", err)
	}

	return nil
}

func validateWorkflowIntegrationHistory(proof workflowstate.IntegrationProof) error {
	if proof.Method != "rebase" && len(proof.History) != 1 {
		return workflowstate.ErrConflict
	}

	previous := proof.BaseCommit
	for _, entry := range proof.History {
		if entry.Parent != previous || !validWorkflowIntegrationCommit(entry) {
			return workflowstate.ErrConflict
		}

		previous = entry.Commit
	}

	last := proof.History[len(proof.History)-1]
	if last.Source != proof.SourceCommit || !workflowIntegrationHistoryTipMatches(proof, last) {
		return workflowstate.ErrConflict
	}

	return validateWorkflowIntegrationProvenance(proof)
}

func validWorkflowIntegrationCommit(entry workflowstate.IntegrationCommit) bool {
	for _, id := range []string{entry.Source, entry.Commit, entry.Tree, entry.Parent} {
		if !workflowstate.ValidCommitID(id) {
			return false
		}
	}

	return true
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

func workflowIntegrationHistoryHash(proof workflowstate.IntegrationProof) (string, error) {
	data, err := json.Marshal(struct {
		Method, Source, Base, Tree string
		History                    []workflowstate.IntegrationCommit
		Provenance                 *workflowstate.IntegrationProvenance `json:",omitempty"`
	}{proof.Method, proof.SourceCommit, proof.BaseCommit, proof.PreparedTree, proof.History, proof.Provenance})

	return contenthash.Sum(data), err
}

func validateWorkflowIntegrationProof(query workflowstate.IntegrationQuery, proof workflowstate.IntegrationProof) error {
	query.SourceDirectory = ""
	if proof.IntegrationQuery != query || !slices.Contains([]string{"merge", "rebase", "squash"}, proof.Method) || proof.PreparedTree != proof.VerifiedTree || len(proof.History) == 0 || len(proof.History) > workflowstate.RecordLimit {
		return workflowstate.ErrConflict
	}

	for _, id := range []string{proof.BaseCommit, proof.PreparedTree, proof.VerifiedTree, proof.ObservedTarget} {
		if !workflowstate.ValidCommitID(id) {
			return workflowstate.ErrConflict
		}
	}

	hash, err := workflowIntegrationHistoryHash(proof)
	if err != nil || hash != proof.HistoryHash {
		return workflowstate.ErrConflict
	}

	return validateWorkflowIntegrationHistory(proof)
}
