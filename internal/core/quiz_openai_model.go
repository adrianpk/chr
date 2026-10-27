package core

import quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"

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
	Format quizmodel.OutputFormatDefinition `json:"format"`
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
	request quizmodel.QuizModelRequest,
) (quizmodel.QuizModelResponse, error) {
	request.ID = HarnessGenerationRequestID(ctx, request.ID)

	err := quizmodel.ValidateQuizModelRequest(request)
	if err != nil {
		return quizmodel.QuizModelResponse{}, err
	}

	trace := a.newOpenAIQuizTrace(request)

	output, result, err := a.invokeBoundQuizAPI(ctx, request, trace)
	if err != nil {
		a.recordOpenAIQuizResponse(ctx, request, trace, result, "error", err, quizmodel.QuizModelResponse{})

		return quizmodel.QuizModelResponse{}, err
	}

	response, err := quizmodel.ResponseFromOutput(request, output)
	err = quizmodel.ValidateParsedResponse(response, err)

	err = ValidateHarnessGenerationOutput(ctx, a.quizAPIRunRequest(request, trace), err)
	if err != nil {
		a.recordOpenAIQuizResponse(ctx, request, trace, result, "parse_error", err, quizmodel.QuizModelResponse{})

		return quizmodel.QuizModelResponse{}, err
	}

	a.recordOpenAIQuizResponse(ctx, request, trace, result, "ready", nil, response)

	return response, nil
}

func (a OpenAIQuizModelAdapter) invokeOpenAI(
	ctx context.Context,
	request quizmodel.QuizModelRequest,
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

func (a OpenAIQuizModelAdapter) openAIRequestBody(request quizmodel.QuizModelRequest) ([]byte, error) {
	content, err := json.Marshal(quizmodel.ProviderPayload(request))
	if err != nil {
		return nil, fmt.Errorf("marshal quiz model payload: %w", err)
	}

	body := openAIResponsesRequest{
		Model:        a.openAIModel(),
		Instructions: quizmodel.Instructions(request),
		Input:        string(content),
		Text: openAIResponsesText{
			Format: quizmodel.OutputFormat(request),
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

func (a OpenAIQuizModelAdapter) executeOpenAIRequest(
	request quizmodel.QuizModelRequest,
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

func (a OpenAIQuizModelAdapter) newOpenAIQuizTrace(request quizmodel.QuizModelRequest) openAIQuizTrace {
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
	request quizmodel.QuizModelRequest,
	trace openAIQuizTrace,
) {
	RecordLLMTraceEvent(ctx, LoadLLMTraceConfigOptions{ProjectDir: a.ProjectDir}, map[string]any{
		"event":                 "llm_request",
		"llm_call_id":           trace.CallID,
		"request_id":            request.ID,
		"action":                string(request.Action),
		"provider":              quizmodel.QuizModelProviderOpenAI,
		"model":                 trace.Model,
		"request_payload_bytes": len(trace.RequestBody),
		"request_payload":       string(trace.RequestBody),
	})
}

func (a OpenAIQuizModelAdapter) recordOpenAIQuizResponse(
	ctx context.Context,
	request quizmodel.QuizModelRequest,
	trace openAIQuizTrace,
	result openAIQuizExecutionResult,
	status string,
	responseErr error,
	response quizmodel.QuizModelResponse,
) {
	event := map[string]any{
		"event":                  "llm_response",
		"llm_call_id":            trace.CallID,
		"request_id":             request.ID,
		"response_id":            response.ID,
		"action":                 string(request.Action),
		"provider":               quizmodel.QuizModelProviderOpenAI,
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

func (a OpenAIQuizModelAdapter) httpClient(request quizmodel.QuizModelRequest) *http.Client {
	if a.Client != nil {
		return a.Client
	}

	return &http.Client{Timeout: openAIQuizModelTimeout(request.Action)}
}

func openAIQuizModelTimeout(action quizmodel.QuizModelAction) time.Duration {
	switch action {
	case quizmodel.QuizModelActionGenerateSpec,
		quizmodel.QuizModelActionGenerateAcceptanceFeature,
		quizmodel.QuizModelActionGeneratePlan,
		quizmodel.QuizModelActionGeneratePlanOutline,
		quizmodel.QuizModelActionGeneratePlanSliceDetail,
		quizmodel.QuizModelActionMaterializePlanTracking:
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
