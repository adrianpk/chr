// Package state owns shared authoring identity values.
package state

import (
	"errors"
	"regexp"
)

// ValidateSessionID checks the existing opaque authoring session identity.
func ValidateSessionID(id string) error {
	if !regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`).MatchString(id) {
		return errors.New("invalid authoring session identity")
	}

	return nil
}

// Intent separates a requested explanation, preparation and development.
type Intent string

// Only an explicit development decision can authorize a prepared scope.
const (
	IntentExplain Intent = "explain"
	IntentPrepare Intent = "prepare"
	IntentDevelop Intent = "develop"
)
