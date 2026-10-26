package tui

import (
	"testing"
	"time"

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

// Background IO returns a message while all model mutations remain on the event
// loop. Rendering and navigation can continue before the message is applied.
func TestModelAsyncRefresh(t *testing.T) {
	started := make(chan AgentWorkRefreshRequest, 1)
	release := make(chan struct{})

	defer func() {
		select {
		case <-release:
		default:
			close(release)
		}
	}()

	model := NewModel(Options{AgentWorkExecutor: modelRefreshExecutor(func(request AgentWorkRefreshRequest) (AgentWorkRefreshResult, error) {
		started <- request

		<-release

		return AgentWorkRefreshResult{Observability: ObservabilityView{DaemonStatus: "completed refresh"}}, nil
	})})
	command := model.startAgentWorkRefresh()

	messages := make(chan tea.Msg, 1)
	go func() { messages <- command() }()

	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("background refresh did not start")
	}

	for i := range 20 {
		model.Update(tea.WindowSizeMsg{Width: 80 + i, Height: 30})
		model.Update(tea.KeyMsg{Type: tea.KeyTab})

		_ = model.View()
	}

	if model.observability.DaemonStatus != "" {
		t.Fatal("background IO mutated the model without a message")
	}

	close(release)

	select {
	case message := <-messages:
		model.Update(message)
	case <-time.After(5 * time.Second):
		t.Fatal("background refresh did not finish")
	}

	if model.observability.DaemonStatus != "completed refresh" {
		t.Fatal("event loop did not apply the completed observation")
	}
}
