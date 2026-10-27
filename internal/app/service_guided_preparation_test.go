package app

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestGuidedUseCasePreparationCommitted materializes only the frozen specification and seals initial contracts.
func TestGuidedUseCasePreparationCommitted(t *testing.T) {
	for _, namespace := range []string{"default", "research"} {
		t.Run(namespace, func(t *testing.T) {
			operations, session, model := guidedPreparationFixture(t, namespace)
			replaceGuidedWebDocuments(t, operations.projectContext.ProjectDir, session)

			command := guidedQuizCommand(session, "prepare", "prepare", "")

			prepared, err := operations.PrepareGuidedWork(t.Context(), command)
			if err != nil {
				t.Fatal(err)
			}

			assertGuidedPreparation(t, operations, prepared)

			assertGuidedPlanSources(t, model)

			replay, err := operations.PrepareGuidedWork(t.Context(), command)
			if err != nil || replay.Revision != prepared.Revision || len(model.base.requests) != 1 || model.calls != 1 {
				t.Fatal("preparation replay repeated generation or effects", err)
			}
		})
	}
}

func assertGuidedPreparation(t *testing.T, operations service, session authoringcore.Session) {
	t.Helper()

	work := session.Guided.Work
	if work == nil || work.Publication == nil || work.Materialization == nil || len(session.Guided.Checkpoints) != 4 {
		t.Fatal("preparation evidence incomplete")
	}

	checkpoint := session.Guided.Checkpoints[3]
	prefix := "ops/" + session.Namespace + "/"

	ref := work.Materialization
	assertGuidedInitialContracts(t, operations, checkpoint, prefix, ref)

	if checkpoint.Stage != "preparation" || len(checkpoint.Sources) <= len(session.Guided.Checkpoints[2].Artifacts) || ref.PlanRevision.BodyHash != work.Publication.Binding.BodyHash {
		t.Fatal("source or plan revision binding changed")
	}
}

// TestGuidedUseCasePreparationLostAck reconciles original plan and materialization receipts.
func TestGuidedUseCasePreparationLostAck(t *testing.T) {
	for _, phase := range []string{"plan", "materialization"} {
		t.Run(phase, func(t *testing.T) {
			operations, session, model := guidedPreparationFixture(t, "default")
			lost := errors.New("lost preparation acknowledgement")
			operations.afterGuidedPreparation = func(current string) error {
				if current == phase {
					return lost
				}

				return nil
			}

			command := guidedQuizCommand(session, "prepare", "prepare", "")
			if _, err := operations.PrepareGuidedWork(t.Context(), command); !errors.Is(err, lost) {
				t.Fatal("fault not reached", err)
			}

			operations.afterGuidedPreparation = nil

			prepared, err := operations.PrepareGuidedWork(t.Context(), command)
			if err != nil || len(model.base.requests) != 1 || model.calls != 1 {
				t.Fatal("original owner effect repeated", err)
			}

			assertGuidedPreparation(t, operations, prepared)
			assertOneGuidedMaterialization(t, operations)
		})
	}
}

// TestGuidedUseCasePreparationRetry keeps a saved plan when native contract generation fails.
func TestGuidedUseCasePreparationRetry(t *testing.T) {
	operations, session, model := guidedPreparationFixture(t, "default")
	model.fail = true
	command := guidedQuizCommand(session, "prepare", "prepare", "")

	blocked, err := operations.PrepareGuidedWork(t.Context(), command)
	if err == nil || blocked.Guided.Work.Publication == nil || blocked.Guided.Work.Materialization != nil {
		t.Fatal("partial preparation was not retained", err)
	}

	model.fail = false

	recovered, err := operations.PrepareGuidedWork(t.Context(), command)
	if err != nil || len(model.base.requests) != 1 || model.calls != 2 {
		t.Fatal("retry replaced the saved plan", err)
	}

	assertGuidedPreparation(t, operations, recovered)
}

// TestGuidedUseCasePreparationExisting cannot adopt an independently generated plan.
func TestGuidedUseCasePreparationExisting(t *testing.T) {
	operations, session, model := guidedPreparationFixture(t, "default")
	scope := quizcore.Scope{ProjectID: session.ProjectID, Namespace: session.Namespace, CaseSlug: session.Guided.CaseSlug}
	plan := serviceQuizPlanResponse("manual-plan", operations.utcNow()).Plan
	plan.Scope, plan.CreatedAt = scope, operations.utcNow()

	paths, err := authoringcore.WriteGeneratedPlan(t.Context(), operations.projectContext.ProjectDir, plan)
	if err != nil {
		t.Fatal(err)
	}

	before, err := os.ReadFile(paths.Path)
	if err != nil {
		t.Fatal(err)
	}

	_, err = operations.PrepareGuidedWork(t.Context(), guidedQuizCommand(session, "prepare", "prepare", ""))
	if !errors.Is(err, workflowstate.ErrConflict) || len(model.base.requests) != 0 {
		t.Fatal("manual plan adopted or replaced", err)
	}

	after, err := os.ReadFile(paths.Path)
	if err != nil || string(before) != string(after) {
		t.Fatal("manual plan changed", err)
	}
}

