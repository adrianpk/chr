package chc

import (
	"fmt"
	"strings"
	"time"
)

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

func ValidateQuizRuntimeState(state QuizRuntimeState) error {
	if strings.TrimSpace(state.ID) == "" {
		return fmt.Errorf("quiz runtime id is required")
	}

	if strings.TrimSpace(state.ProjectID) == "" {
		return fmt.Errorf("quiz runtime project id is required")
	}

	if !state.Status.Valid() {
		return fmt.Errorf("quiz runtime status is invalid: %s", state.Status)
	}

	if !state.Mode.Valid() {
		return fmt.Errorf("quiz runtime mode is invalid: %s", state.Mode)
	}

	err := validateQuizRuntimeCounts(state)
	if err != nil {
		return err
	}

	err = validateQuizRuntimeActiveQuestion(state)
	if err != nil {
		return err
	}

	for index, topic := range state.Coverage {
		err = ValidateQuizCoverageTopic(topic)
		if err != nil {
			return fmt.Errorf("quiz runtime coverage topic %d: %w", index+1, err)
		}
	}

	for index, node := range state.Trail {
		err = ValidateQuizTrailNode(node)
		if err != nil {
			return fmt.Errorf("quiz runtime trail node %d: %w", index+1, err)
		}
	}

	for index, candidate := range state.Candidates {
		err = ValidateQuizCandidate(candidate)
		if err != nil {
			return fmt.Errorf("quiz runtime candidate %d: %w", index+1, err)
		}
	}

	if strings.TrimSpace(state.RawPath) == "" {
		return fmt.Errorf("quiz runtime raw path is required")
	}

	if strings.TrimSpace(state.CuratedPath) == "" {
		return fmt.Errorf("quiz runtime curated path is required")
	}

	if state.UpdatedAt.IsZero() {
		return fmt.Errorf("quiz runtime updated time is required")
	}

	return nil
}

func ValidateQuizScope(scope QuizScope) error {
	if strings.TrimSpace(scope.ProjectID) == "" {
		return fmt.Errorf("quiz scope project id is required")
	}

	return nil
}

func ValidateQuizQuestion(question QuizQuestion) error {
	if strings.TrimSpace(question.ID) == "" {
		return fmt.Errorf("quiz question id is required")
	}

	if strings.TrimSpace(question.Prompt) == "" {
		return fmt.Errorf("quiz question prompt is required")
	}

	if strings.TrimSpace(question.Topic) == "" {
		return fmt.Errorf("quiz question topic is required")
	}

	err := validateQuizDepth(question.Depth)
	if err != nil {
		return err
	}

	if question.Depth == 1 {
		if strings.TrimSpace(question.ParentID) != "" {
			return fmt.Errorf("quiz trunk question parent id must be empty")
		}

		if question.TrunkID != question.ID {
			return fmt.Errorf("quiz trunk question trunk id must match question id")
		}
	}

	if question.Depth > 1 {
		if strings.TrimSpace(question.ParentID) == "" {
			return fmt.Errorf("quiz follow-up question parent id is required")
		}

		if strings.TrimSpace(question.TrunkID) == "" {
			return fmt.Errorf("quiz follow-up question trunk id is required")
		}
	}

	if question.CreatedAt.IsZero() {
		return fmt.Errorf("quiz question created time is required")
	}

	return nil
}

func ValidateQuizEntry(entry QuizEntry) error {
	if strings.TrimSpace(entry.ID) == "" {
		return fmt.Errorf("quiz entry id is required")
	}

	if strings.TrimSpace(entry.QuestionID) == "" {
		return fmt.Errorf("quiz entry question id is required")
	}

	if strings.TrimSpace(entry.Question) == "" {
		return fmt.Errorf("quiz entry question is required")
	}

	if !entry.Status.Valid() {
		return fmt.Errorf("quiz entry status is invalid: %s", entry.Status)
	}

	if !entry.Mode.Valid() {
		return fmt.Errorf("quiz entry mode is invalid: %s", entry.Mode)
	}

	if strings.TrimSpace(entry.Topic) == "" {
		return fmt.Errorf("quiz entry topic is required")
	}

	err := validateQuizDepth(entry.Depth)
	if err != nil {
		return err
	}

	if entry.Status == QuizEntryStatusAnswered && strings.TrimSpace(entry.Answer) == "" {
		return fmt.Errorf("quiz answered entry answer is required")
	}

	if entry.Status == QuizEntryStatusSkipped && strings.TrimSpace(entry.Answer) != "" {
		return fmt.Errorf("quiz skipped entry answer must be empty")
	}

	if entry.Depth > 1 && strings.TrimSpace(entry.ParentID) == "" {
		return fmt.Errorf("quiz follow-up entry parent id is required")
	}

	if entry.CreatedAt.IsZero() {
		return fmt.Errorf("quiz entry created time is required")
	}

	if entry.AnsweredAt.IsZero() {
		return fmt.Errorf("quiz entry answered time is required")
	}

	if entry.AnsweredAt.Before(entry.CreatedAt) {
		return fmt.Errorf("quiz entry answered time is before created time")
	}

	return nil
}

