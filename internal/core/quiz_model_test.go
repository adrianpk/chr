package core

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	harnesstrace "forge.adrianpk.com/chr/internal/core/harness/trace"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
	openaiquiz "forge.adrianpk.com/chr/internal/core/quiz/openai"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

func TestInvokeQuizModelValidatesRequestAndResponse(t *testing.T) {
	request := validQuizModelRequest()
	adapter := &fakeQuizModelAdapter{
		response: quizmodel.Response{
			ID:        "response-1",
			RequestID: request.ID,
			Provider:  "fake",
			Status:    quizmodel.ResponseReady,
			Question:  validQuizQuestion(),
			Candidates: []quizcore.Candidate{
				validQuizCandidate(),
			},
			RawText:   "question",
			CreatedAt: quizRuntimeTestTime(),
		},
	}

	response, err := quizmodel.Invoke(context.Background(), adapter, request)
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

	_, err := quizmodel.Invoke(context.Background(), &fakeQuizModelAdapter{}, request)
	if err == nil {
		t.Fatal("InvokeQuizModel returned nil error")
	}
}

func TestInvokeQuizModelRejectsInvalidResponse(t *testing.T) {
	request := validQuizModelRequest()

	_, err := quizmodel.Invoke(context.Background(), &fakeQuizModelAdapter{
		response: quizmodel.Response{
			ID:        "response-1",
			RequestID: request.ID,
			Provider:  "fake",
			Status:    quizmodel.ResponseReady,
			CreatedAt: quizRuntimeTestTime(),
		},
	}, request)
	if err == nil {
		t.Fatal("InvokeQuizModel returned nil error")
	}
}

func TestInvokeQuizModelAcceptsDraftSpecResponse(t *testing.T) {
	request := validQuizModelRequest()
	request.Action = quizmodel.ActionDraftSpec
	request.Strategy = "draft_spec"
	request.OutputContract.RequiredFields = []string{"title", "body", "status"}
	request.Payload.OutputContract = quizmodel.NewOutputContract(quizmodel.ActionDraftSpec, request.Mode)

	adapter := &fakeQuizModelAdapter{
		response: quizmodel.Response{
			ID:        "response-1",
			RequestID: request.ID,
			Provider:  "fake",
			Status:    quizmodel.ResponseReady,
			DraftSpec: generationcore.QuizDraftSpec{
				Scope:          quizcore.Scope{ProjectID: "project-1"},
				Title:          "Generated Spec",
				Body:           "Generated behavior.",
				SourceQuizPath: ".chr/default/quizzes/project-quiz.md",
				SourceRawPath:  ".chr/default/quizzes/project-quiz-raw.md",
				CreatedAt:      quizRuntimeTestTime(),
			},
			CreatedAt: quizRuntimeTestTime(),
		},
	}

	response, err := quizmodel.Invoke(context.Background(), adapter, request)
	if err != nil {
		t.Fatalf("InvokeQuizModel returned error: %v", err)
	}

	if response.DraftSpec.Title != "Generated Spec" {
		t.Fatalf("draft spec title = %q", response.DraftSpec.Title)
	}
}

func TestInvokeQuizModelAcceptsPlanOutlineResponse(t *testing.T) {
	request := validQuizModelRequest()
	request.Action = quizmodel.ActionGeneratePlanOutline
	request.Strategy = "plan_outline_generation"
	request.OutputContract.RequiredFields = quizmodel.NewOutputContract(
		quizmodel.ActionGeneratePlanOutline,
		request.Mode,
	).RequiredFields
	request.Payload.OutputContract = quizmodel.NewOutputContract(quizmodel.ActionGeneratePlanOutline, request.Mode)

	adapter := &fakeQuizModelAdapter{
		response: quizmodel.Response{
			ID:          "response-1",
			RequestID:   request.ID,
			Provider:    "fake",
			Status:      quizmodel.ResponseReady,
			PlanOutline: validGeneratedPlanOutline(request.CreatedAt),
			CreatedAt:   quizRuntimeTestTime(),
		},
	}

	response, err := quizmodel.Invoke(context.Background(), adapter, request)
	if err != nil {
		t.Fatalf("InvokeQuizModel returned error: %v", err)
	}

	if response.PlanOutline.Slices[0].ID != "Slice 1" {
		t.Fatalf("outline slice id = %q, want Slice 1", response.PlanOutline.Slices[0].ID)
	}
}

func TestInvokeQuizModelAcceptsPlanSliceDetailResponse(t *testing.T) {
	request := validQuizModelRequest()
	request.Action = quizmodel.ActionGeneratePlanSliceDetail
	request.Strategy = "plan_slice_detail_generation"
	request.OutputContract.RequiredFields = quizmodel.NewOutputContract(
		quizmodel.ActionGeneratePlanSliceDetail,
		request.Mode,
	).RequiredFields
	request.Payload.OutputContract = quizmodel.NewOutputContract(quizmodel.ActionGeneratePlanSliceDetail, request.Mode)

	adapter := &fakeQuizModelAdapter{
		response: quizmodel.Response{
			ID:          "response-1",
			RequestID:   request.ID,
			Provider:    "fake",
			Status:      quizmodel.ResponseReady,
			SliceDetail: validGeneratedPlanSliceDetail(request.CreatedAt),
			CreatedAt:   quizRuntimeTestTime(),
		},
	}

	response, err := quizmodel.Invoke(context.Background(), adapter, request)
	if err != nil {
		t.Fatalf("InvokeQuizModel returned error: %v", err)
	}

	if response.SliceDetail.SliceID != "Slice 1" {
		t.Fatalf("slice detail id = %q, want Slice 1", response.SliceDetail.SliceID)
	}
}

