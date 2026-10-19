package core

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// QuizModelProviderOpenAI defines a package boundary constant.
	QuizModelProviderOpenAI = "openai"

	// QuizModelFastBatchSize defines a package boundary constant.
	QuizModelFastBatchSize = 10

	defaultOpenAIResponsesBaseURL  = "https://api.openai.com/v1"
	defaultOpenAIQuizModel         = "gpt-5-mini"
	defaultOpenAIQuizTimeout       = 60 * time.Second
	defaultOpenAIGenerationTimeout = 5 * time.Minute
)

// OpenAIQuizModelAdapter describes a package boundary value.
type OpenAIQuizModelAdapter struct {
	APIKey     string
	BaseURL    string
	Model      string
	ProjectDir string
	Client     *http.Client
}

type openAIQuizTrace struct {
	CallID      string
	Model       string
	RequestBody []byte
	StartedAt   time.Time
}

type openAIQuizExecutionResult struct {
	Output       []byte
	ResponseBody []byte
	StatusCode   int
	Latency      time.Duration
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

// InvokeQuizModel invokes the configured model adapter.
func (a OpenAIQuizModelAdapter) InvokeQuizModel(
	ctx context.Context,
	request QuizModelRequest,
) (QuizModelResponse, error) {
	request.ID = HarnessGenerationRequestID(ctx, request.ID)

	err := ValidateQuizModelRequest(request)
	if err != nil {
		return QuizModelResponse{}, err
	}

	trace := a.newOpenAIQuizTrace(request)

	output, result, err := a.invokeBoundQuizAPI(ctx, request, trace)
	if err != nil {
		a.recordOpenAIQuizResponse(ctx, request, trace, result, "error", err, QuizModelResponse{})

		return QuizModelResponse{}, err
	}

	response, err := quizModelResponseFromOutput(request, output)
	err = validateParsedQuizModelResponse(response, err)

	err = ValidateHarnessGenerationOutput(ctx, a.quizAPIRunRequest(request, trace), err)
	if err != nil {
		a.recordOpenAIQuizResponse(ctx, request, trace, result, "parse_error", err, QuizModelResponse{})

		return QuizModelResponse{}, err
	}

	a.recordOpenAIQuizResponse(ctx, request, trace, result, "ready", nil, response)

	return response, nil
}

func (a OpenAIQuizModelAdapter) invokeOpenAI(
	ctx context.Context,
	request QuizModelRequest,
	trace openAIQuizTrace,
) ([]byte, openAIQuizExecutionResult, error) {
	if a.openAIAPIKey() == "" {
		return nil, openAIQuizExecutionResult{}, fmt.Errorf("OPENAI_API_KEY is required for quiz model generation")
	}

	if len(trace.RequestBody) == 0 {
		body, err := a.openAIRequestBody(request)
		if err != nil {
			return nil, openAIQuizExecutionResult{}, err
		}

		trace.RequestBody = body
	}

	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		strings.TrimRight(a.openAIBaseURL(), "/")+"/responses",
		bytes.NewReader(trace.RequestBody),
	)
	if err != nil {
		return nil, openAIQuizExecutionResult{}, fmt.Errorf("build OpenAI quiz request: %w", err)
	}

	httpRequest.Header.Set("Authorization", "Bearer "+a.openAIAPIKey())
	httpRequest.Header.Set("Content-Type", "application/json")

	a.recordOpenAIQuizRequest(ctx, request, trace)

	result, err := a.executeOpenAIRequest(request, httpRequest, trace.StartedAt)

	return result.Output, result, err
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
			"chr_request_id": request.ID,
			"chr_action":     string(request.Action),
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
	if request.Payload.AcceptanceReadiness != nil {
		payload["acceptance_readiness"] = request.Payload.AcceptanceReadiness
	}

	hasContextAction := request.Action == QuizModelActionNextQuestion ||
		request.Action == QuizModelActionDraftSpec ||
		request.Action == QuizModelActionGenerateSpec ||
		request.Action == QuizModelActionGeneratePlan ||
		request.Action == QuizModelActionGeneratePlanOutline ||
		request.Action == QuizModelActionGeneratePlanSliceDetail ||
		request.Action == QuizModelActionMaterializePlanTracking
	if hasContextAction && len(request.ContextSections) > 0 {
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

func (a OpenAIQuizModelAdapter) executeOpenAIRequest(
	request QuizModelRequest,
	httpRequest *http.Request,
	startedAt time.Time,
) (openAIQuizExecutionResult, error) {
	response, err := a.httpClient(request).Do(httpRequest)
	if err != nil {
		return openAIQuizExecutionResult{Latency: time.Since(startedAt)}, fmt.Errorf("invoke OpenAI quiz model: %w", err)
	}

	defer func() {
		//nolint:errcheck // Response body read errors are returned below.
		_ = response.Body.Close()
	}()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 4*1024*1024))

	result := openAIQuizExecutionResult{
		ResponseBody: responseBody,
		StatusCode:   response.StatusCode,
		Latency:      time.Since(startedAt),
	}
	if err != nil {
		return result, fmt.Errorf("read OpenAI quiz response: %w", err)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return result, fmt.Errorf("OpenAI quiz model returned HTTP %d: %s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	output, err := openAIResponseText(responseBody)
	if err != nil {
		return result, err
	}

	result.Output = []byte(output)

	return result, nil
}

func (a OpenAIQuizModelAdapter) newOpenAIQuizTrace(request QuizModelRequest) openAIQuizTrace {
	body, err := a.openAIRequestBody(request)
	if err != nil {
		body = nil
	}

	return openAIQuizTrace{
		CallID:      newLLMCallID(),
		Model:       a.openAIModel(),
		RequestBody: body,
		StartedAt:   time.Now().UTC(),
	}
}

func (a OpenAIQuizModelAdapter) recordOpenAIQuizRequest(
	ctx context.Context,
	request QuizModelRequest,
	trace openAIQuizTrace,
) {
	RecordLLMTraceEvent(ctx, LoadLLMTraceConfigOptions{ProjectDir: a.ProjectDir}, map[string]any{
		"event":                 "llm_request",
		"llm_call_id":           trace.CallID,
		"request_id":            request.ID,
		"action":                string(request.Action),
		"provider":              QuizModelProviderOpenAI,
		"model":                 trace.Model,
		"request_payload_bytes": len(trace.RequestBody),
		"request_payload":       string(trace.RequestBody),
	})
}

func (a OpenAIQuizModelAdapter) recordOpenAIQuizResponse(
	ctx context.Context,
	request QuizModelRequest,
	trace openAIQuizTrace,
	result openAIQuizExecutionResult,
	status string,
	responseErr error,
	response QuizModelResponse,
) {
	event := map[string]any{
		"event":                  "llm_response",
		"llm_call_id":            trace.CallID,
		"request_id":             request.ID,
		"response_id":            response.ID,
		"action":                 string(request.Action),
		"provider":               QuizModelProviderOpenAI,
		"model":                  trace.Model,
		"status":                 status,
		"http_status":            result.StatusCode,
		"latency_ms":             result.Latency.Milliseconds(),
		"response_payload_bytes": len(result.ResponseBody),
		"output_payload_bytes":   len(result.Output),
		"response_payload":       string(result.ResponseBody),
		"output_payload":         string(result.Output),
	}
	if responseErr != nil {
		event["error"] = responseErr.Error()
	}

	RecordLLMTraceEvent(ctx, LoadLLMTraceConfigOptions{ProjectDir: a.ProjectDir}, event)
}

func newLLMCallID() string {
	var value [8]byte

	_, err := rand.Read(value[:])
	if err != nil {
		return "llm-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	}

	return "llm-" + hex.EncodeToString(value[:])
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
		model = strings.TrimSpace(os.Getenv("CHR_OPENAI_MODEL"))
	}

	if model == "" {
		return defaultOpenAIQuizModel
	}

	return model
}

// ResolvedModel exposes the non-secret legacy model identity for admission.
func (a OpenAIQuizModelAdapter) ResolvedModel() string {
	return a.openAIModel()
}

func (a OpenAIQuizModelAdapter) openAIBaseURL() string {
	baseURL := strings.TrimSpace(a.BaseURL)
	if baseURL == "" {
		baseURL = strings.TrimSpace(os.Getenv("CHR_OPENAI_BASE_URL"))
	}

	if baseURL == "" {
		return defaultOpenAIResponsesBaseURL
	}

	return baseURL
}

func (a OpenAIQuizModelAdapter) httpClient(request QuizModelRequest) *http.Client {
	if a.Client != nil {
		return a.Client
	}

	return &http.Client{Timeout: openAIQuizModelTimeout(request.Action)}
}

func openAIQuizModelTimeout(action QuizModelAction) time.Duration {
	switch action {
	case QuizModelActionGenerateSpec,
		QuizModelActionGenerateAcceptanceFeature,
		QuizModelActionGeneratePlan,
		QuizModelActionGeneratePlanOutline,
		QuizModelActionGeneratePlanSliceDetail,
		QuizModelActionMaterializePlanTracking:
		return defaultOpenAIGenerationTimeout
	default:
		return defaultOpenAIQuizTimeout
	}
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
	case QuizModelActionDraftSpec, QuizModelActionGenerateSpec:
		return quizModelDraftSpecResponseFromOutput(request, output)
	case QuizModelActionGenerateAcceptanceFeature:
		return quizModelFeatureResponseFromOutput(request, output)
	case QuizModelActionGeneratePlan:
		return quizModelPlanResponseFromOutput(request, output)
	case QuizModelActionGeneratePlanOutline:
		return quizModelPlanOutlineResponseFromOutput(request, output)
	case QuizModelActionGeneratePlanSliceDetail:
		return quizModelPlanSliceDetailResponseFromOutput(request, output)
	case QuizModelActionMaterializePlanTracking:
		return quizModelPlanTrackingResponseFromOutput(request, output)
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
			Feature:   quizModelFeatureGeneration(payload),
			CreatedAt: request.CreatedAt.UTC(),
		},
		RawText:   string(output),
		CreatedAt: request.CreatedAt.UTC(),
	}, nil
}

