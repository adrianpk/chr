package workflow

import (
	"context"
	"encoding/json"
	"slices"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
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

	if !workflowintegrate.ValidWorkflowProvenanceHeader(proof, *p) {
		return workflowstate.ErrConflict
	}

	if err := workflowintegrate.ValidateWorkflowProvenanceCommits(proof, *p); err != nil {
		return err
	}

	if p.Preparation != nil && !workflowintegrate.ValidWorkflowPreparationMetadata(proof, *p.Preparation) {
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
