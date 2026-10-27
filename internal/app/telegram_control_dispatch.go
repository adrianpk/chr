package app

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (r telegramReceiver) respondControl(ctx context.Context, user telegramcore.AllowedUser, config telegramcore.ReceiverConfig, update telegramcore.ReceivedUpdate) (bool, error) {
	operationLookup, err := r.store.Operation(ctx, update, 0)
	retained, found := operationLookup.Request, operationLookup.Found

	if err != nil {
		return true, err
	}

	if found && retained.OperationID == "work.control" {
		return true, r.prepareControl(ctx, retained, config)
	}

	parsedControl, err := parseTelegramControl(update)

	command, recognized := parsedControl.Command, parsedControl.Recognized
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

	conversationLookup, err := r.store.ConversationReference(ctx, command.Input.Update.Peer, command.Invocation)

	c, found := conversationLookup.Conversation, conversationLookup.Found
	if err != nil || !found {
		return false, err
	}

	if c.InvocationID == "" {
		return true, nil
	}

	command.Input.Conversation, command.Invocation = c.ID, c.InvocationID

	return false, nil
}

func (r telegramReceiver) dispatchControl(ctx context.Context, user telegramcore.AllowedUser, command telegramControlCommand) error {
	target, err := r.resolveControlTarget(ctx, command)
	if err != nil {
		return r.recordDiagnostic(ctx, command.Input.Update, command.Input.Config, telegramIntentTargetFailure(err))
	}

	id, err := r.controlInvocation(ctx, command, target)
	if err != nil {
		if command.Action == "status" {
			return r.recordDiagnostic(ctx, command.Input.Update, command.Input.Config, "Which execution do you want to inspect? Reply to its progress or completion message; a preparation conversation has no execution yet.")
		}

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

	request, err := r.store.ReserveOperation(ctx, telegramcore.OperationRequest{Source: command.Input.Update, Operator: user.OperatorID, Target: target, OperationID: "work.control", Arguments: arguments}, time.Now())
	if err != nil {
		return err
	}

	return r.prepareControl(ctx, request, command.Input.Config)
}

func (r telegramReceiver) controlInvocation(ctx context.Context, command telegramControlCommand, target telegramcore.OperationTarget) (string, error) {
	id := command.Invocation
	if target.ConversationID != "" {
		conversationLookup2, err := r.store.ConversationReference(ctx, command.Input.Update.Peer, target.ConversationID)

		c, found := conversationLookup2.Conversation, conversationLookup2.Found
		if err != nil || !found || id != "" && c.InvocationID != "" && id != c.InvocationID {
			return "", workflowstate.ErrConflict
		}

		if c.InvocationID != "" {
			id = c.InvocationID
		}
	}

	if !validWorkflowControlIdentity(id) {
		return "", workflowstate.ErrConflict
	}

	return id, nil
}

func (r telegramReceiver) resolveControlTarget(ctx context.Context, command telegramControlCommand) (telegramcore.OperationTarget, error) {
	if command.Action == "status" {
		return r.store.ResolveWorkStatusTarget(ctx, command.Input, time.Now())
	}

	return r.store.ResolveControlTarget(ctx, command.Input, time.Now())
}

func (r telegramReceiver) respondNativeStatus(ctx context.Context, operations telegramControlService, target telegramcore.OperationTarget, id string, input telegramcore.ConversationInput) error {
	status, err := operations.ReadTelegramControl(ctx, target, id)
	if err != nil {
		return r.recordDiagnostic(ctx, input.Update, input.Config, "The exact native execution is unavailable in this project and namespace.")
	}

	if status.ID != id || status.ProjectID != target.ProjectID || status.Namespace != target.Namespace {
		return workflowstate.ErrConflict
	}

	if status.Integration != nil && status.Integration.Decision != nil && len(status.Integration.Decision.Requests) > 0 {
		response, err := telegramQuestionDocument(target, status)
		if err != nil {
			return err
		}

		err = r.store.QueueDocumentResponse(ctx, input.Update, response, input.Config.Digest(), time.Now())
		if errors.Is(err, tgmessage.ErrDocumentUnavailable) {
			return r.recordDiagnostic(ctx, input.Update, input.Config, "The complete questions export could not be attached under the document or storage limits. Inspect canonical work status locally before answering.")
		}

		return err
	}

	return r.store.QueueWorkStatus(ctx, input, target, id, telegramWorkProgressMessage(status, telegramWorkProjectName(input.Projects, status.ProjectID)), time.Now())
}
