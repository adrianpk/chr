package app

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type guidedAdmissionModel struct{ planDeliveryTestModel }

func (model *guidedAdmissionModel) Invoke(ctx context.Context, request quizmodel.Request) (quizmodel.Response, error) {
	response, err := model.planDeliveryTestModel.Invoke(ctx, request)
	if err != nil || response.PlanTracking.Workflow == nil {
		return response, err
	}

	plan := response.PlanTracking.Workflow
	gate := workflowstate.Validation{Required: true, Commands: [][]string{{"git", "show", "HEAD:result.txt"}}, TimeoutSeconds: 10}

	plan.ClosingValidation = gate
	for i := range plan.Slices {
		plan.Slices[i].Validation = gate
		for j := range plan.Slices[i].Tasks {
			plan.Slices[i].Tasks[j].Validation = gate
		}
	}

	return response, nil
}

type guidedAdmissionHost struct{ ids []string }

func (*guidedAdmissionHost) ready() error { return nil }
func (host *guidedAdmissionHost) activate(_ service, id string) error {
	host.ids = append(host.ids, id)

	return nil
}

func guidedAdmissionFixture(t *testing.T) (service, authoringcore.Session, *guidedAdmissionHost) {
	t.Helper()

	operations, session, _ := guidedPreparationFixture(t, "default")
	project := operations.projectContext.ProjectDir
	writeCLIWorkspaceConfig(t, project, "workflow:\n  version: 1\nreview:\n  mode: human_review\nticket:\n  key_prefix: PRJ\nwork:\n  state_authority: tickets\n  include_later_tickets: true\n")

	operations.quizModel = &guidedAdmissionModel{}

	remote := filepath.Join(t.TempDir(), "canonical.git")
	repository := "https://forge.example/test/guided"

	serviceWorkflowDeliveryGit(t, project, "init", "--bare", remote)
	serviceWorkflowDeliveryGit(t, project, "remote", "add", "origin", repository+".git")

	provider := &serviceWorkEntryProvider{serviceWorkflowSubmissionFixture: &serviceWorkflowSubmissionFixture{serviceWorkflowPreparationFixture: &serviceWorkflowPreparationFixture{&serviceWorkflowReviewFixture{&serviceWorkflowPRFixture{}}}}, remote: remote, repository: repository}
	operations.workflowPullRequests = provider
	operations.workflowIntegrationGit = workflowcore.NewWorkflowIntegrationGit(repository, remote)
	operations.workflowPreparationGit = workflowcore.NewWorkflowPreparationGit(repository, remote)
	operations.workflowWorker = []string{"sh", "-c", "printf 'Delivered\\n' > result.txt && git add result.txt && git -c user.name=Test -c user.email=test@example.com commit -qm 'Task result'"}
	host := &guidedAdmissionHost{}
	operations.guidedWorkHost = host

	prepared, err := operations.PrepareGuidedWork(t.Context(), guidedQuizCommand(session, "prepare", "prepare", ""))
	if err != nil {
		t.Fatal(err)
	}

	serviceWorkflowPublishCoordinator(t, project)
	serviceWorkflowDeliveryGit(t, project, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-qm", "Register native coordinator")
	serviceWorkflowDeliveryGit(t, project, "push", remote, "HEAD:refs/heads/dev")

	return operations, prepared, host
}

// TestGuidedUseCaseAdmissionExact excludes older and later unrelated tickets and freezes original finite policy.
func TestGuidedUseCaseAdmissionExact(t *testing.T) {
	t.Parallel()

	operations, prepared, host := guidedAdmissionFixture(t)
	addGuidedUnrelatedTicket(t, operations, prepared, "PRJ-90")

	command := guidedQuizCommand(prepared, "admit", "admit", "")

	started, err := operations.AdmitGuidedWork(t.Context(), command)
	if err != nil {
		t.Fatal(err)
	}

	addGuidedUnrelatedTicket(t, operations, prepared, "PRJ-91")
	assertGuidedNativeAdmission(t, operations, started)

	replayed, err := operations.AdmitGuidedWork(t.Context(), command)
	if err != nil || replayed.Revision != started.Revision || len(host.ids) != 1 {
		t.Fatal("replayed command repeated activation", err)
	}

	fresh := guidedQuizCommand(started, "different-admission", "admit", "")
	if _, err := operations.AdmitGuidedWork(t.Context(), fresh); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("new admission replaced the original", err)
	}
}

