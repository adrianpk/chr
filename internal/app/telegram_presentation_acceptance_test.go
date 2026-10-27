package app

import (
	"bytes"
	"encoding/json"
	"maps"
	"reflect"
	"strings"
	"testing"
	"unicode/utf16"
	"unicode/utf8"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
)

// Rare navigation and malformed requests traverse HTTP dispatch and SQLite.
// Their correction text cannot allocate work, change capture or offer an action.
func TestTelegramPresentationCoverageRequests(t *testing.T) {
	for _, item := range []struct{ name, command, want string }{
		{"unknown command", "/invent", "I could not use that command"},
		{"unknown help group", "/help missing", "Send /help all to choose a command group"},
		{"malformed record", "/tickets update {invalid}", "No record was changed"},
		{"missing focus", "/focus", "Usage: /focus <conversation>"},
		{"no conversations", "/sessions", "No retained sessions"},
		{"record owner unavailable", "/tickets list", "The record owner is unavailable"},
	} {
		t.Run(item.name, func(t *testing.T) {
			f := newTelegramConversationAppFixture(t)
			before := f.sessions(t)

			notice := f.send(t, item.command, 0)
			if !strings.Contains(notice.Text, item.want) || !utf8.ValidString(notice.Text) || len(utf16.Encode([]rune(notice.Text))) > 1800 {
				t.Fatal("missing bounded correction", notice)
			}

			if !reflect.DeepEqual(before, f.sessions(t)) || len(f.documents) != 0 {
				t.Fatal("diagnostic changed capture or offered an attachment")
			}

			checkTelegramNoWork(t, f.receiver.env)
		})
	}
}

// Local HTTP/model/SQLite/Git certification joins complete artifact uploads,
// natural-language work preview, exact confirmation and canonical integration.
// This fixture does not establish mobile/desktop or user-operated acceptance.
func TestTelegramPresentationJourney(t *testing.T) {
	f, provider := telegramPresentationJourneyFixture(t)
	conversation, key := prepareTelegramJourney(t, f)
	checkTelegramPresentationJourneyArtifacts(t, f, conversation)
	telegramAuthoringReopen(t, f)

	operations := telegramOperationConcreteService(t, f.receiver)
	action, prompt := telegramWorkOffer(t, f, "Work on ticket "+key)
	checkTelegramWorkRunCount(t, operations, 0)
	checkTelegramPresentationJourneyPreview(t, f, prompt, key, action.Token)

	admitted := f.send(t, "start", prompt.ID)
	if !strings.Contains(admitted.Text, "Work admitted:") {
		t.Fatal("confirmation did not distinguish admission", admitted)
	}

	admission := append(json.RawMessage(nil), f.last...)
	run := checkTelegramWorkRunCount(t, operations, 1)
	checkTelegramPresentationJourneyPending(t, operations, run.ID)

	stop := runTelegramJourneyReceiver(t, f.receiver)
	telegramWorkWaitTerminal(t, f)
	stop()
	checkTelegramJourneyReceipt(t, f, operations, provider, run.ID, key)

	finishedMessage := checkTelegramPresentationJourneyDrain(t, f)

	status := f.send(t, "/status "+run.ID, 0)
	if !strings.Contains(status.Text, "Integration confirmed by retained native receipts") || !strings.Contains(status.Text, "confirmed tickets: 1") || strings.Contains(status.Text, "0 unit(s)") || len(utf16.Encode([]rune(status.Text))) > 1800 {
		t.Fatal("final status lost canonical truth or its bound", status)
	}

	checkTelegramJourneyNaturalStatus(t, f, operations, finishedMessage)

	f.updates <- admission

	if _, err := f.receiver.step(t.Context()); err != nil {
		t.Fatal(err)
	}

	checkTelegramWorkRunCount(t, operations, 1)

	if provider.creates != 1 || len(provider.submissions) != 1 {
		t.Fatal("replayed confirmation repeated native publication")
	}
}

func telegramPresentationJourneyFixture(t *testing.T) (*telegramConversationAppFixture, *telegramWorkAutomaticProvider) {
	t.Helper()
	t.Setenv(opsnamespace.UserEnv, "default")

	f, operations, _ := telegramWorkAdmissionFixture(t, "ticket")
	provider := telegramJourneyProvider(t, operations)
	operations.workflowPullRequests = provider
	f.receiver.operations = operations
	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "tickets:\n  key_prefix: JRN\nruntime:\n  validation_command: git show HEAD:result.txt\nwork:\n  state_authority: tickets\nreview:\n  mode: trusted_auto_merge\n")
	telegramAuthoringBridgeHTTP(t, f, func(input authoringModelInput) string { return telegramAuthoringBridgeGeneration(t, input) }, func(input ConversationIntentContext) ConversationIntent {
		if strings.HasPrefix(input.Message, "Work on ticket ") {
			conversation := input.FocusedConversation
			if conversation == "" && len(input.Sessions) == 1 {
				conversation = input.Sessions[0].ID
			}

			return ConversationIntent{OperationID: "work", Arguments: ConversationIntentArguments{Action: "ticket", Text: strings.TrimPrefix(input.Message, "Work on ticket "), Conversation: conversation}}
		}

		if input.Message == "Show me the status of that work." {
			return telegramJourneyStatusIntent(input)
		}

		return telegramAuthoringBridgeCapture(input)
	})

	return f, provider
}

