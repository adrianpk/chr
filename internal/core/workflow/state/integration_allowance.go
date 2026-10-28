package state

import (
	"fmt"
	"math"
)

// ValidTokenAllowance checks immutable allowance fields without reserving or admitting a call.
func ValidTokenAllowance(a TokenAllowance) bool {
	maximum, err := TokenSum(a.Input, a.Output)

	return err == nil && maximum > 0 && a.Maximum == maximum && a.WholeCall && ValidFingerprint(a.CapabilityRevision) && !a.Deadline.IsZero()
}

// TokenSum adds nonnegative retained token/counter values with overflow rejection.
func TokenSum(a, b int64) (int64, error) {
	if a < 0 || b < 0 || a > math.MaxInt64-b {
		return 0, IntegrationConflict("token or counter arithmetic overflow")
	}

	return a + b, nil
}

// IntegrationConflict preserves the integration-entry conflict identity and message context.
func IntegrationConflict(reason string) error {
	return fmt.Errorf("%w: integration entry %s", ErrConflict, reason)
}
