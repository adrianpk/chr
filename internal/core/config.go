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
	Evidence        EvidenceConfig
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
	Workspace    WorkWorkspaceConfig
}

// EvidenceConfig describes a package boundary value.
type EvidenceConfig struct {
	Strategy EvidenceStrategyConfig
}

// EvidenceStrategyConfig describes a package boundary value.
type EvidenceStrategyConfig struct {
	Enabled              bool
	Mode                 EvidenceStrategyMode
	TicketChecklist      EvidenceTicketChecklistPolicy
	GherkinPolicy        EvidenceGherkinPolicy
	StrictReadyForReview bool
	AllowUserEdits       bool
	ExecutionReportCopy  bool
}

// EvidenceStrategyMode defines a package boundary type.
type EvidenceStrategyMode string

// EvidenceTicketChecklistPolicy defines a package boundary type.
type EvidenceTicketChecklistPolicy string

// EvidenceGherkinPolicy defines a package boundary type.
type EvidenceGherkinPolicy string

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

// WorkWorkspaceConfig describes workspace ownership for agent execution.
type WorkWorkspaceConfig struct {
	Mode            WorkWorkspaceMode
	WorktreeRoot    string
	WorktreeBase    WorktreeBasePolicy
	WorktreeCleanup WorktreeCleanupPolicy
}

// WorkWorkspaceMode defines a package boundary type.
type WorkWorkspaceMode string

// WorktreeBasePolicy defines a package boundary type.
type WorktreeBasePolicy string

// WorktreeCleanupPolicy defines a package boundary type.
type WorktreeCleanupPolicy string

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

	// EvidenceStrategyModeInline defines a package boundary constant.
	EvidenceStrategyModeInline EvidenceStrategyMode = "inline"
	// EvidenceStrategyModeSeparatePass defines a package boundary constant.
	EvidenceStrategyModeSeparatePass EvidenceStrategyMode = "separate_pass"

	// EvidenceTicketChecklistPolicyRequired defines a package boundary constant.
	EvidenceTicketChecklistPolicyRequired EvidenceTicketChecklistPolicy = "required"
	// EvidenceTicketChecklistPolicyOptional defines a package boundary constant.
	EvidenceTicketChecklistPolicyOptional EvidenceTicketChecklistPolicy = "optional"

	// EvidenceGherkinPolicyActiveWhenApplicable defines a package boundary constant.
	EvidenceGherkinPolicyActiveWhenApplicable EvidenceGherkinPolicy = "active_when_applicable"
	// EvidenceGherkinPolicyDisabled defines a package boundary constant.
	EvidenceGherkinPolicyDisabled EvidenceGherkinPolicy = "disabled"
	// EvidenceGherkinPolicyManualOnly defines a package boundary constant.
	EvidenceGherkinPolicyManualOnly EvidenceGherkinPolicy = "manual_only"
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

// DefaultEvidenceConfig executes the corresponding package operation.
func DefaultEvidenceConfig() EvidenceConfig {
	return EvidenceConfig{
		Strategy: DefaultEvidenceStrategyConfig(),
	}
}

// DefaultEvidenceStrategyConfig executes the corresponding package operation.
func DefaultEvidenceStrategyConfig() EvidenceStrategyConfig {
	return EvidenceStrategyConfig{
		Enabled:              true,
		Mode:                 EvidenceStrategyModeInline,
		TicketChecklist:      EvidenceTicketChecklistPolicyRequired,
		GherkinPolicy:        EvidenceGherkinPolicyActiveWhenApplicable,
		StrictReadyForReview: true,
		AllowUserEdits:       true,
		ExecutionReportCopy:  true,
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
		Evidence:        DefaultEvidenceConfig(),
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
	case "evidence", "evidence.strategy":
		setEvidenceConfigValue(&config.Evidence, key, value)
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

	if indent > 0 && strings.HasSuffix(line, ":") && !strings.Contains(section, ".") {
		return "", "", section + "." + strings.TrimSuffix(line, ":"), false
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
	case "workspace_mode":
		if value != "" {
			config.Workspace.Mode = WorkWorkspaceMode(value)
		}
	case "worktree_root":
		setString(&config.Workspace.WorktreeRoot, value)
	case "worktree_base":
		if value != "" {
			config.Workspace.WorktreeBase = WorktreeBasePolicy(value)
		}
	case "worktree_cleanup":
		if value != "" {
			config.Workspace.WorktreeCleanup = WorktreeCleanupPolicy(value)
		}
	}
}

func setEvidenceConfigValue(config *EvidenceConfig, key, value string) {
	setEvidenceStrategyConfigValue(&config.Strategy, key, value)
}

func setEvidenceStrategyConfigValue(config *EvidenceStrategyConfig, key, value string) {
	if setEvidenceStrategyBoolConfigValue(config, key, value) {
		return
	}

	setEvidenceStrategyEnumConfigValue(config, key, value)
}

func setEvidenceStrategyBoolConfigValue(config *EvidenceStrategyConfig, key, value string) bool {
	switch key {
	case "enabled":
		config.Enabled = configBoolValue(value)
	case "strict_ready_for_review":
		config.StrictReadyForReview = configBoolValue(value)
	case "allow_user_edits":
		config.AllowUserEdits = configBoolValue(value)
	case "execution_report_copy":
		config.ExecutionReportCopy = configBoolValue(value)
	default:
		return false
	}

	return true
}

func setEvidenceStrategyEnumConfigValue(config *EvidenceStrategyConfig, key, value string) {
	if value == "" {
		return
	}

	switch key {
	case "mode":
		config.Mode = EvidenceStrategyMode(value)
	case "ticket_checklist":
		config.TicketChecklist = EvidenceTicketChecklistPolicy(value)
	case "gherkin_policy":
		config.GherkinPolicy = EvidenceGherkinPolicy(value)
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
