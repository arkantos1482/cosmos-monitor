package html

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	sessionCookie = "pmtop_session"
	sessionTTL    = 24 * time.Hour
)

// Auth is the shared operator login. Empty User or Pass disables the gate (local/dev).
type Auth struct {
	User string
	Pass string
}

func (a Auth) enabled() bool {
	return strings.TrimSpace(a.User) != "" && a.Pass != ""
}

type sessionStore struct {
	mu   sync.Mutex
	live map[string]time.Time
}

func newSessionStore() *sessionStore {
	return &sessionStore{live: make(map[string]time.Time)}
}

func (s *sessionStore) issue() string {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("pmtop: session id: " + err.Error())
	}
	id := hex.EncodeToString(b[:])
	s.mu.Lock()
	s.live[id] = time.Now().Add(sessionTTL)
	s.mu.Unlock()
	return id
}

func (s *sessionStore) valid(id string) bool {
	if id == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	exp, ok := s.live[id]
	if !ok || time.Now().After(exp) {
		delete(s.live, id)
		return false
	}
	return true
}

func (s *sessionStore) revoke(id string) {
	s.mu.Lock()
	delete(s.live, id)
	s.mu.Unlock()
}

func (s *dashServer) withLoginGate(next http.Handler) http.Handler {
	if !s.auth.enabled() {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" {
			next.ServeHTTP(w, r)
			return
		}
		if r.URL.Path == "/logout" && r.Method == http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		if s.sessions.valid(sessionID(r)) {
			next.ServeHTTP(w, r)
			return
		}
		if r.Header.Get("HX-Request") != "" {
			w.Header().Set("HX-Redirect", "/login")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Redirect(w, r, "/login?next="+url.QueryEscape(r.URL.RequestURI()), http.StatusSeeOther)
	})
}

func sessionID(r *http.Request) string {
	c, err := r.Cookie(sessionCookie)
	if err != nil {
		return ""
	}
	return c.Value
}

func setSessionCookie(w http.ResponseWriter, r *http.Request, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   forwardedHTTPS(r),
	})
}

func forwardedHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")), "https")
}

func credsOK(auth Auth, user, pass string) bool {
	wantU := []byte(strings.TrimSpace(auth.User))
	gotU := []byte(strings.TrimSpace(user))
	return subtle.ConstantTimeCompare(gotU, wantU) == 1 &&
		subtle.ConstantTimeCompare([]byte(pass), []byte(auth.Pass)) == 1
}

func safeNext(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "/login" || raw == "/logout" {
		return "/"
	}
	if !strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "//") {
		return "/"
	}
	if strings.ContainsAny(raw, "\r\n") {
		return "/"
	}
	return raw
}

func formNext(r *http.Request) string {
	_ = r.ParseForm()
	if v := r.Form.Get("next"); v != "" {
		return safeNext(v)
	}
	return safeNext(r.URL.Query().Get("next"))
}
