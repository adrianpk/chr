package core

import (
	"fmt"
	"strconv"
)

// WorkflowConfig selects prospective defaults without reinterpreting legacy data.
type WorkflowConfig struct {
	Version int `json:"version"`
}

// WorkflowConfigurationValue reports one effective value and its policy source.
type WorkflowConfigurationValue struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Source string `json:"source"`
}

// WorkflowConfiguration is a bounded, credential-free effective configuration.
type WorkflowConfiguration struct {
	Version  int                          `json:"version"`
	Revision string                       `json:"revision"`
	Values   []WorkflowConfigurationValue `json:"values"`
}

type workflowConfigInteger struct {
	key     string
	value   *int
	minimum int
	maximum int
}

func workflowConfigIntegers(config *WorkspaceConfig) []workflowConfigInteger {
	work := &config.Work

	return []workflowConfigInteger{
		{"workflow.version", &config.Workflow.Version, 1, 1},
		{"work.max_active_workers", &work.MaxActiveWorkers, 1, 8},
		{"work.ready_queue_capacity", &work.ReadyQueueCapacity, 1, 1024},
		{"work.integration.max_prepared_per_target", &work.Integration.MaxPreparedPerTarget, 1, 8},
		{"work.worker_timeout_seconds", &work.WorkerTimeoutSeconds, 1, 86400},
		{"work.validation_timeout_seconds", &work.ValidationTimeoutSeconds, 1, 86400},
		{"work.checkpoint_step_limit", &work.CheckpointStepLimit, 1, WorkflowPageLimit},
		{"work.max_worker_attempts", &work.MaxWorkerAttempts, 1, 128},
		{"work.max_integration_attempts", &work.MaxIntegrationAttempts, 1, 10},
	}
}

func parseWorkflowConfig(data []byte) (WorkspaceConfig, error) {
	config := defaultWorkspaceConfig()

	values, err := readWorkflowConfigValues(data)
	if err != nil {
		return config, err
	}

	if value, present := values["workflow.version"]; present {
		if value != "1" {
			return config, fmt.Errorf("workflow.version must be 1; absence selects legacy defaults")
		}

		config.Workflow.Version = 1
		config.Review.Mode = ReviewWorkflowModeTrustedAutoMerge
		config.Work.StateAuthority = WorkStateAuthorityTickets
		config.Work.Workspace.Mode = WorkWorkspaceModeWorktree
		config.Work.ValidationTimeoutSeconds = 3600
	}

	if err := applyWorkspaceConfig(data, &config); err != nil {
		return config, err
	}

	if err := applyWorkflowConfigValues(&config, values); err != nil {
		return config, err
	}

	if err := applyHarnessSelection(&config, values); err != nil {
		return config, err
	}

	if err := applyHarnessSelectionAudit(&config, values); err != nil {
		return config, err
	}

	config.workflowRevision = workflowHash(data)
	config.workflowExplicit = make(map[string]bool, len(values))

	for key := range values {
		config.workflowExplicit[key] = true
	}

	return config, validateWorkflowConfigValues(config, values)
}

func applyWorkflowConfigValues(config *WorkspaceConfig, values map[string]string) error {
	for _, field := range workflowConfigIntegers(config) {
		value, present := values[field.key]
		if !present {
			continue
		}

		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < field.minimum || parsed > field.maximum {
			return fmt.Errorf("%s must be %d..%d", field.key, field.minimum, field.maximum)
		}

		*field.value = parsed
	}

	if value, present := values["work.merge_strategy"]; present {
		config.Work.MergeStrategy = value
	}

	if value, present := values["work.integration.publication_profile"]; present {
		config.Work.Integration.PublicationProfile = value
	}

	return applyWorkflowTicketSelection(config, values)
}

func applyWorkflowTicketSelection(config *WorkspaceConfig, values map[string]string) error {
	if value, present := values["work.include_later_tickets"]; present {
		if value != "true" && value != "false" {
			return fmt.Errorf("work.include_later_tickets must be true or false")
		}

		config.Work.IncludeLaterTickets = value == "true"
	}

	return nil
}

