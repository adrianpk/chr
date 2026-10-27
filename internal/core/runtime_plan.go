package core

import (
	"fmt"
	"path/filepath"
	"strings"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	"forge.adrianpk.com/chr/internal/harness"
)

// RuntimePlanOptions describes a package boundary value.
type RuntimePlanOptions struct {
	ExtraMounts           []harness.RuntimeMount
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
	Mounts                []harness.RuntimeMount
	ArtifactDir           string
	RunArtifactDir        string
	WorkerContextPath     string
	ResultPath            string
	LogPath               string
	ValidationPath        string
	ValidationCommand     string
}

// PlanRuntime executes the corresponding package operation.
func PlanRuntime(opts RuntimePlanOptions) (RuntimePlan, error) {
	projectDir, err := projectcore.ResolveDirectory(opts.ProjectDir)
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

	artifactDir, runArtifactDir, workerContextPath, resultPath, logPath, validationPath, err := runtimeArtifactPaths(projectDir, executionWorkspaceDir, config.ArtifactDir, opts)
	if err != nil {
		return RuntimePlan{}, err
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
		Mounts:                append(runtimeMounts(projectDir, executionWorkspaceDir, workspace.Mode, config.Workdir), opts.ExtraMounts...),
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

func runtimeExecutionWorkspaceDir(projectDir string, mode WorkWorkspaceMode, configuredPath string) string {
	if mode != WorkWorkspaceModeWorktree || strings.TrimSpace(configuredPath) == "" {
		return projectDir
	}

	return resolveProjectPath(projectDir, configuredPath)
}

type runtimeArtifactPathSet struct {
	artifactDir       string
	runArtifactDir    string
	workerContextPath string
	resultPath        string
	logPath           string
	validationPath    string
}

func runtimeArtifactPaths(
	projectDir string,
	executionWorkspaceDir string,
	configuredArtifactDir string,
	opts RuntimePlanOptions,
) (string, string, string, string, string, string, error) {
	paths, err := runtimeArtifactPathSetFor(projectDir, executionWorkspaceDir, configuredArtifactDir, opts)
	if err != nil {
		return "", "", "", "", "", "", err
	}

	return paths.artifactDir,
		paths.runArtifactDir,
		paths.workerContextPath,
		paths.resultPath,
		paths.logPath,
		paths.validationPath,
		nil
}

func runtimeArtifactPathSetFor(projectDir, executionWorkspaceDir, configuredArtifactDir string, opts RuntimePlanOptions) (runtimeArtifactPathSet, error) {
	layout, artifactID, canonical, err := runtimeOpsLayout(projectDir, configuredArtifactDir, opts)
	if err != nil {
		return runtimeArtifactPathSet{}, err
	}

	if canonical {
		return canonicalRuntimeArtifactPaths(layout, artifactID, opts)
	}

	return privateRuntimeArtifactPaths(executionWorkspaceDir, configuredArtifactDir, opts), nil
}

func canonicalRuntimeArtifactPaths(layout OpsLayout, artifactID string, opts RuntimePlanOptions) (runtimeArtifactPathSet, error) {
	artifactDir, err := layout.ArtifactRoot("run")
	if err != nil {
		return runtimeArtifactPathSet{}, err
	}

	runArtifactDir, err := canonicalRunArtifactDir(layout, artifactDir, artifactID, opts)
	if err != nil {
		return runtimeArtifactPathSet{}, err
	}

	return runtimeArtifactPathSet{
		artifactDir:       artifactDir,
		runArtifactDir:    runArtifactDir,
		workerContextPath: filepath.Join(runArtifactDir, "files", "context.md"),
		resultPath:        filepath.Join(runArtifactDir, "evidence", "result.json"),
		logPath:           filepath.Join(runArtifactDir, "evidence", "runtime.log"),
		validationPath:    filepath.Join(runArtifactDir, "evidence", "validation.log"),
	}, nil
}

func canonicalRunArtifactDir(layout OpsLayout, artifactDir, artifactID string, opts RuntimePlanOptions) (string, error) {
	switch {
	case artifactID != "":
		return layout.ArtifactDirectory("run", artifactID)
	case opts.TicketKey != "" && opts.RunID != "":
		return layout.ArtifactDirectory("run", opts.TicketKey+"-"+opts.RunID)
	case opts.RunID != "":
		return layout.ArtifactDirectory("run", opts.RunID)
	default:
		return artifactDir, nil
	}
}

func privateRuntimeArtifactPaths(executionWorkspaceDir, configuredArtifactDir string, opts RuntimePlanOptions) runtimeArtifactPathSet {
	artifactDir := resolveProjectPath(executionWorkspaceDir, configuredArtifactDir)
	runArtifactDir := privateRunArtifactDir(artifactDir, opts)

	return runtimeArtifactPathSet{
		artifactDir:       artifactDir,
		runArtifactDir:    runArtifactDir,
		workerContextPath: filepath.Join(runArtifactDir, "context.md"),
		resultPath:        filepath.Join(runArtifactDir, "result.json"),
		logPath:           filepath.Join(runArtifactDir, "runtime.log"),
		validationPath:    filepath.Join(runArtifactDir, "validation.log"),
	}
}

func privateRunArtifactDir(artifactDir string, opts RuntimePlanOptions) string {
	if opts.TicketKey != "" && opts.RunID != "" {
		return filepath.Join(artifactDir, opts.TicketKey, opts.RunID)
	}

	if opts.RunID != "" {
		return filepath.Join(artifactDir, opts.RunID)
	}

	return artifactDir
}

func runtimeOpsLayout(projectDir, configuredArtifactDir string, opts RuntimePlanOptions) (OpsLayout, string, bool, error) {
	if opts.OpsLayout != nil || opts.ArtifactID != "" {
		if opts.OpsLayout == nil || strings.TrimSpace(opts.ArtifactID) == "" {
			return OpsLayout{}, "", false, fmt.Errorf("canonical ops layout and artifact id are required together")
		}

		return *opts.OpsLayout, strings.TrimSpace(opts.ArtifactID), true, nil
	}

	artifactDir := filepath.ToSlash(filepath.Clean(filepath.FromSlash(strings.TrimSpace(configuredArtifactDir))))
	if artifactDir == PrivateRuntimeArtifactDir {
		layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: projectDir})
		if err != nil {
			return OpsLayout{}, "", false, err
		}

		return layout, "", true, nil
	}

	return OpsLayout{}, "", false, nil
}

func runtimeMounts(projectDir, executionWorkspaceDir string, mode WorkWorkspaceMode, workdir string) []harness.RuntimeMount {
	mountSource := projectDir
	mountDescription := "project workspace"

	if mode == WorkWorkspaceModeWorktree {
		mountSource = executionWorkspaceDir
		mountDescription = "execution workspace"
	}

	mounts := []harness.RuntimeMount{
		{
			Source:      mountSource,
			Target:      cleanContainerPath(workdir),
			ReadOnly:    false,
			Description: mountDescription,
		},
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