func guidedPreparationFixture(t *testing.T, namespace string) (service, authoringcore.Session, *planDeliveryTestModel) {
	t.Helper()

	operations, session, _ := guidedQuizFixture(t, namespace)

	for _, step := range []struct{ id, action, content string }{{"quiz", "quiz", ""}, {"answer", "answer", "Preserve the existing bounded behavior."}} {
		var err error

		session, err = operations.ConductGuidedQuiz(t.Context(), guidedQuizCommand(session, step.id, step.action, step.content))
		if err != nil {
			t.Fatal(err)
		}
	}

	session, err := operations.GenerateGuidedSpec(t.Context(), guidedQuizCommand(session, "spec", "spec", ""))
	if err != nil {
		t.Fatal(err)
	}

	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "workflow:\n  version: 1\nticket:\n  key_prefix: PRJ\n")

	model := &planDeliveryTestModel{}
	operations.quizModel = model
	operations.conversationModel = sufficientPlanReadinessProvider()

	return operations, session, model
}

func assertOneGuidedMaterialization(t *testing.T, operations service) {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join(operations.projectContext.ProjectDir, "ops", "default", "ticket", "*", "metadata.json"))
	if err != nil || len(paths) != 1 {
		t.Fatal("duplicate native tickets", paths, err)
	}
}

// TestGuidedUseCasePreparationCommit reconciles a real successful Git ref update with a lost acknowledgement.
func TestGuidedUseCasePreparationCommit(t *testing.T) {
	operations, session, _ := guidedPreparationFixture(t, "default")
	originalPath := os.Getenv("PATH")

	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}

	wrapper := t.TempDir()
	marker := filepath.Join(wrapper, "lost-once")

	script := `#!/bin/sh
"$CHR_GUIDED_REAL_GIT" "$@"
result=$?
operation=
ref=
for argument do
 case "$argument" in
  update-ref) operation=update-ref ;;
  refs/chr/guided/*) ref=guided ;;
 esac
done
if [ "$operation" = update-ref ] && [ "$ref" = guided ] && [ "$result" = 0 ] && [ ! -e "$CHR_GUIDED_LOST_ACK" ]; then
 : > "$CHR_GUIDED_LOST_ACK"
 exit 17
fi
exit "$result"
`
	if err := os.WriteFile(filepath.Join(wrapper, "git"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CHR_GUIDED_REAL_GIT", realGit)
	t.Setenv("CHR_GUIDED_LOST_ACK", marker)
	t.Setenv("PATH", wrapper+string(os.PathListSeparator)+originalPath)

	command := guidedQuizCommand(session, "prepare", "prepare", "")

	pending, err := operations.PrepareGuidedWork(t.Context(), command)
	if err == nil || pending.Guided.PendingCheckpoint == nil || pending.Guided.Work.Materialization == nil {
		t.Fatal("Git acknowledgement fault not retained", err)
	}

	t.Setenv("PATH", originalPath)

	root := operations.projectContext.ProjectDir
	original := serviceWorkflowDependencyGit(t, root, "rev-parse", pending.Guided.PendingCheckpoint.Ref)
	member := pending.Guided.Work.Materialization.Members[0]

	path := filepath.Join(root, "ops", session.Namespace, filepath.FromSlash(member.Path))
	if err := os.WriteFile(path, []byte("Later edited native working copy"), 0o600); err != nil {
		t.Fatal(err)
	}

	recovered, err := operations.PrepareGuidedWork(t.Context(), command)
	if err != nil || recovered.Guided.Checkpoints[3].Commit != original {
		t.Fatal("recovery substituted later contracts or repeated the commit", err)
	}

	assertGuidedPreparation(t, operations, recovered)
}

func assertGuidedPlanSources(t *testing.T, model *planDeliveryTestModel) {
	t.Helper()

	for _, request := range model.base.requests {
		for _, section := range request.ContextSections {
			if strings.Contains(section.Content, "Later replacement") {
				t.Fatal("planning consumed mutable documents")
			}
		}
	}
}

func assertGuidedInitialContracts(t *testing.T, operations service, checkpoint authoringcore.GuidedStageCheckpoint, prefix string, ref *authoringcore.GuidedMaterializationReference) {
	t.Helper()

	for _, member := range ref.Members {
		content, err := authoringcore.ReadGuidedCommittedArtifact(t.Context(), operations.projectContext.ProjectDir, checkpoint, prefix+member.Path)
		if err != nil || !strings.Contains(string(content), member.ContractHash) {
			t.Fatal("initial native contract not sealed", err)
		}
	}
}

// TestGuidedUseCasePreparationRepeated refuses a fresh preparation identity after the stage is sealed.
func TestGuidedUseCasePreparationRepeated(t *testing.T) {
	operations, session, model := guidedPreparationFixture(t, "default")

	prepared, err := operations.PrepareGuidedWork(t.Context(), guidedQuizCommand(session, "prepare", "prepare", ""))
	if err != nil {
		t.Fatal(err)
	}

	rejected, err := operations.PrepareGuidedWork(t.Context(), guidedQuizCommand(prepared, "replacement-prepare", "prepare", ""))
	if !errors.Is(err, workflowstate.ErrConflict) || rejected.Revision != prepared.Revision || len(model.base.requests) != 1 || model.calls != 1 {
		t.Fatal("sealed preparation was dispatched again", err)
	}
}
