package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

func guidedTelegramFixture(t *testing.T) (*telegramConversationAppFixture, service) {
	t.Helper()

	if testing.Short() {
		t.Skip("real guided project, Git and Telegram transport fixture")
	}

	f := newTelegramConversationAppFixture(t)

	operations, ok := f.receiver.operations.(service)
	if !ok {
		t.Fatal("missing shared operation service")
	}

	projects, err := operations.ListRegisteredProjects([]string{"example-project"})
	if err != nil || len(projects) != 1 {
		t.Fatal("missing registered guided project", err)
	}

	root := projects[0].Path
	// Keep real project/Git owners while replacing only content-provider IO.
	if err := os.MkdirAll(filepath.Join(root, ".chr"), 0o755); err != nil {
		t.Fatal(err)
	}

	initializeGuidedWebGit(t, root)

	now := operations.utcNow()
	operations.quizModel = &fakeServiceQuizModel{responses: []core.QuizModelResponse{serviceQuizModelResponse("question", "question-1", now)}}
	f.receiver.operations = operations

	return f, operations
}

// TestGuidedUseCaseTelegramCapture routes ordinary messages through the same
// canonical run and proves that replay and stale peer replies cannot append twice.
func TestGuidedUseCaseTelegramCapture(t *testing.T) {
	f, operations := guidedTelegramFixture(t)

	start := f.send(t, "Develop a new use case", 0)
	if !strings.Contains(start.Text, "use-case name") {
		t.Fatal("natural start did not ask for a name", start.Text)
	}

	name := f.send(t, "Bounded audit", start.ID)
	if !strings.Contains(name.Text, "Stage: dump") {
		t.Fatal("name did not reach canonical capture", name.Text)
	}

	capture := f.send(t, "Retain exact validation and delivery evidence.", name.ID)
	c := f.sessions(t)[0]

	id, err := telegramAuthoringSessionID(c.AuthoringSessionRef)
	if err != nil {
		t.Fatal(err)
	}

	scope := core.GuidedUseCaseScope{RunID: id, ProjectID: c.ProjectID, Namespace: c.Namespace}
	root := guidedTelegramProject(t, operations, c.ProjectID)

	before, err := core.ReadGuidedUseCase(t.Context(), root, scope)
	if err != nil || strings.Count(before.Guided.Capture, "Retain exact") != 1 {
		t.Fatal("capture was not retained once", err)
	}

	assertGuidedTelegramReplay(t, f, root, scope, before, start.ID)

	if capture.ID == 0 {
		t.Fatal("capture response was not delivered")
	}
}

func assertGuidedTelegramReplay(t *testing.T, f *telegramConversationAppFixture, root string, scope core.GuidedUseCaseScope, before core.AuthoringSession, startID int64) {
	t.Helper()

	f.updates <- f.last

	if _, err := f.receiver.step(t.Context()); err != nil {
		t.Fatal(err)
	}

	stale := f.send(t, "Replace the old name", startID)
	if !strings.Contains(stale.Text, "Saved details are unchanged") {
		t.Fatal("stale prompt was not rejected", stale.Text)
	}

	after, err := core.ReadGuidedUseCase(t.Context(), root, scope)
	if err != nil || after.Revision != before.Revision || after.Guided.Capture != before.Guided.Capture {
		t.Fatal("replay or stale prompt changed canonical state", err)
	}
}

func guidedTelegramProject(t *testing.T, operations service, id string) string {
	t.Helper()

	projects, err := operations.ListRegisteredProjects([]string{id})
	if err != nil || len(projects) != 1 {
		t.Fatal(err)
	}

	return projects[0].Path
}

// TestGuidedUseCaseTelegramStartIntent excludes incidental or negated model text
// from the deterministic development instruction accepted by the guided bridge.
func TestGuidedUseCaseTelegramStartIntent(t *testing.T) {
	cases := []struct {
		text string
		want bool
	}{
		{"Develop a new use case", true},
		{"Start guided development.", true},
		{"Do not develop a new use case", false},
		{"Explain how to develop a new use case", false},
		{"/develop manual topic", false},
	}
	for _, tt := range cases {
		t.Run(tt.text, func(t *testing.T) {
			if got := telegramGuidedStart(tt.text); got != tt.want {
				t.Fatal("development intent changed", got, tt.want)
			}
		})
	}
}
