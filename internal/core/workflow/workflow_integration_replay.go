package workflow

import (
	"context"
	"fmt"
	"slices"

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
		if err != nil || i >= len(p.Source) || !workflowReplayAttributionMatches(p.Source[i], prepared) {
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

func validateWorkflowRebaseProvenance(proof workflowstate.IntegrationProof, p workflowstate.IntegrationProvenance) error {
	if len(proof.History) != len(p.Source) {
		return workflowstate.ErrConflict
	}

	original, prepared := proof.OriginCommit, proof.BaseCommit
	seen := map[string]bool{}

	for i, source := range p.Source {
		entry := p.Prepared[i]
		if !validWorkflowReplayMapping(proof.History[i], p.Replays[i], workflowReplayMapping{Source: source, Prepared: entry, OriginalParent: original, PreparedParent: prepared}) || seen[entry.Commit] {
			return workflowstate.ErrConflict
		}

		seen[entry.Commit] = true
		original, prepared = source.Commit, entry.Commit
	}

	for i, fix := range p.IntegrationFixes {
		entry := p.Prepared[len(p.Source)+i]
		if !validWorkflowIntegrationFix(fix, entry, prepared) || seen[entry.Commit] {
			return workflowstate.ErrConflict
		}

		seen[entry.Commit] = true
		prepared = entry.Commit
	}

	last := p.Prepared[len(p.Prepared)-1]
	if last.Commit != proof.ResultCommit || last.Tree != proof.PreparedTree {
		return workflowstate.ErrConflict
	}

	return nil
}

func validWorkflowReplayMapping(history workflowstate.IntegrationCommit, replay workflowstate.IntegrationReplay, mapping workflowReplayMapping) bool {
	source, prepared := mapping.Source, mapping.Prepared
	originalParent, preparedParent := mapping.OriginalParent, mapping.PreparedParent

	if !validWorkflowIntegrationAttribution(prepared) || !slices.Equal(source.Parents, []string{originalParent}) || !slices.Equal(prepared.Parents, []string{preparedParent}) {
		return false
	}

	if !workflowReplayAttributionMatches(source, prepared) || history != (workflowstate.IntegrationCommit{Source: source.Commit, Commit: prepared.Commit, Tree: prepared.Tree, Parent: preparedParent}) {
		return false
	}

	return replay.Source == source.Commit && replay.Commit == prepared.Commit && slices.Contains([]string{"nonempty", "source_empty", "already_applied"}, replay.EmptyDecision)
}

func validWorkflowIntegrationFix(fix workflowstate.IntegrationFix, entry workflowstate.IntegrationAttribution, parent string) bool {
	if fix.Commit != entry.Commit || !validWorkflowIntegrationAttribution(entry) || !slices.Equal(entry.Parents, []string{parent}) {
		return false
	}

	return workflowstate.ValidateReviewFinding(fix.Finding) == nil && fix.Finding.Category != workflowstate.AgenticReviewFindingCosmetic
}

func workflowReplayAttributionMatches(source, prepared workflowstate.IntegrationAttribution) bool {
	return prepared.Author == source.Author && prepared.MessageHash == source.MessageHash && prepared.Encoding == source.Encoding
}
