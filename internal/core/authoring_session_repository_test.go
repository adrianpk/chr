package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

func TestAuthoringSessionRepositoryCreatesSessionFiles(t *testing.T) {
	ctx := context.Background()
	projectDir := t.TempDir()
	repository := NewAuthoringSessionRepository(projectDir)
	session := validAuthoringSession()

	paths, err := repository.CreateSession(ctx, session)
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	for _, path := range []string{paths.SessionPath, paths.TranscriptPath, paths.ArtifactsPath} {
		_, err = os.Stat(path)
		if err != nil {
			t.Fatalf("expected file %s: %v", path, err)
		}
	}

	wantDir := filepath.Join(projectDir, ".chr", "authoring", "sessions", session.ID)
	if paths.SessionDir != wantDir {
		t.Fatalf("SessionDir = %q, want %q", paths.SessionDir, wantDir)
	}
}

func TestAuthoringSessionRepositoryLoadsCreatedSession(t *testing.T) {
	ctx := context.Background()
	projectDir := t.TempDir()
	repository := NewAuthoringSessionRepository(projectDir)
	session := validAuthoringSession()

	_, err := repository.CreateSession(ctx, session)
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	loaded, err := repository.GetSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("GetSession returned error: %v", err)
	}

	assertAuthoringSessionRoundTrip(t, loaded, session)
}

func TestAuthoringSessionRepositoryUpdatesSession(t *testing.T) {
	ctx := context.Background()
	projectDir := t.TempDir()
	repository := NewAuthoringSessionRepository(projectDir)
	session := validAuthoringSession()

	_, err := repository.CreateSession(ctx, session)
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	session.Topic = "Updated authoring topic"

	_, err = repository.UpdateSession(ctx, session)
	if err != nil {
		t.Fatalf("UpdateSession returned error: %v", err)
	}

	loaded, err := repository.GetSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("GetSession returned error: %v", err)
	}

	if loaded.Topic != session.Topic {
		t.Fatalf("Topic = %q, want %q", loaded.Topic, session.Topic)
	}
}

func TestAuthoringSessionRepositoryListsSessions(t *testing.T) {
	ctx := context.Background()
	projectDir := t.TempDir()
	repository := NewAuthoringSessionRepository(projectDir)
	first := validAuthoringSession()
	second := validAuthoringSession()
	second.ID = "auth-20260521-130000"
	second.Topic = "Second authoring topic"

	_, err := repository.CreateSession(ctx, first)
	if err != nil {
		t.Fatalf("CreateSession first returned error: %v", err)
	}

	_, err = repository.CreateSession(ctx, second)
	if err != nil {
		t.Fatalf("CreateSession second returned error: %v", err)
	}

	sessions, err := repository.ListSessions(ctx)
	if err != nil {
		t.Fatalf("ListSessions returned error: %v", err)
	}

	if len(sessions) != 2 {
		t.Fatalf("len(sessions) = %d, want 2", len(sessions))
	}
}

func TestAuthoringSessionRepositoryRejectsDuplicateSession(t *testing.T) {
	ctx := context.Background()
	projectDir := t.TempDir()
	repository := NewAuthoringSessionRepository(projectDir)
	session := validAuthoringSession()

	_, err := repository.CreateSession(ctx, session)
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	_, err = repository.CreateSession(ctx, session)
	if err == nil {
		t.Fatal("CreateSession returned nil error")
	}
}

func TestAuthoringSessionRepositoryReportsMissingSession(t *testing.T) {
	ctx := context.Background()
	projectDir := t.TempDir()
	repository := NewAuthoringSessionRepository(projectDir)

	_, err := repository.GetSession(ctx, "missing")
	if !errors.Is(err, ErrAuthoringSessionNotFound) {
		t.Fatalf("GetSession error = %v, want ErrAuthoringSessionNotFound", err)
	}
}

func TestAuthoringSessionRepositoryReportsMissingSessionOnUpdate(t *testing.T) {
	ctx := context.Background()
	projectDir := t.TempDir()
	repository := NewAuthoringSessionRepository(projectDir)
	session := validAuthoringSession()

	_, err := repository.UpdateSession(ctx, session)
	if !errors.Is(err, ErrAuthoringSessionNotFound) {
		t.Fatalf("UpdateSession error = %v, want ErrAuthoringSessionNotFound", err)
	}
}

func TestRenderAndParseAuthoringSession(t *testing.T) {
	session := validAuthoringSession()
	paths := AuthoringSessionLayout("/workspace/project", session.ID)

	content, err := RenderAuthoringSession(session, paths)
	if err != nil {
		t.Fatalf("RenderAuthoringSession returned error: %v", err)
	}

	parsed, err := ParseAuthoringSession(content)
	if err != nil {
		t.Fatalf("ParseAuthoringSession returned error: %v", err)
	}

	assertAuthoringSessionRoundTrip(t, parsed, session)
}

func assertAuthoringSessionRoundTrip(t *testing.T, got, want authoringcore.Session) {
	t.Helper()
	assertAuthoringSessionMetadata(t, got, want)
	assertAuthoringSessionArtifacts(t, got, want)
	assertAuthoringSessionProvenance(t, got, want)
}

func assertAuthoringSessionMetadata(t *testing.T, got, want authoringcore.Session) {
	t.Helper()

	if got.ID != want.ID {
		t.Fatalf("ID = %q, want %q", got.ID, want.ID)
	}

	if got.ProjectID != want.ProjectID {
		t.Fatalf("ProjectID = %q, want %q", got.ProjectID, want.ProjectID)
	}

	if got.Status != want.Status {
		t.Fatalf("Status = %q, want %q", got.Status, want.Status)
	}

	if got.Mode != want.Mode {
		t.Fatalf("Mode = %q, want %q", got.Mode, want.Mode)
	}

	if got.Target != want.Target {
		t.Fatalf("Target = %q, want %q", got.Target, want.Target)
	}

	if got.Topic != want.Topic {
		t.Fatalf("Topic = %q, want %q", got.Topic, want.Topic)
	}
}

func assertAuthoringSessionArtifacts(t *testing.T, got, want authoringcore.Session) {
	t.Helper()

	if len(got.Inputs) != len(want.Inputs) {
		t.Fatalf("len(Inputs) = %d, want %d", len(got.Inputs), len(want.Inputs))
	}

	if got.Inputs[0].Path != want.Inputs[0].Path {
		t.Fatalf("Inputs[0].Path = %q, want %q", got.Inputs[0].Path, want.Inputs[0].Path)
	}

	if len(got.Outputs) != len(want.Outputs) {
		t.Fatalf("len(Outputs) = %d, want %d", len(got.Outputs), len(want.Outputs))
	}

	if got.Outputs[0].Path != want.Outputs[0].Path {
		t.Fatalf("Outputs[0].Path = %q, want %q", got.Outputs[0].Path, want.Outputs[0].Path)
	}
}

func assertAuthoringSessionProvenance(t *testing.T, got, want authoringcore.Session) {
	t.Helper()

	if len(got.Provenance) != len(want.Provenance) {
		t.Fatalf("len(Provenance) = %d, want %d", len(got.Provenance), len(want.Provenance))
	}

	if got.Provenance[0].Relation != want.Provenance[0].Relation {
		t.Fatalf("Provenance[0].Relation = %q, want %q", got.Provenance[0].Relation, want.Provenance[0].Relation)
	}
}
