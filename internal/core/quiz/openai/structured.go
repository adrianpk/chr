package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

// InvokeStructured performs the original bounded transport call after caller-owned admission.
func (a Adapter) InvokeStructured(ctx context.Context, request StructuredRequest) ([]byte, error) {
	if a.apiKey() == "" {
		return nil, errors.New("configured model credentials are unavailable")
	}

	body, err := json.Marshal(responsesRequest{
		Model: a.model(), Instructions: request.Instructions,
		Input: string(request.Input), Store: false,
		Text: responsesText{Format: responseFormat(quizmodel.SchemaDefinition{Name: request.Purpose, Description: "Typed application response", Schema: request.Schema})},
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(a.baseURL(), "/")+"/responses", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+a.apiKey())
	req.Header.Set("Content-Type", "application/json")

	return a.readStructured(ctx, req)
}

func (a Adapter) readStructured(ctx context.Context, req *http.Request) ([]byte, error) {
	client := a.Client
	if client == nil {
		client = &http.Client{}
	}

	response, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.New("configured model request failed")
	}

	defer func() {
		//nolint:errcheck // The bounded body read is checked below.
		_ = response.Body.Close()
	}()

	body, err := io.ReadAll(io.LimitReader(response.Body, 1024*1024+1))
	if err != nil || len(body) > 1024*1024 {
		return nil, errors.New("configured model response exceeds its boundary")
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("configured model HTTP status %d", response.StatusCode)
	}

	text, err := responseText(body)
	if err != nil {
		return nil, errors.New("configured model response has no usable text")
	}

	return []byte(text), nil
}

// StructuredRequest is the transport input after caller-owned admission and bounds.
type StructuredRequest struct {
	Purpose, Instructions string
	Input                 json.RawMessage
	Schema                map[string]any
}
