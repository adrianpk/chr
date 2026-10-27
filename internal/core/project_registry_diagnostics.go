package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
)

// ProjectRegistryDiagnostic describes a package boundary value.
type ProjectRegistryDiagnostic struct {
	Project         Project
	Active          bool
	PathExists      bool
	WorkspaceExists bool
	Status          ProjectRegistryDiagnosticStatus
}

// ProjectRegistryDiagnosticStatus defines a package boundary type.
type ProjectRegistryDiagnosticStatus string

const (
	// ProjectRegistryDiagnosticOK defines a package boundary constant.
	ProjectRegistryDiagnosticOK ProjectRegistryDiagnosticStatus = "ok"
	// ProjectRegistryDiagnosticDisabled defines a package boundary constant.
	ProjectRegistryDiagnosticDisabled ProjectRegistryDiagnosticStatus = "disabled"
	// ProjectRegistryDiagnosticMissingPath defines a package boundary constant.
	ProjectRegistryDiagnosticMissingPath ProjectRegistryDiagnosticStatus = "missing_path"
	// ProjectRegistryDiagnosticNoWorkspace defines a package boundary constant.
	ProjectRegistryDiagnosticNoWorkspace ProjectRegistryDiagnosticStatus = "missing_workspace"
)

// ProjectRegistryDiagnosticsService describes a package boundary value.
type ProjectRegistryDiagnosticsService struct {
	registry           ProjectRegistry
	store              ActiveProjectStore
	workspaceDirectory string
}

// NewProjectRegistryDiagnosticsService constructs the corresponding package service or value.
func NewProjectRegistryDiagnosticsService(
	registry ProjectRegistry,
	store ActiveProjectStore,
	workspaceDirectory string,
) ProjectRegistryDiagnosticsService {
	return ProjectRegistryDiagnosticsService{
		registry:           registry,
		store:              store,
		workspaceDirectory: workspaceDirectory,
	}
}

// ListDiagnostics lists the requested data.
func (s ProjectRegistryDiagnosticsService) ListDiagnostics(
	ctx context.Context,
) ([]ProjectRegistryDiagnostic, error) {
	projects, err := s.registry.ListProjects(ctx)
	if err != nil {
		return nil, err
	}

	activeProject, err := s.store.LoadActiveProject()
	if err != nil && !errors.Is(err, ErrActiveProjectNotSelected) {
		return nil, err
	}

	diagnostics := make([]ProjectRegistryDiagnostic, 0, len(projects))
	for _, project := range projects {
		diagnostics = append(diagnostics, diagnoseProjectRegistryEntry(project, activeProject.ProjectID, s.workspaceDirectory))
	}

	return diagnostics, nil
}

func diagnoseProjectRegistryEntry(project Project, activeProjectID, workspaceDirectory string) ProjectRegistryDiagnostic {
	diagnostic := ProjectRegistryDiagnostic{
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

	diagnostic.Status = projectRegistryDiagnosticStatus(project, diagnostic)

	return diagnostic
}

func projectRegistryDiagnosticStatus(
	project Project,
	diagnostic ProjectRegistryDiagnostic,
) ProjectRegistryDiagnosticStatus {
	if !project.Enabled {
		return ProjectRegistryDiagnosticDisabled
	}

	if !diagnostic.PathExists {
		return ProjectRegistryDiagnosticMissingPath
	}

	if !diagnostic.WorkspaceExists {
		return ProjectRegistryDiagnosticNoWorkspace
	}

	return ProjectRegistryDiagnosticOK
}
