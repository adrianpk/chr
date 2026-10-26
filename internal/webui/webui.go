// Package webui owns the local Web UI handler and embedded assets.
//
//nolint:wsl_v5 // Handler methods keep request decoding and guarded rendering steps together.
package webui

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"

	"forge.adrianpk.com/chr/internal/webui/render"
	"forge.adrianpk.com/chr/internal/webui/ui"
)

//go:embed assets
var assetsFS embed.FS

// ShellData contains the presentation data needed by the shell template.
type ShellData struct {
	Title            string
	ProjectDir       string
	ProjectID        string
	ProjectName      string
	URL              string
	Host             string
	Port             int
	RequireToken     string
	RefreshSeconds   int
	Navigation       []ui.NavLink
	LoginEnabled     bool
	HarnessesEnabled bool
}

// HandlerOptions configures the Web UI HTTP handler.
type HandlerOptions struct {
	Data        ShellData
	Provider    DataProvider
	Access      AccessPolicy
	LogRequests io.Writer
	Login       *LoginCredentials
}

// AccessPolicy describes request access controls for the local Web UI.
type AccessPolicy struct {
	RequireToken   bool
	Token          string
	Origin         string
	AllowedOrigins []string
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
		renderer:    templates,
		static:      http.FileServer(http.FS(staticFiles)),
		data:        opts.Data,
		provider:    opts.Provider,
		access:      opts.Access,
		logRequests: opts.LogRequests,
	}
	if handler.provider == nil {
		handler.provider = newStaticDataProvider(opts.Data)
	}

	if err := handler.configureLogin(opts.Login); err != nil {
		return nil, err
	}

	mux.HandleFunc("/", handler.shell)
	mux.HandleFunc("/project", handler.project)
	mux.HandleFunc("/project/fragment", handler.projectFragment)
	mux.HandleFunc("/project/harnesses", handler.harnesses)
	mux.HandleFunc("/board", handler.board)
	mux.HandleFunc("/board/fragment", handler.boardFragment)
	mux.HandleFunc("/board/ticket", handler.ticketDetail)
	mux.HandleFunc("/board/ticket/comment", handler.ticketComment)
	mux.HandleFunc("/board/ticket/save", handler.ticketSave)
	mux.HandleFunc("/dash", handler.dash)
	mux.HandleFunc("/dash/fragment", handler.dashFragment)

	for _, family := range []string{"cases", "dumps", "quizzes", "specs", "plans"} {
		mux.HandleFunc("/"+family, handler.catalog(family))
		mux.HandleFunc("/"+family+"/fragment", handler.catalogFragment(family))
		mux.HandleFunc("/"+family+"/detail", handler.artifactDetail(family))
		mux.HandleFunc("/"+family+"/detail/fragment", handler.artifactDetailFragment(family))
	}

	handler.registerGuidedWorkflowRoutes(mux)
	mux.HandleFunc("/bootstrap", handler.projectFragment)
	mux.HandleFunc("/healthz", healthz)
	mux.HandleFunc("/logout", handler.logout)
	mux.Handle("/assets/", http.StripPrefix("/assets/", handler.static))

	return handler.withRequestLogging(handler.withAccessControl(mux)), nil
}

type handler struct {
	login       *loginState
	renderer    render.Renderer
	static      http.Handler
	data        ShellData
	provider    DataProvider
	access      AccessPolicy
	logRequests io.Writer
}

func (h handler) registerGuidedWorkflowRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/cases/workflow", h.guidedUseCase)
	mux.HandleFunc("/cases/workflow/state", h.guidedUseCaseState)
	mux.HandleFunc("/guided", h.guidedUseCase)
	mux.HandleFunc("/guided/actions", h.guidedUseCaseAction)
	mux.HandleFunc("/actions", h.writeAction)
	mux.HandleFunc("/workflows", h.workflows)
	mux.HandleFunc("/workflows/fragment", h.workflows)
	mux.HandleFunc("/workflows/actions", h.workflowAction)
}