func quizModelFeatureGeneration(payload QuizModelDraftSpecPayload) AcceptanceFeatureGeneration {
	return AcceptanceFeatureGeneration{}
}

func quizModelFeatureResponseFromOutput(request QuizModelRequest, output []byte) (QuizModelResponse, error) {
	payload, err := DecodeQuizModelFeaturePayload(extractJSONObject(output))
	if err != nil {
		return QuizModelResponse{}, err
	}

	return QuizModelResponse{
		ID:        quizModelResponseID(request),
		RequestID: request.ID,
		Provider:  "openai",
		Status:    QuizModelResponseReady,
		Feature:   AcceptanceFeatureGeneration(payload),
		RawText:   strings.TrimSpace(string(output)),
		CreatedAt: request.CreatedAt,
	}, nil
}

func quizModelPlanResponseFromOutput(
	request QuizModelRequest,
	output []byte,
) (QuizModelResponse, error) {
	payload, err := DecodeQuizModelPlanPayload(extractJSONObject(output))
	if err != nil {
		return QuizModelResponse{}, err
	}

	return QuizModelResponse{
		ID:        quizModelResponseID(request),
		RequestID: request.ID,
		Provider:  QuizModelProviderOpenAI,
		Status:    QuizModelResponseReady,
		Plan: GeneratedPlan{
			Scope:     request.Scope,
			Title:     payload.Title,
			Body:      payload.Body,
			Order:     payload.Order,
			CreatedAt: request.CreatedAt.UTC(),
		},
		RawText:   string(output),
		CreatedAt: request.CreatedAt.UTC(),
	}, nil
}

