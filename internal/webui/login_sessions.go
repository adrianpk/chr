package webui

import (
	"crypto/rand"
	"crypto/sha256"
	"net/http"
	"sync"
	"time"
)

const (
	loginSessionCookie = "chr_web_session"
	loginLifetime      = 12 * time.Hour
	loginSessionLimit  = 32
)

type loginState struct {
	credentials LoginCredentials
	mu          sync.Mutex
	sessions    map[[32]byte]time.Time
	attempts    int
	window      time.Time
	now         func() time.Time
}

func newLoginState(credentials *LoginCredentials) (*loginState, error) {
	if credentials == nil {
		return nil, nil
	}

	if err := credentials.validate(); err != nil {
		return nil, err
	}

	return &loginState{credentials: *credentials, sessions: make(map[[32]byte]time.Time), now: time.Now}, nil
}

func (s *loginState) authenticated(r *http.Request) bool {
	cookie, err := r.Cookie(loginSessionCookie)
	if err != nil || len(cookie.Value) < 26 || len(cookie.Value) > 128 {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := sha256.Sum256([]byte(cookie.Value))

	expires, ok := s.sessions[key]
	if !ok || !s.now().Before(expires) {
		delete(s.sessions, key)

		return false
	}

	return true
}

func (s *loginState) attemptAllowed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	if now.Sub(s.window) >= time.Minute {
		s.window, s.attempts = now, 0
	}

	if s.attempts >= 5 {
		return false
	}

	s.attempts++

	return true
}

func (s *loginState) begin(w http.ResponseWriter, r *http.Request) {
	token := rand.Text()
	now := s.now()

	s.mu.Lock()

	for key, expires := range s.sessions {
		if !now.Before(expires) {
			delete(s.sessions, key)
		}
	}

	if cookie, err := r.Cookie(loginSessionCookie); err == nil {
		delete(s.sessions, sha256.Sum256([]byte(cookie.Value)))
	}

	if len(s.sessions) >= loginSessionLimit {
		for key := range s.sessions {
			delete(s.sessions, key)

			break
		}
	}

	s.sessions[sha256.Sum256([]byte(token))] = now.Add(loginLifetime)
	s.mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name: loginSessionCookie, Value: token, Path: "/", MaxAge: int(loginLifetime.Seconds()),
		Expires: now.Add(loginLifetime), HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteStrictMode,
	})
}

func (s *loginState) end(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(loginSessionCookie); err == nil {
		s.mu.Lock()
		delete(s.sessions, sha256.Sum256([]byte(cookie.Value)))
		s.mu.Unlock()
	}

	http.SetCookie(w, &http.Cookie{
		Name: loginSessionCookie, Path: "/", MaxAge: -1, Expires: time.Unix(1, 0),
		HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteStrictMode,
	})
}
