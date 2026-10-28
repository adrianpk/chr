package workflow

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) workflowClosingInput(ctx context.Context, set Record, request SetClosureRequest, ports FrontierPorts) (workflowstate.ClosingInput, error) {
	seconds, err := s.workflowClosingValidationTimeout(ctx, request.InvocationID, set.Set.ClosingValidation.TimeoutSeconds)
	if err != nil {
		return workflowstate.ClosingInput{}, err
	}

	target, err := s.observeWorkflowClosingTarget(ctx, request, ports)
	if err != nil {
		return workflowstate.ClosingInput{}, err
	}

	contract, err := json.Marshal(set.Set.ClosingValidation)
	if err != nil {
		return workflowstate.ClosingInput{}, err
	}

	input := workflowstate.ClosingInput{Schema: "chr-set-closing-input/v1", SetID: set.ID, SetRevision: set.Revision, SetSourceHash: set.SourceHash, ProjectID: set.ProjectID, Namespace: set.Namespace, Plan: workflowstate.ClosingEvidence{Ref: set.Set.PlanRef, Hash: set.Set.PlanHash}, Validation: set.Set.ClosingValidation, ContractHash: contenthash.Sum(contract), Target: target, Branch: request.Branch, EvidenceHashes: map[string]string{set.Set.PlanRef: set.Set.PlanHash}}
	input.ValidationTimeoutSeconds = seconds
	reader := workflowProofReader{store: s}

	input.Tree, err = reader.gitOutput(ctx, "rev-parse", "--verify", target.Commit+"^{tree}")
	if err != nil {
		return input, err
	}

	input.Members, err = s.workflowClosingMembers(ctx, set.ID)
	if err != nil {
		return input, err
	}

	descriptors := slices.Clone(set.Set.Slices)
	slices.SortFunc(descriptors, func(a, b workflowstate.WorkflowSlice) int { return cmp.Compare(a.Order, b.Order) })

	for _, descriptor := range descriptors {
		proof, err := s.workflowClosingSliceProof(ctx, input, descriptor, ports)
		if err != nil {
			return input, err
		}

		input.Slices = append(input.Slices, proof)
	}

	return input, s.checkWorkflowPrerequisiteEvidence(input.EvidenceHashes)
}

func (s *Store) workflowClosingMembers(ctx context.Context, id string) ([]workflowstate.ClosingMember, error) {
	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return nil, err
	}

	members := []workflowstate.ClosingMember{}

	for _, record := range snapshot.Records {
		if record.Task == nil || record.Task.SetID != id {
			continue
		}

		if record.Task.Phase != "integrated" || record.Task.Status != string(ticketcore.StatusDone) || record.Task.AttemptID != "" {
			return nil, fmt.Errorf("%w: actual member %s requires integration", workflowstate.ErrConflict, record.Task.Key)
		}

		members = append(members, workflowstate.ClosingMember{Identity: record.Identity, Key: record.Task.Key, Path: record.Path, SourceHash: record.SourceHash, SliceID: record.Task.SliceID, Order: record.Task.Order})
	}

	slices.SortFunc(members, func(a, b workflowstate.ClosingMember) int { return strings.Compare(a.Key, b.Key) })

	return members, nil
}
