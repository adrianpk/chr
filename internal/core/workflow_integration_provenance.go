package core

import (
	"context"
	"encoding/json"
	"slices"
)

// WorkflowIntegrationProvenance supplements the original history proof without
// rewriting historical receipts. Source includes every original reachable commit.
type WorkflowIntegrationProvenance struct {
	Schema            string                           `json:"schema"`
	ObjectFormat      string                           `json:"object_format"`
	AttributionPolicy string                           `json:"attribution_policy"`
	CombinedDiffHash  string                           `json:"combined_diff_hash"`
	Source            []WorkflowIntegrationAttribution `json:"source"`
	Prepared          []WorkflowIntegrationAttribution `json:"prepared"`
}

func (reader workflowIntegrationReader) combinedProvenance(ctx context.Context, proof WorkflowIntegrationProof) (*WorkflowIntegrationProvenance, error) {
	source, err := reader.sourceAttributions(ctx, proof)
	if err != nil {
		return nil, err
	}

	prepared, err := reader.commitAttribution(ctx, proof.ResultCommit)
	if err != nil {
		return nil, err
	}

	format, err := reader.output(ctx, "rev-parse", "--show-object-format")
	if err != nil {
		return nil, err
	}

	diff, err := reader.rawOutput(ctx, "-c", "core.quotePath=true", "-c", "color.ui=false", "diff", "--binary", "--full-index", "--no-ext-diff", "--no-textconv", "--no-renames", "--diff-algorithm=myers", "--no-indent-heuristic", "--unified=3", "--src-prefix=a/", "--dst-prefix=b/", proof.BaseCommit, proof.ResultCommit, "--")
	if err != nil {
		return nil, err
	}

	return &WorkflowIntegrationProvenance{Schema: "chr-integration-provenance/v1", ObjectFormat: format, AttributionPolicy: "host_author_with_original_source_manifest", CombinedDiffHash: workflowHash([]byte(diff)), Source: source, Prepared: []WorkflowIntegrationAttribution{prepared}}, nil
}

func (reader workflowIntegrationReader) verifyCombinedProvenance(ctx context.Context, proof *WorkflowIntegrationProof) error {
	if proof.Provenance == nil {
		return nil
	}

	actual, err := reader.combinedProvenance(ctx, *proof)
	if err != nil {
		return err
	}

	if !workflowNativeJSONEqual(actual, proof.Provenance) {
		return ErrWorkflowConflict
	}

	return nil
}

func validateWorkflowIntegrationProvenance(proof WorkflowIntegrationProof) error {
	p := proof.Provenance
	if p == nil {
		return nil
	}

	if !validWorkflowProvenanceHeader(proof, *p) {
		return ErrWorkflowConflict
	}

	if err := validateWorkflowProvenanceCommits(proof, *p); err != nil {
		return err
	}

	data, err := json.Marshal(p)
	if err != nil || len(data) >= WorkflowFileLimit {
		return ErrWorkflowConflict
	}

	return nil
}

func validWorkflowProvenanceHeader(proof WorkflowIntegrationProof, p WorkflowIntegrationProvenance) bool {
	if p.Schema != "chr-integration-provenance/v1" || proof.Method == "rebase" || p.AttributionPolicy != "host_author_with_original_source_manifest" || !validWorkflowNativeFingerprint(p.CombinedDiffHash) {
		return false
	}

	expectedFormat := "sha1"
	if len(proof.ResultCommit) == 64 {
		expectedFormat = "sha256"
	}

	return p.ObjectFormat == expectedFormat && len(p.Source) > 0 && len(p.Source) <= WorkflowRecordLimit && len(p.Prepared) == 1
}

func validWorkflowAttributedSources(proof WorkflowIntegrationProof, sources []WorkflowIntegrationAttribution) bool {
	seen := map[string]bool{}
	for _, entry := range sources {
		if !validWorkflowIntegrationAttribution(entry) || len(entry.Commit) != len(proof.SourceCommit) || seen[entry.Commit] {
			return false
		}

		seen[entry.Commit] = true
	}

	return sources[len(sources)-1].Commit == proof.SourceCommit
}

func validateWorkflowProvenanceCommits(proof WorkflowIntegrationProof, p WorkflowIntegrationProvenance) error {
	if !validWorkflowAttributedSources(proof, p.Source) {
		return ErrWorkflowConflict
	}

	result := p.Prepared[0]

	parents := []string{proof.BaseCommit}
	if proof.Method == "merge" {
		parents = append(parents, proof.SourceCommit)
	}

	if !validWorkflowIntegrationAttribution(result) || result.Commit != proof.ResultCommit || result.Tree != proof.PreparedTree || !slices.Equal(result.Parents, parents) {
		return ErrWorkflowConflict
	}

	return nil
}
