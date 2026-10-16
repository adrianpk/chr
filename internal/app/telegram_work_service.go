package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramWorkService interface {
	PreviewTelegramWork(context.Context, core.TelegramOperationTarget, core.WorkflowSelector) (WorkflowBoundPreview, error)
	StartTelegramWork(context.Context, core.TelegramOperationTarget, WorkflowBoundPreview, WorkflowBoundAdmission) (AgentWorkResult, error)
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

	return scoped.StartBoundWorkExtent(ctx, preview.Request, preview.Selected, 0, binding)
}
