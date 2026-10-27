package execution

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// WorktreeMetadataSchemaVersion defines a package boundary constant.
	WorktreeMetadataSchemaVersion = 1
)

// WorktreeMetadata describes a package boundary value.
type WorktreeMetadata struct {
	SchemaVersion       int       `json:"schema_version"`
	TicketKey           string    `json:"ticket_key"`
	Branch              string    `json:"branch"`
	BaseRef             string    `json:"base_ref"`
	BaseSHA             string    `json:"base_sha"`
	WorktreePath        string    `json:"worktree_path"`
	LastExecutionID     string    `json:"last_execution_id,omitempty"`
	LastExecutionStatus string    `json:"last_execution_status,omitempty"`
	LastExecutionRunDir string    `json:"last_execution_run_dir,omitempty"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// NewWorktreeMetadata constructs the corresponding package service or value.
func NewWorktreeMetadata(plan WorktreePlan, updatedAt time.Time) WorktreeMetadata {
	return WorktreeMetadata{
		SchemaVersion: WorktreeMetadataSchemaVersion,
		TicketKey:     plan.TicketKey,
		Branch:        plan.Branch,
		BaseRef:       plan.BaseRef,
		BaseSHA:       plan.BaseSHA,
		WorktreePath:  plan.WorktreePath,
		UpdatedAt:     updatedAt.UTC(),
	}
}

// WorktreeMetadataPath executes the corresponding package operation.
func WorktreeMetadataPath(plan WorktreePlan) (string, error) {
	component := WorktreePathComponent(plan.TicketKey)
	if component == "" {
		return "", fmt.Errorf("ticket key %q cannot be used as a metadata path component", plan.TicketKey)
	}

	return filepath.Join(plan.WorktreeRoot, ".metadata", component+".json"), nil
}

// ReadWorktreeMetadata loads the corresponding package data.
func ReadWorktreeMetadata(path string) (WorktreeMetadata, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return WorktreeMetadata{}, fmt.Errorf("read worktree metadata: %w", err)
	}

	var metadata WorktreeMetadata
	if err := json.Unmarshal(content, &metadata); err != nil {
		return WorktreeMetadata{}, fmt.Errorf("decode worktree metadata: %w", err)
	}

	return metadata, nil
}

// WriteWorktreeMetadata writes the corresponding package data.
func WriteWorktreeMetadata(path string, metadata WorktreeMetadata) error {
	content, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("encode worktree metadata: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create worktree metadata parent: %w", err)
	}

	return writeWorktreeMetadataFile(path, content)
}

func writeWorktreeMetadataFile(path string, content []byte) error {
	content = append(content, '\n')
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("write worktree metadata: %w", err)
	}

	return nil
}
