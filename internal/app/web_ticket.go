package app

import (
	"crypto/rand"
	"errors"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/webui"
)

func (p webViewProvider) writeTicketAction(operations service, request webui.WriteActionRequest) (webWriteDispatch, error) {
	switch request.Action {
	case webActionTicketCreate, webActionTicketUpdate, webActionTicketMove, webActionTicketComment:
		response := writeWebTicket(operations, request)

		return webWriteDispatch{Response: response, Handled: true}, nil
	default:
		return webWriteDispatch{Response: webui.WriteActionResponse{}, Handled: false}, nil
	}
}

func writeWebTicket(operations service, request webui.WriteActionRequest) webui.WriteActionResponse {
	mutation := core.TicketEditRequest{Key: request.ID, Revision: request.Revision, Binding: webTicketMutationBinding(request.Fields["request_id"])}
	if request.Action == webActionTicketComment {
		if strings.TrimSpace(request.Fields["content"]) == "" {
			return webValidationWrite(errors.New("ticket comment body is required"))
		}

		mutation.Comment = request.Fields["content"]
	} else {
		mutation.Body = fieldPointer(request.Fields, "description")
		mutation.Fields = webTicketPatch(request.Fields)
	}

	var (
		document core.TicketDocument
		err      error
	)

	if request.Action == webActionTicketCreate {
		document, err = operations.CreateBoardTicket(mutation)
	} else {
		document, err = operations.SaveTicketDocument(mutation)
	}

	if err != nil {
		return webTicketFailure(err)
	}

	return webui.WriteActionResponse{
		TicketRevision: document.CommittedRevision,
		Status:         webui.WriteStatusOK, Summary: "Ticket changes saved.",
		ChangedFiles: compactChangedFiles(document.Ticket.RelativePath), RedirectURL: "/board",
		Operation: webOperation(request.Action, "completed", document.Ticket.Key, document.Ticket.Status),
	}
}

func webTicketPatch(fields map[string]string) map[string]string {
	patch := map[string]string{}

	for name, value := range fields {
		if name == "description" || name == "request_id" {
			continue
		}

		if name == "synopsis" && fields["description"] != "" {
			continue
		}

		patch[name] = value
	}

	return patch
}

func webTicketMutationBinding(requestID string) core.WorkflowMutationBinding {
	if strings.TrimSpace(requestID) == "" {
		requestID = rand.Text()
	}

	return core.WorkflowMutationBinding{RequestID: requestID, Operator: "web"}
}

func webTicketFailure(err error) webui.WriteActionResponse {
	if errors.Is(err, core.ErrWorkflowConflict) {
		response := webui.WriteActionResponse{Status: webui.WriteStatusStale, Summary: "Write action rejected stale data.", BlockedReason: "The ticket changed after it was loaded. Reload before saving."}
		response.ReloadURL, response.DiscardURL = "/board", "/board"

		return response
	}

	if errors.Is(err, core.ErrWorkflowOwned) || errors.Is(err, core.ErrWorkflowAuthority) {
		return webBlockedWrite(err.Error())
	}

	return webValidationWrite(err)
}
