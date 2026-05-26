package chc

import "time"

type IndexedTicket struct {
	ProjectID    string
	Path         string
	RelativePath string
	Key          string
	Title        string
	Synopsis     string
	Status       string
	Priority     string
	Severity     string
	Type         string
	Tags         []string
	UseCase      string
	ProjectArea  string
	SourceKind   string
	SourceRef    string
	Slice        string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
	FileMTime    time.Time
	FileSize     int64
	ContentHash  string
	IndexedAt    time.Time
}
