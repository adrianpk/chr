package app

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramWorkCompletion struct {
	Reference core.TelegramWorkReference
	Progress  RetainedWorkProgress
	Failure   error
}

func (r telegramReceiver) activateWork(ctx context.Context, ref core.TelegramWorkReference) telegramWorkCompletion {
	result := telegramWorkCompletion{Reference: ref}
	result.Progress, result.Failure = r.resumeWorkReference(ctx, ref)

	return result
}

func (r telegramReceiver) resumeWorkReference(ctx context.Context, ref core.TelegramWorkReference) (RetainedWorkProgress, error) {
	operations, ok := r.operations.(telegramWorkContinuationService)
	if !ok {
		return RetainedWorkProgress{}, ErrRecordUnsupported
	}

	if err := r.validateWorkReference(ctx, ref); err != nil {
		return RetainedWorkProgress{}, err
	}

	preview, err := telegramRetainedWorkPreview(ref.Request)
	if err != nil {
		return RetainedWorkProgress{}, err
	}

	var admission telegramWorkAdmissionResult
	if err := decodeOperationJSON(string(ref.Request.Result), &admission); err != nil {
		return RetainedWorkProgress{}, err
	}

	if admission.Invocation.ID != ref.Invocation.ID || admission.RequestKey == "" || admission.Selector != preview.Request.Selector {
		return RetainedWorkProgress{}, core.ErrWorkflowConflict
	}

	check := func(ctx context.Context) error { return r.validateWorkReference(ctx, ref) }
	bound := WorkflowRetainedReference{ID: ref.Invocation.ID, RequestKey: admission.RequestKey}

	return operations.ResumeTelegramWork(ctx, ref.Request.Target, bound, preview, check)
}

func (r telegramReceiver) validateWorkReference(ctx context.Context, ref core.TelegramWorkReference) error {
	request := ref.Request
	request.Target.ConversationRevision = ref.ConversationRevision

	if err := r.validateRecordAuthority(ctx, request); err != nil {
		return err
	}

	return r.store.ValidateTelegramWorkReference(ctx, ref)
}

func (r telegramReceiver) completeWorkActivation(ctx context.Context, result telegramWorkCompletion) error {
	if result.Failure != nil && !errors.Is(result.Failure, context.Canceled) {
		level := slog.LevelWarn

		code := telegramWorkActivationCode(result.Failure)
		if code != "native_work_unavailable" {
			level = slog.LevelDebug
		}

		slog.Log(ctx, level, "Telegram work activation stopped", "invocation", result.Reference.Invocation.ID, "code", code)
	}

	status := result.Progress.Status
	if status.ID == "" {
		return nil
	}

	ref := result.Reference
	if status.ID != ref.Invocation.ID || status.ProjectID != ref.Request.Target.ProjectID || status.Namespace != ref.Request.Target.Namespace {
		return core.ErrWorkflowConflict
	}

	return r.projectWorkProgress(ctx, ref, status)
}

func (r telegramReceiver) projectWorkProgress(ctx context.Context, ref core.TelegramWorkReference, status WorkflowWorkStatus) error {
	if err := r.validateWorkReference(ctx, ref); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if errors.Is(err, core.ErrWorkflowConflict) {
			return nil
		}

		return err
	}

	config, err := r.currentConfig()
	if err != nil {
		return err
	}

	return r.store.ProjectTelegramWork(ctx, ref, telegramWorkObservation(status), config.Digest(), time.Now())
}

func telegramWorkActivationCode(err error) string {
	switch {
	case errors.Is(err, core.ErrWorkflowOwned):
		return "current_owner"
	case errors.Is(err, core.ErrWorkflowCapacity):
		return "waiting_capacity"
	case errors.Is(err, core.ErrWorkflowConflict):
		return "binding_changed"
	default:
		return "native_work_unavailable"
	}
}
