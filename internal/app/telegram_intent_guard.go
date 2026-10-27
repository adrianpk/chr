package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

const telegramIntentDecisionSchema = "chr.telegram.intent.v1"

type telegramIntentDecision struct {
	Schema   string                                `json:"schema"`
	Operator string                                `json:"operator"`
	Intent   ConversationIntent                    `json:"intent"`
	Target   *telegramcore.TelegramOperationTarget `json:"target,omitempty"`
}

func telegramIntentClarification(text string) ConversationIntent {
	return ConversationIntent{OperationID: "clarify", Explanation: text, Unresolved: []string{"operation_or_target"}}
}

func (r telegramReceiver) freezeIntent(ctx context.Context, input telegramcore.TelegramConversationInput, intent ConversationIntent) (json.RawMessage, error) {
	user, _ := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType)
	decision := telegramIntentDecision{Schema: telegramIntentDecisionSchema, Intent: intent, Operator: user.OperatorID}

	target, needed, err := r.intentTarget(ctx, input, intent)
	if err != nil {
		text := telegramIntentTargetFailure(err)
		decision.Intent = telegramIntentClarification(text)
	} else if needed {
		decision.Target = &target
		decision.Intent = freezeTelegramIntentArguments(intent, target)
		// Namespace is already frozen by the guard. Conversation controls do not
		// accept an independently selectable namespace.
	}

	return json.Marshal(decision)
}

func (r telegramReceiver) intentTarget(ctx context.Context, input telegramcore.TelegramConversationInput, intent ConversationIntent) (telegramcore.TelegramOperationTarget, bool, error) {
	if target, needed, err := r.intentPreparationTarget(ctx, input, intent); needed || err != nil {
		return target, needed, err
	}

	conversation, recognized, err := telegramIntentConversation(input.Update, intent)
	if recognized {
		if err != nil {
			return telegramcore.TelegramOperationTarget{}, false, err
		}

		conversation.Config, conversation.Projects = input.Config, input.Projects

		conversation.ContextDigest = input.ContextDigest
		switch conversation.Action {
		case telegramcore.TelegramConversationList:
			return telegramcore.TelegramOperationTarget{}, false, nil
		case telegramcore.TelegramConversationStart, telegramcore.TelegramConversationSelect:
			target, err := r.store.ResolveTelegramOperationTarget(ctx, conversation, time.Now())

			return target, true, err
		default:
			target, err := r.store.ResolveTelegramConversationTarget(ctx, conversation)

			return target, true, err
		}
	}

	record, recognized, err := telegramIntentRecord(input.Update, intent)
	if recognized {
		if err != nil {
			return telegramcore.TelegramOperationTarget{}, false, err
		}

		record.Input.Config, record.Input.Projects = input.Config, input.Projects
		record.Input.ContextDigest = input.ContextDigest
		target, err := r.store.ResolveTelegramOperationTarget(ctx, record.Input, time.Now())

		return target, true, err
	}

	return telegramcore.TelegramOperationTarget{}, false, nil
}

func telegramIntentTargetFailure(err error) string {
	if errors.Is(err, telegramcore.ErrTelegramRequestCapacity) {
		return telegramRequestCapacityText
	}

	if diagnostic, ok := telegramcore.TelegramConversationFailure(err); ok {
		return diagnostic
	}

	return "Which current project or conversation should I use? Send /sessions to see your conversations; a reply must match its original target."
}

func freezeTelegramIntentArguments(intent ConversationIntent, target telegramcore.TelegramOperationTarget) ConversationIntent {
	switch intent.OperationID {
	case "work", "control":
		intent.Arguments.Project, intent.Arguments.Namespace = target.ProjectID, target.Namespace
	}

	if intent.OperationID == "prepare" || intent.OperationID == "develop" || intent.OperationID == "project" || intent.OperationID == "explain" || intent.OperationID == "prepare-work" {
		intent.Arguments.Project = target.ProjectID
	}

	for _, descriptor := range telegramRecordRegistry() {
		if intent.OperationID == strings.TrimPrefix(descriptor.Name, "/") {
			intent.Arguments.Project = target.ProjectID
			intent.Arguments.Namespace = target.Namespace
		}
	}

	if target.ConversationID != "" {
		intent.Arguments.Conversation = target.ConversationID
	}

	return intent
}

func (r telegramReceiver) intentPreparationTarget(ctx context.Context, input telegramcore.TelegramConversationInput, intent ConversationIntent) (telegramcore.TelegramOperationTarget, bool, error) {
	if target, needed, err := r.intentNativeWorkTarget(ctx, input, intent); needed || err != nil {
		return target, needed, err
	}

	work, recognized, err := telegramIntentPreparedWork(input.Update, intent)
	if recognized {
		if err != nil {
			return telegramcore.TelegramOperationTarget{}, false, err
		}

		work.Config, work.Projects = input.Config, input.Projects
		work.ContextDigest = input.ContextDigest
		_, target, err := r.store.ResolveTelegramAuthoringTarget(ctx, work)

		return target, true, err
	}

	explanation, recognized, err := telegramIntentExplanation(input.Update, intent)
	if recognized {
		if err != nil {
			return telegramcore.TelegramOperationTarget{}, false, err
		}

		explanation.Input.Config, explanation.Input.Projects = input.Config, input.Projects
		explanation.Input.ContextDigest = input.ContextDigest
		target, err := r.store.ResolveTelegramOperationTarget(ctx, explanation.Input, time.Now())

		return target, true, err
	}

	authoring, recognized, err := telegramIntentAuthoring(input.Update, intent)
	if recognized {
		if err != nil {
			return telegramcore.TelegramOperationTarget{}, false, err
		}

		authoring.Input.Config, authoring.Input.Projects = input.Config, input.Projects
		authoring.Input.ContextDigest = input.ContextDigest
		_, target, err := r.store.ResolveTelegramAuthoringTarget(ctx, authoring.Input)

		return target, true, err
	}

	return telegramcore.TelegramOperationTarget{}, false, nil
}
