package core

import (
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

// TestNativeMethodObjects rejects actual prepared objects with false combined
// trees even when their attribution, history and candidate hashes agree.
func TestNativeMethodObjects(t *testing.T) {
	for _, method := range []string{"merge", "rebase", "squash"} {
		t.Run(method, func(t *testing.T) {
			f := newWorkflowGitFixture(t, method)
			proof, lane := nativeCombinedPreparation(t, f)
			proof = nativeAlteredTree(t, lane.Worktree, proof)
			rejectNativeReplay(t, f, proof, lane.Worktree)
		})
	}
}

func nativeAlteredTree(t *testing.T, directory string, proof WorkflowIntegrationProof) WorkflowIntegrationProof {
	t.Helper()

	entry := proof.Provenance.Prepared[0]
	reader := workflowIntegrationReader{dir: directory, preparation: proof.Provenance.Preparation}

	raw, err := reader.rawOutput(t.Context(), "cat-file", "commit", entry.Commit)
	if err != nil {
		t.Fatal(err)
	}

	_, message, _ := strings.Cut(raw, "\n\n")

	commit, err := reader.preparationObject(t.Context(), proof.SourceTree, entry.Parents, entry.Author, entry.Encoding, message)
	if err != nil {
		t.Fatal(err)
	}

	proof.ResultCommit, proof.PreparedTree, proof.VerifiedTree = commit, proof.SourceTree, proof.SourceTree
	proof.History[0].Commit, proof.History[0].Tree = commit, proof.SourceTree

	proof.Provenance, err = reader.integrationProvenance(t.Context(), proof)
	if err != nil {
		t.Fatal(err)
	}

	proof.HistoryHash, err = workflowIntegrationHistoryHash(proof)
	if err != nil || validateWorkflowIntegrationProof(proof.WorkflowIntegrationQuery, proof) != nil {
		t.Fatal("altered-object fixture did not carry a structurally consistent proof", err)
	}

	return proof
}

// TestNativeMethodMetadata rejects forged prepared dates, signature hashes and
// source trees across all methods, including recomputed preparation identities.
func TestNativeMethodMetadata(t *testing.T) {
	cases := []struct {
		name string
		edit func(*WorkflowIntegrationProof)
	}{
		{"source_tree", func(p *WorkflowIntegrationProof) { p.SourceTree = p.PreparedTree }},
		{"date", nativeAlteredPreparationDate},
		{"signature", func(p *WorkflowIntegrationProof) {
			p.Provenance.Preparation.Signatures[0].SignatureHash = contenthash.Sum([]byte("forged signature"))
		}},
	}

	for _, method := range []string{"merge", "rebase", "squash"} {
		t.Run(method, func(t *testing.T) {
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					f := newWorkflowGitFixture(t, method)
					q := nativePolicyQuery(t, f)
					signer, profile := nativeSigningFixture(t)
					q.RequiredSignatures, q.SigningProfile = true, profile.Hash
					git := NewWorkflowPreparationGit(q.Repository, f.remote).WithSSHCommitSigning(signer)

					proof, err := git.PrepareWorkflowIntegration(t.Context(), q)
					if err != nil {
						t.Fatal(err)
					}

					tc.edit(&proof)
					rejectNativeReplay(t, f, proof, q.Lane.Worktree)
				})
			}
		})
	}
}

func nativeAlteredPreparationDate(proof *WorkflowIntegrationProof) {
	p := proof.Provenance.Preparation
	p.InitialCommitterDate, p.CommitterDate = "1700000000 +0000", "1700000000 +0000"

	p.Identity = workflowPreparationIdentity(*proof, *p)
	for i := range proof.Provenance.Prepared {
		proof.Provenance.Prepared[i].Committer.Date = p.CommitterDate
	}
}

// TestNativeMethodNoOp rejects source-equal and base-equal publication claims
// for every configured method before either canonical ref receives a command.
func TestNativeMethodNoOp(t *testing.T) {
	for _, method := range []string{"merge", "rebase", "squash"} {
		t.Run(method, func(t *testing.T) {
			for _, identity := range []string{"source", "base"} {
				t.Run(identity, func(t *testing.T) {
					f := newWorkflowGitFixture(t, method)
					proof, lane := nativeCombinedPreparation(t, f)

					proof.ResultCommit = proof.SourceCommit
					if identity == "base" {
						proof.ResultCommit = proof.BaseCommit
					}

					rejectNativeReplay(t, f, proof, lane.Worktree)
				})
			}
		})
	}
}
