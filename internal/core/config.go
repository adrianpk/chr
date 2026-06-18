package core

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// WorkspaceConfig describes a package boundary value.
type WorkspaceConfig struct {
	TicketKeyPrefix string
	Runtime         RuntimeConfig
	Notifications   NotificationConfig
	Review          ReviewWorkflowConfig
	Work            WorkConfig
}

// RuntimeConfig describes a package boundary value.
type RuntimeConfig struct {
	Image             string
	DockerfilePath    string
	Workdir           string
	ArtifactDir       string
	ValidationCommand string
}

// NotificationConfig describes a package boundary value.
type NotificationConfig struct {
	Telegram TelegramNotificationConfig
}

// ReviewWorkflowConfig describes a package boundary value.
type ReviewWorkflowConfig struct {
	Mode                    ReviewWorkflowMode
	NonBlockingCommentLimit int
	Providers               ReviewProviderConfig
}

// ReviewWorkflowMode defines a package boundary type.
type ReviewWorkflowMode string

// ReviewProviderConfig describes a package boundary value.
type ReviewProviderConfig struct {
	Forgejo ReviewProviderTargetConfig
	GitHub  ReviewProviderTargetConfig
}

// ReviewProviderTargetConfig describes a package boundary value.
type ReviewProviderTargetConfig struct {
	Enabled  bool
	CLI      string
	TokenEnv string
}

// WorkConfig describes a package boundary value.
type WorkConfig struct {
	BranchPolicy WorkBranchPolicy
}

// WorkBranchPolicy describes a package boundary value.
type WorkBranchPolicy struct {
	Mode                  WorkBranchPolicyMode
	IntegrationBranch     string
	CanonicalRemote       string
	RequireCleanWorktree  bool
	RequireRemoteTracking bool
	AllowSafeRepair       bool
	AllowBranchCreation   bool
}

// WorkBranchPolicyMode defines a package boundary type.
type WorkBranchPolicyMode string

// TelegramNotificationConfig describes a package boundary value.
type TelegramNotificationConfig struct {
	Enabled     bool
	BotTokenEnv string
	ChatID      string
}

const (
	// DefaultRuntimeImage defines a package boundary constant.
	DefaultRuntimeImage = "chc-agent-runtime:local"
	// DefaultRuntimeDockerfilePath defines a package boundary constant.
	DefaultRuntimeDockerfilePath = ".chc/user/runtime/agent/Dockerfile"
	// DefaultRuntimeWorkdir defines a package boundary constant.
	DefaultRuntimeWorkdir = "/workspace"
	// DefaultRuntimeArtifactDir defines a package boundary constant.
	DefaultRuntimeArtifactDir = ".chc/user/runtime/runs"
	// DefaultRuntimeValidationCommand defines a package boundary constant.
	DefaultRuntimeValidationCommand = "make verify"

	// ReviewWorkflowModeAgenticReview defines a package boundary constant.
	ReviewWorkflowModeAgenticReview ReviewWorkflowMode = "agentic_review"
	// ReviewWorkflowModeHumanReview defines a package boundary constant.
	ReviewWorkflowModeHumanReview ReviewWorkflowMode = "human_review"
	// ReviewWorkflowModeTrustedAutoMerge defines a package boundary constant.
	ReviewWorkflowModeTrustedAutoMerge ReviewWorkflowMode = "trusted_auto_merge"

	// DefaultReviewWorkflowMode defines a package boundary constant.
	DefaultReviewWorkflowMode = ReviewWorkflowModeHumanReview
	// DefaultReviewWorkflowNonBlockingCommentLimit defines a package boundary constant.
	DefaultReviewWorkflowNonBlockingCommentLimit = 3

	// DefaultReviewProviderForgejoCLI defines a package boundary constant.
	DefaultReviewProviderForgejoCLI = "tea"
	// DefaultReviewProviderForgejoTokenEnv defines a package boundary constant.
	DefaultReviewProviderForgejoTokenEnv = "FORGEJO_TOKEN"
	// DefaultReviewProviderGitHubCLI defines a package boundary constant.
	DefaultReviewProviderGitHubCLI = "gh"
	// DefaultReviewProviderGitHubTokenEnv defines a package boundary constant.
	DefaultReviewProviderGitHubTokenEnv = "GITHUB_TOKEN"

	// WorkBranchPolicyModeIntegrationBranch defines a package boundary constant.
	WorkBranchPolicyModeIntegrationBranch WorkBranchPolicyMode = "integration_branch"
	// WorkBranchPolicyModeCurrentBranch defines a package boundary constant.
	WorkBranchPolicyModeCurrentBranch WorkBranchPolicyMode = "current_branch"
	// WorkBranchPolicyModeNoGitEnforcement defines a package boundary constant.
	WorkBranchPolicyModeNoGitEnforcement WorkBranchPolicyMode = "no_git_enforcement"

	// DefaultWorkIntegrationBranch defines a package boundary constant.
	DefaultWorkIntegrationBranch = "dev"
	// DefaultWorkCanonicalRemote defines a package boundary constant.
	DefaultWorkCanonicalRemote = "origin"
)

