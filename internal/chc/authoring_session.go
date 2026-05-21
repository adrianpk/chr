package chc

import (
	"fmt"
	"time"
)

type AuthoringSessionStatus string

const (
	AuthoringSessionStatusDraft     AuthoringSessionStatus = "draft"
	AuthoringSessionStatusActive    AuthoringSessionStatus = "active"
	AuthoringSessionStatusFinalized AuthoringSessionStatus = "finalized"
	AuthoringSessionStatusDiscarded AuthoringSessionStatus = "discarded"
)

type AuthoringSessionMode string

const (
	AuthoringSessionModeMarkdownDirect AuthoringSessionMode = "markdown_direct"
	AuthoringSessionModeStatic         AuthoringSessionMode = "static"
	AuthoringSessionModeAssisted       AuthoringSessionMode = "assisted"
	AuthoringSessionModeTUI            AuthoringSessionMode = "tui"
)

type AuthoringTarget string

const (
	AuthoringTargetDump     AuthoringTarget = "dump"
	AuthoringTargetQuiz     AuthoringTarget = "quiz"
	AuthoringTargetDraft    AuthoringTarget = "draft"
	AuthoringTargetSpec     AuthoringTarget = "spec"
	AuthoringTargetPlan     AuthoringTarget = "plan"
	AuthoringTargetSliceSet AuthoringTarget = "slice_set"
	AuthoringTargetSlice    AuthoringTarget = "slice"
	AuthoringTargetTracker  AuthoringTarget = "tracker"
	AuthoringTargetTicket   AuthoringTarget = "ticket"
)

type AuthoringArtifactKind string

const (
	AuthoringArtifactKindSession    AuthoringArtifactKind = "session"
	AuthoringArtifactKindTranscript AuthoringArtifactKind = "transcript"
	AuthoringArtifactKindDump       AuthoringArtifactKind = "dump"
	AuthoringArtifactKindQuiz       AuthoringArtifactKind = "quiz"
	AuthoringArtifactKindDraft      AuthoringArtifactKind = "draft"
	AuthoringArtifactKindSpec       AuthoringArtifactKind = "spec"
	AuthoringArtifactKindPlan       AuthoringArtifactKind = "plan"
	AuthoringArtifactKindSliceSet   AuthoringArtifactKind = "slice_set"
	AuthoringArtifactKindSlice      AuthoringArtifactKind = "slice"
	AuthoringArtifactKindTracker    AuthoringArtifactKind = "tracker"
	AuthoringArtifactKindTicket     AuthoringArtifactKind = "ticket"
)

type AuthoringArtifactRole string

const (
	AuthoringArtifactRoleInput  AuthoringArtifactRole = "input"
	AuthoringArtifactRoleOutput AuthoringArtifactRole = "output"
)

type AuthoringProvenanceRelation string

const (
	AuthoringProvenanceRelationSeededBy  AuthoringProvenanceRelation = "seeded_by"
	AuthoringProvenanceRelationGenerated AuthoringProvenanceRelation = "generated"
	AuthoringProvenanceRelationPromoted  AuthoringProvenanceRelation = "promoted"
	AuthoringProvenanceRelationRefined   AuthoringProvenanceRelation = "refined"
)

type AuthoringSession struct {
	ID          string
	ProjectID   string
	Status      AuthoringSessionStatus
	Mode        AuthoringSessionMode
	Target      AuthoringTarget
	Topic       string
	Inputs      []AuthoringArtifactRef
	Outputs     []AuthoringArtifactRef
	Provenance  []AuthoringProvenanceLink
	CreatedAt   time.Time
	StartedAt   *time.Time
	FinishedAt  *time.Time
	UpdatedAt   time.Time
	Description string
}

type AuthoringArtifactRef struct {
	ID          string
	Kind        AuthoringArtifactKind
	Role        AuthoringArtifactRole
	Path        string
	Description string
}

type AuthoringProvenanceLink struct {
	Relation   AuthoringProvenanceRelation
	Source     AuthoringArtifactRef
	Target     AuthoringArtifactRef
	Decision   string
	RecordedAt time.Time
}

func ValidateAuthoringSession(session AuthoringSession) error {
	if session.ID == "" {
		return fmt.Errorf("authoring session id is required")
	}

	if session.ProjectID == "" {
		return fmt.Errorf("authoring session project id is required")
	}

	if !session.Status.Valid() {
		return fmt.Errorf("authoring session status is invalid: %s", session.Status)
	}

	if !session.Mode.Valid() {
		return fmt.Errorf("authoring session mode is invalid: %s", session.Mode)
	}

	if !session.Target.Valid() {
		return fmt.Errorf("authoring session target is invalid: %s", session.Target)
	}

	if session.Topic == "" {
		return fmt.Errorf("authoring session topic is required")
	}

	err := validateAuthoringArtifactRefs("input", session.Inputs, AuthoringArtifactRoleInput)
	if err != nil {
		return err
	}

	err = validateAuthoringArtifactRefs("output", session.Outputs, AuthoringArtifactRoleOutput)
	if err != nil {
		return err
	}

	err = validateAuthoringProvenanceLinks(session.Provenance)
	if err != nil {
		return err
	}

	return validateAuthoringSessionTimes(session)
}

