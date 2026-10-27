package authoring

import (
	"context"
	"errors"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// QuestionBinding identifies the canonical current prompt, independently
// of any adapter's message/reply associations.
type QuestionBinding struct {
	SessionID     string `json:"session_id"`
	TurnID        string `json:"turn_id"`
	QuestionID    string `json:"question_id"`
	Revision      int64  `json:"revision"`
	ContentDigest string `json:"content_digest"`
}

// ArtifactObservation binds a body and metadata image read by the owner.
type ArtifactObservation struct {
	Ref          ArtifactRef
	SourceHash   string
	MetadataHash string
	Body         string
}

// Snapshot contains bounded canonical inputs and their complete image guard.
type Snapshot struct {
	Session             Session
	SessionRef          string
	SourceHash          string
	SessionHash         string
	TranscriptHash      string
	Events              []AssistedTranscriptEvent
	Artifacts           []ArtifactObservation
	Question            *QuestionBinding
	projectMetadataHash string
	dumpRevision        string
	artifactBytes       int
	dumpAliases         []ArtifactObservation
}

// ContinuationAction is one supported shared authoring transition.
type ContinuationAction string

// Shared authoring transitions never authorize execution.
const (
	ContinuationCreate   ContinuationAction = "create"
	ContinuationAsk      ContinuationAction = "ask"
	ContinuationAnswer   ContinuationAction = "answer"
	ContinuationRespond  ContinuationAction = "respond"
	ContinuationGenerate ContinuationAction = "generate"
	ContinuationRefine   ContinuationAction = "refine"
)

// ContinuationRequest binds creation, a question, an answer, generation
// or refinement to one immutable caller-neutral request and observed image.
type ContinuationRequest struct {
	Binding             workflowstate.MutationBinding `json:"binding"`
	ProjectID           string                        `json:"project_id"`
	Namespace           string                        `json:"namespace"`
	SessionID           string                        `json:"session_id"`
	Action              ContinuationAction            `json:"action"`
	Target              ArtifactKind                  `json:"target"`
	Topic               string                        `json:"topic"`
	Capture             string                        `json:"capture"`
	OriginalInstruction string                        `json:"original_instruction,omitempty"`
	Instructions        string                        `json:"instructions"`
	ExpectedRevision    int64                         `json:"expected_revision"`
	ExpectedSourceHash  string                        `json:"expected_source_hash"`
	Question            *QuestionBinding              `json:"question,omitempty"`
}

// Draft is generated data; the owner assigns identities, paths, revisions
// and provenance. A prompt and an artifact body are mutually exclusive.
type Draft struct {
	Body      string
	Question  string
	Rationale string
}

// Generator runs with a bounded context and canonical snapshot. It must
// use that snapshot rather than reentering this repository's mutation owner.
type Generator func(context.Context, Snapshot, ContinuationRequest) (Draft, error)

// CommitGuard rechecks the caller's current authority after provider
// preparation and before any new canonical effect, under the authoring owner.
type CommitGuard func(context.Context) error

// ContinuationResult retains original confirmed facts without transcripts
// or artifact bodies. It can be reconciled after an adapter loses acknowledgement.
type ContinuationResult struct {
	RequestID      string           `json:"request_id"`
	ArgumentDigest string           `json:"argument_digest"`
	ReceiptRef     string           `json:"receipt_ref"`
	Session        Session          `json:"session"`
	SessionRef     string           `json:"session_ref"`
	SourceHash     string           `json:"source_hash"`
	Question       *QuestionBinding `json:"question,omitempty"`
	Prompt         string           `json:"prompt"`
}

// The shared authoring owner retains provider budgets independently of transport.
var (
	ErrProviderWait      = errors.New("authoring provider retry is not due")
	ErrProviderExhausted = errors.New("authoring provider attempts exhausted")
	ErrProviderFailure   = errors.New("authoring provider continuation failed")
)

type authoringRequestBinding struct {
	Schema    string
	Digest    string
	Request   ContinuationRequest
	Attempts  int
	RetryAt   time.Time
	ErrorCode string
	CreatedAt time.Time
}

type authoringMutationReceipt struct {
	Schema   string
	Digest   string
	Result   ContinuationResult
	Checksum string
}
