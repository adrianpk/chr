// Package gate owns Git evidence, validation and release-gate reports.
package gate

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"forge.adrianpk.com/chr/internal/core/configvalue"
)

// TraceabilityMode defines a package boundary type.
type TraceabilityMode string

const (
	// TraceabilityModeAgentOnly defines a package boundary constant.
	TraceabilityModeAgentOnly TraceabilityMode = "agent_only"
)

// Config describes a package boundary value.
type Config struct {
	DevMain DevMainConfig
}

// DevMainConfig describes a package boundary value.
type DevMainConfig struct {
	ValidationCommand string
	Traceability      TraceabilityConfig
	ManualChanges     ManualChangesConfig
	HighRiskPaths     []string
}

// TraceabilityConfig describes a package boundary value.
type TraceabilityConfig struct {
	Mode                       TraceabilityMode
	FailOnMissingAgentEvidence bool
	WarnOnManualChanges        bool
}

// ManualChangesConfig describes a package boundary value.
type ManualChangesConfig struct {
	List                bool
	WarnOnHighRiskPaths bool
}

// DefaultConfig executes the corresponding package operation.
func DefaultConfig() Config {
	return Config{
		DevMain: DevMainConfig{
			ValidationCommand: DefaultValidationCommand,
			Traceability: TraceabilityConfig{
				Mode:                       TraceabilityModeAgentOnly,
				FailOnMissingAgentEvidence: true,
			},
			ManualChanges: ManualChangesConfig{
				List: true,
			},
		},
	}
}

// LoadConfig loads the corresponding package data.
func LoadConfig(path string) (Config, error) {
	config := DefaultConfig()

	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return config, nil
	}

	if err != nil {
		return Config{}, fmt.Errorf("open gate config: %w", err)
	}

	defer func() {
		//nolint:errcheck // Read-only file; scanner errors are checked before returning.
		_ = file.Close()
	}()

	err = readConfig(file, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

func readConfig(file *os.File, config *Config) error {
	scanner := bufio.NewScanner(file)
	section := ""
	listKey := ""

	for scanner.Scan() {
		line := scanner.Text()
		lineResult := parseConfigLine(line, section, listKey)
		updatedSection, updatedListKey, ok := lineResult.Section, lineResult.ListKey, lineResult.Present
		section = updatedSection
		listKey = updatedListKey

		if !ok {
			continue
		}

		setConfigValue(config, lineResult)
	}

	err := scanner.Err()
	if err != nil {
		return fmt.Errorf("read gate config: %w", err)
	}

	return nil
}

func parseConfigLine(rawLine, section, listKey string) configLine {
	line := strings.TrimSpace(rawLine)
	if line == "" || strings.HasPrefix(line, "#") {
		return configLine{Key: "", Value: "", Section: section, ListKey: listKey, Present: false}
	}

	indent := len(rawLine) - len(strings.TrimLeft(rawLine, " "))
	if indent == 4 && line == "high_risk_paths:" {
		return configLine{Key: "", Value: "", Section: section, ListKey: "high_risk_paths", Present: false}
	}

	if strings.HasSuffix(line, ":") {
		return configLine{Key: "", Value: "", Section: configSection(indent, strings.TrimSuffix(line, ":")), ListKey: "", Present: false}
	}

	if strings.HasPrefix(line, "- ") {
		return configLine{Key: "", Value: configvalue.String(strings.TrimPrefix(line, "- ")), Section: section, ListKey: listKey, Present: true}
	}

	key, value, ok := strings.Cut(line, ":")
	if !ok {
		return configLine{Key: "", Value: "", Section: section, ListKey: listKey, Present: false}
	}

	key = strings.TrimSpace(key)
	value = configvalue.String(value)

	return configLine{Key: key, Value: value, Section: section, ListKey: "", Present: true}
}

func configSection(indent int, key string) string {
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

func setConfigValue(config *Config, line configLine) {
	if line.ListKey == "high_risk_paths" && line.Value != "" {
		config.DevMain.HighRiskPaths = append(config.DevMain.HighRiskPaths, line.Value)

		return
	}

	switch line.Section {
	case "dev_main":
		setDevMainConfigValue(&config.DevMain, line.Key, line.Value)
	case "traceability":
		setTraceabilityConfigValue(&config.DevMain.Traceability, line.Key, line.Value)
	case "manual_changes":
		setManualChangesConfigValue(&config.DevMain.ManualChanges, line.Key, line.Value)
	}
}

func setDevMainConfigValue(config *DevMainConfig, key, value string) {
	if key == "validation_command" {
		configvalue.SetString(&config.ValidationCommand, value)
	}
}

func setTraceabilityConfigValue(config *TraceabilityConfig, key, value string) {
	switch key {
	case "mode":
		if value != "" {
			config.Mode = TraceabilityMode(value)
		}
	case "fail_on_missing_agent_evidence":
		config.FailOnMissingAgentEvidence = configvalue.Bool(value)
	case "warn_on_manual_changes":
		config.WarnOnManualChanges = configvalue.Bool(value)
	}
}

func setManualChangesConfigValue(config *ManualChangesConfig, key, value string) {
	switch key {
	case "list":
		config.List = configvalue.Bool(value)
	case "warn_on_high_risk_paths":
		config.WarnOnHighRiskPaths = configvalue.Bool(value)
	}
}

type configLine struct {
	Key, Value, Section, ListKey string
	Present                      bool
}
