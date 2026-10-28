package workflow

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
func (candidate SliceCandidate) UnitInput() workflowstate.SliceInput {
	if candidate.Ticket != nil {
		return *candidate.Ticket
	}

	return candidate.Slice
}

func workflowMemberOwnsUnit(member Member, unit workflowstate.Ref) bool {
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

func (s *Store) readWorkflowTicketMembership(run workflowInvocationFile) (workflowSliceObservation, error) {
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

	return workflowSliceObservation{tasks: []Record{record}}, nil
}

// PrepareWorkflowTicketCandidate binds existing host-accepted task evidence to
// its integration unit. It performs no worker execution or additional source
// gate. Review, combined-result validation and delivery use the shared lifecycle.
func (s *Store) PrepareWorkflowTicketCandidate(ctx context.Context, id string) (candidate SliceCandidate, err error) {
	err = s.withLock(ctx, func() error {
		invocationLocation, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		if invocationLocation.Run.Lane == nil || invocationLocation.Run.Lane.Unit.Kind != "ticket" || invocationLocation.Run.StopIntent != "" || !workflowSliceValidationReady(invocationLocation.Run.Invocation) {
			return ErrWorkflowOwned
		}

		observation, readErr := invocationLocation.Owner.observeWorkflowSlice(ctx, invocationLocation.Run)
		if readErr != nil {
			return readErr
		}

		if invocationLocation.Run.CandidateRef != "" {
			candidate, readErr = invocationLocation.Owner.readWorkflowSliceCandidate(ctx, invocationLocation.Run, observation)

			return readErr
		}

		candidate, readErr = invocationLocation.Owner.workflowTicketCandidate(ctx, invocationLocation.Run, observation)
		if readErr != nil {
			return readErr
		}

		return invocationLocation.Owner.commitWorkflowTicketCandidate(&invocationLocation.Run, candidate)
	})

	return candidate, err
}

func (s *Store) workflowTicketCandidate(ctx context.Context, run workflowInvocationFile, observation workflowSliceObservation) (SliceCandidate, error) {
	proof, err := s.workflowSliceTaskProof(ctx, run, observation.tasks[0])
	if err != nil {
		return SliceCandidate{}, err
	}

	data, err := s.readWorkflowReviewImage(proof.EvidenceRef, proof.EvidenceHash)
	if err != nil {
		return SliceCandidate{}, err
	}

	var result workflowstate.TaskResult
	if err := json.Unmarshal(data, &result); err != nil {
		return SliceCandidate{}, err
	}

	input := workflowstate.SliceInput{Unit: run.Lane.Unit, ContractHash: observation.tasks[0].Task.ContractHash, Tasks: []workflowstate.SliceTaskProof{proof}}
	candidate := SliceCandidate{Schema: "chr-ticket-candidate/v1", ID: workflowTicketCandidateID(run.ID, proof), InvocationID: run.ID, Ticket: &input, Target: run.ExecutionSelection().Target, Lane: *run.Lane, Result: result}

	return candidate, s.checkWorkflowTicketCandidate(ctx, run, observation, candidate)
}

func workflowTicketCandidateID(id string, proof workflowstate.SliceTaskProof) string {
	return contenthash.Sum([]byte(id + "\x00" + proof.EvidenceHash + "\x00" + proof.InputHash))
}

func workflowTicketCandidatePath(id, candidateID string) string {
	return filepath.ToSlash(filepath.Join("run", id, "extension", "chr", "ticket-candidate", candidateID+".json"))
}

func validWorkflowTicketCandidateRef(run workflowInvocationFile) bool {
	id := strings.TrimSuffix(filepath.Base(run.CandidateRef), ".json")

	return run.Lane != nil && run.Lane.Unit.Kind == "ticket" && slices.ContainsFunc(run.Members, func(member Member) bool { return workflowMemberOwnsUnit(member, run.Lane.Unit) }) && len(run.CandidateHash) == 64 && len(id) == 64 && run.CandidateRef == workflowTicketCandidatePath(run.ID, id)
}

func workflowAttemptOwnsCandidate(run workflowInvocationFile, attempt Attempt) bool {
	return attempt.Purpose == workflowSliceValidationPurpose && attempt.Outcome == "validated" && attempt.ResultRef == run.CandidateRef && attempt.ResultHash == run.CandidateHash && len(run.CandidateHash) == 64
}

func (s *Store) commitWorkflowTicketCandidate(run *workflowInvocationFile, candidate SliceCandidate) error {
	data, err := json.MarshalIndent(candidate, "", "  ")
	if err != nil || len(data) >= workflowstate.FileLimit {
		return fmt.Errorf("ticket candidate exceeds file bounds: %w", errorsOrConflict(err))
	}

	data = append(data, '\n')
	run.CandidateRef, run.CandidateHash = workflowTicketCandidatePath(run.ID, candidate.ID), contenthash.Sum(data)
	run.Outcome = "ticket_candidate_ready"

	return s.commitWorkflowPullRequestFiles(run, nil, workflowstate.FileChange{Path: run.CandidateRef, After: data})
}

func (s *Store) checkWorkflowTicketCandidate(ctx context.Context, run workflowInvocationFile, observation workflowSliceObservation, candidate SliceCandidate) error {
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

	var result workflowstate.TaskResult
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}

	return checkWorkflowTicketResult(run, proof, result, candidate.Result)
}

func checkWorkflowTicketResult(run workflowInvocationFile, proof workflowstate.SliceTaskProof, result, candidate workflowstate.TaskResult) error {
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

func workflowTicketCandidateMatchesRun(run workflowInvocationFile, observation workflowSliceObservation, candidate SliceCandidate) bool {
	if candidate.Schema != "chr-ticket-candidate/v1" || candidate.InvocationID != run.ID || candidate.Target != run.ExecutionSelection().Target || candidate.Lane != *run.Lane || candidate.Ticket == nil || len(observation.tasks) != 1 {
		return false
	}

	return workflowTicketCandidateInputMatches(run.Lane.Unit, observation.tasks[0].Task.ContractHash, candidate)
}

func workflowTicketCandidateInputMatches(unit workflowstate.Ref, contractHash string, candidate SliceCandidate) bool {
	input := candidate.Ticket

	return input.Unit == unit && input.SetID == "" && input.ContractRef == "" && input.ContractHash == contractHash && len(input.Tasks) == 1 && workflowCandidateHasNoSlice(candidate.Slice)
}

func workflowCandidateHasNoSlice(input workflowstate.SliceInput) bool {
	return input.Unit == (workflowstate.Ref{}) && len(input.Tasks) == 0 && input.SetID == "" && input.ContractHash == "" && input.ContractRef == ""
}

func (s *Store) workflowTicketValidationInput(ctx context.Context, run workflowInvocationFile, observation workflowSliceObservation, input TaskInput) (TaskInput, error) {
	proof, err := s.workflowSliceTaskProof(ctx, run, observation.tasks[0])
	if err != nil {
		return input, err
	}

	input.ExecutionMode = "aggregate"
	input.Slice = &workflowstate.SliceInput{Unit: run.Lane.Unit, ContractHash: observation.tasks[0].Task.ContractHash, Tasks: []workflowstate.SliceTaskProof{proof}}
	input.EvidenceHashes[proof.EvidenceRef], input.EvidenceHashes[proof.InputRef] = proof.EvidenceHash, proof.InputHash
	input.Prerequisites = append(input.Prerequisites, InputImage{Path: observation.tasks[0].Path, Revision: observation.tasks[0].Revision, SourceHash: observation.tasks[0].SourceHash})

	return input, nil
}
