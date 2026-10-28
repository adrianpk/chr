package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) beginWorkflowClosing(ctx context.Context, request SetClosureRequest, ports FrontierPorts) (claim *workflowClosingClaim, result SetClosureResult, err error) {
	set, err := s.readRecord(request.SetPath)
	if err != nil {
		return nil, result, err
	}

	if !s.workflowClosingSetMatches(set, request) {
		return nil, result, workflowstate.ErrConflict
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

	worker, err := s.reserveWorkflowClosing(ctx, set, request.InvocationID)
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

func (s *Store) reserveWorkflowClosing(ctx context.Context, set Record, id string) (*Claim, error) {
	worker, err := s.reserveWorkflowRole(ctx, workflowClosingReservation(set), "closing")
	if !errors.Is(err, ErrWorkflowCapacity) || id == "" {
		return worker, err
	}

	run, readErr := s.readWorkflowInvocation(id)
	if readErr != nil {
		return nil, errors.Join(err, readErr)
	}

	return nil, s.retainWorkflowCapacityWait(run, err)
}

func (s *Store) workflowClosingSetMatches(set Record, request SetClosureRequest) bool {
	return set.Set != nil && set.ProjectID == s.projectID && set.Namespace == s.layout.Namespace() && set.Revision == request.ExpectedRevision && set.SourceHash == request.ExpectedSourceHash
}

func workflowClosingReservation(set Record) workflowInvocationFile {
	count := 0
	if set.Set.Closing != nil {
		count = len(set.Set.Closing.Attempts)
	}

	return workflowInvocationFile{Invocation: Invocation{Identity: workflowstate.Identity{ID: workflowClosingOwnerID(set)}, InvocationExecution: InvocationExecution{Attempts: make([]Attempt, count)}}}
}

func (s *Store) prepareWorkflowClosing(ctx context.Context, set Record, request SetClosureRequest, ports FrontierPorts, worker *Claim) (*workflowClosingClaim, error) {
	input, err := s.workflowClosingInput(ctx, set, request, ports)
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(input, "", "  ")
	if err != nil || len(data) >= workflowstate.FileLimit {
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
	lease := input.Validation
	lease.TimeoutSeconds = input.ValidationTimeoutSeconds
	state := workflowClosingAttempt{Schema: "chr-set-closing-attempt/v1", SetID: set.ID, Generation: generation, BudgetID: workflowClosingBudgetID(set, request.InvocationID), OwnerRoot: s.layout.Root(), Owner: owner, InvocationID: request.InvocationID, TokenHash: contenthash.Sum([]byte(worker.Token)), Outcome: "running", LeaseUntil: time.Now().UTC().Add(workflowClosingLeaseSeconds(lease)), Input: workflowstate.ClosingEvidence{Ref: workflowClosingFile(set.ID, generation, "input.json"), Hash: contenthash.Sum(data)}}
	state.CapacitySlot = workflowWorkerSlotNumber(worker.reservation.Name())

	return &workflowClosingClaim{worker: worker, set: set, state: state, input: input, data: data}, nil
}

func workflowClosingOwnerID(set Record) string {
	return "set-closing-" + contenthash.Sum([]byte(set.Namespace+"\x00"+set.ID))
}

func (s *Store) workflowClosingReady(ctx context.Context, set Record, request SetClosureRequest) error {
	if set.Set.Status == "closing" {
		return ErrWorkflowOwned
	}

	if err := s.checkWorkflowClosingPublications(ctx); err != nil {
		return err
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

func (s *Store) workflowClosingInvocationOwner(ctx context.Context, runs []workflowInvocationFile, set Record, id string) error {
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
