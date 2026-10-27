package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	modelcore "forge.adrianpk.com/chr/internal/core/model"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

type telegramIntentService interface {
	ResolveConversationIntent(context.Context, string, ConversationIntentContext) (ConversationIntent, error)
}

func (r telegramReceiver) respondIntent(ctx context.Context, user telegramcore.AllowedUser, config telegramcore.ReceiverConfig, update telegramcore.ReceivedUpdate) (bool, error) {
	if update.Kind != "message" || strings.HasPrefix(strings.TrimSpace(update.Text), "/") {
		return false, nil
	}

	operations, ok := r.operations.(telegramIntentService)
	if !ok {
		return true, r.recordDiagnostic(ctx, update, config, "Conversational interpretation is unavailable. Explicit commands remain available through /help.")
	}

	input := telegramcore.ConversationInput{Update: update, Config: config}

	projects, err := r.operations.ListRegisteredProjects(user.ProjectIDs)
	if err != nil {
		return true, err
	}

	input.Projects = projects

	resolution, err := r.resolveIntent(ctx, operations, input)
	if err != nil {
		return true, err
	}

	if len(resolution.Decision) == 0 {
		if resolution.Attempts >= 5 {
			return true, r.recordDiagnostic(ctx, update, config, "I could not understand this request after five attempts. Prior conversations are preserved. Send /help for commands, or send a new corrected instruction.")
		}

		return true, nil
	}

	return true, r.applyIntentDecision(ctx, update, resolution.Decision)
}

func (r telegramReceiver) resolveIntent(ctx context.Context, operations telegramIntentService, input telegramcore.ConversationInput) (telegramcore.Interpretation, error) {
	update := input.Update

	return r.store.ResolveInterpretation(ctx, r.env, update, time.Now(), func(callCtx context.Context) (json.RawMessage, error) {
		intentObservation, err := r.intentContext(callCtx, input)
		modelInput, digest := intentObservation.Context, intentObservation.Digest

		if err != nil {
			return nil, err
		}

		hash := sha256.Sum256(fmt.Appendf(nil, "%d:%d", update.Peer.BotID, update.UpdateID))

		result, err := operations.ResolveConversationIntent(callCtx, "intent-"+hex.EncodeToString(hash[:16]), modelInput)
		if errors.Is(err, errConversationIntentInvalid) || errors.Is(err, modelcore.ErrOutput) {
			result = telegramIntentClarification("Which operation should I use, and for which project or conversation?")
			err = nil
		}

		if err != nil {
			return nil, err
		}

		input.ContextDigest = digest

		return r.freezeIntent(callCtx, input, result)
	})
}

func (r telegramReceiver) applyIntentDecision(ctx context.Context, update telegramcore.ReceivedUpdate, payload json.RawMessage) error {
	input := telegramcore.ConversationInput{Update: update}

	var decision telegramIntentDecision
	if err := decodeOperationJSON(string(payload), &decision); err != nil {
		return err
	}

	if decision.Schema != telegramIntentDecisionSchema {
		return errors.New("unsupported retained intent schema")
	}

	current, err := r.currentConfig()
	if err != nil {
		return err
	}

	currentUser, allowed := current.AuthorizedUser(update.Peer.ChatID, update.Peer.TopicID, update.Peer.UserID, update.ChatType)
	if !allowed {
		return r.store.RejectPendingUpdate(ctx, update, time.Now())
	}

	input.Config = current
	if decision.Operator != currentUser.OperatorID {
		return r.intentCorrection(ctx, update, current)
	}

	if err := validateConversationIntent(decision.Intent, telegramIntentRegistry()); err != nil {
		return r.intentCorrection(ctx, update, current)
	}

	input.Projects, err = r.operations.ListRegisteredProjects(currentUser.ProjectIDs)
	if err != nil {
		return err
	}

	if decision.Target != nil {
		input.Namespace = decision.Target.Namespace
		if err := r.store.ValidateOperationTarget(ctx, input, *decision.Target); err != nil {
			return r.intentCorrection(ctx, update, current)
		}
	}

	return r.dispatchIntent(ctx, currentUser, input, decision.Intent)
}

func (r telegramReceiver) intentContext(ctx context.Context, input telegramcore.ConversationInput) (telegramIntentObservation, error) {
	result := ConversationIntentContext{Message: input.Update.Text, Operations: telegramIntentRegistry()}

	snapshot, err := r.store.IntentContext(ctx, input)
	if err != nil {
		return telegramIntentObservation{Context: result, Digest: ""}, err
	}

	var allowedIDs []string

	user, _ := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType)
	for _, project := range input.Projects {
		if project.Enabled && slices.Contains(user.ProjectIDs, project.ID) {
			result.Projects = append(result.Projects, ConversationIntentProject{ID: project.ID, Name: project.Name})
			allowedIDs = append(allowedIDs, project.ID)
		}
	}

	if slices.Contains(allowedIDs, snapshot.Selected) {
		result.SelectedProject = snapshot.Selected
	}

	for _, session := range snapshot.Sessions {
		if !slices.Contains(allowedIDs, session.ProjectID) {
			continue
		}

		appendTelegramIntentSession(&result, session, snapshot)
	}

	return telegramIntentObservation{Context: result, Digest: snapshot.Digest}, nil
}