func TestQuizModelContextSectionsForPayloadMatchesRedWrightShape(t *testing.T) {
	request := validQuizModelRequest()

	sections, err := quizmodel.ContextSectionsForPayload(request.Payload)
	if err != nil {
		t.Fatalf("QuizModelContextSectionsForPayload returned error: %v", err)
	}

	names := []string{}
	for _, section := range sections {
		names = append(names, section.Name)
	}

	want := []string{"project", "quiz", "recent_entries", "current_question"}
	for index, name := range want {
		if names[index] != name {
			t.Fatalf("section names = %#v, want %#v", names, want)
		}
	}

	if sections[0].Authority != "project_context" {
		t.Fatalf("project authority = %q, want project_context", sections[0].Authority)
	}

	if sections[2].Authority != "quiz_context" {
		t.Fatalf("recent entries authority = %q, want quiz_context", sections[2].Authority)
	}
}

func TestQuizModelProviderPayloadIncludesNextQuestionContext(t *testing.T) {
	request := validQuizModelRequest()
	request.ContextSections = append(request.ContextSections, quizmodel.ContextSection{
		Name:      "project_dump",
		Layer:     "dump",
		Authority: ".chr/default/dumps/project-context.md",
		Content:   "Detailed project dump that should guide next-question generation.",
	})

	payload := quizmodel.ProviderPayload(request)

	contextSections, ok := payload["context_sections"].([]map[string]any)
	if !ok {
		t.Fatalf("context_sections = %#v, want context section payloads", payload["context_sections"])
	}

	if len(contextSections) == 0 {
		t.Fatal("context_sections is empty")
	}

	last := contextSections[len(contextSections)-1]
	if last["content"] != "Detailed project dump that should guide next-question generation." {
		t.Fatalf("last context section = %#v", last)
	}
}

func TestQuizModelProviderPayloadCarriesAcceptanceReadiness(t *testing.T) {
	request := validQuizModelRequest()
	request.Payload.AcceptanceReadiness = &generationcore.AcceptanceReadiness{
		Decision:   generationcore.AcceptanceReadinessNeedsMoreInfo,
		Confidence: 0.66,
		Reasons:    []string{"Need actor details."},
		Missing:    []generationcore.AcceptanceReadinessSlot{generationcore.AcceptanceReadinessSlotActor},
	}

	payload := quizmodel.ProviderPayload(request)

	if _, ok := payload["context_sections"]; !ok {
		t.Fatalf("next-question provider payload missing context_sections: %#v", payload)
	}

	readiness, ok := payload["acceptance_readiness"].(*generationcore.AcceptanceReadiness)
	if !ok {
		t.Fatalf("acceptance_readiness = %#v, want readiness payload", payload["acceptance_readiness"])
	}

	if readiness.Decision != generationcore.AcceptanceReadinessNeedsMoreInfo {
		t.Fatalf("decision = %q, want needs_more_info", readiness.Decision)
	}
}

func TestQuizModelProviderPayloadKeepsDraftSpecContextInProviderText(t *testing.T) {
	request := validQuizModelRequest()
	request.Action = quizmodel.ActionDraftSpec
	request.Strategy = "draft_spec"
	request.Payload.OutputContract = quizmodel.NewOutputContract(quizmodel.ActionDraftSpec, request.Mode)
	request.OutputContract.RequiredFields = request.Payload.OutputContract.RequiredFields
	request.ContextSections = append(request.ContextSections, quizmodel.ContextSection{
		Name:      "project_dump",
		Layer:     "dump",
		Authority: ".chr/default/dumps/project-context.md",
		Content:   "Detailed project dump for draft spec synthesis.",
	})

	payload := quizmodel.ProviderPayload(request)

	if _, ok := payload["context_sections"]; !ok {
		t.Fatalf("draft-spec provider payload omitted context_sections: %#v", payload)
	}
}

func TestQuizModelProviderPayloadKeepsPlanOutlineContextInProviderText(t *testing.T) {
	request := validQuizModelRequest()
	request.Action = quizmodel.ActionGeneratePlanOutline
	request.Strategy = "plan_outline_generation"
	request.Payload.OutputContract = quizmodel.NewOutputContract(quizmodel.ActionGeneratePlanOutline, request.Mode)
	request.OutputContract.RequiredFields = request.Payload.OutputContract.RequiredFields
	request.ContextSections = append(request.ContextSections, quizmodel.ContextSection{
		Name:      "spec",
		Layer:     "spec",
		Authority: ".chr/default/specs/blog.md",
		Content:   "Spec context for outline generation.",
	})

	payload := quizmodel.ProviderPayload(request)

	if _, ok := payload["context_sections"]; !ok {
		t.Fatalf("plan-outline provider payload omitted context_sections: %#v", payload)
	}
}

func TestOpenAIQuizModelAdapterRecordsTraceEvents(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(workspacecore.ActiveOperatorIDEnv, "adrian")
	t.Setenv(harnesstrace.LLMTraceEnabledEnv, "1")

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/responses" {
			t.Fatalf("path = %q, want /responses", request.URL.Path)
		}

		writer.Header().Set("Content-Type", "application/json")
		writeQuizModelHTTPResponse(t, writer, `{"output_text":"{\"question\":\"What should happen next?\",\"tags\":[\"workflow\"],\"follow_up_depth\":1,\"completion_recommended\":false}"}`)
	}))
	defer server.Close()

	model := openaiquiz.Adapter{
		Host:       QuizGenerationHost{},
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Model:      "test-model",
		ProjectDir: dir,
		Client:     server.Client(),
	}

	request := validQuizModelRequest()

	response, err := model.Invoke(context.Background(), request)
	if err != nil {
		t.Fatalf("InvokeQuizModel returned error: %v", err)
	}

	if response.Question.Prompt != "What should happen next?" {
		t.Fatalf("question = %q, want traced test response", response.Question.Prompt)
	}

	events := readLLMTraceEvents(t, dir)

	assertOpenAITraceEventsWithoutPayloads(t, events)
}

