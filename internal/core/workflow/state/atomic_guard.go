package state

import (
	"errors"
)

// ErrWorkflowAtomicGuard blocks a provider path without both immutable guards.
var ErrWorkflowAtomicGuard = errors.New("provider cannot atomically guard the source and validated target base")
