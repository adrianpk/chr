package app

import (
	"errors"

	"forge.adrianpk.com/chr/internal/core"
)

// RecordKind identifies the canonical owner view selected by a caller.
type RecordKind string

// Record kinds share one operation boundary and preserve their native owners.
const (
	RecordTicket   RecordKind = "ticket"
	RecordSlice    RecordKind = "slice"
	RecordSliceSet RecordKind = "slice-set"
	RecordUseCase  RecordKind = "use-case"
)

// ErrRecordUnsupported requires an operation supported by the actual record owner.
var ErrRecordUnsupported = errors.New("record operation is unsupported by its lifecycle owner")

// RecordQuery selects one identity or a bounded page in a frozen namespace.
type RecordQuery struct {
	Kind     RecordKind `json:"kind"`
	Identity string     `json:"identity,omitempty"`
	ParentID string     `json:"parent_id,omitempty"`
	Offset   int        `json:"offset,omitempty"`
	Limit    int        `json:"limit,omitempty"`
}

// RecordView returns observed guards and canonical references without changing lifecycle.
type RecordView struct {
	Kind       RecordKind `json:"kind"`
	Identity   string     `json:"identity"`
	ParentID   string     `json:"parent_id,omitempty"`
	Title      string     `json:"title"`
	Phase      string     `json:"phase,omitempty"`
	Revision   uint64     `json:"revision,omitempty"`
	SourceHash string     `json:"source_hash"`
	Reference  string     `json:"reference"`
	Body       string     `json:"body,omitempty"`
}

// RecordPage bounds display output and provides a deterministic next offset.
type RecordPage struct {
	Records    []RecordView `json:"records"`
	NextOffset int          `json:"next_offset"`
	HasMore    bool         `json:"has_more"`
}

// RecordMutationAction is an explicit metadata mutation, separate from lifecycle controls.
type RecordMutationAction string

// Metadata mutation actions are validated before preparing owner input.
const (
	RecordCreate RecordMutationAction = "create"
	RecordUpdate RecordMutationAction = "update"
)

// RecordMutationRequest exposes metadata changes and explicit proposed-record creation.
// Lifecycle status, execution, review and grants require their dedicated owners.
type RecordMutationRequest struct {
	Kind             RecordKind               `json:"kind"`
	Action           RecordMutationAction     `json:"action"`
	Identity         string                   `json:"identity,omitempty"`
	ParentID         string                   `json:"parent_id,omitempty"`
	ExpectedRevision uint64                   `json:"revision,omitempty"`
	ExpectedHash     string                   `json:"source_hash,omitempty"`
	Title            *string                  `json:"title,omitempty"`
	Synopsis         *string                  `json:"synopsis,omitempty"`
	Body             *string                  `json:"body,omitempty"`
	Context          *string                  `json:"context,omitempty"`
	UseCase          string                   `json:"use_case,omitempty"`
	PlanRef          string                   `json:"plan_ref,omitempty"`
	PlanHash         string                   `json:"plan_hash,omitempty"`
	Validation       *core.WorkflowValidation `json:"validation,omitempty"`
	Slice            *core.WorkflowSlice      `json:"slice,omitempty"`
}

// PreparedRecordMutation freezes the exact owner input before any canonical effect.
type PreparedRecordMutation struct {
	ProjectID string                        `json:"project_id"`
	Namespace string                        `json:"namespace"`
	Kind      RecordKind                    `json:"kind"`
	Identity  string                        `json:"identity,omitempty"`
	ParentID  string                        `json:"parent_id,omitempty"`
	Native    *core.WorkflowMutationRequest `json:"native,omitempty"`
	Case      *core.CaseMutationRequest     `json:"case,omitempty"`
}

// RecordMutationResult returns only confirmed owner facts and its replay receipt.
type RecordMutationResult struct {
	Record         RecordView `json:"record"`
	ReceiptRef     string     `json:"receipt_ref"`
	ArgumentDigest string     `json:"argument_digest"`
}