func TestOpenAIQuizModelAdapterRecordsTracePayloadsWhenEnabled(t *testing.T) {
	dir := t.TempDir()
	apiKey := "sk-header-only-test-secret"

	t.Setenv(workspacecore.ActiveOperatorIDEnv, "adrian")
	t.Setenv(harnesstrace.LLMTraceEnabledEnv, "1")
	t.Setenv(harnesstrace.LLMTracePayloadsEnv, "1")

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer "+apiKey {
			t.Error("missing configured authorization header")
		}

		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Error(err)
		}

		if strings.Contains(string(body), apiKey) {
			t.Error("credential leaked into provider payload")
		}

		writer.Header().Set("Content-Type", "application/json")
		writeQuizModelHTTPResponse(t, writer, `{"output_text":"{\"question\":\"What should happen next?\",\"tags\":[\"workflow\"],\"follow_up_depth\":1,\"completion_recommended\":false}"}`)
	}))
	defer server.Close()

	model := openaiquiz.Adapter{
		Host:       QuizGenerationHost{},
		APIKey:     apiKey,
		BaseURL:    server.URL,
		Model:      "test-model",
		ProjectDir: dir,
		Client:     server.Client(),
	}

	_, err := model.Invoke(context.Background(), validQuizModelRequest())
	if err != nil {
		t.Fatalf("InvokeQuizModel returned error: %v", err)
	}

	events := readLLMTraceEvents(t, dir)

	encoded, err := json.Marshal(events)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(encoded), apiKey) {
		t.Fatal("credential leaked into trace payloads")
	}

	for key, event := range map[string]map[string]any{"request_payload": events[0], "response_payload": events[1], "output_payload": events[1]} {
		if _, ok := event[key]; !ok {
			t.Fatalf("%s missing with payload opt-in: %#v", key, event)
		}
	}
}

func TestOpenAIQuizModelAdapterRecordsParseFailureTrace(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(workspacecore.ActiveOperatorIDEnv, "adrian")
	t.Setenv(harnesstrace.LLMTraceEnabledEnv, "1")

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writeQuizModelHTTPResponse(t, writer, `{"output_text":"not-json"}`)
	}))
	defer server.Close()

	model := openaiquiz.Adapter{
		Host:       QuizGenerationHost{},
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Model:      "test-model",
		ProjectDir: dir,
		Client:     server.Client(),
	}

	_, err := model.Invoke(context.Background(), validQuizModelRequest())
	if err == nil {
		t.Fatal("InvokeQuizModel returned nil error")
	}

	events := readLLMTraceEvents(t, dir)
	if len(events) != 2 {
		t.Fatalf("trace event count = %d, want 2: %#v", len(events), events)
	}

	if events[1]["status"] != "parse_error" {
		t.Fatalf("response status = %#v, want parse_error", events[1]["status"])
	}

	if events[1]["error"] == "" {
		t.Fatalf("missing parse error in response event: %#v", events[1])
	}
}

func TestOpenAIQuizModelAdapterRecordsProviderFailureTrace(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(workspacecore.ActiveOperatorIDEnv, "adrian")
	t.Setenv(harnesstrace.LLMTraceEnabledEnv, "1")

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		http.Error(writer, "provider failed", http.StatusBadGateway)
	}))
	defer server.Close()

	model := openaiquiz.Adapter{
		Host:       QuizGenerationHost{},
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Model:      "test-model",
		ProjectDir: dir,
		Client:     server.Client(),
	}

	_, err := model.Invoke(context.Background(), validQuizModelRequest())
	if err == nil {
		t.Fatal("InvokeQuizModel returned nil error")
	}

	events := readLLMTraceEvents(t, dir)
	if len(events) != 2 {
		t.Fatalf("trace event count = %d, want 2: %#v", len(events), events)
	}

	if events[1]["status"] != "error" {
		t.Fatalf("response status = %#v, want error", events[1]["status"])
	}

	if events[1]["http_status"] != float64(http.StatusBadGateway) {
		t.Fatalf("http_status = %#v, want 502", events[1]["http_status"])
	}

	if events[1]["error"] == "" {
		t.Fatalf("missing provider error in response event: %#v", events[1])
	}
}

func TestQuizModelPlanOutlineResponseFromOutput(t *testing.T) {
	request := validQuizModelRequest()
	request.Action = quizmodel.ActionGeneratePlanOutline
	request.Strategy = "plan_outline_generation"
	request.Payload.OutputContract = quizmodel.NewOutputContract(quizmodel.ActionGeneratePlanOutline, request.Mode)
	request.OutputContract.RequiredFields = request.Payload.OutputContract.RequiredFields

	response, err := quizmodel.ResponseFromOutput(request, []byte(`{
		"title":"Staged Plan Generation",
		"status":"planned",
		"summary":"Generate an outline first.",
		"assumptions":["Existing plans remain valid."],
		"open_questions":["Should staged generation be default?"],
		"risks":["More model calls."],
		"validation_strategy":"Validate schema and request construction.",
		"slices":[{
			"id":"Slice 1",
			"short_name":"outline generation contract",
			"objective":"Add outline contracts.",
			"rationale":"Keep the slice map coherent.",
			"dependencies":[],
			"branch_name":"feat/slice-1-outline-generation-contract",
			"pr_title":"feat(slice-1): outline generation contract",
			"report_path":".prj/reports/slices/staged-plan-generation/slice-1-outline-generation-contract.md",
			"complexity":"medium",
			"implementation_boundary":"Contracts only."
		}],
		"dependencies":[]
	}`), "openai")
	if err != nil {
		t.Fatalf("quizModelResponseFromOutput returned error: %v", err)
	}

	if response.PlanOutline.Title != "Staged Plan Generation" {
		t.Fatalf("plan outline title = %q", response.PlanOutline.Title)
	}
}

