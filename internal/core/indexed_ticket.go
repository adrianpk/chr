package core

import "time"

// IndexedTicket describes a package boundary value.
type IndexedTicket struct {
	ProjectID         string
	Namespace         string
	Path              string
	RelativePath      string
	Key               string
	Title             string
	Synopsis          string
	Status            string
	BoardColumn       string
	Priority          string
	Severity          string
	Type              string
	Tags              []string
	UseCase           string
	ProjectArea       string
	SourceKind        string
	SourceRef         string
	Slice             string
	Branch            string
	ReviewRefs        []TicketReviewRef
	ReviewState       string
	ReviewRound       int
	ReviewResult      string
	ReviewCommentsURL string
	BlockingFindings  []string
	MergeState        string
	MergeResult       string
	MergeActor        string
	MergeCommit       string
	MergedAt          *time.Time
	MergeBlocker      string
	CreatedAt         *time.Time
	UpdatedAt         *time.Time
	FileMTime         time.Time
	FileSize          int64
	ContentHash       string
	IndexedAt         time.Time
}
