//go:build !linux

package process

import "context"

// Linux provides the worker runtime's process observation boundary. Other
// platforms retain their existing signal and direct-child wait behavior.
func waitKilledGroup(context.Context, int) error { return nil }
