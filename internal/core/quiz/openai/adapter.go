package openai

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

const (
	defaultBaseURL           = "https://api.openai.com/v1"
	defaultModel             = "gpt-5-mini"
	defaultQuizTimeout       = 60 * time.Second
	defaultGenerationTimeout = 5 * time.Minute
)

// Adapter describes a package boundary value.
type Adapter struct {
	Host       Host
	APIKey     string
	BaseURL    string
	Model      string
	ProjectDir string
	Client     *http.Client
}

// Invoke invokes the configured model adapter.
func (a Adapter) Invoke(ctx context.Context, request quizmodel.Request) (quizmodel.Response, error) {
	if a.Host == nil {
		return quizmodel.Response{}, fmt.Errorf("quiz generation host is required")
	}

	request.ID = a.Host.RequestID(ctx, request.ID)

	err := quizmodel.ValidateRequest(request)
	if err != nil {
		return quizmodel.Response{}, err
	}

	trace := a.newTrace(request)

	call, err := a.invokeBound(ctx, request, trace)
	if err != nil {
		a.recordResponse(ctx, request, trace, responseObservation{Execution: call.Execution, Status: "error", Err: err})

		return quizmodel.Response{}, err
	}

	response, err := quizmodel.ResponseFromOutput(request, call.Output, Provider)
	err = quizmodel.ValidateParsedResponse(response, err)

	err = a.Host.ValidateOutput(ctx, a.runRequest(request, trace), err)
	if err != nil {
		a.recordResponse(ctx, request, trace, responseObservation{Execution: call.Execution, Status: "parse_error", Err: err})

		return quizmodel.Response{}, err
	}

	a.recordResponse(ctx, request, trace, responseObservation{Execution: call.Execution, Status: "ready", Response: response})

	return response, nil
}

func (a Adapter) apiKey() string {
	apiKey := strings.TrimSpace(a.APIKey)
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	}

	return apiKey
}

func (a Adapter) model() string {
	model := strings.TrimSpace(a.Model)
	if model == "" {
		model = strings.TrimSpace(os.Getenv("CHR_OPENAI_MODEL"))
	}

	if model == "" {
		return defaultModel
	}

	return model
}

// ResolvedModel exposes the non-secret legacy model identity for admission.
func (a Adapter) ResolvedModel() string {
	return a.model()
}

func (a Adapter) baseURL() string {
	baseURL := strings.TrimSpace(a.BaseURL)
	if baseURL == "" {
		baseURL = strings.TrimSpace(os.Getenv("CHR_OPENAI_BASE_URL"))
	}

	if baseURL == "" {
		return defaultBaseURL
	}

	return baseURL
}

func (a Adapter) httpClient(request quizmodel.Request) *http.Client {
	if a.Client != nil {
		return a.Client
	}

	return &http.Client{Timeout: timeout(request.Action)}
}

func timeout(action quizmodel.Action) time.Duration {
	switch action {
	case quizmodel.ActionGenerateSpec,
		quizmodel.ActionGenerateAcceptanceFeature,
		quizmodel.ActionGeneratePlan,
		quizmodel.ActionGeneratePlanOutline,
		quizmodel.ActionGeneratePlanSliceDetail,
		quizmodel.ActionMaterializePlanTracking:
		return defaultGenerationTimeout
	default:
		return defaultQuizTimeout
	}
}
