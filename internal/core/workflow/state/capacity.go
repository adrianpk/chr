package state

import (
	"time"
)

// RoleReservation exposes live owners without their claim credentials.
type RoleReservation struct {
	InvocationID string
	Namespace    string
	AttemptID    string
	Role         string
	LeaseUntil   time.Time
	Owner        ProcessIdentity
}

// Capacity is an observation, not permission to launch work.
// All roles consume the same repository slots, including live expired groups.
type Capacity struct {
	Maximum         int
	Available       int
	LiveProcesses   []Process
	UnexpiredOwners []RoleReservation
}
