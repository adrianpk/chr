package app

import (
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

func TestDecodeTelegramUpdatesParsesCommandRecords(t *testing.T) {
	updates, err := telegramcore.DecodeUpdates(strings.NewReader(telegramUpdateFixture()))
	if err != nil {
		t.Fatalf("DecodeTelegramUpdates returned error: %v", err)
	}

	if len(updates) != 2 {
		t.Fatalf("len(updates) = %d, want 2", len(updates))
	}

	record, err := remoteCommandRecordFromTelegramUpdate(updates[0], telegramCommandTestTime())
	if err != nil {
		t.Fatalf("RemoteCommandRecordFromTelegramUpdate returned error: %v", err)
	}

	if record.Status != core.RemoteCommandRecordAccepted {
		t.Fatalf("status = %q, want accepted", record.Status)
	}

	if record.Request.ProjectSelector != "example" || record.Request.Intent != core.RemoteCommandIntentLoopStart {
		t.Fatalf("request = %#v, want example loop start", record.Request)
	}
}

func TestRemoteCommandRecordFromTelegramUpdateRejectsUnsupportedText(t *testing.T) {
	update := telegramcore.Update{
		Message: telegramcore.UpdateMessage{
			From: telegramcore.SenderUser{ID: 10, Username: "ops"},
			Chat: telegramcore.Chat{ID: 20},
			Date: 1_768_000_000,
			Text: "merge dev",
		},
	}

	record, err := remoteCommandRecordFromTelegramUpdate(update, telegramCommandTestTime())
	if err != nil {
		t.Fatalf("RemoteCommandRecordFromTelegramUpdate returned error: %v", err)
	}

	if record.Status != core.RemoteCommandRecordRejected {
		t.Fatalf("status = %q, want rejected", record.Status)
	}

	if !strings.Contains(record.RejectionReason, "unsupported remote command") {
		t.Fatalf("rejection = %q, want unsupported command", record.RejectionReason)
	}
}

func telegramUpdateFixture() string {
	return `{"ok":true,"result":[
{"update_id":1,"message":{"message_id":10,"from":{"id":100,"username":"ops"},"chat":{"id":200},"date":1768000000,"text":"@example loop start --status ready --max-iterations 1"}},
{"update_id":2,"message":{"message_id":11,"from":{"id":101,"first_name":"Ada"},"chat":{"id":201},"date":1768000001,"text":"ticket start CHR-1"}}
]}`
}

func telegramCommandTestTime() time.Time {
	return time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
}
