package app

import (
	"context"
	"errors"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (r telegramReceiver) continuePreparedWork(ctx context.Context, request telegramcore.TelegramOperationRequest, config telegramcore.TelegramReceiverConfig) error {
	if telegramPreparedWorkFinished(request.State) {
		return nil
	}

	current, err := r.validateAuthoringAuthority(ctx, request, request.State == "applying")
	if err != nil {
		return r.finishAuthoringFailure(ctx, request, config, err)
	}

	operations, ok := r.operations.(telegramPreparedWorkService)
	if !ok {
		return r.finishAuthoringFailure(ctx, request, current, ErrRecordUnsupported)
	}

	prepared, err := r.prepareTelegramWork(ctx, operations, request)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		return r.finishAuthoringFailure(ctx, request, current, err)
	}

	retained, found, err := r.store.TelegramOperation(ctx, request.Source, request.Ordinal)
	if err != nil {
		return err
	}

	if !found {
		return workflowstate.ErrConflict
	}

	request = retained
	if prepared.Scope.Intent == authoringstate.IntentDevelop && prepared.Authorization == nil {
		return r.offerPreparedWorkScope(ctx, request, prepared)
	}

	return r.applyPreparedWork(ctx, operations, request, prepared, current)
}

func (r telegramReceiver) applyPreparedWork(ctx context.Context, operations telegramPreparedWorkService, request telegramcore.TelegramOperationRequest, prepared PreparedAuthoringWork, config telegramcore.TelegramReceiverConfig) error {
	if err := r.validateInitialPreparedWorkGrant(ctx, request, prepared); err != nil {
		return r.finishAuthoringFailure(ctx, request, config, err)
	}

	if err := r.store.BeginTelegramOperation(ctx, request, time.Now()); err != nil {
		return err
	}

	result, err := operations.ApplyTelegramAuthoringWork(ctx, request.Target, prepared, func(ctx context.Context) error {
		_, err := r.validateAuthoringAuthority(ctx, request, false)

		return err
	})
	if errors.Is(err, workflowstate.ErrConflict) || errors.Is(err, ErrRecordUnsupported) || errors.Is(err, core.ErrAuthoringAuthorityRequired) {
		return r.finishAuthoringFailure(ctx, request, config, err)
	}

	if err != nil {
		return err
	}

	return r.finishPreparedWork(ctx, request, result, config)
}

func (r telegramReceiver) validateInitialPreparedWorkGrant(ctx context.Context, request telegramcore.TelegramOperationRequest, prepared PreparedAuthoringWork) error {
	if err := core.ValidateAuthoringWorkAuthorization(prepared.Scope, prepared.Authorization); err != nil {
		return err
	}

	if prepared.Authorization == nil {
		return nil
	}

	initial, err := r.store.IsInitialTelegramAuthoringWork(ctx, request)
	if err != nil {
		return err
	}

	if !initial {
		return workflowstate.ErrConflict
	}

	input, err := r.preparedWorkAuthorityInput(request)
	if err != nil {
		return err
	}

	instruction, err := r.store.ReadTelegramConversationInstruction(ctx, input)
	if err != nil {
		return err
	}

	if instruction.DevelopmentRef == "" || prepared.Authorization.Binding != (workflowstate.MutationBinding{RequestID: instruction.DevelopmentRef, Operator: request.Operator}) {
		return workflowstate.ErrConflict
	}

	return nil
}

func (r telegramReceiver) offerPreparedWorkScope(ctx context.Context, request telegramcore.TelegramOperationRequest, prepared PreparedAuthoringWork) error {
	input, err := r.preparedWorkAuthorityInput(request)
	if err != nil {
		return err
	}

	complete := ""
	if prepared.Mutation.Native != nil {
		complete = "## Complete proposed ticket\n\n" + string(prepared.Mutation.Native.Record.Body)
	}

	return r.offerConfirmationDocument(ctx, request, telegramcore.TelegramActionOffer{Input: input, RequestID: request.ID, Choices: []string{"develop"}, ExpiresAt: time.Now().Add(time.Hour), ConversationPhase: "awaiting_confirmation"}, "25", telegramPreparedWorkPreview(prepared), complete)
}

func telegramPreparedWorkFinished(state string) bool {
	switch state {
	case "confirmed", "blocked", "rejected", "awaiting_confirmation":
		return true
	default:
		return false
	}
}
