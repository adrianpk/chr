package app

import (
	"context"
	"encoding/json"
	"time"

	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (r telegramReceiver) respondPreparedWork(ctx context.Context, user telegramcore.AllowedUser, config telegramcore.ReceiverConfig, update telegramcore.ReceivedUpdate) (bool, error) {
	operationLookup, err := r.store.Operation(ctx, update, 0)
	retained, found := operationLookup.Request, operationLookup.Found

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

func (r telegramReceiver) dispatchPreparedWork(ctx context.Context, user telegramcore.AllowedUser, input telegramcore.ConversationInput) error {
	operations, ok := r.operations.(telegramAuthoringService)
	if !ok {
		return r.recordDiagnostic(ctx, input.Update, input.Config, "Shared authoring is unavailable. Retained context is unchanged.")
	}

	bindingResult, err := r.store.ResolveAuthoringTarget(ctx, input)
	c, target := bindingResult.Conversation, bindingResult.Target

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

	request, err := r.store.ReserveOperation(ctx, telegramcore.OperationRequest{Source: input.Update, Operator: user.OperatorID, Target: target, OperationID: "authoring.work", Arguments: arguments}, time.Now())
	if err != nil {
		return err
	}

	return r.continuePreparedWork(ctx, request, input.Config)
}

func (r telegramReceiver) preparedWorkInput(ctx context.Context, operations telegramAuthoringService, input telegramcore.ConversationInput, c telegramcore.Conversation, target telegramcore.OperationTarget, operator string) (AuthoringWorkRequest, error) {
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
		return owner, workflowstate.ErrConflict
	}

	input.Conversation = c.ID

	instruction, err := r.store.ReadConversationInstruction(ctx, input)
	if err != nil {
		return owner, err
	}

	return AuthoringWorkRequest{Binding: workflowstate.MutationBinding{RequestID: telegramcore.OperationID(input.Update, 0), Operator: operator}, SessionID: id, ExpectedRevision: snapshot.Session.Revision, ExpectedSourceHash: snapshot.SourceHash, Intent: authoringstate.Intent(c.Intent), Instruction: instruction.Text}, nil
}

func (r telegramReceiver) reusePreparedWork(ctx context.Context, input telegramcore.ConversationInput, target telegramcore.OperationTarget, owner AuthoringWorkRequest) (bool, error) {
	operationLookup2, err := r.store.LatestAuthoringWork(ctx, input, target)
	previous, found := operationLookup2.Request, operationLookup2.Found

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

func (r telegramReceiver) reusePreparedWorkResult(ctx context.Context, input telegramcore.ConversationInput, previous telegramcore.OperationRequest, prepared PreparedAuthoringWork) (bool, error) {
	if previous.State == "confirmed" {
		return true, r.store.QueuePresentationResponse(ctx, input.Update, *telegramOutcomeMessage("26", telegramPreparedWorkReplay(previous)), input.Config.Digest(), time.Now())
	}

	action, err := r.store.RequestAction(ctx, previous)
	if err != nil {
		return true, err
	}

	if action.State != "pending" || !action.ExpiresAt.After(time.Now()) {
		return false, nil
	}

	text := "This exact scope is awaiting a decision. Use /answer " + action.Token + " develop.\n" + telegramPreparedWorkPreview(prepared)

	return true, r.store.QueuePresentationResponse(ctx, input.Update, *telegramOutcomeMessage("25", text), input.Config.Digest(), time.Now())
}
