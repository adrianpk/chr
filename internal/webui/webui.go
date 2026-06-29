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
	Data ShellData
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
	}

	mux.HandleFunc("/", handler.shell)
	mux.HandleFunc("/bootstrap", handler.bootstrap)
	mux.HandleFunc("/healthz", healthz)
	mux.Handle("/assets/", http.StripPrefix("/assets/", handler.static))

	return mux, nil
}

type handler struct {
	renderer render.Renderer
	static   http.Handler
	data     ShellData
}

func (h handler) shell(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)

		return
	}

	if err := h.renderer.Page(w, "shell.html", h.data); err != nil {
		http.Error(w, "render shell", http.StatusInternalServerError)
	}
}

func (h handler) bootstrap(w http.ResponseWriter, r *http.Request) {
	if !htmx.IsRequest(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	if err := h.renderer.Fragment(w, "bootstrap.html", h.data); err != nil {
		http.Error(w, "render bootstrap", http.StatusInternalServerError)
	}
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte("ok\n")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
