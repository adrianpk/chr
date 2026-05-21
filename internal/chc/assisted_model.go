package chc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type AssistedModelResponseStatus string

const (
	AssistedModelResponseReady   AssistedModelResponseStatus = "ready"
	AssistedModelResponseBlocked AssistedModelResponseStatus = "blocked"
)

type AssistedModelFailureKind string

const (
	AssistedModelFailureInvalidRequest  AssistedModelFailureKind = "invalid_request"
	AssistedModelFailureAdapter         AssistedModelFailureKind = "adapter"
	AssistedModelFailureInvalidResponse AssistedModelFailureKind = "invalid_response"
	AssistedModelFailureArtifact        AssistedModelFailureKind = "artifact"
)

type AssistedModelRequest struct {
	ID        string
	SessionID string
	TurnID    string
	Intent    AssistedAuthoringIntent
	Target    AuthoringTarget
	Context   string
	Prompt    string
	CreatedAt time.Time
}

type AssistedModelResponse struct {
	ID        string
	RequestID string
	SessionID string
	TurnID    string
	Provider  string
	Status    AssistedModelResponseStatus
	Question  AssistedAuthoringQuestion
	Result    AssistedAuthoringTurnResult
	RawText   string
	CreatedAt time.Time
}

type AssistedModelInvocationPaths struct {
	InvocationDir string
	RequestPath   string
	ResponsePath  string
}

type AssistedModelInvocationArtifacts struct {
	Paths    AssistedModelInvocationPaths
	Request  AssistedModelRequest
	Response AssistedModelResponse
}

type AssistedModelAdapter interface {
	InvokeAssistedModel(ctx context.Context, request AssistedModelRequest) (AssistedModelResponse, error)
}

type AssistedModelFailure struct {
	Kind AssistedModelFailureKind
	Err  error
}

type FakeAssistedModelAdapter struct {
	Responses []AssistedModelResponse
	Err       error
	index     int
}

func (s AssistedModelResponseStatus) Valid() bool {
	switch s {
	case AssistedModelResponseReady, AssistedModelResponseBlocked:
		return true
	default:
		return false
	}
}

func (k AssistedModelFailureKind) Valid() bool {
	switch k {
	case AssistedModelFailureInvalidRequest,
		AssistedModelFailureAdapter,
		AssistedModelFailureInvalidResponse,
		AssistedModelFailureArtifact:
		return true
	default:
		return false
	}
}

func (e AssistedModelFailure) Error() string {
	if e.Err == nil {
		return string(e.Kind)
	}

	return fmt.Sprintf("%s: %s", e.Kind, e.Err.Error())
}

func (e AssistedModelFailure) Unwrap() error {
	return e.Err
}

func NewAssistedModelFailure(kind AssistedModelFailureKind, err error) AssistedModelFailure {
	return AssistedModelFailure{Kind: kind, Err: err}
}

func ClassifyAssistedModelFailure(err error) AssistedModelFailureKind {
	var failure AssistedModelFailure

	if errors.As(err, &failure) {
		return failure.Kind
	}

	return AssistedModelFailureAdapter
}

func (a *FakeAssistedModelAdapter) InvokeAssistedModel(
	ctx context.Context,
	request AssistedModelRequest,
) (AssistedModelResponse, error) {
	err := ctx.Err()
	if err != nil {
		return AssistedModelResponse{}, err
	}

	if a.Err != nil {
		return AssistedModelResponse{}, a.Err
	}

	if a.index >= len(a.Responses) {
		return AssistedModelResponse{}, fmt.Errorf("fake assisted model response queue is empty")
	}

	response := a.Responses[a.index]
	a.index++

	if response.RequestID == "" {
		response.RequestID = request.ID
	}

	if response.SessionID == "" {
		response.SessionID = request.SessionID
	}

	if response.TurnID == "" {
		response.TurnID = request.TurnID
	}

	return response, nil
}

func InvokeAssistedModel(
	ctx context.Context,
	adapter AssistedModelAdapter,
	request AssistedModelRequest,
) (AssistedModelResponse, error) {
	err := ValidateAssistedModelRequest(request)
	if err != nil {
		return AssistedModelResponse{}, NewAssistedModelFailure(AssistedModelFailureInvalidRequest, err)
	}

	if adapter == nil {
		return AssistedModelResponse{}, NewAssistedModelFailure(
			AssistedModelFailureAdapter,
			fmt.Errorf("assisted model adapter is required"),
		)
	}

	response, err := adapter.InvokeAssistedModel(ctx, request)
	if err != nil {
		return AssistedModelResponse{}, NewAssistedModelFailure(AssistedModelFailureAdapter, err)
	}

	err = ValidateAssistedModelResponse(response)
	if err != nil {
		return AssistedModelResponse{}, NewAssistedModelFailure(AssistedModelFailureInvalidResponse, err)
	}

	return response, nil
}

