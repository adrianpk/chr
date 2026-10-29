package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/webui"
)

func (p webViewProvider) completeTicketEditor(ctx context.Context, view webui.TicketDetailView) (webui.TicketDetailView, error) {
	view.RequestID = webTicketMutationBinding("").RequestID
	view.CommentRequestID = webTicketMutationBinding("").RequestID

	editor, err := p.service(ctx).completeTicketEditor(ticketEditor{New: view.New, NativeWorkflow: view.NativeWorkflow, Key: view.Key, Status: view.Status, Phase: view.Phase, Fields: view.Fields, ReadOnly: view.ReadOnly})
	if err != nil {
		return view, err
	}

	for _, item := range editor.Cases {
		view.Cases = append(view.Cases, webui.TicketCaseView{Slug: item.Slug, Name: item.Name})
	}

	for _, group := range editor.Groups {
		view.Groups = append(view.Groups, webTicketEditorGroup(group))
	}

	return view, nil
}

func webTicketEditorGroup(group ticketEditorGroup) webui.TicketFieldGroupView {
	result := webui.TicketFieldGroupView{Title: group.Title}
	for _, field := range group.Fields {
		value := webui.TicketFieldView{Name: field.Name, Label: field.Label, Value: field.Value, ReadOnly: field.ReadOnly, Multiline: field.Multiline}
		for _, option := range field.Options {
			value.Options = append(value.Options, webui.TicketOptionView{Value: option.Value, Label: option.Label})
		}

		result.Fields = append(result.Fields, value)
	}

	return result
}
