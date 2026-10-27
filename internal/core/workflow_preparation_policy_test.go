package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func nativePolicyQuery(t *testing.T, f workflowGitFixture) WorkflowPreparationQuery {
	t.Helper()

	q := WorkflowPreparationQuery{WorkflowIntegrationQuery: f.request.Git.WorkflowIntegrationQuery, BaseCommit: f.request.Git.BaseCommit}
	q.SourceDirectory = f.request.Directory
	q.Lane = WorkflowLane{Unit: workflowstate.Ref{Kind: "slice", ID: "fixture"}, Branch: "refs/heads/chr/workflow/prepared", Worktree: filepath.Join(t.TempDir(), "prepared"), BaseCommit: q.BaseCommit, HeadCommit: q.BaseCommit}

	return q
}

func nativeSigningFixture(t *testing.T) (WorkflowSSHCommitSigning, WorkflowCommitSigningProfile) {
	t.Helper()

	key := filepath.Join(t.TempDir(), "signing")
	if _, err := workflowCommitSSHCommand(t.Context(), "", "-q", "-t", "ed25519", "-N", "", "-C", "", "-f", key); err != nil {
		t.Fatal(err)
	}

	public, err := os.ReadFile(key + ".pub")
	if err != nil {
		t.Fatal(err)
	}

	signing := WorkflowSSHCommitSigning{KeyFile: key, PublicKey: strings.TrimSpace(string(public)), Principal: "workflow@example.invalid"}

	profile, err := signing.Profile()
	if err != nil {
		t.Fatal(err)
	}

	return signing, profile
}

// TestNativePreparationSigning verifies every rewritten object using both the
// sealed host profile and Git's independent SSH verifier before publication.
func TestNativePreparationSigning(t *testing.T) {
	for _, method := range []string{"merge", "rebase", "squash"} {
		t.Run(method, func(t *testing.T) {
			f := newWorkflowGitFixture(t, method)
			if method == "rebase" {
				f = nativeReplayFixture(t)
			}

			q := nativePolicyQuery(t, f)
			signing, profile := nativeSigningFixture(t)
			q.RequiredSignatures, q.SigningProfile = true, profile.Hash
			git := NewWorkflowPreparationGit(q.Repository, f.remote).WithSSHCommitSigning(signing)

			proof, err := git.PrepareWorkflowIntegration(t.Context(), q)
			if err != nil || checkWorkflowPreparedProof(q, proof) != nil {
				t.Fatal("certified signed preparation failed", err)
			}

			checkNativeSigningObjects(t, q.Lane.Worktree, proof, profile)

			if method == "rebase" {
				checkNativeReplayMessages(t, q.Lane.Worktree, proof)
			}

			publishNativeReplay(t, f, proof, q.Lane.Worktree)
		})
	}
}

