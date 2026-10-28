package workflow

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

const (
	// ReviewProviderDiagnosticStatusReady defines a package boundary constant.
	ReviewProviderDiagnosticStatusReady = "ready"
	// ReviewProviderDiagnosticStatusBlocked defines a package boundary constant.
	ReviewProviderDiagnosticStatusBlocked = "blocked"
	// ReviewProviderDiagnosticStatusUnsupported defines a package boundary constant.
	ReviewProviderDiagnosticStatusUnsupported = "unsupported"

	// ReviewProviderDiagnosticCheckOK defines a package boundary constant.
	ReviewProviderDiagnosticCheckOK = "ok"
	// ReviewProviderDiagnosticCheckFail defines a package boundary constant.
	ReviewProviderDiagnosticCheckFail = "fail"
	// ReviewProviderDiagnosticCheckSkipped defines a package boundary constant.
	ReviewProviderDiagnosticCheckSkipped = "skipped"
)

// ReviewProviderDiagnosticOptions describes a package boundary value.
type ReviewProviderDiagnosticOptions struct {
	Config workspaceconfig.ReviewProviderConfig
	URL    string
	DryRun bool
	Runner ReviewProviderDiagnosticRunner
}

// ReviewProviderDiagnosticRunner defines a package boundary interface.
type ReviewProviderDiagnosticRunner interface {
	LookPath(file string) (string, error)
	LookupEnv(key string) string
	Run(ctx context.Context, command string, args ...string) error
}

// OSReviewProviderDiagnosticRunner describes a package boundary value.
type OSReviewProviderDiagnosticRunner struct{}

// ReviewProviderDiagnostic describes a package boundary value.
type ReviewProviderDiagnostic struct {
	Status    string
	Provider  workflowforge.Provider
	Reference workflowforge.Reference
	Checks    []ReviewProviderDiagnosticCheck
}

// ReviewProviderDiagnosticCheck describes a package boundary value.
type ReviewProviderDiagnosticCheck struct {
	Name   string
	Status string
	Reason string
	Detail string
}

// LookPath executes the corresponding package operation.
func (OSReviewProviderDiagnosticRunner) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// LookupEnv executes the corresponding package operation.
func (OSReviewProviderDiagnosticRunner) LookupEnv(key string) string {
	return os.Getenv(key)
}

// Run executes the configured operation.
func (OSReviewProviderDiagnosticRunner) Run(ctx context.Context, command string, args ...string) error {
	cmd := exec.CommandContext(ctx, command, args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run %s: %w", command, err)
	}

	return nil
}

// RunReviewProviderDiagnostics executes the corresponding package operation.
func RunReviewProviderDiagnostics(ctx context.Context, opts ReviewProviderDiagnosticOptions) ReviewProviderDiagnostic {
	runner := opts.Runner
	if runner == nil {
		runner = OSReviewProviderDiagnosticRunner{}
	}

	reference, err := workflowforge.ParseReference(opts.URL)
	diagnostic := ReviewProviderDiagnostic{
		Status:    ReviewProviderDiagnosticStatusReady,
		Provider:  reference.Provider,
		Reference: reference,
	}

	if err != nil {
		diagnostic.Status = ReviewProviderDiagnosticStatusUnsupported
		diagnostic.addCheck("provider_url", ReviewProviderDiagnosticCheckFail, "unsupported_provider", err.Error())

		return diagnostic
	}

	diagnostic.addCheck("provider_url", ReviewProviderDiagnosticCheckOK, "detected", string(reference.Provider))

	providerConfig, ok := reviewProviderTargetConfig(opts.Config, reference.Provider)
	if !ok {
		diagnostic.Status = ReviewProviderDiagnosticStatusUnsupported
		diagnostic.addCheck("provider_enabled", ReviewProviderDiagnosticCheckFail, "unsupported_provider", string(reference.Provider))

		return diagnostic
	}

	diagnostic.addProviderChecks(ctx, runner, reference, providerConfig, opts.DryRun)

	diagnostic.Status = reviewProviderDiagnosticStatus(diagnostic.Checks)

	return diagnostic
}

func (diagnostic *ReviewProviderDiagnostic) addProviderChecks(ctx context.Context, runner ReviewProviderDiagnosticRunner, reference workflowforge.Reference, config workspaceconfig.ReviewProviderTargetConfig, dryRun bool) {
	if !diagnostic.checkProviderEnabled(reference, config) {
		return
	}

	if !diagnostic.checkProviderCLI(runner, config) {
		return
	}

	if !diagnostic.checkProviderToken(runner, config) {
		return
	}

	diagnostic.checkProviderPullRequestRead(ctx, runner, reference, config, dryRun)
}

