package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/core"
)

// Native workflow actions select extents or operate on displayed authorization.
const (
	WorkActionWorkflowTicket    WorkActionID = "workflow_ticket"
	WorkActionWorkflowSlice     WorkActionID = "workflow_slice"
	WorkActionWorkflowSet       WorkActionID = "workflow_set"
	WorkActionWorkflowCase      WorkActionID = "workflow_case"
	WorkActionWorkflowLoose     WorkActionID = "workflow_loose"
	WorkActionWorkflowStatus    WorkActionID = "workflow_status"
	WorkActionWorkflowNext      WorkActionID = "workflow_next"
	WorkActionWorkflowPage      WorkActionID = "workflow_page"
	WorkActionWorkflowContinue  WorkActionID = "workflow_continue"
	WorkActionWorkflowPause     WorkActionID = "workflow_pause"
	WorkActionWorkflowCancel    WorkActionID = "workflow_cancel"
	WorkActionWorkflowResume    WorkActionID = "workflow_resume"
	WorkActionWorkflowApprove   WorkActionID = "workflow_approve"
	WorkActionWorkflowReconcile WorkActionID = "workflow_reconcile"
)

// WorkflowActionRequest retains service-supplied identities and revisions.
type WorkflowActionRequest struct {
	Preview          bool
	Selection        core.WorkflowPreviewRequest
	Selected         *core.WorkflowSelectionGuard
	ID               string
	ExpectedRevision uint64
	BindingHash      string
	Offset           int
}

// WorkflowView contains bounded display lines and exact control bindings.
type WorkflowView struct {
	ID           string
	Revision     uint64
	ApprovalHash string
	Lines        []string
}

// WorkflowPageView is a page of retained authorization, never a work queue.
type WorkflowPageView struct {
	Offset     int
	Items      []WorkflowView
	NextOffset int
}

func (m *Model) planNativeWorkAction(action WorkActionID) bool {
	action = m.nativeWorkAlias(action)
	if !strings.HasPrefix(string(action), "workflow_") {
		return false
	}

	if m.workflow.ID == "" {
		m.applyWorkflowPage(m.observability.WorkflowPage)
	}

	request := WorkActionRequest{Action: action, Workflow: &WorkflowActionRequest{}}
	if extent := nativeWorkExtent(action); extent != "" {
		if !m.planNativeWorkExtent(&request, extent) {
			return true
		}
	} else if !m.planNativeWorkControl(&request) {
		return true
	}

	m.workAction = WorkActionState{Pending: true, Request: request, Title: "Inspect native workflow", Details: []string{"Enter loads the read-only preview or retained status."}}
	m.statusMessage = ""

	return true
}

func (m Model) nativeWorkAlias(action WorkActionID) WorkActionID {
	if m.observability.Authority != core.WorkStateAuthorityTickets {
		return action
	}

	switch action {
	case WorkActionStartTicket:
		return WorkActionWorkflowTicket
	case WorkActionRunAllReady, WorkActionStartReadyLoop:
		return WorkActionWorkflowLoose
	default:
		return action
	}
}

func nativeWorkExtent(action WorkActionID) string {
	switch action {
	case WorkActionWorkflowTicket:
		return "ticket"
	case WorkActionWorkflowSlice:
		return "slice"
	case WorkActionWorkflowSet:
		return "slice_set"
	case WorkActionWorkflowCase:
		return "use_case"
	case WorkActionWorkflowLoose:
		return "loose"
	default:
		return ""
	}
}

func (m *Model) planNativeWorkExtent(request *WorkActionRequest, extent string) bool {
	native := request.Workflow
	native.Preview = true

	native.Selection = core.WorkflowPreviewRequest{Selector: core.WorkflowSelector{Extent: extent}, Limit: core.WorkflowPageLimit}
	if extent == "loose" {
		native.Selection.Statuses = []string{core.TicketStatusReadyForDev}

		native.Selection.Order = m.agentWorkOrder
		if m.agentWorkPriorityFilter != "" {
			native.Selection.Filters.Priorities = []string{m.agentWorkPriorityFilter}
		}

		return true
	}

	if extent == "use_case" && m.SelectedView() == ViewCases {
		metadata, ok := m.selectedCaseMetadata()
		if !ok {
			m.statusMessage = "No selected use case is available."

			return false
		}

		native.Selection.Selector.Identity = metadata.Slug

		return true
	}

	ticket, ok := m.selectedTicketView()
	if !ok {
		m.statusMessage = "No selected ticket is available for this native extent."

		return false
	}

	request.TicketKey = ticket.Key

	return true
}

