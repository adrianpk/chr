package namespace

import (
	"fmt"
	"os"
	"strings"
)

const (
	// Default is used when no explicit shared namespace is configured.
	Default = "default"
	// UserEnv selects the shared operational namespace.
	UserEnv = "OPS_USER"
)

// Resolve preserves explicit, environment and default namespace precedence.
func Resolve(value string, lookupEnv func(string) string) (string, error) {
	if lookupEnv == nil {
		lookupEnv = os.Getenv
	}

	name := strings.TrimSpace(value)
	if name == "" {
		name = strings.TrimSpace(lookupEnv(UserEnv))
	}

	if name == "" {
		name = Default
	}

	if !Valid(name) {
		return "", fmt.Errorf("invalid ops namespace: %s", name)
	}

	return name, nil
}
