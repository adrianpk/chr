package core

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowSliceObservation struct {
	descriptor WorkflowSlice
	set        WorkflowRecord
	tasks      []WorkflowRecord
}

func (s *WorkflowStore) observeWorkflowSlice(ctx context.Context, run workflowInvocationFile) (workflowSliceObservation, error) {
	return s.observeWorkflowSlicePrepared(ctx, run, (*WorkflowStore).recoverTransition)
}

func (s *WorkflowStore) observeWorkflowSlicePrepared(ctx context.Context, run workflowInvocationFile, prepare func(*WorkflowStore) error) (workflowSliceObservation, error) {
	observation, err := s.readWorkflowSliceMembershipPrepared(ctx, run, prepare)
	if err != nil {
		return observation, err
	}

	for _, record := range observation.tasks {
		if err := s.checkWorkflowReusableTaskPrepared(ctx, run, record, *run.Lane, prepare); err != nil {
			return observation, err
		}
	}

	return observation, nil
}

func (s *WorkflowStore) readWorkflowSliceMembership(ctx context.Context, run workflowInvocationFile) (workflowSliceObservation, error) {
	return s.readWorkflowSliceMembershipPrepared(ctx, run, (*WorkflowStore).recoverTransition)
}

func (s *WorkflowStore) readWorkflowSliceMembershipPrepared(ctx context.Context, run workflowInvocationFile, prepare func(*WorkflowStore) error) (workflowSliceObservation, error) {
	if run.Lane == nil || run.Cursor < 1 {
		return workflowSliceObservation{}, ErrWorkflowOwned
	}

	if err := s.checkWorkflowInputsPrepared(ctx, run, prepare); err != nil {
		return workflowSliceObservation{}, err
	}

	if run.Lane.Unit.Kind == "ticket" {
		return s.readWorkflowTicketMembership(run)
	}

	if run.Lane.Unit.Kind != "slice" || run.Selection.Selector.Extent == "ticket" {
		return workflowSliceObservation{}, ErrWorkflowOwned
	}

	return s.readWorkflowAggregateMembership(ctx, run)
}

func (s *WorkflowStore) readWorkflowAggregateMembership(ctx context.Context, run workflowInvocationFile) (workflowSliceObservation, error) {
	selection, err := s.readWorkflowSelection(ctx)
	if err != nil {
		return workflowSliceObservation{}, err
	}

	dependencies, err := newWorkflowDependencies(selection)
	if err != nil {
		return workflowSliceObservation{}, err
	}

	observation := workflowSliceObservation{descriptor: selection.slices[run.Lane.Unit.ID], tasks: dependencies.tasks[run.Lane.Unit.ID]}
	if err := checkWorkflowSliceMembership(run, observation); err != nil {
		return observation, err
	}

	observation.set = selection.sets[observation.tasks[0].Task.SetID]
	if observation.descriptor.TaskInputHash != WorkflowTaskInputHash(observation.tasks, observation.descriptor.ID) {
		return observation, fmt.Errorf("%w: aggregate membership changed", workflowstate.ErrConflict)
	}

	return observation, nil
}

func checkWorkflowSliceMembership(run workflowInvocationFile, observation workflowSliceObservation) error {
	start := run.Cursor - len(observation.tasks)
	if len(observation.tasks) == 0 || start < 0 || run.Cursor > len(run.Members) {
		return fmt.Errorf("%w: incomplete aggregate slice", workflowstate.ErrConflict)
	}

	for i, record := range observation.tasks {
		member := run.Members[start+i].Admission
		if member.Key != record.Task.Key || member.SliceID != run.Lane.Unit.ID || !slices.Contains([]string{"validated", "waiting_slice"}, record.Task.Phase) {
			return fmt.Errorf("%w: aggregate requires every ordered task", workflowstate.ErrConflict)
		}
	}

	return nil
}

func (s *WorkflowStore) workflowSliceInput(ctx context.Context, run workflowInvocationFile, record WorkflowRecord, observation workflowSliceObservation) (WorkflowTaskInput, error) {
	input, err := s.workflowTaskInput(ctx, run, record, nil)
	if err != nil {
		return input, err
	}

	if run.Lane.Unit.Kind == "ticket" {
		return s.workflowTicketValidationInput(ctx, run, observation, input)
	}

	contract, err := s.readWorkflowSliceContract(observation.descriptor)
	if err != nil {
		return input, err
	}

	input.ExecutionMode, input.Body = "aggregate", contract.Contract
	input.ContractHash = observation.descriptor.ContractHash
	input.Contract.Validation = contract.Validation
	input.Slice = &WorkflowSliceInput{Unit: run.Lane.Unit, SetID: observation.set.ID, ContractRef: observation.descriptor.ContractRef, ContractHash: observation.descriptor.ContractHash}
	input.EvidenceHashes[observation.descriptor.ContractRef] = observation.descriptor.ContractHash

	for _, task := range observation.tasks {
		proof, err := s.workflowSliceTaskProof(ctx, run, task)
		if err != nil {
			return input, err
		}

		input.Slice.Tasks = append(input.Slice.Tasks, proof)
		input.EvidenceHashes[proof.EvidenceRef] = proof.EvidenceHash
		input.EvidenceHashes[proof.InputRef] = proof.InputHash
		input.Prerequisites = append(input.Prerequisites, WorkflowInputImage{Path: task.Path, Revision: task.Revision, SourceHash: task.SourceHash})
	}

	return input, nil
}

type workflowSliceContract struct {
	Contract   string
	Validation WorkflowValidation
}

func (s *WorkflowStore) readWorkflowSliceContract(descriptor WorkflowSlice) (workflowSliceContract, error) {
	path, err := s.safePath(descriptor.ContractRef)
	if err != nil {
		return workflowSliceContract{}, err
	}

	data, err := readWorkflowFile(path)
	if err != nil || contenthash.Sum(data) != descriptor.ContractHash {
		return workflowSliceContract{}, fmt.Errorf("%w: aggregate contract image changed", workflowstate.ErrConflict)
	}

	var contract workflowSliceContract
	if err := json.Unmarshal(data, &contract); err != nil {
		return contract, err
	}

	return contract, validateWorkflowGate(contract.Validation)
}

func (s *WorkflowStore) workflowSliceTaskProof(ctx context.Context, run workflowInvocationFile, record WorkflowRecord) (WorkflowSliceTaskProof, error) {
	runs, err := s.repositoryWorkflowInvocations(ctx)
	if err != nil {
		return WorkflowSliceTaskProof{}, err
	}

	request := run.ExecutionSelection()
	request.CumulativeBranch, request.CumulativeCommit = run.Lane.Branch, run.Lane.HeadCommit
	reader := workflowProofReader{store: s, request: request}

	for _, previous := range runs {
		for _, attempt := range previous.Attempts {
			if previous.Namespace == run.Namespace && workflowAttemptValidatesTask(attempt, record) && s.workflowRetainedResultMatches(ctx, previous, attempt, record, &reader) {
				return WorkflowSliceTaskProof{Key: record.Task.Key, ContractHash: record.Task.ContractHash, EvidenceRef: attempt.ResultRef, EvidenceHash: attempt.ResultHash, InputRef: attempt.InputRef, InputHash: attempt.InputHash}, nil
			}
		}
	}

	return WorkflowSliceTaskProof{}, workflowstate.ErrConflict
}
