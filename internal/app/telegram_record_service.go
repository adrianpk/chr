package app

import (
	"context"
	"fmt"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type telegramRecordService interface {
	ReadTelegramRecords(context.Context, core.TelegramOperationTarget, RecordQuery) (RecordPage, error)
	PrepareTelegramRecord(context.Context, core.TelegramOperationTarget, core.WorkflowMutationBinding, RecordMutationRequest) (PreparedRecordMutation, error)
	ApplyTelegramRecord(context.Context, core.TelegramOperationTarget, PreparedRecordMutation) (RecordMutationResult, error)
}

func (s service) telegramProjectService(ctx context.Context, target core.TelegramOperationTarget) (service, error) {
	projects, err := s.ListRegisteredProjects([]string{target.ProjectID})
	if err != nil {
		return s, err
	}

	for _, project := range projects {
		if project.ID == target.ProjectID && project.Enabled {
			s.ctx = ctx
			s.projectContext = commandProjectContext{ProjectDir: project.Path, ProjectID: project.ID, ProjectName: project.Name}

			return s, nil
		}
	}

	return s, fmt.Errorf("%w: registered project is unavailable", workflowstate.ErrConflict)
}

// ReadTelegramRecords resolves the frozen project through the shared registry.
func (s service) ReadTelegramRecords(ctx context.Context, target core.TelegramOperationTarget, query RecordQuery) (RecordPage, error) {
	scoped, err := s.telegramProjectService(ctx, target)
	if err != nil {
		return RecordPage{}, err
	}

	return scoped.ReadRecords(ctx, target.Namespace, query)
}

// PrepareTelegramRecord uses the central owner boundary after project resolution.
func (s service) PrepareTelegramRecord(ctx context.Context, target core.TelegramOperationTarget, binding core.WorkflowMutationBinding, request RecordMutationRequest) (PreparedRecordMutation, error) {
	scoped, err := s.telegramProjectService(ctx, target)
	if err != nil {
		return PreparedRecordMutation{}, err
	}

	return scoped.PrepareRecordMutation(ctx, target.Namespace, binding, request)
}

// ApplyTelegramRecord applies only serialized input matching the frozen target.
func (s service) ApplyTelegramRecord(ctx context.Context, target core.TelegramOperationTarget, prepared PreparedRecordMutation) (RecordMutationResult, error) {
	if prepared.ProjectID != target.ProjectID || prepared.Namespace != target.Namespace {
		return RecordMutationResult{}, workflowstate.ErrConflict
	}

	scoped, err := s.telegramProjectService(ctx, target)
	if err != nil {
		return RecordMutationResult{}, err
	}

	return scoped.ApplyPreparedRecordMutation(ctx, prepared)
}
