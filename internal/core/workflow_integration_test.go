package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type workflowIntegrationFixture struct {
	store                               *WorkflowStore
	hierarchy                           WorkflowMaterializationResult
	run                                 WorkflowInvocation
	provider                            *workflowPRProviderFixture
	git                                 OSWorkflowIntegrationGit
	actor, remote, base, result, target string
}

func workflowTestRead(t *testing.T, store *WorkflowStore, id string) WorkflowInvocation {
	t.Helper()

	run, err := store.ReadWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	return run
}

func workflowDeliveryFixture(t *testing.T, method string) workflowIntegrationFixture {
	t.Helper()

	return workflowDeliveryExtentFixture(t, method, "slice")
}

func workflowDeliveryExtentFixture(t *testing.T, method, extent string) workflowIntegrationFixture {
	t.Helper()

	store, hierarchy, run, provider := workflowPRFixtureFor(t, extent)
	workflowPREnsure(t, store, run.ID, provider)

	run = workflowTestRead(t, store, run.ID)
	root := t.TempDir()
	remote, actor := filepath.Join(root, "canonical.git"), filepath.Join(root, "integrator")
	workflowIntegrationCommand(t, root, "init", "--bare", remote)
	workflowIntegrationCommand(t, run.Lane.Worktree, "push", remote, run.Lane.BaseCommit+":refs/heads/dev", run.Lane.HeadCommit+":"+run.Lane.Branch)
	workflowIntegrationCommand(t, root, "clone", "--no-checkout", remote, actor)
	workflowIntegrationCommand(t, actor, "config", "user.name", "Test")
	workflowIntegrationCommand(t, actor, "config", "user.email", "test@example.com")
	workflowIntegrationCommand(t, actor, "checkout", "-b", "dev", "origin/dev")

	base := workflowIntegrationCommit(t, actor, "target.txt", "Concurrent target change")
	workflowIntegrationCommand(t, actor, "branch", "source", "origin/"+provider.pr.HeadBranch)
	workflowFixtureIntegrate(t, actor, method)

	result := workflowIntegrationCommand(t, actor, "rev-parse", "HEAD")
	target := workflowIntegrationCommit(t, actor, "later.txt", "Later target change")
	workflowIntegrationCommand(t, actor, "push", "origin", "dev")

	provider.pr.State, provider.pr.Merged, provider.pr.ResultCommit = "closed", true, result
	// Forgejo exposes the current target here after merge, not the prepared base.
	provider.pr.BaseCommit = target

	return workflowIntegrationFixture{store, hierarchy, run, provider, NewWorkflowIntegrationGit(provider.pr.Repository, remote), actor, remote, base, result, target}
}

func workflowFixtureIntegrate(t *testing.T, actor, method string) {
	t.Helper()

	switch method {
	case "merge":
		workflowIntegrationCommand(t, actor, "merge", "--no-ff", "-m", "Integrate slice", "source")
	case "rebase":
		workflowIntegrationCommand(t, actor, "checkout", "-b", "rebased", "source")
		workflowIntegrationCommand(t, actor, "rebase", "dev")
		workflowIntegrationCommand(t, actor, "checkout", "dev")
		workflowIntegrationCommand(t, actor, "merge", "--ff-only", "rebased")
	case "squash":
		workflowIntegrationCommand(t, actor, "merge", "--squash", "source")
		workflowIntegrationCommand(t, actor, "commit", "-m", "Integrate slice")
	}
}

func workflowIntegrationCommand(t *testing.T, dir string, args ...string) string {
	t.Helper()

	output, err := outputWorktreeGit(t.Context(), dir, args...)
	if err != nil {
		t.Fatal(err)
	}

	return strings.TrimSpace(output)
}

