package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type projectState struct {
	projectDir           string
	projectID            string
	projectMetadata      authoringcore.ProjectMetadata
	selectedProjectField int
	projectEdit          ProjectEditState
	projectDirty         bool
	projectExecutor      ProjectMetadataExecutor
}

type caseState struct {
	cases             []authoringcore.CaseMetadata
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
	dashTicketKey           string
	observability           ObservabilityView
	workReadiness           WorkReadinessView
	agentWorkPriorityFilter string
	agentWorkOrder          workflowstate.TicketOrder
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

type updateOutcome struct {
	Command tea.Cmd
	Handled bool
}
