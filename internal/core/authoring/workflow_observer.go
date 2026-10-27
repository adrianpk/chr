package authoring

import (
	"context"
	"time"
)

// GuidedWorkflowObserver reads verified facts from the existing native owner.
// It never admits work, changes policy, dispatches or splits a durable mutation.
type GuidedWorkflowObserver interface {
	ObserveStop(context.Context, Session, string) (GuidedStopObservation, error)
	FindAdmission(context.Context, Session) (GuidedAdmissionObservation, error)
	InspectResult(context.Context, Session) (GuidedResultObservation, error)
}

// GuidedStopObservation separates accepted stop intent from native quiescence.
type GuidedStopObservation struct {
	Status       string
	Revision     uint64
	Acknowledged bool
}

// GuidedAdmissionObservation distinguishes original admission from normal absence.
type GuidedAdmissionObservation struct {
	Invocation GuidedInvocationReference
	Found      bool
}

// GuidedResultObservation retains original verified references and native update time.
type GuidedResultObservation struct {
	Reference GuidedResultReference
	UpdatedAt time.Time
	Complete  bool
}