func workflowIntegrationCommit(t *testing.T, dir, file, text string) string {
	t.Helper()

	if err := os.WriteFile(filepath.Join(dir, file), []byte(text+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	workflowIntegrationCommand(t, dir, "add", file)
	workflowIntegrationCommand(t, dir, "commit", "-m", text)

	return workflowIntegrationCommand(t, dir, "rev-parse", "HEAD")
}

// TestWorkflowIntegrationMethods uses actual bare canonical remotes, rewritten
// histories, a moved integration base and later target commits for all methods.
// Immutable proof is reusable without another worker, PR or task finalization.
func TestWorkflowIntegrationMethods(t *testing.T) {
	for _, method := range []string{"merge", "rebase", "squash"} {
		t.Run(method, func(t *testing.T) {
			fixture := workflowDeliveryFixture(t, method)

			proof, err := fixture.store.VerifyWorkflowSliceIntegration(t.Context(), fixture.run.ID, method, fixture.provider, fixture.git)
			if err != nil {
				t.Fatal(err)
			}

			checkWorkflowDeliveryProof(t, fixture, proof, method)

			restarted := NewWorkflowStore(fixture.store.layout, fixture.store.projectID)

			again, err := restarted.VerifyWorkflowSliceIntegration(t.Context(), fixture.run.ID, method, fixture.provider, fixture.git)
			if err != nil || !again.ConfirmedAt.Equal(proof.ConfirmedAt) || again.Git.HistoryHash != proof.Git.HistoryHash {
				t.Fatalf("immutable proof was not reused: %v", err)
			}
		})
	}
}

func checkWorkflowDeliveryProof(t *testing.T, fixture workflowIntegrationFixture, proof WorkflowSliceDeliveryProof, method string) {
	t.Helper()
	checkWorkflowDeliveryGit(t, fixture, proof.Git)

	expected := 1
	if method == "rebase" {
		expected = 3
	}

	if len(proof.Git.History) != expected || fixture.provider.creates != 1 {
		t.Fatal("method history or canonical PR count changed")
	}

	run := workflowTestRead(t, fixture.store, fixture.run.ID)
	if run.Status != "waiting_integration" || run.Outcome != "slice_delivery_verified" || run.DeliveryProofRef == "" || len(run.Attempts) != 4 {
		t.Fatal("verified delivery did not retain the finalization barrier")
	}

	checkWorkflowPRTasks(t, fixture.store, fixture.hierarchy.Tasks[:3])
}

func checkWorkflowDeliveryGit(t *testing.T, fixture workflowIntegrationFixture, proof WorkflowIntegrationProof) {
	t.Helper()

	if proof.BaseCommit != fixture.base || proof.ResultCommit != fixture.result || proof.ObservedTarget != fixture.target || proof.PreparedTree != proof.VerifiedTree || proof.SourceCommit != fixture.run.Lane.HeadCommit || proof.SourceDirectory != "" {
		t.Fatal("delivery proof differs from actual canonical Git or leaks local transport")
	}
}

// TestWorkflowIntegrationReject rejects worker completion claims, wrong provider
// identity, wrong actual target branches and results absent from that target.
func TestWorkflowIntegrationReject(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*testing.T, *workflowIntegrationFixture)
	}{
		{"worker claim", func(_ *testing.T, f *workflowIntegrationFixture) {
			f.provider.pr.Merged = false
			f.provider.pr.State = "open"
		}},
		{"wrong branch", func(_ *testing.T, f *workflowIntegrationFixture) { f.provider.pr.BaseBranch = "main" }},
		{"fork head", func(_ *testing.T, f *workflowIntegrationFixture) { f.provider.pr.HeadRepository += "-fork" }},
		{"wrong source", func(_ *testing.T, f *workflowIntegrationFixture) { f.provider.pr.HeadCommit = f.base }},
		{"source ancestry", func(_ *testing.T, f *workflowIntegrationFixture) { f.provider.pr.ResultCommit = f.run.Lane.HeadCommit }},
		{"wrong actual target", func(t *testing.T, f *workflowIntegrationFixture) {
			workflowIntegrationCommand(t, f.actor, "push", "--force", "origin", f.base+":refs/heads/dev")
		}},
		{"wrong repository", func(_ *testing.T, f *workflowIntegrationFixture) { f.git.repository += "-other" }},
		{"forged result tree", func(t *testing.T, f *workflowIntegrationFixture) {
			workflowIntegrationCommand(t, f.actor, "reset", "--hard", f.base)

			f.provider.pr.ResultCommit = workflowIntegrationCommit(t, f.actor, "forged.txt", "Not the source change")
			workflowIntegrationCommand(t, f.actor, "push", "--force", "origin", "dev")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := workflowDeliveryFixture(t, "rebase")
			tc.mutate(t, &fixture)

			if _, err := fixture.store.VerifyWorkflowSliceIntegration(t.Context(), fixture.run.ID, "rebase", fixture.provider, fixture.git); err == nil {
				t.Fatal("unverifiable canonical delivery was accepted")
			}

			run := workflowTestRead(t, fixture.store, fixture.run.ID)
			if run.DeliveryProofRef != "" || run.Status == "completed" {
				t.Fatal("failed verification created delivery proof or completed tasks")
			}

			checkWorkflowPRTasks(t, fixture.store, fixture.hierarchy.Tasks[:3])
		})
	}
}

