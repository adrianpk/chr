package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestGuidedUseCaseSpecApproval uses committed clarification and records exact default approval.
func TestGuidedUseCaseSpecApproval(t *testing.T) {
	operations, session, model := guidedSpecFixture(t)
	for _, checkpoint := range session.Guided.Checkpoints {
		for _, artifact := range checkpoint.Artifacts {
			if !strings.HasSuffix(artifact.Path, "/artifact.md") {
				continue
			}

			if err := os.WriteFile(filepath.Join(operations.projectContext.ProjectDir, filepath.FromSlash(artifact.Path)), []byte("Later working-copy replacement"), 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}

	completed, err := operations.GenerateGuidedSpec(t.Context(), guidedQuizCommand(session, "generate-spec", "spec", ""))
	if err != nil {
		t.Fatal(err)
	}

	assertGuidedSpecApproval(t, operations, completed)

	for _, section := range model.requests[2].ContextSections {
		if strings.Contains(section.Content, "Later working-copy replacement") {
			t.Fatal("specification consumed working copies")
		}
	}
}

func assertGuidedSpecApproval(t *testing.T, operations service, session authoringcore.AuthoringSession) {
	t.Helper()

	approval := session.Guided.SpecApproval
	if session.Guided.Stage != "preparation" || approval == nil || approval.Policy != authoringcore.GuidedSpecApprovalPolicy || approval.Actor != "operator" {
		t.Fatalf("exact approval missing: %#v", session.Guided)
	}

	if approval.GeneratedRevision == approval.Revision || approval.ContentHash == "" || approval.InputDigest == "" {
		t.Fatal("approval did not bind generated content and owner revisions")
	}

	assertGuidedSpecCheckpoint(t, operations, session)
}

func assertGuidedSpecCheckpoint(t *testing.T, operations service, session authoringcore.AuthoringSession) {
	t.Helper()

	approval := session.Guided.SpecApproval
	checkpoint := session.Guided.Checkpoints[2]

	content, err := core.ReadGuidedCommittedArtifact(t.Context(), operations.projectContext.ProjectDir, checkpoint, approval.Reference)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), `status: "approved"`) || !strings.Contains(string(content), `guided_run: "guided-quiz"`) {
		t.Fatalf("approved version not committed: %v\n%s", err, content)
	}

	if len(checkpoint.Sources) <= len(session.Guided.Checkpoints[1].Artifacts) {
		t.Fatal("spec checkpoint lost the dump dependency")
	}
}

// TestGuidedUseCaseSpecLostAck reuses the original generated draft and approval receipts.
func TestGuidedUseCaseSpecLostAck(t *testing.T) {
	for _, phase := range []string{"draft", "approval"} {
		t.Run(phase, func(t *testing.T) {
			operations, session, model := guidedSpecFixture(t)
			lost := errors.New("lost specification owner acknowledgement")
			operations.afterGuidedSpec = func(current string) error {
				if current == phase {
					return lost
				}

				return nil
			}

			command := guidedQuizCommand(session, "generate-spec", "spec", "")
			if _, err := operations.GenerateGuidedSpec(t.Context(), command); !errors.Is(err, lost) {
				t.Fatalf("fault not reached: %v", err)
			}

			original := singleGuidedSpec(t, operations.projectContext.ProjectDir)
			operations.afterGuidedSpec = nil

			recovered, err := operations.GenerateGuidedSpec(t.Context(), command)
			if err != nil || recovered.Guided.SpecApproval.Reference != original || len(model.requests) != 3 {
				t.Fatalf("specification effect repeated: %v", err)
			}
		})
	}
}

