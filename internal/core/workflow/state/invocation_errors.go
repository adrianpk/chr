package state

import (
	"errors"
)

// ErrOwned identifies overlap with retained task or hierarchy ownership.
var ErrOwned = errors.New("workflow work is already owned")

// ErrStop reports a durable pause or cancellation request.
var ErrStop = errors.New("workflow stop requested")

// InvocationTerminal reports the two retained terminal invocation states.
func InvocationTerminal(status InvocationStatus) bool {
	return status == InvocationCompleted || status == InvocationCancelled
}
