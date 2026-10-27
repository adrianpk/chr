package authoring

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPlanAuthoringArtifactLayoutResolvesCanonicalPaths(t *testing.T) {
	projectDir := t.TempDir()

	for _, test := range authoringArtifactKindPathTests(projectDir) {
		t.Run(string(test.kind), func(t *testing.T) {
			layout, err := PlanArtifactLayout(ArtifactLayoutOptions{
				ProjectDir:  projectDir,
				SessionID:   "auth-20260521-120000",
				Kind:        test.kind,
				Role:        ArtifactRoleOutput,
				Topic:       "Billing Sync",
				Description: "Generated artifact.",
			})
			if err != nil {
				t.Fatalf("PlanAuthoringArtifactLayout returned error: %v", err)
			}

			if layout.Ref.Path != test.wantRelativePath {
				t.Fatalf("Ref.Path = %q, want %q", layout.Ref.Path, test.wantRelativePath)
			}

			wantFullPath := filepath.Join(projectDir, filepath.FromSlash(test.wantRelativePath))
			if layout.FullPath != wantFullPath {
				t.Fatalf("FullPath = %q, want %q", layout.FullPath, wantFullPath)
			}
		})
	}
}

func authoringArtifactKindPathTests(projectDir string) []struct {
	kind             ArtifactKind
	wantRelativePath string
} {
	_ = projectDir

	return []struct {
		kind             ArtifactKind
		wantRelativePath string
	}{
		{
			kind:             ArtifactKindDump,
			wantRelativePath: "ops/default/dump/00000-auth-20260521-120000-billing-sync/artifact.md",
		},
		{
			kind:             ArtifactKindQuiz,
			wantRelativePath: "ops/default/quiz/00000-auth-20260521-120000-billing-sync/artifact.md",
		},
		{
			kind:             ArtifactKindDraft,
			wantRelativePath: "ops/default/spec/00000-auth-20260521-120000-billing-sync/artifact.md",
		},
		{
			kind:             ArtifactKindSpec,
			wantRelativePath: "ops/default/spec/00000-auth-20260521-120000-billing-sync/artifact.md",
		},
		{
			kind:             ArtifactKindPlan,
			wantRelativePath: "ops/default/plan/00000-auth-20260521-120000-billing-sync/artifact.md",
		},
		{
			kind:             ArtifactKindSliceSet,
			wantRelativePath: "ops/default/plan/00000-auth-20260521-120000-billing-sync/artifact.md",
		},
		{
			kind:             ArtifactKindSlice,
			wantRelativePath: "ops/default/plan/00000-auth-20260521-120000-billing-sync/artifact.md",
		},
		{
			kind:             ArtifactKindTracker,
			wantRelativePath: "ops/default/tracker/00000-auth-20260521-120000-billing-sync/artifact.md",
		},
		{
			kind:             ArtifactKindTicket,
			wantRelativePath: "ops/default/ticket/00000-auth-20260521-120000-billing-sync/artifact.md",
		},
	}
}

func TestPlanAuthoringArtifactLayoutDetectsCollision(t *testing.T) {
	projectDir := t.TempDir()
	layout := planAuthoringArtifactLayoutForTest(t, projectDir, false)

	err := os.MkdirAll(filepath.Dir(layout.FullPath), 0o755)
	if err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	err = os.WriteFile(layout.FullPath, []byte("# Existing\n"), 0o644)
	if err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	colliding := planAuthoringArtifactLayoutForTest(t, projectDir, false)
	if !colliding.Collision.Exists {
		t.Fatal("Collision.Exists = false, want true")
	}

	if colliding.Collision.Allowed {
		t.Fatal("Collision.Allowed = true, want false")
	}
}

func TestPlanAuthoringArtifactLayoutAllowsExplicitReplace(t *testing.T) {
	projectDir := t.TempDir()
	layout := planAuthoringArtifactLayoutForTest(t, projectDir, true)

	err := os.MkdirAll(filepath.Dir(layout.FullPath), 0o755)
	if err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	err = os.WriteFile(layout.FullPath, []byte("# Existing\n"), 0o644)
	if err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	colliding := planAuthoringArtifactLayoutForTest(t, projectDir, true)
	if !colliding.Collision.Exists {
		t.Fatal("Collision.Exists = false, want true")
	}

	if !colliding.Collision.Allowed {
		t.Fatal("Collision.Allowed = false, want true")
	}
}

func TestPlanAuthoringArtifactLayoutRejectsInvalidInput(t *testing.T) {
	for _, test := range invalidAuthoringArtifactLayoutTests(t.TempDir()) {
		t.Run(test.name, func(t *testing.T) {
			_, err := PlanArtifactLayout(test.opts)
			if err == nil {
				t.Fatal("PlanAuthoringArtifactLayout returned nil error")
			}
		})
	}
}

func invalidAuthoringArtifactLayoutTests(projectDir string) []struct {
	name string
	opts ArtifactLayoutOptions
} {
	return []struct {
		name string
		opts ArtifactLayoutOptions
	}{
		{
			name: "project dir",
			opts: ArtifactLayoutOptions{
				SessionID: "auth-1",
				Kind:      ArtifactKindQuiz,
				Role:      ArtifactRoleOutput,
				Topic:     "Topic",
			},
		},
		{
			name: "session id",
			opts: ArtifactLayoutOptions{
				ProjectDir: projectDir,
				Kind:       ArtifactKindQuiz,
				Role:       ArtifactRoleOutput,
				Topic:      "Topic",
			},
		},
		{
			name: "session-owned kind",
			opts: ArtifactLayoutOptions{
				ProjectDir: projectDir,
				SessionID:  "auth-1",
				Kind:       ArtifactKindSession,
				Role:       ArtifactRoleOutput,
				Topic:      "Topic",
			},
		},
		{
			name: "topic or name",
			opts: ArtifactLayoutOptions{
				ProjectDir: projectDir,
				SessionID:  "auth-1",
				Kind:       ArtifactKindQuiz,
				Role:       ArtifactRoleOutput,
			},
		},
	}
}

func TestAuthoringArtifactInventoryMarksCollisions(t *testing.T) {
	projectDir := t.TempDir()
	layout := planAuthoringArtifactLayoutForTest(t, projectDir, false)
	layout.Collision.Exists = true
	layout.Collision.Allowed = false

	entries := ArtifactInventory("auth-1", []ArtifactLayout{layout})
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}

	if !entries[0].Collision {
		t.Fatal("entries[0].Collision = false, want true")
	}
}

func planAuthoringArtifactLayoutForTest(
	t *testing.T,
	projectDir string,
	allowReplace bool,
) ArtifactLayout {
	t.Helper()

	layout, err := PlanArtifactLayout(ArtifactLayoutOptions{
		ProjectDir:   projectDir,
		SessionID:    "auth-1",
		Kind:         ArtifactKindQuiz,
		Role:         ArtifactRoleOutput,
		Topic:        "Billing Sync",
		AllowReplace: allowReplace,
	})
	if err != nil {
		t.Fatalf("PlanAuthoringArtifactLayout returned error: %v", err)
	}

	return layout
}
