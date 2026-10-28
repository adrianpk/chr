package state

import (
	"errors"
)

// ErrAtomicGuard blocks a provider path without both immutable guards.
var ErrAtomicGuard = errors.New("provider cannot atomically guard the source and validated target base")
