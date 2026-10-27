package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	"forge.adrianpk.com/chr/internal/harness"
)

func TestServiceCreateDumpUsesSuccessfulCurationProvider(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceWorkspaceQueryFixture(t, projectDir)

	operations := newTestService(t, projectDir)
	operations.dumpCurator = fakeDumpCurationProvider{
		result: DumpCurationResult{
			PolishedBody: "Polished dump body.",
			Highlights:   []string{"Important source point."},
		},
	}

	_, err := operations.CreateDump(DumpCreateRequest{
		OperationID: dumpRequestID(""),
		Title:       "Curated Notes",
		Content:     "rough dump body",
	})
	if err != nil {
		t.Fatalf("CreateDump returned error: %v", err)
	}

	curated := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "dumps", "curated-notes.md"))
	if !strings.Contains(curated, "Polished dump body.") {
		t.Fatalf("curated dump = %q, want polished body", curated)
	}

	raw := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "dumps", "curated-notes-raw.md"))
	hasHighlight := strings.Contains(raw, "Important source point.")
	hasRawBody := strings.Contains(raw, "rough dump body")

	if !hasHighlight || !hasRawBody {
		t.Fatalf("raw dump = %q, want highlight and original body", raw)
	}
}

func TestOpenAIDumpCurationProviderRecordsTracePayloads(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceWorkspaceQueryFixture(t, projectDir)

	t.Setenv(core.LLMTraceEnabledEnv, "1")
	t.Setenv(core.LLMTracePayloadsEnv, "1")
	t.Setenv(core.ActiveOperatorIDEnv, "test-user")

	server := newDumpCurationTraceTestServer(t)
	defer server.Close()

	provider := openAIDumpCurationProvider{
		ProjectDir: projectDir,
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Model:      "test-model",
	}

	result, err := provider.CurateDump(t.Context(), DumpCurationRequest{
		ProjectDir: projectDir,
		ProjectID:  "service-project",
		Title:      "Trace Dump",
		Content:    "rough body",
	})
	if err != nil {
		t.Fatalf("CurateDump returned error: %v", err)
	}

	if result.PolishedBody != "Polished body." {
		t.Fatalf("PolishedBody = %q", result.PolishedBody)
	}

	tracePath := filepath.Join(projectDir, workspacecore.DirectoryName, workspacecore.UserDirectoryName, "log", "llm", "llm.ndjson")

	trace, err := os.ReadFile(tracePath)
	if err != nil {
		t.Fatalf("read trace: %v", err)
	}

	assertDumpCurationTracePayloads(t, string(trace))
}

func TestCodexDumpCurationProviderUsesAgentRunner(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceWorkspaceQueryFixture(t, projectDir)

	runner := &fakeDumpCurationAgentRunner{
		result: harness.AgentRunResult{
			ID:     "dump-curation-trace-dump",
			Output: []byte(`{"polished_body":"Polished body.","highlights":["Source point."],"warnings":[]}`),
		},
	}
	provider := harnessDumpCurationProvider{
		ProjectDir: projectDir,
		Runner:     runner,
	}

	result, err := provider.CurateDump(t.Context(), DumpCurationRequest{
		ProjectDir: projectDir,
		ProjectID:  "service-project",
		Title:      "Trace Dump",
		Content:    "rough body",
	})
	if err != nil {
		t.Fatalf("CurateDump returned error: %v", err)
	}

	if result.PolishedBody != "Polished body." {
		t.Fatalf("PolishedBody = %q", result.PolishedBody)
	}

	if runner.request.ID != "dump-curation-trace-dump" {
		t.Fatalf("agent request id = %q, want dump-curation-trace-dump", runner.request.ID)
	}

	if !strings.Contains(runner.request.Prompt, "dump curation backend") {
		t.Fatalf("prompt does not identify dump curation backend:\n%s", runner.request.Prompt)
	}

	if len(runner.request.OutputSchema) == 0 {
		t.Fatal("agent output schema is empty")
	}
}

func newDumpCurationTraceTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/responses" {
			t.Fatalf("path = %q, want /responses", request.URL.Path)
		}

		writer.Header().Set("Content-Type", "application/json")

		_, err := writer.Write([]byte(`{
			"id": "response-dump",
			"output_text": "{\"polished_body\":\"Polished body.\",\"highlights\":[\"Source point.\"],\"warnings\":[]}"
		}`))
		if err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
}

func assertDumpCurationTracePayloads(t *testing.T, trace string) {
	t.Helper()

	for _, want := range []string{
		`"event":"llm_request"`,
		`"event":"llm_response"`,
		`"action":"dump_curation"`,
		`"request_payload"`,
		`"response_payload"`,
		`"output_payload"`,
	} {
		if !strings.Contains(trace, want) {
			t.Fatalf("trace missing %q:\n%s", want, trace)
		}
	}
}

func TestFallbackDumpCurationProviderPreservesBodyAndHighlights(t *testing.T) {
	provider := fallbackDumpCurationProvider{Reason: "llm_unavailable"}

	result, err := provider.CurateDump(t.Context(), DumpCurationRequest{
		Title:   "Rough Notes",
		Content: "Use API_TOKEN=abc123\n\n- Keep the dump useful.\n- Preserve intent.",
	})
	if err != nil {
		t.Fatalf("CurateDump returned error: %v", err)
	}

	if strings.Contains(result.PolishedBody, "abc123") {
		t.Fatalf("PolishedBody contains secret: %q", result.PolishedBody)
	}

	if !strings.Contains(result.PolishedBody, "API_TOKEN=[REDACTED]") {
		t.Fatalf("PolishedBody = %q, want redacted token", result.PolishedBody)
	}

	if len(result.Highlights) != 3 {
		t.Fatalf("Highlights = %#v, want three recoverable lines", result.Highlights)
	}

	if result.Warnings[0] != "llm_unavailable" {
		t.Fatalf("Warnings = %#v, want llm_unavailable", result.Warnings)
	}
}

func TestFallbackDumpCurationProviderHandlesEmptyContent(t *testing.T) {
	provider := fallbackDumpCurationProvider{}

	result, err := provider.CurateDump(t.Context(), DumpCurationRequest{Title: "Empty"})
	if err != nil {
		t.Fatalf("CurateDump returned error: %v", err)
	}

	if result.PolishedBody != "No dump content provided." {
		t.Fatalf("PolishedBody = %q", result.PolishedBody)
	}

	if len(result.Highlights) != 1 {
		t.Fatalf("Highlights = %#v, want one fallback highlight", result.Highlights)
	}
}

type fakeDumpCurationProvider struct {
	result DumpCurationResult
	err    error
}

type fakeDumpCurationAgentRunner struct {
	request harness.AgentRunRequest
	result  harness.AgentRunResult
	err     error
}

func (r *fakeDumpCurationAgentRunner) RunAgent(_ context.Context, request harness.AgentRunRequest) (harness.AgentRunResult, error) {
	r.request = request

	return r.result, r.err
}

func (p fakeDumpCurationProvider) CurateDump(
	_ context.Context,
	_ DumpCurationRequest,
) (DumpCurationResult, error) {
	return p.result, p.err
}
