package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/harness/codex"
)

type resolvedHarnessSelection struct {
	View       core.HarnessSelection
	Profile    core.HarnessProfile
	Executable string
}

// ResolveHarnessSelection observes one coherent persisted role selection.
// Profile setup is explicit; inspection never installs or authenticates a driver.
func (s service) ResolveHarnessSelection(ctx context.Context, role core.HarnessRole, action string) (core.HarnessSelection, error) {
	config, err := core.LoadWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		return core.HarnessSelection{}, err
	}

	resolved, err := s.resolveHarnessSelection(ctx, config, role, action)

	return resolved.View, err
}

func (s service) resolveHarnessSelection(ctx context.Context, config core.WorkspaceConfig, role core.HarnessRole, action string) (resolvedHarnessSelection, error) {
	if err := ctx.Err(); err != nil {
		return resolvedHarnessSelection{}, err
	}

	name, err := config.Harnesses.Profile(role)
	if err != nil {
		return resolvedHarnessSelection{}, err
	}

	if name == "" {
		return legacyHarnessSelection(config, role, action)
	}

	if s.env == nil {
		return resolvedHarnessSelection{}, errors.New("private user configuration is required to resolve harness profiles")
	}

	profiles, err := core.LoadHarnessProfiles(s.env.ConfigDir())
	if err != nil {
		return resolvedHarnessSelection{}, err
	}

	return resolveConfiguredHarnessSelection(config, profiles, name, role, action, s.env.ConfigDir())
}

func resolveConfiguredHarnessSelection(config core.WorkspaceConfig, profiles map[string]core.HarnessProfile, name string, role core.HarnessRole, action, configDir string) (resolvedHarnessSelection, error) {
	if name == "" {
		return legacyHarnessSelection(config, role, action)
	}

	profile, found := profiles[name]
	if !found {
		view := core.HarnessSelection{Role: role, Action: action, Profile: name, Source: "project", SelectionRevision: config.Revision()}

		return resolvedHarnessSelection{View: view}, fmt.Errorf("unknown harness profile %q; configure it in private harnesses.yml", name)
	}

	resolved := prepareHarnessSelection(config, profile, role, action, configDir)
	if !resolved.View.Readiness.Ready() {
		return resolved, fmt.Errorf("harness profile %q is not locally ready for %s", name, role)
	}

	return resolved, nil
}

func prepareHarnessSelection(config core.WorkspaceConfig, profile core.HarnessProfile, role core.HarnessRole, action, configDir string) resolvedHarnessSelection {
	model := profile.Model
	if model == "" {
		model = "cli_default"
	}

	view := core.HarnessSelection{
		Role: role, Action: action, Source: "project", Profile: profile.Name,
		Harness: profile.Harness, Backend: profile.Harness, Model: model, ModelProvider: profile.ModelProvider,
		Thinking: profile.Thinking, SelectionRevision: config.Revision(), ProfileRevision: profile.Revision,
	}

	view.Readiness = harnessProfileReadiness(profile, role, configDir)
	if profile.Harness == codex.Identity {
		view.Capabilities = (codex.Runner{}).Capabilities()
	}

	executable, err := exec.LookPath(profile.Command)
	if err != nil {
		view.Readiness.Executable = "unavailable"
	}

	return resolvedHarnessSelection{View: view, Profile: profile, Executable: executable}
}

func harnessProfileReadiness(profile core.HarnessProfile, role core.HarnessRole, configDir string) core.HarnessReadiness {
	r := core.HarnessReadiness{
		Executable: "available", Credentials: "cli_managed", Runtime: "native",
		Supported: profile.Harness == codex.Identity, RemoteAuthentication: "not_tested",
	}
	if role == core.HarnessGeneration {
		r.Runtime = "not_required"
	} else if profile.RuntimeRef != "" {
		r.Runtime = "unavailable"
	}

	if profile.CredentialRef != "" {
		r.Credentials = "missing"
		if privateHarnessReference(configDir, profile.CredentialRef) {
			r.Credentials = "present"
		}
	}

	if profile.ModelConfigRef != "" && !privateHarnessReference(configDir, profile.ModelConfigRef) {
		r.Supported = false
	}

	// Reasoning, private credential injection and custom model resources require
	// their adapter-owned dispatch path; they cannot be accepted and then ignored.
	if profile.Thinking != "" || profile.CredentialRef != "" || profile.ModelConfigRef != "" {
		r.Supported = false
	}

	return r
}

func privateHarnessReference(configDir, reference string) bool {
	if !filepath.IsAbs(reference) {
		reference = filepath.Join(configDir, reference)
	}

	info, err := os.Stat(reference)

	return err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0o077 == 0
}

func legacyHarnessSelection(config core.WorkspaceConfig, role core.HarnessRole, action string) (resolvedHarnessSelection, error) {
	backend, err := legacyHarnessBackend(role, action)

	view := core.HarnessSelection{
		Role: role, Action: action, Source: "legacy", Backend: backend,
		Model: "cli_default", SelectionRevision: config.Revision(),
		Readiness: core.HarnessReadiness{Executable: "not_required", Credentials: "api_managed", Runtime: "not_required", Supported: true, RemoteAuthentication: "not_tested"},
	}
	if err != nil {
		view.Readiness.Supported = false

		return resolvedHarnessSelection{View: view}, err
	}

	if backend == "openai" {
		return resolvedHarnessSelection{View: view}, nil
	}

	profile := core.HarnessProfile{Harness: codex.Identity, Command: strings.TrimSpace(os.Getenv("CHR_CODEX_COMMAND")), Model: strings.TrimSpace(os.Getenv("CHR_CODEX_MODEL"))}
	if profile.Command == "" || action == "contextual_quiz" || role == core.HarnessWorker {
		profile.Command = codex.Program
	}

	profile.Revision = core.HarnessProfileRevision(profile)
	resolved := prepareHarnessSelection(config, profile, role, action, "")
	resolved.View.Source = "legacy"

	return resolved, nil
}

func legacyHarnessBackend(role core.HarnessRole, action string) (string, error) {
	if role == core.HarnessWorker || action == "contextual_quiz" || action == "agentic_review" {
		return codex.Identity, nil
	}

	variable := map[string]string{
		string(core.QuizModelActionGenerateSpec):              "CHR_SPEC_GENERATION_BACKEND",
		string(core.QuizModelActionGenerateAcceptanceFeature): "CHR_ACCEPTANCE_FEATURE_GENERATION_BACKEND",
		string(core.QuizModelActionGeneratePlan):              "CHR_PLAN_GENERATION_BACKEND",
		string(core.QuizModelActionMaterializePlanTracking):   "CHR_PLAN_TRACKING_GENERATION_BACKEND",
		"dump_curation": "CHR_DUMP_GENERATION_BACKEND", "conversation": "CHR_CONVERSATION_GENERATION_BACKEND",
	}[action]
	if variable == "" {
		return "openai", nil
	}

	value := strings.ToLower(strings.TrimSpace(os.Getenv(variable)))
	if value == "" {
		return codex.Identity, nil
	}

	if value != "openai" && value != codex.Identity {
		return "", errors.New("unsupported legacy generation backend; select a configured harness profile")
	}

	return value, nil
}
