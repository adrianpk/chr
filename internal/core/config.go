package core

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"forge.adrianpk.com/chr/internal/core/configvalue"
)

// WorkspaceConfig describes a package boundary value.
type WorkspaceConfig struct {
	Harnesses       HarnessSelectionConfig
	HarnessAudit    HarnessSelectionAudit
	Workflow        WorkflowConfig
	TicketKeyPrefix string
	Runtime         RuntimeConfig
	Notifications   NotificationConfig
	Review          ReviewWorkflowConfig
	Work            WorkConfig
	Evidence        EvidenceConfig
	Web             WebConfig

	workflowExplicit map[string]bool
	workflowRevision string
}

// RuntimeConfig describes a package boundary value.
type RuntimeConfig struct {
	Image             string
	DockerfilePath    string
	Workdir           string
	ArtifactDir       string
	ValidationCommand string
}

// WebConfig describes local Web UI configuration.
type WebConfig struct {
	Enabled                bool
	Host                   string
	Port                   int
	OpenBrowser            bool
	AllowLAN               bool
	RequireToken           WebTokenMode
	TokenEnv               string
	AllowedOrigins         []string
	RefreshIntervalSeconds int
}

// WebTokenMode defines the local Web UI token requirement policy.
type WebTokenMode string

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
	Integration              WorkIntegrationConfig
	MergeStrategy            string
	WorkerTimeoutSeconds     int
	ValidationTimeoutSeconds int
	CheckpointStepLimit      int
	IncludeLaterTickets      bool
	MaxWorkerAttempts        int
	MaxIntegrationAttempts   int
	MaxActiveWorkers         int
	ReadyQueueCapacity       int
	StateAuthority           WorkStateAuthority
	BranchPolicy             WorkBranchPolicy
	Workspace                WorkWorkspaceConfig
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

