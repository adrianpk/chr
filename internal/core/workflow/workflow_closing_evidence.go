package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowClosingFile(id string, generation int, name string) string {
	return filepath.ToSlash(filepath.Join("plan", id, "extension", "chr", "closing", fmt.Sprintf("%04d", generation), name))
}

func (s *Store) workflowClosingRead(relative string) ([]byte, error) {
	path, err := s.safePath(relative)
	if err != nil {
		return nil, err
	}

	return opsfile.Read(path)
}

func (s *Store) workflowClosingEvidenceHashes(hashes map[string]string, refs []string) error {
	for _, ref := range refs {
		if len(hashes) >= workflowstate.RecordLimit {
			return fmt.Errorf("closing evidence observation bound exceeded")
		}

		hash, err := s.workflowEvidenceHash(ref)
		if err != nil || hash == "missing" {
			return errorsOrConflict(err)
		}

		hashes[ref] = hash
	}

	return nil
}

func (s *Store) checkWorkflowClosingInputs(ctx context.Context, claim *workflowClosingClaim) error {
	if err := s.checkWorkflowClosingPublications(ctx); err != nil {
		return err
	}

	if err := s.checkWorkflowCoordinatorTarget(ctx, claim.input.Target); err != nil {
		return err
	}

	set, err := s.readRecord(claim.set.Path)
	if err != nil {
		return err
	}

	if set.Revision != claim.set.Revision || set.SourceHash != claim.set.SourceHash {
		return workflowstate.ErrConflict
	}

	if err := s.checkWorkflowClosingMembers(ctx, claim, set); err != nil {
		return err
	}

	refs := claim.input.EvidenceHashes
	if err := s.checkWorkflowPrerequisiteEvidence(refs); err != nil {
		return err
	}

	data, err := s.workflowClosingRead(claim.state.Input.Ref)
	if err != nil || contenthash.Sum(data) != claim.state.Input.Hash {
		return errorsOrConflict(err)
	}

	return nil
}

func (s *Store) checkWorkflowClosingMembers(ctx context.Context, claim *workflowClosingClaim, set Record) error {
	members, err := s.workflowClosingMembers(ctx, set.ID)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(members, claim.input.Members) {
		return fmt.Errorf("%w: actual closing members changed", workflowstate.ErrConflict)
	}

	return s.checkWorkflowIntegrationSetReceipts(ctx, set)
}

func (s *Store) readWorkflowClosingAttempt(set Record, index int) (workflowClosingAttempt, error) {
	binding := set.Set.Closing.Attempts[index]

	data, err := s.workflowClosingRead(binding.Ref)
	if err != nil || contenthash.Sum(data) != binding.Hash {
		return workflowClosingAttempt{}, errorsOrConflict(err)
	}

	var state workflowClosingAttempt
	if err := json.Unmarshal(data, &state); err != nil {
		return state, err
	}

	return state, validateWorkflowClosingAttempt(set, index, state)
}

func (s *Store) readWorkflowClosedSet(set Record) (SetClosureResult, error) {
	if set.Set.Closing == nil || len(set.Set.Closing.Attempts) == 0 {
		return SetClosureResult{}, workflowstate.ErrConflict
	}

	state, err := s.readWorkflowClosingAttempt(set, len(set.Set.Closing.Attempts)-1)
	if err != nil || state.Outcome != "completed" || state.Result.Ref != set.Set.ClosureRef {
		return SetClosureResult{}, errorsOrConflict(err)
	}

	receipt, err := s.readWorkflowClosingReceipt(state)
	if err != nil {
		return SetClosureResult{}, err
	}

	return SetClosureResult{Set: set, Receipt: receipt, ReceiptRef: state.Result.Ref, Outcome: state.Outcome}, nil
}

func (s *Store) readWorkflowClosingReceipt(state workflowClosingAttempt) (workflowstate.ClosingReceipt, error) {
	data, err := s.workflowClosingRead(state.Result.Ref)
	if err != nil || contenthash.Sum(data) != state.Result.Hash {
		return workflowstate.ClosingReceipt{}, errorsOrConflict(err)
	}

	var receipt workflowstate.ClosingReceipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		return receipt, err
	}

	if receipt.Schema != "chr-set-closure/v1" || receipt.SetID != state.SetID || receipt.Outcome != state.Outcome || receipt.Input != state.Input || receipt.ConfirmedAt != state.FinishedAt {
		return receipt, workflowstate.ErrConflict
	}

	return receipt, nil
}

func (s *Store) checkWorkflowClosingResult(state workflowClosingAttempt) error {
	if state.Result.Ref == "" {
		return nil
	}

	data, err := s.workflowClosingRead(state.Input.Ref)
	if err != nil {
		return err
	}

	var input workflowstate.ClosingInput
	if err := json.Unmarshal(data, &input); err != nil {
		return err
	}

	receipt, err := s.readWorkflowClosingReceipt(state)
	if err != nil {
		return err
	}

	retained := receipt.ClosingInput

	retained.Schema = input.Schema
	if input.Schema != "chr-set-closing-input/v1" || !reflect.DeepEqual(retained, input) || receipt.Waiver != input.Validation.WaiverReason {
		return workflowstate.ErrConflict
	}

	return validateWorkflowClosingResults(input.Validation, receipt)
}

func validateWorkflowClosingResults(gate workflowstate.Validation, receipt workflowstate.ClosingReceipt) error {
	if len(receipt.Validation) > len(gate.Commands) || receipt.Outcome == "completed" && len(receipt.Validation) != len(gate.Commands) {
		return workflowstate.ErrConflict
	}

	for i, result := range receipt.Validation {
		if !reflect.DeepEqual(result.Command, gate.Commands[i]) || receipt.Outcome == "completed" && result.ExitCode != 0 {
			return workflowstate.ErrConflict
		}
	}

	return nil
}
