package workflow

import (
	"context"
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
