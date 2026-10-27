package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

type responsesRequest struct {
	Model        string            `json:"model"`
	Instructions string            `json:"instructions"`
	Input        string            `json:"input"`
	Text         responsesText     `json:"text"`
	Store        bool              `json:"store"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type responsesText struct {
	Format responsesFormat `json:"format"`
}

func (a Adapter) invoke(ctx context.Context, request quizmodel.Request, trace callTrace) (invocationResult, error) {
	if a.apiKey() == "" {
		return invocationResult{}, fmt.Errorf("OPENAI_API_KEY is required for quiz model generation")
	}

	if len(trace.RequestBody) == 0 {
		body, err := a.requestBody(request)
		if err != nil {
			return invocationResult{}, err
		}

		trace.RequestBody = body
	}

	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		strings.TrimRight(a.baseURL(), "/")+"/responses",
		bytes.NewReader(trace.RequestBody),
	)
	if err != nil {
		return invocationResult{}, fmt.Errorf("build OpenAI quiz request: %w", err)
	}

	httpRequest.Header.Set("Authorization", "Bearer "+a.apiKey())
	httpRequest.Header.Set("Content-Type", "application/json")

	a.recordRequest(ctx, request, trace)

	result, err := a.execute(request, httpRequest, trace.StartedAt)

	return invocationResult{Output: result.Output, Execution: result}, err
}

func (a Adapter) requestBody(request quizmodel.Request) ([]byte, error) {
	content, err := json.Marshal(quizmodel.ProviderPayload(request))
	if err != nil {
		return nil, fmt.Errorf("marshal quiz model payload: %w", err)
	}

	body := responsesRequest{
		Model:        a.model(),
		Instructions: quizmodel.Instructions(request),
		Input:        string(content),
		Text: responsesText{
			Format: responseFormat(quizmodel.OutputSchema(request)),
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
