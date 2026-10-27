package app

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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	harnesstrace "forge.adrianpk.com/chr/internal/core/harness/trace"
)

const (
	dumpCurationProviderOpenAI = "openai"

	defaultOpenAIDumpCurationBaseURL = "https://api.openai.com/v1"
	defaultOpenAIDumpCurationModel   = "gpt-5-mini"
	defaultOpenAIDumpCurationTimeout = 5 * time.Minute
)

type openAIDumpCurationProvider struct {
	ProjectDir string
	APIKey     string
	Model      string
	BaseURL    string
	Client     *http.Client
	Now        func() time.Time
}

type dumpCurationOpenAITrace struct {
	CallID      string
	Model       string
	RequestID   string
	RequestBody []byte
	StartedAt   time.Time
}

type dumpCurationOpenAIExecutionResult struct {
	Output       []byte
	ResponseBody []byte
	ResponseID   string
	StatusCode   int
	Latency      time.Duration
}

type dumpCurationOpenAIRequest struct {
	Model        string                 `json:"model"`
	Instructions string                 `json:"instructions"`
	Input        string                 `json:"input"`
	Text         dumpCurationOpenAIText `json:"text"`
	Store        bool                   `json:"store"`
	Metadata     map[string]string      `json:"metadata,omitempty"`
}

type dumpCurationOpenAIText struct {
	Format dumpCurationOpenAITextFormat `json:"format"`
}

type dumpCurationOpenAITextFormat struct {
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Strict      bool           `json:"strict"`
	Schema      map[string]any `json:"schema"`
}

type dumpCurationOpenAIInput struct {
	Project  dumpCurationOpenAIProject          `json:"project"`
	Dump     dumpCurationOpenAIDump             `json:"dump"`
	Context  []dumpCurationOpenAIContextSection `json:"context_sections"`
	Contract map[string]any                     `json:"output_contract"`
}

type dumpCurationOpenAIProject struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type dumpCurationOpenAIDump struct {
	Title      string `json:"title"`
	CaseSlug   string `json:"case_slug,omitempty"`
	RawContent string `json:"raw_content"`
}

