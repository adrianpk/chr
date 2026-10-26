package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type modelRefreshExecutor func(AgentWorkRefreshRequest) (AgentWorkRefreshResult, error)

func (execute modelRefreshExecutor) RefreshAgentWork(request AgentWorkRefreshRequest) (AgentWorkRefreshResult, error) {
	return execute(request)
}

// Updates retain the one event-loop owner; a queued command keeps its original
// executor and request rather than consulting later mutable model state.
func TestModelOwnership(t *testing.T) {
	model := NewModel(Options{})
	for _, message := range []tea.Msg{tea.WindowSizeMsg{Width: 100, Height: 30}, tea.KeyMsg{Type: tea.KeyTab}, struct{}{}} {
		updated, _ := model.Update(message)
		if updated != model {
			t.Fatal("update replaced the event-loop model owner")
		}
	}

	var observed AgentWorkRefreshRequest

	model.workflow.ID = "original"
	model.workflowPage.Offset = 7
	model.agentWorkRefreshExecutor = modelRefreshExecutor(func(request AgentWorkRefreshRequest) (AgentWorkRefreshResult, error) {
		observed = request

		return AgentWorkRefreshResult{}, nil
	})
	command := model.startAgentWorkRefresh()
	model.workflow.ID, model.workflowPage.Offset = "changed", 99
	model.agentWorkRefreshExecutor = modelRefreshExecutor(func(AgentWorkRefreshRequest) (AgentWorkRefreshResult, error) {
		t.Fatal("queued command read a replacement executor")

		return AgentWorkRefreshResult{}, nil
	})
	message := command()

	if observed.WorkflowID != "original" || observed.WorkflowOffset != 7 {
		t.Fatal("queued command changed its selected workflow page", observed)
	}

	updated, _ := model.Update(message)
	if updated != model || model.workflow.ID != "changed" {
		t.Fatal("refresh completion replaced the model or its current selection")
	}
}
