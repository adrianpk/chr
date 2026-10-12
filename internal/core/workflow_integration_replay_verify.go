package core

import (
	"context"
	"fmt"
)

func workflowIntegrationHistoryTipMatches(proof WorkflowIntegrationProof, last WorkflowIntegrationCommit) bool {
	if proof.Method == "rebase" && proof.Provenance != nil {
		p := proof.Provenance
		if !validWorkflowProvenanceHeader(proof, *p) || len(p.Source) != len(proof.History) {
			return false
		}

		entry := p.Prepared[len(p.Source)-1]

		return entry.Commit == last.Commit && entry.Tree == last.Tree
	}

	return last.Commit == proof.ResultCommit && last.Tree == proof.PreparedTree
}

func (reader workflowIntegrationReader) rebaseResults(ctx context.Context, proof WorkflowIntegrationProof, count int) ([]string, string, error) {
	base := proof.ResultCommit
	if p := proof.Provenance; p != nil {
		if !validWorkflowProvenanceHeader(proof, *p) || len(p.Source) != count {
			return nil, "", ErrWorkflowConflict
		}

		base = p.Prepared[count-1].Commit
	}

	results := make([]string, count)
	for i := count - 1; i >= 0; i-- {
		results[i] = base

		parents, err := reader.parents(ctx, base)
		if err != nil || len(parents) != 1 {
			return nil, "", fmt.Errorf("rebase result requires the complete ordered linear history")
		}

		base = parents[0]
	}

	return results, base, nil
}

func (reader workflowIntegrationReader) verifyRebaseProvenance(ctx context.Context, proof *WorkflowIntegrationProof) error {
	if proof.Provenance == nil {
		return nil
	}

	parent := proof.History[len(proof.History)-1].Commit
	for _, fix := range proof.Provenance.IntegrationFixes {
		entry, err := reader.commitAttribution(ctx, fix.Commit)
		if err != nil || !validWorkflowIntegrationFix(fix, entry, parent) || entry.Tree == proof.PreparedTree {
			return ErrWorkflowConflict
		}

		parent = entry.Commit
		proof.PreparedTree, proof.VerifiedTree = entry.Tree, entry.Tree
	}

	actual, err := reader.rebaseProvenance(ctx, *proof)
	if err != nil {
		return err
	}

	if !workflowNativeJSONEqual(actual, proof.Provenance) {
		return ErrWorkflowConflict
	}

	return nil
}
