package workflow

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ClosingMember freezes each actual ticket, including work excluded by
// an invocation's default membership policy.
type ClosingMember struct {
	Identity
	Key        string `json:"key"`
	Path       string `json:"path"`
	SourceHash string `json:"source_hash"`
	SliceID    string `json:"slice_id"`
	Order      int    `json:"task_order"`
}

// ClosingSlice binds ordered descriptors to actual integration evidence.
type ClosingSlice struct {
	ID           string                        `json:"id"`
	Order        int                           `json:"slice_order"`
	Receipt      workflowstate.ClosingEvidence `json:"receipt"`
	Verification workflowstate.ClosingEvidence `json:"verification"`
	Proof        IntegrationProof              `json:"proof"`
}

// ClosingInput is the immutable member/plan/contract/target snapshot.
type ClosingInput struct {
	ValidationTimeoutSeconds int                           `json:"validation_timeout_seconds,omitempty"`
	Schema                   string                        `json:"schema"`
	SetID                    string                        `json:"set_id"`
	SetRevision              uint64                        `json:"set_revision"`
	SetSourceHash            string                        `json:"set_source_hash"`
	ProjectID                string                        `json:"project_id"`
	Namespace                string                        `json:"namespace"`
	Plan                     workflowstate.ClosingEvidence `json:"plan"`
	ContractHash             string                        `json:"closing_contract_hash"`
	Validation               workflowstate.Validation      `json:"closing_validation"`
	Members                  []ClosingMember               `json:"members"`
	Slices                   []ClosingSlice                `json:"slices"`
	Target                   workflowstate.PreviewTarget   `json:"target"`
	Branch                   string                        `json:"branch"`
	Tree                     string                        `json:"tree"`
	EvidenceHashes           map[string]string             `json:"evidence_hashes"`
}

// ClosingReceipt retains exact inputs and host results, including an
// explicit waiver. Only a successful receipt can become the current closure.
type ClosingReceipt struct {
	ClosingInput
	Input       workflowstate.ClosingEvidence `json:"input"`
	Validation  []CommandResult               `json:"validation"`
	Outcome     string                        `json:"outcome"`
	Waiver      string                        `json:"waiver_reason,omitempty"`
	ConfirmedAt time.Time                     `json:"confirmed_at"`
}

func (s *Store) workflowClosingInput(ctx context.Context, set Record, request SetClosureRequest, ports FrontierPorts) (ClosingInput, error) {
	seconds, err := s.workflowClosingValidationTimeout(ctx, request.InvocationID, set.Set.ClosingValidation.TimeoutSeconds)
	if err != nil {
		return ClosingInput{}, err
	}

	target, err := s.observeWorkflowClosingTarget(ctx, request, ports)
	if err != nil {
		return ClosingInput{}, err
	}

	contract, err := json.Marshal(set.Set.ClosingValidation)
	if err != nil {
		return ClosingInput{}, err
	}

	input := ClosingInput{Schema: "chr-set-closing-input/v1", SetID: set.ID, SetRevision: set.Revision, SetSourceHash: set.SourceHash, ProjectID: set.ProjectID, Namespace: set.Namespace, Plan: workflowstate.ClosingEvidence{Ref: set.Set.PlanRef, Hash: set.Set.PlanHash}, Validation: set.Set.ClosingValidation, ContractHash: contenthash.Sum(contract), Target: target, Branch: request.Branch, EvidenceHashes: map[string]string{set.Set.PlanRef: set.Set.PlanHash}}
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

func (s *Store) workflowClosingMembers(ctx context.Context, id string) ([]ClosingMember, error) {
	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return nil, err
	}

	members := []ClosingMember{}

	for _, record := range snapshot.Records {
		if record.Task == nil || record.Task.SetID != id {
			continue
		}

		if record.Task.Phase != "integrated" || record.Task.Status != string(ticketcore.StatusDone) || record.Task.AttemptID != "" {
			return nil, fmt.Errorf("%w: actual member %s requires integration", workflowstate.ErrConflict, record.Task.Key)
		}

		members = append(members, ClosingMember{Identity: record.Identity, Key: record.Task.Key, Path: record.Path, SourceHash: record.SourceHash, SliceID: record.Task.SliceID, Order: record.Task.Order})
	}

	slices.SortFunc(members, func(a, b ClosingMember) int { return strings.Compare(a.Key, b.Key) })

	return members, nil
}
