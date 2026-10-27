package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
	"forge.adrianpk.com/chr/internal/harness/codex"
	"forge.adrianpk.com/chr/internal/harness/pi"
)

type resolvedHarnessSelection struct {
	View         core.HarnessSelection
	Profile      core.HarnessProfile
	Executable   string
	RuntimeImage string
}

// ResolveHarnessSelection observes one coherent persisted role selection.
// Profile setup is explicit; inspection never installs or authenticates a driver.
func (s service) ResolveHarnessSelection(ctx context.Context, role harnesscore.HarnessRole, action string) (core.HarnessSelection, error) {
	config, err := core.LoadWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		return core.HarnessSelection{}, err
	}

	resolved, err := s.resolveHarnessSelection(ctx, config, role, action)

	return resolved.View, err
}

func (s service) resolveHarnessSelection(ctx context.Context, config core.WorkspaceConfig, role harnesscore.HarnessRole, action string) (resolvedHarnessSelection, error) {
	if err := ctx.Err(); err != nil {
		return resolvedHarnessSelection{}, err
	}

	name, err := config.Harnesses.Profile(role)
	if err != nil {
		return resolvedHarnessSelection{}, err
	}

	if name == "" {
		return legacyHarnessSelection(ctx, config, role, action)
	}

	if s.env == nil {
		return resolvedHarnessSelection{}, errors.New("private user configuration is required to resolve harness profiles")
	}

	profiles, err := core.LoadHarnessProfiles(s.env.ConfigDir())
	if err != nil {
		return resolvedHarnessSelection{}, err
	}

	return resolveConfiguredHarnessSelection(ctx, config, profiles, name, role, action, s.env.ConfigDir())
}

func resolveConfiguredHarnessSelection(ctx context.Context, config core.WorkspaceConfig, profiles map[string]core.HarnessProfile, name string, role harnesscore.HarnessRole, action, configDir string) (resolvedHarnessSelection, error) {
	if name == "" {
		return legacyHarnessSelection(ctx, config, role, action)
	}

	profile, found := profiles[name]
	if !found {
		view := core.HarnessSelection{Role: role, Action: action, Profile: name, Source: "project", SelectionRevision: config.Revision()}

		return resolvedHarnessSelection{View: view}, fmt.Errorf("unknown harness profile %q; configure it in private harnesses.yml", name)
	}

	resolved := prepareHarnessSelection(ctx, config, profile, role, action, configDir)
	if !resolved.View.Readiness.Ready() {
		return resolved, fmt.Errorf("harness profile %q is not locally ready for %s: %s", name, role, resolved.View.Readiness.Problem)
	}

	return resolved, nil
}

func prepareHarnessSelection(ctx context.Context, config core.WorkspaceConfig, profile core.HarnessProfile, role harnesscore.HarnessRole, action, configDir string) resolvedHarnessSelection {
	model := profile.Model
	if model == "" {
		model = "cli_default"
	}

	view := core.HarnessSelection{
		Role: role, Action: action, Source: "project", Profile: profile.Name,
		Harness: profile.Harness, Backend: profile.Harness, Model: model, ModelProvider: profile.ModelProvider,
		Thinking: profile.Thinking, SelectionRevision: config.Revision(), ProfileRevision: profile.Revision,
	}

	var runtimeImage string

	switch profile.Harness {
	case codex.Identity:
		view.Readiness = harnessProfileReadiness(ctx, profile, role, configDir)
		view.Capabilities = (codex.Runner{}).Capabilities()
	case pi.Identity:
		view.Readiness, runtimeImage = piProfileReadiness(ctx, profile, role, configDir)
		view.Capabilities = (pi.Runner{}).Capabilities()
	}

	executable, err := exec.LookPath(profile.Command)
	if err != nil {
		view.Readiness.Executable = "unavailable"
	}

	return resolvedHarnessSelection{View: view, Profile: profile, Executable: executable, RuntimeImage: runtimeImage}
}

