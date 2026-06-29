package core

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadAssistedAuthoringContextPackageIncludesProtocolsDocumentsAndTranscriptSummary(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "AGENTS.md"), "# Workspace Protocol\n")
	writeTestFile(t, filepath.Join(dir, ".chr", "user", "manifest.md"), "# Workspace Manifest\n")
	writeTestFile(t, filepath.Join(dir, ".chr", "user", "config.yml"), "version: 1\n")
	writeTestFile(t, filepath.Join(dir, "docs", "reference", "authoring.md"), "# Authoring Reference\n")

	session := testAssistedContextSession(dir)
	paths := AuthoringSessionLayout(dir, session.ID)
	events := testAssistedContextTranscriptEvents(t, session.ID)

	pkg, err := LoadAssistedAuthoringContextPackage(AssistedAuthoringContextPackageOptions{
		ProjectDir:       dir,
		Session:          session,
		SessionPaths:     paths,
		TranscriptEvents: events,
	})
	if err != nil {
		t.Fatalf("LoadAssistedAuthoringContextPackage returned error: %v", err)
	}

	if pkg.TranscriptSummary.QuestionCount != 2 {
		t.Fatalf("QuestionCount = %d, want 2", pkg.TranscriptSummary.QuestionCount)
	}

	if pkg.TranscriptSummary.AnswerCount != 1 {
		t.Fatalf("AnswerCount = %d, want 1", pkg.TranscriptSummary.AnswerCount)
	}

	if len(pkg.TranscriptSummary.UnansweredQuestions) != 1 {
		t.Fatalf("UnansweredQuestions = %d, want 1", len(pkg.TranscriptSummary.UnansweredQuestions))
	}

	rendered := RenderAssistedAuthoringContextPackage(pkg)
	for _, want := range []string{
		"Assisted Authoring Context",
		"Act as a technical analyst",
		"# Workspace Protocol",
		"# Workspace Manifest",
		"# Authoring Reference",
		"Implicit conversation history is not included",
		"question-2",
		"authoring-session:transcript",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered context missing %q:\n%s", want, rendered)
		}
	}
}

func TestRenderAssistedAuthoringContextPackageIsDeterministic(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "AGENTS.md"), "# Workspace Protocol\n")
	writeTestFile(t, filepath.Join(dir, "docs", "reference", "authoring.md"), "# Authoring Reference\n")

	session := testAssistedContextSession(dir)
	paths := AuthoringSessionLayout(dir, session.ID)
	events := testAssistedContextTranscriptEvents(t, session.ID)

	pkg, err := LoadAssistedAuthoringContextPackage(AssistedAuthoringContextPackageOptions{
		ProjectDir:       dir,
		Session:          session,
		SessionPaths:     paths,
		TranscriptEvents: events,
	})
	if err != nil {
		t.Fatalf("LoadAssistedAuthoringContextPackage returned error: %v", err)
	}

	first := RenderAssistedAuthoringContextPackage(pkg)
	second := RenderAssistedAuthoringContextPackage(pkg)

	if first != second {
		t.Fatalf("rendered context is not deterministic:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestLoadAssistedAuthoringContextPackageRecordsMissingSelectedDocuments(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "AGENTS.md"), "# Workspace Protocol\n")

	session := testAssistedContextSession(dir)
	session.Inputs = append(session.Inputs, AuthoringArtifactRef{
		ID:          "missing-doc",
		Kind:        AuthoringArtifactKindDraft,
		Role:        AuthoringArtifactRoleInput,
		Path:        "docs/missing.md",
		Description: "Missing document.",
	})

	pkg, err := LoadAssistedAuthoringContextPackage(AssistedAuthoringContextPackageOptions{
		ProjectDir:       dir,
		Session:          session,
		TranscriptEvents: testAssistedContextTranscriptEvents(t, session.ID),
	})
	if err != nil {
		t.Fatalf("LoadAssistedAuthoringContextPackage returned error: %v", err)
	}

	var found bool

	for _, source := range pkg.Sources {
		if source.ID == "authoring-document:missing-doc" {
			found = true

			if source.Included {
				t.Fatal("missing selected document source was marked included")
			}

			if source.Reason != "file was not found" {
				t.Fatalf("missing selected document reason = %q", source.Reason)
			}
		}
	}

	if !found {
		t.Fatal("missing selected document source was not recorded")
	}
}

func TestLoadAssistedAuthoringContextPackageRejectsTranscriptSessionMismatch(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "AGENTS.md"), "# Workspace Protocol\n")
	writeTestFile(t, filepath.Join(dir, "docs", "reference", "authoring.md"), "# Authoring Reference\n")

	session := testAssistedContextSession(dir)
	events := testAssistedContextTranscriptEvents(t, "other-session")

	_, err := LoadAssistedAuthoringContextPackage(AssistedAuthoringContextPackageOptions{
		ProjectDir:       dir,
		Session:          session,
		TranscriptEvents: events,
	})
	if err == nil {
		t.Fatal("LoadAssistedAuthoringContextPackage returned nil error")
	}
}

func testAssistedContextSession(dir string) AuthoringSession {
	now := time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC)

	return AuthoringSession{
		ID:        "session-1",
		ProjectID: "project-1",
		Status:    AuthoringSessionStatusActive,
		Mode:      AuthoringSessionModeAssisted,
		Target:    AuthoringTargetSpec,
		Topic:     "Assisted authoring",
		Inputs: []AuthoringArtifactRef{
			{
				ID:          "authoring-reference",
				Kind:        AuthoringArtifactKindDraft,
				Role:        AuthoringArtifactRoleInput,
				Path:        filepath.ToSlash(filepath.Join("docs", "reference", "authoring.md")),
				Description: "Authoring reference.",
			},
		},
		CreatedAt:   now,
		StartedAt:   &now,
		UpdatedAt:   now,
		Description: "Collect context for assisted authoring.",
	}
}

func testAssistedContextTranscriptEvents(t *testing.T, sessionID string) []AssistedTranscriptEvent {
	t.Helper()

	now := time.Date(2026, 5, 21, 11, 0, 0, 0, time.UTC)
	turns := []AssistedAuthoringTurn{
		{
			ID:        "turn-1",
			SessionID: sessionID,
			Index:     1,
			Status:    AssistedAuthoringTurnAnswered,
			Intent:    AssistedAuthoringIntentDiscover,
			Question: AssistedAuthoringQuestion{
				ID:       "question-1",
				Prompt:   "What should the artifact explain?",
				Required: true,
			},
			Answer: AssistedAuthoringAnswer{
				QuestionID: "question-1",
				Text:       "The assisted authoring flow.",
				AnsweredAt: now.Add(1 * time.Minute),
			},
			CreatedAt: now,
			UpdatedAt: now.Add(1 * time.Minute),
		},
		{
			ID:        "turn-2",
			SessionID: sessionID,
			Index:     2,
			Status:    AssistedAuthoringTurnAsked,
			Intent:    AssistedAuthoringIntentClarify,
			Question: AssistedAuthoringQuestion{
				ID:       "question-2",
				Prompt:   "Which gaps remain?",
				Required: true,
			},
			CreatedAt: now.Add(2 * time.Minute),
			UpdatedAt: now.Add(2 * time.Minute),
		},
	}

	events := []AssistedTranscriptEvent{}

	for _, turn := range turns {
		turnEvents, err := AssistedTranscriptEventsFromTurn(turn)
		if err != nil {
			t.Fatalf("AssistedTranscriptEventsFromTurn returned error: %v", err)
		}

		events = append(events, turnEvents...)
	}

	return events
}
