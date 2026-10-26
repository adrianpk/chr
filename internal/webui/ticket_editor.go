package webui

import (
	"net/http"

	"forge.adrianpk.com/chr/internal/webui/htmx"
)

func (h handler) ticketSave(w http.ResponseWriter, r *http.Request) {
	h.ticketWrite(w, r, false)
}

func (h handler) ticketWrite(w http.ResponseWriter, r *http.Request, comment bool) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	request := writeActionRequest(r)
	if !ticketEditorAction(request.Action, comment) {
		http.Error(w, "unsupported ticket action", http.StatusBadRequest)

		return
	}

	result, err := h.provider.WriteAction(r.Context(), request)
	if err != nil {
		http.Error(w, "write ticket", http.StatusInternalServerError)

		return
	}

	result.Operation.Kind = request.Action
	if result.Operation.ID == "" {
		result.Operation.ID = request.ID
	}

	writtenTicket, err := h.ticketWriteData(r, request, result, comment)
	data, fragment := writtenTicket.Page, writtenTicket.Template

	if err != nil {
		http.Error(w, "reload ticket comments", http.StatusInternalServerError)

		return
	}

	if comment && result.Status != WriteStatusOK && htmx.IsRequest(r) {
		htmx.Retarget(w, "#ticket-comment-result")
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := h.renderer.Fragment(w, fragment, data); err != nil {
		http.Error(w, "render ticket result", http.StatusInternalServerError)
	}
}

func (h handler) ticketWriteData(r *http.Request, request WriteActionRequest, result WriteActionResponse, comment bool) (ticketWriteView, error) {
	if !comment || result.Status != WriteStatusOK {
		return ticketWriteView{Page: PageData{Write: result}, Template: "ticket_write_result.html"}, nil
	}

	data, err := h.ticketDetailPageData(ticketDetailRequest(r, request.ID))
	if err != nil {
		return ticketWriteView{Page: PageData{}, Template: ""}, err
	}

	data.Ticket.PreviousRevision = request.Revision
	data.Ticket.Revision = result.TicketRevision
	data.Write = result

	return ticketWriteView{Page: data, Template: "ticket_discussion.html"}, nil
}

func ticketEditorAction(action string, comment bool) bool {
	if comment {
		return action == "ticket.comment"
	}

	switch action {
	case "ticket.create", "ticket.update", "ticket.move":
		return true
	default:
		return false
	}
}

type ticketWriteView struct {
	Page     PageData
	Template string
}
