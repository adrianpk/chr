package chc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestInvokeQuizModelAcceptsDraftSpecResponse(t *testing.T) {
	request := validQuizModelRequest()
	request.Action = QuizModelActionDraftSpec
	request.Strategy = "draft_spec"
	request.OutputContract.RequiredFields = []string{"title", "body", "status"}
	request.Payload.OutputContract = NewQuizModelOutputContract(QuizModelActionDraftSpec, request.Mode)

	adapter := &fakeQuizModelAdapter{
		response: QuizModelResponse{
			ID:        "response-1",
			RequestID: request.ID,
			Provider:  "fake",
			Status:    QuizModelResponseReady,
			DraftSpec: QuizDraftSpec{
				Scope:          QuizScope{ProjectID: "project-1"},
				Title:          "Generated Spec",
				Body:           "Generated behavior.",
				SourceQuizPath: ".chc/quizzes/project-quiz.md",
				SourceRawPath:  ".chc/quizzes/project-quiz-raw.md",
				CreatedAt:      quizRuntimeTestTime(),
			},
			CreatedAt: quizRuntimeTestTime(),
		},
	}

	response, err := InvokeQuizModel(context.Background(), adapter, request)
	if err != nil {
		t.Fatalf("InvokeQuizModel returned error: %v", err)
	}

	if response.DraftSpec.Title != "Generated Spec" {
		t.Fatalf("draft spec title = %q", response.DraftSpec.Title)
	}
}

func TestInvokeQuizModelAcceptsPlanOutlineResponse(t *testing.T) {
	request := validQuizModelRequest()
	request.Action = QuizModelActionGeneratePlanOutline
	request.Strategy = "plan_outline_generation"
	request.OutputContract.RequiredFields = NewQuizModelOutputContract(
		QuizModelActionGeneratePlanOutline,
		request.Mode,
	).RequiredFields
	request.Payload.OutputContract = NewQuizModelOutputContract(QuizModelActionGeneratePlanOutline, request.Mode)

	adapter := &fakeQuizModelAdapter{
		response: QuizModelResponse{
			ID:          "response-1",
			RequestID:   request.ID,
			Provider:    "fake",
			Status:      QuizModelResponseReady,
			PlanOutline: validGeneratedPlanOutline(request.CreatedAt),
			CreatedAt:   quizRuntimeTestTime(),
		},
	}

	response, err := InvokeQuizModel(context.Background(), adapter, request)
	if err != nil {
		t.Fatalf("InvokeQuizModel returned error: %v", err)
	}

	if response.PlanOutline.Slices[0].ID != "Slice 1" {
		t.Fatalf("outline slice id = %q, want Slice 1", response.PlanOutline.Slices[0].ID)
	}
}

func TestInvokeQuizModelAcceptsPlanSliceDetailResponse(t *testing.T) {
	request := validQuizModelRequest()
	request.Action = QuizModelActionGeneratePlanSliceDetail
	request.Strategy = "plan_slice_detail_generation"
	request.OutputContract.RequiredFields = NewQuizModelOutputContract(
		QuizModelActionGeneratePlanSliceDetail,
		request.Mode,
	).RequiredFields
	request.Payload.OutputContract = NewQuizModelOutputContract(QuizModelActionGeneratePlanSliceDetail, request.Mode)

	adapter := &fakeQuizModelAdapter{
		response: QuizModelResponse{
			ID:          "response-1",
			RequestID:   request.ID,
			Provider:    "fake",
			Status:      QuizModelResponseReady,
			SliceDetail: validGeneratedPlanSliceDetail(request.CreatedAt),
			CreatedAt:   quizRuntimeTestTime(),
		},
	}

	response, err := InvokeQuizModel(context.Background(), adapter, request)
	if err != nil {
		t.Fatalf("InvokeQuizModel returned error: %v", err)
	}

	if response.SliceDetail.SliceID != "Slice 1" {
		t.Fatalf("slice detail id = %q, want Slice 1", response.SliceDetail.SliceID)
	}
}

