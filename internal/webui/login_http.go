package webui

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"forge.adrianpk.com/chr/internal/webui/htmx"
)

func (h handler) allowLogin(w http.ResponseWriter, r *http.Request) bool {
	if h.login == nil || tokenExemptPath(r.URL.Path) {
		return true
	}

	w.Header().Set("Cache-Control", "no-store")
	// Same-origin form posts must retain their Origin instead of the opaque "null" origin.
	w.Header().Set("Referrer-Policy", "same-origin")
	w.Header().Set("Content-Security-Policy", "frame-ancestors 'none'; form-action 'self'")

	if !loginOriginAllowed(r) {
		http.Error(w, "same-origin request required", http.StatusForbidden)

		return false
	}

	if r.URL.Path == "/login" {
		h.loginPage(w, r)

		return false
	}

	if h.login.authenticated(r) {
		return true
	}

	if htmx.IsRequest(r) {
		w.Header().Set("HX-Redirect", "/login")
		http.Error(w, "login required", http.StatusUnauthorized)

		return false
	}

	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		http.Redirect(w, r, "/login?next="+url.QueryEscape(loginNext(r.URL.RequestURI())), http.StatusSeeOther)
	} else {
		http.Error(w, "login required", http.StatusUnauthorized)
	}

	return false
}

func loginOriginAllowed(r *http.Request) bool {
	if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
		return true
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	return r.Header.Get("Origin") == scheme+"://"+r.Host && r.Header.Get("Sec-Fetch-Site") != "cross-site"
}

func (h handler) loginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		h.renderLogin(w, loginNext(r.URL.Query().Get("next")), "")

		return
	}

	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "GET, HEAD, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid login form", http.StatusBadRequest)

		return
	}

	if !h.login.attemptAllowed() {
		w.Header().Set("Retry-After", "60")
		http.Error(w, "too many login attempts; try again in one minute", http.StatusTooManyRequests)

		return
	}

	next := loginNext(r.PostForm.Get("next"))
	if !h.login.credentials.matches(r.PostForm.Get("username"), r.PostForm.Get("password")) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		h.renderLogin(w, next, "Incorrect username or password.")

		return
	}

	h.login.begin(w, r)
	http.Redirect(w, r, next, http.StatusSeeOther)
}

func (h handler) renderLogin(w http.ResponseWriter, next, message string) {
	if err := h.renderer.Page(w, "login.html", struct{ Next, Error string }{next, message}); err != nil {
		http.Error(w, "render login", http.StatusInternalServerError)
	}
}

func loginNext(value string) string {
	parsed, err := url.Parse(value)
	if err != nil || !strings.HasPrefix(value, "/") || parsed.Host != "" {
		return "/board"
	}

	// Credentials in legacy query URLs never enter a login form or redirect.
	if parsed.Query().Has("token") || parsed.Path == "/login" || parsed.Path == "/logout" {
		return "/board"
	}

	if strings.HasPrefix(parsed.Path, "//") || strings.ContainsAny(parsed.Path, "\\\r\n") {
		return "/board"
	}

	return parsed.RequestURI()
}

func (h *handler) configureLogin(credentials *LoginCredentials) error {
	state, err := newLoginState(credentials)
	if err != nil {
		return err
	}

	h.login = state

	h.data.LoginEnabled = state != nil
	if state != nil {
		h.access.RequireToken = false
	}

	if h.access.RequireToken && strings.TrimSpace(h.access.Token) == "" {
		return fmt.Errorf("web token is required but empty")
	}

	return nil
}

func (h handler) logout(w http.ResponseWriter, r *http.Request) {
	if h.login == nil {
		http.NotFound(w, r)

		return
	}

	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	h.login.end(w, r)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
