package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func nativeReplayFixture(t *testing.T) workflowGitFixture {
	t.Helper()

	f := newWorkflowGitFixture(t, "rebase")
	actor := f.request.Directory
	workflowIntegrationCommand(t, actor, "checkout", "-B", "dev", f.request.Git.BaseCommit)

	f.request.Git.BaseCommit = workflowIntegrationCommit(t, actor, "source.txt", "Source fixture")
	workflowIntegrationCommand(t, actor, "push", f.remote, "dev:refs/heads/dev")
	workflowIntegrationCommand(t, actor, "checkout", "source")

	source := f.request.Git.SourceCommit
	tree := f.request.Git.SourceTree
	empty := nativeOriginalCommit(t, actor, tree, source, "Empty Author", "978307200 +0530", "", "feat: record empty intent\n")
	workflowIntegrationCommand(t, actor, "reset", "--hard", empty)
	workflowIntegrationCommit(t, actor, "next.txt", "Remaining intent")

	tree = workflowIntegrationCommand(t, actor, "rev-parse", "HEAD^{tree}")
	f.request.Git.SourceCommit = nativeOriginalCommit(t, actor, tree, empty, "Next Author", "978307210 -0430", "ISO-8859-1", "feat: preserve caf\xe9 intent\n\nExact message without terminal newline")
	f.request.Git.SourceTree = tree
	workflowIntegrationCommand(t, actor, "reset", "--hard", f.request.Git.SourceCommit)
	workflowIntegrationCommand(t, actor, "push", f.remote, "HEAD:"+f.request.SourceRef)

	return f
}

func nativeOriginalCommit(t *testing.T, directory, tree, parent, author, date, encoding, message string) string {
	t.Helper()

	header := fmt.Sprintf("tree %s\nparent %s\nauthor %s <source@example.invalid> %s\ncommitter Source Host <source-host@example.invalid> 978307220 +0000\n", tree, parent, author, date)
	if encoding != "" {
		header += "encoding " + encoding + "\n"
	}

	reader := workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: directory}}

	id, err := reader.RawInput(t.Context(), header+"\n"+message, "hash-object", "-t", "commit", "-w", "--stdin")
	if err != nil {
		t.Fatal(err)
	}

	return strings.TrimSpace(id)
}

// TestNativeReplayProof retains duplicate and empty intent with byte-exact
// messages, dates and encoding, and atomically publishes the ordered history.
func TestNativeReplayProof(t *testing.T) {
	f := nativeReplayFixture(t)
	head := workflowIntegrationCommand(t, f.request.Directory, "rev-parse", "HEAD")
	proof, lane := nativeCombinedPreparation(t, f)

	p := proof.Provenance
	if p == nil || len(p.Replays) != 3 || len(p.Prepared) != 3 {
		t.Fatal("original replay sequence was dropped")
	}

	checkNativeReplayAttribution(t, proof)

	checkNativeReplayMessages(t, lane.Worktree, proof)

	if workflowIntegrationCommand(t, f.request.Directory, "rev-parse", "HEAD") != head || workflowIntegrationCommand(t, f.request.Directory, "status", "--porcelain") != "" {
		t.Fatal("preparation changed the source checkout")
	}

	publishNativeReplay(t, f, proof, lane.Worktree)
}

func checkNativeReplayMessages(t *testing.T, directory string, proof workflowstate.IntegrationProof) {
	t.Helper()

	reader := workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: directory}}
	for i, source := range proof.Provenance.Source {
		original, err := reader.RawOutput(t.Context(), "cat-file", "commit", source.Commit)
		if err != nil {
			t.Fatal(err)
		}

		prepared, err := reader.RawOutput(t.Context(), "cat-file", "commit", proof.Provenance.Prepared[i].Commit)
		if err != nil {
			t.Fatal(err)
		}

		_, originalMessage, _ := strings.Cut(original, "\n\n")
		_, preparedMessage, _ := strings.Cut(prepared, "\n\n")

		if originalMessage != preparedMessage {
			t.Fatal("raw original message bytes changed")
		}
	}
}

func publishNativeReplay(t *testing.T, f workflowGitFixture, proof workflowstate.IntegrationProof, directory string) {
	t.Helper()

	f.request.Git, f.request.Directory = proof, directory
	sealWorkflowGitFixture(t, &f.request)

	publisher := NewWorkflowGitPublisher(f.request.Repository, f.remote)

	result, err := publisher.PublishWorkflowGit(t.Context(), f.request)
	if err != nil || result.State != "applied" {
		t.Fatalf("retained replay publication: %+v, %v", result, err)
	}

	checkWorkflowGitRefs(t, f, proof.ResultCommit, proof.ResultCommit)

	observer := NewWorkflowIntegrationGit(proof.Repository, f.remote)
	query := proof.IntegrationQuery
	query.SourceDirectory = directory

	observed, err := observer.VerifyPreparedWorkflowIntegration(t.Context(), query, proof)
	if err != nil || observed.HistoryHash != proof.HistoryHash || observed.ObservedTarget != proof.ResultCommit {
		t.Fatal("published target lost the retained replay or repair proof", err)
	}
}

