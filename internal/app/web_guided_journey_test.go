package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

// The owned canonical fixture applies the exact prepared Git result but loses
// its response. Native observation must reconcile that intent without resubmit.
type guidedWebPublicationFixture struct{ *serviceWorkEntryProvider }

func (p *guidedWebPublicationFixture) SubmitWorkflowIntegration(ctx context.Context, submission core.WorkflowSubmission) (core.WorkflowSubmissionResponse, error) {
	proof := submission.Guard.Git
	branch := "refs/heads/" + p.pr.BaseBranch

	command := exec.CommandContext(ctx, "git", "-C", submission.PreparedLane.Worktree, "push", "--force-with-lease="+branch+":"+proof.BaseCommit, p.remote, proof.ResultCommit+":"+branch)
	if output, err := command.CombinedOutput(); err != nil {
		return core.WorkflowSubmissionResponse{}, fmt.Errorf("fixture publication: %w: %s", err, output)
	}

	p.submissions = append(p.submissions, submission.ID)
	p.pr.State, p.pr.Merged, p.pr.ResultCommit = "closed", true, proof.ResultCommit

	return core.WorkflowSubmissionResponse{}, context.DeadlineExceeded
}

func guidedWebJourneyFixture(t *testing.T, loseAdmission bool) (webViewProvider, *guidedWorkHost, *guidedWebPublicationFixture) {
	t.Helper()

	provider, host, _ := guidedWebWorkFixture(t)
	factory := provider.serviceFactory

	base, ok := factory(t.Context()).workflowPullRequests.(*serviceWorkEntryProvider)
	if !ok {
		t.Fatal("missing owned canonical provider")
	}

	publication := &guidedWebPublicationFixture{base}

	var lose atomic.Bool
	lose.Store(loseAdmission)

	provider.serviceFactory = func(ctx context.Context) service {
		operations := factory(ctx)
		operations.workflowPullRequests = publication
		operations.afterGuidedAdmission = func(phase string) error {
			if phase == "admission" && lose.Swap(false) {
				return errors.New("lost admission response")
			}

			return nil
		}

		return operations
	}

	return provider, host, publication
}

// TestGuidedUseCaseWebJourney carries HTTP input through sealed authoring,
// uncertain admission, native controls and lost publication to verified results.
func TestGuidedUseCaseWebJourney(t *testing.T) {
	provider, host, publication := guidedWebJourneyFixture(t, true)
	handler := guidedWebHandler(t, provider)
	question := captureGuidedWebInput(t, provider, handler)
	replaceGuidedWebDocuments(t, provider.projectContext.ProjectDir, question)
	webGuidedInput(t, handler, question, "answer", "Keep the bounded workflow and original inputs.", http.StatusSeeOther)
	host.wait.Wait()

	operations := provider.service(t.Context())

	pending := readWebGuidedSession(t, operations, question.ID)
	if pending.Guided.PendingOperation == nil || pending.Guided.Work.Admission == nil || pending.Guided.Work.Invocation != nil {
		t.Fatal("admission response fault was not retained")
	}

	original := *pending.Guided.Work.Admission

	retry, ready := guidedProgressCommand(pending)
	if !ready {
		t.Fatal("missing original retry")
	}

	if response := guidedCaptureHTTP(t, handler, http.MethodPost, "/guided/actions", guidedWebCommandForm(retry)); response.Code != http.StatusSeeOther {
		t.Fatal("original admission retry failed", response.Body.String())
	}

	host.wait.Wait()

	started := readWebGuidedSession(t, operations, pending.ID)
	if !reflect.DeepEqual(started.Guided.Work.Admission, &original) {
		t.Fatal("retry replaced admitted scope or policy")
	}

	assertGuidedWorkWebStatus(t, provider, handler, started)

	controlled := guidedWebJourneyControls(t, provider, host, handler, started)
	completeGuidedWebPublication(t, provider, host, publication, controlled)

	completed := readWebGuidedSession(t, operations, started.ID)
	assertGuidedJourneyResult(t, handler, completed, publication)
}