// TestWorkflowIntegrationWrongMethod prevents implicit method fallback even
// when the provider merged the right source and its result is on the target.
func TestWorkflowIntegrationWrongMethod(t *testing.T) {
	for _, method := range []string{"merge", "rebase", "squash"} {
		t.Run(method, func(t *testing.T) {
			fixture := workflowDeliveryFixture(t, method)

			wrong := map[string]string{"merge": "squash", "rebase": "merge", "squash": "rebase"}[method]
			if _, err := fixture.store.VerifyWorkflowSliceIntegration(t.Context(), fixture.run.ID, wrong, fixture.provider, fixture.git); err == nil {
				t.Fatal("configured method silently fell back to actual provider method")
			}
		})
	}
}

type workflowIntegrationHook struct {
	git   WorkflowIntegrationGit
	after func()
}

func (hook workflowIntegrationHook) VerifyWorkflowIntegration(ctx context.Context, query WorkflowIntegrationQuery) (WorkflowIntegrationProof, error) {
	proof, err := hook.git.VerifyWorkflowIntegration(ctx, query)
	hook.after()

	return proof, err
}

// TestWorkflowIntegrationChangedInput rejects edits during trusted host IO and
// rejects later corruption of the immutable proof instead of replacing it.
func TestWorkflowIntegrationChangedInput(t *testing.T) {
	for _, stage := range []string{"during IO", "retained proof"} {
		t.Run(stage, func(t *testing.T) {
			fixture := workflowDeliveryFixture(t, "merge")
			checkWorkflowIntegrationChanged(t, fixture, stage)
		})
	}
}

func checkWorkflowIntegrationChanged(t *testing.T, fixture workflowIntegrationFixture, stage string) {
	t.Helper()

	var git WorkflowIntegrationGit = fixture.git
	if stage == "during IO" {
		git = workflowIntegrationHook{fixture.git, func() {
			workflowIntegrationCommit(t, fixture.run.Lane.Worktree, "changed.txt", "Changed candidate")
		}}
	} else {
		corruptWorkflowDeliveryProof(t, fixture)
	}

	if _, err := fixture.store.VerifyWorkflowSliceIntegration(t.Context(), fixture.run.ID, "merge", fixture.provider, git); err == nil {
		t.Fatal("changed candidate or immutable evidence was accepted")
	}
}

func corruptWorkflowDeliveryProof(t *testing.T, fixture workflowIntegrationFixture) {
	t.Helper()

	if _, err := fixture.store.VerifyWorkflowSliceIntegration(t.Context(), fixture.run.ID, "merge", fixture.provider, fixture.git); err != nil {
		t.Fatal(err)
	}

	run := workflowTestRead(t, fixture.store, fixture.run.ID)
	if err := os.WriteFile(filepath.Join(fixture.store.layout.Root(), run.DeliveryProofRef), []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}
