package app

import (
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func guidedTelegramSharedFixture(t *testing.T) (*telegramConversationAppFixture, service, webViewProvider, *guidedWorkHost, *guidedWebPublicationFixture) {
	t.Helper()

	f := newTelegramConversationAppFixture(t)
	provider, host, publication := guidedWebJourneyFixture(t, false)
	operations := provider.service(t.Context())

	registry, err := core.NewProjectRegistry(t.Context(), f.receiver.env, testAssetsFS())
	if err != nil {
		t.Fatal(err)
	}

	project := core.Project{ID: operations.projectContext.ProjectID, Name: "Shared project", Path: operations.projectContext.ProjectDir, Enabled: true, RegisteredAt: time.Now()}
	if err := registry.AddProject(t.Context(), project); err != nil {
		t.Fatal(err)
	}

	if err := registry.Close(); err != nil {
		t.Fatal(err)
	}

	config, err := core.LoadTelegramReceiverConfig(f.receiver.env)
	if err != nil {
		t.Fatal(err)
	}

	config.AllowedUsers[0].ProjectIDs = append(config.AllowedUsers[0].ProjectIDs, project.ID)
	if err := core.SaveTelegramReceiverConfig(f.receiver.env, config, true); err != nil {
		t.Fatal(err)
	}

	operations.env = f.receiver.env
	f.receiver.operations = operations

	return f, operations, provider, host, publication
}

// TestGuidedUseCaseTelegramJourney attaches the same Web run, lets the shared
// host execute exact work, preserves its native review wait, and projects only
// receipt-backed completion after lost publication acknowledgement recovery.
func TestGuidedUseCaseTelegramJourney(t *testing.T) {
	f, operations, provider, host, publication := guidedTelegramSharedFixture(t)
	original := readWebGuidedSession(t, operations, "guided-quiz")
	prompt := f.send(t, "Show use case Guided clarification in project Shared project", 0)
	c := f.sessions(t)[0]

	after := readGuidedTelegramSession(t, operations, c)
	if c.GuidedRunID != original.ID || after.Revision != original.Revision || c.ID == original.ID || !strings.Contains(prompt.Text, original.Topic) {
		t.Fatal("attachment created or advanced another run", prompt.Text)
	}

	lifecycle := &telegramGuidedHost{receiver: f.receiver, shared: host}
	pumpGuidedTelegram(t, lifecycle)
	host.wait.Wait()

	started := readGuidedTelegramSession(t, operations, f.sessions(t)[0])

	run := assertGuidedNativeAdmission(t, operations, started)
	if run.Review == nil || run.Review.State != "waiting_approval" || serviceWorkTaskAttempts(run) != 1 {
		t.Fatal("shared activation bypassed review or repeated work")
	}

	pumpGuidedTelegram(t, lifecycle)

	if err := f.receiver.drainDeliveries(t.Context()); err != nil {
		t.Fatal(err)
	}

	assertGuidedTelegramCoalesced(t, f, lifecycle)
	assertGuidedTelegramResultExport(t, f, false)
	assertGuidedTelegramNativeControls(t, f, operations, host)
	f.send(t, "Idea: add an export in a later review", 0)
	completeGuidedWebPublication(t, provider, host, publication, started)
	pumpGuidedTelegram(t, lifecycle)

	if err := f.receiver.drainDeliveries(t.Context()); err != nil {
		t.Fatal(err)
	}

	completed := assertGuidedTelegramCompleted(t, f, operations, original)
	assertGuidedTelegramResultExport(t, f, true)
	assertGuidedTelegramPreparedOriginals(t, f, operations, completed)
}

func pumpGuidedTelegram(t *testing.T, lifecycle *telegramGuidedHost) {
	t.Helper()

	for range 2 {
		if err := lifecycle.pump(t.Context()); err != nil {
			t.Fatal(err)
		}
	}
}

func assertGuidedTelegramCoalesced(t *testing.T, f *telegramConversationAppFixture, lifecycle *telegramGuidedHost) {
	t.Helper()

	before := f.sessions(t)[0]
	pumpGuidedTelegram(t, lifecycle)

	after := f.sessions(t)[0]
	if before.Revision != after.Revision || before.QuestionBinding != after.QuestionBinding {
		t.Fatal("unchanged status replaced a prompt or queued another notice")
	}

	var waiting bool

	for len(f.notices) > 0 {
		notice := <-f.notices
		waiting = waiting || strings.Contains(notice.Text, "native review requires approval")
	}

	if !waiting {
		t.Fatal("real native wait was not presented")
	}
}

func assertGuidedTelegramNativeControls(t *testing.T, f *telegramConversationAppFixture, operations service, host *guidedWorkHost) {
	t.Helper()

	before := readGuidedTelegramSession(t, operations, f.sessions(t)[0])

	original := assertGuidedNativeAdmission(t, operations, before)
	for _, action := range []string{"Pause use case", "Resume use case"} {
		f.send(t, action, 0)
		host.wait.Wait()
	}

	current := readGuidedTelegramSession(t, operations, f.sessions(t)[0])

	run := assertGuidedNativeAdmission(t, operations, current)
	if run.ID != original.ID || run.Policy != original.Policy || serviceWorkTaskAttempts(run) != 1 {
		t.Fatal("Telegram controls exchanged native scope or limits")
	}
}

func assertGuidedTelegramCompleted(t *testing.T, f *telegramConversationAppFixture, operations service, original core.AuthoringSession) core.AuthoringSession {
	t.Helper()

	completed := readGuidedTelegramSession(t, operations, f.sessions(t)[0])
	if f.sessions(t)[0].Phase != "completed" || completed.ID != original.ID || len(completed.Guided.LaterIdeas) != 1 || completed.Guided.Actor != original.Guided.Actor {
		t.Fatal("completion exchanged scope, actor or later ideas")
	}

	return completed
}
