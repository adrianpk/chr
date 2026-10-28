package workflow

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// TestWorkflowPreparationGuard revokes source-only and stale combined grants,
// preserving validated tasks when provider protection, policy or base changes.
func TestWorkflowPreparationGuard(t *testing.T) {
	cases := []struct {
		name   string
		change func(*testing.T, workflowPreparationFixture, IntegrationCandidate)
	}{
		{"base", func(t *testing.T, f workflowPreparationFixture, _ IntegrationCandidate) {
			t.Helper()

			f.provider.target = contenthash.Sum([]byte("moved"))[:40]
			f.provider.pr.BaseCommit = f.provider.target
		}},
		{"protection", func(_ *testing.T, f workflowPreparationFixture, _ IntegrationCandidate) {
			f.provider.revision = "guard-v2"
		}},
		{"source", func(_ *testing.T, f workflowPreparationFixture, _ IntegrationCandidate) {
			f.provider.pr.HeadCommit = contenthash.Sum([]byte("changed-source"))[:40]
		}},
		{"lane", func(t *testing.T, _ workflowPreparationFixture, c IntegrationCandidate) {
			t.Helper()

			if err := os.WriteFile(filepath.Join(c.Lane.Worktree, "target.txt"), []byte("changed"), 0o600); err != nil {
				t.Fatal(err)
			}
		}},
		{"evidence", func(t *testing.T, f workflowPreparationFixture, _ IntegrationCandidate) {
			t.Helper()

			run := workflowTestRead(t, f.store, f.run.ID)
			if err := os.WriteFile(filepath.Join(f.store.layout.Root(), run.Preparation.ResultRef), []byte("{}"), 0o600); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			f := newWorkflowPreparationFixture(t, "rebase", workflowstate.Validation{WaiverReason: "Guard fixture checks host proof"})

			c, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, f.provider, f.git)
			if err != nil {
				t.Fatal(err)
			}

			if _, err := f.store.ReviewWorkflowSlice(t.Context(), f.run.ID, f.provider, nil); err != nil {
				t.Fatal(err)
			}

			test.change(t, f, c)

			if _, err := f.store.CheckWorkflowPublicationGuard(t.Context(), f.run.ID, f.provider); err == nil {
				t.Fatal("changed immutable guard was accepted")
			}
		})
	}
}

// TestWorkflowPreparationUnsupported proves that unsupported protection blocks
// before preparation, without changing method, consuming a worker or publishing.
func TestWorkflowPreparationUnsupported(t *testing.T) {
	f := newWorkflowPreparationFixture(t, "squash", workflowstate.Validation{WaiverReason: "Protection fixture"})
	f.provider.guarded = false

	before := workflowTestRead(t, f.store, f.run.ID)
	if _, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, f.provider, f.git); !errors.Is(err, ErrWorkflowAtomicGuard) {
		t.Fatalf("guard: %v", err)
	}

	run := workflowTestRead(t, f.store, f.run.ID)
	if run.Status != "blocked" || run.Blocker.Reason != "atomic_source_base_guard_required" || run.PreparationCount != 0 || len(run.Attempts) != len(before.Attempts) {
		t.Fatal("unsupported guard launched preparation or discarded its blocker")
	}

	assertWorkflowPreparationSource(t, f)
}

// TestWorkflowPreparationProviderPaths rejects the current Forgejo and GitHub
// merge endpoints for all configured methods; capability discovery performs no
// unguarded mutation or automatic method fallback.
func TestWorkflowPreparationProviderPaths(t *testing.T) {
	for _, url := range []string{"https://codeberg.org/example/repo/pulls/12", "https://github.com/example/repo/pull/12"} {
		for _, method := range []string{"merge", "rebase", "squash"} {
			t.Run(url+"/"+method, func(t *testing.T) {
				runner := newFakeRemoteReviewRunner()
				adapter := NewRemoteReviewAdapter(workspaceconfig.DefaultReviewProviderConfig(), runner)
				pr := PullRequestSnapshot{URL: url, Repository: "https://example.invalid/owner/repo"}

				capability, err := adapter.WorkflowIntegrationCapability(t.Context(), pr, method)
				if err != nil {
					t.Fatal(err)
				}

				if err := validateWorkflowPublicationCapability(ReviewInput{PullRequest: pr, Policy: workflowstate.ReviewPolicy{Method: method}}, capability); !errors.Is(err, ErrWorkflowAtomicGuard) {
					t.Fatalf("provider falsely claimed guard: %v", err)
				}

				if len(runner.inputs) != 0 {
					t.Fatal("capability discovery used an unguarded merge")
				}
			})
		}
	}
}

type workflowPreparationGitHook struct {
	git   PreparationGit
	after func()
}

func (g workflowPreparationGitHook) PrepareWorkflowIntegration(ctx context.Context, query PreparationQuery) (IntegrationProof, error) {
	proof, err := g.git.PrepareWorkflowIntegration(ctx, query)
	if err == nil {
		g.after()
	}

	return proof, err
}

// TestWorkflowPreparationIO rechecks provider and current policy after host Git
// IO rather than authorizing against its preflight snapshot.
func TestWorkflowPreparationIO(t *testing.T) {
	f := newWorkflowPreparationFixture(t, "merge", workflowstate.Validation{WaiverReason: "Preparation IO fixture"})

	git := workflowPreparationGitHook{git: f.git, after: func() { f.provider.revision = "changed-after-Git" }}
	if _, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, f.provider, git); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatalf("stale preparation: %v", err)
	}

	run := workflowTestRead(t, f.store, f.run.ID)
	if run.Preparation.State != "stale" || run.Status != "blocked" {
		t.Fatal("stale preparation was retained as valid")
	}

	assertWorkflowPreparationSource(t, f)
}
