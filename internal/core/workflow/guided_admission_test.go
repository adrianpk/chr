package workflow

import (
	"errors"
	"strings"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestGuidedUseCaseAdmissionAuthority binds intent, actor, operation and immutable preparation to one request.
func TestGuidedUseCaseAdmissionAuthority(t *testing.T) {
	for _, variant := range []string{"exact", "intent", "actor", "policy", "operation", "checkpoint", "request", "target", "later tickets"} {
		t.Run(variant, func(t *testing.T) {
			session, request := guidedAdmissionProofFixture(t)
			changeGuidedAdmissionProof(&session, &request, variant)

			err := guardGuidedAdmissionProof(session, "project", "default", request)
			if variant == "exact" && err != nil || variant != "exact" && !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("unexpected progression authority", err)
			}
		})
	}
}

func guidedAdmissionProofFixture(t *testing.T) (authoringcore.Session, StartRequest) {
	t.Helper()

	checkpoint := authoringcore.GuidedStageCheckpoint{Stage: "preparation", Commit: strings.Repeat("a", 40), Artifacts: []authoringcore.GuidedCommittedArtifact{{Path: "ops/default/ticket/one/metadata.json", Hash: strings.Repeat("b", 64)}}}

	digest, err := authoringcore.GuidedCheckpointDigest(checkpoint)
	if err != nil {
		t.Fatal(err)
	}

	proof := workflowstate.GuidedAdmission{RunID: "run", Actor: "operator", Policy: authoringcore.GuidedWorkAdmissionPolicy, OperationDigest: "operation", CheckpointDigest: digest, MaterializationID: "materialized"}

	selection, err := normalizeWorkflowPreviewRequest(workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "slice_set", Identity: "set"}, Limit: workflowstate.PageLimit})
	if err != nil {
		t.Fatal(err)
	}

	request := StartRequest{StartInput: workflowstate.StartInput{
		RequestKey: "admission",
		Selection:  selection,
		Admission:  &workflowstate.AdmissionGuard{Guided: &proof},
	}}
	session := authoringcore.Session{ID: "run", ProjectID: "project", Namespace: "default", Revision: 1, Guided: &authoringcore.GuidedUseCaseState{SchemaVersion: 1, Actor: "operator", ProgressionIntent: true, Stage: "preparation", Work: &authoringcore.GuidedWorkState{Materialization: &authoringcore.GuidedMaterializationReference{ID: "materialized"}, Admission: &request.StartInput}, PendingOperation: &authoringcore.GuidedUseCaseOperation{Action: "admit", Digest: "operation"}, Checkpoints: []authoringcore.GuidedStageCheckpoint{checkpoint}}}

	return session, request
}

func changeGuidedAdmissionProof(session *authoringcore.Session, request *StartRequest, variant string) {
	switch variant {
	case "intent":
		session.Guided.ProgressionIntent = false
	case "actor":
		request.Admission.Guided.Actor = "foreign actor"
	case "policy":
		request.Admission.Guided.Policy = "other"
	case "operation":
		request.Admission.Guided.OperationDigest = "other"
	case "checkpoint":
		session.Guided.Checkpoints[0].Commit = strings.Repeat("b", 40)
	case "request":
		request.RequestKey = "other"
	case "target":
		request.Selection.Target.Commit = strings.Repeat("b", 40)
	case "later tickets":
		request.Policy.IncludeLaterTickets = true
	}
}

// TestGuidedUseCaseControlAdmissionAuthority distinguishes an authorized but
// stopped run from missing progression authority, without dispatching either.
func TestGuidedUseCaseControlAdmissionAuthority(t *testing.T) {
	for _, status := range []string{"paused", "cancelled", "completed", "failed", "stop_intent"} {
		t.Run(status, func(t *testing.T) {
			session, request := guidedAdmissionProofFixture(t)
			if status == "stop_intent" {
				session.Guided.Control = &authoringcore.GuidedControlState{StopIntent: "pause"}
			} else {
				session.Guided.Status = status
			}

			if err := guardGuidedAdmissionProof(session, "project", "default", request); !errors.Is(err, authoringcore.ErrGuidedUseCaseStopped) {
				t.Fatal("authorized stopped admission lost its lifecycle guard", err)
			}
		})
	}
}
