package core

import (
	"context"
	"errors"
	"fmt"
)

// WorkflowAmendmentRef binds an immutable approved amendment to its owning set.
type WorkflowAmendmentRef struct {
	Revision uint64 `json:"revision"`
	Ref      string `json:"ref"`
	Hash     string `json:"hash"`
}

// WorkflowSetAmendmentRequest retains existing identities and contracts while
// adding approved future work. Reopen explicitly preserves the prior closure.
type WorkflowSetAmendmentRequest struct {
	SetPath            string
	ExpectedRevision   uint64
	ExpectedSourceHash string
	Plan               GeneratedWorkflowPlan
	Reopen             bool
	ReserveTicketKey   func(context.Context) (string, error) `json:"-"`
}

// AmendWorkflowSet requires an observed set image and a changed approved plan.
// It cannot remove, reorder or rewrite retained task contracts or delivered
// slices. No live worker, active lane or unresolved publication may own the set.
func (s *WorkflowStore) AmendWorkflowSet(ctx context.Context, request WorkflowSetAmendmentRequest) (result WorkflowMaterializationResult, err error) {
	if err := ValidateGeneratedWorkflowPlan(request.Plan); err != nil {
		return result, err
	}

	if request.ExpectedRevision == 0 || request.ExpectedSourceHash == "" || request.ReserveTicketKey == nil {
		return result, fmt.Errorf("amendment requires an observed image and native key allocator")
	}

	err = s.withLock(ctx, func() (lockErr error) {
		reservation, readErr := s.reserveWorkflowClaim(ctx, workflowInvocationFile{WorkflowInvocation: WorkflowInvocation{WorkflowIdentity: WorkflowIdentity{ID: "set-amendment"}}})
		if readErr != nil {
			return readErr
		}
		defer func() { lockErr = errors.Join(lockErr, reservation.releaseReservation()) }()

		result, readErr = s.amendWorkflowSet(ctx, request)

		return readErr
	})

	return result, err
}

func (s *WorkflowStore) amendWorkflowSet(ctx context.Context, request WorkflowSetAmendmentRequest) (result WorkflowMaterializationResult, err error) {
	set, runs, err := s.observeWorkflowAmendment(ctx, request)
	if err != nil {
		return WorkflowMaterializationResult{}, err
	}

	body, err := s.readApprovedPlan(request.Plan)
	if err != nil {
		return WorkflowMaterializationResult{}, err
	}

	guard, err := s.acquireWorkflowAllocation()
	if err != nil {
		return WorkflowMaterializationResult{}, err
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

func (s *WorkflowStore) readWorkflowAmendmentResult(result WorkflowMaterializationResult) (WorkflowMaterializationResult, error) {
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
