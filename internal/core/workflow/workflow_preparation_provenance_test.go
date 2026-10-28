package workflow

import (
	"encoding/json"
	"path/filepath"
	"slices"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func nativeCombinedPreparation(t *testing.T, f workflowGitFixture) (workflowstate.IntegrationProof, workflowstate.Lane) {
	t.Helper()

	q := workflowstate.PreparationQuery{IntegrationQuery: f.request.Git.IntegrationQuery, BaseCommit: f.request.Git.BaseCommit}
	q.SourceDirectory = f.request.Directory
	q.Lane = workflowstate.Lane{Unit: workflowstate.Ref{Kind: "slice", ID: "fixture"}, Branch: "refs/heads/chr/workflow/prepared", Worktree: filepath.Join(t.TempDir(), "combined"), BaseCommit: q.BaseCommit, HeadCommit: q.BaseCommit}
	git := NewWorkflowPreparationGit(q.Repository, f.remote)

	proof, err := git.PrepareWorkflowIntegration(t.Context(), q)
	if err != nil {
		t.Fatal(err)
	}

	return proof, q.Lane
}

// TestNativeCombinedProof checks real parent objects and complete source DAG
// attribution without altering the caller's checkout or canonical refs.
func TestNativeCombinedProof(t *testing.T) {
	for _, method := range []string{"merge", "squash"} {
		t.Run(method, func(t *testing.T) {
			f := newWorkflowGitFixture(t, method)
			extendNativeCombinedSource(t, &f)

			head := workflowIntegrationCommand(t, f.request.Directory, "rev-parse", "HEAD")
			proof, lane := nativeCombinedPreparation(t, f)

			p := proof.Provenance
			if p == nil || len(p.Source) != 4 || len(p.Prepared) != 1 || p.Source[3].Commit != proof.SourceCommit {
				t.Fatalf("incomplete source attribution: %+v", p)
			}

			checkNativeCombinedObjects(t, lane.Worktree, proof)
			checkWorkflowGitRefs(t, f, proof.SourceCommit, proof.BaseCommit)

			if workflowIntegrationCommand(t, f.request.Directory, "rev-parse", "HEAD") != head || workflowIntegrationCommand(t, f.request.Directory, "status", "--porcelain") != "" {
				t.Fatal("preparation changed the user checkout")
			}
		})
	}
}

func extendNativeCombinedSource(t *testing.T, f *workflowGitFixture) {
	t.Helper()

	actor := f.request.Directory
	workflowIntegrationCommand(t, actor, "checkout", "-b", "side", f.request.Git.OriginCommit)
	workflowIntegrationCommit(t, actor, "side.txt", "Side intent")
	workflowIntegrationCommand(t, actor, "checkout", "source")
	workflowIntegrationCommand(t, actor, "config", "user.name", "Other Author")
	workflowIntegrationCommand(t, actor, "config", "user.email", "other@example.invalid")
	workflowIntegrationCommit(t, actor, "next.txt", "Next intent")
	workflowIntegrationCommand(t, actor, "merge", "--no-ff", "-m", "test(integration): retain side intent", "side")

	f.request.Git.SourceCommit = workflowIntegrationCommand(t, actor, "rev-parse", "HEAD")
	f.request.Git.SourceTree = workflowIntegrationCommand(t, actor, "rev-parse", "HEAD^{tree}")
	workflowIntegrationCommand(t, actor, "push", f.remote, "HEAD:"+f.request.SourceRef)
}

func checkNativeCombinedObjects(t *testing.T, directory string, proof workflowstate.IntegrationProof) {
	t.Helper()

	reader := workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: directory}}
	parents, err := reader.parents(t.Context(), proof.ResultCommit)

	expected := []string{proof.BaseCommit}
	if proof.Method == "merge" {
		expected = append(expected, proof.SourceCommit)
	}

	if err != nil || !slices.Equal(parents, expected) || !slices.Equal(proof.Provenance.Prepared[0].Parents, expected) {
		t.Fatal("result parents differ from the exact configured method")
	}

	for _, source := range proof.Provenance.Source {
		actual, err := reader.commitAttribution(t.Context(), source.Commit)
		if err != nil || !workflowNativeJSONEqual(actual, source) {
			t.Fatal("original author, date, message or parent proof differs from Git")
		}
	}
}

// TestNativeCombinedGuard rejects recomputed hashes concealing false provenance
// before receive-pack can mutate either source or target.
func TestNativeCombinedGuard(t *testing.T) {
	cases := []struct {
		name   string
		change func(*workflowstate.IntegrationProvenance)
	}{
		{"author", func(p *workflowstate.IntegrationProvenance) { p.Source[0].Author.Name = "Forged Author" }},
		{"date", func(p *workflowstate.IntegrationProvenance) { p.Source[0].Author.Date = "1700000000 +0000" }},
		{"message", func(p *workflowstate.IntegrationProvenance) {
			p.Source[0].MessageHash = contenthash.Sum([]byte("forged"))
		}},
		{"diff", func(p *workflowstate.IntegrationProvenance) { p.CombinedDiffHash = contenthash.Sum([]byte("forged")) }},
		{"committer", func(p *workflowstate.IntegrationProvenance) { p.Prepared[0].Committer.Name = "Forged Host" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newWorkflowGitFixture(t, "squash")
			proof, lane := nativeCombinedPreparation(t, f)
			tc.change(proof.Provenance)

			var err error

			proof.HistoryHash, err = workflowintegrate.WorkflowIntegrationHistoryHash(proof)
			if err != nil {
				t.Fatal(err)
			}

			f.request.Git, f.request.Directory = proof, lane.Worktree
			sealWorkflowGitFixture(t, &f.request)

			publisher := NewWorkflowGitPublisher(f.request.Repository, f.remote)

			result, err := publisher.PublishWorkflowGit(t.Context(), f.request)
			if err == nil || result.CommandsSent {
				t.Fatal("false source attribution authorized publication")
			}

			checkWorkflowGitRefs(t, f, proof.SourceCommit, proof.BaseCommit)
		})
	}
}

// TestNativeCombinedLegacy keeps the original hash representation available for
// immutable historical receipts while new preparation carries richer evidence.
func TestNativeCombinedLegacy(t *testing.T) {
	f := newWorkflowGitFixture(t, "merge")
	proof := f.request.Git

	data, err := json.Marshal(struct {
		Method, Source, Base, Tree string
		History                    []workflowstate.IntegrationCommit
	}{proof.Method, proof.SourceCommit, proof.BaseCommit, proof.PreparedTree, proof.History})
	if err != nil || contenthash.Sum(data) != proof.HistoryHash || workflowintegrate.ValidateWorkflowIntegrationProof(proof.IntegrationQuery, proof, workflowIntegrationRepairs{}) != nil {
		t.Fatal("historical proof hash or read contract changed")
	}
}
