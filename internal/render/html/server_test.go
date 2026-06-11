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
	req := httptest.NewRequest(http.MethodGet, "/s/economics", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	serveView(rec, req, panel.ViewEconomics, stubRender, panel.Options{})
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
	req := httptest.NewRequest(http.MethodGet, "/s/economics", nil)
	req.Header.Set("HX-Request", "true")
	req.Header.Set("HX-Boosted", "true")
	rec := httptest.NewRecorder()
	serveView(rec, req, panel.ViewEconomics, stubRender, panel.Options{})
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
}
