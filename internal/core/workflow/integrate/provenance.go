package integrate

import (
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ValidProvenanceHeader checks the bounded retained envelope and method-specific source/prepared counts.
func ValidProvenanceHeader(proof workflowstate.IntegrationProof, p workflowstate.IntegrationProvenance) bool {
	if !validProvenanceEnvelope(proof, p) || !validProvenanceCount(p) {
		return false
	}

	if proof.Method == "rebase" {
		return p.AttributionPolicy == "preserve_source_authorship" && len(p.Replays) == len(p.Source) && len(p.Prepared) == len(p.Source)+len(p.IntegrationFixes)
	}

	return p.AttributionPolicy == "host_author_with_original_source_manifest" && len(p.Prepared) == 1 && len(p.Replays) == 0 && len(p.IntegrationFixes) == 0
}

func validProvenanceEnvelope(proof workflowstate.IntegrationProof, p workflowstate.IntegrationProvenance) bool {
	expectedFormat := "sha1"
	if len(proof.ResultCommit) == 64 {
		expectedFormat = "sha256"
	}

	return p.Schema == "chr-integration-provenance/v1" && workflowstate.ValidFingerprint(p.CombinedDiffHash) && p.ObjectFormat == expectedFormat
}

func validProvenanceCount(p workflowstate.IntegrationProvenance) bool {
	return len(p.Source) > 0 && len(p.Source) <= workflowstate.RecordLimit && len(p.Prepared) > 0 && len(p.Prepared) <= workflowstate.RecordLimit
}

func validAttributedSources(proof workflowstate.IntegrationProof, sources []workflowstate.IntegrationAttribution) bool {
	seen := map[string]bool{}
	for _, entry := range sources {
		if !validAttribution(entry) || len(entry.Commit) != len(proof.SourceCommit) || seen[entry.Commit] {
			return false
		}

		seen[entry.Commit] = true
	}

	return sources[len(sources)-1].Commit == proof.SourceCommit
}

// CheckProvenanceCommits checks source/prepared attribution and replay identity after the provenance header is checked.
func CheckProvenanceCommits(proof workflowstate.IntegrationProof, p workflowstate.IntegrationProvenance) error {
	if !validAttributedSources(proof, p.Source) {
		return workflowstate.ErrConflict
	}

	if proof.Method == "rebase" {
		return checkRebaseProvenance(proof, p)
	}

	result := p.Prepared[0]

	parents := []string{proof.BaseCommit}
	if proof.Method == "merge" {
		parents = append(parents, proof.SourceCommit)
	}

	if !validAttribution(result) || result.Commit != proof.ResultCommit || result.Tree != proof.PreparedTree || !slices.Equal(result.Parents, parents) {
		return workflowstate.ErrConflict
	}

	return nil
}
