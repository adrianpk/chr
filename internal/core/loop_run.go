package core

import (
	"fmt"
	"time"
)

// LoopRunStatus defines a package boundary type.
type LoopRunStatus string

const (
	// LoopRunStatusPending defines a package boundary constant.
	LoopRunStatusPending LoopRunStatus = "pending"
	// LoopRunStatusRunning defines a package boundary constant.
	LoopRunStatusRunning LoopRunStatus = "running"
	// LoopRunStatusSucceeded defines a package boundary constant.
	LoopRunStatusSucceeded LoopRunStatus = "succeeded"
	// LoopRunStatusFailed defines a package boundary constant.
	LoopRunStatusFailed LoopRunStatus = "failed"
	// LoopRunStatusBlocked defines a package boundary constant.
	LoopRunStatusBlocked LoopRunStatus = "blocked"
)

// LoopRunRecord describes a package boundary value.
type LoopRunRecord struct {
	ID             string
	ArtifactID     string
	ArtifactDir    string
	ProjectID      string
	Status         LoopRunStatus
	EligibleCount  int
	ProcessedCount int
	BlockedCount   int
	FailedCount    int
	FailureReason  string
	CreatedAt      time.Time
	StartedAt      *time.Time
	FinishedAt     *time.Time
	UpdatedAt      time.Time
}

// ValidateLoopRunRecord validates the corresponding package boundary value.
func ValidateLoopRunRecord(record LoopRunRecord) error {
	if record.ID == "" {
		return fmt.Errorf("loop run id is required")
	}

	if record.ProjectID == "" {
		return fmt.Errorf("loop run project id is required")
	}

	if !record.Status.Valid() {
		return fmt.Errorf("loop run status is invalid: %s", record.Status)
	}

	err := validateLoopRunCounts(record)
	if err != nil {
		return err
	}

	return validateLoopRunTimes(record)
}

// Valid reports whether the value is one of the supported constants.
func (s LoopRunStatus) Valid() bool {
	switch s {
	case LoopRunStatusPending,
		LoopRunStatusRunning,
		LoopRunStatusSucceeded,
		LoopRunStatusFailed,
		LoopRunStatusBlocked:
		return true
	default:
		return false
	}
}

func validateLoopRunCounts(record LoopRunRecord) error {
	for _, count := range []struct {
		name  string
		value int
	}{
		{name: "eligible", value: record.EligibleCount},
		{name: "processed", value: record.ProcessedCount},
		{name: "blocked", value: record.BlockedCount},
		{name: "failed", value: record.FailedCount},
	} {
		if count.value < 0 {
			return fmt.Errorf("loop run %s count cannot be negative", count.name)
		}
	}

	return nil
}

func validateLoopRunTimes(record LoopRunRecord) error {
	if record.CreatedAt.IsZero() {
		return fmt.Errorf("loop run created time is required")
	}

	if record.UpdatedAt.IsZero() {
		return fmt.Errorf("loop run updated time is required")
	}

	if record.StartedAt != nil && record.FinishedAt != nil && record.FinishedAt.Before(*record.StartedAt) {
		return fmt.Errorf("loop run finished before it started")
	}

	return nil
}
