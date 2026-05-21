package chc

import (
	"context"
	"errors"
	"os"
	"path/filepath"
)

type ProjectRegistryDiagnostic struct {
	Project         Project
	Active          bool
	PathExists      bool
	WorkspaceExists bool
	Status          ProjectRegistryDiagnosticStatus
}

type ProjectRegistryDiagnosticStatus string

const (
	ProjectRegistryDiagnosticOK          ProjectRegistryDiagnosticStatus = "ok"
	ProjectRegistryDiagnosticDisabled    ProjectRegistryDiagnosticStatus = "disabled"
	ProjectRegistryDiagnosticMissingPath ProjectRegistryDiagnosticStatus = "missing_path"
	ProjectRegistryDiagnosticNoWorkspace ProjectRegistryDiagnosticStatus = "missing_workspace"
)

type ProjectRegistryDiagnosticsService struct {
	registry ProjectRegistry
	store    ActiveProjectStore
}

func NewProjectRegistryDiagnosticsService(
	registry ProjectRegistry,
	store ActiveProjectStore,
) ProjectRegistryDiagnosticsService {
	return ProjectRegistryDiagnosticsService{
		registry: registry,
		store:    store,
	}
}

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
		diagnostics = append(diagnostics, diagnoseProjectRegistryEntry(project, activeProject.ProjectID))
	}

	return diagnostics, nil
}

func diagnoseProjectRegistryEntry(project Project, activeProjectID string) ProjectRegistryDiagnostic {
	diagnostic := ProjectRegistryDiagnostic{
		Project: project,
		Active:  project.ID == activeProjectID,
	}

	workspacePath := filepath.Join(project.Path, DirectoryName)
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
