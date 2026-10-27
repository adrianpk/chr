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

// PlanOptions describes a package boundary value.
type PlanOptions struct {
	ExtraMounts           []harness.RuntimeMount
	ProjectDir            string
	Config                Config
	Workspace             workspacecore.ExecutionConfig
	ExecutionWorkspaceDir string
	TicketKey             string
	TicketBranch          string
	WorkspaceBaseRef      string
	WorkspaceBaseSHA      string
	RunID                 string
	OpsLayout             *opscore.Layout
	ArtifactID            string
}

// Plan describes a package boundary value.
type Plan struct {
	ProjectDir            string
	ProjectCheckoutDir    string
	ExecutionWorkspaceDir string
	WorkspaceMode         workspacecore.ExecutionMode
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

// ResolvePlan executes the corresponding package operation.
func ResolvePlan(opts PlanOptions) (Plan, error) {
	projectDir, err := projectcore.ResolveDirectory(opts.ProjectDir)
	if err != nil {
		return Plan{}, err
	}

	config := opts.Config
	if config == (Config{}) {
		config = DefaultConfig()
	}

	config = FillConfigDefaults(config)

	workspace := opts.Workspace
	if workspace == (workspacecore.ExecutionConfig{}) {
		workspace = workspacecore.DefaultExecutionConfig()
	} else {
		workspace = workspacecore.FillExecutionDefaults(workspace)
	}

	if err := workspacecore.ValidateExecutionConfig(workspace); err != nil {
		return Plan{}, err
	}

	executionWorkspaceDir := executionWorkspaceDir(projectDir, workspace.Mode, opts.ExecutionWorkspaceDir)

	workspaceDir := filepath.Join(executionWorkspaceDir, workspacecore.DirectoryName)
	dockerfilePath := projectcore.ResolvePath(projectDir, config.DockerfilePath)

	paths, err := artifactPaths(projectDir, executionWorkspaceDir, config.ArtifactDir, opts)
	if err != nil {
		return Plan{}, err
	}

	return Plan{
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
		Mounts:                append(mounts(projectDir, executionWorkspaceDir, workspace.Mode, config.Workdir), opts.ExtraMounts...),
		ArtifactDir:           paths.artifactDir,
		RunArtifactDir:        paths.runArtifactDir,
		WorkerContextPath:     paths.workerContextPath,
		ResultPath:            paths.resultPath,
		LogPath:               paths.logPath,
		ValidationPath:        paths.validationPath,
		ValidationCommand:     config.ValidationCommand,
	}, nil
}

// FillConfigDefaults preserves the existing field defaults.
func FillConfigDefaults(config Config) Config {
	defaults := DefaultConfig()

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

func executionWorkspaceDir(projectDir string, mode workspacecore.ExecutionMode, configuredPath string) string {
	if mode != workspacecore.ExecutionWorktree || strings.TrimSpace(configuredPath) == "" {
		return projectDir
	}

	return projectcore.ResolvePath(projectDir, configuredPath)
}

type artifactPathSet struct {
	artifactDir       string
	runArtifactDir    string
	workerContextPath string
	resultPath        string
	logPath           string
	validationPath    string
}

func artifactPaths(projectDir, executionWorkspaceDir, configuredArtifactDir string, opts PlanOptions) (artifactPathSet, error) {
	selected, err := opsLayout(projectDir, configuredArtifactDir, opts)
	if err != nil {
		return artifactPathSet{}, err
	}

	if selected.canonical {
		return canonicalArtifactPaths(selected.layout, selected.artifactID, opts)
	}

	return privateArtifactPaths(executionWorkspaceDir, configuredArtifactDir, opts), nil
}

func canonicalArtifactPaths(layout opscore.Layout, artifactID string, opts PlanOptions) (artifactPathSet, error) {
	artifactDir, err := layout.ArtifactRoot("run")
	if err != nil {
		return artifactPathSet{}, err
	}

	runArtifactDir, err := canonicalRunArtifactDir(layout, artifactDir, artifactID, opts)
	if err != nil {
		return artifactPathSet{}, err
	}

	return artifactPathSet{
		artifactDir:       artifactDir,
		runArtifactDir:    runArtifactDir,
		workerContextPath: filepath.Join(runArtifactDir, "files", "context.md"),
		resultPath:        filepath.Join(runArtifactDir, "evidence", "result.json"),
		logPath:           filepath.Join(runArtifactDir, "evidence", "runtime.log"),
		validationPath:    filepath.Join(runArtifactDir, "evidence", "validation.log"),
	}, nil
}

func canonicalRunArtifactDir(layout opscore.Layout, artifactDir, artifactID string, opts PlanOptions) (string, error) {
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

func privateArtifactPaths(executionWorkspaceDir, configuredArtifactDir string, opts PlanOptions) artifactPathSet {
	artifactDir := projectcore.ResolvePath(executionWorkspaceDir, configuredArtifactDir)
	runArtifactDir := privateRunArtifactDir(artifactDir, opts)

	return artifactPathSet{
		artifactDir:       artifactDir,
		runArtifactDir:    runArtifactDir,
		workerContextPath: filepath.Join(runArtifactDir, "context.md"),
		resultPath:        filepath.Join(runArtifactDir, "result.json"),
		logPath:           filepath.Join(runArtifactDir, "runtime.log"),
		validationPath:    filepath.Join(runArtifactDir, "validation.log"),
	}
}

func privateRunArtifactDir(artifactDir string, opts PlanOptions) string {
	if opts.TicketKey != "" && opts.RunID != "" {
		return filepath.Join(artifactDir, opts.TicketKey, opts.RunID)
	}

	if opts.RunID != "" {
		return filepath.Join(artifactDir, opts.RunID)
	}

	return artifactDir
}

type artifactLayout struct {
	layout     opscore.Layout
	artifactID string
	canonical  bool
}

func opsLayout(projectDir, configuredArtifactDir string, opts PlanOptions) (artifactLayout, error) {
	if opts.OpsLayout != nil || opts.ArtifactID != "" {
		if opts.OpsLayout == nil || strings.TrimSpace(opts.ArtifactID) == "" {
			return artifactLayout{}, fmt.Errorf("canonical ops layout and artifact id are required together")
		}

		return artifactLayout{layout: *opts.OpsLayout, artifactID: strings.TrimSpace(opts.ArtifactID), canonical: true}, nil
	}

	artifactDir := filepath.ToSlash(filepath.Clean(filepath.FromSlash(strings.TrimSpace(configuredArtifactDir))))
	if artifactDir == PrivateArtifactDir {
		layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: projectDir})
		if err != nil {
			return artifactLayout{}, err
		}

		return artifactLayout{layout: layout, canonical: true}, nil
	}

	return artifactLayout{}, nil
}

func mounts(projectDir, executionWorkspaceDir string, mode workspacecore.ExecutionMode, workdir string) []harness.RuntimeMount {
	mountSource := projectDir
	mountDescription := "project workspace"

	if mode == workspacecore.ExecutionWorktree {
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
