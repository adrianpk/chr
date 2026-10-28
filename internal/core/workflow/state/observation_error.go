package state

import (
	"errors"
)

// ErrObservationRequired retains target exclusion until a submitted
// operation is verified delivered or definitively sealed without publication.
var ErrObservationRequired = errors.New("workflow submission requires canonical observation")