func validateWorkflowConfigValues(config WorkspaceConfig, values map[string]string) error {
	choices := []struct {
		key     string
		allowed []string
	}{
		{"work.state_authority", []string{"tickets", "tracker_compat"}},
		{"review.mode", []string{"human_review", "agentic_review", "trusted_auto_merge"}},
		{"work.workspace_mode", []string{"project", "worktree"}},
		{"work.merge_strategy", []string{"merge", "rebase", "squash"}},
		{"work.integration.publication_profile", []string{WorkflowPublicationDisabled, WorkflowPublicationGitAtomicDualRef}},
	}
	for _, choice := range choices {
		if value, present := values[choice.key]; present && !workflowConfigChoice(value, choice.allowed) {
			return fmt.Errorf("unsupported %s", choice.key)
		}
	}

	if err := validateWorkflowPolicyScalars(values); err != nil {
		return err
	}

	if values["work.workspace_mode"] == "project" && (WorkAuthority(config.Work) != WorkStateAuthorityTrackerCompat || config.Work.MaxActiveWorkers != 1) {
		return fmt.Errorf("work.workspace_mode project requires explicitly selected serial tracker_compat work")
	}

	return ValidateWorkConfig(config.Work)
}

func validateWorkflowPolicyScalars(values map[string]string) error {
	for _, key := range []string{"work.require_clean_worktree", "work.require_remote_tracking", "work.allow_safe_repair", "work.allow_branch_creation", "review.forgejo_enabled", "review.github_enabled"} {
		value, present := values[key]
		if present && !workflowConfigChoice(value, []string{"true", "false", "yes", "no", "1", "0", "on", "off"}) {
			return fmt.Errorf("unsupported boolean: %s", key)
		}
	}

	if value, present := values["review.non_blocking_comment_limit"]; present {
		limit, err := strconv.Atoi(value)
		if err != nil || limit < 0 {
			return fmt.Errorf("review.non_blocking_comment_limit must be a nonnegative integer")
		}
	}

	for _, key := range []string{"work.mode", "work.integration_branch", "work.canonical_remote", "work.worktree_root", "work.worktree_base", "work.worktree_cleanup"} {
		if value, present := values[key]; present && value == "" {
			return fmt.Errorf("%s cannot be empty", key)
		}
	}

	return nil
}

func workflowConfigChoice(value string, allowed []string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}

	return false
}

// ResolveWorkflowConfiguration observes configuration without upgrading a project.
// Only a human-review restriction may override the current publication policy.
func ResolveWorkflowConfiguration(projectDir string, restriction ReviewWorkflowMode) (WorkflowConfiguration, error) {
	if restriction != "" && restriction != ReviewWorkflowModeHumanReview {
		return WorkflowConfiguration{}, fmt.Errorf("an invocation may restrict review to human_review only")
	}

	config, err := LoadWorkspaceConfig(projectDir)
	if err != nil {
		return WorkflowConfiguration{}, err
	}

	return effectiveWorkflowConfiguration(config, restriction), nil
}

func effectiveWorkflowConfiguration(config WorkspaceConfig, restriction ReviewWorkflowMode) WorkflowConfiguration {
	revision := config.workflowRevision
	if revision == "" {
		revision = workflowHash(nil)
	}

	result := WorkflowConfiguration{Version: config.Workflow.Version, Revision: revision}
	for _, field := range workflowConfigIntegers(&config) {
		result.Values = append(result.Values, workflowConfigurationValue(config, field.key, strconv.Itoa(*field.value)))
	}

	for _, field := range []struct{ key, value string }{
		{"work.state_authority", string(WorkAuthority(config.Work))},
		{"review.mode", string(config.Review.Mode)},
		{"work.workspace_mode", string(config.Work.Workspace.Mode)},
		{"work.merge_strategy", config.Work.MergeStrategy},
		{"work.integration.publication_profile", config.Work.Integration.PublicationProfile},
		{"work.include_later_tickets", strconv.FormatBool(config.Work.IncludeLaterTickets)},
	} {
		value := workflowConfigurationValue(config, field.key, field.value)
		if field.key == "review.mode" && restriction != "" {
			value.Value, value.Source = string(restriction), "invocation_restriction"
		}

		result.Values = append(result.Values, value)
	}

	return result
}

func workflowConfigurationValue(config WorkspaceConfig, key, value string) WorkflowConfigurationValue {
	source := "legacy_default"
	if config.Workflow.Version == 1 {
		source = "version_default"
	}

	if config.workflowExplicit[key] {
		source = "project"
	}

	return WorkflowConfigurationValue{Key: key, Value: value, Source: source}
}
