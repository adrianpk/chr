package app

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"time"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

type telegramAuthoringArguments struct {
	PayloadHash string `json:"payload_hash"`
}

func (r telegramReceiver) reserveAuthoring(ctx context.Context, operations telegramAuthoringService, user telegramcore.AllowedUser, command telegramAuthoringCommand) (telegramcore.OperationRequest, error) {
	bindingResult, err := r.store.ResolveAuthoringTarget(ctx, command.Input)
	c, target := bindingResult.Conversation, bindingResult.Target

	if err != nil {
		return telegramcore.OperationRequest{}, err
	}

	owner, err := r.authoringOwnerInput(ctx, operations, command, c, target)
	if err != nil {
		return telegramcore.OperationRequest{}, err
	}

	owner.Binding = workflowstate.MutationBinding{RequestID: telegramcore.OperationID(command.Input.Update, 0), Operator: user.OperatorID}

	input, err := json.Marshal(owner)
	if err != nil {
		return telegramcore.OperationRequest{}, err
	}

	arguments, err := json.Marshal(telegramAuthoringArguments{PayloadHash: recordDigest(input)})
	if err != nil {
		return telegramcore.OperationRequest{}, err
	}

	return r.store.ReservePreparedOperation(ctx, telegramcore.OperationRequest{Source: command.Input.Update, Operator: user.OperatorID, Target: target, OperationID: "authoring.continue", Arguments: arguments}, input, time.Now())
}

func (r telegramReceiver) authoringOwnerInput(ctx context.Context, operations telegramAuthoringService, command telegramAuthoringCommand, c telegramcore.Conversation, target telegramcore.OperationTarget) (authoringcore.ContinuationRequest, error) {
	request := authoringcore.ContinuationRequest{ProjectID: target.ProjectID, Namespace: target.Namespace, Action: command.Action, Target: command.Target, Instructions: command.Instructions}
	if c.AuthoringSessionRef == "" {
		return r.initialAuthoringInput(ctx, command, c, request)
	}

	id, err := telegramAuthoringSessionID(c.AuthoringSessionRef)
	if err != nil {
		return request, err
	}

	snapshot, err := operations.ObserveTelegramAuthoring(ctx, target, id)
	if err != nil {
		return request, err
	}

	if snapshot.SessionRef != c.AuthoringSessionRef || snapshot.Session.Namespace != target.Namespace || snapshot.Session.ProjectID != target.ProjectID {
		return request, workflowstate.ErrConflict
	}

	request.SessionID, request.ExpectedRevision, request.ExpectedSourceHash = id, snapshot.Session.Revision, snapshot.SourceHash
	if command.Action == authoringcore.ContinuationAnswer {
		return telegramAuthoringAnswerInput(request, command.Input.Text, c.QuestionBinding, snapshot)
	}

	return request, nil
}

func (r telegramReceiver) initialAuthoringInput(ctx context.Context, command telegramAuthoringCommand, c telegramcore.Conversation, request authoringcore.ContinuationRequest) (authoringcore.ContinuationRequest, error) {
	if command.Action == authoringcore.ContinuationRefine || command.Action == authoringcore.ContinuationAnswer {
		return request, workflowstate.ErrConflict
	}

	input := command.Input
	input.Conversation = c.ID

	instruction, err := r.store.ReadConversationInstruction(ctx, input)
	if err != nil {
		return request, err
	}

	request.Action, request.Topic, request.Capture = authoringcore.ContinuationCreate, c.Topic, c.Capture

	request.OriginalInstruction = instruction.Text
	if request.Capture == "" {
		request.Capture = c.Topic
	}

	return request, nil
}

func telegramAuthoringAnswerInput(request authoringcore.ContinuationRequest, answer, binding string, snapshot authoringcore.Snapshot) (authoringcore.ContinuationRequest, error) {
	if snapshot.Question == nil {
		return request, workflowstate.ErrConflict
	}

	var question telegramcore.ConversationQuestion
	if err := decodeOperationJSON(binding, &question); err != nil {
		return request, err
	}

	expected := telegramAuthoringQuestion(snapshot.SessionRef, snapshot.Question)
	if question != *expected {
		return request, workflowstate.ErrConflict
	}

	request.Target = authoringcore.ArtifactKind(snapshot.Session.Target)

	request.Question, request.Instructions = snapshot.Question, answer
	if request.Target != authoringcore.ArtifactKindQuiz {
		request.Action = authoringcore.ContinuationRespond
	}

	return request, nil
}

func telegramAuthoringQuestion(ref string, question *authoringcore.QuestionBinding) *telegramcore.ConversationQuestion {
	if question == nil {
		return nil
	}

	return &telegramcore.ConversationQuestion{SessionRef: ref, QuestionID: question.QuestionID, TurnID: question.TurnID, Revision: question.Revision, ContentDigest: question.ContentDigest}
}

func telegramAuthoringSessionID(ref string) (string, error) {
	parts := strings.Split(ref, "/")
	if len(parts) != 5 || !filepath.IsLocal(ref) || filepath.Clean(ref) != ref {
		return "", workflowstate.ErrConflict
	}

	if parts[0] != workspacecore.DirectoryName || parts[1] != "authoring" || parts[2] != "sessions" || parts[4] != "session.md" {
		return "", workflowstate.ErrConflict
	}

	return parts[3], nil
}
