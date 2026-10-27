package quiz

import (
	"fmt"
	"strings"
	"time"

	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
)

const (
	// DefaultQuestionBudget defines a package boundary constant.
	DefaultQuestionBudget = 50
	// MaxFollowUpDepth defines a package boundary constant.
	MaxFollowUpDepth = 3

	// BranchReturnReasonMaxDepth defines a package boundary constant.
	BranchReturnReasonMaxDepth = "max_depth_reached"
	// BranchReturnReasonSkipped defines a package boundary constant.
	BranchReturnReasonSkipped = "skipped"
)

// Scope describes a package boundary value.
type Scope struct {
	Namespace string `json:",omitempty"`
	ProjectID string
	CaseSlug  string
}

// RuntimeStatus defines a package boundary type.
type RuntimeStatus string

const (
	// RuntimeStatusIdle defines a package boundary constant.
	RuntimeStatusIdle RuntimeStatus = "idle"
	// RuntimeStatusActive defines a package boundary constant.
	RuntimeStatusActive RuntimeStatus = "active"
	// RuntimeStatusDone defines a package boundary constant.
	RuntimeStatusDone RuntimeStatus = "done"
	// RuntimeStatusCancelled defines a package boundary constant.
	RuntimeStatusCancelled RuntimeStatus = "cancelled"
	// RuntimeStatusBlocked defines a package boundary constant.
	RuntimeStatusBlocked RuntimeStatus = "blocked"
)

// Mode defines a package boundary type.
type Mode string

const (
	// ModeDeep defines a package boundary constant.
	ModeDeep Mode = "deep"
	// ModeFast defines a package boundary constant.
	ModeFast Mode = "fast"
)

// EntryStatus defines a package boundary type.
type EntryStatus string

const (
	// EntryStatusAnswered defines a package boundary constant.
	EntryStatusAnswered EntryStatus = "answered"
	// EntryStatusSkipped defines a package boundary constant.
	EntryStatusSkipped EntryStatus = "skipped"
)

// CoverageStatus defines a package boundary type.
type CoverageStatus string

const (
	// CoverageStatusPending defines a package boundary constant.
	CoverageStatusPending CoverageStatus = "pending"
	// CoverageStatusActive defines a package boundary constant.
	CoverageStatusActive CoverageStatus = "active"
	// CoverageStatusCovered defines a package boundary constant.
	CoverageStatusCovered CoverageStatus = "covered"
	// CoverageStatusDeferred defines a package boundary constant.
	CoverageStatusDeferred CoverageStatus = "deferred"
)

// RuntimeState describes a package boundary value.
type RuntimeState struct {
	ClarificationComplete bool `json:",omitempty"`
	SchemaVersion         int
	SessionID             string
	Revision              string `json:"-"`
	FastBatchID           string
	FastSourceHash        string
	FastQueue             []Candidate
	FastConsumed          []string
	ContinuationPending   bool
	ContinuationError     string
	GenerationOperationID string `json:",omitempty"`
	ID                    string
	ProjectID             string
	CaseSlug              string
	Status                RuntimeStatus
	Mode                  Mode
	QuestionBudget        int
	QuestionsAsked        int
	AnswersRecorded       int
	SkipsRecorded         int
	ActiveTopic           string
	Coverage              []CoverageTopic
	ActiveQuestion        *Question
	Trail                 []TrailNode
	BranchReturnReason    string
	Candidates            []Candidate
	RawPath               string
	CuratedPath           string
	UpdatedAt             time.Time
}

