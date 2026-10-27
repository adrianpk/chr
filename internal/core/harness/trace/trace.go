// Package trace owns bounded host model traces, payload policy and secret redaction.
package trace

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

	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

const (
	// EnabledEnv defines a package boundary constant.
	EnabledEnv = "CHR_LLM_TRACE"
	// PayloadsEnv defines a package boundary constant.
	PayloadsEnv = "CHR_LLM_TRACE_PAYLOADS"
	// MaxBytesEnv defines a package boundary constant.
	MaxBytesEnv = "CHR_LLM_TRACE_MAX_BYTES"
	// KeepFilesEnv defines a package boundary constant.
	KeepFilesEnv = "CHR_LLM_TRACE_KEEP_FILES"

	defaultFileName  = "llm.ndjson"
	defaultMaxBytes  = 10 * 1024 * 1024
	defaultKeepFiles = 5
)

var secretLikeFieldPattern = regexp.MustCompile(`(?i)(token|secret|password|authorization|api[_-]?key)`)

// Config describes a package boundary value.
type Config struct {
	Enabled         bool
	IncludePayloads bool
	Dir             string
	FilePath        string
	MaxBytes        int64
	KeepFiles       int
	Environ         func() []string
	Now             func() time.Time
}

// LoadOptions describes a package boundary value.
type LoadOptions struct {
	ProjectDir string
	LookupEnv  func(string) string
	Environ    func() []string
	Now        func() time.Time
}

// Load loads the corresponding package data.
func Load(opts LoadOptions) (Config, error) {
	lookupEnv := opts.LookupEnv
	if lookupEnv == nil {
		lookupEnv = os.Getenv
	}

	config := Config{
		Enabled:         truthyEnv(lookupEnv(EnabledEnv)),
		IncludePayloads: truthyEnv(lookupEnv(PayloadsEnv)),
		MaxBytes:        envInt64(lookupEnv(MaxBytesEnv), defaultMaxBytes),
		KeepFiles:       envInt(lookupEnv(KeepFilesEnv), defaultKeepFiles),
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

	userspace, err := workspacecore.ResolveUserspace(workspacecore.ResolveUserspaceOptions{
		ProjectDir: opts.ProjectDir,
		LookupEnv:  lookupEnv,
	})
	if err != nil {
		return Config{}, fmt.Errorf("resolve LLM trace userspace: %w", err)
	}

	config.Dir = filepath.Join(userspace.Path, "log", "llm")
	config.FilePath = filepath.Join(config.Dir, defaultFileName)

	return config, nil
}

// Record executes the corresponding package operation.
func Record(ctx context.Context, opts LoadOptions, event map[string]any) {
	config, err := Load(opts)
	if err != nil {
		return
	}

	if err := Write(ctx, config, event); err != nil {
		return
	}
}

// Write executes the corresponding package operation.
func Write(ctx context.Context, config Config, event map[string]any) error {
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

	line, err := encodeEvent(config, event)
	if err != nil {
		return err
	}

	err = os.MkdirAll(config.Dir, 0o755)
	if err != nil {
		return fmt.Errorf("create LLM trace directory: %w", err)
	}

	err = rotateFile(config.FilePath, config.MaxBytes, config.KeepFiles, int64(len(line)))
	if err != nil {
		return fmt.Errorf("rotate LLM trace file: %w", err)
	}

	err = appendLine(config.FilePath, line)
	if err != nil {
		return err
	}

	return nil
}

func appendLine(filePath string, line []byte) error {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open LLM trace file: %w", err)
	}

	_, err = file.Write(line)
	if err != nil {
		closeErr := file.Close()
		if closeErr != nil {
			return fmt.Errorf("write LLM trace event: %w; close LLM trace file: %w", err, closeErr)
		}

		return fmt.Errorf("write LLM trace event: %w", err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("close LLM trace file: %w", err)
	}

	return nil
}

func encodeEvent(config Config, event map[string]any) ([]byte, error) {
	normalized := make(map[string]any, len(event)+1)
	secretValues := secretValues(config.Environ)

	for key, value := range event {
		if value == nil {
			continue
		}

		if !config.IncludePayloads && payloadField(key) {
			continue
		}

		normalized[key] = redactValue(key, value, secretValues)
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

func rotateFile(filePath string, maxBytes int64, keepFiles int, incomingBytes int64) error {
	maxBytes = normalizedMaxBytes(maxBytes)
	keepFiles = normalizedKeepFiles(keepFiles)

	info, err := os.Stat(filePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if os.IsNotExist(err) {
		return nil
	}

	if info.Size()+incomingBytes <= maxBytes {
		return nil
	}

	return rotateExistingFile(filePath, keepFiles)
}

func normalizedMaxBytes(maxBytes int64) int64 {
	if maxBytes > 0 {
		return maxBytes
	}

	return defaultMaxBytes
}

func normalizedKeepFiles(keepFiles int) int {
	if keepFiles > 0 {
		return keepFiles
	}

	return defaultKeepFiles
}

func rotateExistingFile(filePath string, keepFiles int) error {
	err := removeRotation(filePath, keepFiles)
	if err != nil {
		return err
	}

	for index := keepFiles - 1; index >= 1; index-- {
		err = renameRotation(filePath, index, index+1)
		if err != nil {
			return err
		}
	}

	return renameRotation(filePath, 0, 1)
}

func removeRotation(filePath string, index int) error {
	err := os.Remove(rotationPath(filePath, index))
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func renameRotation(filePath string, sourceIndex, targetIndex int) error {
	err := os.Rename(rotationPath(filePath, sourceIndex), rotationPath(filePath, targetIndex))
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func rotationPath(filePath string, index int) string {
	if index == 0 {
		return filePath
	}

	return filePath + "." + strconv.Itoa(index)
}

func payloadField(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))

	return normalized == "payload" ||
		normalized == "request_payload" ||
		normalized == "response_payload" ||
		normalized == "output_payload" ||
		normalized == "raw_request" ||
		normalized == "raw_response"
}

func redactValue(key string, value any, secretValues []string) any {
	if secretLikeFieldPattern.MatchString(key) {
		return "[REDACTED]"
	}

	switch typed := value.(type) {
	case map[string]any:
		redacted := make(map[string]any, len(typed))
		for childKey, childValue := range typed {
			redacted[childKey] = redactValue(childKey, childValue, secretValues)
		}

		return redacted
	case []any:
		redacted := make([]any, 0, len(typed))
		for _, childValue := range typed {
			redacted = append(redacted, redactValue("", childValue, secretValues))
		}

		return redacted
	case string:
		return redactString(typed, secretValues)
	default:
		return value
	}
}

func redactString(value string, secretValues []string) string {
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

func secretValues(environ func() []string) []string {
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
