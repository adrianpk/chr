package state

import (
	"fmt"
	"strings"
)

// ValidateGate checks ordered argv commands or an explicit waiver.
func ValidateGate(g Validation) error {
	if !g.Required {
		if strings.TrimSpace(g.WaiverReason) == "" {
			return fmt.Errorf("waived workflow gate requires a reason")
		}

		return nil
	}

	if len(g.Commands) == 0 || len(g.Commands) > 16 || g.TimeoutSeconds < 1 || g.TimeoutSeconds > 86400 {
		return fmt.Errorf("workflow gate command/deadline bounds exceeded")
	}

	seen := map[string]bool{}

	for _, argv := range g.Commands {
		key := strings.Join(argv, "\x00")
		if err := ValidateCommand(argv); err != nil {
			return err
		}

		if seen[key] {
			return fmt.Errorf("duplicate workflow validation command")
		}

		seen[key] = true
	}

	return nil
}

// ValidateCommand checks one bounded argv command.
func ValidateCommand(argv []string) error {
	if len(argv) == 0 || len(argv) > 64 || argv[0] == "" || len(strings.Join(argv, " ")) > 4096 || strings.Contains(strings.Join(argv, ""), "\x00") {
		return fmt.Errorf("invalid workflow validation command")
	}

	return nil
}
