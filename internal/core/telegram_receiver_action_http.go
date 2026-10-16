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
	if !validTelegramReceiverText(text) || peer.ChatID == 0 || peer.TopicID < 0 {
		return 0, errors.New("invalid telegram delivery")
	}

	keyboard, err := telegramActionKeyboard(actionToken, choices)
	if err != nil {
		return 0, err
	}

	payload := struct {
		ChatID   int64                   `json:"chat_id"`
		TopicID  int64                   `json:"message_thread_id,omitempty"`
		Text     string                  `json:"text"`
		Keyboard *telegramInlineKeyboard `json:"reply_markup,omitempty"`
	}{peer.ChatID, peer.TopicID, text, keyboard}

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

		keyboard.Rows = append(keyboard.Rows, []telegramInlineButton{{Text: choice, CallbackData: data}})
	}

	return keyboard, nil
}
