package tui

import "forge.adrianpk.com/chr/internal/core"

type projectState struct {
	projectDir           string
	projectID            string
	projectMetadata      core.ProjectMetadata
	selectedProjectField int
	projectEdit          ProjectEditState
	projectDirty         bool
	projectExecutor      ProjectMetadataExecutor
}

type caseState struct {
	cases             []core.CaseMetadata
	selectedCase      int
	selectedCaseField int
	caseEdit          ProjectEditState
	caseDirty         bool
	caseFilter        string
	caseFilterEditing bool
	caseExecutor      CaseMetadataExecutor
}

type navigationState struct {
	views        []View
	selected     int
	selectedPane int
	detailScroll int
}

type layoutState struct {
	maximizedPane          int
	verticalSplitPercent   int
	horizontalSplitPercent int
	layoutDrag             layoutDrag
	width                  int
	height                 int
}

type ticketState struct {
	tickets             []TicketView
	selectedTicket      int
	boardColumn         string
	previousBoardColumn string
	ticketFilter        string
	filterEntryMode     bool
	ticketAction        TicketActionState
	ticketExecutor      TicketActionExecutor
}

type documentState struct {
	documents        []DocumentView
	selectedDocument int
}

type dumpState struct {
	dumps        []DumpView
	selectedDump int
	dumpEdit     DumpEditState
	dumpAction   DumpActionState
	dumpExecutor DumpActionExecutor
}

type quizState struct {
	quizzes      []QuizView
	selectedQuiz int
	quizAction   QuizActionState
	quizAnswer   QuizAnswerState
	quizBusy     bool
	quizExecutor QuizActionExecutor
}

type specState struct {
	specs        []SpecView
	selectedSpec int
	specAction   SpecActionState
	specBusy     bool
	specExecutor SpecActionExecutor
}

type planState struct {
	plans        []PlanView
	selectedPlan int
	planAction   PlanActionState
	planBusy     bool
	planExecutor PlanActionExecutor
}

type authoringState struct {
	authoringSessions        []AuthoringSessionView
	selectedAuthoringSession int
}

type workspaceState struct {
	workspaceGraph       core.WorkspaceGraph
	workspaceIndex       core.WorkspaceIndex
	semanticDigestStates []core.WorkspaceSemanticDigestState
	workspaceExecutor    WorkspaceRefreshExecutor
}

type interactionState struct {
	commandPalette CommandPaletteState
	artifactAction ArtifactActionState
	actionExecutor ArtifactActionExecutor
	sourceExecutor SourceOpenExecutor
	statusMessage  string
}

type workState struct {
	observability           ObservabilityView
	workReadiness           WorkReadinessView
	agentWorkPriorityFilter string
	agentWorkOrder          core.TicketLoopOrder
	workflow                WorkflowView
	workflowPage            WorkflowPageView
	workAction              WorkActionState
	workBusy                bool
	workExecutor            WorkActionExecutor
}

type backgroundWorkState struct {
	workflowExecutionBusy    bool
	agentWorkRefreshExecutor AgentWorkRefreshExecutor
	agentWorkRefreshInFlight bool
}

type harnessState struct {
	harnessExecutor  HarnessSelectionExecutor
	harnessSelection harnessSelectionState
	harnessSequence  uint64
}

type guidedModelState struct {
	guidedExecutor GuidedExecutor
	guided         guidedState
	guidedSequence uint64
}