// Question describes a package boundary value.
type Question struct {
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

// Entry describes a package boundary value.
type Entry struct {
	ID           string
	QuestionID   string
	Question     string
	Answer       string
	Status       EntryStatus
	Topic        string
	Depth        int
	ParentID     string
	Mode         Mode
	Tags         []string
	PromptIntent string
	CreatedAt    time.Time
	AnsweredAt   time.Time
}

// CoverageTopic describes a package boundary value.
type CoverageTopic struct {
	ID            string
	Label         string
	Status        CoverageStatus
	QuestionCount int
	AnswerCount   int
	SkipCount     int
}

// TrailNode describes a package boundary value.
type TrailNode struct {
	QuestionID string
	Prompt     string
	Topic      string
	Depth      int
}

// Candidate describes a package boundary value.
type Candidate struct {
	ID        string
	Prompt    string
	Topic     string
	Depth     int
	ParentID  string
	RouteID   string
	Hints     []string
	CreatedAt time.Time
}

// ValidateRuntimeState validates the corresponding package boundary value.
func ValidateRuntimeState(state RuntimeState) error {
	err := validateRuntimeIdentity(state)
	if err != nil {
		return err
	}

	err = validateRuntimeCounts(state)
	if err != nil {
		return err
	}

	err = validateRuntimeActiveQuestion(state)
	if err != nil {
		return err
	}

	err = validateRuntimeCollections(state)
	if err != nil {
		return err
	}

	return validateRuntimeArtifacts(state)
}

// ValidateScope validates the corresponding package boundary value.
func ValidateScope(scope Scope) error {
	if scope.Namespace != "" && !opsnamespace.Valid(scope.Namespace) {
		return fmt.Errorf("invalid quiz namespace")
	}

	if strings.TrimSpace(scope.ProjectID) == "" {
		return fmt.Errorf("quiz scope project id is required")
	}

	return nil
}

// ValidateQuestion validates the corresponding package boundary value.
func ValidateQuestion(question Question) error {
	err := validateQuestionCore(question)
	if err != nil {
		return err
	}

	return validateQuestionRoute(question)
}

// ValidateEntry validates the corresponding package boundary value.
func ValidateEntry(entry Entry) error {
	err := validateEntryCore(entry)
	if err != nil {
		return err
	}

	err = validateEntryAnswer(entry)
	if err != nil {
		return err
	}

	return validateEntryTimes(entry)
}

// ValidateCoverageTopic validates the corresponding package boundary value.
func ValidateCoverageTopic(topic CoverageTopic) error {
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

// ValidateTrailNode validates the corresponding package boundary value.
func ValidateTrailNode(node TrailNode) error {
	if strings.TrimSpace(node.QuestionID) == "" {
		return fmt.Errorf("quiz trail node question id is required")
	}

	if strings.TrimSpace(node.Prompt) == "" {
		return fmt.Errorf("quiz trail node prompt is required")
	}

	if strings.TrimSpace(node.Topic) == "" {
		return fmt.Errorf("quiz trail node topic is required")
	}

	return validateDepth(node.Depth)
}

// ValidateCandidate validates the corresponding package boundary value.
func ValidateCandidate(candidate Candidate) error {
	if strings.TrimSpace(candidate.ID) == "" {
		return fmt.Errorf("quiz candidate id is required")
	}

	if strings.TrimSpace(candidate.Prompt) == "" {
		return fmt.Errorf("quiz candidate prompt is required")
	}

	if strings.TrimSpace(candidate.Topic) == "" {
		return fmt.Errorf("quiz candidate topic is required")
	}

	err := validateDepth(candidate.Depth)
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
func (s RuntimeStatus) Valid() bool {
	switch s {
	case RuntimeStatusIdle,
		RuntimeStatusActive,
		RuntimeStatusDone,
		RuntimeStatusCancelled,
		RuntimeStatusBlocked:
		return true
	default:
		return false
	}
}

// Valid reports whether the value is one of the supported constants.
func (m Mode) Valid() bool {
	switch m {
	case ModeDeep, ModeFast:
		return true
	default:
		return false
	}
}

// Valid reports whether the value is one of the supported constants.
func (s EntryStatus) Valid() bool {
	switch s {
	case EntryStatusAnswered, EntryStatusSkipped:
		return true
	default:
		return false
	}
}

// Valid reports whether the value is one of the supported constants.
func (s CoverageStatus) Valid() bool {
	switch s {
	case CoverageStatusPending,
		CoverageStatusActive,
		CoverageStatusCovered,
		CoverageStatusDeferred:
		return true
	default:
		return false
	}
}

func validateRuntimeCounts(state RuntimeState) error {
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

func validateRuntimeIdentity(state RuntimeState) error {
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

func validateRuntimeActiveQuestion(state RuntimeState) error {
	if state.Status == RuntimeStatusActive && state.ActiveQuestion == nil {
		return fmt.Errorf("quiz runtime active question is required when active")
	}

	if quizRuntimeHasInvalidInactiveQuestion(state) {
		return fmt.Errorf("quiz runtime active question must be empty unless active or blocked")
	}

	if state.ActiveQuestion == nil {
		return nil
	}

	return ValidateQuestion(*state.ActiveQuestion)
}

func quizRuntimeHasInvalidInactiveQuestion(state RuntimeState) bool {
	return state.Status != RuntimeStatusActive &&
		state.Status != RuntimeStatusBlocked &&
		state.ActiveQuestion != nil
}

func validateRuntimeCollections(state RuntimeState) error {
	err := validateRuntimeCoverage(state.Coverage)
	if err != nil {
		return err
	}

	err = validateRuntimeTrail(state.Trail)
	if err != nil {
		return err
	}

	return validateRuntimeCandidates(state.Candidates)
}

func validateRuntimeCoverage(topics []CoverageTopic) error {
	for index, topic := range topics {
		err := ValidateCoverageTopic(topic)
		if err != nil {
			return fmt.Errorf("quiz runtime coverage topic %d: %w", index+1, err)
		}
	}

	return nil
}

func validateRuntimeTrail(nodes []TrailNode) error {
	for index, node := range nodes {
		err := ValidateTrailNode(node)
		if err != nil {
			return fmt.Errorf("quiz runtime trail node %d: %w", index+1, err)
		}
	}

	return nil
}

func validateRuntimeCandidates(candidates []Candidate) error {
	for index, candidate := range candidates {
		err := ValidateCandidate(candidate)
		if err != nil {
			return fmt.Errorf("quiz runtime candidate %d: %w", index+1, err)
		}
	}

	return nil
}

func validateRuntimeArtifacts(state RuntimeState) error {
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

func validateQuestionCore(question Question) error {
	if strings.TrimSpace(question.ID) == "" {
		return fmt.Errorf("quiz question id is required")
	}

	if strings.TrimSpace(question.Prompt) == "" {
		return fmt.Errorf("quiz question prompt is required")
	}

	if strings.TrimSpace(question.Topic) == "" {
		return fmt.Errorf("quiz question topic is required")
	}

	err := validateDepth(question.Depth)
	if err != nil {
		return err
	}

	if question.CreatedAt.IsZero() {
		return fmt.Errorf("quiz question created time is required")
	}

	return nil
}

func validateQuestionRoute(question Question) error {
	if question.Depth == 1 {
		return validateTrunkQuestion(question)
	}

	return validateFollowUpQuestion(question)
}

func validateTrunkQuestion(question Question) error {
	if strings.TrimSpace(question.ParentID) != "" {
		return fmt.Errorf("quiz trunk question parent id must be empty")
	}

	if question.TrunkID != question.ID {
		return fmt.Errorf("quiz trunk question trunk id must match question id")
	}

	return nil
}

func validateFollowUpQuestion(question Question) error {
	if strings.TrimSpace(question.ParentID) == "" {
		return fmt.Errorf("quiz follow-up question parent id is required")
	}

	if strings.TrimSpace(question.TrunkID) == "" {
		return fmt.Errorf("quiz follow-up question trunk id is required")
	}

	return nil
}

func validateEntryCore(entry Entry) error {
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

	err := validateDepth(entry.Depth)
	if err != nil {
		return err
	}

	if entry.Depth > 1 && strings.TrimSpace(entry.ParentID) == "" {
		return fmt.Errorf("quiz follow-up entry parent id is required")
	}

	return nil
}

func validateEntryAnswer(entry Entry) error {
	if entry.Status == EntryStatusAnswered && strings.TrimSpace(entry.Answer) == "" {
		return fmt.Errorf("quiz answered entry answer is required")
	}

	if entry.Status == EntryStatusSkipped && strings.TrimSpace(entry.Answer) != "" {
		return fmt.Errorf("quiz skipped entry answer must be empty")
	}

	return nil
}

func validateEntryTimes(entry Entry) error {
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

func validateDepth(depth int) error {
	if depth < 1 || depth > MaxFollowUpDepth {
		return fmt.Errorf("quiz depth must be between 1 and %d", MaxFollowUpDepth)
	}

	return nil
}
