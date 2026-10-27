package authoring

import (
	"time"
)

// SessionStatus defines a package boundary type.
type SessionStatus string

const (
	// SessionStatusDraft defines a package boundary constant.
	SessionStatusDraft SessionStatus = "draft"
	// SessionStatusActive defines a package boundary constant.
	SessionStatusActive SessionStatus = "active"
	// SessionStatusFinalized defines a package boundary constant.
	SessionStatusFinalized SessionStatus = "finalized"
	// SessionStatusDiscarded defines a package boundary constant.
	SessionStatusDiscarded SessionStatus = "discarded"
)

// SessionMode defines a package boundary type.
type SessionMode string

const (
	// SessionModeMarkdownDirect defines a package boundary constant.
	SessionModeMarkdownDirect SessionMode = "markdown_direct"
	// SessionModeContextual defines a package boundary constant.
	SessionModeContextual SessionMode = "contextual"
	// SessionModeAssisted defines a package boundary constant.
	SessionModeAssisted SessionMode = "assisted"
	// SessionModeTUI defines a package boundary constant.
	SessionModeTUI SessionMode = "tui"
)

// Target defines a package boundary type.
type Target string

const (
	// TargetDump defines a package boundary constant.
	TargetDump Target = "dump"
	// TargetQuiz defines a package boundary constant.
	TargetQuiz Target = "quiz"
	// TargetDraft defines a package boundary constant.
	TargetDraft Target = "draft"
	// TargetSpec defines a package boundary constant.
	TargetSpec Target = "spec"
	// TargetPlan defines a package boundary constant.
	TargetPlan Target = "plan"
	// TargetSliceSet defines a package boundary constant.
	TargetSliceSet Target = "slice_set"
	// TargetSlice defines a package boundary constant.
	TargetSlice Target = "slice"
	// TargetTracker defines a package boundary constant.
	TargetTracker Target = "tracker"
	// TargetTicket defines a package boundary constant.
	TargetTicket Target = "ticket"
)

// ArtifactKind defines a package boundary type.
type ArtifactKind string

const (
	// ArtifactKindSession defines a package boundary constant.
	ArtifactKindSession ArtifactKind = "session"
	// ArtifactKindTranscript defines a package boundary constant.
	ArtifactKindTranscript ArtifactKind = "transcript"
	// ArtifactKindDump defines a package boundary constant.
	ArtifactKindDump ArtifactKind = "dump"
	// ArtifactKindQuiz defines a package boundary constant.
	ArtifactKindQuiz ArtifactKind = "quiz"
	// ArtifactKindDraft defines a package boundary constant.
	ArtifactKindDraft ArtifactKind = "draft"
	// ArtifactKindSpec defines a package boundary constant.
	ArtifactKindSpec ArtifactKind = "spec"
	// ArtifactKindPlan defines a package boundary constant.
	ArtifactKindPlan ArtifactKind = "plan"
	// ArtifactKindSliceSet defines a package boundary constant.
	ArtifactKindSliceSet ArtifactKind = "slice_set"
	// ArtifactKindSlice defines a package boundary constant.
	ArtifactKindSlice ArtifactKind = "slice"
	// ArtifactKindTracker defines a package boundary constant.
	ArtifactKindTracker ArtifactKind = "tracker"
	// ArtifactKindTicket defines a package boundary constant.
	ArtifactKindTicket ArtifactKind = "ticket"
)

// ArtifactRole defines a package boundary type.
type ArtifactRole string

const (
	// ArtifactRoleInput defines a package boundary constant.
	ArtifactRoleInput ArtifactRole = "input"
	// ArtifactRoleOutput defines a package boundary constant.
	ArtifactRoleOutput ArtifactRole = "output"
)

// ProvenanceRelation defines a package boundary type.
type ProvenanceRelation string

const (
	// ProvenanceRelationSeededBy defines a package boundary constant.
	ProvenanceRelationSeededBy ProvenanceRelation = "seeded_by"
	// ProvenanceRelationGenerated defines a package boundary constant.
	ProvenanceRelationGenerated ProvenanceRelation = "generated"
	// ProvenanceRelationPromoted defines a package boundary constant.
	ProvenanceRelationPromoted ProvenanceRelation = "promoted"
	// ProvenanceRelationRefined defines a package boundary constant.
	ProvenanceRelationRefined ProvenanceRelation = "refined"
)

// Session describes a package boundary value.
type Session struct {
	Guided      *GuidedUseCaseState `json:",omitempty"`
	ID          string
	Namespace   string
	Revision    int64
	ProjectID   string
	Status      SessionStatus
	Mode        SessionMode
	Target      Target
	Topic       string
	Inputs      []ArtifactRef
	Outputs     []ArtifactRef
	Provenance  []ProvenanceLink
	CreatedAt   time.Time
	StartedAt   *time.Time
	FinishedAt  *time.Time
	UpdatedAt   time.Time
	Description string
}

// ArtifactRef describes a package boundary value.
type ArtifactRef struct {
	ID          string
	Kind        ArtifactKind
	Role        ArtifactRole
	Path        string
	Description string
}

// ProvenanceLink describes a package boundary value.
type ProvenanceLink struct {
	Relation   ProvenanceRelation
	Source     ArtifactRef
	Target     ArtifactRef
	Decision   string
	RecordedAt time.Time
}

// Valid reports whether the value is one of the supported constants.
func (s SessionStatus) Valid() bool {
	switch s {
	case SessionStatusDraft,
		SessionStatusActive,
		SessionStatusFinalized,
		SessionStatusDiscarded:
		return true
	default:
		return false
	}
}

// Valid reports whether the value is one of the supported constants.
func (m SessionMode) Valid() bool {
	switch m {
	case SessionModeMarkdownDirect,
		SessionModeContextual,
		SessionModeAssisted,
		SessionModeTUI:
		return true
	default:
		return false
	}
}

// Valid reports whether the value is one of the supported constants.
func (t Target) Valid() bool {
	switch t {
	case TargetDump,
		TargetQuiz,
		TargetDraft,
		TargetSpec,
		TargetPlan,
		TargetSliceSet,
		TargetSlice,
		TargetTracker,
		TargetTicket:
		return true
	default:
		return false
	}
}

// Valid reports whether the value is one of the supported constants.
func (k ArtifactKind) Valid() bool {
	switch k {
	case ArtifactKindSession,
		ArtifactKindTranscript,
		ArtifactKindDump,
		ArtifactKindQuiz,
		ArtifactKindDraft,
		ArtifactKindSpec,
		ArtifactKindPlan,
		ArtifactKindSliceSet,
		ArtifactKindSlice,
		ArtifactKindTracker,
		ArtifactKindTicket:
		return true
	default:
		return false
	}
}

// Valid reports whether the value is one of the supported constants.
func (r ArtifactRole) Valid() bool {
	switch r {
	case ArtifactRoleInput, ArtifactRoleOutput:
		return true
	default:
		return false
	}
}

// Valid reports whether the value is one of the supported constants.
func (r ProvenanceRelation) Valid() bool {
	switch r {
	case ProvenanceRelationSeededBy,
		ProvenanceRelationGenerated,
		ProvenanceRelationPromoted,
		ProvenanceRelationRefined:
		return true
	default:
		return false
	}
}
