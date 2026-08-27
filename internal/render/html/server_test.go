package html

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/panel"
)

func stubRender(v panel.View) model.Report {
	return model.Report{Moniker: "node1"}
}

func TestServeViewPollReturnsFragment(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/s/rewards", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	serveView(rec, req, panel.ViewRewards, stubRender, panel.Options{})
	body := rec.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") || strings.Contains(body, `id="dash-nav"`) {
		t.Fatal("poll request should return fragment only")
	}
	if !strings.Contains(body, "dash-section") && !strings.Contains(body, "dash-overview") {
		t.Fatal("poll response should contain rendered view content")
	}
	if !strings.Contains(body, `id="dash-status"`) || !strings.Contains(body, `hx-swap-oob="true"`) {
		t.Fatal("poll response should include OOB status bar")
	}
}

func TestServeViewBoostReturnsFullPage(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/s/rewards", nil)
	req.Header.Set("HX-Request", "true")
	req.Header.Set("HX-Boosted", "true")
	rec := httptest.NewRecorder()
	serveView(rec, req, panel.ViewRewards, stubRender, panel.Options{})
	body := rec.Body.String()
	for _, want := range []string{"<!DOCTYPE html>", `hx-boost="true"`, `id="dash-status"`, `id="dash-nav"`, `id="data"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("boost request missing %q", want)
		}
	}
}

func TestServeViewDirectReturnsFullPage(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	serveView(rec, req, panel.ViewHome, stubRender, panel.Options{})
	body := rec.Body.String()
	if !strings.Contains(body, `<!DOCTYPE html>`) || !strings.Contains(body, `hx-boost="true"`) {
		t.Fatal("direct load should return full boosted page")
	}
	if !strings.Contains(body, `id="dash-status"`) {
		t.Fatal("direct load should include global status bar")
	}
	assertSecurityHeaders(t, rec.Result())
}

func TestServeViewSetsSecurityHeadersOnFragments(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/s/rewards", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	serveView(rec, req, panel.ViewRewards, stubRender, panel.Options{})
	assertSecurityHeaders(t, rec.Result())
}

func TestStaticJS(t *testing.T) {
	mux := http.NewServeMux()
	registerRoutes(mux, stubRender, panel.Options{})
	for _, path := range []string{
		"/static/htmx.min.js",
		"/static/htmx-init.js",
		"/static/ethers.umd.min.js",
		"/static/delegate.js",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		res := rec.Result()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("%s status %d", path, res.StatusCode)
		}
		ct := res.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "application/javascript") {
			t.Fatalf("%s content-type %q", path, ct)
		}
		assertSecurityHeaders(t, res)
		if rec.Body.Len() == 0 {
			t.Fatalf("%s empty body", path)
		}
	}
}

func TestDelegatePageThroughMux(t *testing.T) {
	mux := http.NewServeMux()
	registerRoutes(mux, stubRender, panel.Options{})
	req := httptest.NewRequest(http.MethodGet, "/delegate", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	assertSecurityHeaders(t, res)
	body := rec.Body.String()
	if !strings.Contains(body, `src="/static/ethers.umd.min.js"`) || !strings.Contains(body, `src="/static/delegate.js"`) {
		t.Fatal("delegate page should load first-party wallet scripts")
	}
}

func assertSecurityHeaders(t *testing.T, res *http.Response) {
	t.Helper()
	csp := res.Header.Get("Content-Security-Policy")
	if !strings.Contains(csp, "script-src 'self'") || !strings.Contains(csp, "frame-ancestors 'none'") {
		t.Fatalf("CSP missing script-src 'self' or frame-ancestors 'none': %q", csp)
	}
	if strings.Contains(csp, "'unsafe-inline'") && strings.Contains(csp, "script-src") {
		scriptPart := csp
		if i := strings.Index(csp, "script-src"); i >= 0 {
			scriptPart = csp[i:]
			if j := strings.Index(scriptPart, ";"); j >= 0 {
				scriptPart = scriptPart[:j]
			}
		}
		if strings.Contains(scriptPart, "'unsafe-inline'") {
			t.Fatalf("script-src must not allow unsafe-inline: %q", scriptPart)
		}
	}
	if res.Header.Get("X-Frame-Options") != "DENY" {
		t.Fatal("X-Frame-Options should be DENY")
	}
	if res.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("X-Content-Type-Options should be nosniff")
	}
	if res.Header.Get("Referrer-Policy") != "no-referrer" {
		t.Fatal("Referrer-Policy should be no-referrer")
	}
}
