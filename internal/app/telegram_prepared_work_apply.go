package app

import (
	"context"
	"errors"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func (r telegramReceiver) continuePreparedWork(ctx context.Context, request core.TelegramOperationRequest, config core.TelegramReceiverConfig) error {
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
		return core.ErrWorkflowConflict
	}

	request = retained
	if prepared.Scope.Intent == core.AuthoringIntentDevelop && prepared.Authorization == nil {
		return r.offerPreparedWorkScope(ctx, request, prepared)
	}

	return r.applyPreparedWork(ctx, operations, request, prepared, current)
}

func (r telegramReceiver) applyPreparedWork(ctx context.Context, operations telegramPreparedWorkService, request core.TelegramOperationRequest, prepared PreparedAuthoringWork, config core.TelegramReceiverConfig) error {
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
	if errors.Is(err, core.ErrWorkflowConflict) || errors.Is(err, ErrRecordUnsupported) || errors.Is(err, core.ErrAuthoringAuthorityRequired) {
		return r.finishAuthoringFailure(ctx, request, config, err)
	}

	if err != nil {
		return err
	}

	return r.finishPreparedWork(ctx, request, result, config)
}

func (r telegramReceiver) validateInitialPreparedWorkGrant(ctx context.Context, request core.TelegramOperationRequest, prepared PreparedAuthoringWork) error {
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
		return core.ErrWorkflowConflict
	}

	input, err := r.preparedWorkAuthorityInput(request)
	if err != nil {
		return err
	}

	instruction, err := r.store.ReadTelegramConversationInstruction(ctx, input)
	if err != nil {
		return err
	}

	if instruction.DevelopmentRef == "" || prepared.Authorization.Binding != (core.WorkflowMutationBinding{RequestID: instruction.DevelopmentRef, Operator: request.Operator}) {
		return core.ErrWorkflowConflict
	}

	return nil
}

func (r telegramReceiver) offerPreparedWorkScope(ctx context.Context, request core.TelegramOperationRequest, prepared PreparedAuthoringWork) error {
	input, err := r.preparedWorkAuthorityInput(request)
	if err != nil {
		return err
	}

	_, err = r.store.OfferTelegramAction(ctx, core.TelegramActionOffer{Input: input, RequestID: request.ID, Choices: []string{"develop"}, Text: telegramPreparedWorkPreview(prepared) + "\nAuthorize development of this exact scope?", ExpiresAt: time.Now().Add(time.Hour), ConversationPhase: "awaiting_confirmation"}, time.Now())

	return err
}

func telegramPreparedWorkFinished(state string) bool {
	switch state {
	case "confirmed", "blocked", "rejected", "awaiting_confirmation":
		return true
	default:
		return false
	}
}
