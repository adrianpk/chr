package app

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/core/database"
)

// Legacy dispatch uses the persisted original payload and old sender contract.
// A retained delivered status blocks another send even with a stale caller copy.
func TestTelegramPresentationNotificationLegacy(t *testing.T) {
	for _, manifest := range []string{"", `{"version":0}`} {
		t.Run(manifest, func(t *testing.T) { checkTelegramLegacyNotificationDispatch(t, manifest) })
	}
}

func checkTelegramLegacyNotificationDispatch(t *testing.T, manifest string) {
	t.Helper()

	repository, input, project, env := newTelegramNotificationAppInput(t)

	input.Record.Event.Payload.Detail = "Original legacy detail"
	if err := repository.CreateNotificationEvent(t.Context(), input.Record); err != nil {
		t.Fatal(err)
	}

	if manifest != "" {
		setTelegramLegacyNotificationManifest(t, env, manifest)
	}

	want := core.RenderTelegramNotificationMessage(input.Record.Event).Text

	var calls atomic.Int32

	installTelegramNotificationSender(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)

		var payload core.TelegramMessage
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}

		if payload.Text != want {
			t.Error("legacy payload was rerendered from caller data", payload.Text)
		}

		if _, err := w.Write([]byte(`{"ok":true}`)); err != nil {
			t.Error(err)
		}
	})

	input.Record.Event.Payload.Detail = "Caller changed after persistence"
	for range 2 {
		if _, err := dispatchTelegramNotification(t.Context(), repository, input.Record, project, input.Target); err != nil {
			t.Fatal(err)
		}
	}

	if calls.Load() != 1 {
		t.Fatal("legacy status acquired a new retry", calls.Load())
	}
}

func setTelegramLegacyNotificationManifest(t *testing.T, env core.UserEnvironment, manifest string) {
	t.Helper()

	db, err := sql.Open("sqlite", database.UserDatabasePath(env))
	if err != nil {
		t.Fatal(err)
	}

	_, writeErr := db.ExecContext(t.Context(), "UPDATE notification_events SET telegram_presentation=?", manifest)

	closeErr := db.Close()
	if writeErr != nil || closeErr != nil {
		t.Fatal(writeErr, closeErr)
	}
}
