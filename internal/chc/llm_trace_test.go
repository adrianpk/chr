package chc

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadLLMTraceConfigDisabledDoesNotRequireUserspace(t *testing.T) {
	dir := t.TempDir()

	config, err := LoadLLMTraceConfig(LoadLLMTraceConfigOptions{
		ProjectDir: dir,
		LookupEnv:  func(string) string { return "" },
	})
	if err != nil {
		t.Fatalf("LoadLLMTraceConfig returned error: %v", err)
	}

	if config.Enabled {
		t.Fatal("Enabled = true, want false")
	}
	if config.FilePath != "" {
		t.Fatalf("FilePath = %q, want empty", config.FilePath)
	}

	if _, err := os.Stat(filepath.Join(dir, DirectoryName, "user", "log", "llm")); !os.IsNotExist(err) {
		t.Fatalf("trace directory exists while tracing is disabled: %v", err)
	}
}

func TestLoadLLMTraceConfigEnabledResolvesUserspaceLogPath(t *testing.T) {
	dir := t.TempDir()

	config, err := LoadLLMTraceConfig(LoadLLMTraceConfigOptions{
		ProjectDir: dir,
		LookupEnv: traceTestLookupEnv(map[string]string{
			ActiveOperatorIDEnv: "adrian",
			LLMTraceEnabledEnv:  "1",
		}),
	})
	if err != nil {
		t.Fatalf("LoadLLMTraceConfig returned error: %v", err)
	}

	wantDir := filepath.Join(dir, DirectoryName, "user", "log", "llm")
	if config.Dir != wantDir {
		t.Fatalf("Dir = %q, want %q", config.Dir, wantDir)
	}
	if config.FilePath != filepath.Join(wantDir, defaultLLMTraceFileName) {
		t.Fatalf("FilePath = %q, want llm trace file under userspace", config.FilePath)
	}
}

func TestLoadLLMTraceConfigReadsEnvironmentOverrides(t *testing.T) {
	dir := t.TempDir()

	config, err := LoadLLMTraceConfig(LoadLLMTraceConfigOptions{
		ProjectDir: dir,
		LookupEnv: traceTestLookupEnv(map[string]string{
			ActiveOperatorIDEnv:     "adrian",
			LLMTraceEnabledEnv:      "true",
			LLMTracePayloadsEnv:     "yes",
			LLMTraceMaxBytesEnv:     "2048",
			LLMTraceKeepFilesEnv:    "3",
			"UNRELATED_ENVIRONMENT": "ignored",
		}),
	})
	if err != nil {
		t.Fatalf("LoadLLMTraceConfig returned error: %v", err)
	}

	if !config.Enabled {
		t.Fatal("Enabled = false, want true")
	}
	if !config.IncludePayloads {
		t.Fatal("IncludePayloads = false, want true")
	}
	if config.MaxBytes != 2048 {
		t.Fatalf("MaxBytes = %d, want 2048", config.MaxBytes)
	}
	if config.KeepFiles != 3 {
		t.Fatalf("KeepFiles = %d, want 3", config.KeepFiles)
	}
}

func TestWriteLLMTraceEventWritesJSONL(t *testing.T) {
	dir := t.TempDir()
	config := traceTestConfig(t, dir, true)

	err := WriteLLMTraceEvent(context.Background(), config, map[string]any{
		"event":       "llm_request",
		"llm_call_id": "llm-001",
		"payload":     map[string]any{"input": "hidden by default"},
	})
	if err != nil {
		t.Fatalf("WriteLLMTraceEvent returned error: %v", err)
	}

	content := readTraceFile(t, config.FilePath)
	if !strings.HasSuffix(content, "\n") {
		t.Fatalf("trace file does not end with newline: %q", content)
	}

	decoded := decodeTraceLine(t, content)
	if decoded["event"] != "llm_request" {
		t.Fatalf("event = %#v, want llm_request", decoded["event"])
	}
	if decoded["llm_call_id"] != "llm-001" {
		t.Fatalf("llm_call_id = %#v, want llm-001", decoded["llm_call_id"])
	}
	if _, ok := decoded["payload"]; ok {
		t.Fatalf("payload was written while payload capture is disabled: %#v", decoded)
	}
	if decoded["timestamp"] != "2026-05-27T10:00:00Z" {
		t.Fatalf("timestamp = %#v, want fixed test timestamp", decoded["timestamp"])
	}
}

