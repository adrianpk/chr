package closeout

import (
	"cmp"
	"context"
	"encoding/json"
	"slices"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// BuildInput freezes the original contract and ordered actual delivery evidence.
func BuildInput(ctx context.Context, set workflowstate.Record, request Request, source InputSource) (workflowstate.ClosingInput, error) {
	seconds, err := source.ValidationTimeout(ctx, request.InvocationID, set.Set.ClosingValidation.TimeoutSeconds)
	if err != nil {
		return workflowstate.ClosingInput{}, err
	}

	target, err := source.Target(ctx, request)
	if err != nil {
		return workflowstate.ClosingInput{}, err
	}

	contract, err := json.Marshal(set.Set.ClosingValidation)
	if err != nil {
		return workflowstate.ClosingInput{}, err
	}

	input := workflowstate.ClosingInput{Schema: "chr-set-closing-input/v1", SetID: set.ID, SetRevision: set.Revision, SetSourceHash: set.SourceHash, ProjectID: set.ProjectID, Namespace: set.Namespace, Plan: workflowstate.ClosingEvidence{Ref: set.Set.PlanRef, Hash: set.Set.PlanHash}, Validation: set.Set.ClosingValidation, ContractHash: contenthash.Sum(contract), Target: target, Branch: request.Branch, EvidenceHashes: map[string]string{set.Set.PlanRef: set.Set.PlanHash}}
	input.ValidationTimeoutSeconds = seconds

	input.Tree, err = source.Tree(ctx, target.Commit)
	if err != nil {
		return input, err
	}

	snapshot, err := source.Snapshot(ctx)
	if err != nil {
		return input, err
	}

	input.Members, err = Members(snapshot.Records, set.ID)
	if err != nil {
		return input, err
	}

	descriptors := slices.Clone(set.Set.Slices)
	slices.SortFunc(descriptors, func(a, b workflowstate.WorkflowSlice) int { return cmp.Compare(a.Order, b.Order) })

	for _, descriptor := range descriptors {
		proof, err := source.Slice(ctx, input, descriptor)
		if err != nil {
			return input, err
		}

		input.Slices = append(input.Slices, proof)
	}

	return input, source.Check(input.EvidenceHashes)
}
