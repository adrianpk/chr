package generation

import (
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// PlanPublicationOperation binds retries to caller-owned intent, independently
// of model output, timestamps, or the newly observed target after a commit.
type PlanPublicationOperation struct {
	TicketDelivery     bool `json:",omitempty"`
	ID                 string
	Scope              quizcore.Scope
	ExpectedRevision   string
	Replace            bool
	TrackerReplaceHash string
}
