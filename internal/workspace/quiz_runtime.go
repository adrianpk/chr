package workspace

import (
	"fmt"
	"strings"
	"time"
)

const (
	// DefaultQuizQuestionBudget defines a package boundary constant.
	DefaultQuizQuestionBudget = 50
	// MaxQuizFollowUpDepth defines a package boundary constant.
	MaxQuizFollowUpDepth = 3

	// QuizBranchReturnReasonMaxDepth defines a package boundary constant.
	QuizBranchReturnReasonMaxDepth = "max_depth_reached"
	// QuizBranchReturnReasonSkipped defines a package boundary constant.
	QuizBranchReturnReasonSkipped = "skipped"
)

// QuizScope describes a package boundary value.
type QuizScope struct {
	ProjectID string
	CaseSlug  string
}

// QuizRuntimeStatus defines a package boundary type.
type QuizRuntimeStatus string

const (
	// QuizRuntimeStatusIdle defines a package boundary constant.
	QuizRuntimeStatusIdle QuizRuntimeStatus = "idle"
	// QuizRuntimeStatusActive defines a package boundary constant.
	QuizRuntimeStatusActive QuizRuntimeStatus = "active"
	// QuizRuntimeStatusDone defines a package boundary constant.
	QuizRuntimeStatusDone QuizRuntimeStatus = "done"
	// QuizRuntimeStatusCancelled defines a package boundary constant.
	QuizRuntimeStatusCancelled QuizRuntimeStatus = "cancelled"
	// QuizRuntimeStatusBlocked defines a package boundary constant.
	QuizRuntimeStatusBlocked QuizRuntimeStatus = "blocked"
)

// QuizMode defines a package boundary type.
type QuizMode string

const (
	// QuizModeDeep defines a package boundary constant.
	QuizModeDeep QuizMode = "deep"
	// QuizModeFast defines a package boundary constant.
	QuizModeFast QuizMode = "fast"
)

// QuizEntryStatus defines a package boundary type.
type QuizEntryStatus string

const (
	// QuizEntryStatusAnswered defines a package boundary constant.
	QuizEntryStatusAnswered QuizEntryStatus = "answered"
	// QuizEntryStatusSkipped defines a package boundary constant.
	QuizEntryStatusSkipped QuizEntryStatus = "skipped"
)

// QuizCoverageStatus defines a package boundary type.
type QuizCoverageStatus string

const (
	// QuizCoverageStatusPending defines a package boundary constant.
	QuizCoverageStatusPending QuizCoverageStatus = "pending"
	// QuizCoverageStatusActive defines a package boundary constant.
	QuizCoverageStatusActive QuizCoverageStatus = "active"
	// QuizCoverageStatusCovered defines a package boundary constant.
	QuizCoverageStatusCovered QuizCoverageStatus = "covered"
	// QuizCoverageStatusDeferred defines a package boundary constant.
	QuizCoverageStatusDeferred QuizCoverageStatus = "deferred"
)

// QuizRuntimeState describes a package boundary value.
type QuizRuntimeState struct {
	ID                 string
	ProjectID          string
	CaseSlug           string
	Status             QuizRuntimeStatus
	Mode               QuizMode
	QuestionBudget     int
	QuestionsAsked     int
	AnswersRecorded    int
	SkipsRecorded      int
	ActiveTopic        string
	Coverage           []QuizCoverageTopic
	ActiveQuestion     *QuizQuestion
	Trail              []QuizTrailNode
	BranchReturnReason string
	Candidates         []QuizCandidate
	RawPath            string
	CuratedPath        string
	UpdatedAt          time.Time
}

// QuizQuestion describes a package boundary value.
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

// QuizEntry describes a package boundary value.
type QuizEntry struct {
	ID           string
	QuestionID   string
	Question     string
	Answer       string
	Status       QuizEntryStatus
	Topic        string
	Depth        int
	ParentID     string
	Mode         QuizMode
	Tags         []string
	PromptIntent string
	CreatedAt    time.Time
	AnsweredAt   time.Time
}

// QuizCoverageTopic describes a package boundary value.
type QuizCoverageTopic struct {
	ID            string
	Label         string
	Status        QuizCoverageStatus
	QuestionCount int
	AnswerCount   int
	SkipCount     int
}

// QuizTrailNode describes a package boundary value.
type QuizTrailNode struct {
	QuestionID string
	Prompt     string
	Topic      string
	Depth      int
}

// QuizCandidate describes a package boundary value.
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

// ValidateQuizRuntimeState validates the corresponding package boundary value.
func ValidateQuizRuntimeState(state QuizRuntimeState) error {
	err := validateQuizRuntimeIdentity(state)
	if err != nil {
		return err
	}

	err = validateQuizRuntimeCounts(state)
	if err != nil {
		return err
	}

	err = validateQuizRuntimeActiveQuestion(state)
	if err != nil {
		return err
	}

	err = validateQuizRuntimeCollections(state)
	if err != nil {
		return err
	}

	return validateQuizRuntimeArtifacts(state)
}

// ValidateQuizScope validates the corresponding package boundary value.
func ValidateQuizScope(scope QuizScope) error {
	if strings.TrimSpace(scope.ProjectID) == "" {
		return fmt.Errorf("quiz scope project id is required")
	}

	return nil
}

// ValidateQuizQuestion validates the corresponding package boundary value.
func ValidateQuizQuestion(question QuizQuestion) error {
	err := validateQuizQuestionCore(question)
	if err != nil {
		return err
	}

	return validateQuizQuestionRoute(question)
}

