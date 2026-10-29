package tui

import (
	"crypto/rand"
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/app/operation"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

func (m *Model) openWorkflowControlForm(action WorkActionID) {
	if m.observability.Authority != workspaceconfig.WorkStateAuthorityTickets || m.workflow.ID == "" {
		m.statusMessage = "Inspect a retained native workflow first."

		return
	}

	if !m.workflowActionAvailable(action) {
		return
	}

	current := m.workflow
	native := &WorkflowActionRequest{ID: current.ID, ExpectedRevision: current.Revision, BindingHash: current.ApprovalHash, Integration: current.Integration, Questions: current.Questions, DecisionRevision: current.DecisionRevision, DecisionHash: current.DecisionHash, RequestKey: rand.Text(), Preview: true}
	form := textFormState{Title: "Extend Workflow Allowance", SubmitLabel: "Preview", Submit: (*Model).submitWorkflowControl, WorkRequest: WorkActionRequest{Action: action, Workflow: native}}

	form.Details = []string{fmt.Sprintf("Workflow %s revision %d", current.ID, current.Revision), "The owner rechecks the displayed candidate and allowance; consumed work is preserved."}
	if action == WorkActionWorkflowGrant {
		form.Fields = []textFormField{workChoiceField("Kind", current.GrantKinds, "submission"), newTextFormField("Amount", false, 20), newTextFormField("Actor", false, 128), newTextFormField("Reason", true, 128)}
	} else {
		form.Title = "Answer Integration Decision"
		form.Fields = workflowDecisionFormFields(current)
	}

	m.openTextForm(form)
}

func workflowDecisionFormFields(current WorkflowView) []textFormField {
	question := newTextFormField("Question", false, 128)

	var lines []string

	for _, item := range current.Questions {
		question.Choices = append(question.Choices, textFormChoice{Label: item.Issue, Value: item.ID})

		lines = append(lines, item.ID+": "+item.Issue+"; authority: "+item.Authority)
		for index, option := range item.Options {
			lines = append(lines, fmt.Sprintf("  %d. %s", index+1, option))
		}
	}

	history := newTextFormField("Retained questions and options", true, 1<<20)
	history.ReadOnly = "Use the question selector and option number below."
	history.Input.SetValue(strings.Join(lines, "\n"))

	option := newTextFormField("Option number", false, 6)
	option.Input.SetValue("1")

	fields := []textFormField{history, question, option, workChoiceField("Action", []string{"resolve", "defer"}, "resolve"), newTextFormField("Actor", false, 128), newTextFormField("Reason", true, 8192), newTextFormField("Original commit (optional)", false, 64), newTextFormField("Fix category (optional)", false, 128), newTextFormField("Fix title (optional)", false, 8192), newTextFormField("Fix detail (optional)", true, 8192)}

	return fields
}

func (m *Model) submitWorkflowControl() tea.Cmd {
	request := m.textForm.WorkRequest
	native := *request.Workflow
	request.Workflow = &native

	native.Fields = map[string]string{}
	if request.Action == WorkActionWorkflowGrant {
		for index, key := range []string{"kind", "amount", "actor", "reason"} {
			native.Fields[key] = m.textForm.Fields[index].value()
		}
	} else if !m.captureWorkflowDecision(&native) {
		return nil
	}

	return m.startWorkAction(request)
}

func (m *Model) captureWorkflowDecision(native *WorkflowActionRequest) bool {
	fields := m.textForm.Fields

	number, err := strconv.Atoi(fields[2].value())
	for _, question := range native.Questions {
		if question.ID != fields[1].value() {
			continue
		}

		if err != nil || number < 1 || number > len(question.Options) {
			break
		}

		native.Fields["decision_id"], native.Fields["choice"], native.Fields["authority"] = question.ID, question.Options[number-1], question.Authority
		for index, key := range []string{"answer_action", "actor", "reason", "original_commit", "fix_category", "fix_title", "fix_detail"} {
			native.Fields[key] = fields[index+3].value()
		}

		return true
	}

	m.recordOperation(operation.Result{Status: operation.ValidationFailed, Summary: "Choose a retained question and one of its displayed option numbers."})

	return false
}
