package core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	"forge.adrianpk.com/chr/internal/core/database"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

type telegramConversationFixture struct {
	env   UserEnvironment
	store *telegramFixtureStore
	input telegramcore.TelegramConversationInput
	next  int64
}

func newTelegramConversationFixture(t *testing.T) *telegramConversationFixture {
	t.Helper()

	config := telegramReceiverTestConfig()
	config.AllowedUsers[0].ProjectIDs = []string{"alpha", "beta"}
	config.DefaultProjectID = ""

	env := telegramReceiverTestEnv(t)

	return &telegramConversationFixture{env: env, store: openTelegramReceiverTestStore(t, env), input: telegramcore.TelegramConversationInput{
		Config: config, Projects: []projectcore.Project{{ID: "alpha", Name: "Alpha Project", Path: t.TempDir(), Enabled: true}, {ID: "beta", Name: "Beta Project", Path: t.TempDir(), Enabled: true}}, Intent: "prepare", Text: "Topic",
	}, next: 1}
}

func (f *telegramConversationFixture) apply(t *testing.T, action telegramcore.TelegramConversationAction, change func(*telegramcore.TelegramConversationInput)) {
	t.Helper()

	input := f.input
	input.Action = action
	input.Update = telegramRetainedTestUpdate(t, f.next, fmt.Sprintf("source %d", f.next))
	f.next++

	if change != nil {
		change(&input)
	}

	if err := f.store.RetainBatch(t.Context(), 12345, []telegramcore.TelegramReceivedUpdate{input.Update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	if err := f.store.ApplyTelegramConversation(t.Context(), input, time.Now()); err != nil {
		t.Fatal(err)
	}
}

func (f *telegramConversationFixture) sessions(t *testing.T) []telegramcore.TelegramConversation {
	t.Helper()

	sessions, err := f.store.Conversations(t.Context(), telegramcore.TelegramPeerIdentity{BotID: 12345, ChatID: 100001, UserID: 100001})
	if err != nil {
		t.Fatal(err)
	}

	return sessions
}

func telegramReceiverTestConfig() telegramcore.TelegramReceiverConfig {
	return telegramcore.TelegramReceiverConfig{
		Enabled: true, BotTokenEnv: "EXAMPLE_BOT_TOKEN", PollTimeoutSeconds: 30,
		AllowedUsers: []telegramcore.TelegramAllowedUser{{TelegramUserID: 100001, OperatorID: "maintainer", ProjectIDs: []string{"example-project"}}},
	}
}

func telegramReceiverTestEnv(t *testing.T) UserEnvironment {
	t.Helper()

	env, err := NewUserEnvironment(UserEnvironmentOptions{DevMode: true, DevRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}

	return env
}

func openTelegramReceiverTestStore(t *testing.T, env UserEnvironment) *telegramFixtureStore {
	t.Helper()

	store, err := newTelegramFixtureStore(t.Context(), env, os.DirFS("../../cmd/chr"))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})

	if err := store.BindReceiver(t.Context(), 12345, telegramReceiverTestConfig(), time.Now()); err != nil {
		t.Fatal(err)
	}

	return store
}

func seedTelegramDocumentQuota(t *testing.T, store *telegramFixtureStore, count int) {
	t.Helper()

	zero := make([]byte, tgmessage.DocumentLimit)

	_, err := store.db.ExecContext(t.Context(), `WITH RECURSIVE n(x) AS (SELECT 1 UNION ALL SELECT x+1 FROM n WHERE x<?)
 INSERT INTO telegram_document_snapshots(id,owner_kind,owner_id,bot_id,chat_id,topic_id,user_id,config_digest,project_id,namespace,project_binding_digest,source_ref,source_hash,content_hash,filename,media_type,byte_count,payload,created_at)
 SELECT 'quota-'||x,'receiver','quota-owner-'||x,12345,100001,0,100001,?,'example-project','default',?,'exports/quota.md',?,?,'quota.md','text/markdown',2097152,zeroblob(2097152),? FROM n`, count, telegramReceiverTestConfig().Digest(), strings.Repeat("a", 64), contenthash.Sum(zero), contenthash.Sum(zero), database.FormatTime(time.Now()))
	if err != nil {
		t.Fatal(err)
	}
}

func telegramTestDocumentResponse() tgmessage.DocumentResponse {
	payload := []byte("---\nstatus: draft\n---\n# Original\n```literal```\n")

	return tgmessage.DocumentResponse{Message: tgmessage.Blocks{InitialText: "Draft ready", ContentKind: "markdown", Content: tgmessage.DocumentExcerpt(payload), SecondaryText: "Complete document attached."}, Document: tgmessage.Document{ProjectID: "example-project", Namespace: "default", ProjectBindingDigest: strings.Repeat("a", 64), SourceRef: "ops/default/spec/example.md", SourceHash: contenthash.Sum(payload), Filename: "example.md", MediaType: "text/markdown", Payload: payload}}
}

func telegramRecoveryMigrationAssets(t *testing.T) fstest.MapFS {
	t.Helper()

	source := os.DirFS("../../cmd/chr")

	entries, err := fs.ReadDir(source, "assets/migration/sqlite")
	if err != nil {
		t.Fatal(err)
	}

	assets := fstest.MapFS{}

	for _, entry := range entries {
		path := "assets/migration/sqlite/" + entry.Name()

		data, err := fs.ReadFile(source, path)
		if err != nil {
			t.Fatal(err)
		}

		assets[path] = &fstest.MapFile{Data: data}
	}

	return assets
}

// telegramFixtureStore keeps SQL setup/inspection in the integration fixture;
// production consumers receive only the receiver's public capabilities.
type telegramFixtureStore struct {
	*telegramcore.TelegramReceiverStore
	db *sql.DB
}

func newTelegramFixtureStore(ctx context.Context, env telegramcore.TelegramStoreDirectories, assets fs.FS) (*telegramFixtureStore, error) {
	owner, err := telegramcore.NewTelegramReceiverStore(ctx, env, assets)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", database.UserDatabasePath(env))
	if err != nil {
		return nil, errors.Join(err, owner.Close())
	}

	if err := database.Configure(ctx, db); err != nil {
		return nil, errors.Join(err, db.Close(), owner.Close())
	}

	return &telegramFixtureStore{TelegramReceiverStore: owner, db: db}, nil
}

func (s *telegramFixtureStore) Close() error {
	return errors.Join(s.TelegramReceiverStore.Close(), s.db.Close())
}

// The wire fixture feeds the public polling codec without recreating its rules.
type telegramFixtureWire string

func (w telegramFixtureWire) Do(request *http.Request) (*http.Response, error) {
	if err := request.Body.Close(); err != nil {
		return nil, err
	}

	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(string(w)))}, nil
}

func telegramRetainedTestUpdate(t *testing.T, id int64, text string) telegramcore.TelegramReceivedUpdate {
	t.Helper()

	wire := fmt.Sprintf(`{"ok":true,"result":[{"update_id":%d,"message":{"message_id":1,"from":{"id":100001},"chat":{"id":100001,"type":"private"},"text":%q}}]}`, id, text)

	updates, err := (telegramcore.TelegramReceiverAPI{Client: telegramFixtureWire(wire), BaseURL: "https://telegram.invalid"}).Poll(t.Context(), "12345:example_test_token_123456789", nil, 1)
	if err != nil || len(updates) != 1 {
		t.Fatal("fixture wire update unavailable", updates, err)
	}

	return telegramcore.AuthorizeTelegramReceiverUpdate(telegramReceiverTestConfig(), 12345, updates[0])
}
