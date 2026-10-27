package config

import "fmt"

// WorkStateAuthority selects the sole state owner for a work entrypoint.
type WorkStateAuthority string

// Explicit work authorities never grant two owners to the same task.
const (
	WorkStateAuthorityTickets       WorkStateAuthority = "tickets"
	WorkStateAuthorityTrackerCompat WorkStateAuthority = "tracker_compat"
)

// WorkAuthority preserves legacy defaults until versioned activation is delivered.
func WorkAuthority(config WorkConfig) WorkStateAuthority {
	if config.StateAuthority == "" {
		return WorkStateAuthorityTrackerCompat
	}

	return config.StateAuthority
}

// ValidateWorkStateAuthority rejects ambiguous or unsupported authority choices.
func ValidateWorkStateAuthority(authority WorkStateAuthority) error {
	if authority == "" || authority == WorkStateAuthorityTickets || authority == WorkStateAuthorityTrackerCompat {
		return nil
	}

	return fmt.Errorf("invalid work state_authority: %s", authority)
}