func appendTelegramIntentSession(result *ConversationIntentContext, session telegramcore.Conversation, snapshot telegramcore.IntentSnapshot) {
	result.Sessions = append(result.Sessions, ConversationIntentSession{ID: session.ID, ProjectID: session.ProjectID, Namespace: session.Namespace, Intent: session.Intent, Phase: string(session.Phase), Topic: session.Topic, QuestionBinding: session.QuestionBinding, InvocationID: session.InvocationID, Revision: session.Revision})
	if session.ID == snapshot.Focused {
		result.FocusedConversation = session.ID
	}

	if session.ID == snapshot.LastWorkConversation {
		result.LastWorkConversation = session.ID
	}

	if session.ID == snapshot.RepliedWorkConversation {
		result.RepliedWorkConversation = session.ID
	}
}

func (r telegramReceiver) intentCorrection(ctx context.Context, update telegramcore.ReceivedUpdate, config telegramcore.ReceiverConfig) error {
	return r.recordDiagnostic(ctx, update, config, "Which current project or conversation should this request use? The prior target or arguments are no longer valid. Send /sessions to see your conversations.")
}

func (r telegramReceiver) dispatchIntent(ctx context.Context, user telegramcore.AllowedUser, input telegramcore.ConversationInput, intent ConversationIntent) error {
	if intent.OperationID == "clarify" {
		return r.recordDiagnostic(ctx, input.Update, input.Config, intent.Explanation)
	}

	if handled, err := r.dispatchIntentPreparation(ctx, user, input, intent); handled || err != nil {
		return err
	}

	parsedInput, err := telegramIntentConversation(input.Update, intent)

	conversation, recognized := parsedInput.Input, parsedInput.Recognized
	if recognized {
		if err != nil {
			return r.intentCorrection(ctx, input.Update, input.Config)
		}

		conversation.Config, conversation.Projects, conversation.Namespace = input.Config, input.Projects, input.Namespace
		if handled, err := r.dispatchAuthoringConversation(ctx, user, conversation); handled || err != nil {
			return err
		}

		return r.store.ApplyConversation(ctx, conversation, time.Now())
	}

	parsedRecord, err := telegramIntentRecord(input.Update, intent)

	record, recognized := parsedRecord.Command, parsedRecord.Recognized
	if recognized {
		if err != nil {
			return r.intentCorrection(ctx, input.Update, input.Config)
		}

		record.Input.Config, record.Input.Projects = input.Config, input.Projects

		return r.dispatchRecord(ctx, user, record)
	}

	return r.dispatchIntentRead(ctx, user, input, intent)
}

func (r telegramReceiver) dispatchIntentRead(ctx context.Context, user telegramcore.AllowedUser, input telegramcore.ConversationInput, intent ConversationIntent) error {
	if intent.Arguments != (ConversationIntentArguments{}) {
		return r.intentCorrection(ctx, input.Update, input.Config)
	}

	for _, operation := range telegramReadRegistry() {
		if intent.OperationID != strings.TrimPrefix(operation.Command, "/") {
			continue
		}

		reply, err := telegramReadOperationResponse(r.operations, user, input.Update, operation.ID)
		if err != nil {
			return err
		}

		return r.queueReadReply(ctx, input.Update, input.Config, reply)
	}

	return r.intentCorrection(ctx, input.Update, input.Config)
}

func (r telegramReceiver) dispatchIntentPreparation(ctx context.Context, user telegramcore.AllowedUser, input telegramcore.ConversationInput, intent ConversationIntent) (bool, error) {
	if handled, err := r.dispatchIntentNativeWork(ctx, user, input, intent); handled || err != nil {
		return handled, err
	}

	parsedInput2, err := telegramIntentPreparedWork(input.Update, intent)

	work, recognized := parsedInput2.Input, parsedInput2.Recognized
	if recognized {
		if err != nil {
			return true, r.intentCorrection(ctx, input.Update, input.Config)
		}

		work.Config, work.Projects, work.Namespace = input.Config, input.Projects, input.Namespace

		return true, r.dispatchPreparedWork(ctx, user, work)
	}

	parsedExplanation, err := telegramIntentExplanation(input.Update, intent)

	explanation, recognized := parsedExplanation.Command, parsedExplanation.Recognized
	if recognized {
		if err != nil {
			return true, r.intentCorrection(ctx, input.Update, input.Config)
		}

		explanation.Input.Config, explanation.Input.Projects, explanation.Input.Namespace = input.Config, input.Projects, input.Namespace

		return true, r.dispatchExplanation(ctx, user, explanation)
	}

	parsedAuthoring, err := telegramIntentAuthoring(input.Update, intent)

	authoring, recognized := parsedAuthoring.Command, parsedAuthoring.Recognized
	if recognized {
		if err != nil {
			return true, r.intentCorrection(ctx, input.Update, input.Config)
		}

		authoring.Input.Config, authoring.Input.Projects, authoring.Input.Namespace = input.Config, input.Projects, input.Namespace

		return true, r.dispatchAuthoring(ctx, user, authoring)
	}

	return false, nil
}
