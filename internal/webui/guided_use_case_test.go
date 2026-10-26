package webui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func (p guidedCaptureProvider) CaseWorkflowView(ctx context.Context, _, namespace string) (GuidedUseCaseView, error) {
	return p.GuidedUseCaseView(ctx, "", namespace)
}

type guidedCaptureProvider struct {
	staticDataProvider
	view GuidedUseCaseView
}

func (p guidedCaptureProvider) GuidedUseCaseView(context.Context, string, string) (GuidedUseCaseView, error) {
	p.view.Enabled = true

	return p.view, nil
}

// TestGuidedUseCaseCaptureView renders only current input and escapes captured content.
func TestGuidedUseCaseCaptureView(t *testing.T) {
	for _, test := range []struct {
		name   string
		view   GuidedUseCaseView
		want   string
		absent string
	}{
		{"start", GuidedUseCaseView{GuidedRunView: GuidedRunView{Namespace: "default", RequestID: "start"}}, `name="content"`, "Continue to quiz"},
		{"name", GuidedUseCaseView{GuidedRunView: GuidedRunView{RunID: "run", Stage: "name"}}, `name="content"`, "Continue to quiz"},
		{"dump", GuidedUseCaseView{GuidedRunView: GuidedRunView{RunID: "run", Stage: "dump", Capture: "<script>bad()</script>"}}, "Continue to quiz", "<script>bad()"},
		{"sealed", GuidedUseCaseView{GuidedRunView: GuidedRunView{RunID: "run", Stage: "quiz", Commit: "commit", CommittedDump: "Original capture"}}, "Original capture", "Continue to quiz"},
	} {
		t.Run(test.name, func(t *testing.T) {
			handler, err := NewHandler(HandlerOptions{Provider: guidedCaptureProvider{view: test.view}})
			if err != nil {
				t.Fatal(err)
			}

			response := httptest.NewRecorder()

			location := "/guided"
			if test.name == "sealed" {
				location = "/dumps?case=saved"
			}

			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, location, nil))

			if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), test.want) || strings.Contains(response.Body.String(), test.absent) {
				t.Fatal("incorrect current-stage form", response.Code, response.Body.String())
			}
		})
	}
}

// TestGuidedUseCaseCaptureBounds rejects duplicate parameters and oversized capture requests.
func TestGuidedUseCaseCaptureBounds(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{Provider: guidedCaptureProvider{}})
	if err != nil {
		t.Fatal(err)
	}

	for _, body := range []string{"action=guided.start&action=guided.input", "content=" + strings.Repeat("x", 73<<10)} {
		request := httptest.NewRequest(http.MethodPost, "/guided/actions", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		if response.Code != http.StatusBadRequest {
			t.Fatalf("unbounded or duplicate form accepted: %d", response.Code)
		}
	}
}

// TestGuidedUseCaseCaptureUnavailable keeps static navigation useful without commands.
func TestGuidedUseCaseCaptureUnavailable(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{})
	if err != nil {
		t.Fatal(err)
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/guided", nil))

	body := response.Body.String()
	if response.Code != http.StatusOK || !strings.Contains(body, "Workflows are unavailable") || strings.Contains(body, `action="/guided/actions"`) {
		t.Fatal("unavailable commands exposed", response.Code, body)
	}
}

// TestGuidedUseCaseQuizView binds the sole current input and retains the original pending reply on retry.
func TestGuidedUseCaseQuizView(t *testing.T) {
	for _, test := range []struct {
		name         string
		view         GuidedUseCaseView
		want, absent string
	}{
		{"question", GuidedUseCaseView{GuidedRunView: GuidedRunView{RunID: "run", Stage: "quiz"}, GuidedQuestionView: GuidedQuestionView{Quiz: &QuizAnswerView{Question: "What is the intended result?", QuestionID: "question-current", CanAnswer: true, Open: true, Started: true, Mode: "deep"}}, GuidedStepView: GuidedStepView{QuestionID: "question-current", InputDigest: "original-inputs"}}, `name="question_id" value="question-current"`, "Retry this step"},
		{"pending", GuidedUseCaseView{GuidedRunView: GuidedRunView{RunID: "run", Stage: "quiz"}, GuidedStepView: GuidedStepView{Question: "Superseded question", PendingAction: "answer", PendingContent: "<script>retained</script>", QuestionID: "question-original", InputDigest: "original-inputs"}}, `name="question_id" value="question-original"`, "Superseded question"},
	} {
		t.Run(test.name, func(t *testing.T) {
			handler, err := NewHandler(HandlerOptions{Provider: guidedCaptureProvider{view: test.view}})
			if err != nil {
				t.Fatal(err)
			}

			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/guided", nil))

			body := response.Body.String()
			if response.Code != http.StatusOK || !strings.Contains(body, test.want) || strings.Contains(body, test.absent) || strings.Contains(body, "<script>") {
				t.Fatal("incorrect current question or retry form", body)
			}
		})
	}
}

// TestGuidedUseCaseSpecView exposes the complete approved document without a routine approval prompt.
func TestGuidedUseCaseSpecView(t *testing.T) {
	view := GuidedUseCaseView{GuidedRunView: GuidedRunView{RunID: "run", Stage: "preparation"}, GuidedStepView: GuidedStepView{Blocker: "Configure native workflow version 1 before retrying work preparation."}, Artifacts: []GuidedUseCaseArtifact{{Stage: "specification", Commit: "immutable-commit", Reference: "ops/default/spec/original/artifact.md", Content: "Complete original approved specification <script>escaped</script>"}}}

	handler, err := NewHandler(HandlerOptions{Provider: guidedCaptureProvider{view: view}})
	if err != nil {
		t.Fatal(err)
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/specs?case=checkout", nil))

	body := response.Body.String()
	for _, text := range []string{"Complete original approved specification"} {
		if response.Code != http.StatusOK || !strings.Contains(body, text) {
			t.Fatal("missing approved source", text)
		}
	}

	if strings.Contains(body, "immutable-commit") || strings.Contains(body, "ops/default/spec/original") || strings.Contains(body, "<script>escaped") || strings.Contains(body, "Approve specification") || strings.Contains(body, `action="/guided/actions"`) {
		t.Fatal("unexpected approval form or unsafe original content")
	}
}
