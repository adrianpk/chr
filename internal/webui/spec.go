package webui

import (
	"net/http"
	"net/url"

	"forge.adrianpk.com/chr/internal/webui/htmx"
)

func (h handler) artifactActionResult(w http.ResponseWriter, r *http.Request, request WriteActionRequest, result WriteActionResponse) bool {
	return h.quizActionResult(w, r, request, result) || h.specActionResult(w, r, request, result) || h.planActionResult(w, r, request, result)
}

func (h handler) specActionResult(w http.ResponseWriter, r *http.Request, request WriteActionRequest, result WriteActionResponse) bool {
	if request.Surface != "spec" {
		return false
	}

	if result.Status == WriteStatusOK {
		h.specWriteSuccess(w, r, request, result)

		return true
	}

	data := PageData{Write: result, SpecModal: request.Fields["modal"] == "true"}
	if data.SpecModal && result.ReloadURL != "" {
		data.Write.ReloadURL = "/specs/detail/fragment?id=" + url.QueryEscape(request.ID)
	}

	if htmx.IsRequest(r) {
		if err := h.renderer.Fragment(w, "spec_write_result.html", data); err != nil {
			http.Error(w, "render spec result", http.StatusInternalServerError)
		}

		return true
	}

	if request.Action == "spec.generate" {
		h.renderWriteResult(w, r, result)
	} else {
		h.specFailurePage(w, r, request, result)
	}

	return true
}

func (h handler) specWriteSuccess(w http.ResponseWriter, r *http.Request, request WriteActionRequest, result WriteActionResponse) {
	if !htmx.IsRequest(r) || request.Fields["modal"] != "true" {
		h.redirectAfterWrite(w, r, "/specs")

		return
	}

	htmx.Retarget(w, "#modal-root")
	htmx.Reswap(w, htmx.SwapInner)

	if request.Action != "spec.generate" {
		w.WriteHeader(http.StatusOK)

		return
	}

	r.URL.RawQuery = "id=" + url.QueryEscape(result.Operation.ID)

	data, err := h.artifactDetailPageData(r, "specs")
	if err != nil {
		http.Error(w, "Spec saved. Reload Specs to open it.", http.StatusInternalServerError)

		return
	}

	data.SpecModal, data.Write = true, result
	if err := h.renderer.Fragment(w, "spec_modal.html", data); err != nil {
		http.Error(w, "Spec saved. Reload Specs to open it.", http.StatusInternalServerError)
	}
}

func (h handler) specFailurePage(w http.ResponseWriter, r *http.Request, request WriteActionRequest, result WriteActionResponse) {
	r.URL.RawQuery = "id=" + url.QueryEscape(request.ID)

	data, err := h.artifactDetailPageData(r, "specs")
	if err != nil || data.Detail.Spec == nil {
		http.Error(w, "reload spec editor", http.StatusInternalServerError)

		return
	}

	data.Detail.Title, data.Detail.Content, data.Detail.Revision = request.Fields["title"], request.Fields["content"], request.Revision
	data.Detail.Spec.Draft = request.Fields["draft"] == "true"

	data.Detail.Spec.RequestID, data.Write = request.Fields["request_id"], result
	if err := h.renderer.Page(w, "shell.html", data); err != nil {
		http.Error(w, "render spec editor", http.StatusInternalServerError)
	}
}