func TestValidateQuizModelPlanOutlinePayloadRejectsNonSequentialSlices(t *testing.T) {
	_, err := quizmodel.DecodePlanOutlinePayload([]byte(`{
		"title":"Staged Plan Generation",
		"status":"planned",
		"summary":"Generate an outline first.",
		"assumptions":[],
		"open_questions":[],
		"risks":[],
		"validation_strategy":"Validate schema and request construction.",
		"slices":[{
			"id":"Slice 2",
			"short_name":"outline generation contract",
			"objective":"Add outline contracts.",
			"rationale":"Keep the slice map coherent.",
			"dependencies":[],
			"branch_name":"feat/slice-1-outline-generation-contract",
			"pr_title":"feat(slice-1): outline generation contract",
			"report_path":".prj/reports/slices/staged-plan-generation/slice-1-outline-generation-contract.md",
			"complexity":"medium",
			"implementation_boundary":"Contracts only."
		}],
		"dependencies":[]
	}`))
	if err == nil {
		t.Fatal("DecodeQuizModelPlanOutlinePayload returned nil error")
	}
}

func TestQuizModelPlanSliceDetailResponseFromOutput(t *testing.T) {
	request := validQuizModelRequest()
	request.Action = quizmodel.ActionGeneratePlanSliceDetail
	request.Strategy = "plan_slice_detail_generation"
	request.Payload.OutputContract = quizmodel.NewOutputContract(quizmodel.ActionGeneratePlanSliceDetail, request.Mode)
	request.OutputContract.RequiredFields = request.Payload.OutputContract.RequiredFields

	response, err := quizmodel.ResponseFromOutput(request, []byte(`{
		"slice_id":"Slice 1",
		"short_name":"outline generation contract",
		"objective":"Add outline contracts.",
		"implementation_notes":["Keep this scoped to contracts."],
		"tasks":[{
			"id":"T1.1",
			"summary":"Add the model action.",
			"commit_hint":"feat(chr): add plan outline action"
		}],
		"acceptance_criteria":["The action validates."],
		"ticket_metadata":{
			"title":"Add outline contract",
			"synopsis":"Adds outline generation contracts.",
			"priority":"medium",
			"severity":"normal",
			"type":"feature",
			"tags":["planning"]
		},
		"branch_metadata":{"branch_name":"feat/slice-1-outline-generation-contract"},
		"pr_expectations":["Open one PR against dev."],
		"validation_expectations":["Run focused tests."],
		"dependencies":[],
		"non_goals":["Do not connect runtime orchestration."]
	}`), "openai")
	if err != nil {
		t.Fatalf("quizModelResponseFromOutput returned error: %v", err)
	}

	if response.SliceDetail.Tasks[0].ID != "T1.1" {
		t.Fatalf("slice detail task id = %q, want T1.1", response.SliceDetail.Tasks[0].ID)
	}
}

func TestValidateGeneratedPlanSliceDetailRejectsRenamedSlice(t *testing.T) {
	outline := validGeneratedPlanOutline(quizRuntimeTestTime())
	detail := validGeneratedPlanSliceDetail(quizRuntimeTestTime())
	detail.ShortName = "renamed slice"

	err := generationcore.ValidateGeneratedPlanSliceDetailAgainstOutline(detail, outline, "Slice 1")
	if err == nil {
		t.Fatal("ValidateGeneratedPlanSliceDetailAgainstOutline returned nil error")
	}
}

func TestQuizModelSpecGenerationInstructionsAvoidTopLevelBodyHeading(t *testing.T) {
	instructions := quizmodel.Instructions(quizmodel.Request{Action: quizmodel.ActionGenerateSpec})

	if !strings.Contains(instructions, "body without a top-level title heading") {
		t.Fatalf("spec generation instructions do not prevent duplicate title headings: %q", instructions)
	}
}

func TestQuizModelSpecGenerationInstructionsSeparateGroundedRequirementsFromProposals(t *testing.T) {
	instructions := quizmodel.Instructions(quizmodel.Request{Action: quizmodel.ActionGenerateSpec})

	for _, want := range []string{
		"Do not convert unresolved open questions into canonical requirements",
		"label any chosen default as proposed",
		"Do not introduce new commands, flags, APIs, exit codes, or persistence fields as requirements",
		"label useful ungrounded additions as proposed optional extensions",
	} {
		if !strings.Contains(instructions, want) {
			t.Fatalf("spec generation instructions missing %q in: %q", want, instructions)
		}
	}
}

func TestQuizModelSpecGenerationInstructionsFenceMultilineExamples(t *testing.T) {
	instructions := quizmodel.Instructions(quizmodel.Request{Action: quizmodel.ActionGenerateSpec})

	for _, want := range []string{
		"Fence every multi-line example in a Markdown code block",
		"Never emit raw front matter, YAML, JSON, command output, or code snippets as plain paragraphs or headings",
	} {
		if !strings.Contains(instructions, want) {
			t.Fatalf("spec generation instructions missing %q in: %q", want, instructions)
		}
	}
}

func TestQuizModelSpecGenerationInstructionsDoNotMentionFeatureGeneration(t *testing.T) {
	instructions := quizmodel.Instructions(quizmodel.Request{Action: quizmodel.ActionGenerateSpec})

	for _, forbidden := range []string{
		"Gherkin",
		"gherkin",
		"Feature/Scenario",
		"acceptance feature workflow",
	} {
		if strings.Contains(instructions, forbidden) {
			t.Fatalf("spec generation instructions contain %q in: %q", forbidden, instructions)
		}
	}
}

func TestQuizModelFeatureSchemaRequiresAllStrictProperties(t *testing.T) {
	schema := quizmodel.OutputSchema(quizmodel.Request{Action: quizmodel.ActionGenerateAcceptanceFeature}).Schema

	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatalf("required = %#v, want string list", schema["required"])
	}

	for _, want := range []string{"status", "body", "rejection_reason", "clarification_prompts"} {
		if !slices.Contains(required, want) {
			t.Fatalf("required = %#v, missing %q", required, want)
		}
	}
}

func TestQuizModelFeatureInstructionsRequireEmptyFallbackFields(t *testing.T) {
	instructions := quizmodel.Instructions(quizmodel.Request{Action: quizmodel.ActionGenerateAcceptanceFeature})

	for _, want := range []string{
		"Always include status, body, rejection_reason, and clarification_prompts",
		"Use an empty string for body or rejection_reason when not applicable",
		"an empty array for clarification_prompts when not needed",
	} {
		if !strings.Contains(instructions, want) {
			t.Fatalf("feature generation instructions missing %q in: %q", want, instructions)
		}
	}
}

