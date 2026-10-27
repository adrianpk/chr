package core

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// UnitInput returns the actual ticket or slice integration contract. Independent
// candidates have a ticket input and never persist a slice or set identity.
func (candidate WorkflowSliceCandidate) UnitInput() WorkflowSliceInput {
	if candidate.Ticket != nil {
		return *candidate.Ticket
	}

	return candidate.Slice
}

func workflowMemberOwnsUnit(member WorkflowMember, unit workflowstate.Ref) bool {
	if unit.Kind == "ticket" {
		return member.Admission.Key == unit.ID && member.Admission.SliceID == "" && member.Admission.SetID == ""
	}

	return unit.Kind == "slice" && member.Admission.SliceID == unit.ID
}

func workflowUnitContractHash(observation workflowSliceObservation) string {
	if observation.set.Set == nil && len(observation.tasks) == 1 {
		return observation.tasks[0].Task.ContractHash
	}

	return observation.descriptor.ContractHash
}

func (s *WorkflowStore) readWorkflowTicketMembership(run workflowInvocationFile) (workflowSliceObservation, error) {
	if run.Cursor > len(run.Members) || !workflowMemberOwnsUnit(run.Members[run.Cursor-1], run.Lane.Unit) {
		return workflowSliceObservation{}, workflowstate.ErrConflict
	}

	record, err := s.readRecord(run.Members[run.Cursor-1].Admission.Path)
	if err != nil {
		return workflowSliceObservation{}, err
	}

	if record.Task == nil || !slices.Contains([]string{"validated", "waiting_slice", "integrated"}, record.Task.Phase) {
		return workflowSliceObservation{}, workflowstate.ErrConflict
	}

	return workflowSliceObservation{tasks: []WorkflowRecord{record}}, nil
}

// PrepareWorkflowTicketCandidate binds existing host-accepted task evidence to
// its integration unit. It performs no worker execution or additional source
// gate. Review, combined-result validation and delivery use the shared lifecycle.
func (s *WorkflowStore) PrepareWorkflowTicketCandidate(ctx context.Context, id string) (candidate WorkflowSliceCandidate, err error) {
	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		if run.Lane == nil || run.Lane.Unit.Kind != "ticket" || run.StopIntent != "" || !workflowSliceValidationReady(run.WorkflowInvocation) {
			return ErrWorkflowOwned
		}

		observation, readErr := owner.observeWorkflowSlice(ctx, run)
		if readErr != nil {
			return readErr
		}

		if run.CandidateRef != "" {
			candidate, readErr = owner.readWorkflowSliceCandidate(ctx, run, observation)

			return readErr
		}

		candidate, readErr = owner.workflowTicketCandidate(ctx, run, observation)
		if readErr != nil {
			return readErr
		}

		return owner.commitWorkflowTicketCandidate(&run, candidate)
	})

	return candidate, err
}

func (s *WorkflowStore) workflowTicketCandidate(ctx context.Context, run workflowInvocationFile, observation workflowSliceObservation) (WorkflowSliceCandidate, error) {
	proof, err := s.workflowSliceTaskProof(ctx, run, observation.tasks[0])
	if err != nil {
		return WorkflowSliceCandidate{}, err
	}

	data, err := s.readWorkflowReviewImage(proof.EvidenceRef, proof.EvidenceHash)
	if err != nil {
		return WorkflowSliceCandidate{}, err
	}

	var result WorkflowTaskResult
	if err := json.Unmarshal(data, &result); err != nil {
		return WorkflowSliceCandidate{}, err
	}

	input := WorkflowSliceInput{Unit: run.Lane.Unit, ContractHash: observation.tasks[0].Task.ContractHash, Tasks: []WorkflowSliceTaskProof{proof}}
	candidate := WorkflowSliceCandidate{Schema: "chr-ticket-candidate/v1", ID: workflowTicketCandidateID(run.ID, proof), InvocationID: run.ID, Ticket: &input, Target: run.ExecutionSelection().Target, Lane: *run.Lane, Result: result}

	return candidate, s.checkWorkflowTicketCandidate(ctx, run, observation, candidate)
}

func workflowTicketCandidateID(id string, proof WorkflowSliceTaskProof) string {
	return contenthash.Sum([]byte(id + "\x00" + proof.EvidenceHash + "\x00" + proof.InputHash))
}

func workflowTicketCandidatePath(id, candidateID string) string {
	return filepath.ToSlash(filepath.Join("run", id, "extension", "chr", "ticket-candidate", candidateID+".json"))
}

func validWorkflowTicketCandidateRef(run workflowInvocationFile) bool {
	id := strings.TrimSuffix(filepath.Base(run.CandidateRef), ".json")

	return run.Lane != nil && run.Lane.Unit.Kind == "ticket" && slices.ContainsFunc(run.Members, func(member WorkflowMember) bool { return workflowMemberOwnsUnit(member, run.Lane.Unit) }) && len(run.CandidateHash) == 64 && len(id) == 64 && run.CandidateRef == workflowTicketCandidatePath(run.ID, id)
}

