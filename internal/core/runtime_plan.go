package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RuntimePlanOptions describes a package boundary value.
type RuntimePlanOptions struct {
	ProjectDir            string
	Config                RuntimeConfig
	Workspace             WorkWorkspaceConfig
	ExecutionWorkspaceDir string
	TicketKey             string
	RunID                 string
}

// RuntimePlan describes a package boundary value.
type RuntimePlan struct {
	ProjectDir            string
	ProjectCheckoutDir    string
	ExecutionWorkspaceDir string
	WorkspaceMode         WorkWorkspaceMode
	WorkspaceDir          string
	DockerfilePath        string
	Image                 string
	ContainerWorkdir      string
	Mounts                []RuntimeMount
	ArtifactDir           string
	RunArtifactDir        string
	WorkerContextPath     string
	ResultPath            string
	LogPath               string
	ValidationPath        string
	ValidationCommand     string
}

// RuntimeMount describes a package boundary value.
type RuntimeMount struct {
	Source      string
	Target      string
	ReadOnly    bool
	Description string
}

// PlanRuntime executes the corresponding package operation.
func PlanRuntime(opts RuntimePlanOptions) (RuntimePlan, error) {
	projectDir, err := resolveProjectDir(opts.ProjectDir)
	if err != nil {
		return RuntimePlan{}, err
	}

	config := opts.Config
	if config == (RuntimeConfig{}) {
		config = DefaultRuntimeConfig()
	}

	config = fillRuntimeConfigDefaults(config)

	workspace := opts.Workspace
	if workspace == (WorkWorkspaceConfig{}) {
		workspace = DefaultWorkWorkspaceConfig()
	} else {
		workspace = fillWorkWorkspaceConfigDefaults(workspace)
	}

	if err := ValidateWorkWorkspaceConfig(workspace); err != nil {
		return RuntimePlan{}, err
	}

	executionWorkspaceDir := opts.ExecutionWorkspaceDir
	if executionWorkspaceDir == "" {
		executionWorkspaceDir = projectDir
	} else {
		executionWorkspaceDir = resolveProjectPath(projectDir, executionWorkspaceDir)
	}

	workspaceDir := filepath.Join(projectDir, DirectoryName)
	dockerfilePath := resolveProjectPath(projectDir, config.DockerfilePath)
	artifactDir := resolveProjectPath(projectDir, config.ArtifactDir)
	runArtifactDir := artifactDir

	if opts.TicketKey != "" && opts.RunID != "" {
		runArtifactDir = filepath.Join(artifactDir, opts.TicketKey, opts.RunID)
	} else if opts.RunID != "" {
		runArtifactDir = filepath.Join(artifactDir, opts.RunID)
	}

	return RuntimePlan{
		ProjectDir:            projectDir,
		ProjectCheckoutDir:    projectDir,
		ExecutionWorkspaceDir: executionWorkspaceDir,
		WorkspaceMode:         workspace.Mode,
		WorkspaceDir:          workspaceDir,
		DockerfilePath:        dockerfilePath,
		Image:                 config.Image,
		ContainerWorkdir:      config.Workdir,
		Mounts:                runtimeMounts(projectDir, config.Workdir),
		ArtifactDir:           artifactDir,
		RunArtifactDir:        runArtifactDir,
		WorkerContextPath:     filepath.Join(runArtifactDir, "context.md"),
		ResultPath:            filepath.Join(runArtifactDir, "result.json"),
		LogPath:               filepath.Join(runArtifactDir, "runtime.log"),
		ValidationPath:        filepath.Join(runArtifactDir, "validation.log"),
		ValidationCommand:     config.ValidationCommand,
	}, nil
}

func fillRuntimeConfigDefaults(config RuntimeConfig) RuntimeConfig {
	defaults := DefaultRuntimeConfig()

	if config.Image == "" {
		config.Image = defaults.Image
	}

	if config.DockerfilePath == "" {
		config.DockerfilePath = defaults.DockerfilePath
	}

	if config.Workdir == "" {
		config.Workdir = defaults.Workdir
	}

	if config.ArtifactDir == "" {
		config.ArtifactDir = defaults.ArtifactDir
	}

	if config.ValidationCommand == "" {
		config.ValidationCommand = defaults.ValidationCommand
	}

	return config
}

func fillWorkWorkspaceConfigDefaults(config WorkWorkspaceConfig) WorkWorkspaceConfig {
	defaults := DefaultWorkWorkspaceConfig()

	if config.Mode == "" {
		config.Mode = defaults.Mode
	}

	if config.WorktreeRoot == "" {
		config.WorktreeRoot = defaults.WorktreeRoot
	}

	if config.WorktreeBase == "" {
		config.WorktreeBase = defaults.WorktreeBase
	}

	if config.WorktreeCleanup == "" {
		config.WorktreeCleanup = defaults.WorktreeCleanup
	}

	return config
}

func resolveProjectPath(projectDir, configuredPath string) string {
	path := filepath.Clean(filepath.FromSlash(configuredPath))
	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(projectDir, path)
}

func runtimeMounts(projectDir, workdir string) []RuntimeMount {
	mounts := []RuntimeMount{
		{
			Source:      projectDir,
			Target:      cleanContainerPath(workdir),
			ReadOnly:    false,
			Description: "project workspace",
		},
	}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		mounts = append(mounts, codexCredentialMounts(homeDir)...)
	}

	return mounts
}

func codexCredentialMounts(homeDir string) []RuntimeMount {
	candidates := []struct {
		source string
		target string
	}{
		{
			source: filepath.Join(homeDir, ".codex", "auth.json"),
			target: "/root/.codex/auth.json",
		},
		{
			source: filepath.Join(homeDir, ".codex", "config.toml"),
			target: "/root/.codex/config.toml",
		},
	}

	mounts := make([]RuntimeMount, 0, len(candidates))
	for _, candidate := range candidates {
		_, err := os.Stat(candidate.source)
		if err != nil {
			continue
		}

		mounts = append(mounts, RuntimeMount{
			Source:      candidate.source,
			Target:      candidate.target,
			ReadOnly:    true,
			Description: "codex credential",
		})
	}

	return mounts
}

func cleanContainerPath(path string) string {
	cleaned := filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
	if strings.HasPrefix(cleaned, "/") {
		return cleaned
	}

	return fmt.Sprintf("/%s", cleaned)
}
