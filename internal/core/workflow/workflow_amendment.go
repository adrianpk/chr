package workflow

import (
	"context"
	"errors"
	"fmt"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// SetAmendmentRequest retains existing identities and contracts while
// adding approved future work. Reopen explicitly preserves the prior closure.
type SetAmendmentRequest struct {
	SetPath            string
	ExpectedRevision   uint64
	ExpectedSourceHash string
	Plan               generationcore.GeneratedWorkflowPlan
	Reopen             bool
	ReserveTicketKey   func(context.Context) (string, error) `json:"-"`
}

// AmendWorkflowSet requires an observed set image and a changed approved plan.
// It cannot remove, reorder or rewrite retained task contracts or delivered
// slices. No live worker, active lane or unresolved publication may own the set.
func (s *Store) AmendWorkflowSet(ctx context.Context, request SetAmendmentRequest) (result MaterializationResult, err error) {
	if err := generationcore.ValidateGeneratedWorkflowPlan(request.Plan); err != nil {
		return result, err
	}

	if request.ExpectedRevision == 0 || request.ExpectedSourceHash == "" || request.ReserveTicketKey == nil {
		return result, fmt.Errorf("amendment requires an observed image and native key allocator")
	}

	err = s.withWorkflowPlanMutation(ctx, func(ctx context.Context) (lockErr error) {
		reservation, readErr := s.reserveWorkflowRole(ctx, workflowInvocationFile{Invocation: Invocation{Identity: Identity{ID: "set-amendment"}}}, "set_amendment")
		if readErr != nil {
			return readErr
		}
		defer func() { lockErr = errors.Join(lockErr, reservation.releaseReservation()) }()

		result, readErr = s.amendWorkflowSet(ctx, request)

		return readErr
	})

	return result, err
}

func (s *Store) amendWorkflowSet(ctx context.Context, request SetAmendmentRequest) (result MaterializationResult, err error) {
	set, runs, err := s.observeWorkflowAmendment(ctx, request)
	if err != nil {
		return MaterializationResult{}, err
	}

	body, err := s.readApprovedPlan(request.Plan)
	if err != nil {
		return MaterializationResult{}, err
	}

	if set.Set.PlanRevision != nil && request.Plan.PlanRevision == nil {
		return result, workflowstate.ErrConflict
	}

	if err := s.validateApprovedOrder(request.Plan, set.ID); err != nil {
		return result, err
	}

	guard, err := s.acquireWorkflowAllocation()
	if err != nil {
		return MaterializationResult{}, err
	}
	defer func() { err = errors.Join(err, s.releaseWorkflowAllocation(guard)) }()

	result, files, err := s.buildWorkflowAmendment(ctx, set, request, body)
	if err != nil {
		return result, err
	}

	files, err = s.workflowAmendmentInvocationFiles(ctx, runs, set, result.Set, files)
	if err != nil {
		return result, err
	}

	if err := s.commitAllocatedFiles(files, guard); err != nil {
		return result, err
	}

	return s.readWorkflowAmendmentResult(result)
}

func (s *Store) readWorkflowAmendmentResult(result MaterializationResult) (MaterializationResult, error) {
	set, err := s.readRecord(result.Set.Path)
	if err != nil {
		return result, err
	}

	result.Set = set
	for i, task := range result.Tasks {
		result.Tasks[i], err = s.readRecord(task.Path)
		if err != nil {
			return result, err
		}
	}

	return result, nil
}
