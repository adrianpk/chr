package webui

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

const loginTestPassword = "example-only-password-2026"

func loginTestCredentials(t *testing.T) LoginCredentials {
	t.Helper()

	credentials, err := CreateLoginCredentials("operator", loginTestPassword)
	if err != nil {
		t.Fatal(err)
	}

	return credentials
}

func loginTestHandler(t *testing.T) http.Handler {
	t.Helper()

	credentials := loginTestCredentials(t)

	h, err := NewHandler(HandlerOptions{
		Data: testShellData(), Login: &credentials,
		Access: AccessPolicy{RequireToken: true, Token: "legacy-access-token"},
	})
	if err != nil {
		t.Fatal(err)
	}

	return h
}

// A browser cookie protects every route and replaces token-bearing navigation.
func TestLoginJourney(t *testing.T) {
	server := httptest.NewServer(loginTestHandler(t))
	defer server.Close()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	client := &http.Client{Jar: jar, Timeout: 5 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	response := loginTestPost(t, client, server.URL, "/login", url.Values{"username": {"operator"}, "password": {loginTestPassword}, "next": {"/specs"}})

	cookies := response.Cookies()
	if response.StatusCode != http.StatusSeeOther || response.Header.Get("Location") != "/specs" || len(cookies) != 1 {
		t.Fatal("login did not create a session and preserve the destination")
	}

	loginTestBody(t, response)

	cookie := cookies[0]
	loginTestCookie(t, cookie)

	loginTestPages(t, client, server.URL)

	response = loginTestPost(t, client, server.URL, "/logout", nil)
	loginTestBody(t, response)

	if response.StatusCode != http.StatusSeeOther {
		t.Fatal("logout failed")
	}

	loginTestRevoked(t, server, cookie)
}

func loginTestCookie(t *testing.T, cookie *http.Cookie) {
	t.Helper()

	if !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode || cookie.Path != "/" || cookie.MaxAge != 43200 || cookie.Domain != "" {
		t.Fatal("session cookie lost its scope or lifetime")
	}
}

// Public paths stay usable; direct URLs, fragments, writes and legacy tokens cannot bypass login.
func TestLoginGuards(t *testing.T) {
	h := loginTestHandler(t)

	tests := []struct {
		name, method, path, origin string
		htmx                       bool
		want                       int
	}{
		{"page", "GET", "/board", "", false, 303},
		{"details", "GET", "/specs/detail?id=private", "", false, 303},
		{"fragment", "GET", "/board/fragment", "", true, 401},
		{"artifact fragment", "GET", "/dumps/detail/fragment?id=private", "", true, 401},
		{"write", "POST", "/actions", "http://example.com", false, 401},
		{"ticket write", "POST", "/board/ticket/save", "http://example.com", false, 401},
		{"workflow write", "POST", "/workflows/actions", "http://example.com", false, 401},
		{"cross origin", "POST", "/login", "http://other.example", false, 403},
		{"missing origin", "POST", "/login", "", false, 403},
		{"health", "GET", "/healthz", "", false, 200},
		{"asset", "GET", "/assets/chr.css", "", false, 200},
		{"login", "GET", "/login", "", false, 200},
		{"legacy token", "GET", "/board?token=legacy-access-token", "", false, 303},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, nil)
			request.Header.Set("Origin", test.origin)
			request.Header.Set("Authorization", "Bearer legacy-access-token")

			if test.htmx {
				request.Header.Set("HX-Request", "true")
			}

			recorder := httptest.NewRecorder()
			h.ServeHTTP(recorder, request)

			if recorder.Code != test.want || strings.Contains(recorder.Body.String(), "project-1") {
				t.Fatalf("unprotected response: %d", recorder.Code)
			}

			if test.htmx && recorder.Header().Get("HX-Redirect") != "/login" {
				t.Fatal("expired fragment did not request full login navigation")
			}

			if strings.Contains(recorder.Header().Get("Location"), "legacy-access-token") {
				t.Fatal("legacy token leaked into login redirect")
			}
		})
	}
}

