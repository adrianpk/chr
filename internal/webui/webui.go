// Package webui owns the local Web UI handler and embedded assets.
//
//nolint:wsl_v5 // Handler methods keep request decoding and guarded rendering steps together.
package webui

import (
	"crypto/subtle"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/webui/htmx"
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

	mux.HandleFunc("/actions", handler.writeAction)
	mux.HandleFunc("/workflows", handler.workflows)
	mux.HandleFunc("/workflows/fragment", handler.workflows)
	mux.HandleFunc("/workflows/actions", handler.workflowAction)
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

func (h handler) writeAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	request := writeActionRequest(r)
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

	if r.PostFormValue("surface") == "modal" && r.PostFormValue("action") != "dump.create" || r.URL.Path == "/board/ticket/comment" {
		return "write_result_modal.html"
	}

	return "write_result.html"
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
	case "case.update", "dump.create", "dump.save", "dump.append":
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

func (h handler) withAccessControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !h.allowOrigin(w, r) {
			return
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)

			return
		}

		if !h.allowLogin(w, r) || !h.allowToken(w, r) {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h handler) allowOrigin(w http.ResponseWriter, r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}

	if !h.originAllowed(r, origin) {
		http.Error(w, "forbidden origin", http.StatusForbidden)

		return false
	}

	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, HX-Request, X-CHR-Web-Token")
	w.Header().Add("Vary", "Origin")

	return true
}

func (h handler) originAllowed(r *http.Request, origin string) bool {
	if origin == requestOrigin(r) {
		return true
	}

	if origin == strings.TrimSpace(h.access.Origin) {
		return true
	}

	for _, allowed := range h.access.AllowedOrigins {
		if origin == strings.TrimSpace(allowed) {
			return true
		}
	}

	return false
}

func requestOrigin(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); forwarded != "" {
		scheme = forwarded
	}

	return scheme + "://" + r.Host
}

func (h handler) allowToken(w http.ResponseWriter, r *http.Request) bool {
	if !h.access.RequireToken || tokenExemptPath(r.URL.Path) {
		return true
	}

	if secureTokenEqual(h.requestToken(r), h.access.Token) {
		return true
	}

	http.Error(w, "web token required", http.StatusUnauthorized)

	return false
}

func tokenExemptPath(path string) bool {
	return path == "/healthz" || strings.HasPrefix(path, "/assets/")
}

func (h handler) requestToken(r *http.Request) string {
	if token := strings.TrimSpace(r.Header.Get("X-CHR-Web-Token")); token != "" {
		return token
	}

	if token := bearerToken(r.Header.Get("Authorization")); token != "" {
		return token
	}

	return strings.TrimSpace(r.URL.Query().Get("token"))
}

func secureTokenEqual(got, want string) bool {
	if got == "" || want == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}

func bearerToken(value string) string {
	const prefix = "Bearer "

	if !strings.HasPrefix(value, prefix) {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(value, prefix))
}

func (h handler) withRequestLogging(next http.Handler) http.Handler {
	if h.logRequests == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		started := time.Now()

		next.ServeHTTP(recorder, r)

		_, err := fmt.Fprintf(
			h.logRequests,
			"web request: method=%s path=%s status=%d duration=%s\n",
			r.Method,
			r.URL.EscapedPath(),
			recorder.status,
			time.Since(started).Round(time.Millisecond),
		)
		if err != nil {
			return
		}
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (h handler) projectPageData(r *http.Request) (PageData, error) {
	project, err := h.provider.ProjectView(r.Context())
	if err != nil {
		return PageData{}, err
	}

	return PageData{
		Shell:    h.shellData("project"),
		View:     "project",
		Title:    "Project",
		Revision: project.Revision,
		Project:  project,
	}, nil
}

func (h handler) boardPageData(r *http.Request) (PageData, error) {
	offset := 0
	if value := r.URL.Query().Get("offset"); value != "" {
		var err error
		offset, err = strconv.Atoi(value)
		if err != nil || offset < 0 || offset > 4096 {
			return PageData{}, fmt.Errorf("invalid Board page offset")
		}
	}
	board, err := h.provider.BoardView(r.Context(), BoardRequest{Offset: offset})
	if err != nil {
		return PageData{}, err
	}

	refresh := "/board/fragment"
	if offset != 0 {
		refresh += "?offset=" + strconv.Itoa(offset)
	}

	return PageData{
		Shell:      h.shellData("board"),
		View:       "board",
		Title:      "Board",
		Revision:   board.Revision,
		RefreshURL: refresh,
		Board:      board,
	}, nil
}

func (h handler) ticketDetailPageData(r *http.Request) (PageData, error) {
	ticket, err := h.provider.TicketDetailView(r.Context(), TicketDetailRequest{
		Key: strings.TrimSpace(r.URL.Query().Get("key")),
		New: r.URL.Query().Get("new") == "1",
	})
	if err != nil {
		return PageData{}, err
	}

	return PageData{
		Shell:    h.shellData("board"),
		View:     "ticket_detail",
		Title:    ticket.Title,
		Revision: ticket.Revision,
		Ticket:   ticket,
	}, nil
}

func ticketDetailRequest(r *http.Request, key string) *http.Request {
	clone := r.Clone(r.Context())
	query := clone.URL.Query()
	query.Set("key", key)
	clone.URL.RawQuery = query.Encode()

	return clone
}

func (h handler) dashPageData(r *http.Request) (PageData, error) {
	dash, err := h.readDashPage(r)
	if err != nil {
		return PageData{}, err
	}

	return h.dashPageProjection(r, dash), nil
}

func (h handler) catalogPageData(r *http.Request, family string) (PageData, error) {
	catalog, err := h.provider.CatalogView(r.Context(), family)
	if err != nil {
		return PageData{}, err
	}

	return PageData{
		Shell:      h.shellData(family),
		View:       "catalog",
		Title:      catalog.Title,
		Revision:   catalog.Revision,
		RefreshURL: "/" + family + "/fragment",
		Catalog:    catalog,
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

	title := detail.Title
	if detail.Quiz != nil {
		title = "Quizzes"
	}
	if detail.Spec != nil {
		title = "Specs"
	}

	return PageData{
		Shell:      h.shellData(family),
		View:       "artifact_detail",
		Title:      title,
		Revision:   detail.Revision,
		RefreshURL: artifactDetailRefreshURL(family, r.URL.Query().Get("id")),
		Detail:     detail,
	}, nil
}

func artifactDetailRefreshURL(family, id string) string {
	if family == "cases" || family == "dumps" || family == "quizzes" || family == "specs" {
		return ""
	}

	return "/" + family + "/detail/fragment?id=" + url.QueryEscape(id)
}

func (h handler) shellData(active string) ShellData {
	data := h.data
	_, data.HarnessesEnabled = h.provider.(HarnessSelectionProvider)
	if strings.TrimSpace(data.ProjectName) == "" {
		data.ProjectName = data.ProjectID
	}
	data.ProjectName = displayProjectName(data.ProjectName)

	data.Navigation = make([]ui.NavLink, 0, len(h.data.Navigation))

	for _, link := range h.data.Navigation {
		link.Active = navLinkActive(active, link.Href)
		data.Navigation = append(data.Navigation, link)
	}

	return data
}

func displayProjectName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	first := name[0]
	if first >= 'a' && first <= 'z' {
		return string(first-'a'+'A') + name[1:]
	}

	return name
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