func TestQuizModelContextSectionsForPayloadMatchesRedWrightShape(t *testing.T) {
	request := validQuizModelRequest()

	sections, err := QuizModelContextSectionsForPayload(request.Payload)
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

func TestQuizModelProviderPayloadKeepsNextQuestionContextOutOfProviderText(t *testing.T) {
	request := validQuizModelRequest()
	request.ContextSections = append(request.ContextSections, QuizModelContextSection{
		Name:      "project_dump",
		Layer:     "dump",
		Authority: ".chc/dumps/project-context.md",
		Content:   "Detailed project dump that should not be sent to next-question generation.",
	})

	payload := quizModelProviderPayload(request)

	if _, ok := payload["context_sections"]; ok {
		t.Fatalf("next-question provider payload includes context_sections: %#v", payload)
	}
}

func TestOpenAIQuizModelRequestBodySerializesFocusedNextQuestionPayload(t *testing.T) {
	request := validQuizModelRequest()
	request.ContextSections = append(request.ContextSections, QuizModelContextSection{
		Name:      "project_dump",
		Layer:     "dump",
		Authority: ".chc/dumps/project-context.md",
		Content:   "Detailed project dump that should not be sent to next-question generation.",
	})

	body, err := (OpenAIQuizModelAdapter{APIKey: "test-key", Model: "test-model"}).openAIRequestBody(request)
	if err != nil {
		t.Fatalf("openAIRequestBody returned error: %v", err)
	}

	envelope, input := decodeOpenAIQuizRequestBody(t, body)

	if envelope.Model != "test-model" {
		t.Fatalf("model = %q, want test-model", envelope.Model)
	}

	if envelope.Store {
		t.Fatal("store = true, want false")
	}

	if envelope.Metadata["chc_action"] != string(QuizModelActionNextQuestion) {
		t.Fatalf("metadata = %#v, want next-question action", envelope.Metadata)
	}

	if _, ok := input["context_sections"]; ok {
		t.Fatalf("next-question input includes context_sections: %#v", input)
	}

	for _, key := range []string{"project", "quiz", "recent_entries", "current_question", "output_contract"} {
		if _, ok := input[key]; !ok {
			t.Fatalf("next-question input missing key %q: %#v", key, input)
		}
	}

	if envelope.Text.Format.Name != "quiz_next_question" {
		t.Fatalf("format name = %q, want quiz_next_question", envelope.Text.Format.Name)
	}
}

func TestQuizModelProviderPayloadKeepsDraftSpecContextInProviderText(t *testing.T) {
	request := validQuizModelRequest()
	request.Action = QuizModelActionDraftSpec
	request.Strategy = "draft_spec"
	request.Payload.OutputContract = NewQuizModelOutputContract(QuizModelActionDraftSpec, request.Mode)
	request.OutputContract.RequiredFields = request.Payload.OutputContract.RequiredFields
	request.ContextSections = append(request.ContextSections, QuizModelContextSection{
		Name:      "project_dump",
		Layer:     "dump",
		Authority: ".chc/dumps/project-context.md",
		Content:   "Detailed project dump for draft spec synthesis.",
	})

	payload := quizModelProviderPayload(request)

	if _, ok := payload["context_sections"]; !ok {
		t.Fatalf("draft-spec provider payload omitted context_sections: %#v", payload)
	}
}

func TestOpenAIQuizModelRequestBodySerializesDraftSpecContextSections(t *testing.T) {
	request := validQuizModelRequest()
	request.Action = QuizModelActionDraftSpec
	request.Strategy = "draft_spec"
	request.Payload.OutputContract = NewQuizModelOutputContract(QuizModelActionDraftSpec, request.Mode)
	request.OutputContract.RequiredFields = request.Payload.OutputContract.RequiredFields
	request.ContextSections = append(request.ContextSections, QuizModelContextSection{
		Name:      "project_dump",
		Layer:     "dump",
		Authority: ".chc/dumps/project-context.md",
		Content:   "Detailed project dump for draft spec synthesis.",
		Metadata:  map[string]string{"source": "dump"},
	})

	body, err := (OpenAIQuizModelAdapter{APIKey: "test-key", Model: "test-model"}).openAIRequestBody(request)
	if err != nil {
		t.Fatalf("openAIRequestBody returned error: %v", err)
	}

	envelope, input := decodeOpenAIQuizRequestBody(t, body)

	if envelope.Metadata["chc_action"] != string(QuizModelActionDraftSpec) {
		t.Fatalf("metadata = %#v, want draft-spec action", envelope.Metadata)
	}

	sections, ok := input["context_sections"].([]any)
	if !ok {
		t.Fatalf("context_sections = %#v, want array", input["context_sections"])
	}

	if len(sections) != len(request.ContextSections) {
		t.Fatalf("context_sections length = %d, want %d", len(sections), len(request.ContextSections))
	}

	if envelope.Text.Format.Name != "quiz_draft_spec" {
		t.Fatalf("format name = %q, want quiz_draft_spec", envelope.Text.Format.Name)
	}
}

func TestQuizModelProviderPayloadKeepsPlanOutlineContextInProviderText(t *testing.T) {
	request := validQuizModelRequest()
	request.Action = QuizModelActionGeneratePlanOutline
	request.Strategy = "plan_outline_generation"
	request.Payload.OutputContract = NewQuizModelOutputContract(QuizModelActionGeneratePlanOutline, request.Mode)
	request.OutputContract.RequiredFields = request.Payload.OutputContract.RequiredFields
	request.ContextSections = append(request.ContextSections, QuizModelContextSection{
		Name:      "spec",
		Layer:     "spec",
		Authority: ".chc/specs/blog.md",
		Content:   "Spec context for outline generation.",
	})

	payload := quizModelProviderPayload(request)

	if _, ok := payload["context_sections"]; !ok {
		t.Fatalf("plan-outline provider payload omitted context_sections: %#v", payload)
	}
}

func TestOpenAIQuizModelRequestBodySerializesPlanOutlineContract(t *testing.T) {
	request := validQuizModelRequest()
	request.Action = QuizModelActionGeneratePlanOutline
	request.Strategy = "plan_outline_generation"
	request.Payload.OutputContract = NewQuizModelOutputContract(QuizModelActionGeneratePlanOutline, request.Mode)
	request.OutputContract.RequiredFields = request.Payload.OutputContract.RequiredFields

	body, err := (OpenAIQuizModelAdapter{APIKey: "test-key", Model: "test-model"}).openAIRequestBody(request)
	if err != nil {
		t.Fatalf("openAIRequestBody returned error: %v", err)
	}

	envelope, input := decodeOpenAIQuizRequestBody(t, body)

	if envelope.Metadata["chc_action"] != string(QuizModelActionGeneratePlanOutline) {
		t.Fatalf("metadata = %#v, want plan-outline action", envelope.Metadata)
	}

	if envelope.Text.Format.Name != "generated_plan_outline" {
		t.Fatalf("format name = %q, want generated_plan_outline", envelope.Text.Format.Name)
	}

	if envelope.Text.Format.Schema["required"] == nil {
		t.Fatalf("plan outline schema missing required fields: %#v", envelope.Text.Format.Schema)
	}

	if _, ok := input["context_sections"]; !ok {
		t.Fatalf("plan-outline input omitted context_sections: %#v", input)
	}
}

func TestOpenAIQuizModelRequestBodySerializesPlanSliceDetailContract(t *testing.T) {
	request := validQuizModelRequest()
	request.Action = QuizModelActionGeneratePlanSliceDetail
	request.Strategy = "plan_slice_detail_generation"
	request.Payload.OutputContract = NewQuizModelOutputContract(QuizModelActionGeneratePlanSliceDetail, request.Mode)
	request.OutputContract.RequiredFields = request.Payload.OutputContract.RequiredFields
	request.ContextSections = append(request.ContextSections, QuizModelContextSection{
		Name:      "plan_outline",
		Layer:     "plan",
		Authority: "generated_plan_outline",
		Content:   "Outline context.",
	})

	body, err := (OpenAIQuizModelAdapter{APIKey: "test-key", Model: "test-model"}).openAIRequestBody(request)
	if err != nil {
		t.Fatalf("openAIRequestBody returned error: %v", err)
	}

	envelope, input := decodeOpenAIQuizRequestBody(t, body)

	if envelope.Metadata["chc_action"] != string(QuizModelActionGeneratePlanSliceDetail) {
		t.Fatalf("metadata = %#v, want plan-slice-detail action", envelope.Metadata)
	}

	if envelope.Text.Format.Name != "generated_plan_slice_detail" {
		t.Fatalf("format name = %q, want generated_plan_slice_detail", envelope.Text.Format.Name)
	}

	required, ok := envelope.Text.Format.Schema["required"].([]any)
	if !ok {
		t.Fatalf("schema required = %#v, want []any", envelope.Text.Format.Schema["required"])
	}

	if !containsAnyString(required, "ticket_metadata") {
		t.Fatalf("slice detail schema required fields = %#v, want ticket_metadata", required)
	}

	if _, ok := input["context_sections"]; !ok {
		t.Fatalf("plan-slice-detail input omitted context_sections: %#v", input)
	}
}

func TestOpenAIQuizModelAdapterRecordsTraceEvents(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(ActiveOperatorIDEnv, "adrian")
	t.Setenv(LLMTraceEnabledEnv, "1")

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/responses" {
			t.Fatalf("path = %q, want /responses", request.URL.Path)
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"output_text":"{\"question\":\"What should happen next?\",\"tags\":[\"workflow\"],\"follow_up_depth\":1,\"completion_recommended\":false}"}`))
	}))
	defer server.Close()

	model := OpenAIQuizModelAdapter{
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Model:      "test-model",
		ProjectDir: dir,
		Client:     server.Client(),
	}

	request := validQuizModelRequest()
	response, err := model.InvokeQuizModel(context.Background(), request)
	if err != nil {
		t.Fatalf("InvokeQuizModel returned error: %v", err)
	}

	if response.Question.Prompt != "What should happen next?" {
		t.Fatalf("question = %q, want traced test response", response.Question.Prompt)
	}

	events := readLLMTraceEvents(t, dir)
	if len(events) != 2 {
		t.Fatalf("trace event count = %d, want 2: %#v", len(events), events)
	}

	requestEvent := events[0]
	responseEvent := events[1]
	if requestEvent["event"] != "llm_request" {
		t.Fatalf("request event = %#v", requestEvent["event"])
	}
	if responseEvent["event"] != "llm_response" {
		t.Fatalf("response event = %#v", responseEvent["event"])
	}
	if requestEvent["llm_call_id"] == "" {
		t.Fatalf("missing llm_call_id in request event: %#v", requestEvent)
	}
	if requestEvent["llm_call_id"] != responseEvent["llm_call_id"] {
		t.Fatalf("trace call ids differ: request=%#v response=%#v", requestEvent["llm_call_id"], responseEvent["llm_call_id"])
	}
	if responseEvent["status"] != "ready" {
		t.Fatalf("response status = %#v, want ready", responseEvent["status"])
	}
	if responseEvent["response_id"] == "" {
		t.Fatalf("missing response_id in response event: %#v", responseEvent)
	}
	if _, ok := requestEvent["request_payload"]; ok {
		t.Fatalf("request payload captured without payload opt-in: %#v", requestEvent)
	}
	if _, ok := responseEvent["response_payload"]; ok {
		t.Fatalf("response payload captured without payload opt-in: %#v", responseEvent)
	}
}

func TestOpenAIQuizModelAdapterRecordsTracePayloadsWhenEnabled(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(ActiveOperatorIDEnv, "adrian")
	t.Setenv(LLMTraceEnabledEnv, "1")
	t.Setenv(LLMTracePayloadsEnv, "1")

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"output_text":"{\"question\":\"What should happen next?\",\"tags\":[\"workflow\"],\"follow_up_depth\":1,\"completion_recommended\":false}"}`))
	}))
	defer server.Close()

	model := OpenAIQuizModelAdapter{
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Model:      "test-model",
		ProjectDir: dir,
		Client:     server.Client(),
	}

	_, err := model.InvokeQuizModel(context.Background(), validQuizModelRequest())
	if err != nil {
		t.Fatalf("InvokeQuizModel returned error: %v", err)
	}

	events := readLLMTraceEvents(t, dir)
	if _, ok := events[0]["request_payload"]; !ok {
		t.Fatalf("request payload missing with payload opt-in: %#v", events[0])
	}
	if _, ok := events[1]["response_payload"]; !ok {
		t.Fatalf("response payload missing with payload opt-in: %#v", events[1])
	}
	if _, ok := events[1]["output_payload"]; !ok {
		t.Fatalf("output payload missing with payload opt-in: %#v", events[1])
	}
}

