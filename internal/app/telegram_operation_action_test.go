package app

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

func telegramOperationOfferRecord(t *testing.T, f *telegramConversationAppFixture) (telegramcore.TelegramAction, telegramConversationNotice) {
	t.Helper()

	mutation := RecordMutationRequest{Kind: RecordTicket, Action: RecordCreate, Title: telegramRecordText("One confirmed task"), Synopsis: telegramRecordText("Exactly one result")}

	return telegramOperationOfferMutation(t, f, mutation)
}

func telegramOperationOfferMutation(t *testing.T, f *telegramConversationAppFixture, mutation RecordMutationRequest) (telegramcore.TelegramAction, telegramConversationNotice) {
	t.Helper()

	f.send(t, "/prepare record scope", 0)

	operations := telegramOperationConcreteService(t, f.receiver)
	input := telegramOperationRetainInput(t, f, "Observed record mutation")
	input.Conversation = f.sessions(t)[0].ID

	target, err := f.receiver.store.ResolveTelegramOperationTarget(t.Context(), input, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	arguments, err := json.Marshal(mutation)
	if err != nil {
		t.Fatal(err)
	}

	request, err := f.receiver.store.ReserveTelegramOperation(t.Context(), telegramcore.TelegramOperationRequest{Source: input.Update, Operator: input.Config.AllowedUsers[0].OperatorID, Target: target, OperationID: fmt.Sprintf("records.%s.%s", mutation.Kind, mutation.Action), Arguments: arguments}, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	if _, err := f.receiver.prepareRecord(t.Context(), operations, request); err != nil {
		t.Fatal(err)
	}

	action, err := f.receiver.store.OfferTelegramAction(t.Context(), telegramcore.TelegramActionOffer{Input: input, RequestID: request.ID, Choices: []string{"confirm"}, Text: "Create the displayed pending ticket?", ExpiresAt: time.Now().Add(time.Hour)}, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	if err := f.receiver.drain(t.Context()); err != nil {
		t.Fatal(err)
	}

	return action, telegramOperationNotice(t, f)
}

func telegramOperationRetainInput(t *testing.T, f *telegramConversationAppFixture, text string) telegramcore.TelegramConversationInput {
	t.Helper()

	payload, err := json.Marshal([]any{map[string]any{"update_id": f.next, "message": map[string]any{"message_id": f.next, "from": map[string]any{"id": 100001}, "chat": map[string]any{"id": 100001, "type": "private"}, "text": text}}})
	if err != nil {
		t.Fatal(err)
	}

	f.next++
	f.updates <- payload

	updates, err := f.receiver.api.Poll(t.Context(), f.receiver.token, nil, 1)
	if err != nil || len(updates) != 1 {
		t.Fatal("operation input unavailable", updates, err)
	}

	config, err := f.receiver.currentConfig()
	if err != nil {
		t.Fatal(err)
	}

	update := telegramcore.AuthorizeTelegramReceiverUpdate(config, 12345, updates[0])
	if err := f.receiver.store.RetainBatch(t.Context(), 12345, []telegramcore.TelegramReceivedUpdate{update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	projects, err := f.receiver.operations.ListRegisteredProjects(config.AllowedUsers[0].ProjectIDs)
	if err != nil {
		t.Fatal(err)
	}

	return telegramcore.TelegramConversationInput{Update: update, Config: config, Projects: projects}
}

func queueTelegramOperationCallback(t *testing.T, f *telegramConversationAppFixture, action telegramcore.TelegramAction) {
	t.Helper()

	data, err := telegramcore.TelegramActionCallbackData(telegramcore.TelegramActionAnswer{Token: action.Token, Choice: "confirm"})
	if err != nil {
		t.Fatal(err)
	}

	payload, err := json.Marshal([]any{map[string]any{"update_id": f.next, "callback_query": map[string]any{"from": map[string]any{"id": 100001}, "data": data, "message": map[string]any{"message_id": 100, "date": 1, "from": map[string]any{"id": 12345, "is_bot": true}, "chat": map[string]any{"id": 100001, "type": "private"}, "text": "quoted prompt"}}}})
	if err != nil {
		t.Fatal(err)
	}

	f.next++
	f.updates <- payload
}

// TestTelegramOperationActionJourney gives reply, callback and explicit text answers one confirmed domain result.
func TestTelegramOperationActionJourney(t *testing.T) {
	f := telegramOperationAppFixture(t)
	action, prompt := telegramOperationOfferRecord(t, f)

	reply := f.send(t, "confirm", prompt.ID)
	if !strings.Contains(reply.Text, "Record saved") {
		t.Fatal("exact action reply failed", reply)
	}

	queueTelegramOperationCallback(t, f, action)

	if _, err := f.receiver.step(t.Context()); err != nil {
		t.Fatal(err)
	}

	callback := telegramOperationNotice(t, f)
	if !strings.Contains(callback.Text, "already confirmed") {
		t.Fatal("duplicate callback reapplied action", callback)
	}

	explicit := f.send(t, "/answer "+action.Token+" confirm", 0)
	if !strings.Contains(explicit.Text, "already confirmed") {
		t.Fatal("explicit answer replaced consumed binding", explicit)
	}

	checkTelegramOperationOneRecord(t, f)
}

func checkTelegramOperationOneRecord(t *testing.T, f *telegramConversationAppFixture) {
	t.Helper()

	page := telegramOperationReadPage(t, f, `/tickets list`)
	if len(page.Records) != 1 || page.Records[0].Identity != "PRJ-1" {
		t.Fatal("action duplicated record", page)
	}

	checkTelegramNoWork(t, f.receiver.env)
}

func telegramOperationNotice(t *testing.T, f *telegramConversationAppFixture) telegramConversationNotice {
	t.Helper()

	select {
	case notice := <-f.notices:
		return notice
	default:
		t.Fatal("expected operation notice was not delivered")
	}

	return telegramConversationNotice{}
}

// TestTelegramOperationActionChangedRecord supersedes an observed update when another owner changes its source image.
func TestTelegramOperationActionChangedRecord(t *testing.T) {
	f := telegramOperationAppFixture(t)
	f.send(t, `/use-cases create {"title":"Shared case","synopsis":"Observed before"}`, 0)

	page := telegramOperationReadPage(t, f, `/use-cases show {"identity":"shared-case"}`)
	before := page.Records[0]
	action, _ := telegramOperationOfferMutation(t, f, RecordMutationRequest{Kind: RecordUseCase, Action: RecordUpdate, Identity: before.Identity, ExpectedHash: before.SourceHash, Synopsis: telegramRecordText("Stale proposed update")})
	operations := telegramOperationConcreteService(t, f.receiver)

	scoped, err := operations.telegramProjectService(t.Context(), telegramcore.TelegramOperationTarget{ProjectID: "example-project", Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	telegramApplyRecord(t, scoped, "independent-case-edit", RecordMutationRequest{Kind: RecordUseCase, Action: RecordUpdate, Identity: before.Identity, ExpectedHash: before.SourceHash, Synopsis: telegramRecordText("Changed independently")})
	queueTelegramOperationCallback(t, f, action)

	if _, err := f.receiver.step(t.Context()); err != nil {
		t.Fatal(err)
	}

	if notice := telegramOperationNotice(t, f); !strings.Contains(notice.Text, "Record conflict") {
		t.Fatal("stale record action accepted", notice)
	}

	if notice := f.send(t, "/answer "+action.Token+" confirm", 0); !strings.Contains(notice.Text, "unavailable") {
		t.Fatal("blocked action was reusable", notice)
	}

	page = telegramOperationReadPage(t, f, `/use-cases show {"identity":"shared-case"}`)
	if !strings.Contains(page.Records[0].Body, "Changed independently") {
		t.Fatal("stale action replaced independent edit", page)
	}
}
