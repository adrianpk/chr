package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowcloseout "forge.adrianpk.com/chr/internal/core/workflow/closeout"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowClosingAdmission struct {
	Claim  *workflowClosingClaim
	Result workflowcloseout.Result
}

func (s *Store) beginWorkflowClosing(ctx context.Context, request workflowcloseout.Request, ports FrontierPorts) (admission workflowClosingAdmission, err error) {
	set, err := s.readRecord(request.SetPath)
	if err != nil {
		return admission, err
	}

	if !s.workflowClosingSetMatches(set, request) {
		return admission, workflowstate.ErrConflict
	}

	if err := s.checkWorkflowSetAmendmentEvidence(set); err != nil {
		return admission, err
	}

	if set.Set.Status == "completed" {
		admission.Result, err = workflowcloseout.ReadClosedSet(set, workflowClosingEvidence{store: s})

		return admission, err
	}

	if err := s.workflowClosingReady(ctx, set, request); err != nil {
		return admission, err
	}

	worker, err := s.reserveWorkflowClosing(ctx, set, request.InvocationID)
	if err != nil {
		return admission, err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, worker.releaseReservation())
		}
	}()

	claim, err := s.prepareWorkflowClosing(ctx, set, request, ports, worker)
	if err != nil {
		return admission, err
	}

	admission.Claim = claim
	err = s.commitWorkflowClosingStart(claim)

	return admission, err
}

func (s *Store) reserveWorkflowClosing(ctx context.Context, set workflowstate.Record, id string) (*Claim, error) {
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

func (s *Store) workflowClosingSetMatches(set workflowstate.Record, request workflowcloseout.Request) bool {
	return set.Set != nil && set.ProjectID == s.projectID && set.Namespace == s.layout.Namespace() && set.Revision == request.ExpectedRevision && set.SourceHash == request.ExpectedSourceHash
}

func workflowClosingReservation(set workflowstate.Record) workflowInvocationFile {
	count := 0
	if set.Set.Closing != nil {
		count = len(set.Set.Closing.Attempts)
	}

	return workflowInvocationFile{Invocation: workflowstate.Invocation{Identity: workflowstate.Identity{ID: workflowClosingOwnerID(set)}, InvocationExecution: workflowstate.InvocationExecution{Attempts: make([]workflowstate.Attempt, count)}}}
}

func (s *Store) prepareWorkflowClosing(ctx context.Context, set workflowstate.Record, request workflowcloseout.Request, ports FrontierPorts, worker *Claim) (*workflowClosingClaim, error) {
	input, err := workflowcloseout.BuildInput(ctx, set, request, workflowClosingInputSource{store: s, ports: ports})
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
	state := workflowcloseout.Attempt{Schema: "chr-set-closing-attempt/v1", SetID: set.ID, Generation: generation, BudgetID: workflowcloseout.BudgetID(*set.Set, request.InvocationID), OwnerRoot: s.layout.Root(), Owner: owner, InvocationID: request.InvocationID, TokenHash: contenthash.Sum([]byte(worker.Token)), Outcome: "running", LeaseUntil: time.Now().UTC().Add(workflowcloseout.LeaseDuration(lease)), Input: workflowstate.ClosingEvidence{Ref: workflowcloseout.File(set.ID, generation, "input.json"), Hash: contenthash.Sum(data)}}
	state.CapacitySlot = workflowWorkerSlotNumber(worker.reservation.Name())

	return &workflowClosingClaim{worker: worker, set: set, state: state, input: input, data: data}, nil
}

func workflowClosingOwnerID(set workflowstate.Record) string {
	return "set-closing-" + contenthash.Sum([]byte(set.Namespace+"\x00"+set.ID))
}

func (s *Store) workflowClosingReady(ctx context.Context, set workflowstate.Record, request workflowcloseout.Request) error {
	if set.Set.Status == "closing" {
		return workflowstate.ErrOwned
	}

	if err := s.checkWorkflowClosingPublications(ctx); err != nil {
		return err
	}

	if err := workflowcloseout.CheckHistory(set.ID, set.Set.Closing, workflowClosingEvidence{store: s}); err != nil {
		return err
	}

	if err := workflowcloseout.CheckBudget(set.ID, *set.Set, request.InvocationID, workflowClosingEvidence{store: s}); err != nil {
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

		return workflowcloseout.CheckReady(ctx, workflowcloseout.Readiness{Run: run.Invocation, SetID: set.ID, OwnerRoot: run.OwnerRoot, Root: s.layout.Root()}, workflowClosingRunGuard{store: s, run: run})
	}

	return nil
}

func (s *Store) workflowClosingInvocationOwner(ctx context.Context, runs []workflowInvocationFile, set workflowstate.Record, id string) error {
	for _, run := range runs {
		if run.Namespace != set.Namespace || workflowstate.InvocationTerminal(run.Status) || !workflowClaimResources(run)["set:"+set.ID] {
			continue
		}

		if run.ID != id || run.OwnerRoot != s.layout.Root() {
			return workflowstate.ErrOwned
		}

		if err := workflowcloseout.CheckReady(ctx, workflowcloseout.Readiness{Run: run.Invocation, SetID: set.ID, OwnerRoot: run.OwnerRoot, Root: s.layout.Root()}, workflowClosingRunGuard{store: s, run: run}); err != nil {
			return err
		}
	}

	return nil
}
