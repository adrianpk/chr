package app

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type telegramControlArguments struct {
	Action     string                     `json:"action"`
	ID         string                     `json:"invocation"`
	Parameters *telegramControlParameters `json:"parameters,omitempty"`
}

type telegramControlPreview struct {
	Request WorkflowControlRequest `json:"control"`
	Status  WorkflowWorkStatus     `json:"status"`
}

func (r telegramReceiver) prepareControl(ctx context.Context, request telegramcore.OperationRequest, config telegramcore.ReceiverConfig) error {
	if telegramPreparedWorkFinished(request.State) {
		return nil
	}

	operations, ok := r.operations.(telegramControlService)
	if !ok {
		return r.finishControlPreparation(ctx, request, config)
	}

	if err := r.validateRecordAuthority(ctx, request); err != nil {
		return r.finishControlPreparation(ctx, request, config)
	}

	if request.State == "prepared" {
		input, err := r.previewControl(ctx, operations, request)
		if err != nil {
			return r.finishControlPreparation(ctx, request, config)
		}

		request.PreparedInput, request.State = input, "ready"
	}

	return r.offerControl(ctx, request)
}

func (r telegramReceiver) previewControl(ctx context.Context, operations telegramControlService, request telegramcore.OperationRequest) (json.RawMessage, error) {
	var arguments telegramControlArguments
	if err := decodeOperationJSON(string(request.Arguments), &arguments); err != nil {
		return nil, err
	}

	status, err := operations.ReadTelegramControl(ctx, request.Target, arguments.ID)
	if err != nil {
		return nil, err
	}

	control, err := telegramControlRequest(request, arguments, status)
	if err != nil {
		return nil, err
	}

	preview := telegramControlPreview{Request: control, Status: status}

	input, err := json.Marshal(preview)
	if err != nil {
		return nil, err
	}

	return input, r.store.PrepareOperation(ctx, request, input, time.Now())
}

func telegramRetainedControl(request telegramcore.OperationRequest) (telegramControlPreview, error) {
	var (
		preview   telegramControlPreview
		arguments telegramControlArguments
	)
	if err := decodeBoundedOperationJSON(string(request.PreparedInput), &preview, 1048576); err != nil {
		return preview, err
	}

	if err := decodeOperationJSON(string(request.Arguments), &arguments); err != nil {
		return preview, err
	}

	control := preview.Request

	expected, err := telegramControlRequest(request, arguments, preview.Status)
	if err != nil || !reflect.DeepEqual(control, expected) {
		return preview, workflowstate.ErrConflict
	}

	if !telegramControlStatusMatches(preview.Status, control.ID, request.Target) {
		return preview, workflowstate.ErrConflict
	}

	return preview, nil
}

func telegramControlStatusMatches(status WorkflowWorkStatus, id string, target telegramcore.OperationTarget) bool {
	return status.ID == id && status.ProjectID == target.ProjectID && status.Namespace == target.Namespace
}

func (r telegramReceiver) offerControl(ctx context.Context, request telegramcore.OperationRequest) error {
	preview, err := telegramRetainedControl(request)
	if err != nil {
		return err
	}

	input, err := r.preparedWorkAuthorityInput(request)
	if err != nil {
		return err
	}

	text := telegramControlPreviewText(preview)

	return r.offerConfirmationDocument(ctx, request, telegramcore.ActionOffer{Input: input, RequestID: request.ID, Choices: []string{preview.Request.Action}, ExpiresAt: time.Now().Add(24 * time.Hour)}, "35", text, "The complete retained control, decision options, grant and candidate binding follow.")
}

func (r telegramReceiver) finishControlPreparation(ctx context.Context, request telegramcore.OperationRequest, config telegramcore.ReceiverConfig) error {
	return r.store.FinishOperation(ctx, request, "blocked", "", json.RawMessage(`{"state":"blocked"}`), "The exact native control is unavailable for this peer, project or revision. Read current state before requesting a new control.", config.Digest(), time.Now())
}
