package webui

import (
	"context"
	"net/http"

	"forge.adrianpk.com/chr/internal/webui/htmx"
)

// DashRequest selects a display page without changing scheduling or readiness.
type DashRequest struct{ QueueOffset int }

// DashPageProvider supports bounded shared observation without changing older
// providers that expose only the default page.
type DashPageProvider interface {
	DashPageView(context.Context, DashRequest) (DashView, error)
}

func (h handler) readDashPage(r *http.Request) (DashView, error) {
	offset, err := workflowQueryNumber(r.URL.Query().Get("queue_offset"), 4096)
	if err != nil {
		return DashView{}, err
	}

	if provider, ok := h.provider.(DashPageProvider); ok {
		return provider.DashPageView(r.Context(), DashRequest{QueueOffset: int(offset)})
	}

	return h.provider.DashView(r.Context())
}

func (h handler) dashActionResult(w http.ResponseWriter, r *http.Request, request WriteActionRequest, result WriteActionResponse) bool {
	if request.Surface != "dash" || !htmx.IsRequest(r) {
		return false
	}

	if request.Action != "work.preview" && request.Action != "work.daemon.stop" {
		return false
	}

	h.renderWriteResult(w, r, result)

	return true
}

func dashRefreshQuery(r *http.Request) string {
	if value := r.URL.Query().Get("queue_offset"); value != "" {
		return "?queue_offset=" + value
	}

	return ""
}
