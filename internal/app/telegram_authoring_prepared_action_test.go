package app

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

// TestTelegramAuthoringPreparedOfferReplay keeps a consumed prompt's token,
// expiry and prepared conversation projection unchanged when the offer replays.
func TestTelegramAuthoringPreparedOfferReplay(t *testing.T) {
	f, id, _ := telegramAuthoringPreparedFixture(t, "Develop bounded retries.")
	action, prompt := telegramAuthoringPreparedOffer(t, f, id)
	telegramAuthoringPreparedAnswer(t, f, action, prompt, "text")

	request := telegramAuthoringPreparedLatest(t, f, id)

	input, err := f.receiver.preparedWorkAuthorityInput(request)
	if err != nil {
		t.Fatal(err)
	}

	again, err := f.receiver.store.OfferTelegramAction(t.Context(), core.TelegramActionOffer{Input: input, RequestID: request.ID, Choices: []string{"develop"}, Text: "The retained scope.", ExpiresAt: time.Now().Add(time.Hour), ConversationPhase: "awaiting_confirmation"}, time.Now())
	if err != nil || again.Token != action.Token || !again.ExpiresAt.Equal(action.ExpiresAt) || f.session(t, id).Phase != "prepared" {
		t.Fatal("offer replay regressed consumed state", again, err)
	}
}

func telegramAuthoringPreparedOffer(t *testing.T, f *telegramConversationAppFixture, id string) (core.TelegramAction, telegramConversationNotice) {
	t.Helper()

	prompt := f.send(t, "/prepare-work "+id, 0)
	request := telegramAuthoringPreparedLatest(t, f, id)

	action, err := f.receiver.store.TelegramRequestAction(t.Context(), request)
	if err != nil || request.State != "awaiting_confirmation" || f.session(t, id).Phase != "awaiting_confirmation" || !strings.Contains(prompt.Text, "complete scope attached") {
		t.Fatal("scope decision was not retained", prompt, request, err)
	}

	return action, prompt
}

func telegramAuthoringPreparedAnswer(t *testing.T, f *telegramConversationAppFixture, action core.TelegramAction, prompt telegramConversationNotice, mode string) telegramConversationNotice {
	t.Helper()

	switch mode {
	case "reply":
		return f.send(t, "develop", prompt.ID)
	case "text":
		return f.send(t, "/answer "+action.Token+" develop", 0)
	default:
		data, err := core.TelegramActionCallbackData(core.TelegramActionAnswer{Token: action.Token, Choice: "develop"})
		if err != nil {
			t.Fatal(err)
		}

		payload, err := json.Marshal([]any{map[string]any{"update_id": f.next, "callback_query": map[string]any{"from": map[string]any{"id": 100001}, "data": data, "message": map[string]any{"message_id": prompt.ID, "date": 1, "from": map[string]any{"id": 12345, "is_bot": true}, "chat": map[string]any{"id": 100001, "type": "private"}, "text": "Untrusted quoted description"}}}})
		if err != nil {
			t.Fatal(err)
		}

		f.next++
		f.updates <- payload

		if _, err := f.receiver.step(t.Context()); err != nil {
			t.Fatal(err)
		}

		return telegramOperationNotice(t, f)
	}
}

// TestTelegramAuthoringPreparedDecision requires a real host-bound scope decision
// for model-proposed development and shares one grant across button/reply/text replay.
func TestTelegramAuthoringPreparedDecision(t *testing.T) {
	for _, mode := range []string{"text", "reply", "button"} {
		t.Run(mode, func(t *testing.T) {
			f, id, _ := telegramAuthoringPreparedFixture(t, "Develop bounded retries.")
			action, prompt := telegramAuthoringPreparedOffer(t, f, id)
			checkTelegramPreparedCaption(t, prompt.Text)

			_, records := telegramAuthoringPreparedStore(t, f)
			if len(records) != 0 {
				t.Fatal("a model proposed development authority")
			}

			notice := telegramAuthoringPreparedAnswer(t, f, action, prompt, mode)
			if !strings.Contains(notice.Text, "Ticket prepared in backlog") {
				t.Fatal("scope decision failed", notice)
			}

			original := telegramAuthoringPreparedCheck(t, f, id, true, 1)
			checkTelegramPreparedResult(t, notice.Text, original.Result.Record.Reference)
			telegramAuthoringPreparedReplay(t, f, id, action, original)
		})
	}
}

