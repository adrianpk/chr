package chc

import "time"

const (
	DefaultQuizQuestionBudget = 50
	MaxQuizFollowUpDepth      = 3
)

type QuizScope struct {
	ProjectID string
	CaseSlug  string
}

type QuizRuntimeStatus string

const (
	QuizRuntimeStatusIdle      QuizRuntimeStatus = "idle"
	QuizRuntimeStatusActive    QuizRuntimeStatus = "active"
	QuizRuntimeStatusDone      QuizRuntimeStatus = "done"
	QuizRuntimeStatusCancelled QuizRuntimeStatus = "cancelled"
	QuizRuntimeStatusBlocked   QuizRuntimeStatus = "blocked"
)

type QuizMode string

const (
	QuizModeDeep QuizMode = "deep"
	QuizModeFast QuizMode = "fast"
)

type QuizEntryStatus string

const (
	QuizEntryStatusAnswered QuizEntryStatus = "answered"
	QuizEntryStatusSkipped  QuizEntryStatus = "skipped"
)

type QuizCoverageStatus string

const (
	QuizCoverageStatusPending  QuizCoverageStatus = "pending"
	QuizCoverageStatusActive   QuizCoverageStatus = "active"
	QuizCoverageStatusCovered  QuizCoverageStatus = "covered"
	QuizCoverageStatusDeferred QuizCoverageStatus = "deferred"
)

type QuizRuntimeState struct {
	ID              string
	ProjectID       string
	CaseSlug        string
	Status          QuizRuntimeStatus
	Mode            QuizMode
	QuestionBudget  int
	QuestionsAsked  int
	AnswersRecorded int
	SkipsRecorded   int
	ActiveTopic     string
	Coverage        []QuizCoverageTopic
	ActiveQuestion  *QuizQuestion
	Trail           []QuizTrailNode
	Candidates      []QuizCandidate
	RawPath         string
	CuratedPath     string
	UpdatedAt       time.Time
}

type QuizQuestion struct {
	ID        string
	Prompt    string
	Required  bool
	Topic     string
	Depth     int
	ParentID  string
	TrunkID   string
	RouteID   string
	Rationale string
	CreatedAt time.Time
}

type QuizEntry struct {
	ID         string
	QuestionID string
	Question   string
	Answer     string
	Status     QuizEntryStatus
	Topic      string
	Depth      int
	ParentID   string
	Mode       QuizMode
	Tags       []string
	CreatedAt  time.Time
	AnsweredAt time.Time
}

type QuizCoverageTopic struct {
	ID            string
	Label         string
	Status        QuizCoverageStatus
	QuestionCount int
	AnswerCount   int
	SkipCount     int
}

type QuizTrailNode struct {
	QuestionID string
	Prompt     string
	Topic      string
	Depth      int
}

type QuizCandidate struct {
	ID        string
	Prompt    string
	Topic     string
	Depth     int
	ParentID  string
	RouteID   string
	Hints     []string
	CreatedAt time.Time
}
