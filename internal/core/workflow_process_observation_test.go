package core

import (
	"os"
	"syscall"
	"testing"
)

// TestWorkflowProcessObservation covers the kernel's two process-exit errors,
// while retaining fail-closed behavior for unreadable or invalid observations.
func TestWorkflowProcessObservation(t *testing.T) {
	cases := []struct {
		name    string
		err     error
		missing bool
	}{
		{"removed entry", os.ErrNotExist, true},
		{"exit during read", &os.PathError{Op: "read", Path: "/proc/123/stat", Err: syscall.ESRCH}, true},
		{"unreadable entry", os.ErrPermission, false},
		{"invalid entry", ErrWorkflowConflict, false},
		{"observed entry", nil, false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if workflowProcessMissing(test.err) != test.missing {
				t.Fatal("process observation changed ownership safety")
			}
		})
	}
}
