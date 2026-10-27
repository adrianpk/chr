package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

func assertGuidedTelegramOriginals(t *testing.T, f *telegramConversationAppFixture, operations service, session authoringcore.AuthoringSession) {
	t.Helper()

	c := f.sessions(t)[0]
	root := guidedTelegramProject(t, operations, c.ProjectID)
	replaceGuidedWebDocuments(t, root, session)

	for _, kind := range []string{"dump", "quiz", "specification"} {
		scoped := operations
		scoped.projectContext = commandProjectContext{ProjectID: c.ProjectID, ProjectDir: root}

		original, err := scoped.ReadGuidedUseCaseOriginal(t.Context(), session.ID, session.Namespace, kind)
		if err != nil {
			t.Fatal(err)
		}

		response := f.send(t, "Show "+kind, 0)

		upload := readGuidedTelegramUpload(t, f)
		if !bytes.Equal(upload.Payload, []byte(original.Content)) || strings.Contains(string(upload.Payload), "Later replacement") || !strings.Contains(response.Text, "Original "+kind) {
			t.Fatal("attachment replaced or shortened its original", kind, response.Text)
		}
	}
}

func readGuidedTelegramUpload(t *testing.T, f *telegramConversationAppFixture) telegramDocumentUpload {
	t.Helper()

	select {
	case upload := <-f.documents:
		return upload
	default:
		t.Fatal("complete original was not attached")

		return telegramDocumentUpload{}
	}
}

func assertGuidedTelegramResultExport(t *testing.T, f *telegramConversationAppFixture, complete bool) {
	t.Helper()

	notice := f.send(t, "Show results", 0)
	if !strings.Contains(notice.Text, "Result evidence") {
		telegramWaitNotice(t, f, "Result evidence", 3*time.Second)
	}

	upload := readGuidedTelegramUpload(t, f)

	var result guidedResultExport
	if err := json.Unmarshal(upload.Payload, &result); err != nil {
		t.Fatal(err)
	}

	if result.Complete != complete {
		t.Fatal("export invented or lost confirmed completion")
	}

	if complete && (len(result.Deliveries) == 0 || len(result.Closures) == 0) {
		t.Fatal("result omitted original native receipts")
	}
}

func assertGuidedTelegramPreparedOriginals(t *testing.T, f *telegramConversationAppFixture, operations service, session authoringcore.AuthoringSession) {
	t.Helper()

	for _, kind := range []string{"plan", "tracker"} {
		original, err := operations.ReadGuidedUseCaseOriginal(t.Context(), session.ID, session.Namespace, kind)
		if err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(operations.projectContext.ProjectDir, filepath.FromSlash(original.Reference)), []byte("Later replacement"), 0o600); err != nil {
			t.Fatal(err)
		}

		f.send(t, "Show "+kind, 0)

		upload := readGuidedTelegramUpload(t, f)
		if !bytes.Equal(upload.Payload, []byte(original.Content)) {
			t.Fatal("prepared original followed its current working copy", kind)
		}
	}
}

// TestGuidedUseCaseTelegramResultGuard refuses a terminal display claim without
// a verified native completion result, including incomplete closing evidence.
func TestGuidedUseCaseTelegramResultGuard(t *testing.T) {
	for _, test := range []struct {
		name   string
		result *core.WorkflowResultInspection
	}{
		{name: "missing result"},
		{name: "partial evidence", result: &core.WorkflowResultInspection{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			session := authoringcore.AuthoringSession{Topic: "Result guard", Guided: &authoringcore.GuidedUseCaseState{Stage: "results", Status: "completed"}}
			observed := GuidedUseCaseInspection{Session: &session, Result: test.result}

			text := telegramGuidedPrompt(observed)
			if !strings.Contains(text, "Stage: results · blocked") || !strings.Contains(text, "unconfirmed") {
				t.Fatal("terminal state replaced native receipt verification", text)
			}
		})
	}
}
