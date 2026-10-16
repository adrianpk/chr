package app

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

const telegramIntentDecisionSchema = "chr.telegram.intent.v1"

type telegramIntentDecision struct {
	Schema   string                        `json:"schema"`
	Operator string                        `json:"operator"`
	Intent   ConversationIntent            `json:"intent"`
	Target   *core.TelegramOperationTarget `json:"target,omitempty"`
}

func telegramIntentClarification(text string) ConversationIntent {
	return ConversationIntent{OperationID: "clarify", Explanation: text, Unresolved: []string{"operation_or_target"}}
}

func (r telegramReceiver) freezeIntent(ctx context.Context, input core.TelegramConversationInput, intent ConversationIntent) (json.RawMessage, error) {
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

func (r telegramReceiver) intentTarget(ctx context.Context, input core.TelegramConversationInput, intent ConversationIntent) (core.TelegramOperationTarget, bool, error) {
	if target, needed, err := r.intentPreparationTarget(ctx, input, intent); needed || err != nil {
		return target, needed, err
	}

	conversation, recognized, err := telegramIntentConversation(input.Update, intent)
	if recognized {
		if err != nil {
			return core.TelegramOperationTarget{}, false, err
		}

		conversation.Config, conversation.Projects = input.Config, input.Projects
		switch conversation.Action {
		case core.TelegramConversationList:
			return core.TelegramOperationTarget{}, false, nil
		case core.TelegramConversationStart, core.TelegramConversationSelect:
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
			return core.TelegramOperationTarget{}, false, err
		}

		record.Input.Config, record.Input.Projects = input.Config, input.Projects
		target, err := r.store.ResolveTelegramOperationTarget(ctx, record.Input, time.Now())

		return target, true, err
	}

	return core.TelegramOperationTarget{}, false, nil
}

func telegramIntentTargetFailure(err error) string {
	if diagnostic, ok := core.TelegramConversationFailure(err); ok {
		return diagnostic
	}

	return "The target or arguments are unavailable, ambiguous or conflict with the reply. Which exact current project/session and operation should be used?"
}

func freezeTelegramIntentArguments(intent ConversationIntent, target core.TelegramOperationTarget) ConversationIntent {
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

func (r telegramReceiver) intentPreparationTarget(ctx context.Context, input core.TelegramConversationInput, intent ConversationIntent) (core.TelegramOperationTarget, bool, error) {
	if target, needed, err := r.intentNativeWorkTarget(ctx, input, intent); needed || err != nil {
		return target, needed, err
	}

	work, recognized, err := telegramIntentPreparedWork(input.Update, intent)
	if recognized {
		if err != nil {
			return core.TelegramOperationTarget{}, false, err
		}

		work.Config, work.Projects = input.Config, input.Projects
		_, target, err := r.store.ResolveTelegramAuthoringTarget(ctx, work)

		return target, true, err
	}

	explanation, recognized, err := telegramIntentExplanation(input.Update, intent)
	if recognized {
		if err != nil {
			return core.TelegramOperationTarget{}, false, err
		}

		explanation.Input.Config, explanation.Input.Projects = input.Config, input.Projects
		target, err := r.store.ResolveTelegramOperationTarget(ctx, explanation.Input, time.Now())

		return target, true, err
	}

	authoring, recognized, err := telegramIntentAuthoring(input.Update, intent)
	if recognized {
		if err != nil {
			return core.TelegramOperationTarget{}, false, err
		}

		authoring.Input.Config, authoring.Input.Projects = input.Config, input.Projects
		_, target, err := r.store.ResolveTelegramAuthoringTarget(ctx, authoring.Input)

		return target, true, err
	}

	return core.TelegramOperationTarget{}, false, nil
}
