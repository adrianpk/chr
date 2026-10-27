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
	Schema   string                        `json:"schema"`
	Operator string                        `json:"operator"`
	Intent   ConversationIntent            `json:"intent"`
	Target   *telegramcore.OperationTarget `json:"target,omitempty"`
}

func telegramIntentClarification(text string) ConversationIntent {
	return ConversationIntent{OperationID: "clarify", Explanation: text, Unresolved: []string{"operation_or_target"}}
}

func (r telegramReceiver) freezeIntent(ctx context.Context, input telegramcore.ConversationInput, intent ConversationIntent) (json.RawMessage, error) {
	user, _ := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType)
	decision := telegramIntentDecision{Schema: telegramIntentDecisionSchema, Intent: intent, Operator: user.OperatorID}

	targetResolution, err := r.intentTarget(ctx, input, intent)

	target, needed := targetResolution.Target, targetResolution.Needed
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

func (r telegramReceiver) intentTarget(ctx context.Context, input telegramcore.ConversationInput, intent ConversationIntent) (telegramTargetResolution, error) {
	if targetResolution2, err := r.intentPreparationTarget(ctx, input, intent); targetResolution2.Needed || err != nil {
		return telegramTargetResolution{Target: targetResolution2.Target, Needed: targetResolution2.Needed}, err
	}

	parsedInput, err := telegramIntentConversation(input.Update, intent)

	conversation, recognized := parsedInput.Input, parsedInput.Recognized
	if recognized {
		if err != nil {
			return telegramTargetResolution{Target: telegramcore.OperationTarget{}, Needed: false}, err
		}

		conversation.Config, conversation.Projects = input.Config, input.Projects

		conversation.ContextDigest = input.ContextDigest
		switch conversation.Action {
		case telegramcore.ConversationList:
			return telegramTargetResolution{Target: telegramcore.OperationTarget{}, Needed: false}, nil
		case telegramcore.ConversationStart, telegramcore.ConversationSelect:
			target, err := r.store.ResolveOperationTarget(ctx, conversation, time.Now())

			return telegramTargetResolution{Target: target, Needed: true}, err
		default:
			target, err := r.store.ResolveConversationTarget(ctx, conversation)

			return telegramTargetResolution{Target: target, Needed: true}, err
		}
	}

	parsedRecord, err := telegramIntentRecord(input.Update, intent)

	record, recognized := parsedRecord.Command, parsedRecord.Recognized
	if recognized {
		if err != nil {
			return telegramTargetResolution{Target: telegramcore.OperationTarget{}, Needed: false}, err
		}

		record.Input.Config, record.Input.Projects = input.Config, input.Projects
		record.Input.ContextDigest = input.ContextDigest
		target, err := r.store.ResolveOperationTarget(ctx, record.Input, time.Now())

		return telegramTargetResolution{Target: target, Needed: true}, err
	}

	return telegramTargetResolution{Target: telegramcore.OperationTarget{}, Needed: false}, nil
}

func telegramIntentTargetFailure(err error) string {
	if errors.Is(err, telegramcore.ErrRequestCapacity) {
		return telegramRequestCapacityText
	}

	if diagnostic, ok := telegramcore.ConversationFailure(err); ok {
		return diagnostic
	}

	return "Which current project or conversation should I use? Send /sessions to see your conversations; a reply must match its original target."
}

func freezeTelegramIntentArguments(intent ConversationIntent, target telegramcore.OperationTarget) ConversationIntent {
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

func (r telegramReceiver) intentPreparationTarget(ctx context.Context, input telegramcore.ConversationInput, intent ConversationIntent) (telegramTargetResolution, error) {
	if targetResolution3, err := r.intentNativeWorkTarget(ctx, input, intent); targetResolution3.Needed || err != nil {
		return telegramTargetResolution{Target: targetResolution3.Target, Needed: targetResolution3.Needed}, err
	}

	parsedInput2, err := telegramIntentPreparedWork(input.Update, intent)

	work, recognized := parsedInput2.Input, parsedInput2.Recognized
	if recognized {
		if err != nil {
			return telegramTargetResolution{Target: telegramcore.OperationTarget{}, Needed: false}, err
		}

		work.Config, work.Projects = input.Config, input.Projects
		work.ContextDigest = input.ContextDigest
		bindingResult, err := r.store.ResolveAuthoringTarget(ctx, work)
		target := bindingResult.Target

		return telegramTargetResolution{Target: target, Needed: true}, err
	}

	parsedExplanation, err := telegramIntentExplanation(input.Update, intent)

	explanation, recognized := parsedExplanation.Command, parsedExplanation.Recognized
	if recognized {
		if err != nil {
			return telegramTargetResolution{Target: telegramcore.OperationTarget{}, Needed: false}, err
		}

		explanation.Input.Config, explanation.Input.Projects = input.Config, input.Projects
		explanation.Input.ContextDigest = input.ContextDigest
		target, err := r.store.ResolveOperationTarget(ctx, explanation.Input, time.Now())

		return telegramTargetResolution{Target: target, Needed: true}, err
	}

	parsedAuthoring, err := telegramIntentAuthoring(input.Update, intent)

	authoring, recognized := parsedAuthoring.Command, parsedAuthoring.Recognized
	if recognized {
		if err != nil {
			return telegramTargetResolution{Target: telegramcore.OperationTarget{}, Needed: false}, err
		}

		authoring.Input.Config, authoring.Input.Projects = input.Config, input.Projects
		authoring.Input.ContextDigest = input.ContextDigest
		bindingResult2, err := r.store.ResolveAuthoringTarget(ctx, authoring.Input)
		target := bindingResult2.Target

		return telegramTargetResolution{Target: target, Needed: true}, err
	}

	return telegramTargetResolution{Target: telegramcore.OperationTarget{}, Needed: false}, nil
}
