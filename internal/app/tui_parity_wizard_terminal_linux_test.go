package app

import (
	"maps"
	"os"
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
)

// A fresh production terminal carries Name/Create, exact capture, clarification,
// native review and approval through real Git publication and verified results.
func tuiParityWizardTerminal(t *testing.T) {
	t.Helper()

	operations, seed, _ := guidedAdmissionFixture(t)
	if _, err := operations.ChangeGuidedUseCase(t.Context(), guidedQuizCommand(seed, "park-fixture", "pause", "")); err != nil {
		t.Fatal(err)
	}

	provider, ok := operations.workflowPullRequests.(*serviceWorkEntryProvider)
	if !ok {
		t.Fatal("missing owned canonical provider")
	}

	t.Setenv("CHR_PARITY_REMOTE", provider.remote)

	terminal := startTUITerminal(t, operations, "TestTUIParityProcess")
	captureTUIParityWizard(t, terminal)
	terminal.see(t, "Native review requires approval")

	review := tuiParityWizardSession(t, operations)

	native := assertGuidedNativeAdmission(t, operations, review)
	if native.Review == nil || review.Guided.Capture != "PARITY_EXACT_CAPTURE and final input." || len(review.Guided.Checkpoints) != 4 {
		t.Fatal("terminal did not preserve original input or native review")
	}

	approveTUIParityWizard(t, terminal)
	terminal.stop(t)

	completed := tuiParityWizardSession(t, operations)
	assertTUIParityDelivery(t, operations, completed)
	reopenTUIParityWizard(t, operations, completed)
}

func captureTUIParityWizard(t *testing.T, terminal *guidedTerminal) {
	t.Helper()
	terminal.keys(t, "w")
	terminal.see(t, "[Create]")
	terminal.keys(t, "Acceptance delivery\x13")
	terminal.see(t, "Add the known context")
	terminal.keys(t, "e")
	terminal.see(t, "ctrl+s Continue to quiz")
	terminal.keys(t, "PARITY_EXACT_CAPTURE and final input.\x13")
	terminal.see(t, "What should the workflow preserve")
	terminal.keys(t, "e")
	terminal.see(t, "alt+x cancels quiz")
	terminal.keys(t, "Keep the complete original scope and finite review policy.\r")
}

func approveTUIParityWizard(t *testing.T, terminal *guidedTerminal) {
	t.Helper()
	terminal.keys(t, "c")
	resizeTUIParityTerminal(t, terminal, 160, 60)
	terminal.see(t, "Attention")
	terminal.keys(t, "\t")
	terminal.keys(t, "\x1bm")
	terminal.keys(t, "x")
	terminal.see(t, "Saved documents and completed changes will be kept.")
	terminal.keys(t, "\x1b")
	terminal.see(t, "a Inspect native approval")
	terminal.keys(t, "a")
	terminal.see(t, "Inspect native workflow")
	terminal.keys(t, "\r")
	terminal.see(t, "Confirm approve")
	terminal.keys(t, "\r")
	terminal.see(t, "All admitted work has confirmed delivery")
	terminal.keys(t, "v")
	terminal.see(t, `"Complete": true`)
	terminal.keys(t, "v")
	terminal.see(t, "Confirmed delivery")
}

func tuiParityWizardSession(t *testing.T, operations service) authoringcore.Session {
	t.Helper()

	view, err := operations.InspectGuidedUseCase(t.Context(), "", "default")
	if err != nil {
		t.Fatal(err)
	}

	for _, run := range view.Runs {
		if run.Guided != nil && run.Guided.CaseSlug == "acceptance-delivery" {
			return readWebGuidedSession(t, operations, run.ID)
		}
	}

	t.Fatal("missing fresh acceptance case run")

	return authoringcore.Session{}
}

func assertTUIParityDelivery(t *testing.T, operations service, session authoringcore.Session) {
	t.Helper()

	if session.Guided.Status != "completed" || session.Guided.Stage != "results" || session.Guided.Result == nil {
		t.Fatal("terminal did not reach evidence-backed completion")
	}

	native, err := operations.ReadWorkflow(t.Context(), session.Guided.Work.Invocation.ID)
	if err != nil || serviceWorkTaskAttempts(native) != 1 {
		t.Fatal("terminal duplicated work", err)
	}

	assertGuidedNativeBinding(t, native, session)
	assertTUIParityOriginals(t, operations, session)
}

