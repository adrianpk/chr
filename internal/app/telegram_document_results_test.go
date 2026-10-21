package app

import (
	"bytes"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
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
	telegramAuthoringBridgeHTTP(t, f, func(authoringModelInput) string {
		return telegramAuthoringServiceOutput(t, "# Requirements\n\nKeep the retry ceiling.\n\n## Validation\n\nVerify recovery.", "")
	}, telegramAuthoringBridgeCapture)
	f.send(t, "/prepare readable documents", 0)

	c := f.sessions(t)[0]
	for _, step := range []struct {
		command string
		kind    core.AuthoringArtifactKind
	}{
		{"/done", core.AuthoringArtifactKindDump},
		{"/spec", core.AuthoringArtifactKindSpec},
		{"/refine spec", core.AuthoringArtifactKindSpec},
		{"/plan", core.AuthoringArtifactKindPlan},
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
		})
	}

	checkTelegramNoWork(t, f.receiver.env)
}

func assertTelegramArtifactUpload(t *testing.T, document telegramDocumentUpload, snapshot core.AuthoringSnapshot, kind core.AuthoringArtifactKind) {
	t.Helper()

	for _, artifact := range snapshot.Artifacts {
		if artifact.Ref.Kind == kind && bytes.Equal([]byte(artifact.Body), document.Payload) {
			return
		}
	}

	t.Fatal("attachment differs from its complete canonical artifact", document.Filename)
}