// DefaultRuntimeConfig executes the corresponding package operation.
func DefaultRuntimeConfig() RuntimeConfig {
	return RuntimeConfig{
		Image:             DefaultRuntimeImage,
		DockerfilePath:    DefaultRuntimeDockerfilePath,
		Workdir:           DefaultRuntimeWorkdir,
		ArtifactDir:       DefaultRuntimeArtifactDir,
		ValidationCommand: DefaultRuntimeValidationCommand,
	}
}

// DefaultNotificationConfig executes the corresponding package operation.
func DefaultNotificationConfig() NotificationConfig {
	return NotificationConfig{}
}

// DefaultReviewWorkflowConfig executes the corresponding package operation.
func DefaultReviewWorkflowConfig() ReviewWorkflowConfig {
	return ReviewWorkflowConfig{
		Mode:                    DefaultReviewWorkflowMode,
		NonBlockingCommentLimit: DefaultReviewWorkflowNonBlockingCommentLimit,
		Providers:               DefaultReviewProviderConfig(),
	}
}

// DefaultReviewProviderConfig executes the corresponding package operation.
func DefaultReviewProviderConfig() ReviewProviderConfig {
	return ReviewProviderConfig{
		Forgejo: ReviewProviderTargetConfig{
			Enabled:  true,
			CLI:      DefaultReviewProviderForgejoCLI,
			TokenEnv: DefaultReviewProviderForgejoTokenEnv,
		},
		GitHub: ReviewProviderTargetConfig{
			Enabled:  true,
			CLI:      DefaultReviewProviderGitHubCLI,
			TokenEnv: DefaultReviewProviderGitHubTokenEnv,
		},
	}
}

// DefaultWorkConfig executes the corresponding package operation.
func DefaultWorkConfig() WorkConfig {
	return WorkConfig{
		BranchPolicy: DefaultWorkBranchPolicy(),
	}
}

// DefaultWorkBranchPolicy executes the corresponding package operation.
func DefaultWorkBranchPolicy() WorkBranchPolicy {
	return WorkBranchPolicy{
		Mode:                  WorkBranchPolicyModeIntegrationBranch,
		IntegrationBranch:     DefaultWorkIntegrationBranch,
		CanonicalRemote:       DefaultWorkCanonicalRemote,
		RequireCleanWorktree:  true,
		RequireRemoteTracking: true,
		AllowSafeRepair:       true,
		AllowBranchCreation:   false,
	}
}

// LoadWorkspaceConfig loads the corresponding package data.
func LoadWorkspaceConfig(projectDir string) (WorkspaceConfig, error) {
	config := WorkspaceConfig{
		TicketKeyPrefix: DefaultTicketKeyPrefix,
		Runtime:         DefaultRuntimeConfig(),
		Notifications:   DefaultNotificationConfig(),
		Review:          DefaultReviewWorkflowConfig(),
		Work:            DefaultWorkConfig(),
	}
	path := UserPath(projectDir, "config.yml")

	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return config, nil
	}

	if err != nil {
		return WorkspaceConfig{}, fmt.Errorf("open workspace config: %w", err)
	}

	defer func() {
		//nolint:errcheck // The config file is read-only here; scanner errors are checked below.
		_ = file.Close()
	}()

	err = readWorkspaceConfig(file, &config)
	if err != nil {
		return WorkspaceConfig{}, err
	}

	return config, nil
}

func readWorkspaceConfig(file *os.File, config *WorkspaceConfig) error {
	scanner := bufio.NewScanner(file)
	section := ""

	for scanner.Scan() {
		key, value, updatedSection, ok := parseWorkspaceConfigLine(scanner.Text(), section)
		section = updatedSection

		if !ok {
			continue
		}

		applyWorkspaceConfigValue(config, section, key, value)
	}

	err := scanner.Err()
	if err != nil {
		return fmt.Errorf("read workspace config: %w", err)
	}

	return nil
}

