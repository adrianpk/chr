package app

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Real authoring sources and registry bindings reject missing or changed
// documents instead of generating an answer or following a new source.
func TestTelegramPresentationDocumentReadGuard(t *testing.T) {
	for _, change := range []string{"missing", "source", "root"} {
		t.Run(change, func(t *testing.T) { checkTelegramOriginalDocumentGuard(t, change) })
	}
}

func checkTelegramOriginalDocumentGuard(t *testing.T, change string) {
	t.Helper()

	f, conversation, _ := telegramOriginalDocumentFixture(t)

	kind := "dump"
	if change == "missing" {
		kind = "spec"
	}

	request, err := reserveTelegramOriginalDocument(t, f, conversation, kind)
	if change == "missing" {
		if !errors.Is(err, errAuthoringDocumentUnavailable) {
			t.Fatal("missing document was not rejected", err)
		}

		return
	}

	if err != nil {
		t.Fatal(err)
	}

	prepared, err := telegramRetainedExplanation(request)
	if err != nil {
		t.Fatal(err)
	}

	if change == "source" {
		path := telegramOriginalDocumentPath(t, f, prepared.Sources[0].Reference)
		if err := os.WriteFile(path, []byte("Independent replacement.\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	} else {
		telegramAuthoringChangeAuthority(t, f, "root")
	}

	operations := telegramOperationConcreteService(t, f.receiver)

	_, err = operations.ReadPreparedTelegramDocument(t.Context(), request.Target, prepared)
	if !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("read followed a changed document or registry root", err)
	}

	checkTelegramNoWork(t, f.receiver.env)
}

// A confirmed original-document read keeps its full bytes across SQLite reopen,
// even if the canonical file is edited before the first multipart upload.
func TestTelegramPresentationDocumentReadRecovery(t *testing.T) {
	f, conversation, generations := telegramOriginalDocumentFixture(t)

	request, err := reserveTelegramOriginalDocument(t, f, conversation, "dump")
	if err != nil {
		t.Fatal(err)
	}

	path, original := queueTelegramOriginalDocumentRead(t, f, request)

	later := []byte("Independent later document.\n")
	if err := os.WriteFile(path, later, 0o600); err != nil {
		t.Fatal(err)
	}

	telegramAuthoringReopen(t, f)

	if err := f.receiver.drain(t.Context()); err != nil {
		t.Fatal(err)
	}

	document := telegramLatestDocument(t, f, "dump.md")
	if !bytes.Equal(document.Payload, original) || generations() != 1 {
		t.Fatal("recovery regenerated the answer or changed original attachment bytes")
	}

	current, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(current, later) {
		t.Fatal("document delivery overwrote the independent edit", err)
	}

	checkTelegramNoWork(t, f.receiver.env)
}

func queueTelegramOriginalDocumentRead(t *testing.T, f *telegramConversationAppFixture, request telegramcore.OperationRequest) (string, []byte) {
	t.Helper()

	prepared, err := telegramRetainedExplanation(request)
	if err != nil {
		t.Fatal(err)
	}

	path := telegramOriginalDocumentPath(t, f, prepared.Sources[0].Reference)

	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	config, err := f.receiver.currentConfig()
	if err != nil {
		t.Fatal(err)
	}

	if err := f.receiver.continueExplanation(t.Context(), request, config); err != nil {
		t.Fatal(err)
	}

	return path, original
}

func telegramOriginalDocumentFixture(t *testing.T) (*telegramConversationAppFixture, string, func() int64) {
	t.Helper()

	f := newTelegramConversationAppFixture(t)
	calls := telegramAuthoringBridgeHTTP(t, f, func(authoringModelInput) string {
		return telegramAuthoringServiceOutput(t, "# Saved request\n\nPreserve the retry ceiling.\n", "")
	}, telegramAuthoringBridgeCapture)
	f.send(t, "/prepare original document", 0)

	conversation := f.sessions(t)[0].ID
	f.send(t, "/done "+conversation, 0)
	telegramLatestDocument(t, f, "dump-")

	return f, conversation, calls.Load
}

func reserveTelegramOriginalDocument(t *testing.T, f *telegramConversationAppFixture, conversation, kind string) (telegramcore.OperationRequest, error) {
	t.Helper()

	input := telegramOperationRetainInput(t, f, "Show the "+kind+" document.")
	input.Conversation = conversation
	command := telegramExplanationCommand{Input: input, Question: input.Update.Text, DocumentKind: authoringcore.AuthoringArtifactKind(kind)}
	user := input.Config.AllowedUsers[0]
	operations := telegramOperationConcreteService(t, f.receiver)

	return f.receiver.reserveExplanation(t.Context(), operations, user, command)
}

func telegramOriginalDocumentPath(t *testing.T, f *telegramConversationAppFixture, ref string) string {
	t.Helper()

	operations := telegramOperationConcreteService(t, f.receiver)

	projects, err := operations.ListRegisteredProjects([]string{"example-project"})
	if err != nil || len(projects) != 1 {
		t.Fatal("project unavailable", err)
	}

	return filepath.Join(projects[0].Path, filepath.FromSlash(ref))
}
