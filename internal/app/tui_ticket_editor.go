package app

import (
	"fmt"
	"slices"

	"forge.adrianpk.com/chr/internal/app/operation"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/tui"
)

func (e tuiTicketActionExecutor) executeTicketEditor(request tui.TicketActionRequest) (tui.TicketActionResult, error) {
	if request.Action == tui.TicketActionRefresh {
		board, err := e.operations.ReadTicketBoard(0)

		return tui.TicketActionResult{Tickets: board.Tickets, Message: "Tickets refreshed."}, err
	}

	if request.Action == tui.TicketActionRead {
		return e.readTicketEditor(request)
	}

	if request.OperationID == "" {
		return tui.TicketActionResult{}, fmt.Errorf("ticket operation identity is required")
	}

	mutation := workflowcore.TicketEditRequest{Key: request.Key, Revision: request.Revision, Fields: request.Fields, Body: request.Body, Binding: workflowstate.MutationBinding{RequestID: request.OperationID, Operator: "tui"}}

	var (
		document workflowcore.TicketDocument
		err      error
	)
	if request.Action == tui.TicketActionCreate {
		document, err = e.operations.CreateBoardTicket(mutation)
	} else {
		document, err = e.operations.SaveTicketDocument(mutation)
	}

	if err != nil {
		return tui.TicketActionResult{Committed: document.CommittedRevision != ""}, err
	}

	return e.savedTicketResult(document)
}

func (e tuiTicketActionExecutor) readTicketEditor(request tui.TicketActionRequest) (tui.TicketActionResult, error) {
	var (
		editor ticketEditor
		err    error
	)
	if request.Key == "" {
		editor, err = e.operations.newTicketEditor(request.CaseSlug)
	} else {
		editor, err = e.operations.readTicketEditor(request.Key)
	}

	view := tuiTicketEditor(editor)

	return tui.TicketActionResult{Editor: &view}, err
}

func (e tuiTicketActionExecutor) savedTicketResult(document workflowcore.TicketDocument) (tui.TicketActionResult, error) {
	result := tui.TicketActionResult{Committed: true, Message: "Ticket saved.", Outcome: operation.Result{ChangedFiles: compactChangedFiles(document.Ticket.RelativePath)}}

	editor, err := e.operations.completeTicketEditor(ticketEditor{Key: document.Ticket.Key, Status: string(document.Ticket.Status), Phase: document.Phase, NativeWorkflow: document.Ticket.NativeWorkflow, Fields: document.Fields, ReadOnly: document.ReadOnly, Document: document})
	if err != nil {
		return result, err
	}

	view := tuiTicketEditor(editor)
	result.Editor = &view
	board, err := e.operations.ReadTicketBoard(0)

	result.Tickets = board.Tickets
	if err == nil {
		result.Tickets = includeSavedTicket(result.Tickets, document.Ticket)
	}

	return result, err
}

func tuiTicketEditor(editor ticketEditor) tui.TicketEditorView {
	document := editor.Document
	view := tui.TicketEditorView{New: editor.New, Key: editor.Key, Revision: document.Revision, CaseSlug: editor.Fields["use_case"], Description: document.Description, Discussion: document.Discussion, Phase: document.Phase, Slice: document.Ticket.Slice, ReadOnlyReason: document.ReadOnlyReason}

	view.Fields = []tui.TicketEditorField{
		{Name: "title", Label: "Title", Value: editor.Fields["title"], ReadOnly: editor.ReadOnly["title"]},
		{Name: "description", Label: "Description", Value: document.Description, Multiline: true, ReadOnly: editor.ReadOnly["body"]},
	}
	for _, group := range editor.Groups {
		for _, field := range group.Fields {
			view.Fields = append(view.Fields, tuiTicketEditorField(field))
		}
	}

	return view
}

func tuiTicketEditorField(field ticketEditorField) tui.TicketEditorField {
	value := tui.TicketEditorField{Name: field.Name, Label: field.Label, Value: field.Value, ReadOnly: field.ReadOnly, Multiline: field.Multiline}
	for _, option := range field.Options {
		value.Options = append(value.Options, tui.TicketEditorChoice{Value: option.Value, Label: option.Label})
	}

	return value
}

func includeSavedTicket(tickets []ticketcore.Indexed, saved ticketcore.Indexed) []ticketcore.Indexed {
	if !slices.ContainsFunc(tickets, func(ticket ticketcore.Indexed) bool { return ticket.Key == saved.Key }) {
		return append(tickets, saved)
	}

	return tickets
}