// TestNativeReplayGuard rejects false mapping and metadata even when enclosing
// hashes are recomputed, before either canonical ref is changed.
func TestNativeReplayGuard(t *testing.T) {
	cases := []struct {
		name string
		edit func(*workflowstate.IntegrationProvenance)
	}{
		{"author", func(p *workflowstate.IntegrationProvenance) { p.Source[0].Author.Name = "Forged Author" }},
		{"message", func(p *workflowstate.IntegrationProvenance) {
			p.Prepared[0].MessageHash = contenthash.Sum([]byte("forged"))
		}},
		{"date", func(p *workflowstate.IntegrationProvenance) { p.Prepared[0].Author.Date = "978307200 +0000" }},
		{"empty", func(p *workflowstate.IntegrationProvenance) { p.Replays[0].EmptyDecision = "source_empty" }},
		{"order", func(p *workflowstate.IntegrationProvenance) { p.Replays[0], p.Replays[1] = p.Replays[1], p.Replays[0] }},
		{"missing", func(p *workflowstate.IntegrationProvenance) { p.Prepared = p.Prepared[1:] }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := nativeReplayFixture(t)
			proof, lane := nativeCombinedPreparation(t, f)
			tc.edit(proof.Provenance)
			rejectNativeReplay(t, f, proof, lane.Worktree)
		})
	}
}

func rejectNativeReplay(t *testing.T, f workflowGitFixture, proof workflowstate.IntegrationProof, directory string) {
	t.Helper()

	var err error

	proof.HistoryHash, err = workflowIntegrationHistoryHash(proof)
	if err != nil {
		t.Fatal(err)
	}

	f.request.Git, f.request.Directory = proof, directory
	sealWorkflowGitFixture(t, &f.request)

	publisher := NewWorkflowGitPublisher(f.request.Repository, f.remote)

	result, err := publisher.PublishWorkflowGit(t.Context(), f.request)
	if err == nil || result.CommandsSent {
		t.Fatal("false replay proof reached receive-pack")
	}

	checkWorkflowGitRefs(t, f, proof.SourceCommit, proof.BaseCommit)
}

func nativeReplayFix(t *testing.T, proof workflowstate.IntegrationProof, lane workflowstate.Lane) workflowstate.IntegrationProof {
	t.Helper()

	if err := os.WriteFile(filepath.Join(lane.Worktree, "repair.txt"), []byte("Required integration repair\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	workflowIntegrationCommand(t, lane.Worktree, "add", "repair.txt")

	tree := workflowIntegrationCommand(t, lane.Worktree, "write-tree")
	reader := workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: lane.Worktree}}
	finding := workflowstate.ReviewFinding{Category: workflowstate.AgenticReviewFindingCorrectness, Title: "Retain integration output", Detail: "The combined validation requires this repair output."}

	fixed, err := reader.prepareRebaseFix(t.Context(), proof, tree, finding)
	if err != nil {
		t.Fatal(err)
	}

	workflowIntegrationCommand(t, lane.Worktree, "reset", "--hard", fixed.ResultCommit)

	return fixed
}

// TestNativeReplayFix publishes a separately attributed repair without erasing
// original replay mappings, and rejects hidden or cosmetic repairs.
func TestNativeReplayFix(t *testing.T) {
	for _, mode := range []string{"publish", "undeclared", "cosmetic", "parent"} {
		t.Run(mode, func(t *testing.T) {
			f := nativeReplayFixture(t)
			proof, lane := nativeCombinedPreparation(t, f)
			fixed := nativeReplayFix(t, proof, lane)
			checkNativeReplayFix(t, proof, fixed)

			p := fixed.Provenance

			switch mode {
			case "publish":
				publishNativeReplay(t, f, fixed, lane.Worktree)
			case "undeclared":
				p.IntegrationFixes = nil
			case "cosmetic":
				p.IntegrationFixes[0].Finding.Category = workflowstate.AgenticReviewFindingCosmetic
			case "parent":
				p.Prepared[3].Parents[0] = fixed.BaseCommit
			}

			if mode != "publish" {
				rejectNativeReplay(t, f, fixed, lane.Worktree)
			}
		})
	}
}

