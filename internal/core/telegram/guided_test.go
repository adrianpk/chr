package telegram

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	"forge.adrianpk.com/chr/internal/core/database"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
)

// TestGuidedUseCaseTelegramBinding reserves one peer-scoped association without
// creating a project run, acknowledging its input or bypassing manual isolation.
func TestGuidedUseCaseTelegramBinding(t *testing.T) {
	f := newTelegramConversationFixture(t)
	input := f.input
	input.Project, input.Text = "alpha", "New guided use case"

	input.Update = telegramRetainedTestUpdate(t, 1, "Develop a new use case")
	if err := f.store.RetainBatch(t.Context(), 12345, []ReceivedUpdate{input.Update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	first, target, err := f.store.ReserveGuidedConversation(t.Context(), input, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	second, again, err := f.store.ReserveGuidedConversation(t.Context(), input, time.Now())
	if err != nil || first.ID != second.ID || target != again || authoringEligible(first) {
		t.Fatal("routing replay replaced a binding or adopted manual authoring", err)
	}

	pending, err := f.store.PendingUpdates(t.Context(), 12345)
	if err != nil || len(pending) != 1 {
		t.Fatal("routing acknowledged the source before its owner", err)
	}

	if _, err := os.Stat(filepath.Join(input.Projects[0].Path, first.AuthoringSessionRef)); !os.IsNotExist(err) {
		t.Fatal("routing allocated canonical authoring state", err)
	}

	assertGuidedTelegramPeer(t, f, input, first)
}

func assertGuidedTelegramPeer(t *testing.T, f *telegramConversationFixture, input ConversationInput, c Conversation) {
	t.Helper()

	foreign := input
	foreign.Update.Peer.UserID++
	foreign.Config.AllowedUsers = append(foreign.Config.AllowedUsers, AllowedUser{TelegramUserID: foreign.Update.Peer.UserID, OperatorID: "other", ProjectIDs: []string{"alpha"}})

	foreign.Conversation = c.ID
	if _, _, err := f.store.ResolveGuidedTarget(t.Context(), foreign); err == nil {
		t.Fatal("foreign peer resolved a guided conversation")
	}

	moved := input
	moved.Projects = append([]projectcore.Project(nil), input.Projects...)

	moved.Projects[0].Path = t.TempDir()
	if _, _, err := f.store.ReserveGuidedConversation(t.Context(), moved, time.Now()); err == nil {
		t.Fatal("replay accepted a replaced registered root")
	}
}

// TestGuidedUseCaseTelegramMigration preserves real manual rows when the
// optional guided association is added; historical conversations stay manual.
func TestGuidedUseCaseTelegramMigration(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	assets := telegramRecoveryMigrationAssets(t)
	delete(assets, "assets/migration/sqlite/20261023000029-add_telegram_guided_binding.sql")

	old, err := NewReceiverStore(t.Context(), env, assets)
	if err != nil {
		t.Fatal(err)
	}

	if err := old.BindReceiver(t.Context(), 12345, telegramReceiverTestConfig(), time.Now()); err != nil {
		t.Fatal(err)
	}

	now := database.FormatTime(time.Now())
	execTelegramRecoverySQL(t, old, `INSERT INTO telegram_peers(bot_id,chat_id,topic_id,user_id,updated_at) VALUES(12345,100001,0,100001,?)`, now)
	execTelegramRecoverySQL(t, old, `INSERT INTO telegram_conversations(id,bot_id,chat_id,topic_id,user_id,project_id,project_binding_digest,namespace,intent,phase,topic,capture,revision,created_at,updated_at) VALUES('manual',12345,100001,0,100001,'alpha',?,'default','develop','collecting','Manual topic','Original context',7,?,?)`, contenthash.Sum([]byte("root")), now, now)

	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	current := openTelegramReceiverTestStore(t, env)

	assertGuidedTelegramManualMigration(t, current)
}

func assertGuidedTelegramManualMigration(t *testing.T, current *ReceiverStore) {
	t.Helper()

	rows, err := current.Conversations(t.Context(), PeerIdentity{BotID: 12345, ChatID: 100001, UserID: 100001})
	if err != nil || len(rows) != 1 {
		t.Fatal("migration lost manual conversation", err)
	}

	c := rows[0]
	if c.GuidedRunID != "" || c.Topic != "Manual topic" || c.Capture != "Original context" || c.Revision != 7 || !authoringEligible(c) {
		t.Fatal("migration replaced or adopted manual state")
	}
}
