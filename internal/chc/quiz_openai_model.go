package chc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	QuizModelProviderOpenAI = "openai"

	QuizModelFastBatchSize = 10

	defaultOpenAIResponsesBaseURL = "https://api.openai.com/v1"
	defaultOpenAIQuizModel        = "gpt-5-mini"
)

type OpenAIQuizModelAdapter struct {
	APIKey  string
	BaseURL string
	Model   string
	Client  *http.Client
}

type openAIResponsesRequest struct {
	Model        string              `json:"model"`
	Instructions string              `json:"instructions"`
	Input        string              `json:"input"`
	Text         openAIResponsesText `json:"text"`
	Store        bool                `json:"store"`
	Metadata     map[string]string   `json:"metadata,omitempty"`
}

type openAIResponsesText struct {
	Format openAIResponsesTextFormat `json:"format"`
}

type openAIResponsesTextFormat struct {
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Strict      bool           `json:"strict"`
	Schema      map[string]any `json:"schema"`
}

type openAIResponsesEnvelope struct {
	OutputText string                `json:"output_text"`
	Output     []openAIResponsesItem `json:"output"`
	Error      *openAIResponsesError `json:"error"`
}

type openAIResponsesItem struct {
	Type    string                          `json:"type"`
	Role    string                          `json:"role"`
	Content []openAIResponsesContentSegment `json:"content"`
}

type openAIResponsesContentSegment struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type openAIResponsesError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

func (a OpenAIQuizModelAdapter) InvokeQuizModel(
	ctx context.Context,
	request QuizModelRequest,
) (QuizModelResponse, error) {
	err := ValidateQuizModelRequest(request)
	if err != nil {
		return QuizModelResponse{}, err
	}

	output, err := a.invokeOpenAI(ctx, request)
	if err != nil {
		return QuizModelResponse{}, err
	}

	return quizModelResponseFromOutput(request, output)
}

func (a OpenAIQuizModelAdapter) invokeOpenAI(
	ctx context.Context,
	request QuizModelRequest,
) ([]byte, error) {
	if a.openAIAPIKey() == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required for quiz model generation")
	}

	body, err := a.openAIRequestBody(request)
	if err != nil {
		return nil, err
	}

	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		strings.TrimRight(a.openAIBaseURL(), "/")+"/responses",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("build OpenAI quiz request: %w", err)
	}

	httpRequest.Header.Set("Authorization", "Bearer "+a.openAIAPIKey())
	httpRequest.Header.Set("Content-Type", "application/json")

	return a.executeOpenAIRequest(httpRequest)
}

func (a OpenAIQuizModelAdapter) openAIRequestBody(request QuizModelRequest) ([]byte, error) {
	content, err := json.Marshal(quizModelProviderPayload(request))
	if err != nil {
		return nil, fmt.Errorf("marshal quiz model payload: %w", err)
	}

	body := openAIResponsesRequest{
		Model:        a.openAIModel(),
		Instructions: quizModelInstructions(request),
		Input:        string(content),
		Text: openAIResponsesText{
			Format: quizModelOpenAIFormat(request),
		},
		Store: false,
		Metadata: map[string]string{
			"chc_request_id": request.ID,
			"chc_action":     string(request.Action),
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal OpenAI quiz request: %w", err)
	}

	return payload, nil
}

func quizModelProviderPayload(request QuizModelRequest) map[string]any {
	payload := map[string]any{
		"project":          request.Payload.Project,
		"quiz":             request.Payload.Quiz,
		"recent_entries":   request.Payload.RecentEntries,
		"current_question": request.Payload.CurrentQuestion,
		"output_contract":  request.Payload.OutputContract,
	}

	if len(request.ContextSections) > 0 {
		payload["context_sections"] = quizModelProviderContextSections(request.ContextSections)
	}

	return payload
}

func quizModelProviderContextSections(sections []QuizModelContextSection) []map[string]any {
	payloads := make([]map[string]any, 0, len(sections))

	for _, section := range sections {
		payloads = append(payloads, map[string]any{
			"name":      section.Name,
			"layer":     section.Layer,
			"authority": section.Authority,
			"content":   section.Content,
			"metadata":  section.Metadata,
		})
	}

	return payloads
}

func (a OpenAIQuizModelAdapter) executeOpenAIRequest(httpRequest *http.Request) ([]byte, error) {
	response, err := a.httpClient().Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("invoke OpenAI quiz model: %w", err)
	}

	defer func() {
		//nolint:errcheck // Response body read errors are returned below.
		_ = response.Body.Close()
	}()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 4*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("read OpenAI quiz response: %w", err)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("OpenAI quiz model returned HTTP %d: %s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	output, err := openAIResponseText(responseBody)
	if err != nil {
		return nil, err
	}

	return []byte(output), nil
}

func (a OpenAIQuizModelAdapter) openAIAPIKey() string {
	apiKey := strings.TrimSpace(a.APIKey)
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	}

	return apiKey
}

