package core

import (
	"errors"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestGuidedUseCaseAdmissionMembership refuses replaced, missing, foreign or additional native work.
func TestGuidedUseCaseAdmissionMembership(t *testing.T) {
	member := WorkflowPreviewMember{Key: "PRJ-1", Path: "ticket/one/metadata.json", Revision: 1, SourceHash: strings.Repeat("a", 64), ContractHash: strings.Repeat("b", 64), SetID: "set-one"}
	reference := GuidedMaterializationReference{SetID: member.SetID, Members: []GuidedWorkMember{{Key: member.Key, Path: member.Path, Revision: member.Revision, SourceHash: member.SourceHash, ContractHash: member.ContractHash}}}

	for _, variant := range []string{"exact", "revision", "source", "contract", "path", "set", "missing", "additional"} {
		t.Run(variant, func(t *testing.T) {
			preview := WorkflowPreview{Selector: workflowstate.WorkflowSelector{Extent: "slice_set", Identity: reference.SetID}, Members: []WorkflowPreviewMember{member}, Total: 1}
			changeGuidedPreview(&preview, variant)

			err := ValidateGuidedWorkPreview(reference, preview)
			if variant == "exact" && err != nil || variant != "exact" && !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("unexpected membership authority", err)
			}
		})
	}
}

func changeGuidedPreview(preview *WorkflowPreview, variant string) {
	switch variant {
	case "revision":
		preview.Members[0].Revision++
	case "source":
		preview.Members[0].SourceHash = strings.Repeat("c", 64)
	case "contract":
		preview.Members[0].ContractHash = strings.Repeat("c", 64)
	case "path":
		preview.Members[0].Path = "ticket/foreign/metadata.json"
	case "set":
		preview.Members[0].SetID = "foreign"
	case "missing":
		preview.Members = nil
	case "additional":
		preview.Total++
		preview.Members = append(preview.Members, preview.Members[0])
	}
}

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

func guidedAdmissionProofFixture(t *testing.T) (AuthoringSession, WorkflowStartRequest) {
	t.Helper()

	checkpoint := GuidedStageCheckpoint{Stage: "preparation", Commit: strings.Repeat("a", 40), Artifacts: []GuidedCommittedArtifact{{Path: "ops/default/ticket/one/metadata.json", Hash: strings.Repeat("b", 64)}}}

	digest, err := GuidedCheckpointDigest(checkpoint)
	if err != nil {
		t.Fatal(err)
	}

	proof := workflowstate.GuidedWorkflowAdmission{RunID: "run", Actor: "operator", Policy: GuidedWorkAdmissionPolicy, OperationDigest: "operation", CheckpointDigest: digest, MaterializationID: "materialized"}

	selection, err := normalizeWorkflowPreviewRequest(workflowstate.WorkflowPreviewRequest{Selector: workflowstate.WorkflowSelector{Extent: "slice_set", Identity: "set"}, Limit: workflowstate.PageLimit})
	if err != nil {
		t.Fatal(err)
	}

	request := WorkflowStartRequest{RequestKey: "admission", Selection: selection, Admission: &workflowstate.WorkflowAdmissionGuard{Guided: &proof}}
	session := AuthoringSession{ID: "run", ProjectID: "project", Namespace: "default", Revision: 1, Guided: &GuidedUseCaseState{SchemaVersion: 1, Actor: "operator", ProgressionIntent: true, Stage: "preparation", Work: &GuidedWorkState{Materialization: &GuidedMaterializationReference{ID: "materialized"}, Admission: &request}, PendingOperation: &GuidedUseCaseOperation{Action: "admit", Digest: "operation"}, Checkpoints: []GuidedStageCheckpoint{checkpoint}}}

	return session, request
}

func changeGuidedAdmissionProof(session *AuthoringSession, request *WorkflowStartRequest, variant string) {
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
				session.Guided.Control = &GuidedControlState{StopIntent: "pause"}
			} else {
				session.Guided.Status = status
			}

			if err := guardGuidedAdmissionProof(session, "project", "default", request); !errors.Is(err, ErrGuidedUseCaseStopped) {
				t.Fatal("authorized stopped admission lost its lifecycle guard", err)
			}
		})
	}
}