func checkNativeSigningObjects(t *testing.T, directory string, proof WorkflowIntegrationProof, profile WorkflowCommitSigningProfile) {
	t.Helper()

	p := proof.Provenance.Preparation
	if p == nil || !p.RequiredSignatures || len(p.Signatures) != len(proof.Provenance.Prepared) || p.Signing.Hash != profile.Hash {
		t.Fatal("signature evidence does not cover the entire prepared sequence")
	}

	allowed := filepath.Join(t.TempDir(), "allowed")
	if err := os.WriteFile(allowed, []byte(profile.Principal+" "+profile.PublicKey+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, entry := range proof.Provenance.Prepared {
		workflowIntegrationCommand(t, directory, "-c", "gpg.format=ssh", "-c", "gpg.ssh.allowedSignersFile="+allowed, "verify-commit", entry.Commit)
	}
}

// TestNativeSignedFix preserves the certified signature policy when an
// actionable repair extends the prepared sequence, rejecting an absent signer.
func TestNativeSignedFix(t *testing.T) {
	f := nativeReplayFixture(t)
	q := nativePolicyQuery(t, f)
	signing, profile := nativeSigningFixture(t)
	q.RequiredSignatures, q.SigningProfile = true, profile.Hash
	git := NewWorkflowPreparationGit(q.Repository, f.remote).WithSSHCommitSigning(signing)

	proof, err := git.PrepareWorkflowIntegration(t.Context(), q)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(q.Lane.Worktree, "repair.txt"), []byte("Required repair\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	workflowIntegrationCommand(t, q.Lane.Worktree, "add", "repair.txt")

	tree := workflowIntegrationCommand(t, q.Lane.Worktree, "write-tree")
	finding := workflowstate.ReviewFinding{Category: AgenticReviewFindingCorrectness, Title: "Retain repaired behavior", Detail: "The prepared candidate requires this declared integration repair."}

	reader := workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: q.Lane.Worktree}}
	if _, err := reader.prepareRebaseFix(t.Context(), proof, tree, finding); err == nil {
		t.Fatal("integration repair bypassed the required host signer")
	}

	reader.signing = &signing

	fixed, err := reader.prepareRebaseFix(t.Context(), proof, tree, finding)
	if err != nil || len(proof.Provenance.Prepared) != 3 || len(fixed.Provenance.Prepared) != 4 {
		t.Fatal("signed repair lost the original replay evidence", err)
	}

	workflowIntegrationCommand(t, q.Lane.Worktree, "reset", "--hard", fixed.ResultCommit)
	checkNativeSigningObjects(t, q.Lane.Worktree, fixed, profile)
	publishNativeReplay(t, f, fixed, q.Lane.Worktree)
}

// TestNativeSigningBlocked proves missing, mismatched and unusable host keys
// block before a validation lane is attached or a canonical ref is mutated.
func TestNativeSigningBlocked(t *testing.T) {
	for _, name := range []string{"missing", "profile", "key", "unsigned"} {
		t.Run(name, func(t *testing.T) {
			f := newWorkflowGitFixture(t, "rebase")
			q := nativePolicyQuery(t, f)
			signing, profile := nativeSigningFixture(t)
			q.RequiredSignatures, q.SigningProfile = true, profile.Hash
			git := NewWorkflowPreparationGit(q.Repository, f.remote)

			switch name {
			case "profile":
				q.SigningProfile = contenthash.Sum([]byte("other policy"))
				git = git.WithSSHCommitSigning(signing)
			case "key":
				signing.KeyFile = filepath.Join(t.TempDir(), "absent")
				git = git.WithSSHCommitSigning(signing)
			case "unsigned":
				proof, _ := nativeCombinedPreparation(t, f)
				if checkWorkflowPreparedProof(q, proof) == nil {
					t.Fatal("unsigned proof satisfied required signature policy")
				}

				return
			}

			if _, err := git.PrepareWorkflowIntegration(t.Context(), q); err == nil {
				t.Fatal("unsupported signing policy was accepted")
			}

			if _, err := os.Lstat(q.Lane.Worktree); !os.IsNotExist(err) {
				t.Fatal("failed signing attached a validation lane")
			}

			checkWorkflowGitRefs(t, f, q.SourceCommit, q.BaseCommit)
		})
	}
}

// TestNativePreparationCollision reproduces P == S with actual source metadata
// and proves a timestamp retry preserves the original intent and commit count.
func TestNativePreparationCollision(t *testing.T) {
	f := newWorkflowGitFixture(t, "rebase")
	q := nativePolicyQuery(t, f)
	reader := workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: f.request.Directory}}

	source, err := reader.commitAttribution(t.Context(), q.SourceCommit)
	if err != nil {
		t.Fatal(err)
	}

	seconds := int64(1700000000)
	data := fmt.Sprintf("tree %s\nparent %s\nauthor %s <%s> %s\ncommitter Workflow Host <workflow@example.invalid> %d +0000\n\n", q.SourceTree, q.OriginCommit, source.Author.Name, source.Author.Email, source.Author.Date, seconds)

	raw, err := reader.RawOutput(t.Context(), "cat-file", "commit", q.SourceCommit)
	if err != nil {
		t.Fatal(err)
	}

	_, message, _ := strings.Cut(raw, "\n\n")

	id, err := reader.RawInput(t.Context(), data+message, "hash-object", "-t", "commit", "-w", "--stdin")
	if err != nil {
		t.Fatal(err)
	}

	q.SourceCommit, q.BaseCommit = strings.TrimSpace(id), q.OriginCommit
	q.Lane.BaseCommit, q.Lane.HeadCommit = q.BaseCommit, q.BaseCommit
	workflowIntegrationCommand(t, f.request.Directory, "push", "--force", f.remote, q.SourceCommit+":"+f.request.SourceRef, q.BaseCommit+":refs/heads/dev")

	git := NewWorkflowPreparationGit(q.Repository, f.remote)
	git.committerTime = time.Unix(seconds, 0)

	proof, err := git.PrepareWorkflowIntegration(t.Context(), q)
	if err != nil {
		t.Fatal(err)
	}

	p := proof.Provenance.Preparation
	if proof.ResultCommit == q.SourceCommit || len(proof.History) != 1 || p.NoOpCollisions != 1 || p.CommitterDate != "1700000001 +0000" || proof.Provenance.Prepared[0].Author != source.Author {
		t.Fatal("collision retry modified intent or inserted an artificial commit")
	}

	checkNativeReplayMessages(t, q.Lane.Worktree, proof)
}

// TestNativeCollisionBudget checks the exact initial build plus three retry
// bound and ensures an exhausted or failed build never returns a candidate.
func TestNativeCollisionBudget(t *testing.T) {
	cases := []struct {
		name       string
		calls      int
		collisions int
		result     string
		date       string
	}{
		{"retry", 4, 3, "prepared", "1700000003 +0000"},
		{"exhausted", 4, 3, "", "1700000003 +0000"},
		{"failure", 1, 0, "", "1700000000 +0000"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := &WorkflowIntegrationPreparation{InitialCommitterDate: "1700000000 +0000"}
			calls := 0
			build := func() (string, error) {
				calls++

				return nativeCollisionBuild(tc.name, calls)
			}

			result, err := workflowDistinctPreparation("source", p, build)
			if calls != tc.calls || p.NoOpCollisions != tc.collisions || p.CommitterDate != tc.date || result != tc.result || (tc.result != "") != (err == nil) {
				t.Fatal("collision bound or recorded metadata changed")
			}
		})
	}
}

func nativeCollisionBuild(name string, calls int) (string, error) {
	if name == "failure" {
		return "", workflowstate.ErrConflict
	}

	if name == "retry" && calls == 4 {
		return "prepared", nil
	}

	return "source", nil
}
