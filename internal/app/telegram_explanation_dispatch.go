package app

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func (r telegramReceiver) respondExplanation(ctx context.Context, user core.TelegramAllowedUser, config core.TelegramReceiverConfig, update core.TelegramReceivedUpdate) (bool, error) {
	retained, found, err := r.store.TelegramOperation(ctx, update, 0)
	if err != nil {
		return true, err
	}

	if found && retained.OperationID == "authoring.explain" {
		return true, r.continueExplanation(ctx, retained, config)
	}

	command, recognized, err := parseTelegramExplanation(update)
	if !recognized {
		return false, nil
	}

	if err != nil {
		return true, r.recordDiagnostic(ctx, update, config, telegramExplanationHelp())
	}

	command.Input.Config = config

	command.Input.Projects, err = r.operations.ListRegisteredProjects(user.ProjectIDs)
	if err != nil {
		return true, err
	}

	return true, r.dispatchExplanation(ctx, user, command)
}

func (r telegramReceiver) dispatchExplanation(ctx context.Context, user core.TelegramAllowedUser, command telegramExplanationCommand) error {
	operations, ok := r.operations.(telegramExplanationService)
	if !ok {
		return r.recordDiagnostic(ctx, command.Input.Update, command.Input.Config, "Shared explanation is unavailable. No authoring or work state changed.")
	}

	request, err := r.reserveExplanation(ctx, operations, user, command)
	if err != nil {
		return r.recordDiagnostic(ctx, command.Input.Update, command.Input.Config, telegramIntentTargetFailure(err))
	}

	return r.continueExplanation(ctx, request, command.Input.Config)
}

func (r telegramReceiver) reserveExplanation(ctx context.Context, operations telegramExplanationService, user core.TelegramAllowedUser, command telegramExplanationCommand) (core.TelegramOperationRequest, error) {
	target, err := r.store.ResolveTelegramOperationTarget(ctx, command.Input, time.Now())
	if err != nil {
		return core.TelegramOperationRequest{}, err
	}

	owner, err := r.explanationOwnerRequest(ctx, command, target)
	if err != nil {
		return core.TelegramOperationRequest{}, err
	}

	prepared, err := operations.PrepareTelegramExplanation(ctx, target, owner)
	if err != nil {
		return core.TelegramOperationRequest{}, err
	}

	input, err := json.Marshal(prepared)
	if err != nil {
		return core.TelegramOperationRequest{}, err
	}

	arguments, err := json.Marshal(telegramAuthoringArguments{PayloadHash: recordDigest(input)})
	if err != nil {
		return core.TelegramOperationRequest{}, err
	}

	return r.store.ReservePreparedTelegramOperation(ctx, core.TelegramOperationRequest{Source: command.Input.Update, Operator: user.OperatorID, Target: target, OperationID: "authoring.explain", Arguments: arguments}, input, time.Now())
}

func (r telegramReceiver) continueExplanation(ctx context.Context, request core.TelegramOperationRequest, config core.TelegramReceiverConfig) error {
	if request.State == "confirmed" || request.State == "blocked" || request.State == "rejected" {
		return nil
	}

	current, err := r.validateAuthoringAuthority(ctx, request, false)
	if err != nil {
		return r.finishAuthoringFailure(ctx, request, config, err)
	}

	operations, ok := r.operations.(telegramExplanationService)
	if !ok {
		return r.finishAuthoringFailure(ctx, request, current, ErrRecordUnsupported)
	}

	prepared, err := telegramRetainedExplanation(request)
	if err != nil {
		return r.finishAuthoringFailure(ctx, request, current, err)
	}

	if err := r.store.BeginTelegramOperation(ctx, request, time.Now()); err != nil {
		return err
	}

	model, err := r.store.ResolveTelegramOperationModel(ctx, r.env, request, time.Now(), func(ctx context.Context) (json.RawMessage, error) {
		result, err := operations.ExplainPreparedTelegram(ctx, request.Target, prepared)
		if err != nil {
			return nil, err
		}

		return json.Marshal(result)
	})
	if err != nil {
		return err
	}

	return r.finishExplanationModel(ctx, request, model, current)
}

func (r telegramReceiver) finishExplanationModel(ctx context.Context, request core.TelegramOperationRequest, model core.TelegramOperationModel, config core.TelegramReceiverConfig) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if len(model.Response) == 0 {
		if model.Attempts >= 5 {
			return r.finishAuthoringFailure(ctx, request, config, core.ErrAuthoringProviderExhausted)
		}

		return nil
	}

	current, err := r.validateAuthoringAuthority(ctx, request, false)
	if err != nil {
		return r.finishAuthoringFailure(ctx, request, config, err)
	}

	var result AuthoringExplanation
	if err := decodeOperationJSON(string(model.Response), &result); err != nil {
		return err
	}

	text := result.Text
	if len(result.SourceRefs) > 0 {
		text += "\n\nCanonical sources:\n" + strings.Join(result.SourceRefs, "\n")
	}

	return r.store.FinishTelegramOperation(ctx, request, "confirmed", "telegram-model:"+request.ID, model.Response, text, current.Digest(), time.Now())
}

func telegramRetainedExplanation(request core.TelegramOperationRequest) (PreparedAuthoringExplanation, error) {
	var prepared PreparedAuthoringExplanation
	if err := decodeBoundedOperationJSON(string(request.PreparedInput), &prepared, 1048576); err != nil {
		return prepared, err
	}

	var arguments telegramAuthoringArguments
	if err := decodeOperationJSON(string(request.Arguments), &arguments); err != nil {
		return prepared, err
	}

	if prepared.ProjectID != request.Target.ProjectID || prepared.Request.ID != request.ID || arguments.PayloadHash != recordDigest(request.PreparedInput) {
		return prepared, core.ErrWorkflowConflict
	}

	return prepared, nil
}

func (r telegramReceiver) explanationOwnerRequest(ctx context.Context, command telegramExplanationCommand, target core.TelegramOperationTarget) (AuthoringExplanationRequest, error) {
	owner := AuthoringExplanationRequest{ID: core.TelegramOperationID(command.Input.Update, 0), Question: command.Question}
	if target.ConversationID == "" {
		return owner, nil
	}

	conversation, err := r.store.ReadTelegramConversationBinding(ctx, command.Input, target)
	if err != nil {
		return owner, err
	}

	if conversation.AuthoringSessionRef == "" {
		return owner, nil
	}

	owner.SessionID, err = telegramAuthoringSessionID(conversation.AuthoringSessionRef)

	return owner, err
}
