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

	"forge.adrianpk.com/chr/internal/core"
)

type telegramIntentService interface {
	ResolveConversationIntent(context.Context, string, ConversationIntentContext) (ConversationIntent, error)
}

func (r telegramReceiver) respondIntent(ctx context.Context, user core.TelegramAllowedUser, config core.TelegramReceiverConfig, update core.TelegramReceivedUpdate) (bool, error) {
	if update.Kind != "message" || strings.HasPrefix(strings.TrimSpace(update.Text), "/") {
		return false, nil
	}

	operations, ok := r.operations.(telegramIntentService)
	if !ok {
		return true, r.recordDiagnostic(ctx, update, config, "Conversational interpretation is unavailable. Explicit commands remain available through /help.")
	}

	input := core.TelegramConversationInput{Update: update, Config: config}

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
			return true, r.recordDiagnostic(ctx, update, config, "Intent resolution exhausted five retained attempts. Prior conversations are preserved. Send an explicit command or a new corrected instruction; this source will not restart its retry budget.")
		}

		return true, nil
	}

	return true, r.applyIntentDecision(ctx, update, resolution.Decision)
}

func (r telegramReceiver) resolveIntent(ctx context.Context, operations telegramIntentService, input core.TelegramConversationInput) (core.TelegramInterpretation, error) {
	update := input.Update

	return r.store.ResolveTelegramInterpretation(ctx, r.env, update, time.Now(), func(callCtx context.Context) (json.RawMessage, error) {
		modelInput, err := r.intentContext(callCtx, input)
		if err != nil {
			return nil, err
		}

		hash := sha256.Sum256(fmt.Appendf(nil, "%d:%d", update.Peer.BotID, update.UpdateID))

		result, err := operations.ResolveConversationIntent(callCtx, "intent-"+hex.EncodeToString(hash[:16]), modelInput)
		if errors.Is(err, errConversationIntentInvalid) || errors.Is(err, core.ErrStructuredModelOutput) {
			result = telegramIntentClarification("The model returned an unsupported or malformed decision. Which exact operation and target did you intend?")
			err = nil
		}

		if err != nil {
			return nil, err
		}

		return r.freezeIntent(callCtx, input, result)
	})
}

func (r telegramReceiver) applyIntentDecision(ctx context.Context, update core.TelegramReceivedUpdate, payload json.RawMessage) error {
	input := core.TelegramConversationInput{Update: update}

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
		if err := r.store.ValidateTelegramOperationTarget(ctx, input, *decision.Target); err != nil {
			return r.intentCorrection(ctx, update, current)
		}
	}

	return r.dispatchIntent(ctx, currentUser, input, decision.Intent)
}

func (r telegramReceiver) intentContext(ctx context.Context, input core.TelegramConversationInput) (ConversationIntentContext, error) {
	result := ConversationIntentContext{Message: input.Update.Text, Operations: telegramIntentRegistry()}

	selected, focused, sessions, err := r.store.TelegramIntentContext(ctx, input)
	if err != nil {
		return result, err
	}

	var allowedIDs []string

	user, _ := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType)
	for _, project := range input.Projects {
		if project.Enabled && slices.Contains(user.ProjectIDs, project.ID) {
			result.Projects = append(result.Projects, ConversationIntentProject{ID: project.ID, Name: project.Name})
			allowedIDs = append(allowedIDs, project.ID)
		}
	}

	if slices.Contains(allowedIDs, selected) {
		result.SelectedProject = selected
	}

	for _, session := range sessions {
		if !slices.Contains(allowedIDs, session.ProjectID) {
			continue
		}

		result.Sessions = append(result.Sessions, ConversationIntentSession{ID: session.ID, ProjectID: session.ProjectID, Namespace: session.Namespace, Intent: session.Intent, Phase: session.Phase, Topic: session.Topic, QuestionBinding: session.QuestionBinding, Revision: session.Revision})
		if session.ID == focused {
			result.FocusedConversation = focused
		}
	}

	return result, nil
}