func ValidateQuizCoverageTopic(topic QuizCoverageTopic) error {
	if strings.TrimSpace(topic.ID) == "" {
		return fmt.Errorf("quiz coverage topic id is required")
	}

	if strings.TrimSpace(topic.Label) == "" {
		return fmt.Errorf("quiz coverage topic label is required")
	}

	if !topic.Status.Valid() {
		return fmt.Errorf("quiz coverage topic status is invalid: %s", topic.Status)
	}

	if topic.QuestionCount < 0 {
		return fmt.Errorf("quiz coverage topic question count cannot be negative")
	}

	if topic.AnswerCount < 0 {
		return fmt.Errorf("quiz coverage topic answer count cannot be negative")
	}

	if topic.SkipCount < 0 {
		return fmt.Errorf("quiz coverage topic skip count cannot be negative")
	}

	if topic.AnswerCount+topic.SkipCount > topic.QuestionCount {
		return fmt.Errorf("quiz coverage topic outcomes cannot exceed question count")
	}

	return nil
}

func ValidateQuizTrailNode(node QuizTrailNode) error {
	if strings.TrimSpace(node.QuestionID) == "" {
		return fmt.Errorf("quiz trail node question id is required")
	}

	if strings.TrimSpace(node.Prompt) == "" {
		return fmt.Errorf("quiz trail node prompt is required")
	}

	if strings.TrimSpace(node.Topic) == "" {
		return fmt.Errorf("quiz trail node topic is required")
	}

	return validateQuizDepth(node.Depth)
}

func ValidateQuizCandidate(candidate QuizCandidate) error {
	if strings.TrimSpace(candidate.ID) == "" {
		return fmt.Errorf("quiz candidate id is required")
	}

	if strings.TrimSpace(candidate.Prompt) == "" {
		return fmt.Errorf("quiz candidate prompt is required")
	}

	if strings.TrimSpace(candidate.Topic) == "" {
		return fmt.Errorf("quiz candidate topic is required")
	}

	err := validateQuizDepth(candidate.Depth)
	if err != nil {
		return err
	}

	if candidate.Depth > 1 && strings.TrimSpace(candidate.ParentID) == "" {
		return fmt.Errorf("quiz follow-up candidate parent id is required")
	}

	if candidate.CreatedAt.IsZero() {
		return fmt.Errorf("quiz candidate created time is required")
	}

	return nil
}

func (s QuizRuntimeStatus) Valid() bool {
	switch s {
	case QuizRuntimeStatusIdle,
		QuizRuntimeStatusActive,
		QuizRuntimeStatusDone,
		QuizRuntimeStatusCancelled,
		QuizRuntimeStatusBlocked:
		return true
	default:
		return false
	}
}

func (m QuizMode) Valid() bool {
	switch m {
	case QuizModeDeep,
		QuizModeFast:
		return true
	default:
		return false
	}
}

func (s QuizEntryStatus) Valid() bool {
	switch s {
	case QuizEntryStatusAnswered,
		QuizEntryStatusSkipped:
		return true
	default:
		return false
	}
}

func (s QuizCoverageStatus) Valid() bool {
	switch s {
	case QuizCoverageStatusPending,
		QuizCoverageStatusActive,
		QuizCoverageStatusCovered,
		QuizCoverageStatusDeferred:
		return true
	default:
		return false
	}
}

func validateQuizRuntimeCounts(state QuizRuntimeState) error {
	if state.QuestionBudget <= 0 {
		return fmt.Errorf("quiz runtime question budget must be greater than zero")
	}

	if state.QuestionsAsked < 0 {
		return fmt.Errorf("quiz runtime questions asked cannot be negative")
	}

	if state.AnswersRecorded < 0 {
		return fmt.Errorf("quiz runtime answers recorded cannot be negative")
	}

	if state.SkipsRecorded < 0 {
		return fmt.Errorf("quiz runtime skips recorded cannot be negative")
	}

	if state.QuestionsAsked > state.QuestionBudget {
		return fmt.Errorf("quiz runtime questions asked cannot exceed question budget")
	}

	if state.AnswersRecorded+state.SkipsRecorded > state.QuestionsAsked {
		return fmt.Errorf("quiz runtime outcomes cannot exceed questions asked")
	}

	return nil
}

func validateQuizRuntimeActiveQuestion(state QuizRuntimeState) error {
	if state.Status == QuizRuntimeStatusActive && state.ActiveQuestion == nil {
		return fmt.Errorf("quiz runtime active question is required when active")
	}

	if state.Status != QuizRuntimeStatusActive &&
		state.Status != QuizRuntimeStatusBlocked &&
		state.ActiveQuestion != nil {
		return fmt.Errorf("quiz runtime active question must be empty unless active or blocked")
	}

	if state.ActiveQuestion == nil {
		return nil
	}

	return ValidateQuizQuestion(*state.ActiveQuestion)
}

func validateQuizDepth(depth int) error {
	if depth < 1 || depth > MaxQuizFollowUpDepth {
		return fmt.Errorf("quiz depth must be between 1 and %d", MaxQuizFollowUpDepth)
	}

	return nil
}
