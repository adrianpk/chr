package workflow

import (
	"context"
	"encoding/json"
	"slices"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (reader workflowIntegrationReader) combinedProvenance(ctx context.Context, proof workflowstate.IntegrationProof) (*workflowstate.IntegrationProvenance, error) {
	source, err := reader.sourceAttributions(ctx, proof)
	if err != nil {
		return nil, err
	}

	prepared, err := reader.commitAttribution(ctx, proof.ResultCommit)
	if err != nil {
		return nil, err
	}

	format, err := reader.Output(ctx, "rev-parse", "--show-object-format")
	if err != nil {
		return nil, err
	}

	diff, err := reader.RawOutput(ctx, "-c", "core.quotePath=true", "-c", "color.ui=false", "diff", "--binary", "--full-index", "--no-ext-diff", "--no-textconv", "--no-renames", "--diff-algorithm=myers", "--no-indent-heuristic", "--unified=3", "--src-prefix=a/", "--dst-prefix=b/", proof.BaseCommit, proof.ResultCommit, "--")
	if err != nil {
		return nil, err
	}

	p := &workflowstate.IntegrationProvenance{Schema: "chr-integration-provenance/v1", ObjectFormat: format, AttributionPolicy: "host_author_with_original_source_manifest", CombinedDiffHash: contenthash.Sum([]byte(diff)), Source: source, Prepared: []workflowstate.IntegrationAttribution{prepared}}
	if proof.Provenance != nil {
		p.Repairs = slices.Clone(proof.Provenance.Repairs)
	}

	return p, nil
}

func (reader workflowIntegrationReader) verifyCombinedProvenance(ctx context.Context, proof *workflowstate.IntegrationProof) error {
	if proof.Provenance == nil {
		return nil
	}

	actual, err := reader.integrationProvenance(ctx, *proof)
	if err != nil {
		return err
	}

	if !workflowNativeJSONEqual(actual, proof.Provenance) {
		return workflowstate.ErrConflict
	}

	return nil
}

func validateWorkflowIntegrationProvenance(proof workflowstate.IntegrationProof) error {
	p := proof.Provenance
	if p == nil {
		return nil
	}

	if !validWorkflowProvenanceHeader(proof, *p) {
		return workflowstate.ErrConflict
	}

	if err := validateWorkflowProvenanceCommits(proof, *p); err != nil {
		return err
	}

	if p.Preparation != nil && !validWorkflowPreparationMetadata(proof, *p.Preparation) {
		return workflowstate.ErrConflict
	}

	if err := validateWorkflowTreeRepairs(proof); err != nil {
		return err
	}

	if err := validateWorkflowRepairHistory(proof); err != nil {
		return err
	}

	data, err := json.Marshal(p)
	if err != nil || len(data) >= workflowstate.FileLimit {
		return workflowstate.ErrConflict
	}

	return nil
}

func validWorkflowProvenanceHeader(proof workflowstate.IntegrationProof, p workflowstate.IntegrationProvenance) bool {
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

	return p.Schema == "chr-integration-provenance/v1" && validWorkflowNativeFingerprint(p.CombinedDiffHash) && p.ObjectFormat == expectedFormat
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

func validateWorkflowProvenanceCommits(proof workflowstate.IntegrationProof, p workflowstate.IntegrationProvenance) error {
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
