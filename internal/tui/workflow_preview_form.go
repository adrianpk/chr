package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/app/operation"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

func (m *Model) openWorkPreviewForm() {
	if m.observability.Authority != workspaceconfig.WorkStateAuthorityTickets {
		m.statusMessage = "Bounded workflow preview requires tickets authority."

		return
	}

	ticket, _ := m.selectedTicketView()
	form := textFormState{Title: "Preview Work Scope", SubmitLabel: "Preview", Submit: (*Model).submitWorkPreview, WorkRequest: WorkActionRequest{TicketKey: ticket.Key, TicketRevision: ticket.Revision, Workflow: &WorkflowActionRequest{Preview: true}}}
	form.Details = []string{"Read-only preview; confirmation admits at most one checkpoint.", "Ticket is used for ticket/slice/set; Case is used for case scope.", "Filters and rank order apply only to loose work; other scopes retain hierarchy order."}
	form.Fields = []textFormField{workChoiceField("Scope", []string{"ticket", "slice", "slice_set", "use_case", "loose"}, "ticket"), newTextFormField("Ticket key", false, 128), newTextFormField("Case", false, 128), newTextFormField("Loose priority (optional)", false, 128), newTextFormField("Loose severity (optional)", false, 128), newTextFormField("Loose project area (optional)", false, 128), workChoiceField("Loose order", []string{"", "created_at", "priority", "severity", "priority_severity", "key"}, "")}
	form.Fields[1].Input.SetValue(ticket.Key)

	for _, item := range m.cases {
		form.Fields[2].Choices = append(form.Fields[2].Choices, textFormChoice{Label: item.Name, Value: item.Slug})
		if item.Slug == m.caseContext.CaseSlug {
			form.Fields[2].Choice = len(form.Fields[2].Choices) - 1
		}
	}

	m.openTextForm(form)
}

func workChoiceField(label string, values []string, current string) textFormField {
	field := newTextFormField(label, false, 128)
	for _, value := range values {
		field.Choices = append(field.Choices, textFormChoice{Label: value, Value: value})
		if value == current {
			field.Choice = len(field.Choices) - 1
		}
	}

	return field
}

func (m *Model) submitWorkPreview() tea.Cmd {
	fields := m.textForm.Fields
	request := m.textForm.WorkRequest
	native := *request.Workflow
	request.Workflow = &native
	extent := fields[0].value()
	native.Selection = workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: extent}, Filters: workflowstate.TicketFilters{Priorities: workFilterValue(fields[3].value()), Severities: workFilterValue(fields[4].value()), ProjectAreas: workFilterValue(fields[5].value())}, Order: workflowstate.TicketOrder(fields[6].value())}

	request.Action = workExtentAction(extent)
	if extent == "use_case" {
		native.Selection.Selector.Identity = fields[2].value()
	}

	if extent == "loose" {
		native.Selection.Statuses = []string{"ready_for_dev"}
	}

	if extent == "ticket" || extent == "slice" || extent == "slice_set" {
		if !m.captureWorkPreviewTicket(&request, fields[1].value()) {
			return nil
		}
	} else {
		request.TicketKey, request.TicketRevision = "", ""
	}

	return m.startWorkAction(request)
}

func (m *Model) captureWorkPreviewTicket(request *WorkActionRequest, key string) bool {
	if key == request.TicketKey && request.TicketRevision != "" {
		return true
	}

	if m.ticketExecutor != nil {
		result, err := m.ticketExecutor.ExecuteTicketAction(TicketActionRequest{Action: TicketActionRead, Key: key})
		if err == nil && result.Editor != nil && result.Editor.Key == key && key != "" {
			request.TicketKey, request.TicketRevision = key, result.Editor.Revision

			return true
		}
	}

	m.recordOperation(operation.Result{Status: operation.ValidationFailed, Summary: "Choose a concrete native ticket before previewing."})

	return false
}

func workFilterValue(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	return []string{strings.TrimSpace(value)}
}

func workExtentAction(extent string) WorkActionID {
	for _, action := range []WorkActionID{WorkActionWorkflowTicket, WorkActionWorkflowSlice, WorkActionWorkflowSet, WorkActionWorkflowCase, WorkActionWorkflowLoose} {
		if nativeWorkExtent(action) == extent {
			return action
		}
	}

	return ""
}
