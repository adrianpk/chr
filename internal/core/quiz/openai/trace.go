package openai

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"time"

	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

type callTrace struct {
	CallID      string
	Model       string
	RequestBody []byte
	StartedAt   time.Time
}

func (a Adapter) newTrace(request quizmodel.Request) callTrace {
	body, err := a.requestBody(request)
	if err != nil {
		body = nil
	}

	return callTrace{
		CallID:      newCallID(),
		Model:       a.model(),
		RequestBody: body,
		StartedAt:   time.Now().UTC(),
	}
}

func (a Adapter) recordRequest(ctx context.Context, request quizmodel.Request, trace callTrace) {
	a.Host.Trace(ctx, a.ProjectDir, map[string]any{
		"event":                 "llm_request",
		"llm_call_id":           trace.CallID,
		"request_id":            request.ID,
		"action":                string(request.Action),
		"provider":              Provider,
		"model":                 trace.Model,
		"request_payload_bytes": len(trace.RequestBody),
		"request_payload":       string(trace.RequestBody),
	})
}

func (a Adapter) recordResponse(ctx context.Context, request quizmodel.Request, trace callTrace, observed responseObservation) {
	event := map[string]any{
		"event":                  "llm_response",
		"llm_call_id":            trace.CallID,
		"request_id":             request.ID,
		"response_id":            observed.Response.ID,
		"action":                 string(request.Action),
		"provider":               Provider,
		"model":                  trace.Model,
		"status":                 observed.Status,
		"http_status":            observed.Execution.StatusCode,
		"latency_ms":             observed.Execution.Latency.Milliseconds(),
		"response_payload_bytes": len(observed.Execution.ResponseBody),
		"output_payload_bytes":   len(observed.Execution.Output),
		"response_payload":       string(observed.Execution.ResponseBody),
		"output_payload":         string(observed.Execution.Output),
	}
	if observed.Err != nil {
		event["error"] = observed.Err.Error()
	}

	a.Host.Trace(ctx, a.ProjectDir, event)
}

func newCallID() string {
	var value [8]byte

	_, err := rand.Read(value[:])
	if err != nil {
		return "llm-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	}

	return "llm-" + hex.EncodeToString(value[:])
}

type responseObservation struct {
	Execution executionResult
	Status    string
	Err       error
	Response  quizmodel.Response
}
