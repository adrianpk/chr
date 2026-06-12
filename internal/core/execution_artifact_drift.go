package core

import (
	"context"
	"fmt"
	"os"
)

// ExecutionArtifactDriftStatus defines a package boundary type.
type ExecutionArtifactDriftStatus string

const (
	// ExecutionArtifactDriftStatusOK defines a package boundary constant.
	ExecutionArtifactDriftStatusOK ExecutionArtifactDriftStatus = "ok"
	// ExecutionArtifactDriftStatusMissing defines a package boundary constant.
	ExecutionArtifactDriftStatusMissing ExecutionArtifactDriftStatus = "missing_artifacts"
)

// ExecutionArtifactDriftDiagnostic describes a package boundary value.
type ExecutionArtifactDriftDiagnostic struct {
	Execution ExecutionRecord
	Status    ExecutionArtifactDriftStatus
	Artifacts []ExecutionArtifactPathDiagnostic
}

// ExecutionArtifactPathDiagnostic describes a package boundary value.
type ExecutionArtifactPathDiagnostic struct {
	Name   string
	Path   string
	Exists bool
}

// ExecutionArtifactDriftService describes a package boundary value.
type ExecutionArtifactDriftService struct {
	repository ExecutionRepository
}

// NewExecutionArtifactDriftService constructs the corresponding package service or value.
func NewExecutionArtifactDriftService(repository ExecutionRepository) ExecutionArtifactDriftService {
	return ExecutionArtifactDriftService{repository: repository}
}

// ListDiagnostics lists the requested data.
func (s ExecutionArtifactDriftService) ListDiagnostics(
	ctx context.Context,
	projectID string,
) ([]ExecutionArtifactDriftDiagnostic, error) {
	records, err := s.repository.ListProjectExecutions(ctx, projectID)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]ExecutionArtifactDriftDiagnostic, 0, len(records))

	for _, record := range records {
		diagnostic, err := inspectExecutionArtifacts(record)
		if err != nil {
			return nil, err
		}

		diagnostics = append(diagnostics, diagnostic)
	}

	return diagnostics, nil
}

func inspectExecutionArtifacts(record ExecutionRecord) (ExecutionArtifactDriftDiagnostic, error) {
	artifacts := []ExecutionArtifactPathDiagnostic{
		{Name: "run_dir", Path: record.Artifacts.RunDir, Exists: false},
		{Name: "result", Path: record.Artifacts.ResultPath, Exists: false},
		{Name: "log", Path: record.Artifacts.LogPath, Exists: false},
		{Name: "validation", Path: record.Artifacts.ValidationPath, Exists: false},
		{Name: "report", Path: ExecutionReportPath(record), Exists: false},
	}

	for index, artifact := range artifacts {
		exists, err := executionArtifactExists(artifact.Name, artifact.Path)
		if err != nil {
			return ExecutionArtifactDriftDiagnostic{}, err
		}

		artifacts[index].Exists = exists
	}

	status := ExecutionArtifactDriftStatusOK

	for _, artifact := range artifacts {
		if !artifact.Exists {
			status = ExecutionArtifactDriftStatusMissing

			break
		}
	}

	return ExecutionArtifactDriftDiagnostic{
		Execution: record,
		Status:    status,
		Artifacts: artifacts,
	}, nil
}

func executionArtifactExists(name, path string) (bool, error) {
	if path == "" {
		return false, nil
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, fmt.Errorf("inspect execution artifact %q: %w", path, err)
	}

	if name == "run_dir" {
		return info.IsDir(), nil
	}

	return !info.IsDir(), nil
}
