package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPlanAuthoringArtifactLayoutResolvesCanonicalPaths(t *testing.T) {
	projectDir := t.TempDir()

	for _, test := range authoringArtifactKindPathTests(projectDir) {
		t.Run(string(test.kind), func(t *testing.T) {
			layout, err := PlanAuthoringArtifactLayout(AuthoringArtifactLayoutOptions{
				ProjectDir:  projectDir,
				SessionID:   "auth-20260521-120000",
				Kind:        test.kind,
				Role:        AuthoringArtifactRoleOutput,
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
	kind             AuthoringArtifactKind
	wantRelativePath string
} {
	_ = projectDir

	return []struct {
		kind             AuthoringArtifactKind
		wantRelativePath string
	}{
		{
			kind:             AuthoringArtifactKindDump,
			wantRelativePath: ".chr/default/dumps/auth-20260521-120000-billing-sync.md",
		},
		{
			kind:             AuthoringArtifactKindQuiz,
			wantRelativePath: ".chr/default/quizzes/auth-20260521-120000-billing-sync.md",
		},
		{
			kind:             AuthoringArtifactKindDraft,
			wantRelativePath: ".chr/default/drafts/auth-20260521-120000-billing-sync.md",
		},
		{
			kind:             AuthoringArtifactKindSpec,
			wantRelativePath: ".chr/default/specs/auth-20260521-120000-billing-sync.md",
		},
		{
			kind:             AuthoringArtifactKindPlan,
			wantRelativePath: ".chr/default/plans/auth-20260521-120000-billing-sync.md",
		},
		{
			kind:             AuthoringArtifactKindSliceSet,
			wantRelativePath: ".chr/default/slice-sets/auth-20260521-120000-billing-sync.md",
		},
		{
			kind:             AuthoringArtifactKindSlice,
			wantRelativePath: ".chr/default/slices/auth-20260521-120000-billing-sync.md",
		},
		{
			kind:             AuthoringArtifactKindTracker,
			wantRelativePath: ".chr/default/trackers/auth-20260521-120000-billing-sync.md",
		},
		{
			kind:             AuthoringArtifactKindTicket,
			wantRelativePath: ".chr/default/tickets/auth-20260521-120000-billing-sync.md",
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
			_, err := PlanAuthoringArtifactLayout(test.opts)
			if err == nil {
				t.Fatal("PlanAuthoringArtifactLayout returned nil error")
			}
		})
	}
}

func invalidAuthoringArtifactLayoutTests(projectDir string) []struct {
	name string
	opts AuthoringArtifactLayoutOptions
} {
	return []struct {
		name string
		opts AuthoringArtifactLayoutOptions
	}{
		{
			name: "project dir",
			opts: AuthoringArtifactLayoutOptions{
				SessionID: "auth-1",
				Kind:      AuthoringArtifactKindQuiz,
				Role:      AuthoringArtifactRoleOutput,
				Topic:     "Topic",
			},
		},
		{
			name: "session id",
			opts: AuthoringArtifactLayoutOptions{
				ProjectDir: projectDir,
				Kind:       AuthoringArtifactKindQuiz,
				Role:       AuthoringArtifactRoleOutput,
				Topic:      "Topic",
			},
		},
		{
			name: "session-owned kind",
			opts: AuthoringArtifactLayoutOptions{
				ProjectDir: projectDir,
				SessionID:  "auth-1",
				Kind:       AuthoringArtifactKindSession,
				Role:       AuthoringArtifactRoleOutput,
				Topic:      "Topic",
			},
		},
		{
			name: "topic or name",
			opts: AuthoringArtifactLayoutOptions{
				ProjectDir: projectDir,
				SessionID:  "auth-1",
				Kind:       AuthoringArtifactKindQuiz,
				Role:       AuthoringArtifactRoleOutput,
			},
		},
	}
}

func TestAuthoringArtifactInventoryMarksCollisions(t *testing.T) {
	projectDir := t.TempDir()
	layout := planAuthoringArtifactLayoutForTest(t, projectDir, false)
	layout.Collision.Exists = true
	layout.Collision.Allowed = false

	entries := AuthoringArtifactInventory("auth-1", []AuthoringArtifactLayout{layout})
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
) AuthoringArtifactLayout {
	t.Helper()

	layout, err := PlanAuthoringArtifactLayout(AuthoringArtifactLayoutOptions{
		ProjectDir:   projectDir,
		SessionID:    "auth-1",
		Kind:         AuthoringArtifactKindQuiz,
		Role:         AuthoringArtifactRoleOutput,
		Topic:        "Billing Sync",
		AllowReplace: allowReplace,
	})
	if err != nil {
		t.Fatalf("PlanAuthoringArtifactLayout returned error: %v", err)
	}

	return layout
}
