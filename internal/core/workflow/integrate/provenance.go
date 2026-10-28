package integrate

import (
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ValidWorkflowProvenanceHeader checks the bounded retained envelope and method-specific source/prepared counts.
func ValidWorkflowProvenanceHeader(proof workflowstate.IntegrationProof, p workflowstate.IntegrationProvenance) bool {
	if !validWorkflowProvenanceEnvelope(proof, p) || !validWorkflowProvenanceCount(p) {
		return false
	}

	if proof.Method == "rebase" {
		return p.AttributionPolicy == "preserve_source_authorship" && len(p.Replays) == len(p.Source) && len(p.Prepared) == len(p.Source)+len(p.IntegrationFixes)
	}

	return p.AttributionPolicy == "host_author_with_original_source_manifest" && len(p.Prepared) == 1 && len(p.Replays) == 0 && len(p.IntegrationFixes) == 0
}

func validWorkflowProvenanceEnvelope(proof workflowstate.IntegrationProof, p workflowstate.IntegrationProvenance) bool {
	expectedFormat := "sha1"
	if len(proof.ResultCommit) == 64 {
		expectedFormat = "sha256"
	}

	return p.Schema == "chr-integration-provenance/v1" && workflowstate.ValidFingerprint(p.CombinedDiffHash) && p.ObjectFormat == expectedFormat
}

func validWorkflowProvenanceCount(p workflowstate.IntegrationProvenance) bool {
	return len(p.Source) > 0 && len(p.Source) <= workflowstate.RecordLimit && len(p.Prepared) > 0 && len(p.Prepared) <= workflowstate.RecordLimit
}

func validWorkflowAttributedSources(proof workflowstate.IntegrationProof, sources []workflowstate.IntegrationAttribution) bool {
	seen := map[string]bool{}
	for _, entry := range sources {
		if !validWorkflowIntegrationAttribution(entry) || len(entry.Commit) != len(proof.SourceCommit) || seen[entry.Commit] {
			return false
		}

		seen[entry.Commit] = true
	}

	return sources[len(sources)-1].Commit == proof.SourceCommit
}

// ValidateWorkflowProvenanceCommits checks source/prepared attribution and replay identity after the provenance header is checked.
func ValidateWorkflowProvenanceCommits(proof workflowstate.IntegrationProof, p workflowstate.IntegrationProvenance) error {
	if !validWorkflowAttributedSources(proof, p.Source) {
		return workflowstate.ErrConflict
	}

	if proof.Method == "rebase" {
		return validateWorkflowRebaseProvenance(proof, p)
	}

	result := p.Prepared[0]

	parents := []string{proof.BaseCommit}
	if proof.Method == "merge" {
		parents = append(parents, proof.SourceCommit)
	}

	if !validWorkflowIntegrationAttribution(result) || result.Commit != proof.ResultCommit || result.Tree != proof.PreparedTree || !slices.Equal(result.Parents, parents) {
		return workflowstate.ErrConflict
	}

	return nil
}
