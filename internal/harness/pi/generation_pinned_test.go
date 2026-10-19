package pi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/harness"
)

// An explicitly supplied pinned binary uses only an owned loopback model endpoint.
func TestPiGenerationPinnedInspection(t *testing.T) {
	program := os.Getenv("CHR_TEST_PI_EXECUTABLE")
	if program == "" {
		t.Skip("set CHR_TEST_PI_EXECUTABLE to the prepared Pi 1.3.0 binary for local protocol certification")
	}

	project := t.TempDir()
	private := filepath.Join(project, "selected-auth.json")
	outside := filepath.Join(t.TempDir(), "outside")
	evidence := filepath.Join(t.TempDir(), "evidence")
	writePiFixtureFile(t, filepath.Join(project, "decision.md"), "Retain the bounded host policy.")
	writePiFixtureFile(t, outside, "Outside the inspection capability.")
	writePiFixtureFile(t, evidence, "Exact host-provided evidence.")
	writePiFixtureFile(t, private, `{"chr-fixture":{"type":"api_key","key":"selected-fixture-only"}}`)
	preparePiInspectionFiles(t, project, private, outside)

	for _, scenario := range []struct{ name, action, path, expected string }{
		{"read", "read", "decision.md", "Retain the bounded host policy."},
		{"outside", "read", outside, "Read-only inspection failed"},
		{"evidence", "read", evidence, "Exact host-provided evidence."},
		{"credentials", "read", private, "Read-only inspection failed"},
		{"link escape", "read", "escape", "Read-only inspection failed"},
		{"credential alias", "read", "alias", "Read-only inspection failed"},
		{"oversized", "read", "oversized", "Read-only inspection failed"},
		{"pipe", "read", "pipe", "Read-only inspection failed"},
		{"invalid text", "read", "binary", "Read-only inspection failed"},
		{"list", "list", ".", "decision.md"},
		{"find", "find", ".", "decision.md"},
		{"unsupported", "write", "decision.md", "Validation failed for tool"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			endpoint, requests := pinnedPiEndpoint(t, scenario.action, scenario.path, scenario.expected)
			models := filepath.Join(t.TempDir(), "models.json")
			writePiFixtureFile(t, models, fmt.Sprintf(`{"providers":{"chr-fixture":{"api":"openai-completions","baseUrl":%q,"models":[{"id":"model","contextWindow":32768,"maxTokens":2048}]}}}`, endpoint.URL+"/v1"))

			resources, err := ResolveResources("", private, models, "chr-fixture")
			if err != nil {
				t.Fatal(err)
			}

			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()

			result, err := (Runner{}).RunAgent(ctx, harness.AgentRunRequest{
				ID: "pinned", Purpose: "test", ProjectDir: project, RunDir: t.TempDir(), Program: program,
				FrozenSelection: true, ModelProvider: "chr-fixture", Model: "model", Thinking: "off", Resources: resources,
				Prompt: "Return the host fixture result.", OutputName: "output.json", ReadOnlyFiles: true,
				ReadOnlyRefs: []string{evidence},
			})
			if err != nil || string(result.Output) != `{"ok":true}` || requests.Load() != 2 {
				t.Fatalf("pinned Pi inspection failed: %v; calls=%d; stderr=%s", err, requests.Load(), result.Stderr)
			}
		})
	}
}

func preparePiInspectionFiles(t *testing.T, project, private, outside string) {
	t.Helper()
	writePiFixtureFile(t, filepath.Join(project, "oversized"), strings.Repeat("x", 65537))
	writePiFixtureFile(t, filepath.Join(project, "binary"), "\xff")

	if err := os.Symlink(outside, filepath.Join(project, "escape")); err != nil {
		t.Fatal(err)
	}

	if err := os.Link(private, filepath.Join(project, "alias")); err != nil {
		t.Fatal(err)
	}

	if err := syscall.Mkfifo(filepath.Join(project, "pipe"), 0o600); err != nil {
		t.Fatal(err)
	}

	writePiFixtureFile(t, filepath.Join(project, "AGENTS.md"), "POISON_CONTEXT_MUST_NOT_BE_DISCOVERED")

	poison := filepath.Join(project, ".pi", "extensions")
	if err := os.MkdirAll(poison, 0o700); err != nil {
		t.Fatal(err)
	}

	writePiFixtureFile(t, filepath.Join(poison, "poison.js"), "export default function(){throw new Error('POISON_CONTEXT_MUST_NOT_BE_DISCOVERED')}\n")
}

type pinnedPiRequest struct {
	Tools    []struct{ Function struct{ Name string } }
	Messages []struct {
		Role    string
		Content json.RawMessage
	}
}

func pinnedPiEndpoint(t *testing.T, action, path, expected string) (*httptest.Server, *atomic.Int64) {
	t.Helper()

	var calls atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, request, valid := readPinnedPiRequest(t, r)
		if !valid || calls.Load() >= 2 {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		w.Header().Set("Content-Type", "text/event-stream")

		if calls.Add(1) == 1 {
			arguments := pinnedPiJSON(t, map[string]string{"action": action, "path": path})
			delta := map[string]any{"role": "assistant", "tool_calls": []any{map[string]any{"index": 0, "id": "inspect-1", "type": "function", "function": map[string]any{"name": "chr_inspect", "arguments": string(arguments)}}}}
			writePinnedPiChunk(t, w, delta, "tool_calls")

			return
		}

		var toolResult string

		for _, message := range request.Messages {
			if message.Role == "tool" {
				toolResult = string(message.Content)
			}
		}

		if !strings.Contains(toolResult, expected) || strings.Contains(string(data), "selected-fixture-only") {
			t.Error("read-only result lost its boundary")
		}

		writePinnedPiChunk(t, w, map[string]any{"role": "assistant", "content": `{"ok":true}`}, "stop")
	}))
	t.Cleanup(server.Close)

	return server, &calls
}

func readPinnedPiRequest(t *testing.T, r *http.Request) ([]byte, pinnedPiRequest, bool) {
	t.Helper()

	data, err := io.ReadAll(io.LimitReader(r.Body, harness.MaxInput+1))

	var request pinnedPiRequest

	if err != nil || len(data) > harness.MaxInput || r.URL.Path != "/v1/chat/completions" || r.Method != http.MethodPost {
		t.Error("invalid owned model request")

		return nil, request, false
	}

	if err := json.Unmarshal(data, &request); err != nil || len(request.Tools) != 1 || request.Tools[0].Function.Name != "chr_inspect" {
		t.Error("pinned Pi acquired unexpected tools")

		return nil, request, false
	}

	if r.Header.Get("Authorization") != "Bearer selected-fixture-only" || strings.Contains(string(data), "POISON_CONTEXT_MUST_NOT_BE_DISCOVERED") {
		t.Error("pinned Pi imported unintended context or credentials")

		return nil, request, false
	}

	return data, request, true
}

func writePinnedPiChunk(t *testing.T, w http.ResponseWriter, delta map[string]any, finish string) {
	t.Helper()

	chunk := map[string]any{"id": "owned", "object": "chat.completion.chunk", "model": "model", "choices": []any{map[string]any{"index": 0, "delta": delta, "finish_reason": finish}}}
	if _, err := fmt.Fprintf(w, "data: %s\n\ndata: [DONE]\n\n", pinnedPiJSON(t, chunk)); err != nil {
		t.Error(err)
	}
}

func pinnedPiJSON(t *testing.T, value any) []byte {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Error(err)
	}

	return data
}

func writePiFixtureFile(t *testing.T, path, value string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
		t.Fatal(err)
	}
}
