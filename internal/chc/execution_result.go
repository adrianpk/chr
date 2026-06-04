package chc

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExecutionResultStatus defines a package boundary type.
type ExecutionResultStatus string

const (
	// ExecutionResultStatusSucceeded defines a package boundary constant.
	ExecutionResultStatusSucceeded ExecutionResultStatus = "succeeded"
	// ExecutionResultStatusFailed defines a package boundary constant.
	ExecutionResultStatusFailed ExecutionResultStatus = "failed"
)

// ExecutionResult describes a package boundary value.
type ExecutionResult struct {
	ExecutionID  string                       `json:"execution_id"`
	TicketKey    string                       `json:"ticket_key"`
	Status       ExecutionResultStatus        `json:"status"`
	Summary      string                       `json:"summary"`
	ChangedFiles []ExecutionResultChangedFile `json:"changed_files,omitempty"`
}

// ExecutionResultChangedFile describes a package boundary value.
type ExecutionResultChangedFile struct {
	Path    string `json:"path"`
	Summary string `json:"summary,omitempty"`
}

// LoadExecutionResult loads the corresponding package data.
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

	err = decoder.Decode(&result)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("decode execution result: %w", err)
	}

	err = ValidateExecutionResult(result)
	if err != nil {
		return ExecutionResult{}, err
	}

	return result, nil
}

// ValidateExecutionResult validates the corresponding package boundary value.
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

	err := validateExecutionResultChangedFiles(result.ChangedFiles)
	if err != nil {
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

// Valid reports whether the value is one of the supported constants.
func (s ExecutionResultStatus) Valid() bool {
	switch s {
	case ExecutionResultStatusSucceeded,
		ExecutionResultStatusFailed:
		return true
	default:
		return false
	}
}