func assertGuidedNativeAdmission(t *testing.T, operations service, session authoringcore.Session) workflowcore.Invocation {
	t.Helper()

	work := session.Guided.Work
	if work.Invocation == nil || session.Guided.Stage != "execution" {
		t.Fatal("missing native invocation")
	}

	run, err := operations.ReadWorkflow(t.Context(), work.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	assertGuidedNativeBinding(t, run, session)

	return run
}

func assertGuidedNativeBinding(t *testing.T, run workflowcore.Invocation, session authoringcore.Session) {
	t.Helper()

	work := session.Guided.Work
	if len(run.Members) != len(work.Materialization.Members) || run.Policy.IncludeLaterTickets || run.Policy.WorkerTimeoutSeconds < 1 || run.Policy.MaxWorkerAttempts < 1 {
		t.Fatal("native membership or finite limits changed")
	}

	for i, member := range run.Members {
		if member.Admission.Key != work.Materialization.Members[i].Key || member.Admission.ContractHash != work.Materialization.Members[i].ContractHash {
			t.Fatal("admission substituted a task")
		}
	}

	assertGuidedNativeProvenance(t, run, session)
}

func assertGuidedNativeProvenance(t *testing.T, run workflowcore.Invocation, session authoringcore.Session) {
	t.Helper()

	work := session.Guided.Work

	proof := run.Admission.Guided
	if proof.Actor != session.Guided.Actor || proof.Policy != authoringcore.GuidedWorkAdmissionPolicy || proof.MaterializationID != work.Materialization.ID || run.Policy != work.Admission.Admission.ExecutionPolicy {
		t.Fatal("admission lost authority or policy provenance")
	}
}

func addGuidedUnrelatedTicket(t *testing.T, operations service, session authoringcore.Session, key string) {
	t.Helper()

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	record, err := store.Read(t.Context(), session.Guided.Work.Materialization.Members[0].Path)
	if err != nil {
		t.Fatal(err)
	}

	task := *record.Task

	task.Key, task.SetID, task.SliceID, task.UseCase, task.Order = key, "", "", "", 0
	if _, err := store.Create(t.Context(), "Unrelated task", workflowcore.Record{Kind: "ticket", Task: &task, Domain: record.Domain, Body: record.Body}); err != nil {
		t.Fatal(err)
	}
}

// TestGuidedUseCaseAdmissionLostAck reconciles native admission and start before any new request or budget allocation.
func TestGuidedUseCaseAdmissionLostAck(t *testing.T) {
	t.Parallel()

	for _, phase := range []string{"admission", "start"} {
		t.Run(phase, func(t *testing.T) {
			testGuidedAdmissionLostAck(t, phase)
		})
	}
}

func testGuidedAdmissionLostAck(t *testing.T, phase string) {
	t.Helper()

	operations, prepared, host := guidedAdmissionFixture(t)
	lost := errors.New("lost native acknowledgement")
	operations.afterGuidedAdmission = func(current string) error {
		if current == phase {
			return lost
		}

		return nil
	}

	command := guidedQuizCommand(prepared, "admit", "admit", "")
	if _, err := operations.AdmitGuidedWork(t.Context(), command); !errors.Is(err, lost) {
		t.Fatal("fault not reached", err)
	}

	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "workflow:\n  version: 1\nwork:\n  state_authority: tickets\n  max_worker_attempts: 1\n")

	operations.afterGuidedAdmission = nil

	recovered, err := operations.AdmitGuidedWork(t.Context(), command)
	if err != nil {
		t.Fatal("retained receipt did not reconcile before changed policy", err)
	}

	run := assertGuidedNativeAdmission(t, operations, recovered)
	for _, id := range host.ids {
		if id != run.ID {
			t.Fatal("start acknowledgement recovery substituted an invocation")
		}
	}

	page, err := operations.WorkflowStatusPage(t.Context(), 0, 8)
	if err != nil || len(page.Items) != 1 || len(run.Attempts) != 0 {
		t.Fatal("admission duplicated native work", err)
	}
}

// TestGuidedUseCaseAdmissionHost uses a managed lifetime to execute one real worker and stop at native human review.
func TestGuidedUseCaseAdmissionHost(t *testing.T) {
	t.Parallel()

	operations, prepared, _ := guidedAdmissionFixture(t)
	host := newGuidedWorkHost(t.Context())
	t.Cleanup(host.close)

	operations.guidedWorkHost = host

	requestCtx, cancel := context.WithCancel(t.Context())
	started, err := operations.AdmitGuidedWork(requestCtx, guidedQuizCommand(prepared, "admit", "admit", ""))

	cancel()

	if err != nil {
		t.Fatal(err)
	}

	host.wait.Wait()

	run := assertGuidedNativeAdmission(t, operations, started)
	if serviceWorkTaskAttempts(run) != 1 || run.Review == nil || run.Review.State != workflowstate.ReviewWaitingApproval || len(run.Deliveries) != 0 {
		t.Fatalf("native progression did not stop at its actual review boundary: status=%s outcome=%s review=%+v attempts=%d", run.Status, run.Outcome, run.Review, len(run.Attempts))
	}

	if err := host.activate(operations, run.ID); err != nil {
		t.Fatal(err)
	}

	host.wait.Wait()

	replayed, err := operations.ReadWorkflow(t.Context(), run.ID)
	if err != nil || serviceWorkTaskAttempts(replayed) != 1 || replayed.Policy != run.Policy {
		t.Fatal("duplicate activation replayed worker or reset limits", err)
	}
}

// TestGuidedUseCaseAdmissionStartedAck recovers a lost start response after the actual worker consumed its attempt.
func TestGuidedUseCaseAdmissionStartedAck(t *testing.T) {
	t.Parallel()

	operations, prepared, _ := guidedAdmissionFixture(t)
	host := newGuidedWorkHost(t.Context())
	t.Cleanup(host.close)

	operations.guidedWorkHost = host
	lost := errors.New("lost start acknowledgement")
	operations.afterGuidedAdmission = func(phase string) error {
		if phase == "start" {
			return lost
		}

		return nil
	}
	command := guidedQuizCommand(prepared, "admit", "admit", "")

	pending, err := operations.AdmitGuidedWork(t.Context(), command)
	if !errors.Is(err, lost) || pending.Guided.Work.Invocation == nil || pending.Guided.PendingOperation == nil {
		t.Fatal("lost start was not retained", err)
	}

	host.wait.Wait()

	operations.afterGuidedAdmission = nil

	recovered, err := operations.AdmitGuidedWork(t.Context(), command)
	if err != nil {
		t.Fatal(err)
	}

	host.wait.Wait()

	run := assertGuidedNativeAdmission(t, operations, recovered)
	if run.ID != pending.Guided.Work.Invocation.ID || serviceWorkTaskAttempts(run) != 1 || run.Review == nil || run.Review.State != workflowstate.ReviewWaitingApproval {
		t.Fatal("start reconciliation repeated work or lost the native review wait")
	}
}