const (
	// DefaultRuntimeImage defines a package boundary constant.
	DefaultRuntimeImage = "chr-agent-runtime:local"
	// DefaultRuntimeDockerfilePath defines a package boundary constant.
	DefaultRuntimeDockerfilePath = ".chr/default/runtime/agent/Dockerfile"
	// DefaultRuntimeWorkdir defines a package boundary constant.
	DefaultRuntimeWorkdir = "/workspace"
	// DefaultRuntimeArtifactDir defines the zero-configuration shared run artifact root.
	DefaultRuntimeArtifactDir = "ops/default/run"
	// PrivateRuntimeArtifactDir is the retired private run artifact root accepted only for migration to ops layout.
	PrivateRuntimeArtifactDir = ".chr/default/runtime/runs"
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

	// WebTokenModeAuto requires token behavior according to resolved exposure.
	WebTokenModeAuto WebTokenMode = "auto"
	// WebTokenModeTrue requires a token for local Web UI requests.
	WebTokenModeTrue WebTokenMode = "true"
	// WebTokenModeFalse disables token checks for local Web UI requests.
	WebTokenModeFalse WebTokenMode = "false"
	// WebTokenModeAlways requires a token for every non-static local Web UI request.
	WebTokenModeAlways WebTokenMode = "always"

	// DefaultWebHost defines a package boundary constant.
	DefaultWebHost = "127.0.0.1"
	// DefaultWebPort defines a package boundary constant.
	DefaultWebPort = 7331
	// DefaultWebTokenEnv defines a package boundary constant.
	DefaultWebTokenEnv = "CHR_WEB_TOKEN"
	// DefaultWebRefreshIntervalSeconds defines a package boundary constant.
	DefaultWebRefreshIntervalSeconds = 2

	// WorkBranchPolicyModeIntegrationBranch defines a package boundary constant.
	WorkBranchPolicyModeIntegrationBranch WorkBranchPolicyMode = "integration_branch"
	// WorkBranchPolicyModeCurrentBranch defines a package boundary constant.
	WorkBranchPolicyModeCurrentBranch WorkBranchPolicyMode = "current_branch"
	// WorkBranchPolicyModeNoGitEnforcement defines a package boundary constant.
	WorkBranchPolicyModeNoGitEnforcement WorkBranchPolicyMode = "no_git_enforcement"

	// WorkWorkspaceModeProject defines a package boundary constant.
	WorkWorkspaceModeProject WorkWorkspaceMode = "project"
	// WorkWorkspaceModeWorktree defines a package boundary constant.
	WorkWorkspaceModeWorktree WorkWorkspaceMode = "worktree"

	// WorktreeBasePolicyRemoteIntegrationBranch defines a package boundary constant.
	WorktreeBasePolicyRemoteIntegrationBranch WorktreeBasePolicy = "remote_integration_branch"
	// WorktreeBasePolicyLocalIntegrationBranch defines a package boundary constant.
	WorktreeBasePolicyLocalIntegrationBranch WorktreeBasePolicy = "local_integration_branch"
	// WorktreeBasePolicyCurrentBranch defines a package boundary constant.
	WorktreeBasePolicyCurrentBranch WorktreeBasePolicy = "current_branch"

	// WorktreeCleanupPolicyAfterMerge defines a package boundary constant.
	WorktreeCleanupPolicyAfterMerge WorktreeCleanupPolicy = "after_merge"
	// WorktreeCleanupPolicyManual defines a package boundary constant.
	WorktreeCleanupPolicyManual WorktreeCleanupPolicy = "manual"
	// WorktreeCleanupPolicyNever defines a package boundary constant.
	WorktreeCleanupPolicyNever WorktreeCleanupPolicy = "never"

	// DefaultWorkIntegrationBranch defines a package boundary constant.
	DefaultWorkIntegrationBranch = "dev"
	// DefaultWorkCanonicalRemote defines a package boundary constant.
	DefaultWorkCanonicalRemote = "origin"
	// DefaultWorktreeRoot defines a package boundary constant.
	DefaultWorktreeRoot = "ops/default/worktree"

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

// DefaultWebConfig executes the corresponding package operation.
func DefaultWebConfig() WebConfig {
	return WebConfig{
		Enabled:                false,
		Host:                   DefaultWebHost,
		Port:                   DefaultWebPort,
		OpenBrowser:            false,
		AllowLAN:               false,
		RequireToken:           WebTokenModeAuto,
		TokenEnv:               DefaultWebTokenEnv,
		AllowedOrigins:         nil,
		RefreshIntervalSeconds: DefaultWebRefreshIntervalSeconds,
	}
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
		Integration:            WorkIntegrationConfig{PublicationProfile: WorkflowPublicationDisabled, MaxPreparedPerTarget: 2},
		MergeStrategy:          TicketMergeStrategyRebase,
		WorkerTimeoutSeconds:   3600,
		CheckpointStepLimit:    WorkflowPageLimit,
		MaxWorkerAttempts:      3,
		MaxIntegrationAttempts: 3,
		MaxActiveWorkers:       1,
		ReadyQueueCapacity:     64,
		BranchPolicy:           DefaultWorkBranchPolicy(),
		Workspace:              DefaultWorkWorkspaceConfig(),
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

// DefaultWorkWorkspaceConfig executes the corresponding package operation.
func DefaultWorkWorkspaceConfig() WorkWorkspaceConfig {
	return WorkWorkspaceConfig{
		Mode:            WorkWorkspaceModeProject,
		WorktreeRoot:    DefaultWorktreeRoot,
		WorktreeBase:    WorktreeBasePolicyRemoteIntegrationBranch,
		WorktreeCleanup: WorktreeCleanupPolicyAfterMerge,
	}
}

// LoadWorkspaceConfig loads the corresponding package data.
func defaultWorkspaceConfig() WorkspaceConfig {
	return WorkspaceConfig{
		TicketKeyPrefix: DefaultTicketKeyPrefix,
		Runtime:         DefaultRuntimeConfig(),
		Notifications:   DefaultNotificationConfig(),
		Review:          DefaultReviewWorkflowConfig(),
		Work:            DefaultWorkConfig(),
		Evidence:        DefaultEvidenceConfig(),
		Web:             DefaultWebConfig(),
	}
}

// LoadWorkspaceConfig resolves explicit values over versioned defaults.
func LoadWorkspaceConfig(projectDir string) (WorkspaceConfig, error) {
	if err := requireNoWorkflowUpgrade(projectDir); err != nil {
		return WorkspaceConfig{}, err
	}

	config := defaultWorkspaceConfig()
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
	data, err := io.ReadAll(io.LimitReader(file, WorkflowFileLimit+1))
	if err != nil {
		return fmt.Errorf("read workspace config: %w", err)
	}

	resolved, err := parseWorkflowConfig(data)
	if err != nil {
		return err
	}

	*config = resolved

	return nil
}

func applyWorkspaceConfig(data []byte, config *WorkspaceConfig) error {
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	section := ""

	for scanner.Scan() {
		line := parseWorkspaceConfigLine(scanner.Text(), section)
		section = line.Section

		if !line.Parsed {
			continue
		}

		applyWorkspaceConfigValue(config, section, line.Key, line.Value)
	}

	err := scanner.Err()
	if err != nil {
		return fmt.Errorf("read workspace config: %w", err)
	}

	return nil
}

func applyWorkspaceConfigValue(config *WorkspaceConfig, section, key, value string) {
	if setWorkspaceConfigRootValue(config, section, key, value) {
		return
	}

	switch section {
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
	case "web":
		setWebConfigValue(&config.Web, key, value)
	}
}

func setWorkspaceConfigRootValue(config *WorkspaceConfig, section, key, value string) bool {
	if section != "" && section != "tickets" {
		return false
	}

	if key != "key_prefix" {
		return true
	}

	configvalue.SetString(&config.TicketKeyPrefix, value)

	return true
}

type workspaceConfigLine struct {
	Key, Value, Section string
	Parsed              bool
}

func parseWorkspaceConfigLine(rawLine, section string) workspaceConfigLine {
	line := strings.TrimSpace(rawLine)
	if line == "" || strings.HasPrefix(line, "#") {
		return workspaceConfigLine{Section: section}
	}

	indent := len(rawLine) - len(strings.TrimLeft(rawLine, " "))
	section = workspaceConfigSiblingSection(section, indent)

	if indent == 0 && strings.HasSuffix(line, ":") {
		return workspaceConfigLine{Section: strings.TrimSuffix(line, ":")}
	}

	if indent > 0 && strings.HasSuffix(line, ":") && !strings.Contains(section, ".") {
		return workspaceConfigLine{Section: section + "." + strings.TrimSuffix(line, ":")}
	}

	if indent == 0 {
		section = ""
	}

	key, value, ok := strings.Cut(line, ":")
	if !ok {
		return workspaceConfigLine{Section: section}
	}

	return workspaceConfigLine{Key: strings.TrimSpace(key), Value: configvalue.String(value), Section: section, Parsed: true}
}

func workspaceConfigSiblingSection(section string, indent int) string {
	if indent == 2 && section == "work.integration" {
		return "work"
	}

	return section
}

func setRuntimeConfigValue(config *RuntimeConfig, key, value string) {
	switch key {
	case "image":
		configvalue.SetString(&config.Image, value)
	case "dockerfile":
		configvalue.SetString(&config.DockerfilePath, value)
	case "workdir":
		configvalue.SetString(&config.Workdir, value)
	case "artifact_dir":
		configvalue.SetString(&config.ArtifactDir, value)
	case "validation_command":
		configvalue.SetString(&config.ValidationCommand, value)
	}
}

func setNotificationConfigValue(config *NotificationConfig, key, value string) {
	switch key {
	case "telegram_enabled":
		config.Telegram.Enabled = configvalue.Bool(value)
	case "telegram_bot_token_env":
		configvalue.SetString(&config.Telegram.BotTokenEnv, value)
	case "telegram_chat_id":
		configvalue.SetString(&config.Telegram.ChatID, value)
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
		config.Providers.Forgejo.Enabled = configvalue.Bool(value)
	case "forgejo_cli":
		configvalue.SetString(&config.Providers.Forgejo.CLI, value)
	case "forgejo_token_env":
		configvalue.SetString(&config.Providers.Forgejo.TokenEnv, value)
	case "github_enabled":
		config.Providers.GitHub.Enabled = configvalue.Bool(value)
	case "github_cli":
		configvalue.SetString(&config.Providers.GitHub.CLI, value)
	case "github_token_env":
		configvalue.SetString(&config.Providers.GitHub.TokenEnv, value)
	}
}

func setWorkConfigValue(config *WorkConfig, key, value string) {
	if setWorkCapacityConfigValue(config, key, value) {
		return
	}

	if key == "state_authority" {
		config.StateAuthority = WorkStateAuthority(value)

		return
	}

	if setWorkBranchPolicyConfigValue(&config.BranchPolicy, key, value) {
		return
	}

	setWorkWorkspaceConfigValue(&config.Workspace, key, value)
}

func setWorkBranchPolicyConfigValue(config *WorkBranchPolicy, key, value string) bool {
	switch key {
	case "mode":
		if value != "" {
			config.Mode = WorkBranchPolicyMode(value)
		}
	case "integration_branch":
		configvalue.SetString(&config.IntegrationBranch, value)
	case "canonical_remote":
		configvalue.SetString(&config.CanonicalRemote, value)
	case "require_clean_worktree":
		config.RequireCleanWorktree = configvalue.Bool(value)
	case "require_remote_tracking":
		config.RequireRemoteTracking = configvalue.Bool(value)
	case "allow_safe_repair":
		config.AllowSafeRepair = configvalue.Bool(value)
	case "allow_branch_creation":
		config.AllowBranchCreation = configvalue.Bool(value)
	default:
		return false
	}

	return true
}

func setWorkWorkspaceConfigValue(config *WorkWorkspaceConfig, key, value string) {
	switch key {
	case "workspace_mode":
		if value != "" {
			config.Mode = WorkWorkspaceMode(value)
		}
	case "worktree_root":
		configvalue.SetString(&config.WorktreeRoot, value)
	case "worktree_base":
		if value != "" {
			config.WorktreeBase = WorktreeBasePolicy(value)
		}
	case "worktree_cleanup":
		if value != "" {
			config.WorktreeCleanup = WorktreeCleanupPolicy(value)
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
		config.Enabled = configvalue.Bool(value)
	case "strict_ready_for_review":
		config.StrictReadyForReview = configvalue.Bool(value)
	case "allow_user_edits":
		config.AllowUserEdits = configvalue.Bool(value)
	case "execution_report_copy":
		config.ExecutionReportCopy = configvalue.Bool(value)
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

func setWebConfigValue(config *WebConfig, key, value string) {
	switch key {
	case "enabled":
		config.Enabled = configvalue.Bool(value)
	case "host":
		configvalue.SetString(&config.Host, value)
	case "port":
		setPositiveInt(&config.Port, value)
	case "open_browser":
		config.OpenBrowser = configvalue.Bool(value)
	case "allow_lan":
		config.AllowLAN = configvalue.Bool(value)
	case "require_token":
		mode := string(config.RequireToken)
		configvalue.SetString(&mode, value)

		config.RequireToken = WebTokenMode(mode)
	case "token_env":
		configvalue.SetString(&config.TokenEnv, value)
	case "allowed_origins":
		config.AllowedOrigins = configStringListValue(value)
	case "refresh_interval_seconds":
		setPositiveInt(&config.RefreshIntervalSeconds, value)
	}
}

func setPositiveInt(target *int, value string) {
	parsed, ok := configPositiveIntValue(value)
	if ok {
		*target = parsed
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

func configStringListValue(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" || value == "[]" {
		return nil
	}

	value = strings.TrimPrefix(strings.TrimSuffix(value, "]"), "[")
	parts := strings.Split(value, ",")

	result := make([]string, 0, len(parts))
	for _, part := range parts {
		item := configvalue.String(part)
		if item != "" {
			result = append(result, item)
		}
	}

	return result
}
