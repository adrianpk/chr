//go:build !linux

package process

import (
	"context"
	"errors"
)

// EnableChildReaping requires the Linux worker runtime.
func EnableChildReaping() error { return errors.New("owned worker child reaping requires Linux") }

// ReapChildren requires the Linux worker runtime.
func ReapChildren(context.Context) error {
	return errors.New("owned worker child reaping requires Linux")
}
