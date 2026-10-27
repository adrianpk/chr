package core

import (
	"encoding/json"
	"path/filepath"
	"slices"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

func nativeCombinedPreparation(t *testing.T, f workflowGitFixture) (WorkflowIntegrationProof, WorkflowLane) {
	t.Helper()

	q := WorkflowPreparationQuery{WorkflowIntegrationQuery: f.request.Git.WorkflowIntegrationQuery, BaseCommit: f.request.Git.BaseCommit}
	q.SourceDirectory = f.request.Directory
	q.Lane = WorkflowLane{Unit: WorkflowRef{Kind: "slice", ID: "fixture"}, Branch: "refs/heads/chr/workflow/prepared", Worktree: filepath.Join(t.TempDir(), "combined"), BaseCommit: q.BaseCommit, HeadCommit: q.BaseCommit}
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

func checkNativeCombinedObjects(t *testing.T, directory string, proof WorkflowIntegrationProof) {
	t.Helper()

	reader := workflowIntegrationReader{dir: directory}
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
		change func(*WorkflowIntegrationProvenance)
	}{
		{"author", func(p *WorkflowIntegrationProvenance) { p.Source[0].Author.Name = "Forged Author" }},
		{"date", func(p *WorkflowIntegrationProvenance) { p.Source[0].Author.Date = "1700000000 +0000" }},
		{"message", func(p *WorkflowIntegrationProvenance) { p.Source[0].MessageHash = contenthash.Sum([]byte("forged")) }},
		{"diff", func(p *WorkflowIntegrationProvenance) { p.CombinedDiffHash = contenthash.Sum([]byte("forged")) }},
		{"committer", func(p *WorkflowIntegrationProvenance) { p.Prepared[0].Committer.Name = "Forged Host" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newWorkflowGitFixture(t, "squash")
			proof, lane := nativeCombinedPreparation(t, f)
			tc.change(proof.Provenance)

			var err error

			proof.HistoryHash, err = workflowIntegrationHistoryHash(proof)
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
		History                    []WorkflowIntegrationCommit
	}{proof.Method, proof.SourceCommit, proof.BaseCommit, proof.PreparedTree, proof.History})
	if err != nil || contenthash.Sum(data) != proof.HistoryHash || validateWorkflowIntegrationProof(proof.WorkflowIntegrationQuery, proof) != nil {
		t.Fatal("historical proof hash or read contract changed")
	}
}