func (a OpenAIQuizModelAdapter) openAIModel() string {
	model := strings.TrimSpace(a.Model)
	if model == "" {
		model = strings.TrimSpace(os.Getenv("CHC_OPENAI_MODEL"))
	}

	if model == "" {
		return defaultOpenAIQuizModel
	}

	return model
}

func (a OpenAIQuizModelAdapter) openAIBaseURL() string {
	baseURL := strings.TrimSpace(a.BaseURL)
	if baseURL == "" {
		baseURL = strings.TrimSpace(os.Getenv("CHC_OPENAI_BASE_URL"))
	}

	if baseURL == "" {
		return defaultOpenAIResponsesBaseURL
	}

	return baseURL
}

func (a OpenAIQuizModelAdapter) httpClient() *http.Client {
	if a.Client != nil {
		return a.Client
	}

	return &http.Client{Timeout: 60 * time.Second}
}

func openAIResponseText(content []byte) (string, error) {
	var envelope openAIResponsesEnvelope

	err := json.Unmarshal(content, &envelope)
	if err != nil {
		return "", fmt.Errorf("decode OpenAI quiz response: %w", err)
	}

	if envelope.Error != nil {
		return "", fmt.Errorf("OpenAI quiz model error: %s", envelope.Error.Message)
	}

	if strings.TrimSpace(envelope.OutputText) != "" {
		return envelope.OutputText, nil
	}

	var output strings.Builder

	for _, item := range envelope.Output {
		if item.Type != "message" {
			continue
		}

		for _, segment := range item.Content {
			if segment.Type == "output_text" || segment.Type == "text" {
				output.WriteString(segment.Text)
			}
		}
	}

	text := strings.TrimSpace(output.String())
	if text == "" {
		return "", fmt.Errorf("OpenAI quiz model returned no output text")
	}

	return text, nil
}

func quizModelResponseFromOutput(request QuizModelRequest, output []byte) (QuizModelResponse, error) {
	switch request.Action {
	case QuizModelActionNextQuestion:
		return quizModelQuestionResponseFromOutput(request, output)
	case QuizModelActionDraftSpec:
		return quizModelDraftSpecResponseFromOutput(request, output)
	default:
		return QuizModelResponse{}, fmt.Errorf("quiz model action is unsupported: %s", request.Action)
	}
}

func quizModelQuestionResponseFromOutput(
	request QuizModelRequest,
	output []byte,
) (QuizModelResponse, error) {
	if request.Mode == QuizModeFast {
		return quizModelFastQuestionResponseFromOutput(request, output)
	}

	payload, err := DecodeQuizModelQuestionPayload(extractJSONObject(output))
	if err != nil {
		return QuizModelResponse{}, err
	}

	return QuizModelResponse{
		ID:        quizModelResponseID(request),
		RequestID: request.ID,
		Provider:  QuizModelProviderOpenAI,
		Status:    QuizModelResponseReady,
		Question:  quizModelQuestionFromPayload(request, payload, 1),
		RawText:   string(output),
		CreatedAt: request.CreatedAt.UTC(),
	}, nil
}

func quizModelFastQuestionResponseFromOutput(
	request QuizModelRequest,
	output []byte,
) (QuizModelResponse, error) {
	payload, err := DecodeQuizModelFastQuestionPayload(extractJSONObject(output))
	if err != nil {
		return QuizModelResponse{}, err
	}

	questions := payload.Questions
	if len(questions) > QuizModelFastBatchSize {
		questions = questions[:QuizModelFastBatchSize]
	}

	return QuizModelResponse{
		ID:         quizModelResponseID(request),
		RequestID:  request.ID,
		Provider:   QuizModelProviderOpenAI,
		Status:     QuizModelResponseReady,
		Question:   quizModelQuestionFromPayload(request, questions[0], 1),
		Candidates: quizModelCandidatesFromPayloads(request, questions[1:]),
		RawText:    string(output),
		CreatedAt:  request.CreatedAt.UTC(),
	}, nil
}

func quizModelDraftSpecResponseFromOutput(
	request QuizModelRequest,
	output []byte,
) (QuizModelResponse, error) {
	payload, err := DecodeQuizModelDraftSpecPayload(extractJSONObject(output))
	if err != nil {
		return QuizModelResponse{}, err
	}

	return QuizModelResponse{
		ID:        quizModelResponseID(request),
		RequestID: request.ID,
		Provider:  QuizModelProviderOpenAI,
		Status:    QuizModelResponseReady,
		DraftSpec: QuizDraftSpec{
			Scope:     request.Scope,
			Title:     payload.Title,
			Body:      payload.Body,
			CreatedAt: request.CreatedAt.UTC(),
		},
		RawText:   string(output),
		CreatedAt: request.CreatedAt.UTC(),
	}, nil
}

