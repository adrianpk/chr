package app

import (
	"context"
	"errors"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s service) prepareSliceRecordMutation(ctx context.Context, store *workflowcore.Store, binding workflowstate.MutationBinding, request RecordMutationRequest) (workflowcore.MutationRequest, error) {
	parent, err := s.nativeMutationTarget(ctx, store, RecordSliceSet, request.ParentID)
	if err != nil {
		return workflowcore.MutationRequest{}, err
	}

	if err := guardRecordObservation(parent, request); err != nil {
		return workflowcore.MutationRequest{}, err
	}

	if parent.Set.Status != "proposed" {
		return workflowcore.MutationRequest{}, ErrRecordUnsupported
	}

	if unsupportedSliceRecordFields(request) {
		return workflowcore.MutationRequest{}, ErrRecordUnsupported
	}

	if request.Action == "create" {
		if err := prepareSliceDescriptor(&parent, request); err != nil {
			return workflowcore.MutationRequest{}, err
		}
	} else if err := updateSliceDisplay(&parent, request); err != nil {
		return workflowcore.MutationRequest{}, err
	}

	return workflowcore.MutationRequest{Binding: binding, Record: parent, ExpectedRevision: request.ExpectedRevision}, nil
}

func prepareSliceDescriptor(parent *workflowcore.Record, request RecordMutationRequest) error {
	if request.Title != nil {
		return ErrRecordUnsupported
	}

	if request.Slice == nil {
		return errors.New("slice creation requires a complete descriptor")
	}

	descriptor := *request.Slice
	if !validRecordIdentity(descriptor.ID) || !pendingSliceDescriptor(descriptor) {
		return ErrRecordUnsupported
	}

	for _, existing := range parent.Set.Slices {
		if existing.ID == descriptor.ID {
			return workflowstate.ErrConflict
		}
	}

	descriptor.Phase = "pending"
	descriptor.TaskInputHash = workflowcore.TaskInputHash(nil, descriptor.ID)
	parent.Set.Slices = append(parent.Set.Slices, descriptor)

	return nil
}

func updateSliceDisplay(parent *workflowcore.Record, request RecordMutationRequest) error {
	if !validRecordIdentity(request.Identity) || request.Title == nil || request.Slice != nil {
		return ErrRecordUnsupported
	}

	for index, descriptor := range parent.Set.Slices {
		if descriptor.ID != request.Identity {
			continue
		}

		if descriptor.Phase != "pending" {
			return ErrRecordUnsupported
		}

		parent.Set.Slices[index].Title = *request.Title

		return nil
	}

	return ErrRecordUnsupported
}

func unsupportedSliceRecordFields(request RecordMutationRequest) bool {
	return request.Body != nil || request.Context != nil || request.Synopsis != nil || request.Validation != nil || request.PlanRef != "" || request.PlanHash != "" || request.UseCase != ""
}

func pendingSliceDescriptor(descriptor workflowstate.WorkflowSlice) bool {
	return (descriptor.Phase == "" || descriptor.Phase == "pending") && descriptor.Branch == "" && descriptor.CandidateRef == "" && descriptor.PullRequest == "" && descriptor.DeliveryRef == ""
}
