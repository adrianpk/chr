package webui

import (
	"net/http"
	"net/url"

	"forge.adrianpk.com/chr/internal/webui/htmx"
)

func (h handler) planActionResult(w http.ResponseWriter, r *http.Request, request WriteActionRequest, result WriteActionResponse) bool {
	if request.Surface != "case-plan" {
		return false
	}

	if htmx.IsRequest(r) && result.Status != WriteStatusOK {
		if err := h.renderer.Fragment(w, "case_plan_result.html", PageData{Write: result}); err != nil {
			http.Error(w, "render plan result", http.StatusInternalServerError)
		}

		return true
	}

	if !htmx.IsRequest(r) && result.Status == WriteStatusOK {
		h.redirectAfterWrite(w, r, "/cases/detail?id="+url.QueryEscape(request.ID))

		return true
	}

	h.renderCasePlanResult(w, r, request, result)

	return true
}

func (h handler) renderCasePlanResult(w http.ResponseWriter, r *http.Request, request WriteActionRequest, result WriteActionResponse) {
	r.URL.RawQuery = "id=" + url.QueryEscape(request.ID)

	data, err := h.artifactDetailPageData(r, "cases")
	if err != nil || data.Detail.Plan == nil {
		http.Error(w, "reload case planning", http.StatusInternalServerError)

		return
	}

	data.Write, data.CaseModal = result, request.Fields["modal"] == "true"
	if result.Status != WriteStatusOK {
		data.Detail.Plan.Revision = request.Revision
	}

	if htmx.IsRequest(r) {
		htmx.Retarget(w, "#case-plan-panel")
		htmx.Reswap(w, htmx.SwapOuter)

		if err := h.renderer.Fragment(w, "case_plan_panel.html", data); err != nil {
			http.Error(w, "render case planning", http.StatusInternalServerError)
		}
	} else if err := h.renderer.Page(w, "shell.html", data); err != nil {
		http.Error(w, "render case planning", http.StatusInternalServerError)
	}
}