func guidedWebJourneyControls(t *testing.T, provider webViewProvider, host *guidedWorkHost, handler http.Handler, session authoringcore.Session) authoringcore.Session {
	t.Helper()

	for _, action := range []string{"pause", "resume"} {
		page := guidedCaptureHTTP(t, handler, http.MethodGet, guidedUseCaseURL(session.ID, session.Namespace), nil)

		response := guidedCaptureHTTP(t, handler, http.MethodPost, "/guided/actions", guidedWebControlForm(t, page.Body.String(), action))
		if response.Code != http.StatusSeeOther {
			t.Fatal("journey control failed", action, response.Body.String())
		}

		host.wait.Wait()
	}

	current := readWebGuidedSession(t, provider.service(t.Context()), session.ID)
	webGuidedInput(t, handler, current, "idea", "Add an audit export in separately reviewed work.", http.StatusSeeOther)

	return readWebGuidedSession(t, provider.service(t.Context()), session.ID)
}

func completeGuidedWebPublication(t *testing.T, provider webViewProvider, host *guidedWorkHost, publication *guidedWebPublicationFixture, session authoringcore.Session) {
	t.Helper()

	operations := provider.service(t.Context())

	run := assertGuidedNativeAdmission(t, operations, session)
	if _, err := operations.ApproveWorkflowSlice(t.Context(), run.ID, run.Review.InputHash); err != nil {
		t.Fatal(err)
	}

	publishGuidedWebResult(t, operations, host, publication, run.ID)
}

func publishGuidedWebResult(t *testing.T, operations service, host *guidedWorkHost, publication *guidedWebPublicationFixture, id string) {
	t.Helper()

	if err := host.activate(operations, id); err != nil {
		t.Fatal(err)
	}

	host.wait.Wait()

	submission, err := operations.ReadWorkflowSubmission(t.Context(), id)
	if err != nil || submission.Response != "error_unknown" || len(publication.submissions) != 1 {
		t.Fatal("publication response fault was not retained", err)
	}

	assertGuidedUnknownPublication(t, operations, id)

	if err := host.activate(operations, id); err != nil {
		t.Fatal(err)
	}

	host.wait.Wait()
}

func assertGuidedUnknownPublication(t *testing.T, operations service, id string) {
	t.Helper()

	run, err := operations.ReadWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	observed, err := operations.InspectGuidedUseCase(t.Context(), run.Admission.Guided.RunID, run.Namespace)
	if err != nil || observed.Result == nil || observed.Result.Complete || len(observed.Result.Receipts) != 0 {
		t.Fatal("unknown publication claimed delivery before canonical reconciliation", err)
	}
}

func assertGuidedJourneyResult(t *testing.T, handler http.Handler, session authoringcore.Session, publication *guidedWebPublicationFixture) {
	t.Helper()

	if session.Guided.Stage != "results" || session.Guided.Result == nil || len(publication.submissions) != 1 || publication.creates != 1 {
		t.Fatal("journey did not reconcile one publication to native results", session.Guided.Stage)
	}

	original := guidedCaptureHTTP(t, handler, http.MethodGet, "/dumps?"+url.Values{"case": {session.Guided.CaseSlug}, "namespace": {session.Namespace}}.Encode(), nil)
	if original.Code != http.StatusOK || !strings.Contains(original.Body.String(), "Deliver a bounded workflow.") {
		t.Fatal("original dump missing from its normal screen")
	}

	page := guidedCaptureHTTP(t, handler, http.MethodGet, guidedUseCaseURL(session.ID, session.Namespace), nil)
	for _, text := range []string{"All admitted work has confirmed delivery", "0 remaining", "1 delivered", "View delivered changes", "Exit code: 0"} {
		if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), text) {
			t.Fatal("journey result omitted evidence", text, page.Body.String())
		}
	}
}
