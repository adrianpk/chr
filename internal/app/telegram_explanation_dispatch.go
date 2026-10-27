package app

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (r telegramReceiver) respondExplanation(ctx context.Context, user telegramcore.AllowedUser, config telegramcore.ReceiverConfig, update telegramcore.ReceivedUpdate) (bool, error) {
	route := telegramCommandRoute{
		OperationID: "authoring.explain", Continue: r.continueExplanation, Help: telegramExplanationHelp(),
		Parse: func(update telegramcore.ReceivedUpdate) (telegramParsedCommand, error) {
			parsedExplanation, err := parseTelegramExplanation(update)
			command, recognized := parsedExplanation.Command, parsedExplanation.Recognized

			return telegramParsedCommand{
				Recognized: recognized, Input: &command.Input,
				Dispatch: func(ctx context.Context, user telegramcore.AllowedUser) error {
					return r.dispatchExplanation(ctx, user, command)
				},
			}, err
		},
	}

	return r.respondCommand(ctx, user, config, update, route)
}

func (r telegramReceiver) dispatchExplanation(ctx context.Context, user telegramcore.AllowedUser, command telegramExplanationCommand) error {
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

func (r telegramReceiver) reserveExplanation(ctx context.Context, operations telegramExplanationService, user telegramcore.AllowedUser, command telegramExplanationCommand) (telegramcore.OperationRequest, error) {
	target, err := r.store.ResolveOperationTarget(ctx, command.Input, time.Now())
	if err != nil {
		return telegramcore.OperationRequest{}, err
	}

	owner, err := r.explanationOwnerRequest(ctx, command, target)
	if err != nil {
		return telegramcore.OperationRequest{}, err
	}

	prepared, err := operations.PrepareTelegramExplanation(ctx, target, owner)
	if err != nil {
		return telegramcore.OperationRequest{}, err
	}

	input, err := json.Marshal(prepared)
	if err != nil {
		return telegramcore.OperationRequest{}, err
	}

	arguments, err := json.Marshal(telegramAuthoringArguments{PayloadHash: recordDigest(input)})
	if err != nil {
		return telegramcore.OperationRequest{}, err
	}

	return r.store.ReservePreparedOperation(ctx, telegramcore.OperationRequest{Source: command.Input.Update, Operator: user.OperatorID, Target: target, OperationID: "authoring.explain", Arguments: arguments}, input, time.Now())
}

func (r telegramReceiver) continueExplanation(ctx context.Context, request telegramcore.OperationRequest, config telegramcore.ReceiverConfig) error {
	switch request.State {
	case telegramcore.OperationConfirmed, telegramcore.OperationBlocked, telegramcore.OperationRejected:
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

	if err := r.store.BeginOperation(ctx, request, time.Now()); err != nil {
		return err
	}

	if prepared.Request.DocumentKind != "" {
		return r.finishOriginalDocument(ctx, operations, request, prepared, current)
	}

	model, err := r.store.ResolveOperationModel(ctx, r.env, request, time.Now(), func(ctx context.Context) (json.RawMessage, error) {
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

func (r telegramReceiver) finishExplanationModel(ctx context.Context, request telegramcore.OperationRequest, model telegramcore.OperationModel, config telegramcore.ReceiverConfig) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if len(model.Response) == 0 {
		if model.Attempts >= 5 {
			return r.finishAuthoringFailure(ctx, request, config, authoringcore.ErrAuthoringProviderExhausted)
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

	return r.store.FinishOperation(ctx, request, telegramcore.OperationCompletion{State: telegramcore.OperationConfirmed, ReceiptRef: "telegram-model:" + request.ID, Facts: model.Response, Text: text, ConfigDigest: current.Digest(), At: time.Now()})
}

func telegramRetainedExplanation(request telegramcore.OperationRequest) (PreparedAuthoringExplanation, error) {
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

func (r telegramReceiver) explanationOwnerRequest(ctx context.Context, command telegramExplanationCommand, target telegramcore.OperationTarget) (AuthoringExplanationRequest, error) {
	owner := AuthoringExplanationRequest{ID: telegramcore.OperationID(command.Input.Update, 0), Question: command.Question, DocumentKind: command.DocumentKind}
	if target.ConversationID == "" {
		return owner, nil
	}

	conversation, err := r.store.ReadConversationBinding(ctx, command.Input, target)
	if err != nil {
		return owner, err
	}

	if conversation.AuthoringSessionRef == "" {
		return owner, nil
	}

	owner.SessionID, err = telegramAuthoringSessionID(conversation.AuthoringSessionRef)

	return owner, err
}
