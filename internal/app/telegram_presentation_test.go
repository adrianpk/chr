package app

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
	"unicode/utf16"

	"forge.adrianpk.com/chr/internal/core"
)

// A real HTTP/SQLite conversation keeps the request, human choices and exact
// reply targets while changing only what the user sees at each capture step.
func TestTelegramPresentationConversationJourney(t *testing.T) {
	f := newTelegramConversationAppFixture(t)
	f.send(t, "/project example-project", 0)

	start := f.send(t, "/prepare revise the README 😀", 0)

	c := f.sessions(t)[0]
	if !strings.Contains(start.Text, "revise the README 😀 in example-project is retained") || !strings.Contains(start.Text, "add more details") || c.Capture != "" {
		t.Fatal("request acknowledgement or optional capture is misleading", start, c)
	}

	literal := "# Requirements\n```go\nx := `a_b`\n```"

	saved := f.send(t, literal, start.ID)
	if !strings.Contains(saved.Text, "Details saved") || strings.Contains(saved.Text, "bytes") || f.session(t, c.ID).Capture != literal {
		t.Fatal("capture changed or exposed routine accounting", saved)
	}

	checkTelegramPresentationChoices(t, f, c.ID)
	checkTelegramPresentationCaptureView(t, f, c.ID, literal)
	checkTelegramNoWork(t, f.receiver.env)
}

func checkTelegramPresentationChoices(t *testing.T, f *telegramConversationAppFixture, conversation string) {
	t.Helper()

	f.send(t, "/develop another topic", 0)

	choices := f.send(t, "/sessions", 0)
	if !strings.Contains(choices.Text, "revise the README 😀 in example-project") || !strings.Contains(choices.Text, "/focus "+conversation) || strings.Contains(choices.Text, "revision=") {
		t.Fatal("target choices lost human context or exact binding", choices)
	}
}

func checkTelegramPresentationCaptureView(t *testing.T, f *telegramConversationAppFixture, conversation, literal string) {
	t.Helper()

	focused := f.send(t, "/focus "+conversation, 0)
	if !strings.Contains(focused.Text, "Your next details will go to revise the README") {
		t.Fatal("focus effect is unclear", focused)
	}

	view := f.send(t, "/dump "+conversation, 0)
	if !strings.Contains(view.Text, literal) || !strings.Contains(view.Text, "State: collecting details") || f.session(t, conversation).Capture != literal {
		t.Fatal("inspection lost literal bytes or changed capture", view)
	}

	if text := f.send(t, "/cancel "+conversation, 0).Text; !strings.Contains(text, "collected details were discarded") || f.session(t, conversation).Capture != "" {
		t.Fatal("cancel did not state and perform capture discard", text)
	}
}

// Every command remains discoverable through bounded explicit help groups;
// reading them cannot change the current conversation or create reply authority.
func TestTelegramPresentationConversationHelp(t *testing.T) {
	f := newTelegramConversationAppFixture(t)
	prompt := f.send(t, "/prepare bounded help", 0)
	before := f.sessions(t)[0]

	basic := f.send(t, "/help", prompt.ID)
	if !strings.Contains(basic.Text, "/help all") || strings.Contains(basic.Text, "/ticket") {
		t.Fatal("initial help did not defer advanced commands", basic)
	}

	index := f.send(t, "/help all", basic.ID)

	var catalogue strings.Builder

	for _, line := range strings.Split(index.Text, "\n") {
		if !strings.HasPrefix(line, "/help ") {
			continue
		}

		reply := f.send(t, line, 0)
		if len(utf16.Encode([]rune(reply.Text))) > 1800 {
			t.Fatal("a help group exceeds the visible bound", line)
		}

		catalogue.WriteString(reply.Text)
	}

	for _, entry := range telegramHelpMetadata() {
		if !strings.Contains(catalogue.String(), entry.Syntax) {
			t.Fatal("command omitted from navigable help", entry.Syntax)
		}
	}

	if after := f.session(t, before.ID); after != before {
		t.Fatal("help changed a pending conversation", after)
	}
}