func TestOpenAIQuizModelAdapterRecordsParseFailureTrace(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(ActiveOperatorIDEnv, "adrian")
	t.Setenv(LLMTraceEnabledEnv, "1")

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"output_text":"not-json"}`))
	}))
	defer server.Close()

	model := OpenAIQuizModelAdapter{
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Model:      "test-model",
		ProjectDir: dir,
		Client:     server.Client(),
	}

	_, err := model.InvokeQuizModel(context.Background(), validQuizModelRequest())
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
	t.Setenv(ActiveOperatorIDEnv, "adrian")
	t.Setenv(LLMTraceEnabledEnv, "1")

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		http.Error(writer, "provider failed", http.StatusBadGateway)
	}))
	defer server.Close()

	model := OpenAIQuizModelAdapter{
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Model:      "test-model",
		ProjectDir: dir,
		Client:     server.Client(),
	}

	_, err := model.InvokeQuizModel(context.Background(), validQuizModelRequest())
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

func TestOpenAIQuizModelTimeoutUsesLongerTimeoutForGeneration(t *testing.T) {
	if got := openAIQuizModelTimeout(QuizModelActionNextQuestion); got != 60*time.Second {
		t.Fatalf("next question timeout = %s, want 60s", got)
	}

	if got := openAIQuizModelTimeout(QuizModelActionGenerateSpec); got != 5*time.Minute {
		t.Fatalf("spec generation timeout = %s, want 5m", got)
	}

	if got := openAIQuizModelTimeout(QuizModelActionGeneratePlan); got != 5*time.Minute {
		t.Fatalf("plan generation timeout = %s, want 5m", got)
	}

	if got := openAIQuizModelTimeout(QuizModelActionGeneratePlanOutline); got != 5*time.Minute {
		t.Fatalf("plan outline generation timeout = %s, want 5m", got)
	}

	if got := openAIQuizModelTimeout(QuizModelActionGeneratePlanSliceDetail); got != 5*time.Minute {
		t.Fatalf("plan slice detail generation timeout = %s, want 5m", got)
	}
}

func TestQuizModelPlanOutlineResponseFromOutput(t *testing.T) {
	request := validQuizModelRequest()
	request.Action = QuizModelActionGeneratePlanOutline
	request.Strategy = "plan_outline_generation"
	request.Payload.OutputContract = NewQuizModelOutputContract(QuizModelActionGeneratePlanOutline, request.Mode)
	request.OutputContract.RequiredFields = request.Payload.OutputContract.RequiredFields

	response, err := quizModelResponseFromOutput(request, []byte(`{
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
	}`))
	if err != nil {
		t.Fatalf("quizModelResponseFromOutput returned error: %v", err)
	}

	if response.PlanOutline.Title != "Staged Plan Generation" {
		t.Fatalf("plan outline title = %q", response.PlanOutline.Title)
	}
}

func TestValidateQuizModelPlanOutlinePayloadRejectsNonSequentialSlices(t *testing.T) {
	_, err := DecodeQuizModelPlanOutlinePayload([]byte(`{
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
	request.Action = QuizModelActionGeneratePlanSliceDetail
	request.Strategy = "plan_slice_detail_generation"
	request.Payload.OutputContract = NewQuizModelOutputContract(QuizModelActionGeneratePlanSliceDetail, request.Mode)
	request.OutputContract.RequiredFields = request.Payload.OutputContract.RequiredFields

	response, err := quizModelResponseFromOutput(request, []byte(`{
		"slice_id":"Slice 1",
		"short_name":"outline generation contract",
		"objective":"Add outline contracts.",
		"implementation_notes":["Keep this scoped to contracts."],
		"tasks":[{
			"id":"T1.1",
			"summary":"Add the model action.",
			"commit_hint":"feat(chc): add plan outline action"
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
	}`))
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

	err := ValidateGeneratedPlanSliceDetailAgainstOutline(detail, outline, "Slice 1")
	if err == nil {
		t.Fatal("ValidateGeneratedPlanSliceDetailAgainstOutline returned nil error")
	}
}

