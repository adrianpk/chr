package core

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadWorkspaceConfigUsesDefaultsWhenMissing(t *testing.T) {
	dir := t.TempDir()

	config, err := LoadWorkspaceConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig returned error: %v", err)
	}

	if config.TicketKeyPrefix != DefaultTicketKeyPrefix {
		t.Fatalf("TicketKeyPrefix = %q, want %q", config.TicketKeyPrefix, DefaultTicketKeyPrefix)
	}

	assertDefaultRuntimeConfig(t, config.Runtime)

	if config.Notifications != DefaultNotificationConfig() {
		t.Fatalf("Notifications = %#v, want %#v", config.Notifications, DefaultNotificationConfig())
	}

	if config.Review != DefaultReviewWorkflowConfig() {
		t.Fatalf("Review = %#v, want %#v", config.Review, DefaultReviewWorkflowConfig())
	}

	if config.Work != DefaultWorkConfig() {
		t.Fatalf("Work = %#v, want %#v", config.Work, DefaultWorkConfig())
	}

	if config.Evidence != DefaultEvidenceConfig() {
		t.Fatalf("Evidence = %#v, want %#v", config.Evidence, DefaultEvidenceConfig())
	}

	if !reflect.DeepEqual(config.Web, DefaultWebConfig()) {
		t.Fatalf("Web = %#v, want %#v", config.Web, DefaultWebConfig())
	}
}

func TestLoadWorkspaceConfigParsesTicketAndRuntimeConfig(t *testing.T) {
	dir := t.TempDir()
	writeWorkspaceConfig(t, dir, `version: 1
workspace_root: .chc
tickets:
  key_prefix: PRJ
runtime:
  image: custom-agent:latest
  dockerfile: .chc/user/runtime/custom.Dockerfile
  workdir: /repo
  artifact_dir: .chc/user/runtime/custom-runs
  validation_command: go test ./...
`)

	config, err := LoadWorkspaceConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig returned error: %v", err)
	}

	if config.TicketKeyPrefix != "PRJ" {
		t.Fatalf("TicketKeyPrefix = %q, want PRJ", config.TicketKeyPrefix)
	}

	want := RuntimeConfig{
		Image:             "custom-agent:latest",
		DockerfilePath:    ".chc/user/runtime/custom.Dockerfile",
		Workdir:           "/repo",
		ArtifactDir:       ".chc/user/runtime/custom-runs",
		ValidationCommand: "go test ./...",
	}
	if config.Runtime != want {
		t.Fatalf("Runtime = %#v, want %#v", config.Runtime, want)
	}
}

func TestLoadWorkspaceConfigParsesNotificationConfig(t *testing.T) {
	dir := t.TempDir()
	writeWorkspaceConfig(t, dir, `notifications:
  telegram_enabled: true
  telegram_bot_token_env: CHC_TELEGRAM_BOT_TOKEN
  telegram_chat_id: "123456"
`)

	config, err := LoadWorkspaceConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig returned error: %v", err)
	}

	want := NotificationConfig{
		Telegram: TelegramNotificationConfig{
			Enabled:     true,
			BotTokenEnv: "CHC_TELEGRAM_BOT_TOKEN",
			ChatID:      "123456",
		},
	}
	if config.Notifications != want {
		t.Fatalf("Notifications = %#v, want %#v", config.Notifications, want)
	}
}

func TestLoadWorkspaceConfigParsesReviewWorkflowConfig(t *testing.T) {
	tests := []struct {
		name string
		mode ReviewWorkflowMode
	}{
		{name: "agentic", mode: ReviewWorkflowModeAgenticReview},
		{name: "human", mode: ReviewWorkflowModeHumanReview},
		{name: "trusted", mode: ReviewWorkflowModeTrustedAutoMerge},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			writeWorkspaceConfig(t, dir, `review:
  mode: `+string(test.mode)+`
  non_blocking_comment_limit: 2
`)

			config, err := LoadWorkspaceConfig(dir)
			if err != nil {
				t.Fatalf("LoadWorkspaceConfig returned error: %v", err)
			}

			want := ReviewWorkflowConfig{
				Mode:                    test.mode,
				NonBlockingCommentLimit: 2,
				Providers:               DefaultReviewProviderConfig(),
			}
			if config.Review != want {
				t.Fatalf("Review = %#v, want %#v", config.Review, want)
			}
		})
	}
}

