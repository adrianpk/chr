package workflow

import (
	"context"
	"fmt"
	"slices"

	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (reader workflowIntegrationReader) integrationProvenance(ctx context.Context, proof workflowstate.IntegrationProof) (*workflowstate.IntegrationProvenance, error) {
	var (
		p   *workflowstate.IntegrationProvenance
		err error
	)

	if proof.Method == "rebase" {
		p, err = reader.rebaseProvenance(ctx, proof)
	} else {
		p, err = reader.combinedProvenance(ctx, proof)
	}

	if err != nil {
		return nil, err
	}

	return p, reader.addPreparationMetadata(ctx, proof, p)
}

func (reader workflowIntegrationReader) rebaseProvenance(ctx context.Context, proof workflowstate.IntegrationProof) (*workflowstate.IntegrationProvenance, error) {
	p, err := reader.combinedProvenance(ctx, proof)
	if err != nil {
		return nil, err
	}

	p.AttributionPolicy = "preserve_source_authorship"
	p.Prepared = nil

	if proof.Provenance != nil {
		p.IntegrationFixes = slices.Clone(proof.Provenance.IntegrationFixes)
	}

	for i, entry := range proof.History {
		prepared, err := reader.commitAttribution(ctx, entry.Commit)
		if err != nil || i >= len(p.Source) || !workflowintegrate.ReplayAttributionMatches(p.Source[i], prepared) {
			return nil, fmt.Errorf("rebase must preserve original author, date and message")
		}

		decision, err := reader.replayEmptyDecision(ctx, p.Source[i], prepared)
		if err != nil {
			return nil, err
		}

		p.Prepared = append(p.Prepared, prepared)
		p.Replays = append(p.Replays, workflowstate.IntegrationReplay{Source: entry.Source, Commit: entry.Commit, EmptyDecision: decision})
	}

	if err := reader.appendRebaseFixAttributions(ctx, p); err != nil {
		return nil, err
	}

	return p, validateWorkflowIntegrationProvenance(workflowstate.IntegrationProof{IntegrationQuery: proof.IntegrationQuery, BaseCommit: proof.BaseCommit, PreparedTree: proof.PreparedTree, History: proof.History, Provenance: p})
}

func (reader workflowIntegrationReader) replayEmptyDecision(ctx context.Context, source, prepared workflowstate.IntegrationAttribution) (string, error) {
	originalTree, err := reader.tree(ctx, source.Parents[0])
	if err != nil {
		return "", err
	}

	parentTree, err := reader.tree(ctx, prepared.Parents[0])
	if err != nil {
		return "", err
	}

	if prepared.Tree != parentTree {
		return "nonempty", nil
	}

	if source.Tree == originalTree {
		return "source_empty", nil
	}

	return "already_applied", nil
}

func (reader workflowIntegrationReader) appendRebaseFixAttributions(ctx context.Context, p *workflowstate.IntegrationProvenance) error {
	for _, fix := range p.IntegrationFixes {
		entry, err := reader.commitAttribution(ctx, fix.Commit)
		if err != nil {
			return err
		}

		p.Prepared = append(p.Prepared, entry)
	}

	return nil
}