// TestGuidedUseCaseSpecChanged cannot approve bytes edited after generation.
func TestGuidedUseCaseSpecChanged(t *testing.T) {
	operations, session, model := guidedSpecFixture(t)
	lost := errors.New("draft saved before response loss")
	operations.afterGuidedSpec = func(string) error { return lost }

	command := guidedQuizCommand(session, "generate-spec", "spec", "")
	if _, err := operations.GenerateGuidedSpec(t.Context(), command); !errors.Is(err, lost) {
		t.Fatal(err)
	}

	reference := singleGuidedSpec(t, operations.projectContext.ProjectDir)

	document, err := core.ReadSpecDocument(t.Context(), operations.projectContext.ProjectDir, reference)
	if err != nil {
		t.Fatal(err)
	}

	_, err = core.MutateSpec(t.Context(), operations.projectContext.ProjectDir, core.SpecMutationRequest{OperationID: "manual-edit", Action: "save", Reference: reference, ExpectedRevision: document.Revision, ProjectID: session.ProjectID, CaseSlug: session.Guided.CaseSlug, Title: document.Title, Body: "A later manual specification.", Status: "draft"})
	if err != nil {
		t.Fatal(err)
	}

	operations.afterGuidedSpec = nil

	blocked, err := operations.GenerateGuidedSpec(t.Context(), command)
	if !errors.Is(err, workflowstate.ErrConflict) || blocked.Guided.SpecApproval != nil || len(model.requests) != 3 {
		t.Fatalf("later content approved or generation repeated: %v", err)
	}

	document, err = core.ReadSpecDocument(t.Context(), operations.projectContext.ProjectDir, reference)
	if err != nil || document.Status != "draft" {
		t.Fatalf("manual lifecycle changed: %v", err)
	}
}

// TestGuidedUseCaseSpecInvalid cannot approve an invalid provider result.
func TestGuidedUseCaseSpecInvalid(t *testing.T) {
	operations, session, model := guidedSpecFixture(t)
	response := serviceQuizDraftSpecResponse("invalid", operations.utcNow())
	response.DraftSpec.Body = ""
	model.responses = []quizmodel.Response{response}

	blocked, err := operations.GenerateGuidedSpec(t.Context(), guidedQuizCommand(session, "generate-spec", "spec", ""))
	if err == nil || blocked.Guided.SpecApproval != nil || len(blocked.Guided.Checkpoints) != 2 {
		t.Fatalf("invalid generation advanced: %v", err)
	}
}

func guidedSpecFixture(t *testing.T) (service, authoringcore.AuthoringSession, *fakeServiceQuizModel) {
	t.Helper()

	operations, session, model := guidedQuizFixture(t, "default")

	session, err := operations.ConductGuidedQuiz(t.Context(), guidedQuizCommand(session, "start-quiz", "quiz", ""))
	if err != nil {
		t.Fatal(err)
	}

	session, err = operations.ConductGuidedQuiz(t.Context(), guidedQuizCommand(session, "answer", "answer", "Retain the agreed bounded workflow."))
	if err != nil {
		t.Fatal(err)
	}

	return operations, session, model
}

func singleGuidedSpec(t *testing.T, root string) string {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join(root, "ops", "default", "spec", "*", "artifact.md"))
	if err != nil || len(paths) != 1 {
		t.Fatalf("specification identities duplicated: %v %v", paths, err)
	}

	relative, err := filepath.Rel(root, paths[0])
	if err != nil {
		t.Fatal(err)
	}

	return filepath.ToSlash(relative)
}

// TestGuidedUseCaseSpecStage rejects out-of-order generation without reserving a pending effect.
func TestGuidedUseCaseSpecStage(t *testing.T) {
	operations, session, model := guidedQuizFixture(t, "default")

	rejected, err := operations.GenerateGuidedSpec(t.Context(), guidedQuizCommand(session, "premature-spec", "spec", ""))
	if !errors.Is(err, workflowstate.ErrConflict) || rejected.Revision != session.Revision || rejected.Guided.PendingOperation != nil || len(model.requests) != 0 {
		t.Fatal("out-of-order command poisoned the current stage")
	}
}
