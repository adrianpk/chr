package chc

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
	QuizModelProviderOpenAI = "openai"

	QuizModelFastBatchSize = 10

	defaultOpenAIResponsesBaseURL  = "https://api.openai.com/v1"
	defaultOpenAIQuizModel         = "gpt-5-mini"
	defaultOpenAIQuizTimeout       = 60 * time.Second
	defaultOpenAIGenerationTimeout = 5 * time.Minute
)

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

func (a OpenAIQuizModelAdapter) InvokeQuizModel(
	ctx context.Context,
	request QuizModelRequest,
) (QuizModelResponse, error) {
	err := ValidateQuizModelRequest(request)
	if err != nil {
		return QuizModelResponse{}, err
	}

	trace := a.newOpenAIQuizTrace(request)
	output, result, err := a.invokeOpenAI(ctx, request, trace)
	if err != nil {
		a.recordOpenAIQuizResponse(ctx, request, trace, result, "error", err, QuizModelResponse{})

		return QuizModelResponse{}, err
	}

	response, err := quizModelResponseFromOutput(request, output)
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

	hasContextAction := request.Action == QuizModelActionDraftSpec ||
		request.Action == QuizModelActionGenerateSpec ||
		request.Action == QuizModelActionGeneratePlan ||
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
	body, _ := a.openAIRequestBody(request)

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

func (a OpenAIQuizModelAdapter) httpClient(request QuizModelRequest) *http.Client {
	if a.Client != nil {
		return a.Client
	}

	return &http.Client{Timeout: openAIQuizModelTimeout(request.Action)}
}

func openAIQuizModelTimeout(action QuizModelAction) time.Duration {
	switch action {
	case QuizModelActionGenerateSpec, QuizModelActionGeneratePlan, QuizModelActionMaterializePlanTracking:
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
	case QuizModelActionGeneratePlan:
		return quizModelPlanResponseFromOutput(request, output)
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
			CreatedAt: request.CreatedAt.UTC(),
		},
		RawText:   string(output),
		CreatedAt: request.CreatedAt.UTC(),
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
			CreatedAt: request.CreatedAt.UTC(),
		},
		RawText:   string(output),
		CreatedAt: request.CreatedAt.UTC(),
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

	return map[string]any{
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
	case request.Action == QuizModelActionGenerateSpec:
		return quizModelSpecGenerationInstructions()
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
Transform the plan into a tracking Markdown body and one ticket payload per implementation slice.
The tracking body is an execution index: ordered slice list, dependencies, recommended ticket order, expected branch/PR references, and initial status.
Tickets are the detailed execution units and must be ready for later implementation work.
Each ticket must correspond to exactly one slice from the plan.
Each ticket synopsis must be one short sentence that lets the Board identify the ticket without opening the full body.
Each ticket body must include: objective, tasks, dependencies, acceptance criteria, validation expectations, and references to the source plan/tracking.
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
If plan_kind is project_setup_plan, generate setup slices suitable for an empty or near-empty Go workspace.
If status is partial, still generate a plan, but mark it as provisional, include a Current Implementation Baseline section, and list assumptions and unresolved slots before the slice decomposition.
For project-wide plans, always include a Current Implementation Baseline section before the implementation strategy.
For project-wide plans, always include an Assumptions and Low-Confidence Items section when any slot is unresolved or inferred.

The plan is not a roadmap, milestone list, brainstorm, or architecture essay.
The plan is an executable implementation decomposition, sequencing mechanism, reviewable execution strategy, and bridge from specs to tickets.

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
Each slice must map to one ticket, one branch, one pull request, one review cycle, and one merge cycle.
Default to sequential slices unless isolated slices are truly parallelizable.
Assume the integration branch is dev unless context says otherwise.
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

func quizModelSpecGenerationInstructions() string {
	return strings.TrimSpace(`
You generate a formal specification for the selected project or use case.
Use the structured request payload and all context sections as the sole source of truth.
Ground the spec in the available project metadata, use-case metadata, dumps, quiz notes, existing docs, and related artifacts.
Prefer concrete behavior, constraints, decisions, open questions, and implementation implications.
Keep uncertain behavior explicit instead of inventing facts.
Set status to draft because generated specs require user review.
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
		return openAIResponsesTextFormat{
			Type:        "json_schema",
			Name:        "plan_tracking_artifacts",
			Description: "Plan tracking and ticket payloads derived from an existing implementation plan.",
			Strict:      true,
			Schema:      quizModelPlanTrackingJSONSchema(),
		}
	case request.Action == QuizModelActionGeneratePlan:
		return openAIResponsesTextFormat{
			Type:        "json_schema",
			Name:        "generated_plan",
			Description: "Implementation plan generated from project and use-case context.",
			Strict:      true,
			Schema:      quizModelPlanJSONSchema(),
		}
	case request.Action == QuizModelActionGenerateSpec:
		return openAIResponsesTextFormat{
			Type:        "json_schema",
			Name:        "generated_spec",
			Description: "Draft specification generated from project and use-case context.",
			Strict:      true,
			Schema:      quizModelDraftSpecJSONSchema(),
		}
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

func quizModelPlanJSONSchema() map[string]any {
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
				"description": "Markdown implementation plan body without front matter.",
			},
			"status": map[string]any{
				"type": "string",
				"enum": []string{"planned"},
			},
		},
	}
}
