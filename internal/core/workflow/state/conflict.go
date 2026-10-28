// Package state owns shared durable workflow state values.
package state

import (
	"errors"
)

// ErrConflict requires fresh observation or explicit reconciliation.
var ErrConflict = errors.New("workflow revision or image conflict")
