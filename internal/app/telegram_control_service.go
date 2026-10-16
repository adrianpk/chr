package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramControlService interface {
	ReadTelegramControl(context.Context, core.TelegramOperationTarget, string) (WorkflowWorkStatus, error)
	ReadTelegramControlReceipt(context.Context, core.TelegramOperationTarget, core.WorkflowMutationBinding, WorkflowControlRequest) (WorkflowWorkStatus, bool, error)
	ApplyTelegramControl(context.Context, core.TelegramOperationTarget, core.WorkflowMutationBinding, WorkflowControlRequest, core.WorkflowAdmissionCheck) (WorkflowWorkStatus, error)
}

// ReadTelegramControlReceipt scopes receipt recovery to the original project.
func (s service) ReadTelegramControlReceipt(ctx context.Context, target core.TelegramOperationTarget, binding core.WorkflowMutationBinding, request WorkflowControlRequest) (WorkflowWorkStatus, bool, error) {
	scoped, err := s.telegramBoundProjectService(ctx, target)
	if err != nil {
		return WorkflowWorkStatus{}, false, err
	}

	scoped.workflowNamespace = target.Namespace
	request.RequestBinding = &binding

	return scoped.ReadWorkControlReceipt(ctx, request)
}

// ReadTelegramControl observes the frozen project and namespace through native status.
func (s service) ReadTelegramControl(ctx context.Context, target core.TelegramOperationTarget, id string) (WorkflowWorkStatus, error) {
	scoped, err := s.telegramBoundProjectService(ctx, target)
	if err != nil {
		return WorkflowWorkStatus{}, err
	}

	scoped.workflowNamespace = target.Namespace

	return scoped.WorkflowStatus(ctx, id, WorkflowStatusRequest{Limit: 8})
}

// ApplyTelegramControl enters the same guarded operation as other adapters.
// The configured actor is carried by the immutable caller request binding.
func (s service) ApplyTelegramControl(ctx context.Context, target core.TelegramOperationTarget, binding core.WorkflowMutationBinding, request WorkflowControlRequest, check core.WorkflowAdmissionCheck) (WorkflowWorkStatus, error) {
	if err := binding.Validate(); err != nil {
		return WorkflowWorkStatus{}, err
	}

	scoped, err := s.telegramBoundProjectService(ctx, target)
	if err != nil {
		return WorkflowWorkStatus{}, err
	}

	scoped.workflowNamespace = target.Namespace

	request.AuthorizationCheck = check
	request.RequestBinding = &binding

	return scoped.ControlWorkExtent(ctx, request)
}