func TestLoadWorkspaceConfigParsesReviewProviderConfig(t *testing.T) {
	dir := t.TempDir()
	writeWorkspaceConfig(t, dir, `review:
  forgejo_enabled: true
  forgejo_cli: tea-dev
  forgejo_token_env: CHC_FORGEJO_TOKEN
  github_enabled: false
  github_cli: gh-dev
  github_token_env: CHC_GITHUB_TOKEN
`)

	config, err := LoadWorkspaceConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig returned error: %v", err)
	}

	want := ReviewProviderConfig{
		Forgejo: ReviewProviderTargetConfig{
			Enabled:  true,
			CLI:      "tea-dev",
			TokenEnv: "CHC_FORGEJO_TOKEN",
		},
		GitHub: ReviewProviderTargetConfig{
			Enabled:  false,
			CLI:      "gh-dev",
			TokenEnv: "CHC_GITHUB_TOKEN",
		},
	}
	if config.Review.Providers != want {
		t.Fatalf("Review.Providers = %#v, want %#v", config.Review.Providers, want)
	}
}

func TestLoadWorkspaceConfigParsesWorkConfig(t *testing.T) {
	dir := t.TempDir()
	writeWorkspaceConfig(t, dir, `work:
  mode: current_branch
  integration_branch: develop
  canonical_remote: upstream
  require_clean_worktree: false
  require_remote_tracking: false
  allow_safe_repair: false
  allow_branch_creation: true
  workspace_mode: worktree
  worktree_root: .chc/user/runtime/custom-worktrees
  worktree_base: local_integration_branch
  worktree_cleanup: manual
`)

	config, err := LoadWorkspaceConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig returned error: %v", err)
	}

	want := WorkConfig{
		BranchPolicy: WorkBranchPolicy{
			Mode:                  WorkBranchPolicyModeCurrentBranch,
			IntegrationBranch:     "develop",
			CanonicalRemote:       "upstream",
			RequireCleanWorktree:  false,
			RequireRemoteTracking: false,
			AllowSafeRepair:       false,
			AllowBranchCreation:   true,
		},
		Workspace: WorkWorkspaceConfig{
			Mode:            WorkWorkspaceModeWorktree,
			WorktreeRoot:    ".chc/user/runtime/custom-worktrees",
			WorktreeBase:    WorktreeBasePolicyLocalIntegrationBranch,
			WorktreeCleanup: WorktreeCleanupPolicyManual,
		},
	}
	if config.Work != want {
		t.Fatalf("Work = %#v, want %#v", config.Work, want)
	}
}

func TestLoadWorkspaceConfigKeepsWorkWorkspaceDefaultsForMissingKeys(t *testing.T) {
	dir := t.TempDir()
	writeWorkspaceConfig(t, dir, `work:
  workspace_mode: worktree
`)

	config, err := LoadWorkspaceConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig returned error: %v", err)
	}

	want := DefaultWorkConfig()
	want.Workspace.Mode = WorkWorkspaceModeWorktree

	if config.Work != want {
		t.Fatalf("Work = %#v, want %#v", config.Work, want)
	}
}

func TestLoadWorkspaceConfigParsesEvidenceStrategyConfig(t *testing.T) {
	dir := t.TempDir()
	writeWorkspaceConfig(t, dir, `evidence:
  strategy:
    enabled: false
    mode: separate_pass
    ticket_checklist: optional
    gherkin_policy: manual_only
    strict_ready_for_review: false
    allow_user_edits: false
    execution_report_copy: false
`)

	config, err := LoadWorkspaceConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig returned error: %v", err)
	}

	want := EvidenceConfig{
		Strategy: EvidenceStrategyConfig{
			Enabled:              false,
			Mode:                 EvidenceStrategyModeSeparatePass,
			TicketChecklist:      EvidenceTicketChecklistPolicyOptional,
			GherkinPolicy:        EvidenceGherkinPolicyManualOnly,
			StrictReadyForReview: false,
			AllowUserEdits:       false,
			ExecutionReportCopy:  false,
		},
	}
	if config.Evidence != want {
		t.Fatalf("Evidence = %#v, want %#v", config.Evidence, want)
	}
}

