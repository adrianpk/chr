package app

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type telegramGuidedProjectService interface {
	telegramGuidedProject(context.Context, telegramcore.TelegramOperationTarget) (service, error)
}

// The receiver inventories associations; the shared host owns all progression,
// native dispatch, finite limits, recovery and process shutdown.
type telegramGuidedHost struct {
	receiver telegramReceiver
	shared   *guidedWorkHost
	after    string
}

func (h *telegramGuidedHost) run(ctx context.Context) error {
	defer h.shared.close()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for ctx.Err() == nil {
		if err := h.pump(ctx); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
		case <-ticker.C:
		}
	}

	return nil
}

func (h *telegramGuidedHost) pump(ctx context.Context) error {
	refs, err := h.receiver.store.TelegramGuidedReferences(ctx, h.receiver.botID, h.after, 16)
	if err != nil {
		return err
	}

	h.after = ""
	for _, ref := range refs {
		h.after = ref.Conversation.ID
		if err := h.recoverReference(ctx, ref); err != nil && ctx.Err() == nil {
			if !errors.Is(err, workflowstate.ErrConflict) {
				slog.WarnContext(ctx, "Guided Telegram association needs recovery", "code", "guided_association_blocked")
			}
		}
	}

	return nil
}

func (h *telegramGuidedHost) recoverReference(ctx context.Context, ref telegramcore.TelegramGuidedReference) error {
	operations, ok := h.receiver.operations.(telegramGuidedProjectService)
	if !ok {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := h.receiver.validateGuidedHostReference(ctx, ref); err != nil {
		return err
	}

	scoped, err := operations.telegramGuidedProject(ctx, ref.Request.Target)
	if err != nil {
		return err
	}

	scoped.guidedWorkHost = h.shared
	scoped.guidedHostCheck = func(ctx context.Context) error { return h.receiver.validateGuidedHostReference(ctx, ref) }
	scope := core.GuidedUseCaseScope{RunID: ref.Conversation.GuidedRunID, ProjectID: ref.Request.Target.ProjectID, Namespace: ref.Request.Target.Namespace}

	session, err := core.RecoverGuidedHostRun(ctx, scoped.projectContext.ProjectDir, scope)
	if err != nil {
		return err
	}

	h.shared.recoverRun(ctx, scoped, session)

	observed, err := scoped.InspectGuidedUseCase(ctx, scope.RunID, scope.Namespace)
	if err != nil {
		return err
	}

	progress, err := telegramGuidedProgress(observed)
	if err != nil {
		return err
	}

	config, err := h.receiver.validateAuthoringAuthority(ctx, ref.Request, true)
	if err != nil {
		return err
	}

	delivery, err := h.receiver.guidedDocument(ctx, ref.Request, observed, "")
	if err != nil {
		return err
	}

	return h.receiver.store.ProjectTelegramGuided(ctx, ref, progress, *telegramOutcomeMessage("43", telegramGuidedPrompt(observed)), delivery, config.Digest(), time.Now())
}

func (r telegramReceiver) validateGuidedHostReference(ctx context.Context, ref telegramcore.TelegramGuidedReference) error {
	config, err := r.validateAuthoringAuthority(ctx, ref.Request, true)
	if err != nil {
		return err
	}

	return r.validateGuidedRunBinding(ctx, ref.Request, config, ref.Conversation.GuidedRunID)
}

func (r telegramReceiver) boundTelegramGuidedService(operations telegramGuidedService, request telegramcore.TelegramOperationRequest, id string) telegramGuidedService {
	if scoped, ok := operations.(service); ok {
		scoped.guidedHostCheck = func(ctx context.Context) error {
			config, err := r.validateAuthoringAuthority(ctx, request, true)
			if err != nil {
				return err
			}

			return r.validateGuidedRunBinding(ctx, request, config, id)
		}

		return scoped
	}

	return operations
}
