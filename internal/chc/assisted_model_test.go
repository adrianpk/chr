package chc

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestInvokeAssistedModelUsesAdapterBoundary(t *testing.T) {
	request := testAssistedModelRequest()
	response := testAssistedModelResponse(request)
	adapter := &FakeAssistedModelAdapter{Responses: []AssistedModelResponse{response}}

	got, err := InvokeAssistedModel(context.Background(), adapter, request)
	if err != nil {
		t.Fatalf("InvokeAssistedModel returned error: %v", err)
	}

	if got.ID != response.ID {
		t.Fatalf("response ID = %q, want %q", got.ID, response.ID)
	}
}

func TestInvokeAssistedModelClassifiesFailures(t *testing.T) {
	request := testAssistedModelRequest()
	request.ID = ""

	_, err := InvokeAssistedModel(context.Background(), &FakeAssistedModelAdapter{}, request)
	if err == nil {
		t.Fatal("InvokeAssistedModel returned nil error")
	}

	if ClassifyAssistedModelFailure(err) != AssistedModelFailureInvalidRequest {
		t.Fatalf("failure kind = %q", ClassifyAssistedModelFailure(err))
	}

	adapterErr := errors.New("provider unavailable")
	adapter := &FakeAssistedModelAdapter{Err: adapterErr}

	_, err = InvokeAssistedModel(context.Background(), adapter, testAssistedModelRequest())
	if err == nil {
		t.Fatal("InvokeAssistedModel returned nil adapter error")
	}

	if ClassifyAssistedModelFailure(err) != AssistedModelFailureAdapter {
		t.Fatalf("adapter failure kind = %q", ClassifyAssistedModelFailure(err))
	}
}

func TestInvokeAssistedModelClassifiesInvalidResponses(t *testing.T) {
	request := testAssistedModelRequest()
	response := testAssistedModelResponse(request)
	response.Question.ID = ""
	adapter := &FakeAssistedModelAdapter{Responses: []AssistedModelResponse{response}}

	_, err := InvokeAssistedModel(context.Background(), adapter, request)
	if err == nil {
		t.Fatal("InvokeAssistedModel returned nil error")
	}

	if ClassifyAssistedModelFailure(err) != AssistedModelFailureInvalidResponse {
		t.Fatalf("failure kind = %q", ClassifyAssistedModelFailure(err))
	}
}

func TestRecordAssistedModelInvocationWritesRequestAndResponseArtifacts(t *testing.T) {
	dir := t.TempDir()
	request := testAssistedModelRequest()
	response := testAssistedModelResponse(request)

	paths, err := PlanAssistedModelInvocationArtifacts(AuthoringSessionLayout(dir, request.SessionID), request)
	if err != nil {
		t.Fatalf("PlanAssistedModelInvocationArtifacts returned error: %v", err)
	}

	err = RecordAssistedModelInvocation(context.Background(), AssistedModelInvocationArtifacts{
		Paths:    paths,
		Request:  request,
		Response: response,
	})
	if err != nil {
		t.Fatalf("RecordAssistedModelInvocation returned error: %v", err)
	}

	requestContent := readAssistedModelTestFile(t, paths.RequestPath)
	responseContent := readAssistedModelTestFile(t, paths.ResponsePath)

	for _, want := range []string{"Assisted Model Request", "Build the next question", "Assisted Authoring Context"} {
		if !strings.Contains(requestContent, want) {
			t.Fatalf("request artifact missing %q:\n%s", want, requestContent)
		}
	}

	for _, want := range []string{"Assisted Model Response", "fake", "What else matters?"} {
		if !strings.Contains(responseContent, want) {
			t.Fatalf("response artifact missing %q:\n%s", want, responseContent)
		}
	}
}

func TestRecordAssistedModelInvocationClassifiesArtifactFailures(t *testing.T) {
	request := testAssistedModelRequest()
	response := testAssistedModelResponse(request)
	response.RequestID = "other-request"

	err := RecordAssistedModelInvocation(context.Background(), AssistedModelInvocationArtifacts{
		Paths: AssistedModelInvocationPaths{
			InvocationDir: filepath.Join(t.TempDir(), "model", request.ID),
			RequestPath:   "request.md",
			ResponsePath:  "response.md",
		},
		Request:  request,
		Response: response,
	})
	if err == nil {
		t.Fatal("RecordAssistedModelInvocation returned nil error")
	}

	if ClassifyAssistedModelFailure(err) != AssistedModelFailureArtifact {
		t.Fatalf("failure kind = %q", ClassifyAssistedModelFailure(err))
	}
}

func TestRenderAssistedModelArtifactsAreDeterministic(t *testing.T) {
	request := testAssistedModelRequest()
	response := testAssistedModelResponse(request)

	firstRequest := RenderAssistedModelRequest(request)

	secondRequest := RenderAssistedModelRequest(request)
	if firstRequest != secondRequest {
		t.Fatalf("request render is not deterministic:\n%s\n%s", firstRequest, secondRequest)
	}

	firstResponse := RenderAssistedModelResponse(response)

	secondResponse := RenderAssistedModelResponse(response)
	if firstResponse != secondResponse {
		t.Fatalf("response render is not deterministic:\n%s\n%s", firstResponse, secondResponse)
	}
}

func testAssistedModelRequest() AssistedModelRequest {
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)

	return AssistedModelRequest{
		ID:        "request-1",
		SessionID: "session-1",
		TurnID:    "turn-1",
		Intent:    AssistedAuthoringIntentDiscover,
		Target:    AuthoringTargetSpec,
		Context:   "# Assisted Authoring Context\n\nSession context.",
		Prompt:    "Build the next question.",
		CreatedAt: now,
	}
}

func testAssistedModelResponse(request AssistedModelRequest) AssistedModelResponse {
	return AssistedModelResponse{
		ID:        "response-1",
		RequestID: request.ID,
		SessionID: request.SessionID,
		TurnID:    request.TurnID,
		Provider:  "fake",
		Status:    AssistedModelResponseReady,
		Question: AssistedAuthoringQuestion{
			ID:        "question-1",
			Prompt:    "What else matters?",
			Rationale: "The topic needs one more constraint.",
			Required:  true,
		},
		RawText:   "Ask one follow-up question.",
		CreatedAt: request.CreatedAt.Add(time.Minute),
	}
}

func readAssistedModelTestFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	return string(content)
}
