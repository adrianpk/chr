package chc

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ExecutionResultStatus string

const (
	ExecutionResultStatusSucceeded ExecutionResultStatus = "succeeded"
	ExecutionResultStatusFailed    ExecutionResultStatus = "failed"
)

type ExecutionResult struct {
	ExecutionID  string                       `json:"execution_id"`
	TicketKey    string                       `json:"ticket_key"`
	Status       ExecutionResultStatus        `json:"status"`
	Summary      string                       `json:"summary"`
	ChangedFiles []ExecutionResultChangedFile `json:"changed_files,omitempty"`
}

type ExecutionResultChangedFile struct {
	Path    string `json:"path"`
	Summary string `json:"summary,omitempty"`
}

func LoadExecutionResult(path string) (ExecutionResult, error) {
	file, err := os.Open(path)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("open execution result: %w", err)
	}

	defer func() {
		//nolint:errcheck // Read errors are reported by the decoder.
		_ = file.Close()
	}()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()

	var result ExecutionResult
	if err := decoder.Decode(&result); err != nil {
		return ExecutionResult{}, fmt.Errorf("decode execution result: %w", err)
	}

	if err := ValidateExecutionResult(result); err != nil {
		return ExecutionResult{}, err
	}

	return result, nil
}

func ValidateExecutionResult(result ExecutionResult) error {
	if result.ExecutionID == "" {
		return fmt.Errorf("execution result id is required")
	}

	if result.TicketKey == "" {
		return fmt.Errorf("execution result ticket key is required")
	}

	if !result.Status.Valid() {
		return fmt.Errorf("execution result status is invalid: %s", result.Status)
	}

	if err := validateExecutionResultChangedFiles(result.ChangedFiles); err != nil {
		return err
	}

	return nil
}

func validateExecutionResultChangedFiles(changedFiles []ExecutionResultChangedFile) error {
	for _, changedFile := range changedFiles {
		path := strings.TrimSpace(changedFile.Path)
		if path == "" {
			return fmt.Errorf("execution result changed file path is required")
		}

		if filepath.IsAbs(path) || strings.HasPrefix(filepath.Clean(path), "..") {
			return fmt.Errorf("execution result changed file path must be workspace relative: %s", changedFile.Path)
		}
	}

	return nil
}

func (s ExecutionResultStatus) Valid() bool {
	switch s {
	case ExecutionResultStatusSucceeded,
		ExecutionResultStatusFailed:
		return true
	default:
		return false
	}
}
