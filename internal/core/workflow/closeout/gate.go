// Package closeout owns closing policy's store-facing contracts.
package closeout

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Gate executes the host gate for one bound closing claim and exact input.
// Its receipt remains unconfirmed until the owner rechecks and commits closure.
type Gate interface {
	Execute(context.Context) (workflowstate.ClosingReceipt, error)
}
