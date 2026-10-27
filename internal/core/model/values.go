// Package model owns shared provider-neutral failure identities.
package model

import "errors"

// ErrCapacity indicates both bounded provider slots are occupied.
var ErrCapacity = errors.New("structured model capacity exhausted")

// ErrOutput identifies malformed or oversized provider output.
var ErrOutput = errors.New("invalid or oversized structured model output")
