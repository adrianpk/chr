package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

var errTelegramRecordTarget = errors.New("telegram record target is unavailable")

func (r telegramReceiver) respondRecord(ctx context.Context, user telegramcore.AllowedUser, config telegramcore.ReceiverConfig, update telegramcore.ReceivedUpdate) (bool, error) {
	parsedRecord, parseErr := parseTelegramRecord(update)

	command, recognized := parsedRecord.Command, parsedRecord.Recognized
	if !recognized {
		return false, nil
	}

	if parseErr != nil {
		return true, r.recordDiagnostic(ctx, update, config, "The record request is malformed. Use /help for its JSON fields; read the current record before retrying an update. No record was changed.")
	}

	command.Input.Config = config

	projects, err := r.operations.ListRegisteredProjects(user.ProjectIDs)
	if err != nil {
		return true, err
	}

	command.Input.Projects = projects

	return true, r.dispatchRecord(ctx, user, command)
}

func (r telegramReceiver) dispatchRecord(ctx context.Context, user telegramcore.AllowedUser, command telegramRecordCommand) error {
	operations, ok := r.operations.(telegramRecordService)
	if !ok {
		return r.recordDiagnostic(ctx, command.Input.Update, command.Input.Config, "Record operations are unavailable.")
	}

	if command.Action == "create" || command.Action == "update" {
		return r.mutateRecord(ctx, operations, user, command)
	}

	return r.readRecord(ctx, operations, command)
}

func (r telegramReceiver) readRecord(ctx context.Context, operations telegramRecordService, command telegramRecordCommand) error {
	update, config := command.Input.Update, command.Input.Config

	target, err := r.store.ResolveOperationTarget(ctx, command.Input, time.Now())
	if err != nil {
		return r.recordDiagnostic(ctx, update, config, "The record target is unavailable or conflicts with its reply. Select an authorized current project/session explicitly.")
	}

	page, err := operations.ReadTelegramRecords(ctx, target, command.Query)
	if err != nil {
		return r.recordDiagnostic(ctx, update, config, telegramRecordFailure(err))
	}

	if len(page.Records) == 0 {
		return r.store.QueuePresentationResponse(ctx, update, tgmessage.Blocks{Family: "28", InitialText: "No records matched this view."}, config.Digest(), time.Now())
	}

	response, err := telegramRecordDocument(target, command, page)
	if err != nil {
		return err
	}

	err = r.store.QueueDocumentResponse(ctx, update, response, config.Digest(), time.Now())
	if errors.Is(err, tgmessage.ErrDocumentUnavailable) {
		return r.recordDiagnostic(ctx, update, config, "The records export could not be attached: its image is invalid or document storage is full. Use local project commands to inspect the records.")
	}

	return err
}

func (r telegramReceiver) recordDiagnostic(ctx context.Context, update telegramcore.ReceivedUpdate, config telegramcore.ReceiverConfig, text string) error {
	return r.store.QueueReadResponse(ctx, update, text, config.Digest(), time.Now())
}

func telegramRecordFailure(err error) string {
	if errors.Is(err, workflowstate.ErrConflict) {
		return "Record conflict: the observed revision, source image or lifecycle changed. Read the current record before issuing a new request."
	}

	if errors.Is(err, ErrRecordUnsupported) {
		return "This record operation is unsupported by the current lifecycle owner. No lifecycle transition was authorized."
	}

	return "The record owner is unavailable. Inspect local diagnostics; the retained request cannot be treated as successful."
}

func (r telegramReceiver) mutateRecord(ctx context.Context, operations telegramRecordService, user telegramcore.AllowedUser, command telegramRecordCommand) error {
	operationLookup, err := r.store.Operation(ctx, command.Input.Update, 0)
	request, found := operationLookup.Request, operationLookup.Found

	if err != nil {
		return err
	}

	if !found {
		request, err = r.reserveRecord(ctx, user, command)
		if errors.Is(err, errTelegramRecordTarget) {
			return r.recordDiagnostic(ctx, command.Input.Update, command.Input.Config, "The mutation target is unavailable or ambiguous. Select an authorized current project/session explicitly.")
		}

		if err != nil {
			return err
		}
	}

	switch request.State {
	case telegramcore.OperationConfirmed, telegramcore.OperationRejected, telegramcore.OperationBlocked:
		return nil
	}

	if err := r.validateRecordAuthority(ctx, request); err != nil {
		return r.finishRecordFailure(ctx, request, telegramcore.OperationBlocked, err, command.Input.Config)
	}

	if request.State == telegramcore.OperationPrepared {
		request, err = r.prepareRecord(ctx, operations, request)
		if err != nil {
			return r.finishRecordFailure(ctx, request, telegramcore.OperationRejected, err, command.Input.Config)
		}
	}

	return r.applyRecord(ctx, operations, request, command.Input.Config)
}

func (r telegramReceiver) reserveRecord(ctx context.Context, user telegramcore.AllowedUser, command telegramRecordCommand) (telegramcore.OperationRequest, error) {
	target, err := r.store.ResolveOperationTarget(ctx, command.Input, time.Now())
	if err != nil {
		return telegramcore.OperationRequest{}, fmt.Errorf("%w: %w", errTelegramRecordTarget, err)
	}

	arguments, err := json.Marshal(command.Mutation)
	if err != nil {
		return telegramcore.OperationRequest{}, err
	}

	return r.store.ReserveOperation(ctx, telegramcore.OperationRequest{Source: command.Input.Update, Operator: user.OperatorID, Target: target, OperationID: fmt.Sprintf("records.%s.%s", command.Kind, command.Action), Arguments: arguments}, time.Now())
}