func TestLoadWorkspaceConfigKeepsEvidenceDefaultsForMissingKeys(t *testing.T) {
	dir := t.TempDir()
	writeWorkspaceConfig(t, dir, `evidence:
  strategy:
    gherkin_policy: disabled
`)

	config, err := LoadWorkspaceConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig returned error: %v", err)
	}

	want := DefaultEvidenceConfig()
	want.Strategy.GherkinPolicy = EvidenceGherkinPolicyDisabled

	if config.Evidence != want {
		t.Fatalf("Evidence = %#v, want %#v", config.Evidence, want)
	}
}

func TestLoadWorkspaceConfigParsesWebConfig(t *testing.T) {
	dir := t.TempDir()
	writeWorkspaceConfig(t, dir, `web:
  enabled: true
  host: 0.0.0.0
  port: 7444
  open_browser: true
  require_token: true
  token_env: CUSTOM_CHC_WEB_TOKEN
  allowed_origins: "https://example.test,https://ops.example.test"
  refresh_interval_seconds: 5
`)

	config, err := LoadWorkspaceConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig returned error: %v", err)
	}

	want := WebConfig{
		Enabled:                true,
		Host:                   "0.0.0.0",
		Port:                   7444,
		OpenBrowser:            true,
		RequireToken:           WebTokenModeTrue,
		TokenEnv:               "CUSTOM_CHC_WEB_TOKEN",
		AllowedOrigins:         []string{"https://example.test", "https://ops.example.test"},
		RefreshIntervalSeconds: 5,
	}
	if !reflect.DeepEqual(config.Web, want) {
		t.Fatalf("Web = %#v, want %#v", config.Web, want)
	}
}

func TestLoadWorkspaceConfigKeepsWebDefaultsForMissingKeys(t *testing.T) {
	dir := t.TempDir()
	writeWorkspaceConfig(t, dir, `web:
  enabled: true
`)

	config, err := LoadWorkspaceConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig returned error: %v", err)
	}

	want := DefaultWebConfig()
	want.Enabled = true

	if !reflect.DeepEqual(config.Web, want) {
		t.Fatalf("Web = %#v, want %#v", config.Web, want)
	}
}

func TestValidateWebConfigAcceptsDefaultConfig(t *testing.T) {
	err := ValidateWebConfig(DefaultWebConfig())
	if err != nil {
		t.Fatalf("ValidateWebConfig returned error: %v", err)
	}
}

func TestValidateWebConfigRejectsInvalidBindAndTokenConfig(t *testing.T) {
	assertValidateWebConfigRejects(t, []struct {
		name    string
		config  WebConfig
		wantErr string
	}{
		{
			name: "host",
			config: WebConfig{
				Port:                   DefaultWebPort,
				RequireToken:           WebTokenModeAuto,
				TokenEnv:               DefaultWebTokenEnv,
				RefreshIntervalSeconds: DefaultWebRefreshIntervalSeconds,
			},
			wantErr: "web host is required",
		},
		{
			name: "port",
			config: WebConfig{
				Host:                   DefaultWebHost,
				Port:                   0,
				RequireToken:           WebTokenModeAuto,
				TokenEnv:               DefaultWebTokenEnv,
				RefreshIntervalSeconds: DefaultWebRefreshIntervalSeconds,
			},
			wantErr: "web port must be between 1 and 65535",
		},
		{
			name: "token mode",
			config: WebConfig{
				Host:                   DefaultWebHost,
				Port:                   DefaultWebPort,
				RequireToken:           "sometimes",
				TokenEnv:               DefaultWebTokenEnv,
				RefreshIntervalSeconds: DefaultWebRefreshIntervalSeconds,
			},
			wantErr: "invalid web require_token mode: sometimes",
		},
		{
			name: "token env",
			config: WebConfig{
				Host:                   DefaultWebHost,
				Port:                   DefaultWebPort,
				RequireToken:           WebTokenModeTrue,
				RefreshIntervalSeconds: DefaultWebRefreshIntervalSeconds,
			},
			wantErr: "web token_env is required when token mode requires a token",
		},
	})
}