func TestQuizModelFeatureInstructionsDescribeTraceabilityAndRoundAnswers(t *testing.T) {
	instructions := quizmodel.Instructions(quizmodel.Request{Action: quizmodel.ActionGenerateAcceptanceFeature})

	for _, want := range []string{
		"@spec:",
		"@requirement:",
		"@acceptance-criterion:",
		"@behavior:",
		"acceptance_feature_round context section includes answers",
		"treat those answers as resolving their own prompts",
	} {
		if !strings.Contains(instructions, want) {
			t.Fatalf("feature generation instructions missing %q in: %q", want, instructions)
		}
	}
}

func TestQuizModelFeatureInstructionsRejectTrivialHappyPathCoverage(t *testing.T) {
	instructions := quizmodel.Instructions(quizmodel.Request{Action: quizmodel.ActionGenerateAcceptanceFeature})

	for _, want := range []string{
		"trivial hardcoded happy-path implementation",
		"more than one behavior path",
		"one normal success path",
		"rejection, failure, boundary, terminal, or no-change path",
		"Return status needs_more_info instead of generated",
	} {
		if !strings.Contains(instructions, want) {
			t.Fatalf("feature generation instructions missing %q in: %q", want, instructions)
		}
	}
}

func TestQuizModelSpecGenerationInstructionsDoNotOverfitPathPolicy(t *testing.T) {
	instructions := quizmodel.Instructions(quizmodel.Request{Action: quizmodel.ActionGenerateSpec})

	for _, forbidden := range []string{
		"Do not propose arbitrary domain directories",
		"Do not invent new chr persistence locations",
	} {
		if strings.Contains(instructions, forbidden) {
			t.Fatalf("spec generation instructions contain overfitted path policy %q in: %q", forbidden, instructions)
		}
	}
}

func TestQuizModelQuestionInstructionsMatchRedWrightContextBoundary(t *testing.T) {
	instructions := []string{
		quizmodel.Instructions(quizmodel.Request{Mode: quizcore.ModeDeep}),
		quizmodel.Instructions(quizmodel.Request{Mode: quizcore.ModeFast}),
	}

	for _, instruction := range instructions {
		assertQuizModelQuestionInstructionBoundary(t, instruction)
	}
}

func assertQuizModelQuestionInstructionBoundary(t *testing.T, instruction string) {
	t.Helper()

	if strings.Contains(instruction, "selected context sections") {
		t.Fatalf("question instructions mention selected context sections: %q", instruction)
	}

	for _, want := range []string{
		"service has already selected the quiz scope and routing state",
		"functional analyst discovering requirements",
		"routing_state as binding conductor context",
		"one specific operational aspect",
		"one area is becoming overrepresented",
		"acceptance_readiness",
		"do not force Gherkin terminology",
	} {
		if !strings.Contains(instruction, want) {
			t.Fatalf("question instructions missing %q in: %q", want, instruction)
		}
	}
}

func TestDecodeQuizModelQuestionPayloadEnforcesRedWrightContract(t *testing.T) {
	payload, err := quizmodel.DecodeQuestionPayload([]byte(`{
		"question": "What should editorial review clarify before publishing?",
		"tags": ["Editorial Review", "review"],
		"follow_up_depth": 12,
		"completion_recommended": false
	}`))
	if err != nil {
		t.Fatalf("DecodeQuizModelQuestionPayload returned error: %v", err)
	}

	if payload.Tags[0] != "editorial-review" {
		t.Fatalf("normalized tag = %q, want editorial-review", payload.Tags[0])
	}

	if payload.FollowUpDepth != quizcore.MaxFollowUpDepth {
		t.Fatalf("follow up depth = %d, want clamp to %d", payload.FollowUpDepth, quizcore.MaxFollowUpDepth)
	}
}

