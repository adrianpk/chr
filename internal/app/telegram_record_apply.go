package app

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (r telegramReceiver) validateRecordAuthority(ctx context.Context, request core.TelegramOperationRequest) error {
	return r.validateRecordTargetAuthority(ctx, request, false)
}

func (r telegramReceiver) validateRecordTargetAuthority(ctx context.Context, request core.TelegramOperationRequest, receiptOnly bool) error {
	config, err := r.currentConfig()
	if err != nil {
		return err
	}

	user, allowed := config.AuthorizedUser(request.Source.Peer.ChatID, request.Source.Peer.TopicID, request.Source.Peer.UserID, request.Source.ChatType)
	if !allowed || user.OperatorID != request.Operator {
		return workflowstate.ErrConflict
	}

	projects, err := r.operations.ListRegisteredProjects(user.ProjectIDs)
	if err != nil {
		return err
	}

	input := core.TelegramConversationInput{Update: request.Source, Config: config, Projects: projects}
	if receiptOnly {
		return r.store.ValidateTelegramControlReceiptTarget(ctx, input, request)
	}

	return r.store.ValidateTelegramOperationTarget(ctx, input, request.Target)
}

func (r telegramReceiver) prepareRecord(ctx context.Context, operations telegramRecordService, request core.TelegramOperationRequest) (core.TelegramOperationRequest, error) {
	var mutation RecordMutationRequest
	if err := json.Unmarshal(request.Arguments, &mutation); err != nil {
		return request, err
	}

	prepared, err := operations.PrepareTelegramRecord(ctx, request.Target, workflowstate.MutationBinding{RequestID: request.ID, Operator: request.Operator}, mutation)
	if err != nil {
		return request, err
	}

	input, err := json.Marshal(prepared)
	if err != nil {
		return request, err
	}

	if err := r.store.PrepareTelegramOperation(ctx, request, input, time.Now()); err != nil {
		return request, err
	}

	request.PreparedInput, request.State = input, "ready"

	return request, nil
}

func (r telegramReceiver) applyRecord(ctx context.Context, operations telegramRecordService, request core.TelegramOperationRequest, config core.TelegramReceiverConfig) error {
	if err := r.validateRecordAuthority(ctx, request); err != nil {
		return r.finishRecordFailure(ctx, request, "blocked", err, config)
	}

	var prepared PreparedRecordMutation
	if err := json.Unmarshal(request.PreparedInput, &prepared); err != nil {
		return err
	}

	if !telegramPreparedRecordBinding(prepared, request) {
		return workflowstate.ErrConflict
	}

	if err := r.store.BeginTelegramOperation(ctx, request, time.Now()); err != nil {
		return err
	}

	result, err := operations.ApplyTelegramRecord(ctx, request.Target, prepared)
	if errors.Is(err, workflowstate.ErrConflict) || errors.Is(err, ErrRecordUnsupported) {
		return r.finishRecordFailure(ctx, request, "blocked", err, config)
	}

	if err != nil {
		return err
	}

	result.Record.Body = ""

	facts, err := json.Marshal(result)
	if err != nil {
		return err
	}

	text := telegramRecordResultText(result)
	request.Message = telegramOutcomeMessage("29", text)

	return r.store.FinishTelegramOperation(ctx, request, "confirmed", result.ReceiptRef, facts, text, config.Digest(), time.Now())
}

func telegramPreparedRecordBinding(prepared PreparedRecordMutation, request core.TelegramOperationRequest) bool {
	expected := workflowstate.MutationBinding{RequestID: request.ID, Operator: request.Operator}
	if prepared.Native != nil {
		return prepared.Native.Binding == expected && prepared.Case == nil
	}

	return prepared.Case != nil && prepared.Case.Binding == expected
}

func (r telegramReceiver) finishRecordFailure(ctx context.Context, request core.TelegramOperationRequest, state string, failure error, config core.TelegramReceiverConfig) error {
	text := telegramRecordFailure(failure)
	request.Message = telegramOutcomeMessage("30", text)

	facts, err := json.Marshal(struct {
		State   string `json:"state"`
		Message string `json:"message"`
	}{state, text})
	if err != nil {
		return err
	}

	return r.store.FinishTelegramOperation(ctx, request, state, "", facts, text, config.Digest(), time.Now())
}