func checkTelegramPreparedCaption(t *testing.T, text string) {
	t.Helper()

	if !strings.Contains(text, "Prepare ticket: Develop bounded retries.") {
		t.Fatal("ticket preview omitted its human title", text)
	}

	for _, audit := range []string{"source_request_id:", "title:", "created_at:", "in example-project?"} {
		if strings.Contains(text, audit) {
			t.Fatal("ticket preview exposed source metadata or an opaque project identity", audit, text)
		}
	}
}

func checkTelegramPreparedResult(t *testing.T, text, reference string) {
	t.Helper()

	if !strings.Contains(text, "Develop bounded retries.") || strings.Contains(text, reference) {
		t.Fatal("ticket result omitted its human title or exposed its storage path", text)
	}
}

// TestTelegramAuthoringPreparedExpansion asks once for a changed scope after an
// initial explicit instruction and leaves the first canonical ticket unchanged.
func TestTelegramAuthoringPreparedExpansion(t *testing.T) {
	f, id, _ := telegramAuthoringPreparedFixture(t, "/develop retry ceiling")
	f.send(t, "/prepare-work "+id, 0)

	first := telegramAuthoringPreparedCheck(t, f, id, true, 1)
	f.send(t, "/refine dump "+id+" Include a clarified retry requirement.", 0)

	action, prompt := telegramAuthoringPreparedOffer(t, f, id)

	_, records := telegramAuthoringPreparedStore(t, f)
	if len(records) != 1 {
		t.Fatal("expanded scope silently created native work")
	}

	telegramAuthoringPreparedAnswer(t, f, action, prompt, "text")

	second := telegramAuthoringPreparedCheck(t, f, id, true, 2)
	if second.Scope.Hash == first.Scope.Hash || second.Authorization.Binding == first.Authorization.Binding {
		t.Fatal("expanded scope reused initial authority")
	}
}

// TestTelegramAuthoringPreparedStale rejects a scope prompt replaced by a real
// canonical refinement and requires a new decision for the current description.
func TestTelegramAuthoringPreparedStale(t *testing.T) {
	f, id, _ := telegramAuthoringPreparedFixture(t, "Develop bounded retries.")
	action, prompt := telegramAuthoringPreparedOffer(t, f, id)
	f.send(t, "/refine dump "+id+" Clarify the existing retry requirement.", 0)

	notice := telegramAuthoringPreparedAnswer(t, f, action, prompt, "text")
	if !strings.Contains(notice.Text, "unavailable") {
		t.Fatal("a replaced scope prompt was accepted", notice)
	}

	_, records := telegramAuthoringPreparedStore(t, f)
	if len(records) != 0 {
		t.Fatal("a stale scope created native work")
	}

	next, nextPrompt := telegramAuthoringPreparedOffer(t, f, id)
	telegramAuthoringPreparedAnswer(t, f, next, nextPrompt, "reply")
	telegramAuthoringPreparedCheck(t, f, id, true, 1)
}

func telegramAuthoringPreparedReplay(t *testing.T, f *telegramConversationAppFixture, id string, action core.TelegramAction, original AuthoringWorkResult) {
	t.Helper()

	if f.session(t, id).Phase != "prepared" {
		t.Fatal("scope completion did not project prepared state")
	}

	notice := f.send(t, "/answer "+action.Token+" develop", 0)
	if !strings.Contains(notice.Text, "already confirmed") {
		t.Fatal("decision replay was not recognized", notice)
	}

	again := telegramAuthoringPreparedCheck(t, f, id, true, 1)
	if *again.Authorization != *original.Authorization {
		t.Fatal("replay issued another grant")
	}
}