func TestDecodeQuizModelQuestionPayloadRejectsMalformedOutput(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name: "extra field",
			content: `{
				"question": "What should editorial review clarify?",
				"tags": ["review"],
				"follow_up_depth": 0,
				"completion_recommended": false,
				"extra": "not allowed"
			}`,
		},
		{
			name: "secret prompt",
			content: `{
				"question": "What API key should the workflow collect?",
				"tags": ["review"],
				"follow_up_depth": 0,
				"completion_recommended": false
			}`,
		},
		{
			name: "wrong tags type",
			content: `{
				"question": "What should editorial review clarify?",
				"tags": "review",
				"follow_up_depth": 0,
				"completion_recommended": false
			}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := quizmodel.DecodeQuestionPayload([]byte(test.content))
			if err == nil {
				t.Fatal("DecodeQuizModelQuestionPayload returned nil error")
			}
		})
	}
}

func TestDecodeQuizModelPlanTrackingPayloadRequiresTicketSynopsis(t *testing.T) {
	payload, err := quizmodel.DecodePlanTrackingPayload([]byte(`{
		"title": "Tracking",
		"body": "Track implementation slices.",
		"status": "tracking",
		"tickets": [
			{
				"slice_id": "Slice 1",
				"title": "Implement foundation",
				"synopsis": "Build the implementation foundation.",
				"summary": "Implement the foundational work.",
				"body": "## Tasks\n\n- Implement the foundation.",
				"priority": "normal",
				"severity": "normal",
				"type": "task",
				"tags": ["foundation"],
				"dependencies": []
			}
		]
	}`))
	if err != nil {
		t.Fatalf("DecodeQuizModelPlanTrackingPayload returned error: %v", err)
	}

	if payload.Tickets[0].Synopsis != "Build the implementation foundation." {
		t.Fatalf("ticket synopsis = %q, want generated synopsis", payload.Tickets[0].Synopsis)
	}
}

func TestDecodeQuizModelPlanTrackingPayloadRejectsMissingTicketSynopsis(t *testing.T) {
	_, err := quizmodel.DecodePlanTrackingPayload([]byte(`{
		"title": "Tracking",
		"body": "Track implementation slices.",
		"status": "tracking",
		"tickets": [
			{
				"slice_id": "Slice 1",
				"title": "Implement foundation",
				"summary": "Implement the foundational work.",
				"body": "## Tasks\n\n- Implement the foundation.",
				"priority": "normal",
				"severity": "normal",
				"type": "task",
				"tags": ["foundation"],
				"dependencies": []
			}
		]
	}`))
	if err == nil {
		t.Fatal("DecodeQuizModelPlanTrackingPayload returned nil error")
	}
}

func TestDecodeQuizModelFastQuestionPayloadValidatesCandidates(t *testing.T) {
	payload, err := quizmodel.DecodeFastQuestionPayload([]byte(`{
		"questions": [
			{
				"question": "What should editorial review clarify first?",
				"tags": ["review"],
				"follow_up_depth": 0,
				"completion_recommended": false
			},
			{
				"question": "Which publishing gate depends on that review?",
				"tags": ["publishing"],
				"follow_up_depth": 1,
				"completion_recommended": false
			}
		]
	}`))
	if err != nil {
		t.Fatalf("DecodeQuizModelFastQuestionPayload returned error: %v", err)
	}

	if len(payload.Questions) != 2 {
		t.Fatalf("questions = %#v, want two", payload.Questions)
	}
}

func TestNewQuizModelPromptIntentMatchesRedWrightShape(t *testing.T) {
	intent := quizmodel.NewPromptIntent(quizmodel.QuestionPayload{
		Question:              "What should editorial review clarify?",
		Tags:                  []string{"review"},
		FollowUpDepth:         1,
		CompletionRecommended: false,
	}, map[string]any{"entry_count": 2})

	if intent.Kind != "quiz_question" {
		t.Fatalf("kind = %q, want quiz_question", intent.Kind)
	}

	if intent.ExpectedInputShape != "free_text_or_control" {
		t.Fatalf("input shape = %q, want free_text_or_control", intent.ExpectedInputShape)
	}

	if intent.Metadata["follow_up_depth"] != 1 {
		t.Fatalf("metadata = %#v, want follow_up_depth 1", intent.Metadata)
	}
}

func TestQuizModelResponseFromDeepOutput(t *testing.T) {
	request := validQuizModelRequest()
	request.RuntimeState.ActiveQuestion = nil
	request.RuntimeState.Trail = nil

	response, err := quizmodel.ResponseFromOutput(request, []byte(`{
		"question": "What should editorial review clarify before publishing?",
		"tags": ["Editorial Review"],
		"follow_up_depth": 2,
		"completion_recommended": false
	}`), "fixture-provider")
	if err != nil {
		t.Fatalf("quizModelResponseFromOutput returned error: %v", err)
	}

	if response.Provider != "fixture-provider" {
		t.Fatalf("provider = %q", response.Provider)
	}

	if response.Question.Prompt != "What should editorial review clarify before publishing?" {
		t.Fatalf("question = %q", response.Question.Prompt)
	}

	if response.Question.Depth != 1 {
		t.Fatalf("depth = %d, want 1", response.Question.Depth)
	}

	if response.Question.Topic != "editorial-review" {
		t.Fatalf("topic = %q, want editorial-review", response.Question.Topic)
	}
}

func TestQuizModelResponseFromDeepOutputAfterSkippedEntry(t *testing.T) {
	request := validQuizModelRequest()
	request.Payload.RecentEntries = append(request.Payload.RecentEntries, quizmodel.EntryPayload{
		Question: "Which publishing branch should we avoid?",
		Status:   string(quizcore.EntryStatusSkipped),
		Tags:     []string{"publishing"},
	})
	request.Payload.Quiz.EntryCount = len(request.Payload.RecentEntries)

	response, err := quizmodel.ResponseFromOutput(request, []byte(`{
		"question": "What should the next useful publishing checkpoint clarify?",
		"tags": ["publishing"],
		"follow_up_depth": 0,
		"completion_recommended": false
	}`), "openai")
	if err != nil {
		t.Fatalf("quizModelResponseFromOutput returned error: %v", err)
	}

	if response.Question.Prompt != "What should the next useful publishing checkpoint clarify?" {
		t.Fatalf("question = %q", response.Question.Prompt)
	}

	if response.Question.Topic != "publishing" {
		t.Fatalf("topic = %q, want publishing", response.Question.Topic)
	}

	if response.Question.Required != true {
		t.Fatalf("required = %v, want true", response.Question.Required)
	}
}

func TestQuizModelResponseFromDeepOutputCarriesCompletionRecommendation(t *testing.T) {
	request := validQuizModelRequest()

	response, err := quizmodel.ResponseFromOutput(request, []byte(`{
		"question": "Is there anything else this quiz must capture before drafting?",
		"tags": ["completion"],
		"follow_up_depth": 1,
		"completion_recommended": true
	}`), "openai")
	if err != nil {
		t.Fatalf("quizModelResponseFromOutput returned error: %v", err)
	}

	if response.Question.Required {
		t.Fatal("question required = true, want false for completion recommendation")
	}

	if !strings.Contains(response.Question.Rationale, `"completion_recommended":true`) {
		t.Fatalf("rationale = %q, want completion recommendation metadata", response.Question.Rationale)
	}
}

func TestQuizModelResponseFromFastOutputBuffersCandidates(t *testing.T) {
	request := validQuizModelRequest()
	request.Mode = quizcore.ModeFast
	request.Strategy = quizmodel.FastStrategy
	request.Payload.Quiz.Mode = string(quizcore.ModeFast)
	request.Payload.OutputContract = quizmodel.NewOutputContract(quizmodel.ActionNextQuestion, quizcore.ModeFast)

	response, err := quizmodel.ResponseFromOutput(request, []byte(`{
		"questions": [
			{
				"question": "What should editorial review clarify first?",
				"tags": ["review"],
				"follow_up_depth": 0,
				"completion_recommended": false
			},
			{
				"question": "Which publishing gate depends on that review?",
				"tags": ["publishing"],
				"follow_up_depth": 1,
				"completion_recommended": false
			}
		]
	}`), "openai")
	if err != nil {
		t.Fatalf("quizModelResponseFromOutput returned error: %v", err)
	}

	if response.Question.Prompt != "What should editorial review clarify first?" {
		t.Fatalf("question = %q", response.Question.Prompt)
	}

	if len(response.Candidates) != 1 {
		t.Fatalf("candidates = %#v, want one", response.Candidates)
	}

	if response.Candidates[0].Depth != 1 {
		t.Fatalf("candidate depth = %d, want 1", response.Candidates[0].Depth)
	}
}

func TestQuizModelResponseFromDraftSpecOutput(t *testing.T) {
	request := validQuizModelRequest()
	request.Action = quizmodel.ActionDraftSpec
	request.Strategy = "draft_spec"
	request.Payload.OutputContract = quizmodel.NewOutputContract(quizmodel.ActionDraftSpec, request.Mode)
	request.OutputContract.RequiredFields = request.Payload.OutputContract.RequiredFields

	response, err := quizmodel.ResponseFromOutput(request, []byte(`{
		"title": "Editorial Review Spec",
		"body": "The workflow must preserve editorial review.",
		"status": "draft"
	}`), "openai")
	if err != nil {
		t.Fatalf("quizModelResponseFromOutput returned error: %v", err)
	}

	if response.DraftSpec.Title != "Editorial Review Spec" {
		t.Fatalf("title = %q", response.DraftSpec.Title)
	}
}

func TestQuizModelResponseFromGeneratedSpecOutputRejectsFeatureField(t *testing.T) {
	request := validQuizModelRequest()
	request.Action = quizmodel.ActionGenerateSpec
	request.Strategy = "spec_generation"
	request.Payload.OutputContract = quizmodel.NewOutputContract(quizmodel.ActionGenerateSpec, request.Mode)
	request.OutputContract.RequiredFields = request.Payload.OutputContract.RequiredFields

	_, err := quizmodel.ResponseFromOutput(request, []byte(`{
		"title": "Editorial Review Spec",
		"body": "The workflow must preserve editorial review.",
		"status": "draft",
		"feature": {
			"status": "generated",
			"body": "@spec:editorial-review-spec\nFeature: Editorial review\n\n  Scenario: Review is preserved\n    Given an article draft exists\n    When the editor publishes it\n    Then the review status is preserved\n"
		}
	}`), "openai")
	if err == nil {
		t.Fatal("quizModelResponseFromOutput returned nil error")
	}
}

func TestQuizModelResponseFromGeneratedSpecOutputAllowsGherkinExampleInBody(t *testing.T) {
	request := validQuizModelRequest()
	request.Action = quizmodel.ActionGenerateSpec
	request.Strategy = "spec_generation"
	request.Payload.OutputContract = quizmodel.NewOutputContract(quizmodel.ActionGenerateSpec, request.Mode)
	request.OutputContract.RequiredFields = request.Payload.OutputContract.RequiredFields

	response, err := quizmodel.ResponseFromOutput(request, []byte("{\n"+
		`"title": "Editorial Review Spec",`+"\n"+
		`"body": "Acceptance Criteria\n\n`+"```gherkin"+`\nFeature: Editorial review\n`+"```"+`\n",`+"\n"+
		`"status": "draft"`+"\n"+
		"}"), "openai")
	if err != nil {
		t.Fatalf("quizModelResponseFromOutput returned error: %v", err)
	}

	if !strings.Contains(response.DraftSpec.Body, "Feature: Editorial review") {
		t.Fatalf("body = %q, want Gherkin example preserved", response.DraftSpec.Body)
	}
}

func TestValidateQuizModelResponseAcceptsDraftSpecBeforeServiceProvenance(t *testing.T) {
	response := quizmodel.Response{
		ID:        "response-1",
		RequestID: "request-1",
		Provider:  "fake",
		Status:    quizmodel.ResponseReady,
		DraftSpec: generationcore.QuizDraftSpec{
			Scope:     quizcore.Scope{ProjectID: "project-1"},
			Title:     "Draft Spec",
			Body:      "Draft body.",
			CreatedAt: quizRuntimeTestTime(),
		},
		CreatedAt: quizRuntimeTestTime(),
	}

	err := quizmodel.ValidateResponse(response)
	if err != nil {
		t.Fatalf("ValidateQuizModelResponse returned error: %v", err)
	}
}

func validQuizModelRequest() quizmodel.Request {
	state := validQuizRuntimeState()
	payload := quizmodel.Payload{
		Project: quizmodel.ProjectPayload{
			ID:          "project-1",
			Name:        "Project",
			Description: "Project context.",
		},
		Quiz: quizmodel.QuizPayload{
			ID:         state.ID,
			EntryCount: 1,
			Mode:       string(quizcore.ModeDeep),
		},
		RecentEntries: []quizmodel.EntryPayload{
			{
				Question: "What should this workflow preserve?",
				Answer:   "The workflow should preserve review.",
				Status:   string(quizcore.EntryStatusAnswered),
				Tags:     []string{"workflow"},
			},
		},
		CurrentQuestion: quizmodel.CurrentQuestionPayload{
			Question: "What should this workflow preserve?",
			Tags:     []string{"workflow"},
			PromptIntent: map[string]any{
				"metadata": map[string]any{"follow_up_depth": 0},
			},
		},
		OutputContract: quizmodel.NewOutputContract(quizmodel.ActionNextQuestion, quizcore.ModeDeep),
	}

	return quizmodel.Request{
		ID:           "quiz-model-request-1",
		Action:       quizmodel.ActionNextQuestion,
		Strategy:     "deep",
		Scope:        quizcore.Scope{ProjectID: "project-1"},
		Mode:         quizcore.ModeDeep,
		RuntimeState: state,
		Payload:      payload,
		ContextSections: []quizmodel.ContextSection{
			{Name: "project", Layer: "project", Authority: "project_metadata", Content: "Project context."},
		},
		OutputContract: quizmodel.OutputContract{
			RequiredFields: []string{"id", "prompt", "topic", "depth"},
		},
		CreatedAt: quizRuntimeTestTime(),
	}
}

func validGeneratedPlanOutline(createdAt time.Time) generationcore.GeneratedPlanOutline {
	return generationcore.GeneratedPlanOutline{
		Scope:              quizcore.Scope{ProjectID: "project-1"},
		Title:              "Staged Plan Generation",
		Status:             "planned",
		Summary:            "Generate an outline first.",
		Assumptions:        []string{"Existing plans remain valid."},
		OpenQuestions:      []string{"Should staged generation be default?"},
		Risks:              []string{"More model calls."},
		ValidationStrategy: "Validate schema and request construction.",
		Slices: []generationcore.GeneratedPlanOutlineSlice{
			{
				ID:                     "Slice 1",
				ShortName:              "outline generation contract",
				Objective:              "Add outline contracts.",
				Rationale:              "Keep the slice map coherent.",
				BranchName:             "feat/slice-1-outline-generation-contract",
				PRTitle:                "feat(slice-1): outline generation contract",
				ReportPath:             ".prj/reports/slices/staged-plan-generation/slice-1-outline-generation-contract.md",
				Complexity:             "medium",
				ImplementationBoundary: "Contracts only.",
			},
		},
		CreatedAt: createdAt,
	}
}

func validGeneratedPlanSliceDetail(createdAt time.Time) generationcore.GeneratedPlanSliceDetail {
	return generationcore.GeneratedPlanSliceDetail{
		Scope:               quizcore.Scope{ProjectID: "project-1"},
		SliceID:             "Slice 1",
		ShortName:           "outline generation contract",
		Objective:           "Add outline contracts.",
		ImplementationNotes: []string{"Keep this scoped to contracts."},
		Tasks: []generationcore.GeneratedPlanSliceTask{
			{
				ID:         "T1.1",
				Summary:    "Add the model action.",
				CommitHint: "feat(chr): add plan outline action",
			},
		},
		AcceptanceCriteria: []string{"The action validates."},
		TicketMetadata: generationcore.GeneratedPlanSliceTicketMetadata{
			Title:    "Add outline contract",
			Synopsis: "Adds outline generation contracts.",
			Priority: "medium",
			Severity: "normal",
			Type:     "feature",
			Tags:     []string{"planning"},
		},
		BranchMetadata: generationcore.GeneratedPlanSliceBranchMetadata{
			BranchName: "feat/slice-1-outline-generation-contract",
		},
		PRExpectations:         []string{"Open one PR against dev."},
		ValidationExpectations: []string{"Run focused tests."},
		NonGoals:               []string{"Do not connect runtime orchestration."},
		CreatedAt:              createdAt,
	}
}

func readLLMTraceEvents(t *testing.T, projectDir string) []map[string]any {
	t.Helper()

	path := filepath.Join(projectDir, workspacecore.DirectoryName, workspacecore.UserDirectoryName, "log", "llm", "llm.ndjson")

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read LLM trace file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")

	events := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		var event map[string]any

		err = json.Unmarshal([]byte(line), &event)
		if err != nil {
			t.Fatalf("decode LLM trace event: %v", err)
		}

		events = append(events, event)
	}

	return events
}

func writeQuizModelHTTPResponse(t *testing.T, writer http.ResponseWriter, content string) {
	t.Helper()

	_, err := writer.Write([]byte(content))
	if err != nil {
		t.Fatalf("write HTTP response: %v", err)
	}
}

func assertOpenAITraceEventsWithoutPayloads(t *testing.T, events []map[string]any) {
	t.Helper()

	if len(events) != 2 {
		t.Fatalf("trace event count = %d, want 2: %#v", len(events), events)
	}

	requestEvent := events[0]
	responseEvent := events[1]

	assertQuizModelTraceField(t, requestEvent, "event", "llm_request")
	assertQuizModelTraceField(t, responseEvent, "event", "llm_response")
	assertQuizModelTraceField(t, responseEvent, "status", "ready")
	assertQuizModelTraceHasField(t, requestEvent, "llm_call_id")
	assertQuizModelTraceHasField(t, responseEvent, "response_id")
	assertQuizModelTraceMatchingCallID(t, requestEvent, responseEvent)
	assertQuizModelTraceOmitsField(t, requestEvent, "request_payload")
	assertQuizModelTraceOmitsField(t, responseEvent, "response_payload")
}

func assertQuizModelTraceField(t *testing.T, event map[string]any, key string, want any) {
	t.Helper()

	if event[key] != want {
		t.Fatalf("trace %s = %#v, want %#v in %#v", key, event[key], want, event)
	}
}

func assertQuizModelTraceHasField(t *testing.T, event map[string]any, key string) {
	t.Helper()

	if event[key] == "" {
		t.Fatalf("missing %s in trace event: %#v", key, event)
	}
}

func assertQuizModelTraceMatchingCallID(t *testing.T, requestEvent, responseEvent map[string]any) {
	t.Helper()

	if requestEvent["llm_call_id"] != responseEvent["llm_call_id"] {
		t.Fatalf("trace call ids differ: request=%#v response=%#v", requestEvent["llm_call_id"], responseEvent["llm_call_id"])
	}
}

func assertQuizModelTraceOmitsField(t *testing.T, event map[string]any, key string) {
	t.Helper()

	if _, ok := event[key]; ok {
		t.Fatalf("trace event captured %s: %#v", key, event)
	}
}

type fakeQuizModelAdapter struct {
	response quizmodel.Response
	err      error
	request  quizmodel.Request
}

func (a *fakeQuizModelAdapter) Invoke(_ context.Context, request quizmodel.Request) (quizmodel.Response, error) {
	a.request = request
	if a.err != nil {
		return quizmodel.Response{}, a.err
	}

	return a.response, nil
}
