package chc

import (
	"fmt"
	"path/filepath"
	"strings"
)

type RuntimePlanOptions struct {
	ProjectDir string
	Config     RuntimeConfig
	TicketKey  string
	RunID      string
}

type RuntimePlan struct {
	ProjectDir        string
	WorkspaceDir      string
	DockerfilePath    string
	Image             string
	ContainerWorkdir  string
	Mounts            []RuntimeMount
	ArtifactDir       string
	RunArtifactDir    string
	WorkerContextPath string
	ResultPath        string
	LogPath           string
	ValidationPath    string
	ValidationCommand string
}

type RuntimeMount struct {
	Source      string
	Target      string
	ReadOnly    bool
	Description string
}

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
		ProjectDir:        projectDir,
		WorkspaceDir:      workspaceDir,
		DockerfilePath:    dockerfilePath,
		Image:             config.Image,
		ContainerWorkdir:  config.Workdir,
		Mounts:            runtimeMounts(projectDir, config.Workdir),
		ArtifactDir:       artifactDir,
		RunArtifactDir:    runArtifactDir,
		WorkerContextPath: filepath.Join(runArtifactDir, "context.md"),
		ResultPath:        filepath.Join(runArtifactDir, "result.json"),
		LogPath:           filepath.Join(runArtifactDir, "runtime.log"),
		ValidationPath:    filepath.Join(runArtifactDir, "validation.log"),
		ValidationCommand: config.ValidationCommand,
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

func resolveProjectPath(projectDir, configuredPath string) string {
	path := filepath.Clean(filepath.FromSlash(configuredPath))
	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(projectDir, path)
}

func runtimeMounts(projectDir, workdir string) []RuntimeMount {
	return []RuntimeMount{
		{
			Source:      projectDir,
			Target:      cleanContainerPath(workdir),
			ReadOnly:    false,
			Description: "project workspace",
		},
	}
}

func cleanContainerPath(path string) string {
	cleaned := filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
	if strings.HasPrefix(cleaned, "/") {
		return cleaned
	}

	return fmt.Sprintf("/%s", cleaned)
}