func quizModelQuestionFromPayload(
	request QuizModelRequest,
	payload QuizModelQuestionPayload,
	offset int,
) QuizQuestion {
	depth := quizModelRuntimeDepth(request.RuntimeState, payload.FollowUpDepth)
	id := quizModelQuestionID(request, offset)
	parentID, trunkID := quizModelRoute(request.RuntimeState, depth, id)

	return QuizQuestion{
		ID:        id,
		Prompt:    payload.Question,
		Required:  !payload.CompletionRecommended,
		Topic:     quizModelTopic(payload.Tags),
		Depth:     depth,
		ParentID:  parentID,
		TrunkID:   trunkID,
		RouteID:   trunkID,
		Rationale: quizModelPromptIntentRationale(payload, request),
		CreatedAt: request.CreatedAt.UTC(),
	}
}

func quizModelCandidatesFromPayloads(
	request QuizModelRequest,
	payloads []QuizModelQuestionPayload,
) []QuizCandidate {
	candidates := make([]QuizCandidate, 0, len(payloads))

	for index, payload := range payloads {
		offset := index + 2
		depth := quizModelRuntimeDepth(request.RuntimeState, payload.FollowUpDepth)
		id := quizModelQuestionID(request, offset)
		parentID, trunkID := quizModelRoute(request.RuntimeState, depth, id)

		candidates = append(candidates, QuizCandidate{
			ID:        id,
			Prompt:    payload.Question,
			Topic:     quizModelTopic(payload.Tags),
			Depth:     depth,
			ParentID:  parentID,
			RouteID:   trunkID,
			Hints:     append([]string(nil), payload.Tags...),
			CreatedAt: request.CreatedAt.UTC(),
		})
	}

	return candidates
}

func quizModelRuntimeDepth(state QuizRuntimeState, followUpDepth int) int {
	if state.ActiveQuestion == nil && len(state.Trail) == 0 {
		return 1
	}

	if followUpDepth < 1 {
		return 1
	}

	return followUpDepth
}

func quizModelRoute(state QuizRuntimeState, depth int, fallbackTrunkID string) (string, string) {
	if depth <= 1 {
		return "", fallbackTrunkID
	}

	if state.ActiveQuestion != nil {
		return state.ActiveQuestion.ID, quizModelTrunkID(state, fallbackTrunkID)
	}

	if len(state.Trail) > 0 {
		parent := state.Trail[len(state.Trail)-1]

		return parent.QuestionID, quizModelTrunkID(state, fallbackTrunkID)
	}

	return "", fallbackTrunkID
}

func quizModelTrunkID(state QuizRuntimeState, fallback string) string {
	if state.ActiveQuestion != nil && strings.TrimSpace(state.ActiveQuestion.TrunkID) != "" {
		return state.ActiveQuestion.TrunkID
	}

	if len(state.Trail) > 0 {
		return state.Trail[0].QuestionID
	}

	return fallback
}

func quizModelQuestionID(request QuizModelRequest, offset int) string {
	return fmt.Sprintf("question-%03d", request.RuntimeState.QuestionsAsked+offset)
}

func quizModelResponseID(request QuizModelRequest) string {
	return "response-" + request.ID
}

func quizModelTopic(tags []string) string {
	if len(tags) == 0 {
		return "project-context"
	}

	return tags[0]
}

func quizModelPromptIntentRationale(payload QuizModelQuestionPayload, request QuizModelRequest) string {
	intent := NewQuizModelPromptIntent(payload, quizModelRoutingState(request, payload))

	content, err := json.Marshal(intent)
	if err != nil {
		return ""
	}

	return string(content)
}

func quizModelRoutingState(
	request QuizModelRequest,
	payload QuizModelQuestionPayload,
) map[string]any {
	entryCount := request.Payload.Quiz.EntryCount

	return map[string]any{
		"phase":                    quizModelPhase(payload),
		"entry_count":              entryCount,
		"remaining_questions":      quizModelMaxInt(DefaultQuizQuestionBudget-entryCount, 0),
		"follow_up_depth":          payload.FollowUpDepth,
		"max_follow_up_depth":      MaxQuizFollowUpDepth,
		"trunk_return_recommended": payload.FollowUpDepth >= MaxQuizFollowUpDepth,
		"covered_tags":             quizModelCoveredTags(request.Payload.RecentEntries),
		"mode":                     request.Payload.Quiz.Mode,
		"generation_strategy":      request.Strategy,
		"fallback_reason":          "",
		"candidate_count":          len(request.RuntimeState.Candidates),
	}
}