func harnessProfileReadiness(ctx context.Context, profile core.HarnessProfile, role harnesscore.HarnessRole, configDir string) core.HarnessReadiness {
	r := core.HarnessReadiness{
		Executable: "available", Credentials: "cli_managed", Runtime: "native",
		Supported: profile.Harness == codex.Identity, RemoteAuthentication: "not_tested",
	}
	if role == harnesscore.HarnessGeneration {
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

	// Explicit Codex providers, reasoning, credentials and custom model resources require
	// their adapter-owned dispatch path; they cannot be accepted and then ignored.
	if !harnessProfileOptionsSupported(profile) {
		r.Supported = false
	}

	return r
}

func piProfileReadiness(ctx context.Context, profile core.HarnessProfile, role harnesscore.HarnessRole, configDir string) (core.HarnessReadiness, string) {
	r := core.HarnessReadiness{Executable: "available", Credentials: "missing", Runtime: "not_required", Supported: true, RemoteAuthentication: "not_tested"}
	if err := pi.ValidateGenerationProfile(profile.ModelProvider, profile.Model, profile.Thinking); err != nil {
		r.Supported, r.Problem = false, err.Error()

		return r, ""
	}

	if role == harnesscore.HarnessWorker {
		return piWorkerReadiness(ctx, profile, configDir, r)
	}

	if chr, _ := piWorkerBridgePreparation(profile.Command); chr {
		r.Supported, r.Problem = false, "configure the native Pi executable for generation in a separate profile"

		return r, ""
	}

	if _, err := pi.ResolveResources(configDir, profile.CredentialRef, profile.ModelConfigRef, profile.ModelProvider); err != nil {
		r.Problem = err.Error()
	} else {
		r.Credentials = "present"
	}

	return r, ""
}

func piWorkerReadiness(ctx context.Context, profile core.HarnessProfile, configDir string, r core.HarnessReadiness) (core.HarnessReadiness, string) {
	r.Runtime = "unavailable"
	if runtime.GOOS != "linux" {
		r.Supported, r.Problem = false, "prepare the Pi worker bridge on a supported Linux host"

		return r, ""
	}

	if _, prepared := piWorkerBridgePreparation(profile.Command); !prepared {
		r.Supported, r.Problem = false, "prepare the Linux chr worker bridge with CGO_ENABLED=0 and configure its executable as command"

		return r, ""
	}

	if _, err := pi.ResolveWorkerResources(ctx, configDir, profile.CredentialRef, profile.ModelConfigRef, profile.ModelProvider); err != nil {
		r.Problem = err.Error()

		return r, ""
	}

	r.Credentials = "present"

	image, err := pi.InspectRuntime(ctx, nil, profile.RuntimeRef)
	if err != nil {
		r.Problem = err.Error()

		return r, ""
	}

	if image.Architecture != runtime.GOARCH {
		r.Supported, r.Problem = false, "prepare a Pi image matching the Linux bridge architecture"

		return r, ""
	}

	r.Runtime = "available"

	return r, image.ID
}

func harnessProfileOptionsSupported(profile core.HarnessProfile) bool {
	return profile.Thinking == "" && profile.CredentialRef == "" && profile.ModelConfigRef == "" && (profile.Harness != codex.Identity || profile.ModelProvider == "")
}

func privateHarnessReference(configDir, reference string) bool {
	if !filepath.IsAbs(reference) {
		reference = filepath.Join(configDir, reference)
	}

	info, err := os.Stat(reference)

	return err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0o077 == 0
}

func legacyHarnessSelection(ctx context.Context, config core.WorkspaceConfig, role harnesscore.HarnessRole, action string) (resolvedHarnessSelection, error) {
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
	if profile.Command == "" || action == "contextual_quiz" || role == harnesscore.HarnessWorker {
		profile.Command = codex.Program
	}

	// Historical worker argv did not pass CHR_CODEX_MODEL. Preserve CLI-default
	// provenance rather than advertising that generation-only setting as frozen.
	if role == harnesscore.HarnessWorker {
		profile.Model = ""
	}

	profile.Revision = core.HarnessProfileRevision(profile)
	resolved := prepareHarnessSelection(ctx, config, profile, role, action, "")
	resolved.View.Source = "legacy"

	return resolved, nil
}

func legacyHarnessBackend(role harnesscore.HarnessRole, action string) (string, error) {
	if role == harnesscore.HarnessWorker || action == "contextual_quiz" || action == "agentic_review" {
		return codex.Identity, nil
	}

	variable := map[string]string{
		string(quizmodel.ActionGenerateSpec):              "CHR_SPEC_GENERATION_BACKEND",
		string(quizmodel.ActionGenerateAcceptanceFeature): "CHR_ACCEPTANCE_FEATURE_GENERATION_BACKEND",
		string(quizmodel.ActionGeneratePlan):              "CHR_PLAN_GENERATION_BACKEND",
		string(quizmodel.ActionMaterializePlanTracking):   "CHR_PLAN_TRACKING_GENERATION_BACKEND",
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
