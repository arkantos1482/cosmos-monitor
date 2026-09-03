package html

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/panel"
)

func authedHandler() http.Handler {
	return NewHandler(stubRender, panel.Options{}, Auth{User: "ops", Pass: "secret"})
}

func TestUnauthenticatedBrowserGetsLoginPage(t *testing.T) {
	h := authedHandler()
	for _, path := range []string{"/", "/s/feemarket", "/delegate"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusSeeOther {
			t.Fatalf("%s = %d, want 303", path, rec.Code)
		}
		loc := rec.Header().Get("Location")
		if !strings.HasPrefix(loc, "/login?") {
			t.Fatalf("%s Location = %q, want /login?…", path, loc)
		}
		if rec.Header().Get("WWW-Authenticate") != "" {
			t.Fatal("must not use HTTP basic auth")
		}
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/login", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /login = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`pmtop — Sign in`,
		`<form method="post" action="/login"`,
		`type="password"`,
		`name="username"`,
		`Sign in`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("login page missing %q", want)
		}
	}
	if strings.Contains(body, `hx-boost`) {
		t.Fatal("login page must not enable HTMX boost")
	}
}

func TestWrongPasswordStaysOnLoginForm(t *testing.T) {
	h := authedHandler()
	form := url.Values{"username": {"ops"}, "password": {"nope"}, "next": {"/"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("POST /login wrong password = %d, want 401", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Wrong user or password") {
		t.Fatal("login form should show an error")
	}
	if rec.Header().Get("Set-Cookie") != "" && strings.Contains(rec.Header().Get("Set-Cookie"), "pmtop_session=") &&
		!strings.Contains(rec.Header().Get("Set-Cookie"), "Max-Age=0") {
		t.Fatal("wrong password must not set a session")
	}
}

func TestLoginSetsCookieAndOpensPages(t *testing.T) {
	h := authedHandler()
	form := url.Values{"username": {"ops"}, "password": {"secret"}, "next": {"/s/feemarket"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("POST /login = %d, want 303", rec.Code)
	}
	if rec.Header().Get("Location") != "/s/feemarket" {
		t.Fatalf("Location = %q", rec.Header().Get("Location"))
	}
	cookie := sessionFrom(rec)
	if cookie == "" {
		t.Fatal("missing pmtop_session cookie")
	}

	for _, path := range []string{"/", "/s/feemarket", "/delegate"} {
		got := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, path, nil)
		r.AddCookie(&http.Cookie{Name: sessionCookie, Value: cookie})
		h.ServeHTTP(got, r)
		if got.Code != http.StatusOK {
			t.Fatalf("authed GET %s = %d, want 200", path, got.Code)
		}
		if strings.Contains(got.Body.String(), `action="/login"`) {
			t.Fatalf("%s still showing login form", path)
		}
	}
}

func TestLogoutClearsSession(t *testing.T) {
	h := authedHandler()
	cookie := mustLogin(t, h)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: cookie})
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("POST /logout = %d, want 303", rec.Code)
	}

	got := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: sessionCookie, Value: cookie})
	h.ServeHTTP(got, r)
	if got.Code != http.StatusSeeOther {
		t.Fatalf("revoked session still admitted: %d", got.Code)
	}
}

func TestHTMXPollWithoutSessionIsUnauthorized(t *testing.T) {
	h := authedHandler()
	req := httptest.NewRequest(http.MethodGet, "/s/feemarket", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("HX poll = %d, want 401", rec.Code)
	}
	if rec.Header().Get("HX-Redirect") != "/login" {
		t.Fatalf("HX-Redirect = %q", rec.Header().Get("HX-Redirect"))
	}
}

func TestRejectsOpenRedirect(t *testing.T) {
	if safeNext("https://evil.example/") != "/" {
		t.Fatal("absolute URL must not be next")
	}
	if safeNext("//evil.example/") != "/" {
		t.Fatal("protocol-relative URL must not be next")
	}
}

func TestHandlerNoAuthWhenCredentialsEmpty(t *testing.T) {
	h := NewHandler(stubRender, panel.Options{}, Auth{})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("open mode / = %d, want 200", rec.Code)
	}
}

func mustLogin(t *testing.T, h http.Handler) string {
	t.Helper()
	form := url.Values{"username": {"ops"}, "password": {"secret"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	c := sessionFrom(rec)
	if c == "" {
		t.Fatal("login did not set session")
	}
	return c
}

func sessionFrom(rec *httptest.ResponseRecorder) string {
	for _, c := range rec.Result().Cookies() {
		if c.Name == sessionCookie && c.Value != "" && c.MaxAge != -1 {
			return c.Value
		}
	}
	return ""
}
