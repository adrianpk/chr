package core

import (
	"context"
	"errors"
	"time"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// AuthoringQuestionBinding identifies the canonical current prompt, independently
// of any adapter's message/reply associations.
type AuthoringQuestionBinding struct {
	SessionID     string `json:"session_id"`
	TurnID        string `json:"turn_id"`
	QuestionID    string `json:"question_id"`
	Revision      int64  `json:"revision"`
	ContentDigest string `json:"content_digest"`
}

// AuthoringArtifactObservation binds a body and metadata image read by the owner.
type AuthoringArtifactObservation struct {
	Ref          authoringcore.AuthoringArtifactRef
	SourceHash   string
	MetadataHash string
	Body         string
}

// AuthoringSnapshot contains bounded canonical inputs and their complete image guard.
type AuthoringSnapshot struct {
	Session             authoringcore.AuthoringSession
	SessionRef          string
	SourceHash          string
	SessionHash         string
	TranscriptHash      string
	Events              []AssistedTranscriptEvent
	Artifacts           []AuthoringArtifactObservation
	Question            *AuthoringQuestionBinding
	projectMetadataHash string
	dumpRevision        string
	artifactBytes       int
	dumpAliases         []AuthoringArtifactObservation
}

// AuthoringContinuationAction is one supported shared authoring transition.
type AuthoringContinuationAction string

// Shared authoring transitions never authorize execution.
const (
	AuthoringContinuationCreate   AuthoringContinuationAction = "create"
	AuthoringContinuationAsk      AuthoringContinuationAction = "ask"
	AuthoringContinuationAnswer   AuthoringContinuationAction = "answer"
	AuthoringContinuationRespond  AuthoringContinuationAction = "respond"
	AuthoringContinuationGenerate AuthoringContinuationAction = "generate"
	AuthoringContinuationRefine   AuthoringContinuationAction = "refine"
)

// AuthoringContinuationRequest binds creation, a question, an answer, generation
// or refinement to one immutable caller-neutral request and observed image.
type AuthoringContinuationRequest struct {
	Binding             workflowstate.MutationBinding       `json:"binding"`
	ProjectID           string                              `json:"project_id"`
	Namespace           string                              `json:"namespace"`
	SessionID           string                              `json:"session_id"`
	Action              AuthoringContinuationAction         `json:"action"`
	Target              authoringcore.AuthoringArtifactKind `json:"target"`
	Topic               string                              `json:"topic"`
	Capture             string                              `json:"capture"`
	OriginalInstruction string                              `json:"original_instruction,omitempty"`
	Instructions        string                              `json:"instructions"`
	ExpectedRevision    int64                               `json:"expected_revision"`
	ExpectedSourceHash  string                              `json:"expected_source_hash"`
	Question            *AuthoringQuestionBinding           `json:"question,omitempty"`
}

// AuthoringDraft is generated data; the owner assigns identities, paths, revisions
// and provenance. A prompt and an artifact body are mutually exclusive.
type AuthoringDraft struct {
	Body      string
	Question  string
	Rationale string
}

// AuthoringGenerator runs with a bounded context and canonical snapshot. It must
// use that snapshot rather than reentering this repository's mutation owner.
type AuthoringGenerator func(context.Context, AuthoringSnapshot, AuthoringContinuationRequest) (AuthoringDraft, error)

// AuthoringCommitGuard rechecks the caller's current authority after provider
// preparation and before any new canonical effect, under the authoring owner.
type AuthoringCommitGuard func(context.Context) error

// AuthoringContinuationResult retains original confirmed facts without transcripts
// or artifact bodies. It can be reconciled after an adapter loses acknowledgement.
type AuthoringContinuationResult struct {
	RequestID      string                         `json:"request_id"`
	ArgumentDigest string                         `json:"argument_digest"`
	ReceiptRef     string                         `json:"receipt_ref"`
	Session        authoringcore.AuthoringSession `json:"session"`
	SessionRef     string                         `json:"session_ref"`
	SourceHash     string                         `json:"source_hash"`
	Question       *AuthoringQuestionBinding      `json:"question,omitempty"`
	Prompt         string                         `json:"prompt"`
}

// The shared authoring owner retains provider budgets independently of transport.
var (
	ErrAuthoringProviderWait      = errors.New("authoring provider retry is not due")
	ErrAuthoringProviderExhausted = errors.New("authoring provider attempts exhausted")
	ErrAuthoringProviderFailure   = errors.New("authoring provider continuation failed")
)

type authoringRequestBinding struct {
	Schema    string
	Digest    string
	Request   AuthoringContinuationRequest
	Attempts  int
	RetryAt   time.Time
	ErrorCode string
	CreatedAt time.Time
}

type authoringMutationReceipt struct {
	Schema   string
	Digest   string
	Result   AuthoringContinuationResult
	Checksum string
}
