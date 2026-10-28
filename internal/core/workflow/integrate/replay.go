package integrate

import (
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

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
		if !ValidWorkflowIntegrationFix(fix, entry, prepared) || seen[entry.Commit] {
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

	if !WorkflowReplayAttributionMatches(source, prepared) || history != (workflowstate.IntegrationCommit{Source: source.Commit, Commit: prepared.Commit, Tree: prepared.Tree, Parent: preparedParent}) {
		return false
	}

	return replay.Source == source.Commit && replay.Commit == prepared.Commit && slices.Contains([]string{"nonempty", "source_empty", "already_applied"}, replay.EmptyDecision)
}

// ValidWorkflowIntegrationFix checks one retained host fix against its attributed parent and actionable finding.
func ValidWorkflowIntegrationFix(fix workflowstate.IntegrationFix, entry workflowstate.IntegrationAttribution, parent string) bool {
	if fix.Commit != entry.Commit || !validWorkflowIntegrationAttribution(entry) || !slices.Equal(entry.Parents, []string{parent}) {
		return false
	}

	return workflowstate.ValidateReviewFinding(fix.Finding) == nil && fix.Finding.Category != workflowstate.AgenticReviewFindingCosmetic
}

// WorkflowReplayAttributionMatches preserves original author, encoding and message identity across a replay.
func WorkflowReplayAttributionMatches(source, prepared workflowstate.IntegrationAttribution) bool {
	return prepared.Author == source.Author && prepared.MessageHash == source.MessageHash && prepared.Encoding == source.Encoding
}

type workflowReplayMapping struct {
	Source, Prepared               workflowstate.IntegrationAttribution
	OriginalParent, PreparedParent string
}