func (diagnostic *ReviewProviderDiagnostic) checkProviderEnabled(reference workflowforge.Reference, config workspaceconfig.ReviewProviderTargetConfig) bool {
	if !config.Enabled {
		diagnostic.addCheck("provider_enabled", ReviewProviderDiagnosticCheckFail, "provider_disabled", string(reference.Provider))

		return false
	}

	diagnostic.addCheck("provider_enabled", ReviewProviderDiagnosticCheckOK, "enabled", string(reference.Provider))

	return true
}

func (diagnostic *ReviewProviderDiagnostic) checkProviderCLI(runner ReviewProviderDiagnosticRunner, config workspaceconfig.ReviewProviderTargetConfig) bool {
	if strings.TrimSpace(config.CLI) == "" {
		diagnostic.addCheck("cli_available", ReviewProviderDiagnosticCheckFail, "missing_cli_config", "")

		return false
	}

	resolvedCLI, err := runner.LookPath(config.CLI)
	if err != nil {
		diagnostic.addCheck("cli_available", ReviewProviderDiagnosticCheckFail, "missing_cli", config.CLI)

		return false
	}

	diagnostic.addCheck("cli_available", ReviewProviderDiagnosticCheckOK, "found", resolvedCLI)

	return true
}

func (diagnostic *ReviewProviderDiagnostic) checkProviderToken(runner ReviewProviderDiagnosticRunner, config workspaceconfig.ReviewProviderTargetConfig) bool {
	tokenEnv := strings.TrimSpace(config.TokenEnv)
	if tokenEnv == "" {
		diagnostic.addCheck("token_available", ReviewProviderDiagnosticCheckSkipped, "token_env_not_configured", "")

		return true
	}

	if strings.TrimSpace(runner.LookupEnv(tokenEnv)) == "" {
		diagnostic.addCheck("token_available", ReviewProviderDiagnosticCheckFail, "missing_token", tokenEnv)

		return false
	}

	diagnostic.addCheck("token_available", ReviewProviderDiagnosticCheckOK, "found", tokenEnv)

	return true
}

func (diagnostic *ReviewProviderDiagnostic) checkProviderPullRequestRead(ctx context.Context, runner ReviewProviderDiagnosticRunner, reference workflowforge.Reference, config workspaceconfig.ReviewProviderTargetConfig, dryRun bool) {
	if dryRun {
		diagnostic.addCheck("pr_readable", ReviewProviderDiagnosticCheckSkipped, "dry_run", reference.URL)

		return
	}

	if err := runner.Run(ctx, config.CLI, reviewProviderReadArgs(reference)...); err != nil {
		diagnostic.addCheck("pr_readable", ReviewProviderDiagnosticCheckFail, "pr_unreadable", err.Error())

		return
	}

	diagnostic.addCheck("pr_readable", ReviewProviderDiagnosticCheckOK, "readable", reference.URL)
}

func (diagnostic *ReviewProviderDiagnostic) addCheck(name, status, reason, detail string) {
	diagnostic.Checks = append(diagnostic.Checks, ReviewProviderDiagnosticCheck{
		Name:   name,
		Status: status,
		Reason: reason,
		Detail: detail,
	})
}

func reviewProviderTargetConfig(config workspaceconfig.ReviewProviderConfig, provider workflowforge.Provider) (workspaceconfig.ReviewProviderTargetConfig, bool) {
	switch provider {
	case workflowforge.Forgejo:
		return config.Forgejo, true
	case workflowforge.GitHub:
		return config.GitHub, true
	default:
		return workspaceconfig.ReviewProviderTargetConfig{}, false
	}
}

func reviewProviderReadArgs(reference workflowforge.Reference) []string {
	switch reference.Provider {
	case workflowforge.GitHub:
		return []string{"pr", "view", reference.URL, "--json", "number"}
	case workflowforge.Forgejo:
		apiPath := path.Join("/repos", reference.Owner, reference.Repo, "pulls", strconv.Itoa(reference.Number))

		return []string{"api", apiPath}
	default:
		return nil
	}
}

func reviewProviderDiagnosticStatus(checks []ReviewProviderDiagnosticCheck) string {
	for _, check := range checks {
		if check.Status == ReviewProviderDiagnosticCheckFail {
			if check.Reason == "unsupported_provider" {
				return ReviewProviderDiagnosticStatusUnsupported
			}

			return ReviewProviderDiagnosticStatusBlocked
		}
	}

	return ReviewProviderDiagnosticStatusReady
}
