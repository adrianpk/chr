package core

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

type telegramInlineButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}
type telegramInlineKeyboard struct {
	Rows [][]telegramInlineButton `json:"inline_keyboard"`
}

// SendAction sends plain text and optional choices for an already retained action.
// Buttons contain only the same opaque binding used by explicit and reply answers.
func (a TelegramReceiverAPI) SendAction(ctx context.Context, token string, peer TelegramPeerIdentity, text, actionToken string, choices []string) (int64, error) {
	return a.sendAction(ctx, token, peer, text, "", actionToken, choices, nil)
}

// SendMarkdown sends one bounded legacy-Markdown help body without retrying.
// The durable outbox owns format fallback and every actual send attempt.
func (a TelegramReceiverAPI) SendMarkdown(ctx context.Context, token string, peer TelegramPeerIdentity, text string) (int64, error) {
	return a.sendAction(ctx, token, peer, text, "Markdown", "", nil, nil)
}

// SendPresentation sends frozen visible text with explicit formatting entities.
func (a TelegramReceiverAPI) SendPresentation(ctx context.Context, token string, peer TelegramPeerIdentity, text string, entities []TelegramTextEntity) (int64, error) {
	return a.sendAction(ctx, token, peer, text, "", "", nil, entities)
}

func (a TelegramReceiverAPI) sendAction(ctx context.Context, token string, peer TelegramPeerIdentity, text, parseMode, actionToken string, choices []string, entities []TelegramTextEntity) (int64, error) {
	if !validTelegramReceiverText(text) || peer.ChatID == 0 || peer.TopicID < 0 {
		return 0, errors.New("invalid telegram delivery")
	}

	keyboard, err := telegramActionKeyboard(actionToken, choices)
	if err != nil {
		return 0, err
	}

	payload := struct {
		ChatID    int64                   `json:"chat_id"`
		TopicID   int64                   `json:"message_thread_id,omitempty"`
		Text      string                  `json:"text"`
		ParseMode string                  `json:"parse_mode,omitempty"`
		Entities  []TelegramTextEntity    `json:"entities,omitempty"`
		Keyboard  *telegramInlineKeyboard `json:"reply_markup,omitempty"`
	}{peer.ChatID, peer.TopicID, text, parseMode, entities, keyboard}

	data, err := a.call(ctx, token, "sendMessage", payload, 10*time.Second)
	if err != nil {
		return 0, err
	}

	var message struct {
		MessageID int64 `json:"message_id"`
	}
	if err := json.Unmarshal(data, &message); err != nil || message.MessageID <= 0 {
		return 0, &TelegramAPIError{Code: "invalid_response", Permanent: true, Uncertain: true}
	}

	return message.MessageID, nil
}

func telegramActionKeyboard(token string, choices []string) (*telegramInlineKeyboard, error) {
	if token == "" && len(choices) == 0 {
		return nil, nil
	}

	if len(choices) == 0 || len(choices) > 8 {
		return nil, ErrTelegramActionUnavailable
	}

	keyboard := &telegramInlineKeyboard{}

	for _, choice := range choices {
		data, err := TelegramActionCallbackData(TelegramActionAnswer{Token: token, Choice: choice})
		if err != nil {
			return nil, err
		}

		keyboard.Rows = append(keyboard.Rows, []telegramInlineButton{{Text: telegramActionLabel(choice), CallbackData: data}})
	}

	return keyboard, nil
}

func telegramActionLabel(choice string) string {
	labels := map[string]string{"start": "Start work", "develop": "Prepare ticket", "confirm": "Confirm change", "cancel": "Cancel", "pause": "Pause work", "resume": "Resume work", "continue": "Continue one step", "reconcile": "Check work state", "approve": "Approve review", "grant": "Grant allowance", "answer_decision": "Confirm answer"}
	if label := labels[choice]; label != "" {
		return label
	}
	return choice
}
