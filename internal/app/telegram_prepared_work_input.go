package app

import (
	"context"
	"encoding/json"
	"time"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (r telegramReceiver) prepareTelegramWork(ctx context.Context, operations telegramPreparedWorkService, request telegramcore.OperationRequest) (PreparedAuthoringWork, error) {
	if request.State != telegramcore.OperationPrepared {
		return telegramRetainedPreparedWork(request)
	}

	var owner AuthoringWorkRequest
	if err := decodeOperationJSON(string(request.Arguments), &owner); err != nil {
		return PreparedAuthoringWork{}, err
	}

	if owner.Binding != (workflowstate.MutationBinding{RequestID: request.ID, Operator: request.Operator}) {
		return PreparedAuthoringWork{}, workflowstate.ErrConflict
	}

	prepared, err := operations.PrepareTelegramAuthoringWork(ctx, request.Target, owner)
	if err != nil {
		return prepared, err
	}

	prepared, err = r.bindInitialPreparedWorkAuthority(ctx, request, prepared)
	if err != nil {
		return prepared, err
	}

	input, err := json.Marshal(prepared)
	if err != nil {
		return prepared, err
	}

	if err := r.store.PrepareOperation(ctx, request, input, time.Now()); err != nil {
		return prepared, err
	}

	return prepared, nil
}

func (r telegramReceiver) bindInitialPreparedWorkAuthority(ctx context.Context, request telegramcore.OperationRequest, prepared PreparedAuthoringWork) (PreparedAuthoringWork, error) {
	if prepared.Scope.Intent != authoringstate.IntentDevelop {
		return prepared, nil
	}

	initial, err := r.store.IsInitialAuthoringWork(ctx, request)
	if err != nil || !initial {
		return prepared, err
	}

	input, err := r.preparedWorkAuthorityInput(request)
	if err != nil {
		return prepared, err
	}

	instruction, err := r.store.ReadConversationInstruction(ctx, input)
	if err != nil {
		return prepared, err
	}

	if instruction.DevelopmentRef == "" {
		return prepared, nil
	}

	grant, err := authoringcore.AuthorizeAuthoringWork(prepared.Scope, workflowstate.MutationBinding{RequestID: instruction.DevelopmentRef, Operator: request.Operator}, prepared.Scope.Hash)
	prepared.Authorization = &grant

	return prepared, err
}

func (r telegramReceiver) preparedWorkAuthorityInput(request telegramcore.OperationRequest) (telegramcore.ConversationInput, error) {
	config, err := r.currentConfig()
	if err != nil {
		return telegramcore.ConversationInput{}, err
	}

	user, allowed := config.AuthorizedUser(request.Source.Peer.ChatID, request.Source.Peer.TopicID, request.Source.Peer.UserID, request.Source.ChatType)
	if !allowed || user.OperatorID != request.Operator {
		return telegramcore.ConversationInput{}, workflowstate.ErrConflict
	}

	projects, err := r.operations.ListRegisteredProjects(user.ProjectIDs)

	return telegramcore.ConversationInput{Update: request.Source, Config: config, Projects: projects, Conversation: request.Target.ConversationID}, err
}

func telegramRetainedPreparedWork(request telegramcore.OperationRequest) (PreparedAuthoringWork, error) {
	var prepared PreparedAuthoringWork
	if err := decodeBoundedOperationJSON(string(request.PreparedInput), &prepared, 1048576); err != nil {
		return prepared, err
	}

	var owner AuthoringWorkRequest
	if err := decodeOperationJSON(string(request.Arguments), &owner); err != nil {
		return prepared, err
	}

	scope := prepared.Scope
	if err := authoringcore.ValidateAuthoringWorkScope(scope); err != nil {
		return prepared, err
	}

	if scope.Binding != (workflowstate.MutationBinding{RequestID: request.ID, Operator: request.Operator}) || scope.ProjectID != request.Target.ProjectID || scope.Namespace != request.Target.Namespace || telegramWorkScopeRequest(scope) != owner {
		return prepared, workflowstate.ErrConflict
	}

	return prepared, nil
}

func telegramWorkScopeRequest(scope authoringcore.AuthoringWorkScope) AuthoringWorkRequest {
	return AuthoringWorkRequest{Binding: scope.Binding, SessionID: scope.SessionID, ExpectedRevision: scope.SessionRevision, ExpectedSourceHash: scope.SourceHash, Intent: scope.Intent, Instruction: scope.Instruction}
}
