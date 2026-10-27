package execution

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResultStatus defines a package boundary type.
type ResultStatus string

const (
	// ResultStatusSucceeded defines a package boundary constant.
	ResultStatusSucceeded ResultStatus = "succeeded"
	// ResultStatusFailed defines a package boundary constant.
	ResultStatusFailed ResultStatus = "failed"
)

// Result describes a package boundary value.
type Result struct {
	ExecutionID     string                `json:"execution_id"`
	TicketKey       string                `json:"ticket_key"`
	Status          ResultStatus          `json:"status"`
	Summary         string                `json:"summary"`
	EvidenceSummary ResultEvidenceSummary `json:"evidence_summary"`
	ChangedFiles    []ResultChangedFile   `json:"changed_files,omitempty"`
}

// ResultEvidenceSummary describes evidence produced by the worker.
type ResultEvidenceSummary struct {
	Strategy      string   `json:"strategy"`
	Produced      string   `json:"produced"`
	Validation    string   `json:"validation"`
	Deviations    []string `json:"deviations,omitempty"`
	RemainingRisk []string `json:"remaining_risk,omitempty"`
}

// ResultChangedFile describes a package boundary value.
type ResultChangedFile struct {
	Path    string `json:"path"`
	Summary string `json:"summary,omitempty"`
}

// LoadResult loads the corresponding package data.
func LoadResult(path string) (Result, error) {
	file, err := os.Open(path)
	if err != nil {
		return Result{}, fmt.Errorf("open execution result: %w", err)
	}

	defer func() {
		//nolint:errcheck // Read errors are reported by the decoder.
		_ = file.Close()
	}()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()

	var result Result

	err = decoder.Decode(&result)
	if err != nil {
		return Result{}, fmt.Errorf("decode execution result: %w", err)
	}

	err = ValidateResult(result)
	if err != nil {
		return Result{}, err
	}

	return result, nil
}

// ValidateResult validates the corresponding package boundary value.
func ValidateResult(result Result) error {
	if result.ExecutionID == "" {
		return fmt.Errorf("execution result id is required")
	}

	if result.TicketKey == "" {
		return fmt.Errorf("execution result ticket key is required")
	}

	if !result.Status.Valid() {
		return fmt.Errorf("execution result status is invalid: %s", result.Status)
	}

	err := validateExecutionResultEvidenceSummary(result.EvidenceSummary)
	if err != nil {
		return err
	}

	err = validateExecutionResultChangedFiles(result.ChangedFiles)
	if err != nil {
		return err
	}

	return nil
}

func validateExecutionResultEvidenceSummary(summary ResultEvidenceSummary) error {
	if strings.TrimSpace(summary.Strategy) == "" {
		return fmt.Errorf("execution result evidence_summary.strategy is required")
	}

	if strings.TrimSpace(summary.Produced) == "" {
		return fmt.Errorf("execution result evidence_summary.produced is required")
	}

	if strings.TrimSpace(summary.Validation) == "" {
		return fmt.Errorf("execution result evidence_summary.validation is required")
	}

	return nil
}

func validateExecutionResultChangedFiles(changedFiles []ResultChangedFile) error {
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
func (s ResultStatus) Valid() bool {
	switch s {
	case ResultStatusSucceeded,
		ResultStatusFailed:
		return true
	default:
		return false
	}
}
