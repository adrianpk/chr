package chc

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

// ReviewProvider defines a package boundary type.
type ReviewProvider string

const (
	// ReviewProviderForgejo defines a package boundary constant.
	ReviewProviderForgejo ReviewProvider = "forgejo"
	// ReviewProviderGitHub defines a package boundary constant.
	ReviewProviderGitHub ReviewProvider = "github"
	// ReviewProviderUnsupported defines a package boundary constant.
	ReviewProviderUnsupported ReviewProvider = "unsupported"

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

// ReviewProviderReference describes a package boundary value.
type ReviewProviderReference struct {
	Provider ReviewProvider
	URL      string
	Host     string
	Owner    string
	Repo     string
	Number   int
}

// ReviewProviderDiagnosticOptions describes a package boundary value.
type ReviewProviderDiagnosticOptions struct {
	Config ReviewProviderConfig
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
	Provider  ReviewProvider
	Reference ReviewProviderReference
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

// DetectReviewProviderReference executes the corresponding package operation.
func DetectReviewProviderReference(rawURL string) (ReviewProviderReference, error) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return ReviewProviderReference{Provider: ReviewProviderUnsupported}, errors.New("pull request URL is required")
	}

	parsed, segments, err := parseReviewProviderURL(trimmed)

	reference := ReviewProviderReference{
		Provider: ReviewProviderUnsupported,
		URL:      trimmed,
	}
	if err != nil {
		return reference, err
	}

	reference.Host = strings.ToLower(parsed.Hostname())
	reference.Owner = segments[0]
	reference.Repo = segments[1]

	reference.Number, err = parseReviewProviderPRNumber(segments[3], trimmed)
	if err != nil {
		return reference, err
	}

	reference.Provider = detectReviewProvider(reference.Host, segments[2])
	if reference.Provider == ReviewProviderUnsupported {
		return reference, fmt.Errorf("unsupported pull request URL: %s", trimmed)
	}

	return reference, nil
}

// RunReviewProviderDiagnostics executes the corresponding package operation.
func RunReviewProviderDiagnostics(ctx context.Context, opts ReviewProviderDiagnosticOptions) ReviewProviderDiagnostic {
	runner := opts.Runner
	if runner == nil {
		runner = OSReviewProviderDiagnosticRunner{}
	}

	reference, err := DetectReviewProviderReference(opts.URL)
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

func (diagnostic *ReviewProviderDiagnostic) addProviderChecks(
	ctx context.Context,
	runner ReviewProviderDiagnosticRunner,
	reference ReviewProviderReference,
	config ReviewProviderTargetConfig,
	dryRun bool,
) {
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

func (diagnostic *ReviewProviderDiagnostic) checkProviderEnabled(
	reference ReviewProviderReference,
	config ReviewProviderTargetConfig,
) bool {
	if !config.Enabled {
		diagnostic.addCheck("provider_enabled", ReviewProviderDiagnosticCheckFail, "provider_disabled", string(reference.Provider))

		return false
	}

	diagnostic.addCheck("provider_enabled", ReviewProviderDiagnosticCheckOK, "enabled", string(reference.Provider))

	return true
}

func (diagnostic *ReviewProviderDiagnostic) checkProviderCLI(
	runner ReviewProviderDiagnosticRunner,
	config ReviewProviderTargetConfig,
) bool {
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

func (diagnostic *ReviewProviderDiagnostic) checkProviderToken(
	runner ReviewProviderDiagnosticRunner,
	config ReviewProviderTargetConfig,
) bool {
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

func (diagnostic *ReviewProviderDiagnostic) checkProviderPullRequestRead(
	ctx context.Context,
	runner ReviewProviderDiagnosticRunner,
	reference ReviewProviderReference,
	config ReviewProviderTargetConfig,
	dryRun bool,
) {
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

func reviewProviderTargetConfig(config ReviewProviderConfig, provider ReviewProvider) (ReviewProviderTargetConfig, bool) {
	switch provider {
	case ReviewProviderForgejo:
		return config.Forgejo, true
	case ReviewProviderGitHub:
		return config.GitHub, true
	default:
		return ReviewProviderTargetConfig{}, false
	}
}

func reviewProviderReadArgs(reference ReviewProviderReference) []string {
	switch reference.Provider {
	case ReviewProviderGitHub:
		return []string{"pr", "view", reference.URL, "--json", "number"}
	case ReviewProviderForgejo:
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

func cleanPathSegments(rawPath string) []string {
	parts := strings.Split(strings.Trim(rawPath, "/"), "/")
	segments := make([]string, 0, len(parts))

	for _, part := range parts {
		if part != "" {
			segments = append(segments, part)
		}
	}

	return segments
}

func parseReviewProviderURL(rawURL string) (*url.URL, []string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, nil, fmt.Errorf("invalid pull request URL: %s", rawURL)
	}

	segments := cleanPathSegments(parsed.Path)
	if len(segments) < 4 {
		return nil, nil, fmt.Errorf("unsupported pull request URL: %s", rawURL)
	}

	return parsed, segments, nil
}

func parseReviewProviderPRNumber(rawNumber, rawURL string) (int, error) {
	number, err := strconv.Atoi(rawNumber)
	if err != nil || number <= 0 {
		return 0, fmt.Errorf("unsupported pull request URL: %s", rawURL)
	}

	return number, nil
}

func detectReviewProvider(host, pathMarker string) ReviewProvider {
	if strings.EqualFold(host, "github.com") && pathMarker == "pull" {
		return ReviewProviderGitHub
	}

	if pathMarker == "pulls" {
		return ReviewProviderForgejo
	}

	return ReviewProviderUnsupported
}
