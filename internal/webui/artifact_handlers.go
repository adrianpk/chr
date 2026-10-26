// Package webui owns the local Web UI handler and embedded assets.
//
//nolint:wsl_v5 // Handler methods keep request decoding and guarded rendering steps together.
package webui

import (
	"errors"
	"io/fs"
	"net/http"
	"net/url"

	"forge.adrianpk.com/chr/internal/webui/htmx"
)

func (h handler) catalog(family string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("case") != "" || family == "cases" && r.URL.Query().Get("new") == "workflow" {
			h.guidedUseCase(w, r)

			return
		}
		data, err := h.catalogPageData(r, family)
		if err != nil {
			http.Error(w, "load catalog", http.StatusInternalServerError)

			return
		}

		if err := h.renderer.Page(w, "shell.html", data); err != nil {
			http.Error(w, "render catalog", http.StatusInternalServerError)
		}
	}
}

func (h handler) catalogFragment(family string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !htmx.IsRequest(r) {
			http.Redirect(w, r, "/"+family, http.StatusSeeOther)

			return
		}

		data, err := h.catalogPageData(r, family)
		if err != nil {
			http.Error(w, "load catalog", http.StatusInternalServerError)

			return
		}

		if err := h.renderer.Fragment(w, "catalog_content.html", data); err != nil {
			http.Error(w, "render catalog", http.StatusInternalServerError)
		}
	}
}

func (h handler) artifactDetail(family string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := h.artifactDetailPageData(r, family)
		if err != nil {
			artifactReadFailure(w, r, err)

			return
		}

		if err := h.renderer.Page(w, "shell.html", data); err != nil {
			http.Error(w, "render artifact", http.StatusInternalServerError)
		}
	}
}

func (h handler) artifactDetailFragment(family string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !htmx.IsRequest(r) {
			location := "/" + family + "/detail?id=" + url.QueryEscape(r.URL.Query().Get("id"))
			http.Redirect(w, r, location, http.StatusSeeOther)

			return
		}

		data, err := h.artifactDetailPageData(r, family)
		if err != nil {
			artifactReadFailure(w, r, err)

			return
		}

		fragment := "artifact_detail_content.html"

		switch family {
		case "cases":
			data.CaseModal = true
			fragment = "case_modal.html"
		case "specs":
			data.SpecModal = true
			fragment = "spec_modal.html"
		case "dumps":
			data.DumpModal = true
			fragment = "dump_modal.html"
		}

		if err := h.renderer.Fragment(w, fragment, data); err != nil {
			http.Error(w, "render artifact", http.StatusInternalServerError)
		}
	}
}

func artifactReadFailure(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, fs.ErrNotExist) {
		http.NotFound(w, r)

		return
	}
	http.Error(w, "load artifact", http.StatusInternalServerError)
}
