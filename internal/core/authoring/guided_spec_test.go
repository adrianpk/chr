package authoring

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

// TestGuidedUseCaseSpecAuthority binds default approval to the intended run and policy.
func TestGuidedUseCaseSpecAuthority(t *testing.T) {
	for _, variant := range []string{"valid", "actor", "policy", "project", "namespace", "run", "body"} {
		t.Run(variant, func(t *testing.T) {
			approval := &GuidedSpecApproval{RunID: "run", Actor: "operator", Policy: GuidedSpecApprovalPolicy, InputDigest: "inputs"}
			request := SpecMutationRequest{GuidedApproval: approval, ProjectID: "project", Reference: "ops/default/spec/spec/artifact.md", Title: "Generated", Body: "Exact generated body"}
			session := Session{ID: "run", ProjectID: "project", Namespace: "default", Guided: &GuidedUseCaseState{Stage: "specification", Actor: "operator", ProgressionIntent: true, PendingOperation: &GuidedUseCaseOperation{Action: "spec", PreparedSpec: &GuidedPreparedSpec{Title: request.Title, Body: request.Body, InputDigest: "inputs"}}}}

			switch variant {
			case "actor":
				approval.Actor = "other"
			case "policy":
				approval.Policy = "other"
			case "project":
				request.ProjectID = "other"
			case "namespace":
				session.Namespace = "other"
			case "run":
				approval.RunID = "other"
			case "body":
				request.Body = "Modified body"
			}

			err := guardGuidedApprovalRun(session, request)
			if (err == nil) != (variant == "valid") {
				t.Fatalf("unexpected approval authority: %v", err)
			}
		})
	}
}

// TestGuidedUseCaseSpecLegacy preserves existing manual mutation receipt digests.
func TestGuidedUseCaseSpecLegacy(t *testing.T) {
	request := SpecMutationRequest{OperationID: "operation", Action: "create", Namespace: "default", ProjectID: "project", CaseSlug: "case", Title: "Title", Body: "Body", Status: "draft"}
	encoded, err := json.Marshal(request)

	expected := `{"OperationID":"operation","Action":"create","Namespace":"default","Reference":"","ExpectedRevision":"","ProjectID":"project","CaseSlug":"case","Title":"Title","Body":"Body","Status":"draft"}`
	if err != nil || string(encoded) != expected {
		t.Fatalf("manual mutation encoding changed: %v\n%s", err, encoded)
	}
}

// TestGuidedUseCaseSpecCheckpoint reuses the committed approved bytes after a lost seal acknowledgement.
func TestGuidedUseCaseSpecCheckpoint(t *testing.T) {
	root, session := guidedSpecSealFixture(t)
	r := &fileSessionRepository{projectDir: root}
	approval := session.Guided.SpecApproval
	paths := []string{approval.Reference, filepath.ToSlash(filepath.Join(filepath.Dir(approval.Reference), "metadata.json"))}

	var original string

	err := WithMutation(t.Context(), root, func(ctx context.Context) error {
		checkpoint, err := r.observeGuidedCheckpoint(ctx, session.ID, *session.Guided.PendingOperation, "specification", paths)
		if err != nil {
			return err
		}

		session.Guided.PendingCheckpoint = &checkpoint
		if err := RetainGuidedUseCaseProgress(ctx, root, &session); err != nil {
			return err
		}

		committed, err := r.commitGuidedCheckpoint(ctx, checkpoint, session.Guided.PendingOperation.Digest)
		original = committed.Commit

		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	guidedWriteFile(t, root, approval.Reference, "Later working-copy replacement")

	err = WithMutation(t.Context(), root, func(ctx context.Context) error { return SealGuidedSpec(ctx, root, &session) })
	if err != nil || session.Guided.Checkpoints[2].Commit != original {
		t.Fatalf("approved seal replaced or blocked by working copy: %v", err)
	}
}

func guidedSpecSealFixture(t *testing.T) (string, Session) {
	t.Helper()

	root, command := guidedCheckpointFixture(t)

	session, err := FinishGuidedUseCaseDump(t.Context(), root, command)
	if err != nil {
		t.Fatal(err)
	}

	source := session.Guided.Checkpoints[0]
	source.Stage = "quiz"
	session.Guided.Checkpoints = append(session.Guided.Checkpoints, source)

	result, err := MutateSpec(t.Context(), root, SpecMutationRequest{OperationID: "approved-spec", Action: "create", Namespace: session.Namespace, ProjectID: session.ProjectID, CaseSlug: session.Guided.CaseSlug, Title: "Approved specification", Body: "Exact approved content", Status: "approved"})
	if err != nil {
		t.Fatal(err)
	}

	session.Guided.Stage = "specification"
	session.Guided.SpecApproval = &GuidedSpecApprovalReference{Reference: result.Reference, Revision: result.Revision, GuidedSpecApproval: GuidedSpecApproval{ContentHash: contenthash.Sum([]byte("Exact approved content"))}}
	session.Guided.PendingOperation = &GuidedUseCaseOperation{ID: "seal-spec", Action: "spec", Digest: contenthash.Sum([]byte("seal-spec"))}

	return root, session
}
