package authoring

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSynthesizeAssistedArtifactWritesEditableArtifactAndUpdatesProvenance(t *testing.T) {
	dir := t.TempDir()
	repository, session, sourceRef, summary := setupAssistedArtifactSynthesisTest(t, dir)

	result, err := SynthesizeAssistedArtifact(context.Background(), AssistedArtifactSynthesisOptions{
		ProjectDir:     dir,
		Session:        session,
		CuratedSummary: summary,
		SourceRef:      sourceRef,
		TargetKind:     ArtifactKindSpec,
		TargetName:     "assisted spec",
		Repository:     repository,
		Now:            func() time.Time { return time.Date(2026, 5, 21, 22, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("SynthesizeAssistedArtifact returned error: %v", err)
	}

	content := readAssistedArtifactSynthesisFile(t, result.Layout.FullPath)
	for _, want := range []string{
		"Assisted authoring",
		"The user wants a concise assisted authoring workflow.",
		"The assisted authoring flow.",
		"Provenance",
		sourceRef.Path,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("synthesized artifact missing %q:\n%s", want, content)
		}
	}

	updated, err := repository.GetSession(context.Background(), session.ID)
	if err != nil {
		t.Fatalf("GetSession returned error: %v", err)
	}

	if len(updated.Outputs) != 2 {
		t.Fatalf("session output count = %d, want 2", len(updated.Outputs))
	}

	if len(updated.Provenance) != 1 {
		t.Fatalf("session provenance count = %d, want 1", len(updated.Provenance))
	}

	if updated.Provenance[0].Source.ID != sourceRef.ID {
		t.Fatalf("provenance source = %q, want %q", updated.Provenance[0].Source.ID, sourceRef.ID)
	}
}

func TestSynthesizeAssistedArtifactRejectsCollisionWithoutReplace(t *testing.T) {
	dir := t.TempDir()
	repository, session, sourceRef, summary := setupAssistedArtifactSynthesisTest(t, dir)
	layout := planAssistedArtifactSynthesisLayout(t, dir, session)
	writeTestFile(t, layout.FullPath, "# Existing\n")

	_, err := SynthesizeAssistedArtifact(context.Background(), AssistedArtifactSynthesisOptions{
		ProjectDir:     dir,
		Session:        session,
		CuratedSummary: summary,
		SourceRef:      sourceRef,
		TargetKind:     ArtifactKindSpec,
		TargetName:     "assisted spec",
		Repository:     repository,
	})
	if err == nil {
		t.Fatal("SynthesizeAssistedArtifact returned nil error")
	}

	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("collision error = %v", err)
	}
}

func TestSynthesizeAssistedArtifactAllowsExplicitReplace(t *testing.T) {
	dir := t.TempDir()
	repository, session, sourceRef, summary := setupAssistedArtifactSynthesisTest(t, dir)
	layout := planAssistedArtifactSynthesisLayout(t, dir, session)
	writeTestFile(t, layout.FullPath, "# Existing\n")

	_, err := SynthesizeAssistedArtifact(context.Background(), AssistedArtifactSynthesisOptions{
		ProjectDir:     dir,
		Session:        session,
		CuratedSummary: summary,
		SourceRef:      sourceRef,
		TargetKind:     ArtifactKindSpec,
		TargetName:     "assisted spec",
		AllowReplace:   true,
		Repository:     repository,
	})
	if err != nil {
		t.Fatalf("SynthesizeAssistedArtifact returned error: %v", err)
	}

	content := readAssistedArtifactSynthesisFile(t, layout.FullPath)
	if strings.Contains(content, "# Existing") {
		t.Fatalf("artifact was not replaced:\n%s", content)
	}
}

func TestRenderAssistedSynthesizedArtifactIsDeterministic(t *testing.T) {
	_, session, sourceRef, summary := setupAssistedArtifactSynthesisTest(t, t.TempDir())
	targetRef := ArtifactRef{
		ID:          "spec:session-1:assisted-spec",
		Kind:        ArtifactKindSpec,
		Role:        ArtifactRoleOutput,
		Path:        ".chr/default/specs/session-1-assisted-spec.md",
		Description: "Synthesized assisted authoring artifact.",
	}

	first := RenderAssistedSynthesizedArtifact(session, summary, sourceRef, targetRef)
	second := RenderAssistedSynthesizedArtifact(session, summary, sourceRef, targetRef)

	if first != second {
		t.Fatalf("rendered artifact is not deterministic:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func setupAssistedArtifactSynthesisTest(
	t *testing.T,
	dir string,
) (SessionRepository, Session, ArtifactRef, AssistedCuratedAnswerSummary) {
	t.Helper()

	repository := NewSessionRepository(dir)
	session := testAssistedContextSession(dir)
	sourcePath := filepath.Join(dir, ".chr", "authoring", "sessions", session.ID, "curated-summary.md")
	sourceRef := ArtifactRef{
		ID:          "curated-summary-" + session.ID,
		Kind:        ArtifactKindDraft,
		Role:        ArtifactRoleOutput,
		Path:        filepath.ToSlash(sourcePath),
		Description: "Curated assisted authoring answer summary.",
	}
	session.Outputs = append(session.Outputs, sourceRef)

	_, err := repository.CreateSession(context.Background(), session)
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	summary := BuildAssistedCuratedAnswerSummary(
		session.ID,
		testAssistedContextTranscriptEvents(t, session.ID),
		time.Date(2026, 5, 21, 21, 0, 0, 0, time.UTC),
	)
	summary.ModelSummary = "The user wants a concise assisted authoring workflow."

	return repository, session, sourceRef, summary
}

func planAssistedArtifactSynthesisLayout(
	t *testing.T,
	dir string,
	session Session,
) ArtifactLayout {
	t.Helper()

	layout, err := PlanArtifactLayout(ArtifactLayoutOptions{
		ProjectDir: dir,
		SessionID:  session.ID,
		Kind:       ArtifactKindSpec,
		Role:       ArtifactRoleOutput,
		Topic:      session.Topic,
		Name:       "assisted spec",
	})
	if err != nil {
		t.Fatalf("PlanAuthoringArtifactLayout returned error: %v", err)
	}

	return layout
}

func readAssistedArtifactSynthesisFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	return string(content)
}
