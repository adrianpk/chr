package runtime

import (
	"fmt"
	"path/filepath"
	"strings"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	"forge.adrianpk.com/chr/internal/harness"
)

// RuntimePlanOptions describes a package boundary value.
type RuntimePlanOptions struct {
	ExtraMounts           []harness.RuntimeMount
	ProjectDir            string
	Config                RuntimeConfig
	Workspace             workspacecore.WorkWorkspaceConfig
	ExecutionWorkspaceDir string
	TicketKey             string
	TicketBranch          string
	WorkspaceBaseRef      string
	WorkspaceBaseSHA      string
	RunID                 string
	OpsLayout             *opscore.Layout
	ArtifactID            string
}

// RuntimePlan describes a package boundary value.
type RuntimePlan struct {
	ProjectDir            string
	ProjectCheckoutDir    string
	ExecutionWorkspaceDir string
	WorkspaceMode         workspacecore.WorkWorkspaceMode
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

	config = FillRuntimeConfigDefaults(config)

	workspace := opts.Workspace
	if workspace == (workspacecore.WorkWorkspaceConfig{}) {
		workspace = workspacecore.DefaultWorkWorkspaceConfig()
	} else {
		workspace = workspacecore.FillWorkWorkspaceConfigDefaults(workspace)
	}

	if err := workspacecore.ValidateWorkWorkspaceConfig(workspace); err != nil {
		return RuntimePlan{}, err
	}

	executionWorkspaceDir := runtimeExecutionWorkspaceDir(projectDir, workspace.Mode, opts.ExecutionWorkspaceDir)

	workspaceDir := filepath.Join(executionWorkspaceDir, workspacecore.DirectoryName)
	dockerfilePath := projectcore.ResolvePath(projectDir, config.DockerfilePath)

	paths, err := runtimeArtifactPaths(projectDir, executionWorkspaceDir, config.ArtifactDir, opts)
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
		ArtifactDir:           paths.artifactDir,
		RunArtifactDir:        paths.runArtifactDir,
		WorkerContextPath:     paths.workerContextPath,
		ResultPath:            paths.resultPath,
		LogPath:               paths.logPath,
		ValidationPath:        paths.validationPath,
		ValidationCommand:     config.ValidationCommand,
	}, nil
}

// FillRuntimeConfigDefaults preserves the existing field defaults.
func FillRuntimeConfigDefaults(config RuntimeConfig) RuntimeConfig {
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

func runtimeExecutionWorkspaceDir(projectDir string, mode workspacecore.WorkWorkspaceMode, configuredPath string) string {
	if mode != workspacecore.WorkWorkspaceModeWorktree || strings.TrimSpace(configuredPath) == "" {
		return projectDir
	}

	return projectcore.ResolvePath(projectDir, configuredPath)
}

type runtimeArtifactPathSet struct {
	artifactDir       string
	runArtifactDir    string
	workerContextPath string
	resultPath        string
	logPath           string
	validationPath    string
}

func runtimeArtifactPaths(projectDir, executionWorkspaceDir, configuredArtifactDir string, opts RuntimePlanOptions) (runtimeArtifactPathSet, error) {
	selected, err := runtimeOpsLayout(projectDir, configuredArtifactDir, opts)
	if err != nil {
		return runtimeArtifactPathSet{}, err
	}

	if selected.canonical {
		return canonicalRuntimeArtifactPaths(selected.layout, selected.artifactID, opts)
	}

	return privateRuntimeArtifactPaths(executionWorkspaceDir, configuredArtifactDir, opts), nil
}

func canonicalRuntimeArtifactPaths(layout opscore.Layout, artifactID string, opts RuntimePlanOptions) (runtimeArtifactPathSet, error) {
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

func canonicalRunArtifactDir(layout opscore.Layout, artifactDir, artifactID string, opts RuntimePlanOptions) (string, error) {
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
	artifactDir := projectcore.ResolvePath(executionWorkspaceDir, configuredArtifactDir)
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

type runtimeArtifactLayout struct {
	layout     opscore.Layout
	artifactID string
	canonical  bool
}

func runtimeOpsLayout(projectDir, configuredArtifactDir string, opts RuntimePlanOptions) (runtimeArtifactLayout, error) {
	if opts.OpsLayout != nil || opts.ArtifactID != "" {
		if opts.OpsLayout == nil || strings.TrimSpace(opts.ArtifactID) == "" {
			return runtimeArtifactLayout{}, fmt.Errorf("canonical ops layout and artifact id are required together")
		}

		return runtimeArtifactLayout{layout: *opts.OpsLayout, artifactID: strings.TrimSpace(opts.ArtifactID), canonical: true}, nil
	}

	artifactDir := filepath.ToSlash(filepath.Clean(filepath.FromSlash(strings.TrimSpace(configuredArtifactDir))))
	if artifactDir == PrivateRuntimeArtifactDir {
		layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: projectDir})
		if err != nil {
			return runtimeArtifactLayout{}, err
		}

		return runtimeArtifactLayout{layout: layout, canonical: true}, nil
	}

	return runtimeArtifactLayout{}, nil
}

func runtimeMounts(projectDir, executionWorkspaceDir string, mode workspacecore.WorkWorkspaceMode, workdir string) []harness.RuntimeMount {
	mountSource := projectDir
	mountDescription := "project workspace"

	if mode == workspacecore.WorkWorkspaceModeWorktree {
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
