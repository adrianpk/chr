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
	TicketBranch          string
	WorkspaceBaseRef      string
	WorkspaceBaseSHA      string
	RunID                 string
	OpsLayout             *OpsLayout
	ArtifactID            string
}

// RuntimePlan describes a package boundary value.
type RuntimePlan struct {
	ProjectDir            string
	ProjectCheckoutDir    string
	ExecutionWorkspaceDir string
	WorkspaceMode         WorkWorkspaceMode
	TicketKey             string
	TicketBranch          string
	WorkspaceBaseRef      string
	WorkspaceBaseSHA      string
	ArtifactID            string
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

	executionWorkspaceDir := runtimeExecutionWorkspaceDir(projectDir, workspace.Mode, opts.ExecutionWorkspaceDir)

	workspaceDir := filepath.Join(executionWorkspaceDir, DirectoryName)
	dockerfilePath := resolveProjectPath(projectDir, config.DockerfilePath)
	artifactDir := resolveProjectPath(executionWorkspaceDir, config.ArtifactDir)
	runArtifactDir := artifactDir
	workerContextPath := filepath.Join(runArtifactDir, "context.md")
	resultPath := filepath.Join(runArtifactDir, "result.json")
	logPath := filepath.Join(runArtifactDir, "runtime.log")
	validationPath := filepath.Join(runArtifactDir, "validation.log")

	if opts.OpsLayout != nil || opts.ArtifactID != "" {
		if opts.OpsLayout == nil || strings.TrimSpace(opts.ArtifactID) == "" {
			return RuntimePlan{}, fmt.Errorf("canonical ops layout and artifact id are required together")
		}
		artifactDir, err = opts.OpsLayout.ArtifactRoot("run")
		if err != nil {
			return RuntimePlan{}, err
		}
		runArtifactDir, err = opts.OpsLayout.ArtifactDirectory("run", opts.ArtifactID)
		if err != nil {
			return RuntimePlan{}, err
		}
		workerContextPath = filepath.Join(runArtifactDir, "files", "context.md")
		resultPath = filepath.Join(runArtifactDir, "evidence", "result.json")
		logPath = filepath.Join(runArtifactDir, "evidence", "runtime.log")
		validationPath = filepath.Join(runArtifactDir, "evidence", "validation.log")
	}

	if opts.OpsLayout == nil && opts.TicketKey != "" && opts.RunID != "" {
		runArtifactDir = filepath.Join(artifactDir, opts.TicketKey, opts.RunID)
		workerContextPath = filepath.Join(runArtifactDir, "context.md")
		resultPath = filepath.Join(runArtifactDir, "result.json")
		logPath = filepath.Join(runArtifactDir, "runtime.log")
		validationPath = filepath.Join(runArtifactDir, "validation.log")
	} else if opts.OpsLayout == nil && opts.RunID != "" {
		runArtifactDir = filepath.Join(artifactDir, opts.RunID)
		workerContextPath = filepath.Join(runArtifactDir, "context.md")
		resultPath = filepath.Join(runArtifactDir, "result.json")
		logPath = filepath.Join(runArtifactDir, "runtime.log")
		validationPath = filepath.Join(runArtifactDir, "validation.log")
	}

	return RuntimePlan{
		ProjectDir:            projectDir,
		ProjectCheckoutDir:    projectDir,
		ExecutionWorkspaceDir: executionWorkspaceDir,
		WorkspaceMode:         workspace.Mode,
		TicketKey:             strings.TrimSpace(opts.TicketKey),
		TicketBranch:          strings.TrimSpace(opts.TicketBranch),
		WorkspaceBaseRef:      strings.TrimSpace(opts.WorkspaceBaseRef),
		WorkspaceBaseSHA:      strings.TrimSpace(opts.WorkspaceBaseSHA),
		ArtifactID:            strings.TrimSpace(opts.ArtifactID),
		WorkspaceDir:          workspaceDir,
		DockerfilePath:        dockerfilePath,
		Image:                 config.Image,
		ContainerWorkdir:      config.Workdir,
		Mounts:                runtimeMounts(projectDir, executionWorkspaceDir, workspace.Mode, config.Workdir),
		ArtifactDir:           artifactDir,
		RunArtifactDir:        runArtifactDir,
		WorkerContextPath:     workerContextPath,
		ResultPath:            resultPath,
		LogPath:               logPath,
		ValidationPath:        validationPath,
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

func runtimeExecutionWorkspaceDir(projectDir string, mode WorkWorkspaceMode, configuredPath string) string {
	if mode != WorkWorkspaceModeWorktree || strings.TrimSpace(configuredPath) == "" {
		return projectDir
	}

	return resolveProjectPath(projectDir, configuredPath)
}

func runtimeMounts(projectDir, executionWorkspaceDir string, mode WorkWorkspaceMode, workdir string) []RuntimeMount {
	mountSource := projectDir
	mountDescription := "project workspace"

	if mode == WorkWorkspaceModeWorktree {
		mountSource = executionWorkspaceDir
		mountDescription = "execution workspace"
	}

	mounts := []RuntimeMount{
		{
			Source:      mountSource,
			Target:      cleanContainerPath(workdir),
			ReadOnly:    false,
			Description: mountDescription,
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
