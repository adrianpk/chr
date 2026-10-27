package app

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf16"

	"forge.adrianpk.com/chr/internal/core"
	notifycore "forge.adrianpk.com/chr/internal/core/notify"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

type telegramNotificationCapture struct {
	mu       sync.Mutex
	bodies   [][]byte
	captions []string
	entities []string
}

func (c *telegramNotificationCapture) receive(t *testing.T) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(3 << 20); err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		file, header, err := r.FormFile("document")
		if err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		body, readErr := io.ReadAll(io.LimitReader(file, 3<<20))

		closeErr := file.Close()
		if readErr != nil || closeErr != nil || header.Filename != "notification-event.json" {
			t.Error("invalid complete export", readErr, closeErr)
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		c.mu.Lock()
		c.bodies = append(c.bodies, body)
		c.captions = append(c.captions, r.FormValue("caption"))
		c.entities = append(c.entities, r.FormValue("caption_entities"))
		attempt := len(c.bodies)
		c.mu.Unlock()

		if attempt == 1 {
			w.WriteHeader(http.StatusBadRequest)

			_, err = w.Write([]byte(`{"ok":false,"error_code":400,"description":"Bad Request: can't parse entities"}`))
		} else {
			_, err = w.Write([]byte(`{"ok":true,"result":{"message_id":1001,"document":{"file_id":"saved-export"}}}`))
		}

		if err != nil {
			t.Error(err)
		}
	}
}

func newTelegramNotificationAppInput(t *testing.T) (notifycore.Repository, notifycore.TelegramInput, commandProjectContext, core.UserEnvironment) {
	t.Helper()

	dir := newTestWorkspace(t)
	writeTelegramNotifyConfig(t, dir)
	t.Setenv("CHR_TELEGRAM_BOT_TOKEN", "notification-token")

	env, err := core.NewUserEnvironment(core.UserEnvironmentOptions{DevMode: true, DevRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}

	repository, err := notifycore.NewRepository(t.Context(), env, testAssetsFS())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := repository.Close(); err != nil {
			t.Error(err)
		}
	})

	project := commandProjectContext{ProjectID: projectID(dir), ProjectDir: dir, ProjectName: "Example project"}
	now := time.Now().UTC()
	record := notifycore.AuditRecord{Channel: "telegram", Event: notifycore.Event{ID: "notice-original", ProjectID: project.ProjectID, Type: notifycore.EventLoopFinished, Severity: notifycore.SeverityInfo, Status: notifycore.StatusPending, Source: notifycore.SourceTicketLoop, References: notifycore.References{LoopRunID: "loop-original"}, Payload: notifycore.Payload{Summary: "Ticket loop finished", Detail: "Processed: 1; blocked: 0; failed: 0.\nWorker-reported summary: " + strings.Repeat("literal `_*` 😀\n", 500)}, CreatedAt: now, UpdatedAt: now}}

	config, err := core.LoadWorkspaceConfig(dir)
	if err != nil {
		t.Fatal(err)
	}

	input, err := telegramNotificationInput(record, project, config.Notifications.Telegram)
	if err != nil {
		t.Fatal(err)
	}

	input.Message.ContentKind = "code"

	return repository, input, project, env
}

// HTTP and SQLite prove a format fallback sends the same complete export and
// caption through the notification owner without a receiver or another trigger.
func TestTelegramPresentationNotificationDispatch(t *testing.T) {
	repository, input, project, _ := newTelegramNotificationAppInput(t)
	capture := &telegramNotificationCapture{}
	server := httptest.NewServer(capture.receive(t))
	t.Cleanup(server.Close)

	previous := notifyTelegramSender
	notifyTelegramSender = telegramcore.TelegramSender{Client: server.Client(), BaseURL: server.URL}

	t.Cleanup(func() { notifyTelegramSender = previous })

	original := bytes.Clone(input.Document.Payload)
	if err := repository.CreateTelegram(t.Context(), input); err != nil {
		t.Fatal(err)
	}

	input.Document.Payload = []byte("changed after admission")
	if _, err := dispatchTelegramNotification(t.Context(), repository, input.Record, project, input.Target); err != nil {
		t.Fatal(err)
	}

	checkTelegramNotificationCapture(t, capture, original)

	lookup, err := repository.ReadTelegram(t.Context(), input.Record.Event.ID, input.Target)

	saved, found := lookup.Delivery, lookup.Frozen
	if err != nil || !found || saved.Attempts != 2 || !saved.PlainFallback || saved.Acknowledgement.FileID != "saved-export" || saved.Status != notifycore.StatusDelivered {
		t.Fatal("notification lost its own acknowledgement", saved, err)
	}
}

func checkTelegramNotificationCapture(t *testing.T, c *telegramNotificationCapture, original []byte) {
	t.Helper()
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.bodies) != 2 {
		t.Fatal("fallback changed dispatch count", len(c.bodies))
	}

	for _, body := range c.bodies {
		if !bytes.Equal(body, original) || !json.Valid(body) {
			t.Fatal("fallback replaced the complete original export")
		}
	}

	checkTelegramNotificationCaption(t, c)
}

func checkTelegramNotificationCaption(t *testing.T, c *telegramNotificationCapture) {
	t.Helper()

	if c.captions[0] != c.captions[1] || len(utf16.Encode([]rune(c.captions[0]))) > 900 || !strings.Contains(c.captions[0], "Example project") || !strings.Contains(c.captions[0], "canonical receipt") || c.entities[0] == "" || c.entities[1] != "" {
		t.Fatal("fallback lost bounded context or formatting selection", c.captions, c.entities)
	}
}
