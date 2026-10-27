package app

import (
	"bytes"
	"errors"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

type telegramGuidedUploadAck struct {
	base telegramcore.TelegramHTTPClient
	lost atomic.Bool
}

func (transport *telegramGuidedUploadAck) Do(request *http.Request) (*http.Response, error) {
	response, err := transport.base.Do(request)
	if err == nil && strings.HasSuffix(request.URL.Path, "/sendDocument") && !transport.lost.Swap(true) {
		return nil, errors.Join(response.Body.Close(), errTelegramLostRecordAck)
	}

	return response, err
}

// TestGuidedUseCaseTelegramUploadRecovery loses a real multipart upload's reply,
// reopens SQLite and resends its original snapshot with the consumed attempt kept.
func TestGuidedUseCaseTelegramUploadRecovery(t *testing.T) {
	t.Parallel()

	f, operations := guidedTelegramFixture(t)
	start := f.send(t, "Develop a new use case", 0)
	name := f.send(t, "Original upload", start.ID)
	capture := f.send(t, "Keep the complete original.", name.ID)
	f.send(t, "Finish capture", capture.ID)

	before := readGuidedTelegramSession(t, operations, f.sessions(t)[0])

	update := retainTelegramHelpInput(t, f.receiver, 100, "Show dump")
	if err := f.receiver.respond(t.Context(), update); err != nil {
		t.Fatal(err)
	}

	f.receiver.api.Client = &telegramGuidedUploadAck{base: f.receiver.api.Client}
	if err := f.receiver.drainDeliveries(t.Context()); err != nil {
		t.Fatal(err)
	}

	original := readGuidedTelegramUpload(t, f)
	replaceGuidedWebDocuments(t, guidedTelegramProject(t, operations, before.ProjectID), before)
	telegramAuthoringReopen(t, f)
	resendGuidedTelegramOriginal(t, f, original)

	after := readGuidedTelegramSession(t, operations, f.sessions(t)[0])

	request, found, err := f.receiver.store.TelegramOperation(t.Context(), update, 0)
	if err != nil || !found || request.State != "confirmed" || after.Revision != before.Revision {
		t.Fatal("transport recovery replayed or lost the canonical operation", err)
	}
}

func resendGuidedTelegramOriginal(t *testing.T, f *telegramConversationAppFixture, original telegramDocumentUpload) {
	t.Helper()

	now := time.Now().Add(time.Minute)

	due, err := f.receiver.store.DueDeliveries(t.Context(), f.receiver.botID, now)
	if err != nil || len(due) != 1 || due[0].Attempts != 1 || due[0].State != "uncertain" {
		t.Fatal("reopen lost the original uncertain attempt", due, err)
	}

	if err := f.receiver.store.ReserveDelivery(t.Context(), due[0].ID, now); err != nil {
		t.Fatal(err)
	}

	if err := f.receiver.sendReservedDelivery(t.Context(), due[0]); err != nil {
		t.Fatal(err)
	}

	upload := readGuidedTelegramUpload(t, f)
	assertGuidedTelegramUploadImage(t, upload, original)

	due, err = f.receiver.store.DueDeliveries(t.Context(), f.receiver.botID, now)
	if err != nil || len(due) != 0 {
		t.Fatal("definite upload acknowledgement did not finish delivery", err)
	}
}

func assertGuidedTelegramUploadImage(t *testing.T, upload, original telegramDocumentUpload) {
	t.Helper()

	if upload.Filename != original.Filename || !bytes.Equal(upload.Payload, original.Payload) || upload.Fields["chat_id"] != original.Fields["chat_id"] {
		t.Fatal("uncertain retry replaced the complete original or destination")
	}
}
