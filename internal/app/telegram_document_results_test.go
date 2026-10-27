package app

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

func telegramLatestDocument(t *testing.T, f *telegramConversationAppFixture, prefix string) telegramDocumentUpload {
	t.Helper()

	var found telegramDocumentUpload

	for len(f.documents) > 0 {
		document := <-f.documents
		if strings.HasPrefix(document.Filename, prefix) {
			found = document
		}
	}

	if found.Filename == "" {
		t.Fatal("primary document unavailable", prefix)
	}

	return found
}

// Real authoring, receiver HTTP and SQLite send only the primary new artifact,
// preserving its complete source image rather than resending older outputs.
func TestTelegramPresentationDocumentAuthoring(t *testing.T) {
	f := newTelegramConversationAppFixture(t)
	calls := telegramAuthoringBridgeHTTP(t, f, func(authoringModelInput) string {
		return telegramAuthoringServiceOutput(t, "# Requirements\n\n"+strings.Repeat("Keep the retry ceiling.\n", 450)+"\n## Validation\n\nVerify recovery.", "")
	}, func(input ConversationIntentContext) ConversationIntent {
		kind := strings.TrimSuffix(strings.TrimPrefix(input.Message, "Show the "), " document.")
		if len(input.Sessions) != 1 {
			t.Error("document read requires one observed conversation")

			return telegramIntentClarification("Which conversation?")
		}

		return ConversationIntent{OperationID: "explain", Arguments: ConversationIntentArguments{Text: input.Message, Action: kind, Conversation: input.Sessions[0].ID}}
	})
	f.send(t, "/prepare readable documents", 0)

	c := f.sessions(t)[0]
	for _, step := range []struct {
		command string
		kind    authoringcore.ArtifactKind
	}{
		{"/done", authoringcore.ArtifactKindDump},
		{"/spec", authoringcore.ArtifactKindSpec},
		{"/refine spec", authoringcore.ArtifactKindSpec},
		{"/plan", authoringcore.ArtifactKindPlan},
	} {
		t.Run(step.command, func(t *testing.T) {
			command := step.command + " " + c.ID
			if strings.HasPrefix(step.command, "/refine") {
				command += " retain the same scope"
			}

			notice := f.send(t, command, 0)
			if !strings.Contains(notice.Text, "Complete document attached") || len(f.documents) != 1 {
				t.Fatal("missing or repeated primary document", notice, len(f.documents))
			}

			document := telegramLatestDocument(t, f, string(step.kind)+"-")
			snapshot := telegramAuthoringBridgeSnapshot(t, f, c.ID)
			assertTelegramArtifactUpload(t, document, snapshot, step.kind)

			generations := calls.Load()

			checkTelegramOriginalDocumentRead(t, f, snapshot, step.kind, c.ID)

			if calls.Load() != generations {
				t.Fatal("document read regenerated content or changed canonical state")
			}
		})
	}

	checkTelegramNoWork(t, f.receiver.env)
}

func checkTelegramOriginalDocumentRead(t *testing.T, f *telegramConversationAppFixture, snapshot authoringcore.AuthoringSnapshot, kind authoringcore.ArtifactKind, conversation string) {
	t.Helper()

	notice := f.send(t, "Show the "+string(kind)+" document.", 0)
	if !strings.Contains(notice.Text, "Complete original document attached") || len(f.documents) != 1 {
		t.Fatal("document read did not attach the original", notice, len(f.documents))
	}

	document := telegramLatestDocument(t, f, string(kind)+".md")
	assertTelegramArtifactUpload(t, document, snapshot, kind)

	if telegramAuthoringBridgeSnapshot(t, f, conversation).SourceHash != snapshot.SourceHash {
		t.Fatal("document read changed canonical state")
	}
}

func assertTelegramArtifactUpload(t *testing.T, document telegramDocumentUpload, snapshot authoringcore.AuthoringSnapshot, kind authoringcore.ArtifactKind) {
	t.Helper()

	for _, artifact := range snapshot.Artifacts {
		if artifact.Ref.Kind == kind && bytes.Equal([]byte(artifact.Body), document.Payload) {
			return
		}
	}

	t.Fatal("attachment differs from its complete canonical artifact", document.Filename)
}

type telegramAdvanceAfterAuthoring struct {
	service
	advance func() error
}

func (s telegramAdvanceAfterAuthoring) ContinueTelegramAuthoring(ctx context.Context, target telegramcore.OperationTarget, request authoringcore.AuthoringContinuationRequest, guard authoringcore.AuthoringCommitGuard) (authoringcore.AuthoringContinuationResult, error) {
	result, err := s.service.ContinueTelegramAuthoring(ctx, target, request, guard)
	if err == nil {
		err = s.advance()
	}

	return result, err
}

// A confirmed owner result remains downloadable when an independent cancellation
// advances the receiver conversation before its local completion projection.
func TestTelegramPresentationDocumentAdvancedConversation(t *testing.T) {
	f := newTelegramConversationAppFixture(t)
	telegramAuthoringBridgeHTTP(t, f, func(authoringModelInput) string {
		return telegramAuthoringServiceOutput(t, "# Requirements\nKeep the retry ceiling.", "")
	}, telegramAuthoringBridgeCapture)
	f.send(t, "/prepare advanced conversation", 0)

	original := f.sessions(t)[0]
	operations := telegramOperationConcreteService(t, f.receiver)
	f.receiver.operations = telegramAdvanceAfterAuthoring{service: operations, advance: func() error {
		input := telegramOperationRetainInput(t, f, "/cancel "+original.ID)
		input.Action, input.Conversation = telegramcore.ConversationCancel, original.ID

		return f.receiver.store.ApplyConversation(t.Context(), input, time.Now())
	}}

	input := telegramOperationRetainInput(t, f, "/done "+original.ID)
	if err := f.receiver.respond(t.Context(), input.Update); err != nil {
		t.Fatal(err)
	}

	if err := f.receiver.drainDeliveries(t.Context()); err != nil {
		t.Fatal(err)
	}

	operationLookup, err := f.receiver.store.Operation(t.Context(), input.Update, 0)

	request, found := operationLookup.Request, operationLookup.Found
	if err != nil || !found || request.State != "confirmed" || request.ReceiptRef == "" {
		t.Fatal("advanced conversation lost original receipt", request, err)
	}

	checkTelegramAdvancedDocument(t, f, original)
}

func checkTelegramAdvancedDocument(t *testing.T, f *telegramConversationAppFixture, original telegramcore.Conversation) {
	t.Helper()

	current := f.session(t, original.ID)
	if current.Phase != "canceled" || current.Capture != "" || current.AuthoringSessionRef != "" || current.ProjectID != original.ProjectID {
		t.Fatal("completed presentation overwrote current selection", current)
	}

	if len(f.documents) != 1 {
		t.Fatal("advanced result lost or duplicated attachment", len(f.documents))
	}

	document := telegramLatestDocument(t, f, "dump-")
	if !strings.Contains(document.Fields["caption"], "current selection is preserved") {
		t.Fatal("advanced result did not explain preserved selection", document.Fields["caption"])
	}
}
