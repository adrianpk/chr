package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// DaemonProjectStatus defines a package boundary type.
type DaemonProjectStatus string

const (
	// DaemonProjectStatusReady defines a package boundary constant.
	DaemonProjectStatusReady DaemonProjectStatus = "ready"
	// DaemonProjectStatusMissing defines a package boundary constant.
	DaemonProjectStatusMissing DaemonProjectStatus = "missing"
	// DaemonProjectStatusInvalid defines a package boundary constant.
	DaemonProjectStatusInvalid DaemonProjectStatus = "invalid"
)

// DaemonProject describes a package boundary value.
type DaemonProject struct {
	Project Project
	Status  DaemonProjectStatus
	Reason  string
}

// LoadDaemonProjects loads the corresponding package data.
func LoadDaemonProjects(ctx context.Context, registry ProjectRegistry) ([]DaemonProject, error) {
	projects, err := registry.ListProjects(ctx)
	if err != nil {
		return nil, err
	}

	loaded := make([]DaemonProject, 0, len(projects))
	for _, project := range projects {
		if !project.Enabled {
			continue
		}

		loaded = append(loaded, classifyDaemonProject(project))
	}

	return loaded, nil
}

func classifyDaemonProject(project Project) DaemonProject {
	if project.Path == "" {
		return DaemonProject{
			Project: project,
			Status:  DaemonProjectStatusInvalid,
			Reason:  "project path is empty",
		}
	}

	info, err := os.Stat(project.Path)
	if err != nil {
		return classifyDaemonProjectStatError(project, err)
	}

	if !info.IsDir() {
		return DaemonProject{
			Project: project,
			Status:  DaemonProjectStatusInvalid,
			Reason:  "project path is not a directory",
		}
	}

	return classifyDaemonProjectWorkspace(project)
}

func classifyDaemonProjectStatError(project Project, err error) DaemonProject {
	if os.IsNotExist(err) {
		return DaemonProject{
			Project: project,
			Status:  DaemonProjectStatusMissing,
			Reason:  "project path does not exist",
		}
	}

	return DaemonProject{
		Project: project,
		Status:  DaemonProjectStatusInvalid,
		Reason:  fmt.Sprintf("inspect project path: %v", err),
	}
}

func classifyDaemonProjectWorkspace(project Project) DaemonProject {
	workspacePath := filepath.Join(project.Path, DirectoryName)

	info, err := os.Stat(workspacePath)
	if err != nil {
		return classifyDaemonProjectWorkspaceError(project, err)
	}

	if !info.IsDir() {
		return DaemonProject{
			Project: project,
			Status:  DaemonProjectStatusInvalid,
			Reason:  "project workspace is not a directory",
		}
	}

	return DaemonProject{
		Project: project,
		Status:  DaemonProjectStatusReady,
	}
}

func classifyDaemonProjectWorkspaceError(project Project, err error) DaemonProject {
	if os.IsNotExist(err) {
		return DaemonProject{
			Project: project,
			Status:  DaemonProjectStatusInvalid,
			Reason:  "project workspace is missing",
		}
	}

	return DaemonProject{
		Project: project,
		Status:  DaemonProjectStatusInvalid,
		Reason:  fmt.Sprintf("inspect project workspace: %v", err),
	}
}
