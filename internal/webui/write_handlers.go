// Package webui owns the local Web UI handler and embedded assets.
//
//nolint:wsl_v5 // Handler methods keep request decoding and guarded rendering steps together.
package webui

import (
	"net/http"
	"net/url"

	"forge.adrianpk.com/chr/internal/webui/htmx"
)

func (h handler) writeAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	request := writeActionRequest(r)
	if request.Action == HarnessSelectionAction {
		http.Error(w, "use the project harness form", http.StatusBadRequest)

		return
	}

	result, err := h.provider.WriteAction(r.Context(), request)
	if err != nil {
		http.Error(w, "write action", http.StatusInternalServerError)

		return
	}

	if h.dashActionResult(w, r, request, result) || h.artifactActionResult(w, r, request, result) {
		return
	}

	if shouldCloseModalAfterWrite(r, request, result) {
		htmx.Retarget(w, "#modal-root")
		htmx.Reswap(w, htmx.SwapInner)
		w.WriteHeader(http.StatusOK)

		return
	}

	if result.Status == WriteStatusOK && result.RedirectURL != "" && !result.ProjectionStale {
		h.redirectAfterWrite(w, r, result.RedirectURL)

		return
	}

	h.renderWriteResult(w, r, result)
}

func (h handler) redirectAfterWrite(w http.ResponseWriter, r *http.Request, location string) {
	if htmx.IsRequest(r) {
		w.Header().Set(htmx.HeaderRedirect, location)
		w.WriteHeader(http.StatusNoContent)

		return
	}

	http.Redirect(w, r, location, http.StatusSeeOther)
}

func (h handler) renderWriteResult(w http.ResponseWriter, r *http.Request, result WriteActionResponse) {
	status := http.StatusBadRequest
	// Expected write failures need a successful fragment response for htmx to display them.
	if result.Status == WriteStatusOK || htmx.IsRequest(r) {
		status = http.StatusOK
	}

	w.WriteHeader(status)

	data := PageData{
		Shell: h.shellData("project"),
		View:  "write_result",
		Title: "Write Result",
		Write: result,
	}
	if htmx.IsRequest(r) {
		template := writeResultTemplate(r, &data)

		if err := h.renderer.Fragment(w, template, data); err != nil {
			http.Error(w, "render write result", http.StatusInternalServerError)
		}

		return
	}

	if err := h.renderer.Page(w, "shell.html", data); err != nil {
		http.Error(w, "render write result", http.StatusInternalServerError)
	}
}

func writeResultTemplate(r *http.Request, data *PageData) string {
	if r.PostFormValue("action") == "dump.save" {
		data.DumpModal = r.PostFormValue("surface") == "modal"
		if data.DumpModal && data.Write.ReloadURL != "" {
			data.Write.ReloadURL = "/dumps/detail/fragment?id=" + url.QueryEscape(r.PostFormValue("id"))
		}

		return "dump_save_result.html"
	}

	if r.PostFormValue("surface") == "modal" && r.PostFormValue("action") != "dump.create" && r.PostFormValue("action") != "case.create" || r.URL.Path == "/board/ticket/comment" {
		return "write_result_modal.html"
	}

	return "write_result.html"
}

func writeActionRequest(r *http.Request) WriteActionRequest {
	if err := r.ParseForm(); err != nil {
		return WriteActionRequest{Action: "invalid", Confirm: false}
	}

	fields := map[string]string{}

	for key, values := range r.PostForm {
		if writeActionReservedField(key) || len(values) == 0 {
			continue
		}

		fields[key] = values[0]
	}

	return WriteActionRequest{
		Action:   r.PostFormValue("action"),
		Family:   r.PostFormValue("family"),
		ID:       r.PostFormValue("id"),
		Revision: r.PostFormValue("revision"),
		Confirm:  r.PostFormValue("confirm") == "true",
		Surface:  r.PostFormValue("surface"),
		Fields:   fields,
	}
}

func writeActionReservedField(key string) bool {
	switch key {
	case "action", "family", "id", "revision", "confirm", "surface":
		return true
	default:
		return false
	}
}

func shouldCloseModalAfterWrite(r *http.Request, request WriteActionRequest, result WriteActionResponse) bool {
	if !htmx.IsRequest(r) || result.Status != WriteStatusOK || request.Surface != "modal" || result.ProjectionStale {
		return false
	}

	switch request.Action {
	case "case.create", "case.update", "dump.create", "dump.save", "dump.append":
		return true
	default:
		return false
	}
}