func assertTUIParityOriginals(t *testing.T, operations service, session authoringcore.Session) {
	t.Helper()

	for _, kind := range []string{"dump", "quiz", "specification", "plan", "tracker"} {
		original, err := operations.ReadGuidedUseCaseOriginal(t.Context(), session.ID, session.Namespace, kind)
		if err != nil || original.Content == "" || original.Commit == "" || original.Reference == "" {
			t.Fatal("terminal lost a committed original", kind, err)
		}
	}
}

func reopenTUIParityWizard(t *testing.T, operations service, before authoringcore.Session) {
	t.Helper()

	image := workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))
	terminal := startTUITerminal(t, operations, "TestTUIParityProcess")
	terminal.catalog(t)
	terminal.see(t, "Acceptance delivery")
	terminal.keys(t, "\t")
	terminal.keys(t, "\x1bm")
	terminal.see(t, "All admitted work has confirmed delivery")
	terminal.keys(t, "1")
	terminal.see(t, "PARITY_EXACT_CAPTURE and final input.")
	terminal.stop(t)

	after := tuiParityWizardSession(t, operations)
	if after.ID != before.ID || after.Revision != before.Revision || !maps.Equal(image, workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))) {
		t.Fatal("completed restart advanced work or changed its originals")
	}
}

// Only model and external forge responses are deterministic. The child uses the
// real managed host, service adapters, SQLite, Git workers and local publication.
func TestTUIParityProcess(t *testing.T) {
	project := os.Getenv("CHR_GUIDED_TUI_PROJECT")
	if project == "" {
		t.Skip("launched by the owned parity terminal fixture")
	}

	env, err := core.NewUserEnvironment(core.UserEnvironmentOptions{DevMode: true, DevRoot: os.Getenv("CHR_GUIDED_TUI_ROOT")})
	if err != nil {
		t.Fatal(err)
	}

	operations := newService(t.Context(), env, commandProjectContext{ProjectDir: project, ProjectID: os.Getenv("CHR_GUIDED_TUI_ID")}, testAssetsFS())
	configureTUIParityProviders(t, &operations)

	factory := commandFactory{deps: testAppDependencies(), io: IO{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr}}

	options, err := factory.loadTUIOptions(t.Context(), env, operations.projectContext)
	if err != nil {
		t.Fatal(err)
	}

	if err := runManagedTUI(t.Context(), options, operations); err != nil {
		t.Fatal(err)
	}
}

func configureTUIParityProviders(t *testing.T, operations *service) {
	t.Helper()

	now := operations.utcNow()
	completion := serviceQuizModelResponse("complete", "complete", now)
	completion.Question.Required = false
	operations.quizModel = &guidedWebWorkModel{quiz: fakeServiceQuizModel{responses: []quizmodel.Response{serviceQuizModelResponse("first", "question-1", now), completion}}}
	operations.conversationModel = sufficientPlanReadinessProvider()
	operations.generationBinder = testGenerationBinder

	remote, repository := os.Getenv("CHR_PARITY_REMOTE"), "https://forge.example/test/guided"
	provider := &serviceWorkEntryProvider{serviceWorkflowSubmissionFixture: &serviceWorkflowSubmissionFixture{serviceWorkflowPreparationFixture: &serviceWorkflowPreparationFixture{&serviceWorkflowReviewFixture{&serviceWorkflowPRFixture{}}}}, remote: remote, repository: repository}
	operations.workflowPullRequests = &guidedWebPublicationFixture{provider}
	operations.workflowIntegrationGit = workflowcore.NewWorkflowIntegrationGit(repository, remote)
	operations.workflowPreparationGit = workflowcore.NewWorkflowPreparationGit(repository, remote)
	operations.workflowWorker = []string{"sh", "-c", "printf 'Delivered\\n' > result.txt && git add result.txt && git -c user.name=Test -c user.email=test@example.com commit -qm 'Task result'"}
}
