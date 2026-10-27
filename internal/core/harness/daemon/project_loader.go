package daemon

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// ProjectStatus defines a package boundary type.
type ProjectStatus string

const (
	// ProjectReady defines a package boundary constant.
	ProjectReady ProjectStatus = "ready"
	// ProjectMissing defines a package boundary constant.
	ProjectMissing ProjectStatus = "missing"
	// ProjectInvalid defines a package boundary constant.
	ProjectInvalid ProjectStatus = "invalid"
)

// Project describes a package boundary value.
type Project struct {
	Project projectcore.Project
	Status  ProjectStatus
	Reason  string
}

// LoadProjects loads the corresponding package data.
func LoadProjects(ctx context.Context, registry projectcore.Registry) ([]Project, error) {
	projects, err := registry.ListProjects(ctx)
	if err != nil {
		return nil, err
	}

	loaded := make([]Project, 0, len(projects))
	for _, project := range projects {
		if !project.Enabled {
			continue
		}

		loaded = append(loaded, classifyProject(project))
	}

	return loaded, nil
}

func classifyProject(project projectcore.Project) Project {
	if project.Path == "" {
		return Project{
			Project: project,
			Status:  ProjectInvalid,
			Reason:  "project path is empty",
		}
	}

	info, err := os.Stat(project.Path)
	if err != nil {
		return classifyProjectStatError(project, err)
	}

	if !info.IsDir() {
		return Project{
			Project: project,
			Status:  ProjectInvalid,
			Reason:  "project path is not a directory",
		}
	}

	return classifyProjectWorkspace(project)
}

func classifyProjectStatError(project projectcore.Project, err error) Project {
	if os.IsNotExist(err) {
		return Project{
			Project: project,
			Status:  ProjectMissing,
			Reason:  "project path does not exist",
		}
	}

	return Project{
		Project: project,
		Status:  ProjectInvalid,
		Reason:  fmt.Sprintf("inspect project path: %v", err),
	}
}

func classifyProjectWorkspace(project projectcore.Project) Project {
	workspacePath := filepath.Join(project.Path, workspacecore.DirectoryName)

	info, err := os.Stat(workspacePath)
	if err != nil {
		return classifyProjectWorkspaceError(project, err)
	}

	if !info.IsDir() {
		return Project{
			Project: project,
			Status:  ProjectInvalid,
			Reason:  "project workspace is not a directory",
		}
	}

	return Project{
		Project: project,
		Status:  ProjectReady,
	}
}

func classifyProjectWorkspaceError(project projectcore.Project, err error) Project {
	if os.IsNotExist(err) {
		return Project{
			Project: project,
			Status:  ProjectInvalid,
			Reason:  "project workspace is missing",
		}
	}

	return Project{
		Project: project,
		Status:  ProjectInvalid,
		Reason:  fmt.Sprintf("inspect project workspace: %v", err),
	}
}
