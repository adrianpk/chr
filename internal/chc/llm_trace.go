package chc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	LLMTraceEnabledEnv   = "CHC_LLM_TRACE"
	LLMTracePayloadsEnv  = "CHC_LLM_TRACE_PAYLOADS"
	LLMTraceMaxBytesEnv  = "CHC_LLM_TRACE_MAX_BYTES"
	LLMTraceKeepFilesEnv = "CHC_LLM_TRACE_KEEP_FILES"

	defaultLLMTraceFileName  = "llm.ndjson"
	defaultLLMTraceMaxBytes  = 10 * 1024 * 1024
	defaultLLMTraceKeepFiles = 5
)

var secretLikeFieldPattern = regexp.MustCompile(`(?i)(token|secret|password|authorization|api[_-]?key)`)

type LLMTraceConfig struct {
	Enabled         bool
	IncludePayloads bool
	Dir             string
	FilePath        string
	MaxBytes        int64
	KeepFiles       int
	Environ         func() []string
	Now             func() time.Time
}

type LoadLLMTraceConfigOptions struct {
	ProjectDir string
	LookupEnv  func(string) string
	Environ    func() []string
	Now        func() time.Time
}

func LoadLLMTraceConfig(opts LoadLLMTraceConfigOptions) (LLMTraceConfig, error) {
	lookupEnv := opts.LookupEnv
	if lookupEnv == nil {
		lookupEnv = os.Getenv
	}

	config := LLMTraceConfig{
		Enabled:         truthyEnv(lookupEnv(LLMTraceEnabledEnv)),
		IncludePayloads: truthyEnv(lookupEnv(LLMTracePayloadsEnv)),
		MaxBytes:        envInt64(lookupEnv(LLMTraceMaxBytesEnv), defaultLLMTraceMaxBytes),
		KeepFiles:       envInt(lookupEnv(LLMTraceKeepFilesEnv), defaultLLMTraceKeepFiles),
		Environ:         opts.Environ,
		Now:             opts.Now,
	}
	if config.Environ == nil {
		config.Environ = os.Environ
	}
	if config.Now == nil {
		config.Now = time.Now
	}

	if !config.Enabled {
		return config, nil
	}

	userspace, err := ResolveProjectUserspace(ResolveProjectUserspaceOptions{
		ProjectDir: opts.ProjectDir,
		LookupEnv:  lookupEnv,
	})
	if err != nil {
		return LLMTraceConfig{}, fmt.Errorf("resolve LLM trace userspace: %w", err)
	}

	config.Dir = filepath.Join(userspace.Path, "log", "llm")
	config.FilePath = filepath.Join(config.Dir, defaultLLMTraceFileName)
	return config, nil
}

func RecordLLMTraceEvent(ctx context.Context, opts LoadLLMTraceConfigOptions, event map[string]any) {
	config, err := LoadLLMTraceConfig(opts)
	if err != nil {
		return
	}

	_ = WriteLLMTraceEvent(ctx, config, event)
}

func WriteLLMTraceEvent(ctx context.Context, config LLMTraceConfig, event map[string]any) error {
	err := ctx.Err()
	if err != nil {
		return err
	}

	if !config.Enabled {
		return nil
	}

	if config.Dir == "" || config.FilePath == "" {
		return fmt.Errorf("LLM trace path is required when tracing is enabled")
	}

	line, err := encodeLLMTraceEvent(config, event)
	if err != nil {
		return err
	}

	err = os.MkdirAll(config.Dir, 0o755)
	if err != nil {
		return fmt.Errorf("create LLM trace directory: %w", err)
	}

	err = rotateLLMTraceFile(config.FilePath, config.MaxBytes, config.KeepFiles, int64(len(line)))
	if err != nil {
		return fmt.Errorf("rotate LLM trace file: %w", err)
	}

	file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open LLM trace file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(line)
	if err != nil {
		return fmt.Errorf("write LLM trace event: %w", err)
	}

	return nil
}

func encodeLLMTraceEvent(config LLMTraceConfig, event map[string]any) ([]byte, error) {
	normalized := make(map[string]any, len(event)+1)
	secretValues := llmTraceSecretValues(config.Environ)
	for key, value := range event {
		if value == nil {
			continue
		}
		if !config.IncludePayloads && llmTracePayloadField(key) {
			continue
		}

		normalized[key] = redactLLMTraceValue(key, value, secretValues)
	}

	if _, ok := normalized["timestamp"]; !ok {
		now := time.Now
		if config.Now != nil {
			now = config.Now
		}
		normalized["timestamp"] = now().UTC().Format(time.RFC3339Nano)
	}

	content, err := json.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("marshal LLM trace event: %w", err)
	}

	return append(content, '\n'), nil
}

func rotateLLMTraceFile(filePath string, maxBytes int64, keepFiles int, incomingBytes int64) error {
	if maxBytes <= 0 {
		maxBytes = defaultLLMTraceMaxBytes
	}
	if keepFiles <= 0 {
		keepFiles = defaultLLMTraceKeepFiles
	}

	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	if info.Size()+incomingBytes <= maxBytes {
		return nil
	}

	err = os.Remove(filePath + "." + strconv.Itoa(keepFiles))
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	for index := keepFiles - 1; index >= 1; index-- {
		source := filePath + "." + strconv.Itoa(index)
		target := filePath + "." + strconv.Itoa(index+1)

		err = os.Rename(source, target)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	err = os.Rename(filePath, filePath+".1")
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func llmTracePayloadField(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	return normalized == "payload" ||
		normalized == "request_payload" ||
		normalized == "response_payload" ||
		normalized == "output_payload" ||
		normalized == "raw_request" ||
		normalized == "raw_response"
}

func redactLLMTraceValue(key string, value any, secretValues []string) any {
	if secretLikeFieldPattern.MatchString(key) {
		return "[REDACTED]"
	}

	switch typed := value.(type) {
	case map[string]any:
		redacted := make(map[string]any, len(typed))
		for childKey, childValue := range typed {
			redacted[childKey] = redactLLMTraceValue(childKey, childValue, secretValues)
		}
		return redacted
	case []any:
		redacted := make([]any, 0, len(typed))
		for _, childValue := range typed {
			redacted = append(redacted, redactLLMTraceValue("", childValue, secretValues))
		}
		return redacted
	case string:
		return redactLLMTraceString(typed, secretValues)
	default:
		return value
	}
}

func redactLLMTraceString(value string, secretValues []string) string {
	redacted := redactBearerToken(value)
	for _, secret := range secretValues {
		if secret == "" {
			continue
		}

		redacted = strings.ReplaceAll(redacted, secret, "[REDACTED]")
	}

	return redacted
}

func redactBearerToken(value string) string {
	fields := strings.Fields(value)
	if len(fields) >= 2 && strings.EqualFold(fields[0], "Bearer") {
		return "Bearer [REDACTED]"
	}

	return value
}

func llmTraceSecretValues(environ func() []string) []string {
	if environ == nil {
		return nil
	}

	values := make([]string, 0)
	for _, entry := range environ() {
		name, value, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		if !secretLikeFieldPattern.MatchString(name) || value == "" {
			continue
		}

		values = append(values, value)
	}

	return values
}

func truthyEnv(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func envInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}

func envInt64(value string, fallback int64) int64 {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}