func ValidateAssistedModelRequest(request AssistedModelRequest) error {
	if strings.TrimSpace(request.ID) == "" {
		return fmt.Errorf("assisted model request id is required")
	}

	if strings.TrimSpace(request.SessionID) == "" {
		return fmt.Errorf("assisted model request session id is required")
	}

	if strings.TrimSpace(request.TurnID) == "" {
		return fmt.Errorf("assisted model request turn id is required")
	}

	if !request.Intent.Valid() {
		return fmt.Errorf("assisted model request intent is invalid: %s", request.Intent)
	}

	if !request.Target.Valid() {
		return fmt.Errorf("assisted model request target is invalid: %s", request.Target)
	}

	if strings.TrimSpace(request.Context) == "" {
		return fmt.Errorf("assisted model request context is required")
	}

	if strings.TrimSpace(request.Prompt) == "" {
		return fmt.Errorf("assisted model request prompt is required")
	}

	if request.CreatedAt.IsZero() {
		return fmt.Errorf("assisted model request created time is required")
	}

	return nil
}

func ValidateAssistedModelResponse(response AssistedModelResponse) error {
	if strings.TrimSpace(response.ID) == "" {
		return fmt.Errorf("assisted model response id is required")
	}

	if strings.TrimSpace(response.RequestID) == "" {
		return fmt.Errorf("assisted model response request id is required")
	}

	if strings.TrimSpace(response.SessionID) == "" {
		return fmt.Errorf("assisted model response session id is required")
	}

	if strings.TrimSpace(response.TurnID) == "" {
		return fmt.Errorf("assisted model response turn id is required")
	}

	if strings.TrimSpace(response.Provider) == "" {
		return fmt.Errorf("assisted model response provider is required")
	}

	if !response.Status.Valid() {
		return fmt.Errorf("assisted model response status is invalid: %s", response.Status)
	}

	err := validateAssistedModelResponsePayload(response)
	if err != nil {
		return err
	}

	if response.CreatedAt.IsZero() {
		return fmt.Errorf("assisted model response created time is required")
	}

	return nil
}

func PlanAssistedModelInvocationArtifacts(
	sessionPaths AuthoringSessionPaths,
	request AssistedModelRequest,
) (AssistedModelInvocationPaths, error) {
	err := ValidateAssistedModelRequest(request)
	if err != nil {
		return AssistedModelInvocationPaths{}, err
	}

	if strings.TrimSpace(sessionPaths.SessionDir) == "" {
		return AssistedModelInvocationPaths{}, fmt.Errorf("authoring session directory is required")
	}

	invocationDir := filepath.Join(sessionPaths.SessionDir, "model", request.ID)

	return AssistedModelInvocationPaths{
		InvocationDir: invocationDir,
		RequestPath:   filepath.Join(invocationDir, "request.md"),
		ResponsePath:  filepath.Join(invocationDir, "response.md"),
	}, nil
}

func RecordAssistedModelInvocation(ctx context.Context, artifacts AssistedModelInvocationArtifacts) error {
	err := ctx.Err()
	if err != nil {
		return err
	}

	err = validateAssistedModelInvocationArtifacts(artifacts)
	if err != nil {
		return NewAssistedModelFailure(AssistedModelFailureArtifact, err)
	}

	err = os.MkdirAll(artifacts.Paths.InvocationDir, 0o755)
	if err != nil {
		return NewAssistedModelFailure(
			AssistedModelFailureArtifact,
			fmt.Errorf("create assisted model invocation directory: %w", err),
		)
	}

	files := map[string]string{
		artifacts.Paths.RequestPath:  RenderAssistedModelRequest(artifacts.Request),
		artifacts.Paths.ResponsePath: RenderAssistedModelResponse(artifacts.Response),
	}

	for path, content := range files {
		err = os.WriteFile(path, []byte(content), 0o644)
		if err != nil {
			return NewAssistedModelFailure(
				AssistedModelFailureArtifact,
				fmt.Errorf("write assisted model artifact: %w", err),
			)
		}
	}

	return nil
}

