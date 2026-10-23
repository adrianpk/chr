package app

import (
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

// Keep concrete temporary-fixture checkpoint and native evidence identities in
// verbose acceptance output without exposing host paths or raw worker state.
func assertGuidedTUIEvidence(t *testing.T, session core.AuthoringSession, result guidedResultExport) {
	t.Helper()

	if len(result.Deliveries) == 0 || len(result.Closures) == 0 || len(result.Validations) == 0 || len(result.Ideas) != 1 {
		t.Fatal("complete export lacks original receipt, closing, validation or idea evidence")
	}

	for _, checkpoint := range session.Guided.Checkpoints {
		t.Logf("checkpoint stage=%s commit=%s", checkpoint.Stage, checkpoint.Commit)

		for _, artifact := range checkpoint.Artifacts {
			t.Logf("input path=%s hash=%s", artifact.Path, artifact.Hash)
		}
	}

	for _, delivery := range result.Deliveries {
		receipt := delivery.Receipt
		t.Logf("delivery candidate=%s submission=%s result=%s target=%s evidence_hash=%s", receipt.CandidateID, receipt.SubmissionID, receipt.ResultCommit, delivery.Proof.Git.TargetBranch, delivery.Reference.Hash)
	}

	for _, closure := range result.Closures {
		t.Logf("closure outcome=%s evidence_hash=%s", closure.Receipt.Outcome, closure.Reference.Hash)
	}

	for _, validation := range result.Validations {
		t.Logf("validation outcome=%s evidence_hash=%s", validation.Outcome, validation.Reference.Hash)
	}
}
