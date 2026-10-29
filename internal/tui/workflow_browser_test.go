package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Retained selection has its own viewport, member paging and late-refresh fence;
// it never follows queue selection or silently returns to an older workflow page.
func TestWorkflowBrowser(t *testing.T) {
	executor := &fakeWorkActionExecutor{}
	model := NewModel(Options{WorkExecutor: executor})
	model.selectView(ViewDash)

	model.dashWorkflowDetails = true

	page := WorkflowPageView{NextOffset: 16}
	for index := range 16 {
		page.Items = append(page.Items, WorkflowView{ID: fmt.Sprintf("run-%02d", index), State: "pending", Title: fmt.Sprintf("Retained work %02d", index), Revision: 3, NextOffset: 8})
	}

	model.applyWorkflowPage(page)

	model.workflow = page.Items[15]

	view := model.renderWorkflowWorkspace(120, 24)
	if !strings.Contains(view, "Retained work 15") {
		t.Fatal("last retained selection clipped", view)
	}

	command, _ := model.handleWorkflowBrowseKey("n")
	command()

	if executor.request.Workflow.Offset != 8 || executor.request.Workflow.ExpectedRevision != 3 {
		t.Fatal("member page lost revision fence")
	}

	model.applyWorkflowPage(WorkflowPageView{Offset: 16, Items: []WorkflowView{{ID: "last-run", Revision: 1}}, NextOffset: -1})
	model.completeAgentWorkRefresh(agentWorkRefreshFinishedMsg{request: AgentWorkRefreshRequest{WorkflowID: "run-15"}, result: AgentWorkRefreshResult{Observability: ObservabilityView{WorkflowPage: page}}})

	if model.workflowPage.Offset != 16 || model.workflow.ID != "last-run" {
		t.Fatal("late refresh replaced current page")
	}

	model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if model.dashWorkflowDetails {
		t.Fatal("workflow browser did not return to Dash queue")
	}
}