// TestNativeReplayBlocked keeps unsupported merge histories and wholly empty
// integration units out of preparation without changing source or target refs.
func TestNativeReplayBlocked(t *testing.T) {
	for _, reason := range []string{"merge", "already_applied", "empty"} {
		t.Run(reason, func(t *testing.T) {
			f := newWorkflowGitFixture(t, "rebase")
			actor := f.request.Directory

			switch reason {
			case "merge":
				extendNativeCombinedSource(t, &f)
			case "already_applied":
				workflowIntegrationCommand(t, actor, "checkout", "-B", "dev", f.request.Git.BaseCommit)

				f.request.Git.BaseCommit = workflowIntegrationCommit(t, actor, "source.txt", "Source fixture")
				workflowIntegrationCommand(t, actor, "push", f.remote, "dev:refs/heads/dev")
			case "empty":
				tree := workflowIntegrationCommand(t, actor, "rev-parse", f.request.Git.OriginCommit+"^{tree}")
				f.request.Git.SourceCommit = nativeOriginalCommit(t, actor, tree, f.request.Git.OriginCommit, "Empty Author", "978307200 +0000", "", "feat: retain empty intent\n")
				f.request.Git.SourceTree = tree
				workflowIntegrationCommand(t, f.remote, "fetch", "--quiet", "--no-tags", actor, f.request.Git.SourceCommit)
				workflowIntegrationCommand(t, f.remote, "update-ref", f.request.SourceRef, f.request.Git.SourceCommit)
			}

			rejectNativePreparation(t, f)
		})
	}
}

func rejectNativePreparation(t *testing.T, f workflowGitFixture) {
	t.Helper()

	head := workflowIntegrationCommand(t, f.request.Directory, "rev-parse", "HEAD")
	q := PreparationQuery{IntegrationQuery: f.request.Git.IntegrationQuery, BaseCommit: f.request.Git.BaseCommit}
	q.SourceDirectory = f.request.Directory
	q.Lane = workflowstate.Lane{Unit: workflowstate.Ref{Kind: "slice", ID: "fixture"}, Branch: "refs/heads/chr/workflow/rejected", Worktree: filepath.Join(t.TempDir(), "rejected"), BaseCommit: q.BaseCommit, HeadCommit: q.BaseCommit}
	git := NewWorkflowPreparationGit(q.Repository, f.remote)

	if _, err := git.PrepareWorkflowIntegration(t.Context(), q); err == nil {
		t.Fatal("unsupported or empty integration unit was prepared")
	}

	if _, err := os.Lstat(q.Lane.Worktree); !os.IsNotExist(err) || workflowIntegrationCommand(t, f.request.Directory, "rev-parse", "HEAD") != head {
		t.Fatal("rejected preparation changed the caller or created a lane")
	}

	checkWorkflowGitRefs(t, f, q.SourceCommit, q.BaseCommit)
}

func checkNativeReplayAttribution(t *testing.T, proof workflowstate.IntegrationProof) {
	t.Helper()

	p := proof.Provenance

	for i, decision := range []string{"already_applied", "source_empty", "nonempty"} {
		if p.Replays[i].EmptyDecision != decision || p.Source[i].Author != p.Prepared[i].Author || p.Source[i].MessageHash != p.Prepared[i].MessageHash || p.Source[i].Encoding != p.Prepared[i].Encoding || p.Prepared[i].Committer.Name != "Workflow Host" {
			t.Fatal("replay intent or integrator attribution changed")
		}
	}
}

func checkNativeReplayFix(t *testing.T, original, fixed workflowstate.IntegrationProof) {
	t.Helper()

	p := fixed.Provenance
	if len(fixed.History) != 3 || len(p.Prepared) != 4 || len(p.IntegrationFixes) != 1 || p.Prepared[3].Author.Name != "Workflow Host" || p.Prepared[3].Parents[0] != original.ResultCommit {
		t.Fatal("repair erased source history or separate attribution")
	}
}

// TestNativeReplayObservation retains new replay and repair evidence through
// both forge adapters while preserving legacy publication observation behavior.
func TestNativeReplayObservation(t *testing.T) {
	for _, forge := range []string{"forgejo", "github"} {
		t.Run(forge, func(t *testing.T) {
			for _, mode := range []string{"replay", "repair"} {
				t.Run(mode, func(t *testing.T) {
					fixture := newNativeCapabilityFixture(t, forge)
					f := workflowGitFixture{request: fixture.request.Publication, remote: fixture.remote}

					proof, lane := nativeCombinedPreparation(t, f)
					if mode == "repair" {
						proof = nativeReplayFix(t, proof, lane)
					}

					fixture.request.Publication.Git, fixture.request.Publication.Directory = proof, lane.Worktree
					sealWorkflowGitFixture(t, &fixture.request.Publication)

					binding := nativeObservationFixture(t, fixture)
					checkNativeObservationCases(t, fixture, binding)
				})
			}
		})
	}
}
