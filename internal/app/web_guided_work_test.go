package app

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
	"forge.adrianpk.com/chr/internal/webui"
)

type guidedWebWorkModel struct {
	quiz fakeServiceQuizModel
	plan guidedAdmissionModel
}

func (model *guidedWebWorkModel) Invoke(ctx context.Context, request quizmodel.Request) (quizmodel.Response, error) {
	if request.Action == quizmodel.ActionGeneratePlan || request.Strategy == quizmodel.StrategyPlanDelivery {
		return model.plan.Invoke(ctx, request)
	}

	return model.quiz.Invoke(ctx, request)
}

func guidedWebWorkFixture(t *testing.T) (webViewProvider, *guidedWorkHost, *guidedWebWorkModel) {
	t.Helper()

	operations, _, _ := guidedAdmissionFixture(t)
	now := operations.utcNow()
	completion := serviceQuizModelResponse("complete", "complete", now)
	completion.Question.Required = false
	model := &guidedWebWorkModel{quiz: fakeServiceQuizModel{responses: []quizmodel.Response{serviceQuizModelResponse("first", "question-1", now), completion}}}
	operations.quizModel = model
	host := newGuidedWorkHost(t.Context())
	t.Cleanup(host.close)

	provider := webViewProvider{env: operations.env, assetsFS: operations.assetsFS, projectContext: operations.projectContext}
	provider.config.Web = core.DefaultWebConfig()
	provider.config.ProjectDir = operations.projectContext.ProjectDir
	provider.config.guidedHost = host
	provider.serviceFactory = func(ctx context.Context) service { copy := operations; copy.ctx = ctx; return copy }

	return provider, host, model
}

// TestGuidedUseCaseAdmissionWeb advances an eligible HTTP answer through sealed preparation and actual native work.
func TestGuidedUseCaseAdmissionWeb(t *testing.T) {
	provider, host, model := guidedWebWorkFixture(t)
	handler := guidedWebHandler(t, provider)
	current := captureGuidedWebInput(t, provider, handler)

	answer := guidedWebCommandForm(guidedQuizCommand(current, "answer", "answer", "Preserve the exact bounded behavior."))
	for range 2 {
		response := guidedCaptureHTTP(t, handler, http.MethodPost, "/guided/actions", answer)
		if response.Code != http.StatusSeeOther {
			t.Fatal("automatic work progression failed", response.Code, response.Body.String())
		}
	}

	host.wait.Wait()

	operations := provider.service(t.Context())
	started := readWebGuidedSession(t, operations, current.ID)

	run := assertGuidedNativeAdmission(t, operations, started)
	if len(run.Members) != 1 || serviceWorkTaskAttempts(run) != 1 || run.Review == nil || run.Review.State != "waiting_approval" || model.plan.calls != 1 || len(model.plan.base.requests) != 1 {
		t.Fatal("HTTP progression repeated work or skipped the native review boundary")
	}

	assertGuidedWorkWebStatus(t, provider, handler, started)
}

func assertGuidedWorkWebStatus(t *testing.T, provider webViewProvider, handler http.Handler, session core.AuthoringSession) {
	t.Helper()

	before, err := provider.GuidedUseCaseView(t.Context(), session.ID, session.Namespace)
	if err != nil {
		t.Fatal("native query lost the distinction between validation and delivery", err)
	}

	if !guidedWebWorkCountsMatch(before.Work) {
		t.Fatal("native query lost the distinction between validation and delivery")
	}

	for range 2 {
		page := guidedCaptureHTTP(t, handler, http.MethodGet, guidedUseCaseURL(session.ID, session.Namespace), nil)
		if page.Code != http.StatusOK {
			t.Fatal(page.Body.String())
		}

		assertGuidedWorkPage(t, page.Body.String(), session)
	}

	after, err := provider.GuidedUseCaseView(t.Context(), session.ID, session.Namespace)
	if err != nil || after.Work.Attempts != before.Work.Attempts || after.Revision != before.Revision {
		t.Fatal("inspection advanced native or authoring state", err)
	}
}

func guidedWebWorkCountsMatch(work *webui.GuidedUseCaseWork) bool {
	return work != nil && work.Validated == 1 && work.Confirmed == 0 && work.Remaining == 1
}

func assertGuidedWorkPage(t *testing.T, body string, session core.AuthoringSession) {
	t.Helper()

	for _, text := range []string{"1 of 1 tasks validated", "0 delivered", "1 remaining", "The work needs review", "Plan"} {
		if !strings.Contains(body, text) {
			t.Fatalf("missing owner-backed progress %q", text)
		}
	}

	if strings.Contains(body, "Execution has not started.") || strings.Contains(body, session.Guided.Work.Invocation.ID) {
		t.Fatal("Web hid native progress or exposed an opaque invocation identity")
	}
}
