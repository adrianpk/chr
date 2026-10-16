package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

func telegramAuthoringServiceFixture(t *testing.T, reply func(authoringModelInput) string) (service, *atomic.Int64) {
	t.Helper()

	var calls atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)

		var request struct {
			Input string `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}

		var input authoringModelInput
		if err := json.Unmarshal([]byte(request.Input), &input); err != nil {
			t.Error(err)
		}

		if err := json.NewEncoder(w).Encode(map[string]string{"output_text": reply(input)}); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(server.Close)

	operations := newTestService(t, t.TempDir())
	provider := core.NewConfiguredStructuredModel(true)
	provider.API = core.OpenAIQuizModelAdapter{APIKey: "fixture", BaseURL: server.URL, Client: server.Client()}
	operations.conversationModel = provider

	return operations, &calls
}

func telegramAuthoringServiceOutput(t *testing.T, body, question string) string {
	t.Helper()

	data, err := json.Marshal(authoringModelOutput{Body: body, Question: question})
	if err != nil {
		t.Fatal(err)
	}

	return string(data)
}

func telegramAuthoringServiceRequest() core.AuthoringContinuationRequest {
	return core.AuthoringContinuationRequest{Binding: core.WorkflowMutationBinding{RequestID: "service-capture", Operator: "maintainer"}, ProjectID: "service-project", Namespace: "original", Action: "create", Target: core.AuthoringArtifactKindDump, Topic: "Bounded request retries", Capture: "Preserve the retry ceiling and existing defaults."}
}

func telegramAuthoringServiceNext(previous core.AuthoringContinuationResult, id, action string, target core.AuthoringArtifactKind) core.AuthoringContinuationRequest {
	return core.AuthoringContinuationRequest{Binding: core.WorkflowMutationBinding{RequestID: id, Operator: "maintainer"}, ProjectID: previous.Session.ProjectID, Namespace: previous.Session.Namespace, SessionID: previous.Session.ID, ExpectedRevision: previous.Session.Revision, ExpectedSourceHash: previous.SourceHash, Action: core.AuthoringContinuationAction(action), Target: target}
}

// TestTelegramAuthoringServiceJourney generates real canonical dump/quiz/spec/plan
// artifacts through the configured HTTP model, answers a real prompt and refines
// an observed spec without invoking native work or changing namespace.
func TestTelegramAuthoringServiceJourney(t *testing.T) {
	operations, calls := telegramAuthoringServiceFixture(t, func(input authoringModelInput) string {
		if input.Action == "ask" {
			return telegramAuthoringServiceOutput(t, "", "Which timeout must remain unchanged?\nWhich limit applies?")
		}

		return telegramAuthoringServiceOutput(t, "## Requirements\n\nPreserve retries and defaults.\n\n## Validation\n\nCheck timeout behavior.", "")
	})

	first, err := operations.ContinueAuthoring(t.Context(), telegramAuthoringServiceRequest())
	if err != nil {
		t.Fatal(err)
	}

	asked, err := operations.ContinueAuthoring(t.Context(), telegramAuthoringServiceNext(first, "ask", "ask", core.AuthoringArtifactKindQuiz))
	if err != nil || asked.Question == nil {
		t.Fatalf("question=%+v err=%v", asked, err)
	}

	answer := telegramAuthoringServiceNext(asked, "answer", "answer", core.AuthoringArtifactKindQuiz)
	answer.Question = asked.Question
	answer.Instructions = "Keep the configured timeout and five attempts."

	answered, err := operations.ContinueAuthoring(t.Context(), answer)
	if err != nil {
		t.Fatal(err)
	}

	spec := telegramAuthoringServiceGenerate(t, operations, answered, "spec", "generate", core.AuthoringArtifactKindSpec)
	refined := telegramAuthoringServiceGenerate(t, operations, spec, "refine", "refine", core.AuthoringArtifactKindSpec)

	plan := telegramAuthoringServiceGenerate(t, operations, refined, "plan", "generate", core.AuthoringArtifactKindPlan)
	if calls.Load() != 5 || plan.Session.Namespace != "original" {
		t.Fatalf("calls=%d session=%+v", calls.Load(), plan.Session)
	}

	telegramAuthoringCheckArtifacts(t, operations, plan)
	telegramAuthoringServiceNoWork(t, operations)
}

func telegramAuthoringServiceGenerate(t *testing.T, operations service, previous core.AuthoringContinuationResult, id, action string, target core.AuthoringArtifactKind) core.AuthoringContinuationResult {
	t.Helper()

	result, err := operations.ContinueAuthoring(t.Context(), telegramAuthoringServiceNext(previous, id, action, target))
	if err != nil {
		t.Fatal(err)
	}

	return result
}

func telegramAuthoringCheckArtifacts(t *testing.T, operations service, result core.AuthoringContinuationResult) {
	t.Helper()

	snapshot, err := operations.ObserveAuthoring(t.Context(), result.Session.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(snapshot.Artifacts) != 5 || len(snapshot.Events) != 2 {
		t.Fatalf("snapshot=%+v", snapshot)
	}

	for _, artifact := range snapshot.Artifacts {
		if !strings.HasPrefix(artifact.Ref.Path, "ops/original/") || artifact.SourceHash == "" {
			t.Fatalf("noncanonical artifact: %+v", artifact)
		}

		if _, err := os.Stat(filepath.Join(operations.projectContext.ProjectDir, filepath.Dir(artifact.Ref.Path), "extension", "chr")); err != nil {
			t.Fatal(err)
		}

		if artifact.Ref.Kind == core.AuthoringArtifactKindSpec && !strings.Contains(artifact.Body, "source_request_id: \"refine\"") {
			t.Fatalf("missing refinement provenance: %s", artifact.Body)
		}
	}
}

// TestTelegramAuthoringServiceInvalid retains the prior canonical context when
// a real HTTP provider returns malformed, unknown or contradictory output.
func TestTelegramAuthoringServiceInvalid(t *testing.T) {
	tests := []struct{ name, output string }{
		{"unknown", `{"body":"content","question":"","rationale":"","tool":"exec"}`},
		{"missing", `{"body":"content"}`},
		{"null", `{"body":null,"question":"","rationale":""}`},
		{"contradictory", `{"body":"content","question":"Choose?","rationale":""}`},
		{"malformed", `bad json`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operations, calls := telegramAuthoringServiceFixture(t, func(input authoringModelInput) string {
				if input.Action == "create" {
					return telegramAuthoringServiceOutput(t, "Preserve defaults.", "")
				}

				return tt.output
			})

			before, err := operations.ContinueAuthoring(t.Context(), telegramAuthoringServiceRequest())
			if err != nil {
				t.Fatal(err)
			}

			request := telegramAuthoringServiceNext(before, "invalid", "generate", core.AuthoringArtifactKindSpec)
			if _, err := operations.ContinueAuthoring(t.Context(), request); err == nil {
				t.Fatal("invalid output applied")
			}

			after, err := operations.ObserveAuthoring(t.Context(), before.Session.ID)
			if err != nil || after.SourceHash != before.SourceHash || calls.Load() != 2 {
				t.Fatalf("prior context changed: %+v err=%v", after, err)
			}

			telegramAuthoringServiceNoWork(t, operations)
		})
	}
}

// TestTelegramAuthoringServiceObservedEdit rejects stale artifact observations and
// reuses original confirmed receipts after a later ordinary writer changes a body.
func TestTelegramAuthoringServiceObservedEdit(t *testing.T) {
	operations, calls := telegramAuthoringServiceFixture(t, func(authoringModelInput) string { return telegramAuthoringServiceOutput(t, "Preserve defaults.", "") })
	initial := telegramAuthoringServiceRequest()

	before, err := operations.ContinueAuthoring(t.Context(), initial)
	if err != nil {
		t.Fatal(err)
	}

	snapshot, err := operations.ObserveAuthoring(t.Context(), before.Session.ID)
	if err != nil {
		t.Fatal(err)
	}

	path := snapshot.Session.Outputs[0].Path
	if _, err := operations.SaveDump(DumpSaveRequest{RelativePath: path, Content: "Independently edited context"}); err != nil {
		t.Fatal(err)
	}

	request := telegramAuthoringServiceNext(before, "stale", "generate", core.AuthoringArtifactKindSpec)
	if _, err := operations.ContinueAuthoring(t.Context(), request); !errors.Is(err, core.ErrWorkflowConflict) {
		t.Fatalf("stale image err=%v", err)
	}

	replay, err := operations.ContinueAuthoring(t.Context(), initial)
	if err != nil || replay.SourceHash != before.SourceHash || calls.Load() != 1 {
		t.Fatalf("replay=%+v calls=%d err=%v", replay, calls.Load(), err)
	}

	content, err := os.ReadFile(filepath.Join(operations.projectContext.ProjectDir, path))
	if err != nil || string(content) != "Independently edited context\n" {
		t.Fatal("later edit overwritten", err)
	}
}

func telegramAuthoringServiceNoWork(t *testing.T, operations service) {
	t.Helper()

	executions, err := core.NewExecutionRepository(t.Context(), operations.env, testAssetsFS())
	if err != nil {
		t.Fatal(err)
	}

	runs, readErr := executions.ListProjectExecutions(t.Context(), operations.projectContext.ProjectID)
	if err := executions.Close(); err != nil {
		t.Fatal(err)
	}

	if readErr != nil || len(runs) != 0 {
		t.Fatalf("unexpected executions: %+v err=%v", runs, readErr)
	}

	for _, namespace := range []string{"default", "original"} {
		entries, err := os.ReadDir(filepath.Join(operations.projectContext.ProjectDir, "ops", namespace, "run"))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}

		if err != nil || len(entries) != 0 {
			t.Fatalf("unexpected native invocation artifacts: %v err=%v", entries, err)
		}
	}
}

// TestTelegramAuthoringServiceClarification persists a real generated question
// before accepting its answer, then generates the requested artifact from the
// canonical answer and records its transcript provenance without fake inputs.
func TestTelegramAuthoringServiceClarification(t *testing.T) {
	operations, calls := telegramAuthoringServiceFixture(t, func(input authoringModelInput) string {
		if input.Target == core.AuthoringArtifactKindSpec && len(input.Answers) == 0 {
			return telegramAuthoringServiceOutput(t, "", "Which behavior is required?")
		}

		return telegramAuthoringServiceOutput(t, "## Requirements\n\nPreserve the configured limits.", "")
	})

	initial, err := operations.ContinueAuthoring(t.Context(), telegramAuthoringServiceRequest())
	if err != nil {
		t.Fatal(err)
	}

	pending := telegramAuthoringServiceGenerate(t, operations, initial, "spec-question", "generate", core.AuthoringArtifactKindSpec)
	if pending.Question == nil || len(pending.Session.Outputs) != 1 {
		t.Fatalf("question produced an artifact: %+v", pending)
	}

	answer := telegramAuthoringServiceNext(pending, "spec-answer", "answer", core.AuthoringArtifactKindSpec)
	answer.Question = pending.Question
	answer.Instructions = "Preserve the configured limits."

	answered, err := operations.ContinueAuthoring(t.Context(), answer)
	if err != nil {
		t.Fatal(err)
	}

	generated := telegramAuthoringServiceGenerate(t, operations, answered, "spec-result", "generate", core.AuthoringArtifactKindSpec)

	snapshot, err := operations.ObserveAuthoring(t.Context(), generated.Session.ID)
	if err != nil {
		t.Fatal(err)
	}

	if calls.Load() != 3 || !strings.Contains(snapshot.Artifacts[len(snapshot.Artifacts)-1].Body, "transcript.md") {
		t.Fatalf("missing canonical answer provenance: %+v calls=%d", snapshot, calls.Load())
	}

	telegramAuthoringServiceNoWork(t, operations)
}
