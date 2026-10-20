package webui

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Redirect destinations stay local, including encoded paths and old token links.
func TestLoginRedirects(t *testing.T) {
	for _, test := range []struct{ name, input, want string }{
		{"local", "/specs?scope=sample", "/specs?scope=sample"},
		{"external", "https://other.example/", "/board"},
		{"authority", "//other.example/", "/board"},
		{"encoded authority", "/%2fother.example/", "/board"},
		{"backslash", "/%5cother.example/", "/board"},
		{"credentials", "/specs?token=private", "/board"},
		{"loop", "/login", "/board"},
		{"line break", "/%0d%0aLocation:other", "/board"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := loginNext(test.input); got != test.want {
				t.Fatalf("destination = %q, want %q", got, test.want)
			}
		})
	}
}

// Cookie-authenticated writes require the actual same origin, even without an Origin header.
func TestLoginWriteOrigin(t *testing.T) {
	state, err := newLoginState(&LoginCredentials{Version: 1, Username: "operator", Salt: make([]byte, 16), Hash: make([]byte, 32)})
	if err != nil {
		t.Fatal(err)
	}

	cookieResponse := httptest.NewRecorder()
	state.begin(cookieResponse, httptest.NewRequest("GET", "/", nil))

	cookie := cookieResponse.Result().Cookies()[0]
	h := handler{login: state}
	protected := h.withAccessControl(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))

	for _, test := range []struct {
		name, origin, forwarded string
		want                    int
	}{
		{"same origin", "http://example.com", "", 204},
		{"missing", "", "", 403},
		{"different host", "http://other.example", "", 403},
		{"different port", "http://example.com:8080", "", 403},
		{"forged TLS proxy", "https://example.com", "https", 403},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/actions", nil)
			request.AddCookie(cookie)
			request.Header.Set("Origin", test.origin)
			request.Header.Set("X-Forwarded-Proto", test.forwarded)

			recorder := httptest.NewRecorder()
			protected.ServeHTTP(recorder, request)

			if recorder.Code != test.want {
				t.Fatal("write origin guard failed", recorder.Code)
			}
		})
	}
}

// Untrusted login traffic cannot grow verifier work or session memory without a bound.
func TestLoginLimits(t *testing.T) {
	now := time.Now()

	state := &loginState{sessions: make(map[[32]byte]time.Time), now: func() time.Time { return now }}
	for range 5 {
		if !state.attemptAllowed() {
			t.Fatal("login budget rejected an available attempt")
		}
	}

	if state.attemptAllowed() {
		t.Fatal("login budget allowed unlimited password work")
	}

	now = now.Add(time.Minute)

	if !state.attemptAllowed() {
		t.Fatal("login budget did not reopen after one minute")
	}

	for range loginSessionLimit + 1 {
		state.begin(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	}

	if len(state.sessions) > loginSessionLimit {
		t.Fatal("sessions exceeded the memory bound")
	}
}
