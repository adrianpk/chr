package app

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

type telegramAuthoringArguments struct {
	PayloadHash string `json:"payload_hash"`
}

func (r telegramReceiver) reserveAuthoring(ctx context.Context, operations telegramAuthoringService, user telegramcore.TelegramAllowedUser, command telegramAuthoringCommand) (telegramcore.TelegramOperationRequest, error) {
	c, target, err := r.store.ResolveTelegramAuthoringTarget(ctx, command.Input)
	if err != nil {
		return telegramcore.TelegramOperationRequest{}, err
	}

	owner, err := r.authoringOwnerInput(ctx, operations, command, c, target)
	if err != nil {
		return telegramcore.TelegramOperationRequest{}, err
	}

	owner.Binding = workflowstate.MutationBinding{RequestID: telegramcore.TelegramOperationID(command.Input.Update, 0), Operator: user.OperatorID}

	input, err := json.Marshal(owner)
	if err != nil {
		return telegramcore.TelegramOperationRequest{}, err
	}

	arguments, err := json.Marshal(telegramAuthoringArguments{PayloadHash: recordDigest(input)})
	if err != nil {
		return telegramcore.TelegramOperationRequest{}, err
	}

	return r.store.ReservePreparedTelegramOperation(ctx, telegramcore.TelegramOperationRequest{Source: command.Input.Update, Operator: user.OperatorID, Target: target, OperationID: "authoring.continue", Arguments: arguments}, input, time.Now())
}

func (r telegramReceiver) authoringOwnerInput(ctx context.Context, operations telegramAuthoringService, command telegramAuthoringCommand, c telegramcore.TelegramConversation, target telegramcore.TelegramOperationTarget) (core.AuthoringContinuationRequest, error) {
	request := core.AuthoringContinuationRequest{ProjectID: target.ProjectID, Namespace: target.Namespace, Action: command.Action, Target: command.Target, Instructions: command.Instructions}
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
	if command.Action == core.AuthoringContinuationAnswer {
		return telegramAuthoringAnswerInput(request, command.Input.Text, c.QuestionBinding, snapshot)
	}

	return request, nil
}

func (r telegramReceiver) initialAuthoringInput(ctx context.Context, command telegramAuthoringCommand, c telegramcore.TelegramConversation, request core.AuthoringContinuationRequest) (core.AuthoringContinuationRequest, error) {
	if command.Action == core.AuthoringContinuationRefine || command.Action == core.AuthoringContinuationAnswer {
		return request, workflowstate.ErrConflict
	}

	input := command.Input
	input.Conversation = c.ID

	instruction, err := r.store.ReadTelegramConversationInstruction(ctx, input)
	if err != nil {
		return request, err
	}

	request.Action, request.Topic, request.Capture = core.AuthoringContinuationCreate, c.Topic, c.Capture

	request.OriginalInstruction = instruction.Text
	if request.Capture == "" {
		request.Capture = c.Topic
	}

	return request, nil
}

func telegramAuthoringAnswerInput(request core.AuthoringContinuationRequest, answer, binding string, snapshot core.AuthoringSnapshot) (core.AuthoringContinuationRequest, error) {
	if snapshot.Question == nil {
		return request, workflowstate.ErrConflict
	}

	var question telegramcore.TelegramConversationQuestion
	if err := decodeOperationJSON(binding, &question); err != nil {
		return request, err
	}

	expected := telegramAuthoringQuestion(snapshot.SessionRef, snapshot.Question)
	if question != *expected {
		return request, workflowstate.ErrConflict
	}

	request.Target = core.AuthoringArtifactKind(snapshot.Session.Target)

	request.Question, request.Instructions = snapshot.Question, answer
	if request.Target != core.AuthoringArtifactKindQuiz {
		request.Action = core.AuthoringContinuationRespond
	}

	return request, nil
}

func telegramAuthoringQuestion(ref string, question *core.AuthoringQuestionBinding) *telegramcore.TelegramConversationQuestion {
	if question == nil {
		return nil
	}

	return &telegramcore.TelegramConversationQuestion{SessionRef: ref, QuestionID: question.QuestionID, TurnID: question.TurnID, Revision: question.Revision, ContentDigest: question.ContentDigest}
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
