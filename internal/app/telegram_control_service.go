package app

import (
	"context"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type telegramControlService interface {
	ReadTelegramControl(context.Context, telegramcore.OperationTarget, string) (WorkflowWorkStatus, error)
	ReadTelegramControlReceipt(context.Context, telegramcore.OperationTarget, workflowstate.MutationBinding, WorkflowControlRequest) (telegramControlReceipt, error)
	ApplyTelegramControl(context.Context, telegramcore.OperationTarget, workflowstate.MutationBinding, WorkflowControlRequest, workflowcore.AdmissionCheck) (WorkflowWorkStatus, error)
}

// ReadTelegramControlReceipt scopes receipt recovery to the original project.
func (s service) ReadTelegramControlReceipt(ctx context.Context, target telegramcore.OperationTarget, binding workflowstate.MutationBinding, request WorkflowControlRequest) (telegramControlReceipt, error) {
	scoped, err := s.telegramBoundProjectService(ctx, target)
	if err != nil {
		return telegramControlReceipt{Status: WorkflowWorkStatus{}, Found: false}, err
	}

	scoped.workflowNamespace = target.Namespace
	request.RequestBinding = &binding

	status, found, err := scoped.ReadWorkControlReceipt(ctx, request)

	return telegramControlReceipt{Status: status, Found: found}, err
}

// ReadTelegramControl observes the frozen project and namespace through native status.
func (s service) ReadTelegramControl(ctx context.Context, target telegramcore.OperationTarget, id string) (WorkflowWorkStatus, error) {
	scoped, err := s.telegramBoundProjectService(ctx, target)
	if err != nil {
		return WorkflowWorkStatus{}, err
	}

	scoped.workflowNamespace = target.Namespace

	return scoped.WorkflowStatus(ctx, id, WorkflowStatusRequest{Limit: 8})
}

// ApplyTelegramControl enters the same guarded operation as other adapters.
// The configured actor is carried by the immutable caller request binding.
func (s service) ApplyTelegramControl(ctx context.Context, target telegramcore.OperationTarget, binding workflowstate.MutationBinding, request WorkflowControlRequest, check workflowcore.AdmissionCheck) (WorkflowWorkStatus, error) {
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
