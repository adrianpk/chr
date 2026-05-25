package chc

import (
	"context"
	"encoding/json"
	"strings"
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

func TestQuizModelQuestionInstructionsMatchRedWrightContextBoundary(t *testing.T) {
	instructions := []string{
		quizModelDeepQuestionInstructions(),
		quizModelFastQuestionInstructions(),
	}

	for _, instruction := range instructions {
		if strings.Contains(instruction, "selected context sections") {
			t.Fatalf("question instructions mention selected context sections: %q", instruction)
		}

		if !strings.Contains(instruction, "Use the provided project, quiz, recent answer, and routing state context.") {
			t.Fatalf("question instructions do not match RedWright context wording: %q", instruction)
		}
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