func TestWriteLLMTraceEventIncludesPayloadsWhenEnabled(t *testing.T) {
	dir := t.TempDir()
	config := traceTestConfig(t, dir, true)
	config.IncludePayloads = true

	err := WriteLLMTraceEvent(context.Background(), config, map[string]any{
		"event":   "llm_response",
		"payload": map[string]any{"output": "visible"},
	})
	if err != nil {
		t.Fatalf("WriteLLMTraceEvent returned error: %v", err)
	}

	decoded := decodeTraceLine(t, readTraceFile(t, config.FilePath))
	payload, ok := decoded["payload"].(map[string]any)
	if !ok {
		t.Fatalf("payload = %#v, want object", decoded["payload"])
	}
	if payload["output"] != "visible" {
		t.Fatalf("payload output = %#v, want visible", payload["output"])
	}
}

func TestWriteLLMTraceEventRedactsSecretFieldsAndEnvironmentValues(t *testing.T) {
	dir := t.TempDir()
	config := traceTestConfig(t, dir, true)
	config.IncludePayloads = true
	config.Environ = func() []string {
		return []string{
			"OPENAI_API_KEY=env-secret",
			"CHC_OTHER=value",
		}
	}

	err := WriteLLMTraceEvent(context.Background(), config, map[string]any{
		"event": "llm_request",
		"headers": map[string]any{
			"Authorization": "Bearer direct-secret",
		},
		"payload": map[string]any{
			"message": "contains env-secret value",
			"api_key": "direct-key",
		},
	})
	if err != nil {
		t.Fatalf("WriteLLMTraceEvent returned error: %v", err)
	}

	content := readTraceFile(t, config.FilePath)
	if strings.Contains(content, "env-secret") {
		t.Fatalf("trace file leaked environment secret: %s", content)
	}
	if strings.Contains(content, "direct-secret") {
		t.Fatalf("trace file leaked bearer secret: %s", content)
	}
	if strings.Contains(content, "direct-key") {
		t.Fatalf("trace file leaked secret field: %s", content)
	}
	if !strings.Contains(content, "[REDACTED]") {
		t.Fatalf("trace file does not contain redaction marker: %s", content)
	}
}

func TestWriteLLMTraceEventRotatesTraceFile(t *testing.T) {
	dir := t.TempDir()
	config := traceTestConfig(t, dir, true)
	config.IncludePayloads = true
	config.MaxBytes = 150
	config.KeepFiles = 2

	for index := 0; index < 4; index++ {
		err := WriteLLMTraceEvent(context.Background(), config, map[string]any{
			"event":       "llm_request",
			"llm_call_id": "llm-rotation",
			"payload":     strings.Repeat("x", 120),
		})
		if err != nil {
			t.Fatalf("WriteLLMTraceEvent %d returned error: %v", index, err)
		}
	}

	assertFileExists(t, config.FilePath)
	assertFileExists(t, config.FilePath+".1")
	assertFileExists(t, config.FilePath+".2")
	if _, err := os.Stat(config.FilePath + ".3"); !os.IsNotExist(err) {
		t.Fatalf("unexpected third rotated file: %v", err)
	}
}

func TestRecordLLMTraceEventIsBestEffort(t *testing.T) {
	RecordLLMTraceEvent(context.Background(), LoadLLMTraceConfigOptions{
		ProjectDir: t.TempDir(),
		LookupEnv: traceTestLookupEnv(map[string]string{
			LLMTraceEnabledEnv: "1",
		}),
	}, map[string]any{
		"event": "llm_request",
	})
}

func traceTestConfig(t *testing.T, dir string, enabled bool) LLMTraceConfig {
	t.Helper()

	enabledValue := ""
	if enabled {
		enabledValue = "1"
	}

	config, err := LoadLLMTraceConfig(LoadLLMTraceConfigOptions{
		ProjectDir: dir,
		LookupEnv: traceTestLookupEnv(map[string]string{
			ActiveOperatorIDEnv: "adrian",
			LLMTraceEnabledEnv:  enabledValue,
		}),
		Environ: func() []string { return nil },
		Now: func() time.Time {
			return time.Date(2026, 5, 27, 10, 0, 0, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatalf("LoadLLMTraceConfig returned error: %v", err)
	}

	return config
}

func traceTestLookupEnv(values map[string]string) func(string) string {
	return func(key string) string {
		return values[key]
	}
}

func readTraceFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read trace file: %v", err)
	}

	return string(content)
}

func decodeTraceLine(t *testing.T, content string) map[string]any {
	t.Helper()

	var decoded map[string]any
	err := json.Unmarshal([]byte(strings.TrimSpace(content)), &decoded)
	if err != nil {
		t.Fatalf("decode trace line: %v", err)
	}

	return decoded
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file %s to exist: %v", path, err)
	}
}
