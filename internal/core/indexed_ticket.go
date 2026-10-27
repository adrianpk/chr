package core

import (
	"time"

	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
)

// IndexedTicket describes a package boundary value.
type IndexedTicket struct {
	IndexedTicketAttention
	IndexedTicketIdentity
	IndexedTicketSummary
	IndexedTicketClassification
	IndexedTicketSource
	IndexedTicketReview
	IndexedTicketMerge
	IndexedTicketTimes
	IndexedTicketProjection
}

// IndexedTicketAttention groups the retained attention fields without changing their wire order.
type IndexedTicketAttention struct {
	NativeWorkflow   bool
	AttentionReasons []string
}

// IndexedTicketIdentity groups the retained identity fields without changing their wire order.
type IndexedTicketIdentity struct {
	ProjectID    string
	Namespace    string
	Path         string
	RelativePath string
	Key          string
}

// IndexedTicketSummary groups the retained summary fields without changing their wire order.
type IndexedTicketSummary struct {
	Title    string
	Assignee string
	Synopsis string
}

// IndexedTicketClassification groups the retained classification fields without changing their wire order.
type IndexedTicketClassification struct {
	Status      string
	BoardColumn string
	Priority    string
	Severity    string
	Type        string
	Tags        []string
	UseCase     string
	ProjectArea string
}

// IndexedTicketSource groups the retained source fields without changing their wire order.
type IndexedTicketSource struct {
	SourceKind string
	SourceRef  string
	Slice      string
	Branch     string
}

// IndexedTicketReview groups the retained review fields without changing their wire order.
type IndexedTicketReview struct {
	ReviewRefs        []ticketmetadata.ReviewRef
	ReviewState       string
	ReviewRound       int
	ReviewResult      string
	ReviewCommentsURL string
	BlockingFindings  []string
}

// IndexedTicketMerge groups the retained merge fields without changing their wire order.
type IndexedTicketMerge struct {
	MergeState   string
	MergeResult  string
	MergeActor   string
	MergeCommit  string
	MergedAt     *time.Time
	MergeBlocker string
}

// IndexedTicketTimes groups the retained times fields without changing their wire order.
type IndexedTicketTimes struct {
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// IndexedTicketProjection groups the retained projection fields without changing their wire order.
type IndexedTicketProjection struct {
	FileMTime   time.Time
	FileSize    int64
	ContentHash string
	IndexedAt   time.Time
}
