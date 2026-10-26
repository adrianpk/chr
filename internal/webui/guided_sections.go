package webui

// GuidedQuestionView contains the current quiz and its loading observation.
type GuidedQuestionView struct {
	Quiz       *QuizAnswerView
	QuizID     string
	Processing bool
}

// GuidedPageView contains the current screen, navigation guards and visible action flags.
type GuidedPageView struct {
	CaseSlug     string
	Observation  string
	StateURL     string
	StateToken   string
	Surface      string
	CurrentURL   string
	Message      string
	Starting     bool
	Following    bool
	ShowControls bool
	ShowRetry    bool
	ShowReview   bool
	ReadOnly     bool
}

// GuidedIdeasView contains retained ideas and their separate revision-bound request.
type GuidedIdeasView struct {
	Ideas         []GuidedUseCaseIdea
	IdeaRequestID string
	IdeaRevision  string
	CanRetainIdea bool
}

// GuidedLifecycleView contains the displayed lifecycle controls and suspension observation.
type GuidedLifecycleView struct {
	Controls      []GuidedUseCaseControl
	Suspended     bool
	ControlStatus string
}

// GuidedRunView contains the retained journey identity and current committed capture.
type GuidedRunView struct {
	RunID         string
	Namespace     string
	Name          string
	Revision      string
	RequestID     string
	Stage         string
	Status        string
	Capture       string
	Commit        string
	CommittedDump string
}

// GuidedStepView contains the original pending input and current question/retry guards.
type GuidedStepView struct {
	PendingAction  string
	PendingContent string
	Question       string
	QuestionID     string
	InputDigest    string
	RetryAction    string
	Blocker        string
}

// GuidedDiagnosticsView contains observed policy and problems without granting authority.
type GuidedDiagnosticsView struct {
	Policy   []string
	Problems []string
}