// Wrong credentials get one generic error; bounded attempts limit password work.
func TestLoginAttempts(t *testing.T) {
	h := loginTestHandler(t)

	for i := range 6 {
		form := url.Values{"username": {"operator"}, "password": {"incorrect-password"}}
		request := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("Origin", "http://example.com")

		recorder := httptest.NewRecorder()
		h.ServeHTTP(recorder, request)

		want := http.StatusUnauthorized
		if i == 5 {
			want = http.StatusTooManyRequests
		}

		if recorder.Code != want || len(recorder.Result().Cookies()) != 0 {
			t.Fatal("failed login created a session or escaped the attempt budget")
		}
	}
}

// Session expiry, replacement, server restart and concurrent reads preserve denial.
func TestLoginSessions(t *testing.T) {
	credentials := loginTestCredentials(t)

	state, err := newLoginState(&credentials)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	state.now = func() time.Time { return now }
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "https://example.com/board", nil)
	state.begin(recorder, request)

	cookie := recorder.Result().Cookies()[0]
	request.AddCookie(cookie)

	if !cookie.Secure || !state.authenticated(request) {
		t.Fatal("TLS session was not protected and usable")
	}

	loginTestConcurrent(t, state, request)

	fresh, err := newLoginState(&credentials)
	if err != nil || fresh.authenticated(request) {
		t.Fatal("restart retained an in-memory session")
	}

	replacement := httptest.NewRecorder()
	state.begin(replacement, request)

	if state.authenticated(request) {
		t.Fatal("re-login retained the replaced session")
	}

	expiring := httptest.NewRequest(http.MethodGet, "https://example.com/board", nil)
	expiring.AddCookie(replacement.Result().Cookies()[0])

	now = now.Add(loginLifetime)

	if state.authenticated(expiring) {
		t.Fatal("expired server-side session remained valid")
	}

	if state.attemptAllowed() == false {
		t.Fatal("new attempt window did not open")
	}
}

func loginTestPost(t *testing.T, client *http.Client, origin, path string, form url.Values) *http.Response {
	t.Helper()

	request, err := http.NewRequestWithContext(t.Context(), http.MethodPost, origin+path, strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}

	request.Header.Set("Origin", origin)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}

	return response
}

func loginTestBody(t *testing.T, response *http.Response) string {
	t.Helper()

	body, readErr := io.ReadAll(response.Body)

	closeErr := response.Body.Close()
	if readErr != nil || closeErr != nil {
		t.Fatal(readErr, closeErr)
	}

	return string(body)
}

func loginTestPages(t *testing.T, client *http.Client, origin string) {
	t.Helper()

	for _, path := range []string{"/board", "/dash", "/cases", "/dumps", "/quizzes", "/specs", "/plans", "/project"} {
		request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, origin+path, nil)
		if err != nil {
			t.Fatal(err)
		}

		response, err := client.Do(request)
		if err != nil {
			t.Fatal(err)
		}

		body := loginTestBody(t, response)
		if response.StatusCode != http.StatusOK || !strings.Contains(body, "Log out") || response.Header.Get("Cache-Control") != "no-store" {
			t.Fatalf("%s did not load as a protected page: %d", path, response.StatusCode)
		}
	}
}

func loginTestRevoked(t *testing.T, server *httptest.Server, cookie *http.Cookie) {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, "/board", nil)
	request.AddCookie(cookie)

	recorder := httptest.NewRecorder()
	// The client cookie is also denied after server-side logout.
	server.Config.Handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusSeeOther {
		t.Fatal("revoked session remained valid")
	}
}

func loginTestConcurrent(t *testing.T, state *loginState, request *http.Request) {
	t.Helper()

	var group sync.WaitGroup
	for range 20 {
		group.Go(func() {
			if !state.authenticated(request) {
				t.Error("concurrent session read failed")
			}
		})
	}

	group.Wait()
}
