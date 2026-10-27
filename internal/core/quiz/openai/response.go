package openai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

type executionResult struct {
	Output       []byte
	ResponseBody []byte
	StatusCode   int
	Latency      time.Duration
}

type responsesEnvelope struct {
	OutputText string          `json:"output_text"`
	Output     []responsesItem `json:"output"`
	Error      *responsesError `json:"error"`
}

type responsesItem struct {
	Type    string                    `json:"type"`
	Role    string                    `json:"role"`
	Content []responsesContentSegment `json:"content"`
}

type responsesContentSegment struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type responsesError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

func (a Adapter) execute(request quizmodel.Request, httpRequest *http.Request, startedAt time.Time) (executionResult, error) {
	response, err := a.httpClient(request).Do(httpRequest)
	if err != nil {
		return executionResult{Latency: time.Since(startedAt)}, fmt.Errorf("invoke OpenAI quiz model: %w", err)
	}

	defer func() {
		//nolint:errcheck // Response body read errors are returned below.
		_ = response.Body.Close()
	}()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 4*1024*1024))

	result := executionResult{
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

	output, err := responseText(responseBody)
	if err != nil {
		return result, err
	}

	result.Output = []byte(output)

	return result, nil
}

func responseText(content []byte) (string, error) {
	var envelope responsesEnvelope

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
