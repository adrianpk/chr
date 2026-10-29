package tui

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

type interactionContext struct {
	Generation        uint64
	View              View
	Case, Source, Run string
	Draft             [32]byte
}

type operationBinding struct {
	Sequence uint64
	Family   string
	Context  interactionContext
}

func (m *Model) currentInteraction() interactionContext {
	draft := fmt.Sprintf("%#v", []any{m.projectEdit.Active, m.projectEdit.Field, m.projectEdit.Input.Value(), m.caseEdit.Active, m.caseEdit.Field, m.caseEdit.Input.Value(), m.dumpEdit.Active, m.dumpEdit.Path, m.dumpEdit.Input.Value(), m.quizAnswer.Active, m.quizAnswer.Binding, m.quizAnswer.Input.Value(), m.guided.Active, m.guided.Sequence, m.harnessSelection.Active, m.harnessSequence})

	return interactionContext{Generation: m.draftSequence, View: m.SelectedView(), Case: m.caseContext.CaseSlug, Source: m.currentOperationTarget(), Draft: sha256.Sum256([]byte(draft))}
}

func (m *Model) beginOperation(family string) operationBinding {
	m.operationSequence++

	binding := operationBinding{Sequence: m.operationSequence, Family: family, Context: m.operationInteraction(family)}
	if m.pendingOperations == nil {
		m.pendingOperations = make(map[string]operationBinding)
	}

	m.pendingOperations[family] = binding

	return binding
}

func (m *Model) finishOperation(binding operationBinding) (bool, bool) {
	current, ok := m.pendingOperations[binding.Family]
	if !ok || current.Sequence != binding.Sequence {
		return false, false
	}

	delete(m.pendingOperations, binding.Family)

	return true, binding.Context == m.operationInteraction(binding.Family)
}

func (m *Model) currentOperationTarget() string {
	switch m.SelectedView() {
	case ViewDash, ViewBoard, ViewTickets:
		return m.currentTicketTarget()

	case ViewQuizzes:
		if quiz, ok := m.selectedQuizView(); ok {
			return fmt.Sprintf("%s:%v", quiz.Scope.CaseSlug, quiz.Binding)
		}
	case ViewSpecs:
		return m.selectedSpecScope().CaseSlug
	case ViewPlans:
		return m.selectedPlanScope().CaseSlug
	case ViewCases:
		if metadata, ok := m.selectedCaseMetadata(); ok {
			return metadata.Slug + ":" + metadata.SourceHash
		}
	}

	return m.selectedSourcePath()
}

func (m *Model) operationInteraction(family string) interactionContext {
	context := m.currentInteraction()
	if strings.HasPrefix(family, "work:") {
		context.Run = m.workflow.ID
		if m.SelectedView() == ViewDash {
			context.Run += fmt.Sprintf(":%d:%#v", m.queueSequence, m.queueRequest)
		}
	}

	return context
}

func (m *Model) currentTicketTarget() string {
	ticket, ok := m.selectedTicketView()
	if m.SelectedView() == ViewDash {
		ticket, ok = m.selectedDashTicketView()
	}

	if ok {
		return ticket.Key + ":" + ticket.Revision
	}

	return ""
}
