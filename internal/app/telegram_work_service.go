package app

import (
	"context"
	"errors"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramWorkService interface {
	PreviewTelegramWork(context.Context, core.TelegramOperationTarget, core.WorkflowSelector) (WorkflowBoundPreview, error)
	StartTelegramWork(context.Context, core.TelegramOperationTarget, WorkflowBoundPreview, WorkflowBoundAdmission) (AgentWorkResult, error)
}

type telegramWorkContinuationService interface {
	ResumeTelegramWork(context.Context, core.TelegramOperationTarget, WorkflowRetainedReference, WorkflowBoundPreview, core.WorkflowAdmissionCheck) (RetainedWorkProgress, error)
}

// ResumeTelegramWork restores the frozen namespace and enters the same guarded
// retained-work operation used by other hosts, without a replacement selector.
func (s service) ResumeTelegramWork(ctx context.Context, target core.TelegramOperationTarget, reference WorkflowRetainedReference, preview WorkflowBoundPreview, check core.WorkflowAdmissionCheck) (RetainedWorkProgress, error) {
	scoped, err := s.telegramBoundProjectService(ctx, target)
	if err != nil {
		return RetainedWorkProgress{}, err
	}

	scoped.workflowNamespace = target.Namespace

	return scoped.ResumeBoundWork(ctx, reference, preview, check)
}

// PreviewTelegramWork delegates exact selection and policy to the common service.
func (s service) PreviewTelegramWork(ctx context.Context, target core.TelegramOperationTarget, selector core.WorkflowSelector) (WorkflowBoundPreview, error) {
	scoped, err := s.telegramBoundProjectService(ctx, target)
	if err != nil {
		return WorkflowBoundPreview{}, err
	}

	scoped.workflowNamespace = target.Namespace

	selected := ""
	if selector.Extent == "ticket" {
		selected = selector.Identity
	}

	return scoped.PreviewSelectedWork(ctx, core.WorkflowPreviewRequest{Selector: selector}, selected)
}

// StartTelegramWork performs admission only, using the retained common request key.
func (s service) StartTelegramWork(ctx context.Context, target core.TelegramOperationTarget, preview WorkflowBoundPreview, binding WorkflowBoundAdmission) (AgentWorkResult, error) {
	scoped, err := s.telegramBoundProjectService(ctx, target)
	if err != nil {
		return AgentWorkResult{}, err
	}

	scoped.workflowNamespace = target.Namespace

	binding.ReadyTicket = preview.ReadyTicket

	return scoped.StartBoundWorkExtent(ctx, preview.Request, preview.Selected, 0, binding)
}

// ReadTelegramWorkConfirmation reads the exact selected source and contract images
// through the existing scoped workflow owner, without changing readiness.
func (s service) ReadTelegramWorkConfirmation(ctx context.Context, target core.TelegramOperationTarget, preview WorkflowBoundPreview) ([]core.WorkflowRecord, error) {
	scoped, err := s.telegramBoundProjectService(ctx, target)
	if err != nil {
		return nil, err
	}

	scoped.workflowNamespace = target.Namespace

	records, err := scoped.workflowConfirmationRecords(ctx, preview)
	if errors.Is(err, errWorkflowConfirmationBound) {
		return nil, core.ErrTelegramDocumentUnavailable
	}

	return records, err
}