func TestQuizModelSpecGenerationInstructionsAvoidTopLevelBodyHeading(t *testing.T) {
	instructions := quizModelSpecGenerationInstructions()

	if !strings.Contains(instructions, "body without a top-level title heading") {
		t.Fatalf("spec generation instructions do not prevent duplicate title headings: %q", instructions)
	}
}

func TestQuizModelSpecGenerationInstructionsDoNotOverfitPathPolicy(t *testing.T) {
	instructions := quizModelSpecGenerationInstructions()

	for _, forbidden := range []string{
		"Do not propose arbitrary domain directories",
		"Do not invent new ChrysoCoord persistence locations",
	} {
		if strings.Contains(instructions, forbidden) {
			t.Fatalf("spec generation instructions contain overfitted path policy %q in: %q", forbidden, instructions)
		}
	}
}

func TestQuizModelQuestionInstructionsMatchRedWrightContextBoundary(t *testing.T) {
	instructions := []string{
		quizModelDeepQuestionInstructions(),
		quizModelFastQuestionInstructions(),
	}

	for _, instruction := range instructions {
		if strings.Contains(instruction, "selected context sections") {
			t.Fatalf("question instructions mention selected context sections: %q", instruction)
		}

		if !strings.Contains(instruction, "service has already selected the quiz scope and routing state") {
			t.Fatalf("question instructions do not keep service-selected scope: %q", instruction)
		}

		if !strings.Contains(instruction, "functional analyst discovering requirements") {
			t.Fatalf("question instructions do not define analyst role: %q", instruction)
		}

		if !strings.Contains(instruction, "routing_state as binding conductor context") {
			t.Fatalf("question instructions do not bind routing state: %q", instruction)
		}

		if !strings.Contains(instruction, "one specific operational aspect") {
			t.Fatalf("question instructions do not define single-aspect focus: %q", instruction)
		}

		if !strings.Contains(instruction, "one area is becoming overrepresented") {
			t.Fatalf("question instructions do not define topic shift signal: %q", instruction)
		}
	}
}

