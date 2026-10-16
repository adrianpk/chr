package app

import (
	"os"
	"os/exec"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
)

// ListRegisteredProjects is a bounded, read-only registry operation for explicit IDs.
func (s service) ListRegisteredProjects(ids []string) ([]core.Project, error) {
	return core.ReadRegisteredProjects(s.ctx, s.env, ids)
}

// InspectTelegramReceiver exposes transport health to local and remote readers.
func (s service) InspectTelegramReceiver(botID int64) (core.TelegramReceiverState, error) {
	return core.InspectTelegramReceiver(s.ctx, s.env, botID)
}

type projectReadiness struct {
	RuntimeReady           bool
	ModelCredentialPresent bool
	ModelExecutablePresent bool
}

func (s service) InspectProjectReadiness(project core.Project) projectReadiness {
	result := projectReadiness{ModelCredentialPresent: strings.TrimSpace(os.Getenv("OPENAI_API_KEY")) != ""}

	command := os.Getenv("CHR_CODEX_COMMAND")
	if command == "" {
		command = "codex"
	}

	_, commandErr := exec.LookPath(command)
	result.ModelExecutablePresent = commandErr == nil

	config, err := core.LoadWorkspaceConfig(project.Path)
	if err != nil || !project.Enabled {
		return result
	}

	runtime, err := core.ValidateRuntime(core.RuntimeValidationOptions{ProjectDir: project.Path, Config: config.Runtime, Workspace: config.Work.Workspace})
	result.RuntimeReady = err == nil && core.RuntimeValidationPassed(runtime.Checks)

	return result
}
