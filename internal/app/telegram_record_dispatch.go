package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

var errTelegramRecordTarget = errors.New("telegram record target is unavailable")

func (r telegramReceiver) respondRecord(ctx context.Context, user core.TelegramAllowedUser, config core.TelegramReceiverConfig, update core.TelegramReceivedUpdate) (bool, error) {
	command, recognized, parseErr := parseTelegramRecord(update)
	if !recognized {
		return false, nil
	}

	if parseErr != nil {
		return true, r.recordDiagnostic(ctx, update, config, "Unsupported record arguments. Use /help and strict JSON with the displayed revision/source_hash guards.")
	}

	operations, ok := r.operations.(telegramRecordService)
	if !ok {
		return true, r.recordDiagnostic(ctx, update, config, "Record operations are unavailable.")
	}

	command.Input.Config = config

	projects, err := r.operations.ListRegisteredProjects(user.ProjectIDs)
	if err != nil {
		return true, err
	}

	command.Input.Projects = projects
	if command.Action == "create" || command.Action == "update" {
		return true, r.mutateRecord(ctx, operations, user, command)
	}

	return true, r.readRecord(ctx, operations, command)
}

func (r telegramReceiver) readRecord(ctx context.Context, operations telegramRecordService, command telegramRecordCommand) error {
	update, config := command.Input.Update, command.Input.Config

	target, err := r.store.ResolveTelegramOperationTarget(ctx, command.Input, time.Now())
	if err != nil {
		return r.recordDiagnostic(ctx, update, config, "The record target is unavailable or conflicts with its reply. Select an authorized current project/session explicitly.")
	}

	page, err := operations.ReadTelegramRecords(ctx, target, command.Query)
	if err != nil {
		return r.recordDiagnostic(ctx, update, config, telegramRecordFailure(err))
	}

	text, err := json.MarshalIndent(page, "", "  ")
	if err != nil {
		return err
	}

	reference := fmt.Sprintf("project=%s ops/%s %s offset=%d", target.ProjectID, target.Namespace, command.Kind, command.Query.Offset)
	if len(page.Records) == 1 && command.Query.Identity != "" {
		reference = target.ProjectID + " " + page.Records[0].Reference
	}

	return r.store.QueueReferencedResponse(ctx, update, string(text), reference, config.Digest(), time.Now())
}

func (r telegramReceiver) recordDiagnostic(ctx context.Context, update core.TelegramReceivedUpdate, config core.TelegramReceiverConfig, text string) error {
	return r.store.QueueReadResponse(ctx, update, text, config.Digest(), time.Now())
}

func telegramRecordFailure(err error) string {
	if errors.Is(err, core.ErrWorkflowConflict) {
		return "Record conflict: the observed revision, source image or lifecycle changed. Read the current record before issuing a new request."
	}

	if errors.Is(err, ErrRecordUnsupported) {
		return "This record operation is unsupported by the current lifecycle owner. No lifecycle transition was authorized."
	}

	return "The record owner is unavailable. Inspect local diagnostics; the retained request cannot be treated as successful."
}

func (r telegramReceiver) mutateRecord(ctx context.Context, operations telegramRecordService, user core.TelegramAllowedUser, command telegramRecordCommand) error {
	request, found, err := r.store.TelegramOperation(ctx, command.Input.Update, 0)
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
	case "confirmed", "rejected", "blocked":
		return nil
	}

	if err := r.validateRecordAuthority(ctx, request); err != nil {
		return r.finishRecordFailure(ctx, request, "blocked", err, command.Input.Config)
	}

	if request.State == "prepared" {
		request, err = r.prepareRecord(ctx, operations, request)
		if err != nil {
			return r.finishRecordFailure(ctx, request, "rejected", err, command.Input.Config)
		}
	}

	return r.applyRecord(ctx, operations, request, command.Input.Config)
}

func (r telegramReceiver) reserveRecord(ctx context.Context, user core.TelegramAllowedUser, command telegramRecordCommand) (core.TelegramOperationRequest, error) {
	target, err := r.store.ResolveTelegramOperationTarget(ctx, command.Input, time.Now())
	if err != nil {
		return core.TelegramOperationRequest{}, fmt.Errorf("%w: %w", errTelegramRecordTarget, err)
	}

	arguments, err := json.Marshal(command.Mutation)
	if err != nil {
		return core.TelegramOperationRequest{}, err
	}

	return r.store.ReserveTelegramOperation(ctx, core.TelegramOperationRequest{Source: command.Input.Update, Operator: user.OperatorID, Target: target, OperationID: fmt.Sprintf("records.%s.%s", command.Kind, command.Action), Arguments: arguments}, time.Now())
}
