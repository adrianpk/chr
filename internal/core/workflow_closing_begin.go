package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

func (s *WorkflowStore) beginWorkflowClosing(ctx context.Context, request WorkflowSetClosureRequest, ports WorkflowFrontierPorts) (claim *workflowClosingClaim, result WorkflowSetClosureResult, err error) {
	set, err := s.readRecord(request.SetPath)
	if err != nil {
		return nil, result, err
	}

	if !s.workflowClosingSetMatches(set, request) {
		return nil, result, ErrWorkflowConflict
	}

	if err := s.checkWorkflowSetAmendmentEvidence(set); err != nil {
		return nil, result, err
	}

	if set.Set.Status == "completed" {
		result, err = s.readWorkflowClosedSet(set)

		return nil, result, err
	}

	if err := s.workflowClosingReady(ctx, set, request); err != nil {
		return nil, result, err
	}

	worker, err := s.reserveWorkflowClaim(ctx, workflowClosingReservation(set))
	if err != nil {
		return nil, result, err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, worker.releaseReservation())
		}
	}()

	claim, err = s.prepareWorkflowClosing(ctx, set, request, ports, worker)
	if err != nil {
		return nil, result, err
	}

	err = s.commitWorkflowClosingStart(ctx, claim)

	return claim, result, err
}

func (s *WorkflowStore) workflowClosingSetMatches(set WorkflowRecord, request WorkflowSetClosureRequest) bool {
	return set.Set != nil && set.ProjectID == s.projectID && set.Namespace == s.layout.Namespace() && set.Revision == request.ExpectedRevision && set.SourceHash == request.ExpectedSourceHash
}

func workflowClosingReservation(set WorkflowRecord) workflowInvocationFile {
	count := 0
	if set.Set.Closing != nil {
		count = len(set.Set.Closing.Attempts)
	}

	return workflowInvocationFile{WorkflowInvocation: WorkflowInvocation{WorkflowIdentity: WorkflowIdentity{ID: workflowClosingOwnerID(set)}, Attempts: make([]WorkflowAttempt, count)}}
}

func (s *WorkflowStore) prepareWorkflowClosing(ctx context.Context, set WorkflowRecord, request WorkflowSetClosureRequest, ports WorkflowFrontierPorts, worker *WorkflowClaim) (*workflowClosingClaim, error) {
	input, err := s.workflowClosingInput(ctx, set, request, ports)
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(input, "", "  ")
	if err != nil || len(data) >= WorkflowFileLimit {
		return nil, fmt.Errorf("closing input exceeds file bounds: %w", errorsOrConflict(err))
	}

	worker.Token, err = newWorkflowClaimToken()
	if err != nil {
		return nil, err
	}

	owner, err := workflowProcessIdentity(os.Getpid())
	if err != nil {
		return nil, err
	}

	generation := 1
	if set.Set.Closing != nil {
		generation += len(set.Set.Closing.Attempts)
	}

	worker.AttemptID = fmt.Sprintf("%s-attempt-%d", worker.InvocationID, generation)

	data = append(data, '\n')
	state := workflowClosingAttempt{Schema: "chr-set-closing-attempt/v1", SetID: set.ID, Generation: generation, BudgetID: workflowClosingBudgetID(set, request.InvocationID), OwnerRoot: s.layout.Root(), Owner: owner, InvocationID: request.InvocationID, TokenHash: workflowHash([]byte(worker.Token)), Outcome: "running", LeaseUntil: time.Now().UTC().Add(workflowClosingLeaseSeconds(input.Validation)), Input: WorkflowClosingEvidence{Ref: workflowClosingFile(set.ID, generation, "input.json"), Hash: workflowHash(data)}}

	return &workflowClosingClaim{worker: worker, set: set, state: state, input: input, data: data}, nil
}

func workflowClosingOwnerID(set WorkflowRecord) string {
	return "set-closing-" + workflowHash([]byte(set.Namespace+"\x00"+set.ID))
}

func (s *WorkflowStore) workflowClosingReady(ctx context.Context, set WorkflowRecord, request WorkflowSetClosureRequest) error {
	if set.Set.Status == "closing" {
		return ErrWorkflowOwned
	}

	if err := s.checkWorkflowClosingEvidence(set); err != nil {
		return err
	}

	if err := s.workflowClosingBudget(set, request.InvocationID); err != nil {
		return err
	}

	runs, err := s.repositoryWorkflowInvocations(ctx)
	if err != nil {
		return err
	}

	if err := s.workflowClosingInvocationOwner(ctx, runs, set, request.InvocationID); err != nil {
		return err
	}

	if request.InvocationID != "" {
		run, err := s.readWorkflowInvocation(request.InvocationID)
		if err != nil {
			return err
		}

		return s.workflowClosingRunReady(ctx, run, set.ID)
	}

	return nil
}

func (s *WorkflowStore) workflowClosingInvocationOwner(ctx context.Context, runs []workflowInvocationFile, set WorkflowRecord, id string) error {
	for _, run := range runs {
		if run.Namespace != set.Namespace || workflowInvocationTerminal(run.Status) || !workflowClaimResources(run)["set:"+set.ID] {
			continue
		}

		if run.ID != id || run.OwnerRoot != s.layout.Root() {
			return ErrWorkflowOwned
		}

		if err := s.workflowClosingRunReady(ctx, run, set.ID); err != nil {
			return err
		}
	}

	return nil
}
