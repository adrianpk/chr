package core

import (
	"context"
	"encoding/json"
	"slices"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
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
	Replays           []WorkflowIntegrationReplay      `json:"replays,omitempty"`
	IntegrationFixes  []WorkflowIntegrationFix         `json:"integration_fixes,omitempty"`
	Repairs           []WorkflowIntegrationTreeRepair  `json:"repairs,omitempty"`
	Preparation       *WorkflowIntegrationPreparation  `json:"preparation,omitempty"`
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

	p := &WorkflowIntegrationProvenance{Schema: "chr-integration-provenance/v1", ObjectFormat: format, AttributionPolicy: "host_author_with_original_source_manifest", CombinedDiffHash: contenthash.Sum([]byte(diff)), Source: source, Prepared: []WorkflowIntegrationAttribution{prepared}}
	if proof.Provenance != nil {
		p.Repairs = slices.Clone(proof.Provenance.Repairs)
	}

	return p, nil
}

func (reader workflowIntegrationReader) verifyCombinedProvenance(ctx context.Context, proof *WorkflowIntegrationProof) error {
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

func validateWorkflowIntegrationProvenance(proof WorkflowIntegrationProof) error {
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
	if err != nil || len(data) >= workflowstate.WorkflowFileLimit {
		return workflowstate.ErrConflict
	}

	return nil
}

func validWorkflowProvenanceHeader(proof WorkflowIntegrationProof, p WorkflowIntegrationProvenance) bool {
	if !validWorkflowProvenanceEnvelope(proof, p) || !validWorkflowProvenanceCount(p) {
		return false
	}

	if proof.Method == "rebase" {
		return p.AttributionPolicy == "preserve_source_authorship" && len(p.Replays) == len(p.Source) && len(p.Prepared) == len(p.Source)+len(p.IntegrationFixes)
	}

	return p.AttributionPolicy == "host_author_with_original_source_manifest" && len(p.Prepared) == 1 && len(p.Replays) == 0 && len(p.IntegrationFixes) == 0
}

func validWorkflowProvenanceEnvelope(proof WorkflowIntegrationProof, p WorkflowIntegrationProvenance) bool {
	expectedFormat := "sha1"
	if len(proof.ResultCommit) == 64 {
		expectedFormat = "sha256"
	}

	return p.Schema == "chr-integration-provenance/v1" && validWorkflowNativeFingerprint(p.CombinedDiffHash) && p.ObjectFormat == expectedFormat
}

func validWorkflowProvenanceCount(p WorkflowIntegrationProvenance) bool {
	return len(p.Source) > 0 && len(p.Source) <= workflowstate.WorkflowRecordLimit && len(p.Prepared) > 0 && len(p.Prepared) <= workflowstate.WorkflowRecordLimit
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
