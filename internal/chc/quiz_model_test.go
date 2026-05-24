package chc

import (
	"context"
	"testing"
)

func TestInvokeQuizModelValidatesRequestAndResponse(t *testing.T) {
	request := validQuizModelRequest()
	adapter := &fakeQuizModelAdapter{
		response: QuizModelResponse{
			ID:        "response-1",
			RequestID: request.ID,
			Provider:  "fake",
			Status:    QuizModelResponseReady,
			Question:  validQuizQuestion(),
			Candidates: []QuizCandidate{
				validQuizCandidate(),
			},
			RawText:   "question",
			CreatedAt: quizRuntimeTestTime(),
		},
	}

	response, err := InvokeQuizModel(context.Background(), adapter, request)
	if err != nil {
		t.Fatalf("InvokeQuizModel returned error: %v", err)
	}

	if response.Question.ID != "question-1" {
		t.Fatalf("question id = %q", response.Question.ID)
	}

	if len(response.Candidates) != 1 {
		t.Fatalf("candidates = %#v, want one candidate", response.Candidates)
	}
}

func TestInvokeQuizModelRejectsInvalidRequest(t *testing.T) {
	request := validQuizModelRequest()
	request.ContextSections = nil

	_, err := InvokeQuizModel(context.Background(), &fakeQuizModelAdapter{}, request)
	if err == nil {
		t.Fatal("InvokeQuizModel returned nil error")
	}
}

func TestInvokeQuizModelRejectsInvalidResponse(t *testing.T) {
	request := validQuizModelRequest()

	_, err := InvokeQuizModel(context.Background(), &fakeQuizModelAdapter{
		response: QuizModelResponse{
			ID:        "response-1",
			RequestID: request.ID,
			Provider:  "fake",
			Status:    QuizModelResponseReady,
			CreatedAt: quizRuntimeTestTime(),
		},
	}, request)
	if err == nil {
		t.Fatal("InvokeQuizModel returned nil error")
	}
}

func validQuizModelRequest() QuizModelRequest {
	state := validQuizRuntimeState()

	return QuizModelRequest{
		ID:           "quiz-model-request-1",
		Action:       QuizModelActionNextQuestion,
		Strategy:     "deep",
		Scope:        QuizScope{ProjectID: "project-1"},
		Mode:         QuizModeDeep,
		RuntimeState: state,
		ContextSections: []QuizModelContextSection{
			{Name: "project", Layer: "project", Authority: "project_metadata", Content: "Project context."},
		},
		OutputContract: QuizModelOutputContract{
			RequiredFields: []string{"id", "prompt", "topic", "depth"},
		},
		CreatedAt: quizRuntimeTestTime(),
	}
}

type fakeQuizModelAdapter struct {
	response QuizModelResponse
	err      error
	request  QuizModelRequest
}

func (a *fakeQuizModelAdapter) InvokeQuizModel(
	_ context.Context,
	request QuizModelRequest,
) (QuizModelResponse, error) {
	a.request = request
	if a.err != nil {
		return QuizModelResponse{}, a.err
	}

	return a.response, nil
}
