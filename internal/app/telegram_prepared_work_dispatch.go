package app

import (
	"context"
	"encoding/json"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func (r telegramReceiver) respondPreparedWork(ctx context.Context, user core.TelegramAllowedUser, config core.TelegramReceiverConfig, update core.TelegramReceivedUpdate) (bool, error) {
	retained, found, err := r.store.TelegramOperation(ctx, update, 0)
	if err != nil {
		return true, err
	}

	if found && retained.OperationID == "authoring.work" {
		return true, r.continuePreparedWork(ctx, retained, config)
	}

	input, recognized, err := parseTelegramPreparedWork(update)
	if !recognized {
		return false, nil
	}

	if err != nil {
		return true, r.recordDiagnostic(ctx, update, config, telegramPreparedWorkHelp())
	}

	input.Config = config

	input.Projects, err = r.operations.ListRegisteredProjects(user.ProjectIDs)
	if err != nil {
		return true, err
	}

	return true, r.dispatchPreparedWork(ctx, user, input)
}

func (r telegramReceiver) dispatchPreparedWork(ctx context.Context, user core.TelegramAllowedUser, input core.TelegramConversationInput) error {
	operations, ok := r.operations.(telegramAuthoringService)
	if !ok {
		return r.recordDiagnostic(ctx, input.Update, input.Config, "Shared authoring is unavailable. Retained context is unchanged.")
	}

	c, target, err := r.store.ResolveTelegramAuthoringTarget(ctx, input)
	if err != nil {
		return r.recordDiagnostic(ctx, input.Update, input.Config, telegramAuthoringFailure(err))
	}

	owner, err := r.preparedWorkInput(ctx, operations, input, c, target, user.OperatorID)
	if err != nil {
		return r.recordDiagnostic(ctx, input.Update, input.Config, telegramAuthoringFailure(err))
	}

	if reused, err := r.reusePreparedWork(ctx, input, target, owner); reused || err != nil {
		return err
	}

	arguments, err := json.Marshal(owner)
	if err != nil {
		return err
	}

	request, err := r.store.ReserveTelegramOperation(ctx, core.TelegramOperationRequest{Source: input.Update, Operator: user.OperatorID, Target: target, OperationID: "authoring.work", Arguments: arguments}, time.Now())
	if err != nil {
		return err
	}

	return r.continuePreparedWork(ctx, request, input.Config)
}

func (r telegramReceiver) preparedWorkInput(ctx context.Context, operations telegramAuthoringService, input core.TelegramConversationInput, c core.TelegramConversation, target core.TelegramOperationTarget, operator string) (AuthoringWorkRequest, error) {
	var owner AuthoringWorkRequest

	id, err := telegramAuthoringSessionID(c.AuthoringSessionRef)
	if err != nil {
		return owner, err
	}

	snapshot, err := operations.ObserveTelegramAuthoring(ctx, target, id)
	if err != nil {
		return owner, err
	}

	if snapshot.Session.Namespace != target.Namespace || snapshot.Session.ProjectID != target.ProjectID || snapshot.Question != nil {
		return owner, core.ErrWorkflowConflict
	}

	input.Conversation = c.ID

	instruction, err := r.store.ReadTelegramConversationInstruction(ctx, input)
	if err != nil {
		return owner, err
	}

	return AuthoringWorkRequest{Binding: core.WorkflowMutationBinding{RequestID: core.TelegramOperationID(input.Update, 0), Operator: operator}, SessionID: id, ExpectedRevision: snapshot.Session.Revision, ExpectedSourceHash: snapshot.SourceHash, Intent: core.AuthoringIntent(c.Intent), Instruction: instruction.Text}, nil
}

func (r telegramReceiver) reusePreparedWork(ctx context.Context, input core.TelegramConversationInput, target core.TelegramOperationTarget, owner AuthoringWorkRequest) (bool, error) {
	previous, found, err := r.store.LatestTelegramAuthoringWork(ctx, input, target)
	if err != nil {
		return true, err
	}

	if !found || (previous.State != "confirmed" && previous.State != "awaiting_confirmation") {
		return false, nil
	}

	prepared, err := telegramRetainedPreparedWork(previous)
	if err != nil {
		return true, err
	}

	scope := prepared.Scope

	owner.Binding.RequestID = scope.Binding.RequestID
	if previous.Target.ProjectBindingDigest != target.ProjectBindingDigest || telegramWorkScopeRequest(scope) != owner {
		return false, nil
	}

	return r.reusePreparedWorkResult(ctx, input, previous, prepared)
}

func (r telegramReceiver) reusePreparedWorkResult(ctx context.Context, input core.TelegramConversationInput, previous core.TelegramOperationRequest, prepared PreparedAuthoringWork) (bool, error) {
	if previous.State == "confirmed" {
		return true, r.recordDiagnostic(ctx, input.Update, input.Config, "Previously prepared native work is confirmed by receipt "+previous.ReceiptRef+".\n"+string(previous.Result))
	}

	action, err := r.store.TelegramRequestAction(ctx, previous)
	if err != nil {
		return true, err
	}

	if action.State != "pending" || !action.ExpiresAt.After(time.Now()) {
		return false, nil
	}

	text := "This exact scope is awaiting a decision. Use /answer " + action.Token + " develop.\n" + telegramPreparedWorkPreview(prepared)

	return true, r.recordDiagnostic(ctx, input.Update, input.Config, text)
}
