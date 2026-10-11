package core

import (
	"bufio"
	"bytes"
	"fmt"
	"slices"
	"strings"
)

func readWorkflowConfigValues(data []byte) (map[string]string, error) {
	if len(data) > WorkflowFileLimit {
		return nil, fmt.Errorf("workspace config exceeds 1 MiB")
	}

	values, sections := map[string]string{}, map[string]bool{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 4096), WorkflowFileLimit)

	section := ""

	for scanner.Scan() {
		line := scanner.Text()
		previous := strings.Split(section, ".")[0]

		key, value, next, ok := parseWorkspaceConfigLine(line, section)
		if err := guardWorkflowConfigIndent(line, previous, key, ok); err != nil {
			return nil, err
		}

		section = next
		if err := guardWorkflowNativeSection(line, section, ok, sections); err != nil {
			return nil, err
		}

		if err := guardWorkflowConfigLine(line, section, key, ok, sections); err != nil {
			return nil, err
		}

		if !ok || !workflowConfigSection(section) {
			continue
		}

		qualified := section + "." + key
		if _, present := values[qualified]; present {
			return nil, fmt.Errorf("duplicate workspace config key: %s", qualified)
		}

		if err := guardWorkflowConfigValue(qualified, value); err != nil {
			return nil, err
		}

		values[qualified] = value
	}

	return values, scanner.Err()
}

func guardWorkflowConfigValue(key, value string) error {
	if !knownWorkflowConfigKey(key) || len(value) > 4096 {
		return fmt.Errorf("unsupported workspace config key or value: %s", key)
	}

	return nil
}

func workflowConfigSection(section string) bool {
	return section == "workflow" || section == "work" || section == "review" || section == "work.integration"
}

func guardWorkflowConfigIndent(raw, previous, key string, parsed bool) error {
	if !workflowConfigSection(previous) {
		return nil
	}

	if strings.HasPrefix(raw, "\t") {
		return fmt.Errorf("workflow configuration cannot use tab indentation")
	}

	if parsed && !strings.HasPrefix(raw, " ") && !workflowConfigChoice(key, []string{"version", "workspace_root", "key_prefix"}) {
		return fmt.Errorf("workflow configuration keys must be indented")
	}

	return nil
}

func guardWorkflowConfigLine(raw, section, key string, parsed bool, sections map[string]bool) error {
	line := strings.TrimSpace(raw)
	if line == "" || strings.HasPrefix(line, "#") {
		return nil
	}

	indent := len(raw) - len(strings.TrimLeft(raw, " "))

	root := strings.Split(section, ".")[0]
	if indent == 0 {
		return guardWorkflowConfigRoot(line, root, key, parsed, sections)
	}

	if workflowConfigSection(root) {
		return guardWorkflowConfigScalar(raw, section, root, indent, parsed)
	}

	return nil
}

func guardWorkflowConfigRoot(line, root, key string, parsed bool, sections map[string]bool) error {
	if parsed && workflowConfigSection(key) {
		return fmt.Errorf("%s must be a configuration mapping", key)
	}

	if !parsed && workflowConfigSection(root) && line == root+":" {
		if sections[root] {
			return fmt.Errorf("duplicate workspace config section: %s", root)
		}

		sections[root] = true
	}

	return nil
}

func guardWorkflowConfigScalar(raw, section, root string, indent int, parsed bool) error {
	if section == "work.integration" {
		return guardWorkflowNativeScalar(raw, indent, parsed)
	}

	if !parsed || indent != 2 || section != root || strings.Contains(raw, "\t") {
		return fmt.Errorf("%s configuration requires scalar keys indented by two spaces", root)
	}

	return nil
}

func guardWorkflowNativeSection(raw, section string, parsed bool, sections map[string]bool) error {
	if parsed || strings.TrimSpace(raw) != "integration:" || section != "work.integration" {
		return nil
	}

	if sections[section] {
		return fmt.Errorf("duplicate workspace config section: %s", section)
	}

	sections[section] = true

	return nil
}

func guardWorkflowNativeScalar(raw string, indent int, parsed bool) error {
	if !parsed && indent == 2 && strings.TrimSpace(raw) == "integration:" {
		return nil
	}

	if !parsed || indent != 4 || strings.Contains(raw, "\t") {
		return fmt.Errorf("work.integration requires scalar keys indented by four spaces")
	}

	return nil
}

func knownWorkflowConfigKey(key string) bool {
	config := WorkspaceConfig{}
	for _, field := range workflowConfigIntegers(&config) {
		if key == field.key {
			return true
		}
	}

	return slices.Contains([]string{
		"work.integration.publication_profile",
		"work.state_authority", "work.workspace_mode", "work.merge_strategy", "work.include_later_tickets",
		"work.mode", "work.integration_branch", "work.canonical_remote", "work.require_clean_worktree",
		"work.require_remote_tracking", "work.allow_safe_repair", "work.allow_branch_creation",
		"work.worktree_root", "work.worktree_base", "work.worktree_cleanup",
		"review.mode", "review.non_blocking_comment_limit", "review.forgejo_enabled", "review.forgejo_cli",
		"review.forgejo_token_env", "review.github_enabled", "review.github_cli", "review.github_token_env",
	}, key)
}
