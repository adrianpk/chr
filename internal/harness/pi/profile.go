package pi

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// ValidateGenerationProfile performs local checks without probing the executable.
func ValidateGenerationProfile(provider, model, thinking string) error {
	for _, value := range []string{provider, model} {
		if value == "" || len(value) > 1024 || strings.TrimSpace(value) != value || strings.ContainsAny(value, "\x00\r\n") || !utf8.ValidString(value) {
			return fmt.Errorf("pi generation requires bounded exact provider and model identifiers: %w", ErrCapability)
		}
	}

	if model == "cli_default" || !SupportsThinking(thinking) {
		return fmt.Errorf("prepare an exact Pi model and supported thinking level: %w", ErrCapability)
	}

	return nil
}