func TestValidateWebConfigRejectsInvalidRefreshAndOriginConfig(t *testing.T) {
	assertValidateWebConfigRejects(t, []struct {
		name    string
		config  WebConfig
		wantErr string
	}{
		{
			name: "refresh",
			config: WebConfig{
				Host:         DefaultWebHost,
				Port:         DefaultWebPort,
				RequireToken: WebTokenModeAuto,
				TokenEnv:     DefaultWebTokenEnv,
			},
			wantErr: "web refresh_interval_seconds must be positive",
		},
		{
			name: "origin",
			config: WebConfig{
				Host:                   DefaultWebHost,
				Port:                   DefaultWebPort,
				RequireToken:           WebTokenModeAuto,
				TokenEnv:               DefaultWebTokenEnv,
				AllowedOrigins:         []string{""},
				RefreshIntervalSeconds: DefaultWebRefreshIntervalSeconds,
			},
			wantErr: "web allowed_origins cannot contain empty values",
		},
	})
}

func assertValidateWebConfigRejects(t *testing.T, tests []struct {
	name    string
	config  WebConfig
	wantErr string
},
) {
	t.Helper()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateWebConfig(test.config)
			if err == nil {
				t.Fatal("ValidateWebConfig returned nil error")
			}

			if err.Error() != test.wantErr {
				t.Fatalf("ValidateWebConfig error = %q, want %q", err.Error(), test.wantErr)
			}
		})
	}
}

func TestWebHostIsLoopback(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		{host: "127.0.0.1", want: true},
		{host: "::1", want: true},
		{host: "localhost", want: true},
		{host: "0.0.0.0", want: false},
		{host: "192.168.1.10", want: false},
		{host: "example.test", want: false},
	}

	for _, test := range tests {
		t.Run(test.host, func(t *testing.T) {
			if got := WebHostIsLoopback(test.host); got != test.want {
				t.Fatalf("WebHostIsLoopback(%q) = %v, want %v", test.host, got, test.want)
			}
		})
	}
}

func TestValidateReviewWorkflowConfigRejectsInvalidConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  ReviewWorkflowConfig
		wantErr string
	}{
		{
			name:    "mode",
			config:  ReviewWorkflowConfig{Mode: "unknown", NonBlockingCommentLimit: 1},
			wantErr: "invalid review workflow mode: unknown",
		},
		{
			name:    "limit",
			config:  ReviewWorkflowConfig{Mode: ReviewWorkflowModeHumanReview, NonBlockingCommentLimit: -1},
			wantErr: "review non_blocking_comment_limit cannot be negative",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateReviewWorkflowConfig(test.config)
			if err == nil {
				t.Fatal("ValidateReviewWorkflowConfig returned nil error")
			}

			if err.Error() != test.wantErr {
				t.Fatalf("ValidateReviewWorkflowConfig error = %q, want %q", err.Error(), test.wantErr)
			}
		})
	}
}

