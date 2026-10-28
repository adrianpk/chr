package state

import (
	"time"
)

// TokenAllowance describes whole-call accounting, including input.
// Reported output already includes reasoning; cached input remains input.
type TokenAllowance struct {
	Maximum            int64     `json:"maximum"`
	Input              int64     `json:"input"`
	Output             int64     `json:"output"`
	CapabilityRevision string    `json:"capability_revision"`
	WholeCall          bool      `json:"whole_call"`
	Deadline           time.Time `json:"deadline"`
}
