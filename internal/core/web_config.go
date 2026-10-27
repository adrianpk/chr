package core

import (
	"fmt"
	"net"
	"strings"

	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// ValidateWebConfig validates the local Web UI configuration.
func ValidateWebConfig(config workspaceconfig.WebConfig) error {
	for _, check := range []func(workspaceconfig.WebConfig) error{
		validateWebBindConfig,
		validateWebTokenConfig,
		validateWebRefreshConfig,
		validateWebAllowedOrigins,
	} {
		if err := check(config); err != nil {
			return err
		}
	}

	return nil
}

func validateWebBindConfig(config workspaceconfig.WebConfig) error {
	if strings.TrimSpace(config.Host) == "" {
		return fmt.Errorf("web host is required")
	}

	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("web port must be between 1 and 65535")
	}

	return nil
}

func validateWebTokenConfig(config workspaceconfig.WebConfig) error {
	switch config.RequireToken {
	case workspaceconfig.WebTokenModeAuto, workspaceconfig.WebTokenModeTrue, workspaceconfig.WebTokenModeFalse, workspaceconfig.WebTokenModeAlways:
	default:
		return fmt.Errorf("invalid web require_token mode: %s", config.RequireToken)
	}

	if config.TokenEnv == "" && (config.RequireToken == workspaceconfig.WebTokenModeTrue || config.RequireToken == workspaceconfig.WebTokenModeAlways) {
		return fmt.Errorf("web token_env is required when token mode requires a token")
	}

	return nil
}

func validateWebRefreshConfig(config workspaceconfig.WebConfig) error {
	if config.RefreshIntervalSeconds <= 0 {
		return fmt.Errorf("web refresh_interval_seconds must be positive")
	}

	return nil
}

func validateWebAllowedOrigins(config workspaceconfig.WebConfig) error {
	for _, origin := range config.AllowedOrigins {
		if strings.TrimSpace(origin) == "" {
			return fmt.Errorf("web allowed_origins cannot contain empty values")
		}
	}

	return nil
}

// WebHostIsLoopback reports whether a configured Web UI host is loopback-only.
func WebHostIsLoopback(host string) bool {
	host = strings.TrimSpace(host)
	if strings.EqualFold(host, "localhost") {
		return true
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	return ip.IsLoopback()
}