func TestQuizModelRoutingStateCarriesTopicConductorSignals(t *testing.T) {
	request := validQuizModelRequest()
	request.Payload.RecentEntries = []QuizModelEntryPayload{
		quizModelEntryPayloadWithTags("identity"),
		quizModelEntryPayloadWithTags("identity"),
		quizModelEntryPayloadWithTags("storage"),
		quizModelEntryPayloadWithTags("identity"),
		quizModelEntryPayloadWithTags("identity"),
	}

	state := quizModelRoutingState(request, QuizModelQuestionPayload{
		Question:      "What identifier should appear in public paths?",
		Tags:          []string{"identity"},
		FollowUpDepth: 1,
	})

	coverage, ok := state["topic_coverage"].(map[string]int)
	if !ok {
		t.Fatalf("topic_coverage = %#v, want map[string]int", state["topic_coverage"])
	}

	if coverage["identity"] != 4 {
		t.Fatalf("identity coverage = %d, want 4", coverage["identity"])
	}

	streak, ok := state["recent_topic_streak"].(map[string]any)
	if !ok {
		t.Fatalf("recent_topic_streak = %#v, want map[string]any", state["recent_topic_streak"])
	}

	if streak["topic"] != "identity" {
		t.Fatalf("recent topic = %#v, want identity", streak["topic"])
	}

	if streak["count"] != 2 {
		t.Fatalf("recent topic count = %#v, want 2", streak["count"])
	}
}