// ValidateQuizEntry validates the corresponding package boundary value.
func ValidateQuizEntry(entry QuizEntry) error {
	err := validateQuizEntryCore(entry)
	if err != nil {
		return err
	}

	err = validateQuizEntryAnswer(entry)
	if err != nil {
		return err
	}

	return validateQuizEntryTimes(entry)
}

// ValidateQuizCoverageTopic validates the corresponding package boundary value.
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

// ValidateQuizTrailNode validates the corresponding package boundary value.
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

// ValidateQuizCandidate validates the corresponding package boundary value.
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

// Valid reports whether the value is one of the supported constants.
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

// Valid reports whether the value is one of the supported constants.
func (m QuizMode) Valid() bool {
	switch m {
	case QuizModeDeep, QuizModeFast:
		return true
	default:
		return false
	}
}

// Valid reports whether the value is one of the supported constants.
func (s QuizEntryStatus) Valid() bool {
	switch s {
	case QuizEntryStatusAnswered, QuizEntryStatusSkipped:
		return true
	default:
		return false
	}
}

// Valid reports whether the value is one of the supported constants.
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

func validateQuizRuntimeIdentity(state QuizRuntimeState) error {
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

	return nil
}

func validateQuizRuntimeActiveQuestion(state QuizRuntimeState) error {
	if state.Status == QuizRuntimeStatusActive && state.ActiveQuestion == nil {
		return fmt.Errorf("quiz runtime active question is required when active")
	}

	if quizRuntimeHasInvalidInactiveQuestion(state) {
		return fmt.Errorf("quiz runtime active question must be empty unless active or blocked")
	}

	if state.ActiveQuestion == nil {
		return nil
	}

	return ValidateQuizQuestion(*state.ActiveQuestion)
}

func quizRuntimeHasInvalidInactiveQuestion(state QuizRuntimeState) bool {
	return state.Status != QuizRuntimeStatusActive &&
		state.Status != QuizRuntimeStatusBlocked &&
		state.ActiveQuestion != nil
}

func validateQuizRuntimeCollections(state QuizRuntimeState) error {
	err := validateQuizRuntimeCoverage(state.Coverage)
	if err != nil {
		return err
	}

	err = validateQuizRuntimeTrail(state.Trail)
	if err != nil {
		return err
	}

	return validateQuizRuntimeCandidates(state.Candidates)
}

func validateQuizRuntimeCoverage(topics []QuizCoverageTopic) error {
	for index, topic := range topics {
		err := ValidateQuizCoverageTopic(topic)
		if err != nil {
			return fmt.Errorf("quiz runtime coverage topic %d: %w", index+1, err)
		}
	}

	return nil
}

func validateQuizRuntimeTrail(nodes []QuizTrailNode) error {
	for index, node := range nodes {
		err := ValidateQuizTrailNode(node)
		if err != nil {
			return fmt.Errorf("quiz runtime trail node %d: %w", index+1, err)
		}
	}

	return nil
}

func validateQuizRuntimeCandidates(candidates []QuizCandidate) error {
	for index, candidate := range candidates {
		err := ValidateQuizCandidate(candidate)
		if err != nil {
			return fmt.Errorf("quiz runtime candidate %d: %w", index+1, err)
		}
	}

	return nil
}

func validateQuizRuntimeArtifacts(state QuizRuntimeState) error {
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

func validateQuizQuestionCore(question QuizQuestion) error {
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

	if question.CreatedAt.IsZero() {
		return fmt.Errorf("quiz question created time is required")
	}

	return nil
}

func validateQuizQuestionRoute(question QuizQuestion) error {
	if question.Depth == 1 {
		return validateQuizTrunkQuestion(question)
	}

	return validateQuizFollowUpQuestion(question)
}

func validateQuizTrunkQuestion(question QuizQuestion) error {
	if strings.TrimSpace(question.ParentID) != "" {
		return fmt.Errorf("quiz trunk question parent id must be empty")
	}

	if question.TrunkID != question.ID {
		return fmt.Errorf("quiz trunk question trunk id must match question id")
	}

	return nil
}

func validateQuizFollowUpQuestion(question QuizQuestion) error {
	if strings.TrimSpace(question.ParentID) == "" {
		return fmt.Errorf("quiz follow-up question parent id is required")
	}

	if strings.TrimSpace(question.TrunkID) == "" {
		return fmt.Errorf("quiz follow-up question trunk id is required")
	}

	return nil
}

func validateQuizEntryCore(entry QuizEntry) error {
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

	if entry.Depth > 1 && strings.TrimSpace(entry.ParentID) == "" {
		return fmt.Errorf("quiz follow-up entry parent id is required")
	}

	return nil
}

func validateQuizEntryAnswer(entry QuizEntry) error {
	if entry.Status == QuizEntryStatusAnswered && strings.TrimSpace(entry.Answer) == "" {
		return fmt.Errorf("quiz answered entry answer is required")
	}

	if entry.Status == QuizEntryStatusSkipped && strings.TrimSpace(entry.Answer) != "" {
		return fmt.Errorf("quiz skipped entry answer must be empty")
	}

	return nil
}

func validateQuizEntryTimes(entry QuizEntry) error {
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

func validateQuizDepth(depth int) error {
	if depth < 1 || depth > MaxQuizFollowUpDepth {
		return fmt.Errorf("quiz depth must be between 1 and %d", MaxQuizFollowUpDepth)
	}

	return nil
}
