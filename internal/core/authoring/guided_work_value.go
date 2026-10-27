package authoring

import (
	"time"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// GuidedPreparedPlan retains validated generation before revision-bound publication.
type GuidedPreparedPlan struct {
	Plan           generationcore.GeneratedPlan `json:"plan"`
	TargetRevision string                       `json:"target_revision"`
	InputDigest    string                       `json:"input_digest"`
}

// GuidedWorkState links the canonical plan, materialization and native invocation owners.
type GuidedWorkState struct {
	Admission            *workflowstate.StartInput               `json:"admission,omitempty"`
	Invocation           *GuidedInvocationReference              `json:"invocation,omitempty"`
	PublicationOperation generationcore.PlanPublicationOperation `json:"publication_operation"`
	Publication          *PlanPublicationResult                  `json:"publication,omitempty"`
	Materialization      *GuidedMaterializationReference         `json:"materialization,omitempty"`
}

// GuidedWorkMember identifies one exact initial native task contract.
type GuidedWorkMember struct {
	Key, Path, ContractHash, SourceHash string
	Revision                            uint64
}

// GuidedMaterializationReference links the original receipt and exact native membership.
type GuidedMaterializationReference struct {
	ID, ReceiptRef, SetID, SetPath string
	PlanRevision                   generationcore.PlanRevisionBinding
	Members                        []GuidedWorkMember
}

// GuidedInvocationReference addresses the native owner; its mutable state stays there.
type GuidedInvocationReference struct {
	ID, RequestKey string
	Revision       uint64
}

// GuidedResultReference retains only native completion references, not new receipts.
type GuidedResultReference struct {
	InvocationID string
	Revision     uint64
	Receipts     []workflowstate.EvidenceRef
	Closures     []workflowstate.ClosingEvidence
}

// GuidedIdeaReference locates a separately retained request pending later review.
type GuidedIdeaReference struct {
	ID, Digest, Path, Hash, Actor string
	CreatedAt                     time.Time
}

// GuidedControlState keeps stop intent separate from an interrupted stage.
// Native receipts prove acceptance; native quiescence is acknowledged separately.
type GuidedControlState struct {
	StopIntent           string                                `json:"stop_intent,omitempty"`
	ResumeStatus         string                                `json:"resume_status,omitempty"`
	PreviousStatus       string                                `json:"previous_status,omitempty"`
	PreviousIntent       string                                `json:"previous_intent,omitempty"`
	PreviousBlocker      string                                `json:"previous_blocker,omitempty"`
	AcknowledgedRevision uint64                                `json:"acknowledged_revision,omitempty"`
	PendingOperation     *GuidedUseCaseOperation               `json:"pending_operation,omitempty"`
	NativeRequest        *workflowstate.ControlMutationRequest `json:"native_request,omitempty"`
	NativeReceipt        *workflowstate.ControlResult          `json:"native_receipt,omitempty"`
	CompletedOperations  []GuidedControlReceipt                `json:"completed_operations,omitempty"`
}

// GuidedControlReceipt retains the original control outcome across later changes.
type GuidedControlReceipt struct {
	GuidedUseCaseOperation
	Native   *workflowstate.ControlResult `json:"native,omitempty"`
	Rejected bool                         `json:"rejected,omitempty"`
}