func quizModelPlanOutlineResponseFromOutput(
	request QuizModelRequest,
	output []byte,
) (QuizModelResponse, error) {
	payload, err := DecodeQuizModelPlanOutlinePayload(extractJSONObject(output))
	if err != nil {
		return QuizModelResponse{}, err
	}

	outline := generatedPlanOutlineFromPayload(request.Scope, payload)
	outline.CreatedAt = request.CreatedAt.UTC()

	return QuizModelResponse{
		ID:          quizModelResponseID(request),
		RequestID:   request.ID,
		Provider:    QuizModelProviderOpenAI,
		Status:      QuizModelResponseReady,
		PlanOutline: outline,
		RawText:     string(output),
		CreatedAt:   request.CreatedAt.UTC(),
	}, nil
}

func quizModelPlanSliceDetailResponseFromOutput(
	request QuizModelRequest,
	output []byte,
) (QuizModelResponse, error) {
	payload, err := DecodeQuizModelPlanSliceDetailPayload(extractJSONObject(output))
	if err != nil {
		return QuizModelResponse{}, err
	}

	detail := generatedPlanSliceDetailFromPayload(request.Scope, payload)
	detail.CreatedAt = request.CreatedAt.UTC()

	return QuizModelResponse{
		ID:          quizModelResponseID(request),
		RequestID:   request.ID,
		Provider:    QuizModelProviderOpenAI,
		Status:      QuizModelResponseReady,
		SliceDetail: detail,
		RawText:     string(output),
		CreatedAt:   request.CreatedAt.UTC(),
	}, nil
}

