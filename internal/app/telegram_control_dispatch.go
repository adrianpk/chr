package app

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func (r telegramReceiver) respondControl(ctx context.Context, user core.TelegramAllowedUser, config core.TelegramReceiverConfig, update core.TelegramReceivedUpdate) (bool, error) {
	retained, found, err := r.store.TelegramOperation(ctx, update, 0)
	if err != nil {
		return true, err
	}

	if found && retained.OperationID == "work.control" {
		return true, r.prepareControl(ctx, retained, config)
	}

	command, recognized, err := parseTelegramControl(update)
	if !recognized {
		return false, nil
	}

	if err != nil {
		return true, r.recordDiagnostic(ctx, update, config, telegramControlHelp())
	}

	command.Input.Config = config

	command.Input.Projects, err = r.operations.ListRegisteredProjects(user.ProjectIDs)
	if err != nil {
		return true, err
	}

	legacy, err := r.controlCancelConversation(ctx, &command)
	if err != nil || legacy {
		return !legacy, err
	}

	return true, r.dispatchControl(ctx, user, command)
}

func (r telegramReceiver) controlCancelConversation(ctx context.Context, command *telegramControlCommand) (bool, error) {
	if command.Action != "cancel" || command.Invocation == "" || command.Input.Conversation != "" {
		return false, nil
	}

	c, found, err := r.store.TelegramConversationReference(ctx, command.Input.Update.Peer, command.Invocation)
	if err != nil || !found {
		return false, err
	}

	if c.InvocationID == "" {
		return true, nil
	}

	command.Input.Conversation, command.Invocation = c.ID, c.InvocationID

	return false, nil
}

func (r telegramReceiver) dispatchControl(ctx context.Context, user core.TelegramAllowedUser, command telegramControlCommand) error {
	target, err := r.resolveControlTarget(ctx, command)
	if err != nil {
		return r.recordDiagnostic(ctx, command.Input.Update, command.Input.Config, telegramIntentTargetFailure(err))
	}

	id, err := r.controlInvocation(ctx, command, target)
	if err != nil {
		return r.recordDiagnostic(ctx, command.Input.Update, command.Input.Config, "Choose the exact execution or a conversation bound to it; conflicting targets cannot control work.")
	}

	operations, ok := r.operations.(telegramControlService)
	if !ok {
		return r.recordDiagnostic(ctx, command.Input.Update, command.Input.Config, "Native controls are unavailable through this operation service.")
	}

	if command.Action == "status" {
		return r.respondNativeStatus(ctx, operations, target, id, command.Input)
	}

	arguments, err := json.Marshal(telegramControlArguments{Action: command.Action, ID: id, Parameters: command.Parameters})
	if err != nil {
		return err
	}

	request, err := r.store.ReserveTelegramOperation(ctx, core.TelegramOperationRequest{Source: command.Input.Update, Operator: user.OperatorID, Target: target, OperationID: "work.control", Arguments: arguments}, time.Now())
	if err != nil {
		return err
	}

	return r.prepareControl(ctx, request, command.Input.Config)
}

func (r telegramReceiver) controlInvocation(ctx context.Context, command telegramControlCommand, target core.TelegramOperationTarget) (string, error) {
	id := command.Invocation
	if target.ConversationID != "" {
		c, found, err := r.store.TelegramConversationReference(ctx, command.Input.Update.Peer, target.ConversationID)
		if err != nil || !found || id != "" && c.InvocationID != "" && id != c.InvocationID {
			return "", core.ErrWorkflowConflict
		}

		if c.InvocationID != "" {
			id = c.InvocationID
		}
	}

	if !validWorkflowControlIdentity(id) {
		return "", core.ErrWorkflowConflict
	}

	return id, nil
}

func (r telegramReceiver) resolveControlTarget(ctx context.Context, command telegramControlCommand) (core.TelegramOperationTarget, error) {
	if command.Action == "status" {
		return r.store.ResolveTelegramOperationTarget(ctx, command.Input, time.Now())
	}

	return r.store.ResolveTelegramControlTarget(ctx, command.Input, time.Now())
}

func (r telegramReceiver) respondNativeStatus(ctx context.Context, operations telegramControlService, target core.TelegramOperationTarget, id string, input core.TelegramConversationInput) error {
	status, err := operations.ReadTelegramControl(ctx, target, id)
	if err != nil {
		return r.recordDiagnostic(ctx, input.Update, input.Config, "The exact native execution is unavailable in this project and namespace.")
	}

	if status.ID != id || status.ProjectID != target.ProjectID || status.Namespace != target.Namespace {
		return core.ErrWorkflowConflict
	}

	if status.Integration != nil && status.Integration.Decision != nil && len(status.Integration.Decision.Requests) > 0 {
		response, err := telegramQuestionDocument(target, status)
		if err != nil {
			return err
		}

		err = r.store.QueueDocumentResponse(ctx, input.Update, response, input.Config.Digest(), time.Now())
		if errors.Is(err, core.ErrTelegramDocumentUnavailable) {
			return r.recordDiagnostic(ctx, input.Update, input.Config, "The complete questions export could not be attached under the document or storage limits. Inspect canonical work status locally before answering.")
		}

		return err
	}

	return r.store.QueuePresentationResponse(ctx, input.Update, telegramWorkProgressMessage(status, telegramWorkProjectName(input.Projects, status.ProjectID)), input.Config.Digest(), time.Now())
}