func TestValidateEvidenceStrategyConfigRejectsInvalidConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  EvidenceStrategyConfig
		wantErr string
	}{
		{
			name:    "mode",
			config:  EvidenceStrategyConfig{Mode: "unknown", TicketChecklist: EvidenceTicketChecklistPolicyRequired, GherkinPolicy: EvidenceGherkinPolicyActiveWhenApplicable},
			wantErr: "invalid evidence strategy mode: unknown",
		},
		{
			name:    "ticket checklist",
			config:  EvidenceStrategyConfig{Mode: EvidenceStrategyModeInline, TicketChecklist: "unknown", GherkinPolicy: EvidenceGherkinPolicyActiveWhenApplicable},
			wantErr: "invalid evidence ticket_checklist policy: unknown",
		},
		{
			name:    "gherkin policy",
			config:  EvidenceStrategyConfig{Mode: EvidenceStrategyModeInline, TicketChecklist: EvidenceTicketChecklistPolicyRequired, GherkinPolicy: "unknown"},
			wantErr: "invalid evidence gherkin_policy: unknown",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateEvidenceStrategyConfig(test.config)
			if err == nil {
				t.Fatal("ValidateEvidenceStrategyConfig returned nil error")
			}

			if err.Error() != test.wantErr {
				t.Fatalf("ValidateEvidenceStrategyConfig error = %q, want %q", err.Error(), test.wantErr)
			}
		})
	}
}

func TestValidateWorkConfigRejectsInvalidBranchPolicyConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  WorkConfig
		wantErr string
	}{
		{
			name: "mode",
			config: WorkConfig{
				BranchPolicy: WorkBranchPolicy{Mode: "unknown"},
			},
			wantErr: "invalid work branch policy mode: unknown",
		},
		{
			name: "integration branch",
			config: WorkConfig{
				BranchPolicy: WorkBranchPolicy{Mode: WorkBranchPolicyModeIntegrationBranch},
			},
			wantErr: "work integration_branch is required",
		},
		{
			name: "remote",
			config: WorkConfig{
				BranchPolicy: WorkBranchPolicy{
					Mode:                  WorkBranchPolicyModeIntegrationBranch,
					IntegrationBranch:     "dev",
					RequireRemoteTracking: true,
				},
			},
			wantErr: "work canonical_remote is required when remote tracking is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertValidateWorkConfigError(t, test.config, test.wantErr)
		})
	}
}

func TestValidateWorkConfigRejectsInvalidWorkspaceConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  WorkConfig
		wantErr string
	}{
		{
			name: "workspace mode",
			config: WorkConfig{
				BranchPolicy: DefaultWorkBranchPolicy(),
				Workspace:    WorkWorkspaceConfig{Mode: "unknown"},
			},
			wantErr: "invalid work workspace_mode: unknown",
		},
		{
			name: "worktree root",
			config: WorkConfig{
				BranchPolicy: DefaultWorkBranchPolicy(),
				Workspace: WorkWorkspaceConfig{
					Mode:            WorkWorkspaceModeWorktree,
					WorktreeBase:    WorktreeBasePolicyRemoteIntegrationBranch,
					WorktreeCleanup: WorktreeCleanupPolicyAfterMerge,
				},
			},
			wantErr: "work worktree_root is required",
		},
		{
			name: "worktree base",
			config: WorkConfig{
				BranchPolicy: DefaultWorkBranchPolicy(),
				Workspace: WorkWorkspaceConfig{
					Mode:            WorkWorkspaceModeWorktree,
					WorktreeRoot:    DefaultWorktreeRoot,
					WorktreeBase:    "unknown",
					WorktreeCleanup: WorktreeCleanupPolicyAfterMerge,
				},
			},
			wantErr: "invalid work worktree_base: unknown",
		},
		{
			name: "worktree cleanup",
			config: WorkConfig{
				BranchPolicy: DefaultWorkBranchPolicy(),
				Workspace: WorkWorkspaceConfig{
					Mode:            WorkWorkspaceModeWorktree,
					WorktreeRoot:    DefaultWorktreeRoot,
					WorktreeBase:    WorktreeBasePolicyRemoteIntegrationBranch,
					WorktreeCleanup: "unknown",
				},
			},
			wantErr: "invalid work worktree_cleanup: unknown",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertValidateWorkConfigError(t, test.config, test.wantErr)
		})
	}
}

func assertValidateWorkConfigError(t *testing.T, config WorkConfig, wantErr string) {
	t.Helper()

	err := ValidateWorkConfig(config)
	if err == nil {
		t.Fatal("ValidateWorkConfig returned nil error")
	}

	if err.Error() != wantErr {
		t.Fatalf("ValidateWorkConfig error = %q, want %q", err.Error(), wantErr)
	}
}

