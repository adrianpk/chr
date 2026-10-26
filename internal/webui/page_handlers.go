// Package webui owns the local Web UI handler and embedded assets.
//
//nolint:wsl_v5 // Handler methods keep request decoding and guarded rendering steps together.
package webui

import (
	"errors"
	"net/http"

	"forge.adrianpk.com/chr/internal/webui/htmx"
)

func (h handler) shell(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)

		return
	}

	h.renderProject(w, r)
}

func (h handler) project(w http.ResponseWriter, r *http.Request) {
	h.renderProject(w, r)
}

func (h handler) projectFragment(w http.ResponseWriter, r *http.Request) {
	if !htmx.IsRequest(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	data, err := h.projectPageData(r)
	if err != nil {
		http.Error(w, "load project", http.StatusInternalServerError)

		return
	}

	if err := h.renderer.Fragment(w, "project_content.html", data); err != nil {
		http.Error(w, "render project", http.StatusInternalServerError)
	}
}

func (h handler) board(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("case") != "" {
		h.guidedUseCase(w, r)

		return
	}
	data, err := h.boardPageData(r)
	if err != nil {
		http.Error(w, "load board", http.StatusInternalServerError)

		return
	}

	if err := h.renderer.Page(w, "shell.html", data); err != nil {
		http.Error(w, "render board", http.StatusInternalServerError)
	}
}

func (h handler) boardFragment(w http.ResponseWriter, r *http.Request) {
	if !htmx.IsRequest(r) {
		http.Redirect(w, r, "/board", http.StatusSeeOther)

		return
	}

	data, err := h.boardPageData(r)
	if err != nil {
		http.Error(w, "load board", http.StatusInternalServerError)

		return
	}

	w.Header().Set("ETag", `"`+data.Revision+`"`)
	w.Header().Set("Cache-Control", "no-cache")
	if r.Header.Get("If-None-Match") == `"`+data.Revision+`"` {
		w.WriteHeader(http.StatusNotModified)

		return
	}

	if err := h.renderer.Fragment(w, "board_content.html", data); err != nil {
		http.Error(w, "render board", http.StatusInternalServerError)
	}
}

func (h handler) ticketDetail(w http.ResponseWriter, r *http.Request) {
	data, err := h.ticketDetailPageData(r)
	if err != nil {
		http.Error(w, "load ticket", http.StatusInternalServerError)

		return
	}

	if err := h.renderer.Fragment(w, "ticket_modal.html", data); err != nil {
		http.Error(w, "render ticket", http.StatusInternalServerError)
	}
}

func (h handler) ticketComment(w http.ResponseWriter, r *http.Request) {
	h.ticketWrite(w, r, true)
}

func (h handler) dash(w http.ResponseWriter, r *http.Request) {
	data, err := h.dashPageData(r)
	if errors.Is(err, errDashRequest) {
		http.Error(w, "invalid Dash queue page", http.StatusBadRequest)

		return
	}
	if err != nil {
		data = h.dashPageProjection(r, DashView{Unavailable: true, Revision: "unavailable"})
	}
	if err = h.renderer.Page(w, "shell.html", data); err != nil {
		http.Error(w, "render dash", http.StatusInternalServerError)
	}
}

func (h handler) dashFragment(w http.ResponseWriter, r *http.Request) {
	if !htmx.IsRequest(r) {
		http.Redirect(w, r, "/dash"+dashRefreshQuery(r), http.StatusSeeOther)

		return
	}
	data, err := h.dashPageData(r)
	if errors.Is(err, errDashRequest) {
		http.Error(w, "invalid Dash queue page", http.StatusBadRequest)

		return
	}
	if err != nil {
		w.Header().Set("Cache-Control", "no-store")
		http.Error(w, "Work observation is unavailable", http.StatusServiceUnavailable)

		return
	}
	if dashObservationHeaders(w, r, data.Dash) {
		return
	}
	if err = h.renderer.Fragment(w, "dash_content.html", data); err != nil {
		http.Error(w, "render dash", http.StatusInternalServerError)
	}
}

func (h handler) renderProject(w http.ResponseWriter, r *http.Request) {
	data, err := h.projectPageData(r)
	if err != nil {
		http.Error(w, "load project", http.StatusInternalServerError)

		return
	}

	if err := h.renderer.Page(w, "shell.html", data); err != nil {
		http.Error(w, "render project", http.StatusInternalServerError)
	}
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte("ok\n")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
