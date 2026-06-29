// Package webui owns the optional local Web UI handler and embedded assets.
package webui

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

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
	default:
		return false
	}
}
