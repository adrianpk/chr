// Package webui owns the optional local Web UI handler and embedded assets.
package webui

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"

	"forge.adrianpk.com/chc/internal/webui/htmx"
	"forge.adrianpk.com/chc/internal/webui/render"
	"forge.adrianpk.com/chc/internal/webui/ui"
)

//go:embed assets
var assetsFS embed.FS

// ShellData contains the presentation data needed by the shell template.
type ShellData struct {
	Title          string
	ProjectDir     string
	ProjectID      string
	URL            string
	Host           string
	Port           int
	RequireToken   string
	RefreshSeconds int
	Navigation     []ui.NavLink
}

// HandlerOptions configures the Web UI HTTP handler.
type HandlerOptions struct {
	Data     ShellData
	Provider DataProvider
}

// NewHandler creates the Web UI HTTP handler.
func NewHandler(opts HandlerOptions) (http.Handler, error) {
	templates, err := render.New(
		assetsFS,
		"assets/templates/*.html",
	)
	if err != nil {
		return nil, err
	}

	staticFiles, err := fs.Sub(assetsFS, "assets/static")
	if err != nil {
		return nil, fmt.Errorf("load web static assets: %w", err)
	}

	mux := http.NewServeMux()

	handler := handler{
		renderer: templates,
		static:   http.FileServer(http.FS(staticFiles)),
		data:     opts.Data,
		provider: opts.Provider,
	}
	if handler.provider == nil {
		handler.provider = newStaticDataProvider(opts.Data)
	}

	mux.HandleFunc("/", handler.shell)
	mux.HandleFunc("/project", handler.project)
	mux.HandleFunc("/project/fragment", handler.projectFragment)
	mux.HandleFunc("/board", handler.board)
	mux.HandleFunc("/board/fragment", handler.boardFragment)
	mux.HandleFunc("/dash", handler.dash)
	mux.HandleFunc("/dash/fragment", handler.dashFragment)

	for _, family := range []string{"cases", "dumps", "quizzes", "specs", "plans"} {
		mux.HandleFunc("/"+family, handler.catalog(family))
		mux.HandleFunc("/"+family+"/fragment", handler.catalogFragment(family))
		mux.HandleFunc("/"+family+"/detail", handler.artifactDetail(family))
		mux.HandleFunc("/"+family+"/detail/fragment", handler.artifactDetailFragment(family))
	}

	mux.HandleFunc("/actions", handler.writeAction)
	mux.HandleFunc("/bootstrap", handler.projectFragment)
	mux.HandleFunc("/healthz", healthz)
	mux.Handle("/assets/", http.StripPrefix("/assets/", handler.static))

	return mux, nil
}

type handler struct {
	renderer render.Renderer
	static   http.Handler
	data     ShellData
	provider DataProvider
}

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

	if err := h.renderer.Fragment(w, "board_content.html", data); err != nil {
		http.Error(w, "render board", http.StatusInternalServerError)
	}
}

func (h handler) dash(w http.ResponseWriter, r *http.Request) {
	data, err := h.dashPageData(r)
	if err != nil {
		http.Error(w, "load dash", http.StatusInternalServerError)

		return
	}

	if err := h.renderer.Page(w, "shell.html", data); err != nil {
		http.Error(w, "render dash", http.StatusInternalServerError)
	}
}

func (h handler) dashFragment(w http.ResponseWriter, r *http.Request) {
	if !htmx.IsRequest(r) {
		http.Redirect(w, r, "/dash", http.StatusSeeOther)

		return
	}

	data, err := h.dashPageData(r)
	if err != nil {
		http.Error(w, "load dash", http.StatusInternalServerError)

		return
	}

	if err := h.renderer.Fragment(w, "dash_content.html", data); err != nil {
		http.Error(w, "render dash", http.StatusInternalServerError)
	}
}

func (h handler) catalog(family string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			http.Error(w, "load artifact", http.StatusInternalServerError)

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
			http.Error(w, "load artifact", http.StatusInternalServerError)

			return
		}

		if err := h.renderer.Fragment(w, "artifact_detail_content.html", data); err != nil {
			http.Error(w, "render artifact", http.StatusInternalServerError)
		}
	}
}

func (h handler) writeAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	result, err := h.provider.WriteAction(r.Context(), writeActionRequest(r))
	if err != nil {
		http.Error(w, "write action", http.StatusInternalServerError)

		return
	}

	if result.Status == WriteStatusOK && result.RedirectURL != "" {
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
	if result.Status == WriteStatusOK {
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
		if err := h.renderer.Fragment(w, "write_result.html", data); err != nil {
			http.Error(w, "render write result", http.StatusInternalServerError)
		}

		return
	}

	if err := h.renderer.Page(w, "shell.html", data); err != nil {
		http.Error(w, "render write result", http.StatusInternalServerError)
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
		Fields:   fields,
	}
}

func writeActionReservedField(key string) bool {
	switch key {
	case "action", "family", "id", "revision", "confirm":
		return true
	default:
		return false
	}
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte("ok\n")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h handler) projectPageData(r *http.Request) (PageData, error) {
	project, err := h.provider.ProjectView(r.Context())
	if err != nil {
		return PageData{}, err
	}

	return PageData{
		Shell:   h.shellData("project"),
		View:    "project",
		Title:   "Project",
		Project: project,
	}, nil
}

func (h handler) boardPageData(r *http.Request) (PageData, error) {
	board, err := h.provider.BoardView(r.Context())
	if err != nil {
		return PageData{}, err
	}

	return PageData{
		Shell: h.shellData("board"),
		View:  "board",
		Title: "Board",
		Board: board,
	}, nil
}

func (h handler) dashPageData(r *http.Request) (PageData, error) {
	dash, err := h.provider.DashView(r.Context())
	if err != nil {
		return PageData{}, err
	}

	return PageData{
		Shell: h.shellData("dash"),
		View:  "dash",
		Title: "Dash",
		Dash:  dash,
	}, nil
}

func (h handler) catalogPageData(r *http.Request, family string) (PageData, error) {
	catalog, err := h.provider.CatalogView(r.Context(), family)
	if err != nil {
		return PageData{}, err
	}

	return PageData{
		Shell:   h.shellData(family),
		View:    "catalog",
		Title:   catalog.Title,
		Catalog: catalog,
	}, nil
}

func (h handler) artifactDetailPageData(r *http.Request, family string) (PageData, error) {
	detail, err := h.provider.ArtifactDetailView(r.Context(), ArtifactDetailRequest{
		Family: family,
		ID:     r.URL.Query().Get("id"),
	})
	if err != nil {
		return PageData{}, err
	}

	return PageData{
		Shell:  h.shellData(family),
		View:   "artifact_detail",
		Title:  detail.Title,
		Detail: detail,
	}, nil
}

func (h handler) shellData(active string) ShellData {
	data := h.data

	data.Navigation = make([]ui.NavLink, 0, len(h.data.Navigation))

	for _, link := range h.data.Navigation {
		link.Active = navLinkActive(active, link.Href)
		data.Navigation = append(data.Navigation, link)
	}

	return data
}

func navLinkActive(active, href string) bool {
	switch active {
	case "project":
		return href == "/" || href == "/project"
	case "board":
		return href == "/board"
	case "dash":
		return href == "/dash"
	case "cases", "dumps", "quizzes", "specs", "plans":
		return href == "/"+active
	default:
		return false
	}
}