// Explicit entities and definite-rejection-only fallback use the same retained
// body and original attempt allowance at the real HTTP and SQLite boundaries.
func TestTelegramPresentationFormatHTTP(t *testing.T) {
	for _, test := range []struct {
		name, failure, state    string
		status, sends, attempts int
	}{
		{"formatted", "", "", 0, 1, 0},
		{"format fallback", "Bad Request: can't parse entities", "", 400, 2, 0},
		{"uncertain", "Internal Server Error", "uncertain", 503, 1, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			payloads := make(chan telegramHelpHTTPBody, 3)
			server := telegramHelpHTTPServer(t, payloads, test.status, test.failure)
			receiver, config := telegramAppTestReceiver(t, core.TelegramReceiverAPI{Client: server.Client(), BaseURL: server.URL})
			update := retainTelegramHelpInput(t, receiver, 1, "/dump")

			message := core.TelegramMessageBlocks{InitialText: "Saved 😀", ContentKind: "markdown", Content: "# Title\n```literal```\\_*[", SecondaryText: "Details retained."}
			if err := receiver.store.QueuePresentationResponse(t.Context(), update, message, config.Digest(), time.Now()); err != nil {
				t.Fatal(err)
			}

			if err := receiver.drainDeliveries(t.Context()); err != nil {
				t.Fatal(err)
			}

			checkTelegramPresentationPayloads(t, payloads, test.sends, message)
			assertTelegramHelpDelivery(t, receiver, test.state, test.attempts)
		})
	}
}

func checkTelegramPresentationPayloads(t *testing.T, payloads <-chan telegramHelpHTTPBody, sends int, message core.TelegramMessageBlocks) {
	t.Helper()

	if len(payloads) != sends {
		t.Fatal("wrong number of send attempts", len(payloads), sends)
	}

	for index := range sends {
		payload := <-payloads
		if payload.Text != message.InitialText+"\n\n"+message.Content+"\n\n"+message.SecondaryText || payload.ParseMode != "" {
			t.Fatal("literal content was parsed or changed", payload)
		}

		if index == 0 {
			if len(payload.Entities) != 1 || payload.Entities[0] != (core.TelegramTextEntity{Type: "pre", Offset: 10, Length: len(utf16.Encode([]rune(message.Content)))}) {
				t.Fatal("wire entity does not address the final literal content", payload)
			}
		} else if len(payload.Entities) != 0 {
			t.Fatal("plain fallback still carries formatting", payload)
		}
	}
}

// Canonical questions lead with the actual prompt and retain their original
// reply binding; displayed text alone cannot select or answer another question.
func TestTelegramPresentationConversationQuestion(t *testing.T) {
	f := newTelegramConversationAppFixture(t)
	telegramAuthoringBridgeHTTP(t, f, func(input authoringModelInput) string {
		return telegramAuthoringBridgeGeneration(t, input)
	}, telegramAuthoringBridgeCapture)
	f.send(t, "/prepare retry policy", 0)

	c := f.sessions(t)[0]
	f.send(t, "/done "+c.ID, 0)

	question := f.send(t, "/quiz "+c.ID, 0)
	if !strings.HasPrefix(question.Text, "Which retry ceiling applies?") || !strings.Contains(question.Text, "Reply to this message to answer") {
		t.Fatal("question did not lead with the actual prompt", question)
	}

	f.send(t, "Five attempts.", question.ID)

	snapshot := telegramAuthoringBridgeSnapshot(t, f, c.ID)

	data, err := json.Marshal(snapshot.Events)
	if err != nil || !strings.Contains(string(data), "Five attempts.") {
		t.Fatal("answer did not reach the bound canonical question", string(data), err)
	}
}

// Read-only project and health views omit operational identifiers and provide a
// concrete recovery route when the user's registry has no accessible project.
func TestTelegramPresentationConversationReads(t *testing.T) {
	receiver, config := telegramAppTestReceiver(t, core.TelegramReceiverAPI{})
	for _, test := range []struct {
		name      string
		operation telegramReadOperation
		want      string
	}{
		{"empty projects", telegramReadProjects, "No projects are available to you. Project access must be configured locally."},
		{"health", telegramReadStatus, "Receiver:"},
	} {
		t.Run(test.name, func(t *testing.T) {
			reply, err := telegramReadOperationResponse(receiver.operations, config.AllowedUsers[0], core.TelegramReceivedUpdate{Peer: core.TelegramPeerIdentity{BotID: 12345}}, test.operation)
			if err != nil || !strings.Contains(reply.Text, test.want) || strings.Contains(reply.Text, "Cursor epoch") || strings.Contains(reply.Text, "Bot:") {
				t.Fatal("read view leaked routine identifiers or lost recovery guidance", reply, err)
			}
		})
	}
}
