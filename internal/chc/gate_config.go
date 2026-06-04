package chc

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

// GateTraceabilityMode defines a package boundary type.
type GateTraceabilityMode string

const (
	// GateTraceabilityModeAgentOnly defines a package boundary constant.
	GateTraceabilityModeAgentOnly GateTraceabilityMode = "agent_only"
)

// GateConfig describes a package boundary value.
type GateConfig struct {
	DevMain GateDevMainConfig
}

// GateDevMainConfig describes a package boundary value.
type GateDevMainConfig struct {
	ValidationCommand string
	Traceability      GateTraceabilityConfig
	ManualChanges     GateManualChangesConfig
	HighRiskPaths     []string
}

// GateTraceabilityConfig describes a package boundary value.
type GateTraceabilityConfig struct {
	Mode                       GateTraceabilityMode
	FailOnMissingAgentEvidence bool
	WarnOnManualChanges        bool
}

// GateManualChangesConfig describes a package boundary value.
type GateManualChangesConfig struct {
	List                bool
	WarnOnHighRiskPaths bool
}

// DefaultGateConfig executes the corresponding package operation.
func DefaultGateConfig() GateConfig {
	return GateConfig{
		DevMain: GateDevMainConfig{
			ValidationCommand: DefaultGateValidationCommand,
			Traceability: GateTraceabilityConfig{
				Mode:                       GateTraceabilityModeAgentOnly,
				FailOnMissingAgentEvidence: true,
			},
			ManualChanges: GateManualChangesConfig{
				List: true,
			},
		},
	}
}

// LoadGateConfig loads the corresponding package data.
func LoadGateConfig(projectDir string) (GateConfig, error) {
	config := DefaultGateConfig()
	path := UserPath(projectDir, "config", "gates.yml")

	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return config, nil
	}

	if err != nil {
		return GateConfig{}, fmt.Errorf("open gate config: %w", err)
	}

	defer func() {
		//nolint:errcheck // Scanner errors are checked below.
		_ = file.Close()
	}()

	err = readGateConfig(file, &config)
	if err != nil {
		return GateConfig{}, err
	}

	return config, nil
}

func readGateConfig(file *os.File, config *GateConfig) error {
	scanner := bufio.NewScanner(file)
	section := ""
	listKey := ""

	for scanner.Scan() {
		line := scanner.Text()
		key, value, updatedSection, updatedListKey, ok := parseGateConfigLine(line, section, listKey)
		section = updatedSection
		listKey = updatedListKey

		if !ok {
			continue
		}

		setGateConfigValue(config, section, listKey, key, value)
	}

	err := scanner.Err()
	if err != nil {
		return fmt.Errorf("read gate config: %w", err)
	}

	return nil
}

func parseGateConfigLine(rawLine, section, listKey string) (string, string, string, string, bool) {
	line := strings.TrimSpace(rawLine)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", section, listKey, false
	}

	indent := len(rawLine) - len(strings.TrimLeft(rawLine, " "))
	if indent == 4 && line == "high_risk_paths:" {
		return "", "", section, "high_risk_paths", false
	}

	if strings.HasSuffix(line, ":") {
		return "", "", gateConfigSection(indent, strings.TrimSuffix(line, ":")), "", false
	}

	if strings.HasPrefix(line, "- ") {
		return "", configStringValue(strings.TrimPrefix(line, "- ")), section, listKey, true
	}

	key, value, ok := strings.Cut(line, ":")
	if !ok {
		return "", "", section, listKey, false
	}

	key = strings.TrimSpace(key)
	value = configStringValue(value)

	return key, value, section, "", true
}

func gateConfigSection(indent int, key string) string {
	switch {
	case indent == 0 && key == "gates":
		return ""
	case indent == 2 && key == "dev_main":
		return "dev_main"
	case indent == 4 && key == "traceability":
		return "traceability"
	case indent == 4 && key == "manual_changes":
		return "manual_changes"
	default:
		return key
	}
}

func setGateConfigValue(config *GateConfig, section, listKey, key, value string) {
	if listKey == "high_risk_paths" && value != "" {
		config.DevMain.HighRiskPaths = append(config.DevMain.HighRiskPaths, value)

		return
	}

	switch section {
	case "dev_main":
		setGateDevMainConfigValue(&config.DevMain, key, value)
	case "traceability":
		setGateTraceabilityConfigValue(&config.DevMain.Traceability, key, value)
	case "manual_changes":
		setGateManualChangesConfigValue(&config.DevMain.ManualChanges, key, value)
	}
}

func setGateDevMainConfigValue(config *GateDevMainConfig, key, value string) {
	if key == "validation_command" {
		setString(&config.ValidationCommand, value)
	}
}

func setGateTraceabilityConfigValue(config *GateTraceabilityConfig, key, value string) {
	switch key {
	case "mode":
		if value != "" {
			config.Mode = GateTraceabilityMode(value)
		}
	case "fail_on_missing_agent_evidence":
		config.FailOnMissingAgentEvidence = configBoolValue(value)
	case "warn_on_manual_changes":
		config.WarnOnManualChanges = configBoolValue(value)
	}
}

func setGateManualChangesConfigValue(config *GateManualChangesConfig, key, value string) {
	switch key {
	case "list":
		config.List = configBoolValue(value)
	case "warn_on_high_risk_paths":
		config.WarnOnHighRiskPaths = configBoolValue(value)
	}
}
