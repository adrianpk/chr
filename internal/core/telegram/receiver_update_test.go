package telegram

import (
	"fmt"
	"strings"
	"testing"
)

// TestTelegramReceiverAuthorization keeps sender, chat, topic and unsupported content outside dispatch.
func TestTelegramReceiverAuthorization(t *testing.T) {
	config := telegramReceiverTestConfig()
	config.AllowedChatIDs = []int64{-100}
	config.AllowedTopicIDs = []int64{7}

	cases := []struct {
		name, sender, chat, extra, text string
		allowed                         bool
	}{
		{"private", `{"id":100001}`, `{"id":100001,"type":"private"}`, "", "/help", true},
		{"group topic", `{"id":100001}`, `{"id":-100,"type":"supergroup"}`, `,"message_thread_id":7`, "/help", true},
		{"wrong user", `{"id":100002}`, `{"id":100002,"type":"private"}`, "", "private denied body", false},
		{"wrong group", `{"id":100001}`, `{"id":-101,"type":"supergroup"}`, "", "/projects", false},
		{"wrong topic", `{"id":100001}`, `{"id":-100,"type":"supergroup"}`, `,"message_thread_id":8`, "/help", false},
		{"bot sender", `{"id":100001,"is_bot":true}`, `{"id":100001,"type":"private"}`, "", "/help", false},
		{"anonymous", `{}`, `{"id":-100,"type":"supergroup"}`, `,"sender_chat":{"id":-100}`, "/help", false},
		{"channel", `{"id":100001}`, `{"id":-100,"type":"channel"}`, "", "/help", false},
		{"forwarded", `{"id":100001}`, `{"id":100001,"type":"private"}`, `,"forward_origin":{"type":"user"}`, "/help", false},
		{"oversized", `{"id":100001}`, `{"id":100001,"type":"private"}`, "", strings.Repeat("x", 16*1024+1), false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			data := fmt.Sprintf(`{"update_id":1,"message":{"message_id":1,"from":%s,"chat":%s,"text":%q%s}}`, test.sender, test.chat, test.text, test.extra)

			update, err := decodeReceiverUpdate([]byte(data))
			if err != nil {
				t.Fatal(err)
			}

			update = AuthorizeReceiverUpdate(config, 12345, update)
			if (update.Kind == "message") != test.allowed || (!test.allowed && update.Text != "") {
				t.Fatal("authorization or denied payload retention failed", update.Kind)
			}
		})
	}
}

// TestTelegramReceiverCallback authenticates the actual clicker and ignores non-message update kinds.
func TestTelegramReceiverCallback(t *testing.T) {
	data := []byte(`{"update_id":1,"callback_query":{"id":"example","from":{"id":100001},"data":"opaque:help","message":{"message_id":7,"date":1,"from":{"id":12345,"is_bot":true},"chat":{"id":100001,"type":"private"},"text":"old bot message"}}}`)

	update, err := decodeReceiverUpdate(data)
	if err != nil || update.Kind != "callback_query" || update.Peer.UserID != 100001 || update.Text != "opaque:help" {
		t.Fatal("callback sender was not retained", update, err)
	}

	for _, name := range []string{"edited_message", "channel_post", "inline_query"} {
		update, err := decodeReceiverUpdate([]byte(fmt.Sprintf(`{"update_id":2,%q:{"text":"/work"}}`, name)))
		if err != nil || update.Kind != "ignored" || update.Text != "" {
			t.Fatal("unsupported update entered interpretation", name, err)
		}
	}
}

// TestTelegramReceiverOutputBounds protects both Unicode limits for read-only replies.
func TestTelegramReceiverOutputBounds(t *testing.T) {
	cases := map[string]string{
		"ascii":         strings.Repeat("a", 20000),
		"supplementary": strings.Repeat("😀", 3900),
		"short":         "help",
		"invalid UTF-8": string([]byte{0xff}),
	}
	for name, text := range cases {
		t.Run(name, func(t *testing.T) {
			if got := BoundReceiverText(text); !validReceiverText(got) {
				t.Fatal("output exceeded Telegram text bounds")
			}
		})
	}
}