func (m *Model) planNativeWorkControl(request *WorkActionRequest) bool {
	if request.Action == WorkActionWorkflowPage {
		request.Workflow.Offset = max(0, m.workflowPage.NextOffset)

		return true
	}

	if request.Action == WorkActionWorkflowNext {
		m.selectNextWorkflow()

		request.Action = WorkActionWorkflowStatus
	}

	current := m.workflow
	if current.ID == "" && request.Action != WorkActionWorkflowStatus {
		m.statusMessage = "Inspect retained workflows before choosing a control."

		return false
	}

	request.Workflow.ID, request.Workflow.ExpectedRevision = current.ID, current.Revision
	request.Workflow.BindingHash = current.ApprovalHash
	request.Workflow.Preview = request.Action != WorkActionWorkflowStatus

	return true
}

func (m *Model) selectNextWorkflow() {
	items := m.workflowPage.Items
	for i, item := range items {
		if item.ID == m.workflow.ID {
			m.workflow = items[(i+1)%len(items)]

			return
		}
	}
}

func (m *Model) applyWorkflowPage(page WorkflowPageView) {
	m.workflowPage = page
	if len(page.Items) == 0 {
		return
	}

	for _, item := range page.Items {
		if item.ID == m.workflow.ID {
			if item.Revision >= m.workflow.Revision {
				m.workflow = item
			}

			return
		}
	}

	m.workflow = page.Items[0]
}

func (m *Model) startNativeWorkAction(request WorkActionRequest) tea.Cmd {
	if m.workExecutor == nil {
		m.statusMessage = "Work action executor is unavailable."

		return nil
	}

	if nativeWorkExecutes(request) {
		if m.workflowExecutionBusy {
			m.statusMessage = "A native execution action is already running. Status and stop controls remain available."

			return nil
		}

		m.workflowExecutionBusy = true
	}

	m.statusMessage = "Loading shared workflow operation..."

	return m.executeWorkActionCommand(request)
}

func nativeWorkExecutes(request WorkActionRequest) bool {
	return !request.Workflow.Preview && (nativeWorkExtent(request.Action) != "" || request.Action == WorkActionWorkflowContinue)
}

func (m *Model) completeNativeWorkAction(message workActionFinishedMsg) {
	if nativeWorkExecutes(message.Request) {
		m.workflowExecutionBusy = false
	}

	if message.Result.Tickets != nil {
		m.tickets = NewTicketViews(message.Result.Tickets)
		m.selectFirstFilteredTicket()
	}

	if message.Result.WorkflowPage != nil {
		m.applyWorkflowPage(*message.Result.WorkflowPage)
	}

	if message.Result.Workflow != nil {
		m.workflow = *message.Result.Workflow
	}

	if message.Err != nil {
		m.statusMessage = "Work action failed: " + message.Err.Error()

		return
	}

	if message.Result.Confirmation != nil {
		m.workAction = *message.Result.Confirmation
	}

	m.statusMessage = message.Result.Message
}

func (m Model) nativeWorkflowLines() []string {
	if m.workflow.ID != "" {
		return m.workflow.Lines
	}

	if len(m.observability.WorkflowPage.Items) > 0 {
		return m.observability.WorkflowPage.Items[0].Lines
	}

	return nil
}

func (m Model) renderNativeWorkflow() string {
	lines := m.nativeWorkflowLines()
	if len(lines) == 0 {
		return ""
	}

	return fmt.Sprintf("\nNative Workflow\n  %s\n", strings.Join(lines, "\n  "))
}