func checkTelegramPresentationJourneyArtifacts(t *testing.T, f *telegramConversationAppFixture, conversation string) {
	t.Helper()

	snapshot := telegramAuthoringBridgeSnapshot(t, f, conversation)
	matched := map[authoringcore.ArtifactKind]bool{}

	for len(f.documents) > 0 {
		upload := <-f.documents
		for _, artifact := range snapshot.Artifacts {
			if bytes.Equal(upload.Payload, []byte(artifact.Body)) {
				matched[artifact.Ref.Kind] = true
			}
		}

		if len(utf16.Encode([]rune(upload.Fields["caption"]))) > 900 || upload.Fields["reply_markup"] != "" {
			t.Fatal("artifact upload exceeded its bound or acquired an action")
		}
	}

	for _, kind := range []authoringcore.ArtifactKind{authoringcore.ArtifactKindDump, authoringcore.ArtifactKindSpec, authoringcore.ArtifactKindPlan} {
		if !matched[kind] {
			t.Fatal("generated artifact was not attached completely", kind)
		}
	}
}

func checkTelegramPresentationJourneyPreview(t *testing.T, f *telegramConversationAppFixture, prompt telegramConversationNotice, key, token string) {
	t.Helper()

	select {
	case upload := <-f.documents:
		if upload.Filename != "confirmation-scope.md" || !bytes.Contains(upload.Payload, []byte(key)) || !bytes.Contains(upload.Payload, []byte("git show HEAD:result.txt")) {
			t.Fatal("work preview lost its complete scope", upload.Filename)
		}

		if len(utf16.Encode([]rune(prompt.Text))) > 900 || !strings.Contains(prompt.Text, "Confirming admits") {
			t.Fatal("preview omitted consequence or exceeded its bound", prompt)
		}

		checkTelegramConfirmationKeyboard(t, upload.Fields["reply_markup"], token)
	default:
		t.Fatal("work preview has no complete confirmation document")
	}
}

func checkTelegramPresentationJourneyPending(t *testing.T, operations service, id string) {
	t.Helper()

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil || status.ConfirmedTickets != 0 || status.DeliveryRef != "" || status.DeliveryHash != "" {
		t.Fatal("admission invented an integration receipt", status, err)
	}
}

// A valid registered Git owner with no tickets returns an empty view rather than
// an unavailable-owner diagnostic or an unnecessary empty attachment.
func TestTelegramPresentationCoverageEmptyRecords(t *testing.T) {
	f := telegramOperationAppFixture(t)

	notice := f.send(t, "/tickets list", 0)
	if !strings.Contains(notice.Text, "No records matched this view") || len(f.documents) != 0 {
		t.Fatal("empty view changed meaning or offered a document", notice)
	}

	checkTelegramNoWork(t, f.receiver.env)
}

func checkTelegramPresentationJourneyDrain(t *testing.T, f *telegramConversationAppFixture) int64 {
	t.Helper()

	if err := f.receiver.drainDeliveries(t.Context()); err != nil {
		t.Fatal(err)
	}

	var finished int64

	for len(f.notices) > 0 {
		notice := <-f.notices
		if !utf8.ValidString(notice.Text) || len(utf16.Encode([]rune(notice.Text))) > 1800 {
			t.Fatal("progress projection exceeded its bound", notice)
		}

		if strings.Contains(notice.Text, "Work has finished.") {
			finished = notice.ID
		}
	}

	return finished
}

// A status proposal needs the acknowledged work reference and its native
// binding; preparation focus alone cannot identify the completed execution.
func telegramJourneyStatusIntent(input ConversationIntentContext) ConversationIntent {
	var wire struct {
		Recent   string                              `json:"last_work_conversation"`
		Replied  string                              `json:"replied_work_conversation"`
		Sessions []struct{ ID, InvocationID string } `json:"sessions"`
	}

	payload, err := json.Marshal(input)
	if err != nil || json.Unmarshal(payload, &wire) != nil {
		return telegramIntentClarification("The current execution context is unavailable.")
	}

	if wire.Replied != "" {
		wire.Recent = wire.Replied
	}

	for _, session := range wire.Sessions {
		if session.ID == wire.Recent && session.InvocationID != "" {
			return ConversationIntent{OperationID: "control", Arguments: ConversationIntentArguments{Action: "status", Conversation: session.ID}}
		}
	}

	return telegramIntentClarification("The recent work message has no visible execution binding.")
}

func checkTelegramJourneyNaturalStatus(t *testing.T, f *telegramConversationAppFixture, operations service, finished int64) {
	t.Helper()

	if finished == 0 {
		t.Fatal("no acknowledged terminal message available for a status reply")
	}

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)
	message := f.send(t, "Show me the status of that work.", 0)

	for _, reply := range []int64{finished, message.ID} {
		if !strings.Contains(message.Text, "confirmed tickets: 1") {
			t.Fatal("natural status lost the completed execution", message.Text)
		}

		message = f.send(t, "Show me the status of that work.", reply)
	}

	if !strings.Contains(message.Text, "Integration confirmed by retained native receipts") || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("status reply failed or changed canonical work", message.Text)
	}
}
