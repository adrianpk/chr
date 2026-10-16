package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// TelegramPeerIdentity is the complete transport peer, never a local authority grant.
type TelegramPeerIdentity struct {
	BotID   int64
	ChatID  int64
	TopicID int64
	UserID  int64
}

// TelegramReceivedUpdate is normalized input; denied receipts retain no payload.
type TelegramReceivedUpdate struct {
	UpdateID  int64
	Digest    string
	Kind      string
	Peer      TelegramPeerIdentity
	MessageID int64
	Text      string
	ChatType  string
	Reason    string
}

type telegramReceiverMessage struct {
	MessageID int64                `json:"message_id"`
	ThreadID  int64                `json:"message_thread_id"`
	Date      int64                `json:"date"`
	From      telegramReceiverUser `json:"from"`
	Chat      struct {
		ID   int64  `json:"id"`
		Type string `json:"type"`
	} `json:"chat"`
	Text             string          `json:"text"`
	SenderChat       json.RawMessage `json:"sender_chat"`
	ForwardOrigin    json.RawMessage `json:"forward_origin"`
	ForwardFrom      json.RawMessage `json:"forward_from"`
	AutomaticForward bool            `json:"is_automatic_forward"`
	ViaBot           json.RawMessage `json:"via_bot"`
}

type telegramReceiverUser struct {
	ID    int64 `json:"id"`
	IsBot bool  `json:"is_bot"`
}

type telegramReceiverCallback struct {
	From    telegramReceiverUser     `json:"from"`
	Message *telegramReceiverMessage `json:"message"`
	Data    string                   `json:"data"`
}

func decodeTelegramReceiverUpdates(data []byte) ([]TelegramReceivedUpdate, error) {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil || len(raw) > 100 {
		return nil, telegramAPIInvalid()
	}

	updates := make([]TelegramReceivedUpdate, 0, len(raw))
	for _, item := range raw {
		update, err := decodeTelegramReceiverUpdate(item)
		if err != nil {
			return nil, err
		}

		updates = append(updates, update)
	}

	return updates, nil
}

func decodeTelegramReceiverUpdate(data []byte) (TelegramReceivedUpdate, error) {
	var wire struct {
		ID       *int64                    `json:"update_id"`
		Message  *telegramReceiverMessage  `json:"message"`
		Callback *telegramReceiverCallback `json:"callback_query"`
	}
	if err := json.Unmarshal(data, &wire); err != nil || !validTelegramUpdateID(wire.ID) {
		return TelegramReceivedUpdate{}, telegramAPIInvalid()
	}

	update := TelegramReceivedUpdate{UpdateID: *wire.ID, Digest: telegramUpdateDigest(data), Kind: "ignored", Reason: "unsupported_update"}
	if wire.Callback != nil && wire.Message == nil {
		return normalizeTelegramReceiverCallback(update, *wire.Callback), nil
	}

	if wire.Message != nil && wire.Callback == nil {
		return normalizeTelegramReceiverMessage(update, *wire.Message, "message"), nil
	}

	return update, nil
}

func validTelegramUpdateID(id *int64) bool {
	return id != nil && *id >= 0 && *id != math.MaxInt64
}

func normalizeTelegramReceiverCallback(update TelegramReceivedUpdate, callback telegramReceiverCallback) TelegramReceivedUpdate {
	if callback.Message == nil || callback.Message.Date == 0 || len(callback.Data) > 64 {
		return update
	}

	message := *callback.Message
	message.From = callback.From // Authenticate the clicker, never the bot's original message author.
	message.Text = callback.Data

	return normalizeTelegramReceiverMessage(update, message, "callback_query")
}

func normalizeTelegramReceiverMessage(update TelegramReceivedUpdate, message telegramReceiverMessage, kind string) TelegramReceivedUpdate {
	update.Kind = "rejected"

	update.Reason = "unsupported_identity_or_content"
	if !validTelegramMessageSender(message) || forwardedTelegramMessage(message) {
		return update
	}

	if message.Text == "" || len(message.Text) > 16*1024 || !utf8.ValidString(message.Text) {
		return update
	}

	update.Kind, update.Reason = kind, ""
	update.Peer = TelegramPeerIdentity{ChatID: message.Chat.ID, TopicID: message.ThreadID, UserID: message.From.ID}
	update.MessageID, update.Text, update.ChatType = message.MessageID, message.Text, message.Chat.Type

	return update
}

func validTelegramMessageSender(message telegramReceiverMessage) bool {
	return message.From.ID > 0 && !message.From.IsBot && message.MessageID > 0 && message.ThreadID >= 0 && len(message.SenderChat) == 0
}

func forwardedTelegramMessage(message telegramReceiverMessage) bool {
	return len(message.ForwardOrigin) > 0 || len(message.ForwardFrom) > 0 || message.AutomaticForward || len(message.ViaBot) > 0
}

func telegramUpdateDigest(data []byte) string {
	var normalized any
	// UseNumber preserves exact signed-64-bit IDs in the normalized digest.
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	if err := decoder.Decode(&normalized); err != nil {
		panic(err) // The identical payload was decoded successfully above.
	}

	canonical, err := json.Marshal(normalized)
	if err != nil {
		panic(err)
	}

	sum := sha256.Sum256(canonical)

	return hex.EncodeToString(sum[:])
}

// AuthorizeTelegramReceiverUpdate discards denied bodies before persistence or dispatch.
func AuthorizeTelegramReceiverUpdate(config TelegramReceiverConfig, botID int64, update TelegramReceivedUpdate) TelegramReceivedUpdate {
	update.Peer.BotID = botID
	if update.Kind == "ignored" || update.Kind == "rejected" {
		return update
	}

	if _, ok := config.AuthorizedUser(update.Peer.ChatID, update.Peer.TopicID, update.Peer.UserID, update.ChatType); !ok {
		update.Kind, update.Reason = "rejected", "unauthorized_peer"
		update.Text = ""
	}

	return update
}

func validTelegramReceiverText(text string) bool {
	return text != "" && utf8.ValidString(text) && utf8.RuneCountInString(text) <= 3900 && len(utf16.Encode([]rune(text))) <= 3900
}

// BoundTelegramReceiverText truncates read-only output at both Unicode boundaries.
func BoundTelegramReceiverText(text string) string {
	text = strings.ToValidUTF8(text, "�")
	if validTelegramReceiverText(text) {
		return text
	}

	suffix := "\n[truncated; use local project commands for the complete list]"

	units, end := 0, 0
	for index, char := range text {
		units += utf16.RuneLen(char)
		if units > 3800 {
			break
		}

		end = index + utf8.RuneLen(char)
	}

	return text[:end] + suffix
}

var errTelegramReceiverConflict = errors.New("telegram retained identity conflict")
