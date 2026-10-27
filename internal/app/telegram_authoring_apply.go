package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (r telegramReceiver) continueTelegramAuthoring(ctx context.Context, request telegramcore.TelegramOperationRequest, config telegramcore.TelegramReceiverConfig) error {
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

func (r telegramReceiver) validateAuthoringAuthority(ctx context.Context, request telegramcore.TelegramOperationRequest, reconcile bool) (telegramcore.TelegramReceiverConfig, error) {
	config, err := r.currentConfig()
	if err != nil {
		return config, err
	}

	user, allowed := config.AuthorizedUser(request.Source.Peer.ChatID, request.Source.Peer.TopicID, request.Source.Peer.UserID, request.Source.ChatType)
	if !allowed || user.OperatorID != request.Operator {
		return config, workflowstate.ErrConflict
	}

	projects, err := r.operations.ListRegisteredProjects(user.ProjectIDs)
	if err != nil {
		return config, err
	}

	input := telegramcore.TelegramConversationInput{Update: request.Source, Config: config, Projects: projects}
	if reconcile {
		_, err = r.store.ReadTelegramConversationBinding(ctx, input, request.Target)
	} else {
		err = r.store.ValidateTelegramOperationTarget(ctx, input, request.Target)
	}

	return config, err
}

func telegramRetainedAuthoringInput(request telegramcore.TelegramOperationRequest) (core.AuthoringContinuationRequest, error) {
	var owner core.AuthoringContinuationRequest
	if err := decodeBoundedOperationJSON(string(request.PreparedInput), &owner, 1048576); err != nil {
		return owner, err
	}

	var arguments telegramAuthoringArguments
	if err := decodeOperationJSON(string(request.Arguments), &arguments); err != nil {
		return owner, err
	}

	if owner.Binding != (workflowstate.MutationBinding{RequestID: request.ID, Operator: request.Operator}) || owner.ProjectID != request.Target.ProjectID || owner.Namespace != request.Target.Namespace || arguments.PayloadHash != recordDigest(request.PreparedInput) {
		return owner, workflowstate.ErrConflict
	}

	return owner, nil
}

func (r telegramReceiver) authoringAttemptFailure(ctx context.Context, request telegramcore.TelegramOperationRequest, config telegramcore.TelegramReceiverConfig, failure error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if errors.Is(failure, core.ErrAuthoringProviderFailure) || errors.Is(failure, core.ErrAuthoringProviderWait) {
		return nil
	}

	if errors.Is(failure, workflowstate.ErrConflict) || errors.Is(failure, core.ErrAuthoringProviderExhausted) {
		return r.finishAuthoringFailure(ctx, request, config, failure)
	}

	return failure
}

func (r telegramReceiver) finishAuthoringFailure(ctx context.Context, request telegramcore.TelegramOperationRequest, config telegramcore.TelegramReceiverConfig, failure error) error {
	text := telegramAuthoringFailure(failure)

	facts, err := json.Marshal(map[string]string{"state": "blocked", "message": text})
	if err != nil {
		return err
	}

	return r.store.FinishTelegramOperation(ctx, request, "blocked", "", facts, text, config.Digest(), time.Now())
}

func telegramAuthoringFailure(err error) string {
	if errors.Is(err, telegramcore.ErrTelegramRequestCapacity) {
		return telegramRequestCapacityText
	}

	if errors.Is(err, core.ErrAuthoringProviderExhausted) {
		return "Authoring exhausted five retained provider attempts. Prior context and artifacts are preserved. Send a new corrected instruction; this request's budget cannot restart."
	}

	if text, ok := telegramcore.TelegramConversationFailure(err); ok {
		return text
	}

	return "Authoring cannot apply this request to the current session, question or artifact image. Prior context is preserved. Use an exact current session, answer its pending question, or explicitly /refine an existing artifact."
}

type telegramAuthoringFacts struct {
	RequestID  string                                     `json:"request_id"`
	SessionRef string                                     `json:"session_ref"`
	Revision   int64                                      `json:"revision"`
	SourceHash string                                     `json:"source_hash"`
	Outputs    []core.AuthoringArtifactRef                `json:"outputs"`
	Question   *telegramcore.TelegramConversationQuestion `json:"question,omitempty"`
}

func (r telegramReceiver) finishAuthoring(ctx context.Context, request telegramcore.TelegramOperationRequest, result core.AuthoringContinuationResult, config telegramcore.TelegramReceiverConfig) error {
	if result.RequestID != request.ID || result.ReceiptRef == "" {
		return workflowstate.ErrConflict
	}

	question := telegramAuthoringQuestion(result.SessionRef, result.Question)

	facts, err := json.Marshal(telegramAuthoringFacts{RequestID: result.RequestID, SessionRef: result.SessionRef, Revision: result.Session.Revision, SourceHash: result.SourceHash, Outputs: result.Session.Outputs, Question: question})
	if err != nil {
		return err
	}

	text := telegramAuthoringResultText(request.Target.ConversationID, result)

	request.Delivery, err = r.authoringDocument(ctx, request, result)
	if err != nil {
		text = "The authoring result is saved. The complete document could not be attached because its source is unavailable or changed. Use local project commands to inspect the saved result."
	} else if request.Delivery != nil {
		text = request.Delivery.Message.InitialText
	}

	progress := telegramcore.TelegramAuthoringProgress{SessionRef: result.SessionRef, Question: question}

	err = r.store.FinishTelegramAuthoring(ctx, request, progress, result.ReceiptRef, facts, text, config.Digest(), time.Now())
	if errors.Is(err, workflowstate.ErrConflict) {
		text += "\n\nThis conversation has moved on; its current selection is preserved."

		if request.Delivery != nil {
			request.Delivery.Message.Family = "23"
			request.Delivery.Message.SecondaryText = "Complete document attached. This conversation has moved on; its current selection is preserved."
		}

		return r.store.FinishTelegramOperation(ctx, request, "confirmed", result.ReceiptRef, facts, text, config.Digest(), time.Now())
	}

	return err
}

func telegramAuthoringResultText(conversation string, result core.AuthoringContinuationResult) string {
	if result.Question != nil {
		return tgmessage.Render(tgmessage.Blocks{Content: result.Prompt, SecondaryText: "Reply to this message to answer. Use /reply " + conversation + " <answer> when you cannot reply directly."}).PlainFallback
	}

	if result.Session.Target == core.AuthoringTargetDump {
		return tgmessage.Render(tgmessage.Blocks{Family: "15", InitialText: "Your request and details are saved as a dump for " + result.Session.Topic + ".", SecondaryText: "Use /quiz " + conversation + " to clarify the requirements."}).PlainFallback
	}

	var text strings.Builder
	fmt.Fprintf(&text, "Prepared authoring for session %s.\nCanonical session: %s\n", conversation, result.SessionRef)

	for _, output := range result.Session.Outputs {
		fmt.Fprintf(&text, "%s: %s\n", output.Kind, output.Path)
	}

	fmt.Fprintf(&text, "Use /quiz, /spec or /plan with this session; /refine explicitly replaces an observed artifact. /prepare-work prepares a native backlog ticket.")

	return text.String()
}