func RenderAssistedModelRequest(request AssistedModelRequest) string {
	var output bytes.Buffer

	output.WriteString("---\n")
	output.WriteString("id: " + yamlString(request.ID) + "\n")
	output.WriteString("session_id: " + yamlString(request.SessionID) + "\n")
	output.WriteString("turn_id: " + yamlString(request.TurnID) + "\n")
	output.WriteString("intent: " + yamlString(string(request.Intent)) + "\n")
	output.WriteString("target: " + yamlString(string(request.Target)) + "\n")
	output.WriteString("created_at: " + yamlTimestamp(request.CreatedAt) + "\n")
	output.WriteString("---\n\n")
	output.WriteString("# Assisted Model Request\n\n")
	output.WriteString("## Prompt\n\n")
	output.WriteString(strings.TrimSpace(request.Prompt))
	output.WriteString("\n\n")
	output.WriteString("## Context\n\n")
	output.WriteString(strings.TrimSpace(request.Context))
	output.WriteString("\n")

	return output.String()
}

func RenderAssistedModelResponse(response AssistedModelResponse) string {
	var output bytes.Buffer

	output.WriteString("---\n")
	output.WriteString("id: " + yamlString(response.ID) + "\n")
	output.WriteString("request_id: " + yamlString(response.RequestID) + "\n")
	output.WriteString("session_id: " + yamlString(response.SessionID) + "\n")
	output.WriteString("turn_id: " + yamlString(response.TurnID) + "\n")
	output.WriteString("provider: " + yamlString(response.Provider) + "\n")
	output.WriteString("status: " + yamlString(string(response.Status)) + "\n")
	output.WriteString("created_at: " + yamlTimestamp(response.CreatedAt) + "\n")
	output.WriteString("---\n\n")
	output.WriteString("# Assisted Model Response\n\n")
	writeAssistedModelResponseQuestion(&output, response)
	writeAssistedModelResponseResult(&output, response)
	output.WriteString("## Raw Text\n\n")
	output.WriteString(strings.TrimSpace(response.RawText))
	output.WriteString("\n")

	return output.String()
}

func validateAssistedModelResponsePayload(response AssistedModelResponse) error {
	if response.Status == AssistedModelResponseReady {
		return validateAssistedAuthoringQuestion(response.Question)
	}

	if response.Status == AssistedModelResponseBlocked {
		return validateAssistedAuthoringTurnResult(AssistedAuthoringTurnBlocked, response.Result)
	}

	return nil
}

func validateAssistedModelInvocationArtifacts(artifacts AssistedModelInvocationArtifacts) error {
	err := ValidateAssistedModelRequest(artifacts.Request)
	if err != nil {
		return err
	}

	err = ValidateAssistedModelResponse(artifacts.Response)
	if err != nil {
		return err
	}

	if artifacts.Response.RequestID != artifacts.Request.ID {
		return fmt.Errorf("assisted model response request id does not match request")
	}

	required := []struct {
		name  string
		value string
	}{
		{name: "invocation directory", value: artifacts.Paths.InvocationDir},
		{name: "request path", value: artifacts.Paths.RequestPath},
		{name: "response path", value: artifacts.Paths.ResponsePath},
	}

	for _, field := range required {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("assisted model artifact %s is required", field.name)
		}
	}

	return nil
}

func writeAssistedModelResponseQuestion(output *bytes.Buffer, response AssistedModelResponse) {
	output.WriteString("## Question\n\n")

	if response.Question.ID == "" {
		output.WriteString("No question was returned.\n\n")

		return
	}

	output.WriteString("- id: ")
	output.WriteString(response.Question.ID)
	output.WriteString("\n")
	output.WriteString("- prompt: ")
	output.WriteString(response.Question.Prompt)
	output.WriteString("\n")
	output.WriteString("- rationale: ")
	output.WriteString(response.Question.Rationale)
	output.WriteString("\n")
	output.WriteString("- required: ")
	output.WriteString(markdownBool(response.Question.Required))
	output.WriteString("\n\n")
}

func writeAssistedModelResponseResult(output *bytes.Buffer, response AssistedModelResponse) {
	output.WriteString("## Result\n\n")

	if response.Result.CompletionReason == "" && response.Result.BlockedReason == "" && response.Result.Summary == "" {
		output.WriteString("No turn result was returned.\n\n")

		return
	}

	output.WriteString("- completion_reason: ")
	output.WriteString(string(response.Result.CompletionReason))
	output.WriteString("\n")
	output.WriteString("- blocked_reason: ")
	output.WriteString(string(response.Result.BlockedReason))
	output.WriteString("\n")
	output.WriteString("- summary: ")
	output.WriteString(response.Result.Summary)
	output.WriteString("\n")
	output.WriteString("- next_intent: ")
	output.WriteString(string(response.Result.NextIntent))
	output.WriteString("\n\n")
}
