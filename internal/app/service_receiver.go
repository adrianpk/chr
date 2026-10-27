package app

import (
	"os"
	"os/exec"
	"strings"

	harnessruntime "forge.adrianpk.com/chr/internal/core/harness/runtime"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// ListRegisteredProjects is a bounded, read-only registry operation for explicit IDs.
func (s service) ListRegisteredProjects(ids []string) ([]projectcore.Project, error) {
	return projectcore.ReadRegistered(s.ctx, s.env, ids)
}

// InspectTelegramReceiver exposes transport health to local and remote readers.
func (s service) InspectReceiver(botID int64) (telegramcore.ReceiverState, error) {
	return telegramcore.InspectReceiver(s.ctx, s.env, botID)
}

type projectReadiness struct {
	RuntimeReady           bool
	ModelCredentialPresent bool
	ModelExecutablePresent bool
}

func (s service) InspectProjectReadiness(project projectcore.Project) projectReadiness {
	result := projectReadiness{ModelCredentialPresent: strings.TrimSpace(os.Getenv("OPENAI_API_KEY")) != ""}

	command := os.Getenv("CHR_CODEX_COMMAND")
	if command == "" {
		command = "codex"
	}

	_, commandErr := exec.LookPath(command)
	result.ModelExecutablePresent = commandErr == nil

	config, err := workspaceconfig.Load(project.Path)
	if err != nil || !project.Enabled {
		return result
	}

	runtime, err := harnessruntime.Validate(harnessruntime.ValidationOptions{ExtraMounts: legacyCodexMounts(), ProjectDir: project.Path, Config: config.Runtime, Workspace: config.Work.Workspace})
	result.RuntimeReady = err == nil && harnessruntime.ValidationPassed(runtime.Checks)

	return result
}