func workflowAttemptOwnsCandidate(run workflowInvocationFile, attempt WorkflowAttempt) bool {
	return attempt.Purpose == workflowSliceValidationPurpose && attempt.Outcome == "validated" && attempt.ResultRef == run.CandidateRef && attempt.ResultHash == run.CandidateHash && len(run.CandidateHash) == 64
}

func (s *WorkflowStore) commitWorkflowTicketCandidate(run *workflowInvocationFile, candidate WorkflowSliceCandidate) error {
	data, err := json.MarshalIndent(candidate, "", "  ")
	if err != nil || len(data) >= workflowstate.FileLimit {
		return fmt.Errorf("ticket candidate exceeds file bounds: %w", errorsOrConflict(err))
	}

	data = append(data, '\n')
	run.CandidateRef, run.CandidateHash = workflowTicketCandidatePath(run.ID, candidate.ID), contenthash.Sum(data)
	run.Outcome = "ticket_candidate_ready"

	return s.commitWorkflowPullRequestFiles(run, nil, workflowFileChange{Path: run.CandidateRef, After: data})
}

func (s *WorkflowStore) checkWorkflowTicketCandidate(ctx context.Context, run workflowInvocationFile, observation workflowSliceObservation, candidate WorkflowSliceCandidate) error {
	if !workflowTicketCandidateMatchesRun(run, observation, candidate) {
		return workflowstate.ErrConflict
	}

	proof, err := s.workflowSliceTaskProof(ctx, run, observation.tasks[0])
	if err != nil || proof != candidate.Ticket.Tasks[0] || candidate.ID != workflowTicketCandidateID(run.ID, proof) {
		return workflowstate.ErrConflict
	}

	data, err := s.readWorkflowReviewImage(proof.EvidenceRef, proof.EvidenceHash)
	if err != nil {
		return err
	}

	var result WorkflowTaskResult
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}

	return checkWorkflowTicketResult(run, proof, result, candidate.Result)
}

func checkWorkflowTicketResult(run workflowInvocationFile, proof WorkflowSliceTaskProof, result, candidate WorkflowTaskResult) error {
	expected, err := json.Marshal(result)
	if err != nil {
		return err
	}

	actual, err := json.Marshal(candidate)
	if err != nil || string(actual) != string(expected) || !result.Successful || result.Commit != run.Lane.HeadCommit || result.InputRef != proof.InputRef || result.InputHash != proof.InputHash {
		return workflowstate.ErrConflict
	}

	return nil
}

func workflowTicketCandidateMatchesRun(run workflowInvocationFile, observation workflowSliceObservation, candidate WorkflowSliceCandidate) bool {
	if candidate.Schema != "chr-ticket-candidate/v1" || candidate.InvocationID != run.ID || candidate.Target != run.ExecutionSelection().Target || candidate.Lane != *run.Lane || candidate.Ticket == nil || len(observation.tasks) != 1 {
		return false
	}

	return workflowTicketCandidateInputMatches(run.Lane.Unit, observation.tasks[0].Task.ContractHash, candidate)
}

func workflowTicketCandidateInputMatches(unit workflowstate.Ref, contractHash string, candidate WorkflowSliceCandidate) bool {
	input := candidate.Ticket

	return input.Unit == unit && input.SetID == "" && input.ContractRef == "" && input.ContractHash == contractHash && len(input.Tasks) == 1 && workflowCandidateHasNoSlice(candidate.Slice)
}

func workflowCandidateHasNoSlice(input WorkflowSliceInput) bool {
	return input.Unit == (workflowstate.Ref{}) && len(input.Tasks) == 0 && input.SetID == "" && input.ContractHash == "" && input.ContractRef == ""
}

func (s *WorkflowStore) workflowTicketValidationInput(ctx context.Context, run workflowInvocationFile, observation workflowSliceObservation, input WorkflowTaskInput) (WorkflowTaskInput, error) {
	proof, err := s.workflowSliceTaskProof(ctx, run, observation.tasks[0])
	if err != nil {
		return input, err
	}

	input.ExecutionMode = "aggregate"
	input.Slice = &WorkflowSliceInput{Unit: run.Lane.Unit, ContractHash: observation.tasks[0].Task.ContractHash, Tasks: []WorkflowSliceTaskProof{proof}}
	input.EvidenceHashes[proof.EvidenceRef], input.EvidenceHashes[proof.InputRef] = proof.EvidenceHash, proof.InputHash
	input.Prerequisites = append(input.Prerequisites, WorkflowInputImage{Path: observation.tasks[0].Path, Revision: observation.tasks[0].Revision, SourceHash: observation.tasks[0].SourceHash})

	return input, nil
}