type dumpCurationOpenAIContextSection struct {
	Name      string            `json:"name"`
	Layer     string            `json:"layer"`
	Authority string            `json:"authority"`
	Content   string            `json:"content"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type dumpCurationOpenAIEnvelope struct {
	ID         string                         `json:"id"`
	OutputText string                         `json:"output_text"`
	Output     []dumpCurationOpenAIOutputItem `json:"output"`
	Error      *dumpCurationOpenAIError       `json:"error"`
}

type dumpCurationOpenAIOutputItem struct {
	Type    string                             `json:"type"`
	Role    string                             `json:"role"`
	Content []dumpCurationOpenAIContentSegment `json:"content"`
}

type dumpCurationOpenAIContentSegment struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type dumpCurationOpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

type dumpCurationPayload struct {
	PolishedBody string   `json:"polished_body"`
	Highlights   []string `json:"highlights"`
	Warnings     []string `json:"warnings"`
}

func (p openAIDumpCurationProvider) CurateDump(
	ctx context.Context,
	request DumpCurationRequest,
) (DumpCurationResult, error) {
	trace, err := p.newTrace(request)
	if err != nil {
		return DumpCurationResult{}, err
	}

	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		strings.TrimRight(p.openAIBaseURL(), "/")+"/responses",
		bytes.NewReader(trace.RequestBody),
	)
	if err != nil {
		return DumpCurationResult{}, fmt.Errorf("build OpenAI dump curation request: %w", err)
	}

	apiKey := p.openAIAPIKey()
	if apiKey == "" {
		return DumpCurationResult{}, fmt.Errorf("OPENAI_API_KEY is required for dump curation")
	}

	httpRequest.Header.Set("Authorization", "Bearer "+apiKey)
	httpRequest.Header.Set("Content-Type", "application/json")

	p.recordRequest(ctx, trace)

	result, err := p.executeBoundDumpAPI(ctx, request, httpRequest, trace)
	if err != nil {
		p.recordResponse(ctx, trace, result, "error", err)

		return DumpCurationResult{}, err
	}

	curation, err := decodeDumpCurationPayload(result.Output)

	err = core.ValidateHarnessGenerationOutput(ctx, p.dumpAPIRunRequest(request, trace), err)
	if err != nil {
		p.recordResponse(ctx, trace, result, "parse_error", err)

		return DumpCurationResult{}, err
	}

	p.recordResponse(ctx, trace, result, "ready", nil)

	return DumpCurationResult{
		PolishedBody: curation.PolishedBody,
		Highlights:   curation.Highlights,
		Warnings:     curation.Warnings,
	}, nil
}

func (p openAIDumpCurationProvider) newTrace(
	request DumpCurationRequest,
) (dumpCurationOpenAITrace, error) {
	now := p.now()
	requestID := "dump-curation-" + dumpSlug(firstNonEmptyDumpCurationString(request.Title, "dump"))

	body, err := p.openAIRequestBody(request, requestID)
	if err != nil {
		return dumpCurationOpenAITrace{}, err
	}

	return dumpCurationOpenAITrace{
		CallID:      newDumpCurationCallID(),
		Model:       p.openAIModel(),
		RequestID:   requestID,
		RequestBody: body,
		StartedAt:   now,
	}, nil
}

func (p openAIDumpCurationProvider) openAIRequestBody(
	request DumpCurationRequest,
	requestID string,
) ([]byte, error) {
	input, err := json.Marshal(p.openAIInput(request))
	if err != nil {
		return nil, fmt.Errorf("marshal dump curation input: %w", err)
	}

	body := dumpCurationOpenAIRequest{
		Model:        p.openAIModel(),
		Instructions: dumpCurationInstructions(),
		Input:        string(input),
		Text: dumpCurationOpenAIText{
			Format: dumpCurationOpenAITextFormat{
				Type:        "json_schema",
				Name:        "dump_curation",
				Description: "Curated Markdown dump body and highlights derived from a raw capture.",
				Strict:      true,
				Schema:      dumpCurationJSONSchema(),
			},
		},
		Store: false,
		Metadata: map[string]string{
			"chr_request_id": requestID,
			"chr_action":     "dump_curation",
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal OpenAI dump curation request: %w", err)
	}

	return payload, nil
}

func (p openAIDumpCurationProvider) openAIInput(request DumpCurationRequest) dumpCurationOpenAIInput {
	return dumpCurationOpenAIInput{
		Project: p.projectPayload(request),
		Dump: dumpCurationOpenAIDump{
			Title:      strings.TrimSpace(request.Title),
			CaseSlug:   strings.TrimSpace(request.Scope.CaseSlug),
			RawContent: strings.TrimSpace(request.Content),
		},
		Context: p.contextSections(request),
		Contract: map[string]any{
			"required_fields":     []string{"polished_body", "highlights", "warnings"},
			"rendered_text_field": "polished_body",
		},
	}
}

func (p openAIDumpCurationProvider) projectPayload(request DumpCurationRequest) dumpCurationOpenAIProject {
	project := dumpCurationOpenAIProject{ID: request.ProjectID}

	metadata, err := authoringcore.NewProjectMetadataService(request.ProjectDir).Load()
	if err != nil {
		return project
	}

	project.Name = metadata.Name
	project.Description = metadata.Description

	return project
}

func (p openAIDumpCurationProvider) contextSections(
	request DumpCurationRequest,
) []dumpCurationOpenAIContextSection {
	sections := []dumpCurationOpenAIContextSection{}

	sections = appendDumpCurationFileContextSection(
		sections,
		request.ProjectDir,
		"project_document",
		"project",
		authoringcore.ProjectMetadataPath(""),
	)

	if strings.TrimSpace(request.Scope.CaseSlug) != "" {
		sections = appendDumpCurationFileContextSection(
			sections,
			request.ProjectDir,
			"use_case_document",
			"use_case",
			authoringcore.CaseMetadataPath("", request.Scope.CaseSlug),
		)
	}

	return sections
}

func appendDumpCurationFileContextSection(
	sections []dumpCurationOpenAIContextSection,
	projectDir string,
	name string,
	layer string,
	path string,
) []dumpCurationOpenAIContextSection {
	if !filepath.IsAbs(path) {
		path = filepath.Join(projectDir, path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return sections
	}

	return append(sections, dumpCurationOpenAIContextSection{
		Name:      name,
		Layer:     layer,
		Authority: filepath.ToSlash(path),
		Content:   string(content),
		Metadata:  map[string]string{"context_layer": layer},
	})
}

func (p openAIDumpCurationProvider) executeOpenAIRequest(
	request *http.Request,
	startedAt time.Time,
) (dumpCurationOpenAIExecutionResult, error) {
	response, err := p.httpClient().Do(request)
	if err != nil {
		return dumpCurationOpenAIExecutionResult{Latency: time.Since(startedAt)}, fmt.Errorf("invoke OpenAI dump curation: %w", err)
	}

	defer func() {
		closeErr := response.Body.Close()
		if closeErr != nil {
			return
		}
	}()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return dumpCurationOpenAIExecutionResult{
			StatusCode: response.StatusCode,
			Latency:    time.Since(startedAt),
		}, fmt.Errorf("read OpenAI dump curation response: %w", err)
	}

	result := dumpCurationOpenAIExecutionResult{
		ResponseBody: responseBody,
		StatusCode:   response.StatusCode,
		Latency:      time.Since(startedAt),
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return result, fmt.Errorf("OpenAI dump curation returned HTTP %d: %s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	curationResponse, err := dumpCurationOpenAIResponseText(responseBody)
	output, responseID := curationResponse.Text, curationResponse.ResponseID
	result.Output = []byte(output)
	result.ResponseID = responseID

	if err != nil {
		return result, err
	}

	return result, nil
}

func dumpCurationOpenAIResponseText(content []byte) (dumpCurationResponseText, error) {
	var envelope dumpCurationOpenAIEnvelope

	err := json.Unmarshal(content, &envelope)
	if err != nil {
		return dumpCurationResponseText{Text: "", ResponseID: ""}, fmt.Errorf("decode OpenAI dump curation response: %w", err)
	}

	if envelope.Error != nil {
		return dumpCurationResponseText{Text: "", ResponseID: envelope.ID}, fmt.Errorf("OpenAI dump curation error: %s", envelope.Error.Message)
	}

	if strings.TrimSpace(envelope.OutputText) != "" {
		return dumpCurationResponseText{Text: envelope.OutputText, ResponseID: envelope.ID}, nil
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
		return dumpCurationResponseText{Text: "", ResponseID: envelope.ID}, fmt.Errorf("OpenAI dump curation returned no output text")
	}

	return dumpCurationResponseText{Text: text, ResponseID: envelope.ID}, nil
}

func decodeDumpCurationPayload(content []byte) (DumpCurationResult, error) {
	var payload dumpCurationPayload

	err := json.Unmarshal(extractDumpCurationJSONObject(content), &payload)
	if err != nil {
		return DumpCurationResult{}, fmt.Errorf("decode dump curation payload: %w", err)
	}

	result := normalizeDumpCurationResult(DumpCurationResult(payload), "")
	if result.PolishedBody == "" {
		return DumpCurationResult{}, fmt.Errorf("dump curation polished body is required")
	}

	if len(result.Highlights) == 0 {
		return DumpCurationResult{}, fmt.Errorf("dump curation highlights are required")
	}

	return result, nil
}

func extractDumpCurationJSONObject(content []byte) []byte {
	text := strings.TrimSpace(string(content))
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")

	if start >= 0 && end >= start {
		return []byte(text[start : end+1])
	}

	return content
}

func (p openAIDumpCurationProvider) recordRequest(
	ctx context.Context,
	trace dumpCurationOpenAITrace,
) {
	harnesstrace.Record(ctx, harnesstrace.LoadOptions{ProjectDir: p.ProjectDir}, map[string]any{
		"event":                 "llm_request",
		"llm_call_id":           trace.CallID,
		"request_id":            trace.RequestID,
		"action":                "dump_curation",
		"provider":              dumpCurationProviderOpenAI,
		"model":                 trace.Model,
		"request_payload_bytes": len(trace.RequestBody),
		"request_payload":       string(trace.RequestBody),
	})
}

func (p openAIDumpCurationProvider) recordResponse(
	ctx context.Context,
	trace dumpCurationOpenAITrace,
	result dumpCurationOpenAIExecutionResult,
	status string,
	responseErr error,
) {
	event := map[string]any{
		"event":                  "llm_response",
		"llm_call_id":            trace.CallID,
		"request_id":             trace.RequestID,
		"response_id":            result.ResponseID,
		"action":                 "dump_curation",
		"provider":               dumpCurationProviderOpenAI,
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

	harnesstrace.Record(ctx, harnesstrace.LoadOptions{ProjectDir: p.ProjectDir}, event)
}

func dumpCurationInstructions() string {
	return strings.TrimSpace(`
You curate one chr dump capture.
Use only the structured request payload and context sections as the source of truth.
Preserve the user's intent, meaningful wording, decisions, tensions, uncertainty, and open questions.
Correct spelling, punctuation, grammar, paragraph flow, and obvious transcription noise.
Remove conversational filler only when it does not carry meaning.
Do not turn the dump into a spec, plan, ticket, task list, or architecture essay.
Do not invent requirements, implementation details, project facts, or decisions.
Keep uncertainty explicit.
Redact obvious secrets, credentials, tokens, passwords, and private keys.
Return a concise highlights array with recoverable source points from the raw capture.
Return warnings only for non-fatal ambiguity, missing context, redaction, or fallback-worthy concerns.
Return only the JSON object required by the response schema.
`)
}

func dumpCurationJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"polished_body", "highlights", "warnings"},
		"properties": map[string]any{
			"polished_body": map[string]any{
				"type":        "string",
				"description": "Markdown body for the curated dump without front matter.",
			},
			"highlights": map[string]any{
				"type":     "array",
				"minItems": 1,
				"items": map[string]any{
					"type": "string",
				},
			},
			"warnings": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
		},
	}
}

func (p openAIDumpCurationProvider) openAIAPIKey() string {
	apiKey := strings.TrimSpace(p.APIKey)
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	}

	return apiKey
}

func (p openAIDumpCurationProvider) openAIModel() string {
	model := strings.TrimSpace(p.Model)
	if model == "" {
		model = strings.TrimSpace(os.Getenv("CHR_OPENAI_MODEL"))
	}

	if model == "" {
		return defaultOpenAIDumpCurationModel
	}

	return model
}

func (p openAIDumpCurationProvider) openAIBaseURL() string {
	baseURL := strings.TrimSpace(p.BaseURL)
	if baseURL == "" {
		baseURL = strings.TrimSpace(os.Getenv("CHR_OPENAI_BASE_URL"))
	}

	if baseURL == "" {
		return defaultOpenAIDumpCurationBaseURL
	}

	return baseURL
}

func (p openAIDumpCurationProvider) httpClient() *http.Client {
	if p.Client != nil {
		return p.Client
	}

	return &http.Client{Timeout: defaultOpenAIDumpCurationTimeout}
}

func (p openAIDumpCurationProvider) now() time.Time {
	if p.Now != nil {
		return p.Now().UTC()
	}

	return time.Now().UTC()
}

func newDumpCurationCallID() string {
	var value [8]byte

	_, err := rand.Read(value[:])
	if err != nil {
		return "llm-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	}

	return "llm-" + hex.EncodeToString(value[:])
}

func firstNonEmptyDumpCurationString(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}

	return ""
}

type dumpCurationResponseText struct {
	Text       string
	ResponseID string
}
