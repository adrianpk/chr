// Package webui owns the local Web UI handler and embedded assets.
//
//nolint:wsl_v5 // Handler methods keep request decoding and guarded rendering steps together.
package webui

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

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
