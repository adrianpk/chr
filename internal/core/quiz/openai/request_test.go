package openai

import (
	"encoding/json"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"

	"testing"
	"time"
)

func TestOpenAIQuizModelRequestBodySerializesContextualNextQuestionPayload(t *testing.T) {
	request := validQuizModelRequest()
	request.ContextSections = append(request.ContextSections, quizmodel.ContextSection{
		Name:      "project_dump",
		Layer:     "dump",
		Authority: ".chr/default/dumps/project-context.md",
		Content:   "Detailed project dump that should guide next-question generation.",
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

	if envelope.Metadata["chr_action"] != string(quizmodel.ActionNextQuestion) {
		t.Fatalf("metadata = %#v, want next-question action", envelope.Metadata)
	}

	if _, ok := input["context_sections"]; !ok {
		t.Fatalf("next-question input missing context_sections: %#v", input)
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

func TestOpenAIQuizModelRequestBodySerializesDraftSpecContextSections(t *testing.T) {
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
		Metadata:  map[string]string{"source": "dump"},
	})

	body, err := (OpenAIQuizModelAdapter{APIKey: "test-key", Model: "test-model"}).openAIRequestBody(request)
	if err != nil {
		t.Fatalf("openAIRequestBody returned error: %v", err)
	}

	envelope, input := decodeOpenAIQuizRequestBody(t, body)

	if envelope.Metadata["chr_action"] != string(quizmodel.ActionDraftSpec) {
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

func TestOpenAIQuizModelRequestBodySerializesGeneratedSpecSchemaWithoutFeature(t *testing.T) {
	request := validQuizModelRequest()
	request.Action = quizmodel.ActionGenerateSpec
	request.Strategy = "spec_generation"
	request.Payload.OutputContract = quizmodel.NewOutputContract(quizmodel.ActionGenerateSpec, request.Mode)
	request.OutputContract.RequiredFields = request.Payload.OutputContract.RequiredFields

	body, err := (OpenAIQuizModelAdapter{APIKey: "test-key", Model: "test-model"}).openAIRequestBody(request)
	if err != nil {
		t.Fatalf("openAIRequestBody returned error: %v", err)
	}

	envelope, _ := decodeOpenAIQuizRequestBody(t, body)

	if envelope.Text.Format.Name != "generated_spec" {
		t.Fatalf("format name = %q, want generated_spec", envelope.Text.Format.Name)
	}

	properties, ok := envelope.Text.Format.Schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema properties = %#v, want map", envelope.Text.Format.Schema["properties"])
	}

	if _, ok := properties["feature"]; ok {
		t.Fatalf("schema properties includes feature: %#v", properties)
	}
}

func TestOpenAIQuizModelRequestBodySerializesPlanOutlineContract(t *testing.T) {
	request := validQuizModelRequest()
	request.Action = quizmodel.ActionGeneratePlanOutline
	request.Strategy = "plan_outline_generation"
	request.Payload.OutputContract = quizmodel.NewOutputContract(quizmodel.ActionGeneratePlanOutline, request.Mode)
	request.OutputContract.RequiredFields = request.Payload.OutputContract.RequiredFields

	body, err := (OpenAIQuizModelAdapter{APIKey: "test-key", Model: "test-model"}).openAIRequestBody(request)
	if err != nil {
		t.Fatalf("openAIRequestBody returned error: %v", err)
	}

	envelope, input := decodeOpenAIQuizRequestBody(t, body)

	if envelope.Metadata["chr_action"] != string(quizmodel.ActionGeneratePlanOutline) {
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
	request.Action = quizmodel.ActionGeneratePlanSliceDetail
	request.Strategy = "plan_slice_detail_generation"
	request.Payload.OutputContract = quizmodel.NewOutputContract(quizmodel.ActionGeneratePlanSliceDetail, request.Mode)
	request.OutputContract.RequiredFields = request.Payload.OutputContract.RequiredFields
	request.ContextSections = append(request.ContextSections, quizmodel.ContextSection{
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

	if envelope.Metadata["chr_action"] != string(quizmodel.ActionGeneratePlanSliceDetail) {
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

func TestOpenAIQuizModelTimeoutUsesLongerTimeoutForGeneration(t *testing.T) {
	if got := openAIQuizModelTimeout(quizmodel.ActionNextQuestion); got != 60*time.Second {
		t.Fatalf("next question timeout = %s, want 60s", got)
	}

	if got := openAIQuizModelTimeout(quizmodel.ActionGenerateSpec); got != 5*time.Minute {
		t.Fatalf("spec generation timeout = %s, want 5m", got)
	}

	if got := openAIQuizModelTimeout(quizmodel.ActionGeneratePlan); got != 5*time.Minute {
		t.Fatalf("plan generation timeout = %s, want 5m", got)
	}

	if got := openAIQuizModelTimeout(quizmodel.ActionGeneratePlanOutline); got != 5*time.Minute {
		t.Fatalf("plan outline generation timeout = %s, want 5m", got)
	}

	if got := openAIQuizModelTimeout(quizmodel.ActionGeneratePlanSliceDetail); got != 5*time.Minute {
		t.Fatalf("plan slice detail generation timeout = %s, want 5m", got)
	}
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
func containsAnyString(values []any, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}

	return false
}
func validQuizRuntimeState() quizcore.RuntimeState {
	now := quizRuntimeTestTime()
	question := validQuizQuestion()

	return quizcore.RuntimeState{
		ID:              "quiz-project",
		ProjectID:       "project-1",
		Status:          quizcore.RuntimeStatusActive,
		Mode:            quizcore.ModeDeep,
		QuestionBudget:  quizcore.DefaultQuestionBudget,
		QuestionsAsked:  1,
		AnswersRecorded: 0,
		SkipsRecorded:   0,
		ActiveTopic:     "workflow",
		Coverage: []quizcore.CoverageTopic{
			{
				ID:            "workflow",
				Label:         "Workflow",
				Status:        quizcore.CoverageStatusActive,
				QuestionCount: 1,
			},
		},
		ActiveQuestion: &question,
		Trail: []quizcore.TrailNode{
			{
				QuestionID: question.ID,
				Prompt:     question.Prompt,
				Topic:      question.Topic,
				Depth:      question.Depth,
			},
		},
		Candidates: []quizcore.Candidate{
			validQuizCandidate(),
		},
		RawPath:     "ops/default/quiz/00000-project-quiz-raw/artifact.md",
		CuratedPath: "ops/default/quiz/00000-project-quiz/artifact.md",
		UpdatedAt:   now,
	}
}
func quizRuntimeTestTime() time.Time {
	return time.Date(2026, 5, 24, 15, 0, 0, 0, time.UTC)
}
func validQuizQuestion() quizcore.Question {
	now := quizRuntimeTestTime()

	return quizcore.Question{
		ID:        "question-1",
		Prompt:    "What should this workflow preserve?",
		Required:  true,
		Topic:     "workflow",
		Depth:     1,
		TrunkID:   "question-1",
		Rationale: "The runtime needs the primary workflow invariant.",
		CreatedAt: now,
	}
}
func validQuizCandidate() quizcore.Candidate {
	now := quizRuntimeTestTime()

	return quizcore.Candidate{
		ID:        "candidate-1",
		Prompt:    "Which follow-up would clarify the workflow?",
		Topic:     "workflow",
		Depth:     1,
		RouteID:   "workflow-follow-up",
		Hints:     []string{"workflow"},
		CreatedAt: now,
	}
}
