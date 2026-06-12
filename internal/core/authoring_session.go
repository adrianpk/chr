package core

import (
	"fmt"
	"time"
)

// AuthoringSessionStatus defines a package boundary type.
type AuthoringSessionStatus string

const (
	// AuthoringSessionStatusDraft defines a package boundary constant.
	AuthoringSessionStatusDraft AuthoringSessionStatus = "draft"
	// AuthoringSessionStatusActive defines a package boundary constant.
	AuthoringSessionStatusActive AuthoringSessionStatus = "active"
	// AuthoringSessionStatusFinalized defines a package boundary constant.
	AuthoringSessionStatusFinalized AuthoringSessionStatus = "finalized"
	// AuthoringSessionStatusDiscarded defines a package boundary constant.
	AuthoringSessionStatusDiscarded AuthoringSessionStatus = "discarded"
)

// AuthoringSessionMode defines a package boundary type.
type AuthoringSessionMode string

const (
	// AuthoringSessionModeMarkdownDirect defines a package boundary constant.
	AuthoringSessionModeMarkdownDirect AuthoringSessionMode = "markdown_direct"
	// AuthoringSessionModeContextual defines a package boundary constant.
	AuthoringSessionModeContextual AuthoringSessionMode = "contextual"
	// AuthoringSessionModeAssisted defines a package boundary constant.
	AuthoringSessionModeAssisted AuthoringSessionMode = "assisted"
	// AuthoringSessionModeTUI defines a package boundary constant.
	AuthoringSessionModeTUI AuthoringSessionMode = "tui"
)

// AuthoringTarget defines a package boundary type.
type AuthoringTarget string

const (
	// AuthoringTargetDump defines a package boundary constant.
	AuthoringTargetDump AuthoringTarget = "dump"
	// AuthoringTargetQuiz defines a package boundary constant.
	AuthoringTargetQuiz AuthoringTarget = "quiz"
	// AuthoringTargetDraft defines a package boundary constant.
	AuthoringTargetDraft AuthoringTarget = "draft"
	// AuthoringTargetSpec defines a package boundary constant.
	AuthoringTargetSpec AuthoringTarget = "spec"
	// AuthoringTargetPlan defines a package boundary constant.
	AuthoringTargetPlan AuthoringTarget = "plan"
	// AuthoringTargetSliceSet defines a package boundary constant.
	AuthoringTargetSliceSet AuthoringTarget = "slice_set"
	// AuthoringTargetSlice defines a package boundary constant.
	AuthoringTargetSlice AuthoringTarget = "slice"
	// AuthoringTargetTracker defines a package boundary constant.
	AuthoringTargetTracker AuthoringTarget = "tracker"
	// AuthoringTargetTicket defines a package boundary constant.
	AuthoringTargetTicket AuthoringTarget = "ticket"
)

// AuthoringArtifactKind defines a package boundary type.
type AuthoringArtifactKind string

const (
	// AuthoringArtifactKindSession defines a package boundary constant.
	AuthoringArtifactKindSession AuthoringArtifactKind = "session"
	// AuthoringArtifactKindTranscript defines a package boundary constant.
	AuthoringArtifactKindTranscript AuthoringArtifactKind = "transcript"
	// AuthoringArtifactKindDump defines a package boundary constant.
	AuthoringArtifactKindDump AuthoringArtifactKind = "dump"
	// AuthoringArtifactKindQuiz defines a package boundary constant.
	AuthoringArtifactKindQuiz AuthoringArtifactKind = "quiz"
	// AuthoringArtifactKindDraft defines a package boundary constant.
	AuthoringArtifactKindDraft AuthoringArtifactKind = "draft"
	// AuthoringArtifactKindSpec defines a package boundary constant.
	AuthoringArtifactKindSpec AuthoringArtifactKind = "spec"
	// AuthoringArtifactKindPlan defines a package boundary constant.
	AuthoringArtifactKindPlan AuthoringArtifactKind = "plan"
	// AuthoringArtifactKindSliceSet defines a package boundary constant.
	AuthoringArtifactKindSliceSet AuthoringArtifactKind = "slice_set"
	// AuthoringArtifactKindSlice defines a package boundary constant.
	AuthoringArtifactKindSlice AuthoringArtifactKind = "slice"
	// AuthoringArtifactKindTracker defines a package boundary constant.
	AuthoringArtifactKindTracker AuthoringArtifactKind = "tracker"
	// AuthoringArtifactKindTicket defines a package boundary constant.
	AuthoringArtifactKindTicket AuthoringArtifactKind = "ticket"
)

// AuthoringArtifactRole defines a package boundary type.
type AuthoringArtifactRole string

const (
	// AuthoringArtifactRoleInput defines a package boundary constant.
	AuthoringArtifactRoleInput AuthoringArtifactRole = "input"
	// AuthoringArtifactRoleOutput defines a package boundary constant.
	AuthoringArtifactRoleOutput AuthoringArtifactRole = "output"
)

// AuthoringProvenanceRelation defines a package boundary type.
type AuthoringProvenanceRelation string

const (
	// AuthoringProvenanceRelationSeededBy defines a package boundary constant.
	AuthoringProvenanceRelationSeededBy AuthoringProvenanceRelation = "seeded_by"
	// AuthoringProvenanceRelationGenerated defines a package boundary constant.
	AuthoringProvenanceRelationGenerated AuthoringProvenanceRelation = "generated"
	// AuthoringProvenanceRelationPromoted defines a package boundary constant.
	AuthoringProvenanceRelationPromoted AuthoringProvenanceRelation = "promoted"
	// AuthoringProvenanceRelationRefined defines a package boundary constant.
	AuthoringProvenanceRelationRefined AuthoringProvenanceRelation = "refined"
)

// AuthoringSession describes a package boundary value.
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

// AuthoringArtifactRef describes a package boundary value.
type AuthoringArtifactRef struct {
	ID          string
	Kind        AuthoringArtifactKind
	Role        AuthoringArtifactRole
	Path        string
	Description string
}

// AuthoringProvenanceLink describes a package boundary value.
type AuthoringProvenanceLink struct {
	Relation   AuthoringProvenanceRelation
	Source     AuthoringArtifactRef
	Target     AuthoringArtifactRef
	Decision   string
	RecordedAt time.Time
}

// ValidateAuthoringSession validates the corresponding package boundary value.
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

// Valid reports whether the value is one of the supported constants.
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

// Valid reports whether the value is one of the supported constants.
func (m AuthoringSessionMode) Valid() bool {
	switch m {
	case AuthoringSessionModeMarkdownDirect,
		AuthoringSessionModeContextual,
		AuthoringSessionModeAssisted,
		AuthoringSessionModeTUI:
		return true
	default:
		return false
	}
}

// Valid reports whether the value is one of the supported constants.
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

// Valid reports whether the value is one of the supported constants.
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

// Valid reports whether the value is one of the supported constants.
func (r AuthoringArtifactRole) Valid() bool {
	switch r {
	case AuthoringArtifactRoleInput,
		AuthoringArtifactRoleOutput:
		return true
	default:
		return false
	}
}

// Valid reports whether the value is one of the supported constants.
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
