package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func (r telegramReceiver) continueTelegramAuthoring(ctx context.Context, request core.TelegramOperationRequest, config core.TelegramReceiverConfig) error {
	if request.State == "confirmed" || request.State == "rejected" || request.State == "blocked" {
		return nil
	}

	operations, ok := r.operations.(telegramAuthoringService)
	if !ok {
		return r.finishAuthoringFailure(ctx, request, config, ErrRecordUnsupported)
	}

	current, err := r.validateAuthoringAuthority(ctx, request, request.State == "applying")
	if err != nil {
		return r.finishAuthoringFailure(ctx, request, config, err)
	}

	owner, err := telegramRetainedAuthoringInput(request)
	if err != nil {
		return r.finishAuthoringFailure(ctx, request, current, err)
	}

	if err := r.store.BeginTelegramOperation(ctx, request, time.Now()); err != nil {
		return err
	}

	result, err := operations.ContinueTelegramAuthoring(ctx, request.Target, owner, func(ctx context.Context) error {
		_, err := r.validateAuthoringAuthority(ctx, request, false)

		return err
	})
	if err != nil {
		return r.authoringAttemptFailure(ctx, request, current, err)
	}

	return r.finishAuthoring(ctx, request, result, current)
}

func (r telegramReceiver) validateAuthoringAuthority(ctx context.Context, request core.TelegramOperationRequest, reconcile bool) (core.TelegramReceiverConfig, error) {
	config, err := r.currentConfig()
	if err != nil {
		return config, err
	}

	user, allowed := config.AuthorizedUser(request.Source.Peer.ChatID, request.Source.Peer.TopicID, request.Source.Peer.UserID, request.Source.ChatType)
	if !allowed || user.OperatorID != request.Operator {
		return config, core.ErrWorkflowConflict
	}

	projects, err := r.operations.ListRegisteredProjects(user.ProjectIDs)
	if err != nil {
		return config, err
	}

	input := core.TelegramConversationInput{Update: request.Source, Config: config, Projects: projects}
	if reconcile {
		_, err = r.store.ReadTelegramConversationBinding(ctx, input, request.Target)
	} else {
		err = r.store.ValidateTelegramOperationTarget(ctx, input, request.Target)
	}

	return config, err
}

func telegramRetainedAuthoringInput(request core.TelegramOperationRequest) (core.AuthoringContinuationRequest, error) {
	var owner core.AuthoringContinuationRequest
	if err := decodeBoundedOperationJSON(string(request.PreparedInput), &owner, 1048576); err != nil {
		return owner, err
	}

	var arguments telegramAuthoringArguments
	if err := decodeOperationJSON(string(request.Arguments), &arguments); err != nil {
		return owner, err
	}

	if owner.Binding != (core.WorkflowMutationBinding{RequestID: request.ID, Operator: request.Operator}) || owner.ProjectID != request.Target.ProjectID || owner.Namespace != request.Target.Namespace || arguments.PayloadHash != recordDigest(request.PreparedInput) {
		return owner, core.ErrWorkflowConflict
	}

	return owner, nil
}

func (r telegramReceiver) authoringAttemptFailure(ctx context.Context, request core.TelegramOperationRequest, config core.TelegramReceiverConfig, failure error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if errors.Is(failure, core.ErrAuthoringProviderFailure) || errors.Is(failure, core.ErrAuthoringProviderWait) {
		return nil
	}

	if errors.Is(failure, core.ErrWorkflowConflict) || errors.Is(failure, core.ErrAuthoringProviderExhausted) {
		return r.finishAuthoringFailure(ctx, request, config, failure)
	}

	return failure
}

func (r telegramReceiver) finishAuthoringFailure(ctx context.Context, request core.TelegramOperationRequest, config core.TelegramReceiverConfig, failure error) error {
	text := telegramAuthoringFailure(failure)

	facts, err := json.Marshal(map[string]string{"state": "blocked", "message": text})
	if err != nil {
		return err
	}

	return r.store.FinishTelegramOperation(ctx, request, "blocked", "", facts, text, config.Digest(), time.Now())
}

func telegramAuthoringFailure(err error) string {
	if errors.Is(err, core.ErrTelegramRequestCapacity) {
		return telegramRequestCapacityText
	}

	if errors.Is(err, core.ErrAuthoringProviderExhausted) {
		return "Authoring exhausted five retained provider attempts. Prior context and artifacts are preserved. Send a new corrected instruction; this request's budget cannot restart."
	}

	if text, ok := core.TelegramConversationFailure(err); ok {
		return text
	}

	return "Authoring cannot apply this request to the current session, question or artifact image. Prior context is preserved. Use an exact current session, answer its pending question, or explicitly /refine an existing artifact."
}

type telegramAuthoringFacts struct {
	RequestID  string                             `json:"request_id"`
	SessionRef string                             `json:"session_ref"`
	Revision   int64                              `json:"revision"`
	SourceHash string                             `json:"source_hash"`
	Outputs    []core.AuthoringArtifactRef        `json:"outputs"`
	Question   *core.TelegramConversationQuestion `json:"question,omitempty"`
}

func (r telegramReceiver) finishAuthoring(ctx context.Context, request core.TelegramOperationRequest, result core.AuthoringContinuationResult, config core.TelegramReceiverConfig) error {
	if result.RequestID != request.ID || result.ReceiptRef == "" {
		return core.ErrWorkflowConflict
	}

	question := telegramAuthoringQuestion(result.SessionRef, result.Question)

	facts, err := json.Marshal(telegramAuthoringFacts{RequestID: result.RequestID, SessionRef: result.SessionRef, Revision: result.Session.Revision, SourceHash: result.SourceHash, Outputs: result.Session.Outputs, Question: question})
	if err != nil {
		return err
	}

	text := telegramAuthoringResultText(request.Target.ConversationID, result)
	progress := core.TelegramAuthoringProgress{SessionRef: result.SessionRef, Question: question}

	err = r.store.FinishTelegramAuthoring(ctx, request, progress, result.ReceiptRef, facts, text, config.Digest(), time.Now())
	if errors.Is(err, core.ErrWorkflowConflict) {
		text = "Canonical authoring was confirmed, but this conversation has advanced. Its current binding is unchanged.\n" + result.SessionRef

		return r.store.FinishTelegramOperation(ctx, request, "confirmed", result.ReceiptRef, facts, text, config.Digest(), time.Now())
	}

	return err
}

func telegramAuthoringResultText(conversation string, result core.AuthoringContinuationResult) string {
	if result.Question != nil {
		return fmt.Sprintf("Question for session %s:\n\n%s\n\nReply to this message or use /reply %s <answer>.", conversation, result.Prompt, conversation)
	}

	var text strings.Builder
	fmt.Fprintf(&text, "Prepared authoring for session %s.\nCanonical session: %s\n", conversation, result.SessionRef)

	for _, output := range result.Session.Outputs {
		fmt.Fprintf(&text, "%s: %s\n", output.Kind, output.Path)
	}

	fmt.Fprintf(&text, "Use /quiz, /spec or /plan with this session; /refine explicitly replaces an observed artifact. /prepare-work prepares a native backlog ticket.")

	return text.String()
}