func (r telegramReceiver) intentCorrection(ctx context.Context, update core.TelegramReceivedUpdate, config core.TelegramReceiverConfig) error {
	return r.recordDiagnostic(ctx, update, config, "The interpretation has unsupported or incomplete arguments. Name the exact project/session and requested operation, or use /help for deterministic controls.")
}

func (r telegramReceiver) dispatchIntent(ctx context.Context, user core.TelegramAllowedUser, input core.TelegramConversationInput, intent ConversationIntent) error {
	if intent.OperationID == "clarify" {
		return r.recordDiagnostic(ctx, input.Update, input.Config, intent.Explanation)
	}

	if handled, err := r.dispatchIntentPreparation(ctx, user, input, intent); handled || err != nil {
		return err
	}

	conversation, recognized, err := telegramIntentConversation(input.Update, intent)
	if recognized {
		if err != nil {
			return r.intentCorrection(ctx, input.Update, input.Config)
		}

		conversation.Config, conversation.Projects, conversation.Namespace = input.Config, input.Projects, input.Namespace
		if handled, err := r.dispatchAuthoringConversation(ctx, user, conversation); handled || err != nil {
			return err
		}

		return r.store.ApplyTelegramConversation(ctx, conversation, time.Now())
	}

	record, recognized, err := telegramIntentRecord(input.Update, intent)
	if recognized {
		if err != nil {
			return r.intentCorrection(ctx, input.Update, input.Config)
		}

		record.Input.Config, record.Input.Projects = input.Config, input.Projects

		return r.dispatchRecord(ctx, user, record)
	}

	return r.dispatchIntentRead(ctx, user, input, intent)
}

func (r telegramReceiver) dispatchIntentRead(ctx context.Context, user core.TelegramAllowedUser, input core.TelegramConversationInput, intent ConversationIntent) error {
	if intent.Arguments != (ConversationIntentArguments{}) {
		return r.intentCorrection(ctx, input.Update, input.Config)
	}

	for _, operation := range telegramReadRegistry() {
		if intent.OperationID != strings.TrimPrefix(operation.Command, "/") {
			continue
		}

		text, err := telegramReadOperationResponse(r.operations, user, input.Update, operation.ID)
		if err != nil {
			return err
		}

		return r.recordDiagnostic(ctx, input.Update, input.Config, text)
	}

	return r.intentCorrection(ctx, input.Update, input.Config)
}

func (r telegramReceiver) dispatchIntentPreparation(ctx context.Context, user core.TelegramAllowedUser, input core.TelegramConversationInput, intent ConversationIntent) (bool, error) {
	work, recognized, err := telegramIntentPreparedWork(input.Update, intent)
	if recognized {
		if err != nil {
			return true, r.intentCorrection(ctx, input.Update, input.Config)
		}

		work.Config, work.Projects, work.Namespace = input.Config, input.Projects, input.Namespace

		return true, r.dispatchPreparedWork(ctx, user, work)
	}

	explanation, recognized, err := telegramIntentExplanation(input.Update, intent)
	if recognized {
		if err != nil {
			return true, r.intentCorrection(ctx, input.Update, input.Config)
		}

		explanation.Input.Config, explanation.Input.Projects, explanation.Input.Namespace = input.Config, input.Projects, input.Namespace

		return true, r.dispatchExplanation(ctx, user, explanation)
	}

	authoring, recognized, err := telegramIntentAuthoring(input.Update, intent)
	if recognized {
		if err != nil {
			return true, r.intentCorrection(ctx, input.Update, input.Config)
		}

		authoring.Input.Config, authoring.Input.Projects, authoring.Input.Namespace = input.Config, input.Projects, input.Namespace

		return true, r.dispatchAuthoring(ctx, user, authoring)
	}

	return false, nil
}
