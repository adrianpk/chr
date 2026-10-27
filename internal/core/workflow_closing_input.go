package core

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowClosingMember freezes each actual ticket, including work excluded by
// an invocation's default membership policy.
type WorkflowClosingMember struct {
	WorkflowIdentity
	Key        string `json:"key"`
	Path       string `json:"path"`
	SourceHash string `json:"source_hash"`
	SliceID    string `json:"slice_id"`
	Order      int    `json:"task_order"`
}

// WorkflowClosingSlice binds ordered descriptors to actual integration evidence.
type WorkflowClosingSlice struct {
	ID           string                   `json:"id"`
	Order        int                      `json:"slice_order"`
	Receipt      WorkflowClosingEvidence  `json:"receipt"`
	Verification WorkflowClosingEvidence  `json:"verification"`
	Proof        WorkflowIntegrationProof `json:"proof"`
}

// WorkflowClosingInput is the immutable member/plan/contract/target snapshot.
type WorkflowClosingInput struct {
	ValidationTimeoutSeconds int                      `json:"validation_timeout_seconds,omitempty"`
	Schema                   string                   `json:"schema"`
	SetID                    string                   `json:"set_id"`
	SetRevision              uint64                   `json:"set_revision"`
	SetSourceHash            string                   `json:"set_source_hash"`
	ProjectID                string                   `json:"project_id"`
	Namespace                string                   `json:"namespace"`
	Plan                     WorkflowClosingEvidence  `json:"plan"`
	ContractHash             string                   `json:"closing_contract_hash"`
	Validation               workflowstate.Validation `json:"closing_validation"`
	Members                  []WorkflowClosingMember  `json:"members"`
	Slices                   []WorkflowClosingSlice   `json:"slices"`
	Target                   WorkflowPreviewTarget    `json:"target"`
	Branch                   string                   `json:"branch"`
	Tree                     string                   `json:"tree"`
	EvidenceHashes           map[string]string        `json:"evidence_hashes"`
}

// WorkflowClosingReceipt retains exact inputs and host results, including an
// explicit waiver. Only a successful receipt can become the current closure.
type WorkflowClosingReceipt struct {
	WorkflowClosingInput
	Input       WorkflowClosingEvidence `json:"input"`
	Validation  []WorkflowCommandResult `json:"validation"`
	Outcome     string                  `json:"outcome"`
	Waiver      string                  `json:"waiver_reason,omitempty"`
	ConfirmedAt time.Time               `json:"confirmed_at"`
}

func (s *WorkflowStore) workflowClosingInput(ctx context.Context, set WorkflowRecord, request WorkflowSetClosureRequest, ports WorkflowFrontierPorts) (WorkflowClosingInput, error) {
	seconds, err := s.workflowClosingValidationTimeout(ctx, request.InvocationID, set.Set.ClosingValidation.TimeoutSeconds)
	if err != nil {
		return WorkflowClosingInput{}, err
	}

	target, err := s.observeWorkflowClosingTarget(ctx, request, ports)
	if err != nil {
		return WorkflowClosingInput{}, err
	}

	contract, err := json.Marshal(set.Set.ClosingValidation)
	if err != nil {
		return WorkflowClosingInput{}, err
	}

	input := WorkflowClosingInput{Schema: "chr-set-closing-input/v1", SetID: set.ID, SetRevision: set.Revision, SetSourceHash: set.SourceHash, ProjectID: set.ProjectID, Namespace: set.Namespace, Plan: WorkflowClosingEvidence{Ref: set.Set.PlanRef, Hash: set.Set.PlanHash}, Validation: set.Set.ClosingValidation, ContractHash: contenthash.Sum(contract), Target: target, Branch: request.Branch, EvidenceHashes: map[string]string{set.Set.PlanRef: set.Set.PlanHash}}
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
	slices.SortFunc(descriptors, func(a, b WorkflowSlice) int { return compareRank(a.Order, b.Order) })

	for _, descriptor := range descriptors {
		proof, err := s.workflowClosingSliceProof(ctx, input, descriptor, ports)
		if err != nil {
			return input, err
		}

		input.Slices = append(input.Slices, proof)
	}

	return input, s.checkWorkflowPrerequisiteEvidence(input.EvidenceHashes)
}

func (s *WorkflowStore) workflowClosingMembers(ctx context.Context, id string) ([]WorkflowClosingMember, error) {
	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return nil, err
	}

	members := []WorkflowClosingMember{}

	for _, record := range snapshot.Records {
		if record.Task == nil || record.Task.SetID != id {
			continue
		}

		if record.Task.Phase != "integrated" || record.Task.Status != TicketStatusDone || record.Task.AttemptID != "" {
			return nil, fmt.Errorf("%w: actual member %s requires integration", workflowstate.ErrConflict, record.Task.Key)
		}

		members = append(members, WorkflowClosingMember{WorkflowIdentity: record.WorkflowIdentity, Key: record.Task.Key, Path: record.Path, SourceHash: record.SourceHash, SliceID: record.Task.SliceID, Order: record.Task.Order})
	}

	slices.SortFunc(members, func(a, b WorkflowClosingMember) int { return strings.Compare(a.Key, b.Key) })

	return members, nil
}