func quizModelEntryPayloadWithTags(tags ...string) QuizModelEntryPayload {
	return QuizModelEntryPayload{
		Question: "Question?",
		Answer:   "Answer.",
		Status:   string(QuizEntryStatusAnswered),
		Tags:     tags,
	}
}

func TestDecodeQuizModelQuestionPayloadEnforcesRedWrightContract(t *testing.T) {
	payload, err := DecodeQuizModelQuestionPayload([]byte(`{
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

	if payload.FollowUpDepth != MaxQuizFollowUpDepth {
		t.Fatalf("follow up depth = %d, want clamp to %d", payload.FollowUpDepth, MaxQuizFollowUpDepth)
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
			_, err := DecodeQuizModelQuestionPayload([]byte(test.content))
			if err == nil {
				t.Fatal("DecodeQuizModelQuestionPayload returned nil error")
			}
		})
	}
}

func TestDecodeQuizModelPlanTrackingPayloadRequiresTicketSynopsis(t *testing.T) {
	payload, err := DecodeQuizModelPlanTrackingPayload([]byte(`{
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
	_, err := DecodeQuizModelPlanTrackingPayload([]byte(`{
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
	payload, err := DecodeQuizModelFastQuestionPayload([]byte(`{
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
	intent := NewQuizModelPromptIntent(QuizModelQuestionPayload{
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

	response, err := quizModelResponseFromOutput(request, []byte(`{
		"question": "What should editorial review clarify before publishing?",
		"tags": ["Editorial Review"],
		"follow_up_depth": 2,
		"completion_recommended": false
	}`))
	if err != nil {
		t.Fatalf("quizModelResponseFromOutput returned error: %v", err)
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
	request.Payload.RecentEntries = append(request.Payload.RecentEntries, QuizModelEntryPayload{
		Question: "Which publishing branch should we avoid?",
		Status:   string(QuizEntryStatusSkipped),
		Tags:     []string{"publishing"},
	})
	request.Payload.Quiz.EntryCount = len(request.Payload.RecentEntries)

	response, err := quizModelResponseFromOutput(request, []byte(`{
		"question": "What should the next useful publishing checkpoint clarify?",
		"tags": ["publishing"],
		"follow_up_depth": 0,
		"completion_recommended": false
	}`))
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

	response, err := quizModelResponseFromOutput(request, []byte(`{
		"question": "Is there anything else this quiz must capture before drafting?",
		"tags": ["completion"],
		"follow_up_depth": 1,
		"completion_recommended": true
	}`))
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

func TestQuizModelRuntimeDepthForcesTrunkAfterBranchReturn(t *testing.T) {
	state := validQuizRuntimeState()
	state.ActiveQuestion = nil
	state.Trail = nil
	state.BranchReturnReason = QuizBranchReturnReasonSkipped

	depth := quizModelRuntimeDepth(state, MaxQuizFollowUpDepth)
	if depth != 1 {
		t.Fatalf("depth = %d, want trunk depth after branch return", depth)
	}
}

func TestQuizModelResponseFromFastOutputBuffersCandidates(t *testing.T) {
	request := validQuizModelRequest()
	request.Mode = QuizModeFast
	request.Strategy = QuizModelFastStrategy
	request.Payload.Quiz.Mode = string(QuizModeFast)
	request.Payload.OutputContract = NewQuizModelOutputContract(QuizModelActionNextQuestion, QuizModeFast)

	response, err := quizModelResponseFromOutput(request, []byte(`{
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
	request.Action = QuizModelActionDraftSpec
	request.Strategy = "draft_spec"
	request.Payload.OutputContract = NewQuizModelOutputContract(QuizModelActionDraftSpec, request.Mode)
	request.OutputContract.RequiredFields = request.Payload.OutputContract.RequiredFields

	response, err := quizModelResponseFromOutput(request, []byte(`{
		"title": "Editorial Review Spec",
		"body": "The workflow must preserve editorial review.",
		"status": "draft"
	}`))
	if err != nil {
		t.Fatalf("quizModelResponseFromOutput returned error: %v", err)
	}

	if response.DraftSpec.Title != "Editorial Review Spec" {
		t.Fatalf("title = %q", response.DraftSpec.Title)
	}
}

func TestValidateQuizModelResponseAcceptsDraftSpecBeforeServiceProvenance(t *testing.T) {
	response := QuizModelResponse{
		ID:        "response-1",
		RequestID: "request-1",
		Provider:  "fake",
		Status:    QuizModelResponseReady,
		DraftSpec: QuizDraftSpec{
			Scope:     QuizScope{ProjectID: "project-1"},
			Title:     "Draft Spec",
			Body:      "Draft body.",
			CreatedAt: quizRuntimeTestTime(),
		},
		CreatedAt: quizRuntimeTestTime(),
	}

	err := ValidateQuizModelResponse(response)
	if err != nil {
		t.Fatalf("ValidateQuizModelResponse returned error: %v", err)
	}
}

func validQuizModelRequest() QuizModelRequest {
	state := validQuizRuntimeState()
	payload := QuizModelPayload{
		Project: QuizModelProjectPayload{
			ID:          "project-1",
			Name:        "Project",
			Description: "Project context.",
		},
		Quiz: QuizModelQuizPayload{
			ID:         state.ID,
			EntryCount: 1,
			Mode:       string(QuizModeDeep),
		},
		RecentEntries: []QuizModelEntryPayload{
			{
				Question: "What should this workflow preserve?",
				Answer:   "The workflow should preserve review.",
				Status:   string(QuizEntryStatusAnswered),
				Tags:     []string{"workflow"},
			},
		},
		CurrentQuestion: QuizModelCurrentQuestionPayload{
			Question: "What should this workflow preserve?",
			Tags:     []string{"workflow"},
			PromptIntent: map[string]any{
				"metadata": map[string]any{"follow_up_depth": 0},
			},
		},
		OutputContract: NewQuizModelOutputContract(QuizModelActionNextQuestion, QuizModeDeep),
	}

	return QuizModelRequest{
		ID:           "quiz-model-request-1",
		Action:       QuizModelActionNextQuestion,
		Strategy:     "deep",
		Scope:        QuizScope{ProjectID: "project-1"},
		Mode:         QuizModeDeep,
		RuntimeState: state,
		Payload:      payload,
		ContextSections: []QuizModelContextSection{
			{Name: "project", Layer: "project", Authority: "project_metadata", Content: "Project context."},
		},
		OutputContract: QuizModelOutputContract{
			RequiredFields: []string{"id", "prompt", "topic", "depth"},
		},
		CreatedAt: quizRuntimeTestTime(),
	}
}

func validGeneratedPlanOutline(createdAt time.Time) GeneratedPlanOutline {
	return GeneratedPlanOutline{
		Scope:              QuizScope{ProjectID: "project-1"},
		Title:              "Staged Plan Generation",
		Status:             "planned",
		Summary:            "Generate an outline first.",
		Assumptions:        []string{"Existing plans remain valid."},
		OpenQuestions:      []string{"Should staged generation be default?"},
		Risks:              []string{"More model calls."},
		ValidationStrategy: "Validate schema and request construction.",
		Slices: []GeneratedPlanOutlineSlice{
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

func validGeneratedPlanSliceDetail(createdAt time.Time) GeneratedPlanSliceDetail {
	return GeneratedPlanSliceDetail{
		Scope:               QuizScope{ProjectID: "project-1"},
		SliceID:             "Slice 1",
		ShortName:           "outline generation contract",
		Objective:           "Add outline contracts.",
		ImplementationNotes: []string{"Keep this scoped to contracts."},
		Tasks: []GeneratedPlanSliceTask{
			{
				ID:         "T1.1",
				Summary:    "Add the model action.",
				CommitHint: "feat(chc): add plan outline action",
			},
		},
		AcceptanceCriteria: []string{"The action validates."},
		TicketMetadata: GeneratedPlanSliceTicketMetadata{
			Title:    "Add outline contract",
			Synopsis: "Adds outline generation contracts.",
			Priority: "medium",
			Severity: "normal",
			Type:     "feature",
			Tags:     []string{"planning"},
		},
		BranchMetadata: GeneratedPlanSliceBranchMetadata{
			BranchName: "feat/slice-1-outline-generation-contract",
		},
		PRExpectations:         []string{"Open one PR against dev."},
		ValidationExpectations: []string{"Run focused tests."},
		NonGoals:               []string{"Do not connect runtime orchestration."},
		CreatedAt:              createdAt,
	}
}

func containsAnyString(values []any, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}

	return false
}

func decodeOpenAIQuizRequestBody(t *testing.T, body []byte) (openAIResponsesRequest, map[string]any) {
	t.Helper()

	var envelope openAIResponsesRequest

	err := json.Unmarshal(body, &envelope)
	if err != nil {
		t.Fatalf("decode OpenAI request body: %v", err)
	}

	input := map[string]any{}

	err = json.Unmarshal([]byte(envelope.Input), &input)
	if err != nil {
		t.Fatalf("decode OpenAI request input: %v", err)
	}

	return envelope, input
}

func readLLMTraceEvents(t *testing.T, projectDir string) []map[string]any {
	t.Helper()

	path := filepath.Join(projectDir, DirectoryName, "user", "log", "llm", defaultLLMTraceFileName)
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
