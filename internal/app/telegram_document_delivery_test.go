package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramDocumentUpload struct {
	Filename string
	Payload  []byte
	Fields   map[string]string
}

func telegramDocumentHTTPFixture(t *testing.T, uploads chan<- telegramDocumentUpload, status int) *httptest.Server {
	t.Helper()

	var calls atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/sendDocument") {
			t.Error("document used the wrong method", r.URL.Path)
		}

		upload := readTelegramDocumentUpload(t, r)
		uploads <- upload

		if calls.Add(1) == 1 && status != 0 {
			w.WriteHeader(status)

			if _, err := fmt.Fprintf(w, `{"ok":false,"error_code":%d,"description":"Bad Request: can't parse entities"}`, status); err != nil {
				t.Error(err)
			}

			return
		}

		if _, err := io.WriteString(w, `{"ok":true,"result":{"message_id":77,"document":{"file_id":"original-upload"}}}`); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(server.Close)

	return server
}

func readTelegramDocumentUpload(t *testing.T, r *http.Request) telegramDocumentUpload {
	t.Helper()

	upload := telegramDocumentUpload{Fields: map[string]string{}}

	reader, err := r.MultipartReader()
	if err != nil {
		t.Error(err)

		return upload
	}

	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			t.Error(err)

			return upload
		}

		content, err := io.ReadAll(io.LimitReader(part, (2<<20)+1))
		if err != nil {
			t.Error(err)
		}

		if part.FormName() == "document" {
			upload.Filename, upload.Payload = part.FileName(), content
		} else {
			upload.Fields[part.FormName()] = string(content)
		}
	}

	return upload
}

func queueTelegramDocumentFixture(t *testing.T, receiver telegramReceiver, config core.TelegramReceiverConfig) core.TelegramDocumentResponse {
	t.Helper()
	addTelegramAppTestProjects(t, receiver.env)

	update := retainTelegramHelpInput(t, receiver, 1, "show document")

	projects, err := receiver.operations.ListRegisteredProjects(config.AllowedUsers[0].ProjectIDs)
	if err != nil {
		t.Fatal(err)
	}

	target, err := receiver.store.ResolveTelegramOperationTarget(t.Context(), core.TelegramConversationInput{Update: update, Config: config, Projects: projects}, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte("---\nstatus: draft\n---\n# Original 😀\n```literal```\n")

	response := core.TelegramDocumentResponse{Message: core.TelegramMessageBlocks{InitialText: "Draft ready", ContentKind: "markdown", Content: core.TelegramDocumentExcerpt(payload), SecondaryText: "Complete document attached."}, Document: core.TelegramDocument{ProjectID: target.ProjectID, Namespace: target.Namespace, ProjectBindingDigest: target.ProjectBindingDigest, SourceRef: "ops/default/spec/example.md", SourceHash: recordDigest(payload), Filename: "example.md", MediaType: "text/markdown", Payload: payload}}
	if err := receiver.store.QueueDocumentResponse(t.Context(), update, response, config.Digest(), time.Now()); err != nil {
		t.Fatal(err)
	}

	return response
}

// Real multipart HTTP and SQLite preserve the same complete image on definite
// format fallback. A server failure retains uncertainty without an immediate retry.
func TestTelegramPresentationDeliveryHTTP(t *testing.T) {
	for _, test := range []struct {
		name, state             string
		status, sends, attempts int
	}{
		{"upload", "", 0, 1, 0},
		{"plain fallback", "", 400, 2, 0},
		{"uncertain", "uncertain", 503, 1, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			uploads := make(chan telegramDocumentUpload, 3)
			server := telegramDocumentHTTPFixture(t, uploads, test.status)
			receiver, config := telegramAppTestReceiver(t, core.TelegramReceiverAPI{Client: server.Client(), BaseURL: server.URL})

			response := queueTelegramDocumentFixture(t, receiver, config)
			if err := receiver.drainDeliveries(t.Context()); err != nil {
				t.Fatal(err)
			}

			checkTelegramDocumentUploads(t, uploads, test.sends, response.Document.Payload)
			assertTelegramHelpDelivery(t, receiver, test.state, test.attempts)
		})
	}
}

func checkTelegramDocumentUploads(t *testing.T, uploads <-chan telegramDocumentUpload, sends int, want []byte) {
	t.Helper()

	if len(uploads) != sends {
		t.Fatal("wrong send allowance", len(uploads), sends)
	}

	var caption string

	for index := range sends {
		upload := <-uploads
		checkTelegramDocumentFile(t, upload, want)

		if index == 0 {
			caption = upload.Fields["caption"]

			var entities []core.TelegramTextEntity
			if err := json.Unmarshal([]byte(upload.Fields["caption_entities"]), &entities); err != nil || len(entities) != 1 {
				t.Fatal("literal caption entities missing", entities, err)
			}
		} else if upload.Fields["caption_entities"] != "" || upload.Fields["caption"] != caption {
			t.Fatal("plain retry changed content or retained formatting", upload)
		}
	}
}

// Registry revocation after snapshot admission blocks the original upload and
// preserves its bytes for operator-owned recovery rather than leaking a document.
func TestTelegramPresentationDeliveryRevocation(t *testing.T) {
	uploads := make(chan telegramDocumentUpload, 1)
	server := telegramDocumentHTTPFixture(t, uploads, 0)
	receiver, config := telegramAppTestReceiver(t, core.TelegramReceiverAPI{Client: server.Client(), BaseURL: server.URL})
	queueTelegramDocumentFixture(t, receiver, config)

	registry, err := core.NewProjectRegistry(t.Context(), receiver.env, testAssetsFS())
	if err != nil {
		t.Fatal(err)
	}

	if _, err := registry.UpdateProjectPath(t.Context(), "example-project", t.TempDir()); err != nil {
		t.Fatal(err)
	}

	if err := registry.Close(); err != nil {
		t.Fatal(err)
	}

	if err := receiver.drainDeliveries(t.Context()); err != nil || len(uploads) != 0 {
		t.Fatal("stale project image was uploaded", len(uploads), err)
	}

	assertTelegramHelpDelivery(t, receiver, "blocked", 1)
}

func checkTelegramDocumentFile(t *testing.T, upload telegramDocumentUpload, want []byte) {
	t.Helper()

	if upload.Filename != "example.md" || !bytes.Equal(upload.Payload, want) || upload.Fields["chat_id"] != "100001" {
		t.Fatal("upload changed bytes, filename or destination", upload)
	}
}