func applyWorkspaceConfigValue(config *WorkspaceConfig, section, key, value string) {
	switch section {
	case "":
		if key == "key_prefix" {
			setString(&config.TicketKeyPrefix, value)
		}
	case "tickets":
		if key == "key_prefix" {
			setString(&config.TicketKeyPrefix, value)
		}
	case "runtime":
		setRuntimeConfigValue(&config.Runtime, key, value)
	case "notifications":
		setNotificationConfigValue(&config.Notifications, key, value)
	case "review":
		setReviewWorkflowConfigValue(&config.Review, key, value)
	case "work":
		setWorkConfigValue(&config.Work, key, value)
	}
}

func parseWorkspaceConfigLine(rawLine, section string) (string, string, string, bool) {
	line := strings.TrimSpace(rawLine)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", section, false
	}

	indent := len(rawLine) - len(strings.TrimLeft(rawLine, " "))
	if indent == 0 && strings.HasSuffix(line, ":") {
		return "", "", strings.TrimSuffix(line, ":"), false
	}

	if indent == 0 {
		section = ""
	}

	key, value, ok := strings.Cut(line, ":")
	if !ok {
		return "", "", section, false
	}

	return strings.TrimSpace(key), configStringValue(value), section, true
}

func setRuntimeConfigValue(config *RuntimeConfig, key, value string) {
	switch key {
	case "image":
		setString(&config.Image, value)
	case "dockerfile":
		setString(&config.DockerfilePath, value)
	case "workdir":
		setString(&config.Workdir, value)
	case "artifact_dir":
		setString(&config.ArtifactDir, value)
	case "validation_command":
		setString(&config.ValidationCommand, value)
	}
}

func setNotificationConfigValue(config *NotificationConfig, key, value string) {
	switch key {
	case "telegram_enabled":
		config.Telegram.Enabled = configBoolValue(value)
	case "telegram_bot_token_env":
		setString(&config.Telegram.BotTokenEnv, value)
	case "telegram_chat_id":
		setString(&config.Telegram.ChatID, value)
	}
}

func setReviewWorkflowConfigValue(config *ReviewWorkflowConfig, key, value string) {
	switch key {
	case "mode":
		if value != "" {
			config.Mode = ReviewWorkflowMode(value)
		}
	case "non_blocking_comment_limit":
		setPositiveInt(&config.NonBlockingCommentLimit, value)
	case "forgejo_enabled":
		config.Providers.Forgejo.Enabled = configBoolValue(value)
	case "forgejo_cli":
		setString(&config.Providers.Forgejo.CLI, value)
	case "forgejo_token_env":
		setString(&config.Providers.Forgejo.TokenEnv, value)
	case "github_enabled":
		config.Providers.GitHub.Enabled = configBoolValue(value)
	case "github_cli":
		setString(&config.Providers.GitHub.CLI, value)
	case "github_token_env":
		setString(&config.Providers.GitHub.TokenEnv, value)
	}
}

func setWorkConfigValue(config *WorkConfig, key, value string) {
	switch key {
	case "mode":
		if value != "" {
			config.BranchPolicy.Mode = WorkBranchPolicyMode(value)
		}
	case "integration_branch":
		setString(&config.BranchPolicy.IntegrationBranch, value)
	case "canonical_remote":
		setString(&config.BranchPolicy.CanonicalRemote, value)
	case "require_clean_worktree":
		config.BranchPolicy.RequireCleanWorktree = configBoolValue(value)
	case "require_remote_tracking":
		config.BranchPolicy.RequireRemoteTracking = configBoolValue(value)
	case "allow_safe_repair":
		config.BranchPolicy.AllowSafeRepair = configBoolValue(value)
	case "allow_branch_creation":
		config.BranchPolicy.AllowBranchCreation = configBoolValue(value)
	}
}

func setString(target *string, value string) {
	if value != "" {
		*target = value
	}
}

func setPositiveInt(target *int, value string) {
	parsed, ok := configPositiveIntValue(value)
	if ok {
		*target = parsed
	}
}

func configStringValue(value string) string {
	return strings.Trim(strings.TrimSpace(value), `"'`)
}

func configBoolValue(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "yes", "1", "on":
		return true
	default:
		return false
	}
}

func configPositiveIntValue(value string) (int, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0, false
	}

	return parsed, true
}