func quizModelPlanTrackingResponseFromOutput(
	request QuizModelRequest,
	output []byte,
) (QuizModelResponse, error) {
	payload, err := DecodeQuizModelPlanTrackingPayload(extractJSONObject(output))
	if err != nil {
		return QuizModelResponse{}, err
	}

	tickets := make([]GeneratedPlanTrackingTicket, 0, len(payload.Tickets))
	for _, ticket := range payload.Tickets {
		tickets = append(tickets, GeneratedPlanTrackingTicket{
			SliceID:      ticket.SliceID,
			Title:        ticket.Title,
			Synopsis:     ticket.Synopsis,
			Summary:      ticket.Summary,
			Body:         ticket.Body,
			Priority:     ticket.Priority,
			Severity:     ticket.Severity,
			Type:         ticket.Type,
			Tags:         ticket.Tags,
			Dependencies: ticket.Dependencies,
		})
	}

	return QuizModelResponse{
		ID:        quizModelResponseID(request),
		RequestID: request.ID,
		Provider:  QuizModelProviderOpenAI,
		Status:    QuizModelResponseReady,
		PlanTracking: GeneratedPlanTracking{
			Scope:     request.Scope,
			Title:     payload.Title,
			Body:      payload.Body,
			Tickets:   tickets,
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

	state := map[string]any{
		"phase":                    quizModelPhase(payload),
		"entry_count":              entryCount,
		"remaining_questions":      quizModelMaxInt(DefaultQuizQuestionBudget-entryCount, 0),
		"follow_up_depth":          payload.FollowUpDepth,
		"max_follow_up_depth":      MaxQuizFollowUpDepth,
		"trunk_return_recommended": payload.FollowUpDepth >= MaxQuizFollowUpDepth,
		"covered_tags":             quizModelCoveredTags(request.Payload.RecentEntries),
		"topic_coverage":           quizModelTopicCoverage(request.Payload.RecentEntries),
		"recent_topic_streak":      quizModelRecentTopicStreak(request.Payload.RecentEntries),
		"mode":                     request.Payload.Quiz.Mode,
		"generation_strategy":      request.Strategy,
		"fallback_reason":          "",
		"candidate_count":          len(request.RuntimeState.Candidates),
	}
	if request.Payload.AcceptanceReadiness != nil {
		state["acceptance_readiness"] = request.Payload.AcceptanceReadiness

		promptIntent, ok := quizModelAcceptanceReadinessPromptIntent(*request.Payload.AcceptanceReadiness)
		if ok {
			state["acceptance_prompt_intent"] = promptIntent
		}
	}

	return state
}

func quizModelAcceptanceReadinessPromptIntent(
	readiness AcceptanceReadiness,
) (AcceptanceReadinessPromptIntent, bool) {
	if readiness.Decision != AcceptanceReadinessNeedsMoreInfo {
		return AcceptanceReadinessPromptIntent{}, false
	}

	for _, slot := range readiness.Missing {
		intent, err := NewAcceptanceReadinessPromptIntent(slot)
		if err == nil {
			return intent, true
		}
	}

	return AcceptanceReadinessPromptIntent{}, false
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

func quizModelTopicCoverage(entries []QuizModelEntryPayload) map[string]int {
	coverage := map[string]int{}

	for _, entry := range entries {
		for _, tag := range entry.Tags {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}

			coverage[tag]++
		}
	}

	return coverage
}

func quizModelRecentTopicStreak(entries []QuizModelEntryPayload) map[string]any {
	if len(entries) == 0 {
		return map[string]any{
			"topic": "",
			"count": 0,
		}
	}

	lastTopic := quizModelTopic(entries[len(entries)-1].Tags)
	count := 0

	for index := len(entries) - 1; index >= 0; index-- {
		if quizModelTopic(entries[index].Tags) != lastTopic {
			break
		}

		count++
	}

	return map[string]any{
		"topic": lastTopic,
		"count": count,
	}
}

func quizModelInstructions(request QuizModelRequest) string {
	switch {
	case request.Action == QuizModelActionMaterializePlanTracking:
		return quizModelPlanTrackingInstructions()
	case request.Action == QuizModelActionGeneratePlan:
		return quizModelPlanGenerationInstructions()
	case request.Action == QuizModelActionGeneratePlanOutline:
		return quizModelPlanOutlineGenerationInstructions()
	case request.Action == QuizModelActionGeneratePlanSliceDetail:
		return quizModelPlanSliceDetailGenerationInstructions()
	case request.Action == QuizModelActionGenerateSpec:
		return quizModelSpecGenerationInstructions()
	case request.Action == QuizModelActionGenerateAcceptanceFeature:
		return quizModelFeatureGenerationInstructions()
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
The service has already selected the quiz scope and routing state; do not change the scope.
You are acting as a functional analyst discovering requirements for a future specification.
Use routing_state as binding conductor context: keep the phase, follow-up depth, topic coverage, and branch-return guidance.
When acceptance_readiness is present, use it as guidance for whether the use case is naturally acceptance-testable.
If acceptance_readiness.decision is needs_more_info, use the first missing acceptance slot as internal guidance only; ask a natural question about the use case behavior, not about the schema.
If acceptance_readiness.decision is not_applicable, continue normal dump-oriented discovery and do not force Gherkin terminology.
Ask exactly one concise question about one specific operational aspect at a time.
Use follow-ups only when they materially improve clarity.
Shift topics when routing_state shows one area is becoming overrepresented.
Do not include examples or options unless they are truly necessary.
Do not ask the user to rank, pick a top N list, enumerate a full option set, or justify several choices at once.
Do not ask for secrets, credentials, tokens, passwords, private keys, or PII.
Return only the JSON object required by the response schema.
`)
}

func quizModelFastQuestionInstructions() string {
	return strings.TrimSpace(fmt.Sprintf(`
You generate the next %d questions for a project quiz fast mode.
The service has already selected the quiz scope and routing state; do not change the scope.
You are acting as a functional analyst discovering requirements for a future specification.
Use routing_state as binding conductor context: keep the phase, follow-up depth, topic coverage, and branch-return guidance.
When acceptance_readiness is present, use it as guidance for whether the use case is naturally acceptance-testable.
If acceptance_readiness.decision is needs_more_info, bias early questions toward concrete use-case behavior that may clarify missing acceptance slots, without naming Gherkin or the slot schema.
If acceptance_readiness.decision is not_applicable, continue normal dump-oriented discovery and do not force Gherkin terminology.
Prepare concise questions, each about one specific operational aspect.
Use follow-ups only when they materially improve clarity.
Shift topics when routing_state shows one area is becoming overrepresented.
Do not include examples or options unless they are truly necessary.
Prefer low-latency questions that can safely advance the quiz without exposing strategy internals.
Do not mention batch, round, tree, candidate, or strategy choices to the user.
If the next step needs semantic interpretation that is not explicit in the provided context, return a normal broad question.
Do not ask for secrets, credentials, tokens, passwords, private keys, or PII.
Return only the JSON object required by the response schema.
`, QuizModelFastBatchSize))
}

func quizModelPlanTrackingInstructions() string {
	return strings.TrimSpace(`
You materialize execution tracking artifacts from an existing implementation plan.
Use the structured request payload and context sections as the sole source of truth.
Do not execute work, do not create branches, do not open pull requests, and do not claim implementation progress.
Transform the plan into a tracking Markdown body and one ticket payload per task. A slice with three tasks needs three tickets and remains one integration unit.
The tracking body is an execution index: ordered slice list, dependencies, recommended ticket order, expected branch/PR references, and initial status.
Tickets are the detailed execution units and must be ready for later implementation work.
Each ticket must correspond to exactly one task in exactly one owning slice from the plan. Never collapse several tasks into one slice ticket.
Slice N and TN.M are display labels. Preserve supplied stable set, slice and task identities and their positive local orders; never derive identity from a title, display label or checkout path.
Ticket-native materialization requires explicit approved GeneratedWorkflowPlan input with aggregate slice contracts, individual outcome/scope/acceptance/validation contracts and a stable materialization request ID. Ordinary plan reads do not adopt legacy state.
Adoption must select original plan and ticket images, preserve content and history, and reject ambiguous identities or live ownership. A tracking body is a projection and cannot assign task ownership.
Each ticket synopsis must be one short sentence that lets the Board identify the ticket without opening the full body.
Each ticket body must include: task objective, scope, dependencies, acceptance criteria, validation expectations, and references to the source plan/tracking.
When a ticket fixes a bug, include a task to reproduce or observe the defect with a failing test before changing code when feasible.
Prefer table-driven tests for input/output behavior in validation expectations.
Mention property-based, fuzz, race, or benchmark tests only when the slice's use case justifies them.
Keep the content deterministic, concise, implementation-oriented, and review-friendly.
Do not duplicate implementation progress or invent completed work.
Do not ask for secrets, credentials, tokens, passwords, private keys, or PII.
Return only the JSON object required by the response schema.
`)
}

func quizModelPlanGenerationInstructions() string {
	return strings.TrimSpace(`
You generate an implementation plan for the selected project or use case.
Use the structured request payload and all context sections as the sole source of truth.
Synthesize project metadata, project-wide context, use-case metadata, use-case context, dumps, quizzes, existing specs, and related plans before planning.
Do not simply summarize specs. Derive executable implementation work from the available material.
When project_wide_planning_baseline is present, treat it as binding deterministic evidence about the repository state.
When project_wide_planning_readiness is present, obey its plan_kind, status, blocking_slots, and generation_guidance.
If plan_kind is project_gap_plan, generate gap-closure slices for the existing codebase; do not propose greenfield scaffolding, go.mod creation, Makefile creation, command creation, storage creation, export creation, or tests that the baseline shows already exist unless the slice explicitly changes or hardens the existing implementation.
If plan_kind is project_setup_plan, generate setup slices suitable for an empty or near-empty Go core.
If status is partial, still generate a plan, but mark it as provisional, include a Current Implementation Baseline section, and list assumptions and unresolved slots before the slice decomposition.
For project-wide plans, always include a Current Implementation Baseline section before the implementation strategy.
For project-wide plans, always include an Assumptions and Low-Confidence Items section when any slot is unresolved or inferred.

The plan is not a roadmap, milestone list, brainstorm, or architecture essay.
The plan is an executable implementation decomposition, sequencing mechanism, reviewable execution strategy, and bridge from specs to tickets.

Return order as structured slices and tasks matching the Markdown decomposition. Use stable machine IDs (slice-1, task-t1.1), positive slice_order and task_order, titles, and nested tasks. Do not allocate ticket identities or references. This map supplies intended order only; it does not approve execution.

Generate a deterministic Markdown body with these sections:
- High-level implementation strategy
- Slice decomposition
- Slice ordering
- Slice dependency analysis
- Tasks inside slices
- Ticket generation data
- Branch strategy
- Pull request strategy
- Merge sequencing
- Review requirements
- Completion validation criteria

Each slice must include:
- slice identifier
- title
- objective
- implementation rationale
- dependencies
- implementation notes
- acceptance criteria
- estimated execution complexity
- associated tasks
- ticket metadata
- branch metadata
- PR expectations

Tasks must be implementation-oriented and should generally map to one Conventional Commit.
Each task must map to one ticket. Its owning slice keeps one branch, one pull request, one review cycle, and one merge cycle for all its tasks.
Default to sequential slices unless isolated slices are truly parallelizable.
Use the integration branch from project configuration or provided context; only fall back to the default integration branch when no explicit policy is available.
Use deterministic branch names shaped like feature/<use-case>/<slice-id>-<short-name>.
Aim for 4 to 8 slices for normal work and up to 12 for large manageable work.
If more than 12 slices are needed, explicitly flag the use case as unusually large, explain why, and suggest smaller use cases or phases.
Optimize for incremental delivery, low review complexity, isolated change sets, predictable merge flow, agent execution compatibility, traceability, and resumability.
Avoid motivational language, filler, vague abstractions, generic agile terminology, fake precision, unnecessary infrastructure, and unrelated work.
Set status to planned.
Do not ask for secrets, credentials, tokens, passwords, private keys, or PII.
Return only the JSON object required by the response schema.
`)
}

func quizModelPlanOutlineGenerationInstructions() string {
	return strings.TrimSpace(`
You generate a structured implementation plan outline for the selected project or use case.
Use the structured request payload and all context sections as the sole source of truth.
Synthesize project metadata, project-wide context, use-case metadata, use-case context, dumps, quizzes, existing specs, related plans, and project-wide planning baseline/readiness sections before outlining.
Do not write the full implementation plan body.
Produce only the global slice map and planning constraints needed for later slice detail expansion.
The outline is authoritative for slice count, slice order, slice identifiers, branch names, PR titles, report paths, dependencies, assumptions, risks, and validation strategy.
Use sequential slice identifiers starting at Slice 1.
Keep each slice concise and implementation-oriented.
Do not generate tickets, code, or final Markdown plan prose.
Keep uncertain behavior explicit instead of inventing facts.
Set status to planned.
Do not ask for secrets, credentials, tokens, passwords, private keys, or PII.
Return only the JSON object required by the response schema.
`)
}

func quizModelPlanSliceDetailGenerationInstructions() string {
	return strings.TrimSpace(`
You expand exactly one implementation slice from an authoritative staged plan outline.
Use the structured request payload, the full plan_outline context section, and the target_slice context section as the sole source of truth.
Return details only for the target slice.
Do not rename the slice, change its slice identifier, reorder slices, add new slices, remove dependencies, or change branch metadata from the outline.
Expand the target slice into implementation notes, tasks, acceptance criteria, ticket metadata, PR expectations, validation expectations, dependencies, and non-goals.
Tasks must be implementation-oriented and should generally map to one Conventional Commit.
Keep uncertain behavior explicit instead of inventing facts.
Do not generate the final Markdown plan body.
Do not generate tickets, code, branches, commits, or pull requests.
Do not ask for secrets, credentials, tokens, passwords, private keys, or PII.
Return only the JSON object required by the response schema.
`)
}

func quizModelSpecGenerationInstructions() string {
	return strings.TrimSpace(`
You generate a formal specification for the selected project or use case.
Use the structured request payload and all context sections as the sole source of truth.
Ground the spec in the available project metadata, use-case metadata, dumps, quiz notes, existing docs, and related artifacts.
Prefer concrete behavior, constraints, decisions, open questions, and implementation implications.
Keep uncertain behavior explicit instead of inventing facts.
Do not convert unresolved open questions into canonical requirements; keep them explicit or label any chosen default as proposed.
Do not introduce new commands, flags, APIs, exit codes, or persistence fields as requirements unless they are grounded in context; label useful ungrounded additions as proposed optional extensions.
Keep implementation checklists secondary to behavioral requirements and avoid making them more authoritative than the spec decisions.
Fence every multi-line example in a Markdown code block with a language tag, such as yaml, json, shell, or text.
Never emit raw front matter, YAML, JSON, command output, or code snippets as plain paragraphs or headings.
Set status to draft because generated specs require user review.
Return body without a top-level title heading; the renderer adds the H1 from title.
Do not ask for secrets, credentials, tokens, passwords, private keys, or PII.
Return only the JSON object required by the response schema.
`)
}

func quizModelFeatureGenerationInstructions() string {
	return strings.TrimSpace(`
You generate a Gherkin acceptance feature artifact from the selected project or use case context.
Use the structured request payload and all context sections as the sole source of truth.
Return status generated only when the context supports concrete observable behavior and the feature body can pass static checks.
Return status needs_more_info when the context has acceptance potential but scenario details are still missing; explain the missing information in rejection_reason and include default-oriented clarification_prompts.
Return status not_applicable when the topic is not naturally suited to an acceptance feature; explain why in rejection_reason.
Return status rejected when a feature was attempted but would be unsafe, vague, or invalid; explain why in rejection_reason.
Clarification prompts must ask the user to confirm or adjust an inferred default instead of restating schema terms.
Generated feature bodies must include traceability tags and concrete Given, When, and Then steps.
Use at least one supported traceability tag on the feature or on every scenario: @spec:, @requirement:, @acceptance-criterion:, or @behavior:.
Other tags may be included as metadata, but they do not satisfy traceability by themselves.
Do not generate a feature that a trivial hardcoded happy-path implementation could satisfy.
When observable behavior applies, include more than one behavior path: at least one normal success path and at least one rejection, failure, boundary, terminal, or no-change path when the context supports it.
Keep combinatorial internal cases in unit, table-driven, fuzz, or property tests instead of bloating Gherkin.
Return status needs_more_info instead of generated when the available context only supports one weak happy-path scenario.
If an acceptance_feature_round context section includes answers, treat those answers as resolving their own prompts unless an accepted or active decision document directly contradicts them.
Always include status, body, rejection_reason, and clarification_prompts. Use an empty string for body or rejection_reason when not applicable, and an empty array for clarification_prompts when not needed.
Do not ask for secrets, credentials, tokens, passwords, private keys, or PII.
Return only the JSON object required by the response schema.
`)
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
	case request.Action == QuizModelActionMaterializePlanTracking:
		return quizModelJSONSchemaTextFormat(
			"plan_tracking_artifacts",
			"Plan tracking and ticket payloads derived from an existing implementation plan.",
			quizModelPlanTrackingJSONSchema(),
		)
	case request.Action == QuizModelActionGeneratePlan:
		return quizModelJSONSchemaTextFormat(
			"generated_plan",
			"Implementation plan generated from project and use-case context.",
			quizModelPlanJSONSchema(),
		)
	case request.Action == QuizModelActionGeneratePlanOutline:
		return quizModelJSONSchemaTextFormat(
			"generated_plan_outline",
			"Plan outline generated from project and use-case context.",
			quizModelPlanOutlineJSONSchema(),
		)
	case request.Action == QuizModelActionGeneratePlanSliceDetail:
		return quizModelJSONSchemaTextFormat(
			"generated_plan_slice_detail",
			"Implementation details generated for one outline slice.",
			quizModelPlanSliceDetailJSONSchema(),
		)
	case request.Action == QuizModelActionGenerateSpec:
		return quizModelJSONSchemaTextFormat(
			"generated_spec",
			"Draft specification generated from project and use-case context.",
			quizModelSpecJSONSchema(),
		)
	case request.Action == QuizModelActionGenerateAcceptanceFeature:
		return quizModelJSONSchemaTextFormat(
			"acceptance_feature",
			"Acceptance feature generated from project, use-case, quiz, dump, and spec context.",
			quizModelFeatureJSONSchema(),
		)
	case request.Action == QuizModelActionDraftSpec:
		return quizModelJSONSchemaTextFormat(
			"quiz_draft_spec",
			"Draft specification generated from project quiz context.",
			quizModelDraftSpecJSONSchema(),
		)
	case request.Mode == QuizModeFast:
		return quizModelJSONSchemaTextFormat(
			"quiz_fast_questions",
			"A batch of project quiz questions.",
			quizModelFastQuestionJSONSchema(),
		)
	default:
		return quizModelJSONSchemaTextFormat(
			"quiz_next_question",
			"The next project quiz question.",
			quizModelQuestionJSONSchema(),
		)
	}
}

func quizModelJSONSchemaTextFormat(name, description string, schema map[string]any) openAIResponsesTextFormat {
	return openAIResponsesTextFormat{
		Type:        "json_schema",
		Name:        name,
		Description: description,
		Strict:      true,
		Schema:      schema,
	}
}

//nolint:funlen // JSON schema literals are clearer kept together.
func quizModelPlanTrackingJSONSchema() map[string]any {
	ticketSchema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required": []string{
			"slice_id",
			"title",
			"synopsis",
			"summary",
			"body",
			"priority",
			"severity",
			"type",
			"tags",
			"dependencies",
		},
		"properties": map[string]any{
			"slice_id": map[string]any{"type": "string"},
			"title":    map[string]any{"type": "string"},
			"synopsis": map[string]any{
				"type":        "string",
				"description": "One short Board-facing sentence describing the ticket.",
			},
			"summary": map[string]any{"type": "string"},
			"body": map[string]any{
				"type":        "string",
				"description": "Markdown ticket body without front matter.",
			},
			"priority": map[string]any{
				"type": "string",
				"enum": []string{"low", "normal", "high", "critical"},
			},
			"severity": map[string]any{
				"type": "string",
				"enum": []string{"low", "normal", "high", "critical"},
			},
			"type": map[string]any{
				"type": "string",
				"enum": []string{"task", "bug", "feature", "cleanup", "research", "docs"},
			},
			"tags": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
			"dependencies": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
		},
	}

	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"title", "body", "status", "tickets"},
		"properties": map[string]any{
			"title": map[string]any{"type": "string"},
			"body": map[string]any{
				"type":        "string",
				"description": "Markdown tracking body without front matter.",
			},
			"status": map[string]any{
				"type": "string",
				"enum": []string{"tracking"},
			},
			"tickets": map[string]any{
				"type":     "array",
				"minItems": 1,
				"maxItems": 12,
				"items":    ticketSchema,
			},
		},
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
				"description": "One concise user-facing interviewer question.",
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

func quizModelSpecJSONSchema() map[string]any {
	return quizModelDraftSpecJSONSchema()
}

func quizModelFeatureJSONSchema() map[string]any {
	promptSchema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"id", "question", "default_answer", "reason"},
		"properties": map[string]any{
			"id": map[string]any{
				"type":        "string",
				"description": "Stable short prompt id.",
			},
			"question": map[string]any{
				"type":        "string",
				"description": "One concise user-facing question.",
			},
			"default_answer": map[string]any{
				"type":        "string",
				"description": "Best inferred answer the user can confirm or adjust.",
			},
			"reason": map[string]any{
				"type":        "string",
				"description": "Why this answer is needed for acceptance scenarios.",
			},
		},
	}

	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"status", "body", "rejection_reason", "clarification_prompts"},
		"properties": map[string]any{
			"status": map[string]any{
				"type": "string",
				"enum": []string{
					AcceptanceFeatureStatusGenerated,
					AcceptanceFeatureStatusNeedsMoreInfo,
					AcceptanceFeatureStatusRejected,
					AcceptanceFeatureStatusNotApplicable,
				},
			},
			"body": map[string]any{
				"type":        "string",
				"description": "Gherkin feature body when status is generated.",
			},
			"rejection_reason": map[string]any{
				"type":        "string",
				"description": "Reason generation needs more information, was rejected, or is not applicable.",
			},
			"clarification_prompts": map[string]any{
				"type":        "array",
				"description": "Default-oriented prompts required when status is needs_more_info.",
				"items":       promptSchema,
			},
		},
	}
}

func quizModelPlanJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"title", "body", "status", "order"},
		"properties": map[string]any{
			"order": planOrderJSONSchema(),
			"title": map[string]any{
				"type": "string",
			},
			"body": map[string]any{
				"type":        "string",
				"description": "Markdown implementation plan body without front matter.",
			},
			"status": map[string]any{
				"type": "string",
				"enum": []string{"planned"},
			},
		},
	}
}

func quizModelPlanOutlineJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required": []string{
			"title",
			"status",
			"summary",
			"assumptions",
			"open_questions",
			"risks",
			"validation_strategy",
			"slices",
			"dependencies",
		},
		"properties": map[string]any{
			"title":               map[string]any{"type": "string"},
			"status":              map[string]any{"type": "string", "enum": []string{"planned"}},
			"summary":             map[string]any{"type": "string"},
			"assumptions":         stringArrayJSONSchema(),
			"open_questions":      stringArrayJSONSchema(),
			"risks":               stringArrayJSONSchema(),
			"validation_strategy": map[string]any{"type": "string"},
			"slices":              map[string]any{"type": "array", "items": quizModelPlanOutlineSliceJSONSchema()},
			"dependencies":        map[string]any{"type": "array", "items": quizModelPlanOutlineDependencyJSONSchema()},
		},
	}
}

func quizModelPlanOutlineSliceJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required": []string{
			"id",
			"short_name",
			"objective",
			"rationale",
			"dependencies",
			"branch_name",
			"pr_title",
			"report_path",
			"complexity",
			"implementation_boundary",
		},
		"properties": map[string]any{
			"id":                      map[string]any{"type": "string"},
			"short_name":              map[string]any{"type": "string"},
			"objective":               map[string]any{"type": "string"},
			"rationale":               map[string]any{"type": "string"},
			"dependencies":            stringArrayJSONSchema(),
			"branch_name":             map[string]any{"type": "string"},
			"pr_title":                map[string]any{"type": "string"},
			"report_path":             map[string]any{"type": "string"},
			"complexity":              map[string]any{"type": "string"},
			"implementation_boundary": map[string]any{"type": "string"},
		},
	}
}

func quizModelPlanOutlineDependencyJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"from_slice_id", "to_slice_id", "reason"},
		"properties": map[string]any{
			"from_slice_id": map[string]any{"type": "string"},
			"to_slice_id":   map[string]any{"type": "string"},
			"reason":        map[string]any{"type": "string"},
		},
	}
}

func quizModelPlanSliceDetailJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required": []string{
			"slice_id",
			"short_name",
			"objective",
			"implementation_notes",
			"tasks",
			"acceptance_criteria",
			"ticket_metadata",
			"branch_metadata",
			"pr_expectations",
			"validation_expectations",
			"dependencies",
			"non_goals",
		},
		"properties": map[string]any{
			"slice_id":                map[string]any{"type": "string"},
			"short_name":              map[string]any{"type": "string"},
			"objective":               map[string]any{"type": "string"},
			"implementation_notes":    stringArrayJSONSchema(),
			"tasks":                   map[string]any{"type": "array", "items": quizModelPlanSliceTaskJSONSchema()},
			"acceptance_criteria":     stringArrayJSONSchema(),
			"ticket_metadata":         quizModelPlanSliceTicketMetadataJSONSchema(),
			"branch_metadata":         quizModelPlanSliceBranchMetadataJSONSchema(),
			"pr_expectations":         stringArrayJSONSchema(),
			"validation_expectations": stringArrayJSONSchema(),
			"dependencies":            stringArrayJSONSchema(),
			"non_goals":               stringArrayJSONSchema(),
		},
	}
}

func quizModelPlanSliceTaskJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"id", "summary", "commit_hint"},
		"properties": map[string]any{
			"id":          map[string]any{"type": "string"},
			"summary":     map[string]any{"type": "string"},
			"commit_hint": map[string]any{"type": "string"},
		},
	}
}

func quizModelPlanSliceTicketMetadataJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"title", "synopsis", "priority", "severity", "type", "tags"},
		"properties": map[string]any{
			"title":    map[string]any{"type": "string"},
			"synopsis": map[string]any{"type": "string"},
			"priority": map[string]any{"type": "string"},
			"severity": map[string]any{"type": "string"},
			"type":     map[string]any{"type": "string"},
			"tags":     stringArrayJSONSchema(),
		},
	}
}

func quizModelPlanSliceBranchMetadataJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"branch_name"},
		"properties": map[string]any{
			"branch_name": map[string]any{"type": "string"},
		},
	}
}

func stringArrayJSONSchema() map[string]any {
	return map[string]any{
		"type":  "array",
		"items": map[string]any{"type": "string"},
	}
}
