package ticket

import (
	"time"

	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
)

// Indexed describes a package boundary value.
type Indexed struct {
	IndexedAttention
	IndexedIdentity
	IndexedSummary
	IndexedClassification
	IndexedSource
	IndexedReview
	IndexedMerge
	IndexedTimes
	IndexedProjection
}

// IndexedAttention groups the retained attention fields without changing their wire order.
type IndexedAttention struct {
	NativeWorkflow   bool
	AttentionReasons []string
}

// IndexedIdentity groups the retained identity fields without changing their wire order.
type IndexedIdentity struct {
	ProjectID    string
	Namespace    string
	Path         string
	RelativePath string
	Key          string
}

// IndexedSummary groups the retained summary fields without changing their wire order.
type IndexedSummary struct {
	Title    string
	Assignee string
	Synopsis string
}

// IndexedClassification groups the retained classification fields without changing their wire order.
type IndexedClassification struct {
	Status      Status
	BoardColumn Status
	Priority    string
	Severity    string
	Type        string
	Tags        []string
	UseCase     string
	ProjectArea string
}

// IndexedSource groups the retained source fields without changing their wire order.
type IndexedSource struct {
	SourceKind string
	SourceRef  string
	Slice      string
	Branch     string
}

// IndexedReview groups the retained review fields without changing their wire order.
type IndexedReview struct {
	ReviewRefs        []ticketmetadata.ReviewRef
	ReviewState       ReviewState
	ReviewRound       int
	ReviewResult      ReviewResult
	ReviewCommentsURL string
	BlockingFindings  []string
}

// IndexedMerge groups the retained merge fields without changing their wire order.
type IndexedMerge struct {
	MergeState   MergeState
	MergeResult  MergeResult
	MergeActor   string
	MergeCommit  string
	MergedAt     *time.Time
	MergeBlocker string
}

// IndexedTimes groups the retained times fields without changing their wire order.
type IndexedTimes struct {
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// IndexedProjection groups the retained projection fields without changing their wire order.
type IndexedProjection struct {
	FileMTime   time.Time
	FileSize    int64
	ContentHash string
	IndexedAt   time.Time
}
