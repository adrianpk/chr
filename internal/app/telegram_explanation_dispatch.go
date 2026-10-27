package app

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (r telegramReceiver) respondExplanation(ctx context.Context, user core.TelegramAllowedUser, config core.TelegramReceiverConfig, update core.TelegramReceivedUpdate) (bool, error) {
	route := telegramCommandRoute{
		OperationID: "authoring.explain", Continue: r.continueExplanation, Help: telegramExplanationHelp(),
		Parse: func(update core.TelegramReceivedUpdate) (telegramParsedCommand, error) {
			command, recognized, err := parseTelegramExplanation(update)

			return telegramParsedCommand{
				Recognized: recognized, Input: &command.Input,
				Dispatch: func(ctx context.Context, user core.TelegramAllowedUser) error {
					return r.dispatchExplanation(ctx, user, command)
				},
			}, err
		},
	}

	return r.respondCommand(ctx, user, config, update, route)
}

func (r telegramReceiver) dispatchExplanation(ctx context.Context, user core.TelegramAllowedUser, command telegramExplanationCommand) error {
	operations, ok := r.operations.(telegramExplanationService)
	if !ok {
		return r.recordDiagnostic(ctx, command.Input.Update, command.Input.Config, "Shared explanation is unavailable. No authoring or work state changed.")
	}

	request, err := r.reserveExplanation(ctx, operations, user, command)
	if err != nil {
		if errors.Is(err, errAuthoringDocumentUnavailable) {
			return r.recordDiagnostic(ctx, command.Input.Update, command.Input.Config, "There is no saved "+string(command.DocumentKind)+" document in that conversation. No document was generated or changed.")
		}

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
	switch request.State {
	case "confirmed", "blocked", "rejected":
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

	if prepared.Request.DocumentKind != "" {
		return r.finishOriginalDocument(ctx, operations, request, prepared, current)
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

	response := telegramExplanationDocument(request.Target, request.ID, result, model.Response)
	request.Delivery = &response
	text := response.Message.InitialText

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
		return prepared, workflowstate.ErrConflict
	}

	return prepared, nil
}

func (r telegramReceiver) explanationOwnerRequest(ctx context.Context, command telegramExplanationCommand, target core.TelegramOperationTarget) (AuthoringExplanationRequest, error) {
	owner := AuthoringExplanationRequest{ID: core.TelegramOperationID(command.Input.Update, 0), Question: command.Question, DocumentKind: command.DocumentKind}
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