func TestLoadWorkspaceConfigKeepsNotificationsDisabledByDefault(t *testing.T) {
	dir := t.TempDir()
	writeWorkspaceConfig(t, dir, `notifications:
  telegram_bot_token_env: CHC_TELEGRAM_BOT_TOKEN
  telegram_chat_id: "123456"
`)

	config, err := LoadWorkspaceConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig returned error: %v", err)
	}

	if config.Notifications.Telegram.Enabled {
		t.Fatal("Telegram notifications are enabled by default")
	}
}

func TestLoadWorkspaceConfigKeepsRuntimeDefaultsForMissingKeys(t *testing.T) {
	dir := t.TempDir()
	writeWorkspaceConfig(t, dir, `runtime:
  image: custom-agent:latest
`)

	config, err := LoadWorkspaceConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig returned error: %v", err)
	}

	if config.Runtime.Image != "custom-agent:latest" {
		t.Fatalf("Runtime.Image = %q, want custom-agent:latest", config.Runtime.Image)
	}

	if config.Runtime.DockerfilePath != DefaultRuntimeDockerfilePath {
		t.Fatalf("Runtime.DockerfilePath = %q, want %q", config.Runtime.DockerfilePath, DefaultRuntimeDockerfilePath)
	}
}

func TestLoadWorkspaceConfigSupportsLegacyTopLevelTicketKeyPrefix(t *testing.T) {
	dir := t.TempDir()
	writeWorkspaceConfig(t, dir, "key_prefix: LEGACY\n")

	config, err := LoadWorkspaceConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig returned error: %v", err)
	}

	if config.TicketKeyPrefix != "LEGACY" {
		t.Fatalf("TicketKeyPrefix = %q, want LEGACY", config.TicketKeyPrefix)
	}
}

func TestValidateNotificationConfigAcceptsDisabledConfig(t *testing.T) {
	err := ValidateNotificationConfig(NotificationConfig{})
	if err != nil {
		t.Fatalf("ValidateNotificationConfig returned error: %v", err)
	}
}

func TestValidateNotificationConfigAcceptsEnabledTelegramConfig(t *testing.T) {
	config := NotificationConfig{
		Telegram: TelegramNotificationConfig{
			Enabled:     true,
			BotTokenEnv: "CHC_TELEGRAM_BOT_TOKEN",
			ChatID:      "123456",
		},
	}

	err := ValidateNotificationConfig(config)
	if err != nil {
		t.Fatalf("ValidateNotificationConfig returned error: %v", err)
	}
}

func TestValidateNotificationConfigRejectsIncompleteTelegramConfig(t *testing.T) {
	for _, test := range []struct {
		name    string
		config  NotificationConfig
		wantErr string
	}{
		{
			name: "token env",
			config: NotificationConfig{
				Telegram: TelegramNotificationConfig{
					Enabled: true,
					ChatID:  "123456",
				},
			},
			wantErr: "telegram bot token environment variable is required",
		},
		{
			name: "chat id",
			config: NotificationConfig{
				Telegram: TelegramNotificationConfig{
					Enabled:     true,
					BotTokenEnv: "CHC_TELEGRAM_BOT_TOKEN",
				},
			},
			wantErr: "telegram chat id is required",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateNotificationConfig(test.config)
			if err == nil {
				t.Fatal("ValidateNotificationConfig returned nil error")
			}

			if err.Error() != test.wantErr+" when telegram notifications are enabled" {
				t.Fatalf("ValidateNotificationConfig error = %q, want %q", err.Error(), test.wantErr)
			}
		})
	}
}

func assertDefaultRuntimeConfig(t *testing.T, config RuntimeConfig) {
	t.Helper()

	if config != DefaultRuntimeConfig() {
		t.Fatalf("Runtime = %#v, want %#v", config, DefaultRuntimeConfig())
	}
}

func writeWorkspaceConfig(t *testing.T, projectDir, content string) {
	t.Helper()

	path := UserPath(projectDir, "config.yml")

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("create config directory: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}
}
