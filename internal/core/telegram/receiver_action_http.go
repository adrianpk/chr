package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

type inlineButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}
type inlineKeyboard struct {
	Rows [][]inlineButton `json:"inline_keyboard"`
}

// SendAction sends plain text and optional choices for an already retained action.
// Buttons contain only the same opaque binding used by explicit and reply answers.
func (a ReceiverAPI) SendAction(ctx context.Context, token string, send ActionSendRequest) (int64, error) {
	return a.sendAction(ctx, token, messageSend{Peer: send.Peer, Text: send.Text, ParseMode: "", ActionToken: send.ActionToken, Choices: send.Choices, Entities: nil})
}

// SendMarkdown sends one bounded legacy-Markdown help body without retrying.
// The durable outbox owns format fallback and every actual send attempt.
func (a ReceiverAPI) SendMarkdown(ctx context.Context, token string, peer PeerIdentity, text string) (int64, error) {
	return a.sendAction(ctx, token, messageSend{Peer: peer, Text: text, ParseMode: "Markdown", ActionToken: "", Choices: nil, Entities: nil})
}

// SendPresentation sends frozen visible text with explicit formatting entities.
func (a ReceiverAPI) SendPresentation(ctx context.Context, token string, send PresentationSendRequest) (int64, error) {
	return a.sendAction(ctx, token, messageSend{Peer: send.Peer, Text: send.Text, ParseMode: "", ActionToken: "", Choices: nil, Entities: send.Entities})
}

func (a ReceiverAPI) sendAction(ctx context.Context, token string, send messageSend) (int64, error) {
	if !validReceiverText(send.Text) || send.Peer.ChatID == 0 || send.Peer.TopicID < 0 {
		return 0, errors.New("invalid telegram delivery")
	}

	keyboard, err := actionKeyboard(send.ActionToken, send.Choices, send.Labels...)
	if err != nil {
		return 0, err
	}

	payload := struct {
		ChatID    int64              `json:"chat_id"`
		TopicID   int64              `json:"message_thread_id,omitempty"`
		Text      string             `json:"text"`
		ParseMode string             `json:"parse_mode,omitempty"`
		Entities  []tgmessage.Entity `json:"entities,omitempty"`
		Keyboard  *inlineKeyboard    `json:"reply_markup,omitempty"`
	}{send.Peer.ChatID, send.Peer.TopicID, send.Text, send.ParseMode, send.Entities, keyboard}

	data, err := a.call(ctx, token, "sendMessage", payload, 10*time.Second)
	if err != nil {
		return 0, err
	}

	var message struct {
		MessageID int64 `json:"message_id"`
	}
	if err := json.Unmarshal(data, &message); err != nil || message.MessageID <= 0 {
		return 0, &APIError{Code: "invalid_response", Permanent: true, Uncertain: true}
	}

	return message.MessageID, nil
}

func actionKeyboard(token string, choices []string, labels ...[]string) (*inlineKeyboard, error) {
	if token == "" && len(choices) == 0 {
		return nil, nil
	}

	if len(choices) == 0 || len(choices) > 8 {
		return nil, ErrActionUnavailable
	}

	labelValues, err := keyboardLabels(choices, labels)
	if err != nil {
		return nil, err
	}

	keyboard := &inlineKeyboard{}

	for index, choice := range choices {
		data, err := ActionCallbackData(ActionAnswer{Token: token, Choice: choice})
		if err != nil {
			return nil, err
		}

		keyboard.Rows = append(keyboard.Rows, []inlineButton{{Text: labelValues[index], CallbackData: data}})
	}

	return keyboard, nil
}

func actionLabel(choice string) string {
	labels := map[string]string{"finish": "Finish capture", "start": "Start work", "develop": "Prepare ticket", "confirm": "Confirm change", "cancel": "Cancel", "pause": "Pause work", "resume": "Resume work", "continue": "Continue one step", "reconcile": "Check work state", "approve": "Approve review", "grant": "Grant allowance", "answer_decision": "Confirm answer"}
	if label := labels[choice]; label != "" {
		return label
	}

	return choice
}

// SendActionPresentation keeps the retained action keyboard with rich text.
func (a ReceiverAPI) SendActionPresentation(ctx context.Context, token string, send ActionPresentationRequest) (int64, error) {
	return a.sendAction(ctx, token, messageSend{Peer: send.Peer, Text: send.Text, ParseMode: "", ActionToken: send.ActionToken, Choices: send.Choices, Entities: send.Entities, Labels: [][]string{send.Labels}})
}

func keyboardLabels(choices []string, retained [][]string) ([]string, error) {
	if len(retained) > 1 {
		return nil, ErrActionUnavailable
	}

	var labels []string
	if len(retained) == 1 {
		labels = retained[0]
	} else {
		for _, choice := range choices {
			labels = append(labels, actionLabel(choice))
		}
	}

	if len(labels) != len(choices) {
		return nil, ErrActionUnavailable
	}

	for _, label := range labels {
		if label == "" || tgmessage.VisibleUnits(label) > 64 {
			return nil, ErrActionUnavailable
		}
	}

	return labels, nil
}
