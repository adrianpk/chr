package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (r telegramReceiver) respondWork(ctx context.Context, user core.TelegramAllowedUser, config core.TelegramReceiverConfig, update core.TelegramReceivedUpdate) (bool, error) {
	retained, found, err := r.store.TelegramOperation(ctx, update, 0)
	if err != nil {
		return true, err
	}

	if found && retained.OperationID == "work.admit" {
		return true, r.prepareWorkAdmission(ctx, retained, config)
	}

	command, recognized, err := parseTelegramWork(update)
	if !recognized {
		return false, nil
	}

	if err != nil {
		return true, r.recordDiagnostic(ctx, update, config, telegramWorkHelp())
	}

	command.Input.Config = config

	command.Input.Projects, err = r.operations.ListRegisteredProjects(user.ProjectIDs)
	if err != nil {
		return true, err
	}

	return true, r.dispatchWork(ctx, user, command)
}

func (r telegramReceiver) dispatchWork(ctx context.Context, user core.TelegramAllowedUser, command telegramWorkCommand) error {
	target, err := r.store.ResolveTelegramWorkTarget(ctx, command.Input, time.Now())
	if err != nil {
		return r.recordDiagnostic(ctx, command.Input.Update, command.Input.Config, telegramIntentTargetFailure(err))
	}

	arguments, err := json.Marshal(command.Selector)
	if err != nil {
		return err
	}

	request, err := r.store.ReserveTelegramOperation(ctx, core.TelegramOperationRequest{Source: command.Input.Update, Operator: user.OperatorID, Target: target, OperationID: "work.admit", Arguments: arguments}, time.Now())
	if err != nil {
		return err
	}

	return r.prepareWorkAdmission(ctx, request, command.Input.Config)
}

func (r telegramReceiver) prepareWorkAdmission(ctx context.Context, request core.TelegramOperationRequest, config core.TelegramReceiverConfig) error {
	if telegramPreparedWorkFinished(request.State) {
		return nil
	}

	operations, ok := r.operations.(telegramWorkService)
	if !ok {
		return r.finishWorkFailure(ctx, request, "blocked", ErrRecordUnsupported, config)
	}

	if err := r.validateWorkAdmissionAuthority(ctx, request); err != nil {
		return r.finishWorkFailure(ctx, request, "blocked", err, config)
	}

	if request.State == "prepared" {
		prepared, err := r.previewWorkAdmission(ctx, operations, request)
		if err != nil {
			return r.finishWorkFailure(ctx, request, "rejected", err, config)
		}

		request.PreparedInput = prepared
		request.State = "ready"
	}

	return r.offerWorkAdmission(ctx, request)
}

func (r telegramReceiver) previewWorkAdmission(ctx context.Context, operations telegramWorkService, request core.TelegramOperationRequest) (json.RawMessage, error) {
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

	return prepared, r.store.PrepareTelegramOperation(ctx, request, prepared, time.Now())
}

func (r telegramReceiver) offerWorkAdmission(ctx context.Context, request core.TelegramOperationRequest) error {
	preview, err := telegramRetainedWorkPreview(request)
	if err != nil {
		return err
	}

	input, err := r.preparedWorkAuthorityInput(request)
	if err != nil {
		return err
	}

	complete, err := r.workConfirmationContents(ctx, request, preview)
	if errors.Is(err, core.ErrTelegramDocumentUnavailable) {
		return r.unavailableConfirmationDocument(ctx, request, input.Config)
	}

	if errors.Is(err, workflowstate.ErrConflict) {
		return r.finishWorkFailure(ctx, request, "blocked", err, input.Config)
	}

	if err != nil {
		return err
	}

	text := telegramWorkPreviewText(preview)
	text = strings.Replace(text, "\n\n", "\n\n"+strings.Join(complete.Titles[:min(3, len(complete.Titles))], "\n")+"\n", 1)

	return r.offerConfirmationDocument(ctx, request, core.TelegramActionOffer{Input: input, RequestID: request.ID, Choices: []string{"start"}, ExpiresAt: time.Now().Add(24 * time.Hour)}, "31", text, complete.Complete)
}

func (r telegramReceiver) validateWorkAdmissionAuthority(ctx context.Context, request core.TelegramOperationRequest) error {
	if err := r.validateRecordAuthority(ctx, request); err != nil {
		return err
	}

	return r.store.ValidateTelegramWorkConversation(ctx, request)
}
