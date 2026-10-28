package state

import (
	"strconv"
)

// ProcessState names the existing supervised-command lifecycle.
type ProcessState string

// Process states retain the running/stopped journal protocol.
const (
	ProcessRunning ProcessState = "running"
	ProcessStopped ProcessState = "stopped"
)

// Process retains each supervised worker or validation command.
type Process struct {
	CapacitySlot int             `json:"capacity_slot,omitempty"`
	Identity     ProcessIdentity `json:"identity"`
	GroupID      int             `json:"group_id"`
	Command      []string        `json:"command"`
	State        ProcessState    `json:"state"`
	StdoutRef    string          `json:"stdout_ref"`
	StderrRef    string          `json:"stderr_ref"`
	StdoutHash   string          `json:"stdout_hash"`
	StderrHash   string          `json:"stderr_hash"`
	Result       *CommandResult  `json:"result"`
}

// ValidProcessIdentity checks retained identity shape; it never observes the live process.
func ValidProcessIdentity(identity ProcessIdentity) bool {
	birth, err := strconv.ParseUint(identity.StartTicks, 10, 64)

	return identity.PID > 0 && len(identity.BootID) == 36 && err == nil && birth > 0
}

// WorkerSlotLimit bounds the common worker pool and retained process slots.
const WorkerSlotLimit = 8
