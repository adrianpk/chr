package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (r telegramReceiver) respondWork(ctx context.Context, user telegramcore.AllowedUser, config telegramcore.ReceiverConfig, update telegramcore.ReceivedUpdate) (bool, error) {
	route := telegramCommandRoute{
		OperationID: "work.admit", Continue: r.prepareWorkAdmission, Help: telegramWorkHelp(),
		Parse: func(update telegramcore.ReceivedUpdate) (telegramParsedCommand, error) {
			parsedWork, err := parseTelegramWork(update)
			command, recognized := parsedWork.Command, parsedWork.Recognized

			return telegramParsedCommand{
				Recognized: recognized, Input: &command.Input,
				Dispatch: func(ctx context.Context, user telegramcore.AllowedUser) error {
					return r.dispatchWork(ctx, user, command)
				},
			}, err
		},
	}

	return r.respondCommand(ctx, user, config, update, route)
}

func (r telegramReceiver) dispatchWork(ctx context.Context, user telegramcore.AllowedUser, command telegramWorkCommand) error {
	target, err := r.store.ResolveWorkTarget(ctx, command.Input, time.Now())
	if err != nil {
		return r.recordDiagnostic(ctx, command.Input.Update, command.Input.Config, telegramIntentTargetFailure(err))
	}

	arguments, err := json.Marshal(command.Selector)
	if err != nil {
		return err
	}

	request, err := r.store.ReserveOperation(ctx, telegramcore.OperationRequest{Source: command.Input.Update, Operator: user.OperatorID, Target: target, OperationID: "work.admit", Arguments: arguments}, time.Now())
	if err != nil {
		return err
	}

	return r.prepareWorkAdmission(ctx, request, command.Input.Config)
}

func (r telegramReceiver) prepareWorkAdmission(ctx context.Context, request telegramcore.OperationRequest, config telegramcore.ReceiverConfig) error {
	if telegramPreparedWorkFinished(request.State) {
		return nil
	}

	operations, ok := r.operations.(telegramWorkService)
	if !ok {
		return r.finishWorkFailure(ctx, request, telegramcore.OperationBlocked, ErrRecordUnsupported, config)
	}

	if err := r.validateWorkAdmissionAuthority(ctx, request); err != nil {
		return r.finishWorkFailure(ctx, request, telegramcore.OperationBlocked, err, config)
	}

	if request.State == telegramcore.OperationPrepared {
		prepared, err := r.previewWorkAdmission(ctx, operations, request)
		if err != nil {
			return r.finishWorkFailure(ctx, request, telegramcore.OperationRejected, err, config)
		}

		request.PreparedInput = prepared
		request.State = telegramcore.OperationReady
	}

	return r.offerWorkAdmission(ctx, request)
}

func (r telegramReceiver) previewWorkAdmission(ctx context.Context, operations telegramWorkService, request telegramcore.OperationRequest) (json.RawMessage, error) {
	var selector core.WorkflowSelector
	if err := decodeOperationJSON(string(request.Arguments), &selector); err != nil {
		return nil, err
	}

	preview, err := operations.PreviewTelegramWork(ctx, request.Target, selector)
	if err != nil {
		return nil, err
	}

	prepared, err := json.Marshal(preview)
	if err != nil {
		return nil, err
	}

	return prepared, r.store.PrepareOperation(ctx, request, prepared, time.Now())
}

func (r telegramReceiver) offerWorkAdmission(ctx context.Context, request telegramcore.OperationRequest) error {
	preview, err := telegramRetainedWorkPreview(request)
	if err != nil {
		return err
	}

	input, err := r.preparedWorkAuthorityInput(request)
	if err != nil {
		return err
	}

	complete, err := r.workConfirmationContents(ctx, request, preview)
	if errors.Is(err, tgmessage.ErrDocumentUnavailable) {
		return r.unavailableConfirmationDocument(ctx, request, input.Config)
	}

	if errors.Is(err, workflowstate.ErrConflict) {
		return r.finishWorkFailure(ctx, request, telegramcore.OperationBlocked, err, input.Config)
	}

	if err != nil {
		return err
	}

	text := telegramWorkPreviewText(preview)
	text = strings.Replace(text, "\n\n", "\n\n"+strings.Join(complete.Titles[:min(3, len(complete.Titles))], "\n")+"\n", 1)

	return r.offerConfirmationDocument(ctx, request, telegramcore.ActionOffer{Input: input, RequestID: request.ID, Choices: []string{"start"}, ExpiresAt: time.Now().Add(24 * time.Hour)}, telegramConfirmationDocument{Family: "31", Text: text, Complete: complete.Complete})
}

func (r telegramReceiver) validateWorkAdmissionAuthority(ctx context.Context, request telegramcore.OperationRequest) error {
	if err := r.validateRecordAuthority(ctx, request); err != nil {
		return err
	}

	return r.store.ValidateWorkConversation(ctx, request)
}