func (s AuthoringSessionStatus) Valid() bool {
	switch s {
	case AuthoringSessionStatusDraft,
		AuthoringSessionStatusActive,
		AuthoringSessionStatusFinalized,
		AuthoringSessionStatusDiscarded:
		return true
	default:
		return false
	}
}

func (m AuthoringSessionMode) Valid() bool {
	switch m {
	case AuthoringSessionModeMarkdownDirect,
		AuthoringSessionModeStatic,
		AuthoringSessionModeAssisted,
		AuthoringSessionModeTUI:
		return true
	default:
		return false
	}
}

func (t AuthoringTarget) Valid() bool {
	switch t {
	case AuthoringTargetDump,
		AuthoringTargetQuiz,
		AuthoringTargetDraft,
		AuthoringTargetSpec,
		AuthoringTargetPlan,
		AuthoringTargetSliceSet,
		AuthoringTargetSlice,
		AuthoringTargetTracker,
		AuthoringTargetTicket:
		return true
	default:
		return false
	}
}

func (k AuthoringArtifactKind) Valid() bool {
	switch k {
	case AuthoringArtifactKindSession,
		AuthoringArtifactKindTranscript,
		AuthoringArtifactKindDump,
		AuthoringArtifactKindQuiz,
		AuthoringArtifactKindDraft,
		AuthoringArtifactKindSpec,
		AuthoringArtifactKindPlan,
		AuthoringArtifactKindSliceSet,
		AuthoringArtifactKindSlice,
		AuthoringArtifactKindTracker,
		AuthoringArtifactKindTicket:
		return true
	default:
		return false
	}
}

func (r AuthoringArtifactRole) Valid() bool {
	switch r {
	case AuthoringArtifactRoleInput,
		AuthoringArtifactRoleOutput:
		return true
	default:
		return false
	}
}

func (r AuthoringProvenanceRelation) Valid() bool {
	switch r {
	case AuthoringProvenanceRelationSeededBy,
		AuthoringProvenanceRelationGenerated,
		AuthoringProvenanceRelationPromoted,
		AuthoringProvenanceRelationRefined:
		return true
	default:
		return false
	}
}

func validateAuthoringArtifactRefs(name string, refs []AuthoringArtifactRef, expectedRole AuthoringArtifactRole) error {
	for index, ref := range refs {
		err := validateAuthoringArtifactRef(ref)
		if err != nil {
			return fmt.Errorf("authoring %s artifact %d: %w", name, index+1, err)
		}

		if ref.Role != expectedRole {
			return fmt.Errorf("authoring %s artifact %d role must be %s", name, index+1, expectedRole)
		}
	}

	return nil
}

func validateAuthoringArtifactRef(ref AuthoringArtifactRef) error {
	if ref.ID == "" {
		return fmt.Errorf("artifact id is required")
	}

	if !ref.Kind.Valid() {
		return fmt.Errorf("artifact kind is invalid: %s", ref.Kind)
	}

	if !ref.Role.Valid() {
		return fmt.Errorf("artifact role is invalid: %s", ref.Role)
	}

	if ref.Path == "" {
		return fmt.Errorf("artifact path is required")
	}

	return nil
}

func validateAuthoringProvenanceLinks(links []AuthoringProvenanceLink) error {
	for index, link := range links {
		if !link.Relation.Valid() {
			return fmt.Errorf("authoring provenance link %d relation is invalid: %s", index+1, link.Relation)
		}

		err := validateAuthoringArtifactRef(link.Source)
		if err != nil {
			return fmt.Errorf("authoring provenance link %d source: %w", index+1, err)
		}

		err = validateAuthoringArtifactRef(link.Target)
		if err != nil {
			return fmt.Errorf("authoring provenance link %d target: %w", index+1, err)
		}

		if link.RecordedAt.IsZero() {
			return fmt.Errorf("authoring provenance link %d recorded time is required", index+1)
		}
	}

	return nil
}

func validateAuthoringSessionTimes(session AuthoringSession) error {
	if session.CreatedAt.IsZero() {
		return fmt.Errorf("authoring session created time is required")
	}

	if session.UpdatedAt.IsZero() {
		return fmt.Errorf("authoring session updated time is required")
	}

	if session.StartedAt != nil && session.FinishedAt != nil && session.FinishedAt.Before(*session.StartedAt) {
		return fmt.Errorf("authoring session finished before it started")
	}

	return nil
}
