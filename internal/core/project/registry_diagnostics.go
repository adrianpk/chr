package project

import (
	"context"
	"errors"
	"os"
	"path/filepath"
)

// Diagnostic describes a package boundary value.
type Diagnostic struct {
	Project         Project
	Active          bool
	PathExists      bool
	WorkspaceExists bool
	Status          DiagnosticStatus
}

// DiagnosticStatus defines a package boundary type.
type DiagnosticStatus string

const (
	// DiagnosticOK defines a package boundary constant.
	DiagnosticOK DiagnosticStatus = "ok"
	// DiagnosticDisabled defines a package boundary constant.
	DiagnosticDisabled DiagnosticStatus = "disabled"
	// DiagnosticMissingPath defines a package boundary constant.
	DiagnosticMissingPath DiagnosticStatus = "missing_path"
	// DiagnosticNoWorkspace defines a package boundary constant.
	DiagnosticNoWorkspace DiagnosticStatus = "missing_workspace"
)

// DiagnosticsService describes a package boundary value.
type DiagnosticsService struct {
	registry           Registry
	store              ActiveStore
	workspaceDirectory string
}

// NewDiagnosticsService constructs the corresponding package service or value.
func NewDiagnosticsService(registry Registry, store ActiveStore, workspaceDirectory string) DiagnosticsService {
	return DiagnosticsService{
		registry:           registry,
		store:              store,
		workspaceDirectory: workspaceDirectory,
	}
}

// ListDiagnostics lists the requested data.
func (s DiagnosticsService) ListDiagnostics(ctx context.Context) ([]Diagnostic, error) {
	projects, err := s.registry.ListProjects(ctx)
	if err != nil {
		return nil, err
	}

	activeProject, err := s.store.LoadActiveProject()
	if err != nil && !errors.Is(err, ErrActiveNotSelected) {
		return nil, err
	}

	diagnostics := make([]Diagnostic, 0, len(projects))
	for _, project := range projects {
		diagnostics = append(diagnostics, diagnoseRegistryEntry(project, activeProject.ProjectID, s.workspaceDirectory))
	}

	return diagnostics, nil
}

func diagnoseRegistryEntry(project Project, activeProjectID, workspaceDirectory string) Diagnostic {
	diagnostic := Diagnostic{
		Project: project,
		Active:  project.ID == activeProjectID,
	}

	workspacePath := filepath.Join(project.Path, workspaceDirectory)
	projectInfo, projectErr := os.Stat(project.Path)

	if projectErr == nil && projectInfo.IsDir() {
		diagnostic.PathExists = true
	}

	workspaceInfo, workspaceErr := os.Stat(workspacePath)
	if workspaceErr == nil && workspaceInfo.IsDir() {
		diagnostic.WorkspaceExists = true
	}

	diagnostic.Status = registryDiagnosticStatus(project, diagnostic)

	return diagnostic
}

func registryDiagnosticStatus(project Project, diagnostic Diagnostic) DiagnosticStatus {
	if !project.Enabled {
		return DiagnosticDisabled
	}

	if !diagnostic.PathExists {
		return DiagnosticMissingPath
	}

	if !diagnostic.WorkspaceExists {
		return DiagnosticNoWorkspace
	}

	return DiagnosticOK
}