func quizModelMaxInt(left, right int) int {
	if left > right {
		return left
	}

	return right
}

func quizModelPhase(payload QuizModelQuestionPayload) string {
	if payload.FollowUpDepth > 1 {
		return "follow_up"
	}

	return "trunk"
}

func quizModelCoveredTags(entries []QuizModelEntryPayload) []string {
	seen := map[string]bool{}
	tags := []string{}

	for _, entry := range entries {
		for _, tag := range entry.Tags {
			tag = strings.TrimSpace(tag)
			if tag == "" || seen[tag] {
				continue
			}

			seen[tag] = true
			tags = append(tags, tag)
		}
	}

	return tags
}

func quizModelInstructions(request QuizModelRequest) string {
	switch {
	case request.Action == QuizModelActionDraftSpec:
		return quizModelDraftSpecInstructions()
	case request.Mode == QuizModeFast:
		return quizModelFastQuestionInstructions()
	default:
		return quizModelDeepQuestionInstructions()
	}
}

func quizModelDeepQuestionInstructions() string {
	return strings.TrimSpace(`
You generate exactly one next question for a project quiz.
Use the provided project, quiz, recent answer, routing state, and selected context sections.
Ask one concise question that helps capture useful project or use-case knowledge.
Prefer useful progression over repeating prior questions.
If follow-up depth is high, return to broader coverage.
Do not ask for secrets, credentials, tokens, passwords, private keys, or PII.
Return only the JSON object required by the response schema.
`)
}

func quizModelFastQuestionInstructions() string {
	return strings.TrimSpace(fmt.Sprintf(`
You generate the next %d questions for a project quiz fast mode.
Use the provided project, quiz, recent answer, routing state, and selected context sections.
Prefer low-latency questions that can safely advance the quiz without exposing strategy internals.
Do not mention batch, round, tree, candidate, or strategy choices to the user.
If the next step needs semantic interpretation that is not explicit in the provided context, return a normal broad question.
Do not ask for secrets, credentials, tokens, passwords, private keys, or PII.
Return only the JSON object required by the response schema.
`, QuizModelFastBatchSize))
}

func quizModelDraftSpecInstructions() string {
	return strings.TrimSpace(`
You generate a draft specification from a completed project dump and quiz.
Use the structured request payload and all context sections as the sole source of truth.
The output must be a formal draft spec grounded in the provided dump and quiz material.
Do not invent facts not grounded in context.
Do not ask for secrets, credentials, tokens, passwords, private keys, or PII.
Return only the JSON object required by the response schema.
`)
}

func quizModelOpenAIFormat(request QuizModelRequest) openAIResponsesTextFormat {
	switch {
	case request.Action == QuizModelActionDraftSpec:
		return openAIResponsesTextFormat{
			Type:        "json_schema",
			Name:        "quiz_draft_spec",
			Description: "Draft specification generated from project quiz context.",
			Strict:      true,
			Schema:      quizModelDraftSpecJSONSchema(),
		}
	case request.Mode == QuizModeFast:
		return openAIResponsesTextFormat{
			Type:        "json_schema",
			Name:        "quiz_fast_questions",
			Description: "A batch of project quiz questions.",
			Strict:      true,
			Schema:      quizModelFastQuestionJSONSchema(),
		}
	default:
		return openAIResponsesTextFormat{
			Type:        "json_schema",
			Name:        "quiz_next_question",
			Description: "The next project quiz question.",
			Strict:      true,
			Schema:      quizModelQuestionJSONSchema(),
		}
	}
}

func quizModelQuestionJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             QuizModelQuestionRequiredFields,
		"properties": map[string]any{
			"question": map[string]any{
				"type":        "string",
				"description": "The user-facing question text.",
			},
			"tags": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
			"follow_up_depth": map[string]any{
				"type":    "integer",
				"minimum": 0,
				"maximum": MaxQuizFollowUpDepth,
			},
			"completion_recommended": map[string]any{
				"type": "boolean",
			},
		},
	}
}

func quizModelFastQuestionJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"questions"},
		"properties": map[string]any{
			"questions": map[string]any{
				"type":     "array",
				"minItems": 1,
				"maxItems": QuizModelFastBatchSize,
				"items":    quizModelQuestionJSONSchema(),
			},
		},
	}
}

func quizModelDraftSpecJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"title", "body", "status"},
		"properties": map[string]any{
			"title": map[string]any{
				"type": "string",
			},
			"body": map[string]any{
				"type":        "string",
				"description": "Markdown spec body without front matter.",
			},
			"status": map[string]any{
				"type": "string",
				"enum": []string{"draft"},
			},
		},
	}
}
