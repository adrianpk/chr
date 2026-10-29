package app

import (
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type (
	ticketEditorOption struct{ Value, Label string }
	ticketEditorCase   struct{ Slug, Name string }
	ticketEditorField  struct {
		Name, Label, Value, ReadOnly string
		Multiline                    bool
		Options                      []ticketEditorOption
	}
)
type ticketEditorGroup struct {
	Title  string
	Fields []ticketEditorField
}
type ticketEditor struct {
	New, NativeWorkflow bool
	Key, Status, Phase  string
	Document            workflowcore.TicketDocument
	Fields, ReadOnly    map[string]string
	Cases               []ticketEditorCase
	Groups              []ticketEditorGroup
}

func (s service) completeTicketEditor(editor ticketEditor) (ticketEditor, error) {
	cases, err := authoringcore.NewCaseMetadataService(s.projectContext.ProjectDir).LoadBounded(s.ctx, workflowstate.RecordLimit)
	if err != nil {
		return editor, err
	}

	for _, item := range cases {
		editor.Cases = append(editor.Cases, ticketEditorCase{Slug: item.Slug, Name: item.Name})
	}

	editor.Groups = ticketEditorGroups(editor)

	return editor, nil
}

func (s service) readTicketEditor(key string) (ticketEditor, error) {
	document, err := s.ReadTicketDocument(key)
	if err != nil {
		return ticketEditor{}, err
	}

	return s.completeTicketEditor(ticketEditor{Key: document.Ticket.Key, Status: string(document.Ticket.Status), Phase: document.Phase, NativeWorkflow: document.Ticket.NativeWorkflow, Fields: document.Fields, ReadOnly: document.ReadOnly, Document: document})
}

func (s service) newTicketEditor(caseSlug string) (ticketEditor, error) {
	return s.completeTicketEditor(ticketEditor{New: true, Fields: map[string]string{"status": "backlog", "priority": "normal", "severity": "unclassified", "type": "task", "use_case": caseSlug}, ReadOnly: map[string]string{}})
}
