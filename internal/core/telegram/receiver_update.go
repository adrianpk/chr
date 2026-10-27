package telegram

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

// PeerIdentity is the complete transport peer, never a local authority grant.
type PeerIdentity struct {
	BotID   int64
	ChatID  int64
	TopicID int64
	UserID  int64
}

// ReceivedUpdate is normalized input; denied receipts retain no payload.
type ReceivedUpdate struct {
	UpdateID         int64
	Digest           string
	Kind             string
	Peer             PeerIdentity
	MessageID        int64
	ReplyToMessageID int64
	Text             string
	ChatType         string
	Reason           string
}

type receiverMessage struct {
	MessageID int64        `json:"message_id"`
	ThreadID  int64        `json:"message_thread_id"`
	Date      int64        `json:"date"`
	From      receiverUser `json:"from"`
	Chat      struct {
		ID   int64  `json:"id"`
		Type string `json:"type"`
	} `json:"chat"`
	ReplyTo *struct {
		MessageID int64 `json:"message_id"`
	} `json:"reply_to_message"`
	Text             string          `json:"text"`
	SenderChat       json.RawMessage `json:"sender_chat"`
	ForwardOrigin    json.RawMessage `json:"forward_origin"`
	ForwardFrom      json.RawMessage `json:"forward_from"`
	AutomaticForward bool            `json:"is_automatic_forward"`
	ViaBot           json.RawMessage `json:"via_bot"`
}

type receiverUser struct {
	ID    int64 `json:"id"`
	IsBot bool  `json:"is_bot"`
}

type receiverCallback struct {
	From    receiverUser     `json:"from"`
	Message *receiverMessage `json:"message"`
	Data    string           `json:"data"`
}

func decodeReceiverUpdates(data []byte) ([]ReceivedUpdate, error) {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil || len(raw) > 100 {
		return nil, aPIInvalid()
	}

	updates := make([]ReceivedUpdate, 0, len(raw))
	for _, item := range raw {
		update, err := decodeReceiverUpdate(item)
		if err != nil {
			return nil, err
		}

		updates = append(updates, update)
	}

	return updates, nil
}

func decodeReceiverUpdate(data []byte) (ReceivedUpdate, error) {
	var wire struct {
		ID       *int64            `json:"update_id"`
		Message  *receiverMessage  `json:"message"`
		Callback *receiverCallback `json:"callback_query"`
	}
	if err := json.Unmarshal(data, &wire); err != nil || !validUpdateID(wire.ID) {
		return ReceivedUpdate{}, aPIInvalid()
	}

	update := ReceivedUpdate{UpdateID: *wire.ID, Digest: updateDigest(data), Kind: "ignored", Reason: "unsupported_update"}
	if wire.Callback != nil && wire.Message == nil {
		return normalizeReceiverCallback(update, *wire.Callback), nil
	}

	if wire.Message != nil && wire.Callback == nil {
		return normalizeReceiverMessage(update, *wire.Message, "message"), nil
	}

	return update, nil
}

func validUpdateID(id *int64) bool {
	return id != nil && *id >= 0 && *id != math.MaxInt64
}

func normalizeReceiverCallback(update ReceivedUpdate, callback receiverCallback) ReceivedUpdate {
	if callback.Message == nil || callback.Message.Date == 0 || len(callback.Data) > 64 {
		return update
	}

	message := *callback.Message
	message.From = callback.From // Authenticate the clicker, never the bot's original message author.
	message.Text = callback.Data

	return normalizeReceiverMessage(update, message, "callback_query")
}

func normalizeReceiverMessage(update ReceivedUpdate, message receiverMessage, kind string) ReceivedUpdate {
	update.Kind = "rejected"

	update.Reason = "unsupported_identity_or_content"
	if !validMessageSender(message) || forwardedMessage(message) {
		return update
	}

	if message.Text == "" || len(message.Text) > 16*1024 || !utf8.ValidString(message.Text) {
		return update
	}

	if message.ReplyTo != nil {
		update.ReplyToMessageID = message.ReplyTo.MessageID
	}

	update.Kind, update.Reason = kind, ""
	update.Peer = PeerIdentity{ChatID: message.Chat.ID, TopicID: message.ThreadID, UserID: message.From.ID}
	update.MessageID, update.Text, update.ChatType = message.MessageID, message.Text, message.Chat.Type

	return update
}

func validMessageSender(message receiverMessage) bool {
	return message.From.ID > 0 && !message.From.IsBot && message.MessageID > 0 && message.ThreadID >= 0 && len(message.SenderChat) == 0
}

func forwardedMessage(message receiverMessage) bool {
	return len(message.ForwardOrigin) > 0 || len(message.ForwardFrom) > 0 || message.AutomaticForward || len(message.ViaBot) > 0
}

func updateDigest(data []byte) string {
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

// AuthorizeReceiverUpdate discards denied bodies before persistence or dispatch.
func AuthorizeReceiverUpdate(config ReceiverConfig, botID int64, update ReceivedUpdate) ReceivedUpdate {
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

func validReceiverText(text string) bool {
	return text != "" && utf8.ValidString(text) && utf8.RuneCountInString(text) <= 3900 && len(utf16.Encode([]rune(text))) <= 3900
}

// BoundReceiverText truncates read-only output at both Unicode boundaries.
func BoundReceiverText(text string) string {
	text = strings.ToValidUTF8(text, "�")
	if validReceiverText(text) {
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

var errReceiverConflict = errors.New("telegram retained identity conflict")
