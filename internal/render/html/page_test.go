package html

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/panel"
)

func TestFullPageNavAndAutoRefresh(t *testing.T) {
	out := FullPage("node1", panel.ViewEconomics, "<p>body</p>")
	for _, want := range []string{
		`id="dash-nav"`,
		`dash-nav__link--active`,
		`href="/s/economics"`,
		`Economics`,
		`<p>body</p>`,
		`scheduleAutoRefresh`,
		`location.reload`,
		`sessionStorage.setItem`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("page missing %q", want)
		}
	}
	for _, bad := range []string{
		`htmx`,
		`hx-get`,
		`hx-swap`,
		`dash-refresh`,
		`dash-view`,
		`/fragment`,
	} {
		if strings.Contains(out, bad) {
			t.Fatalf("page should not contain %q", bad)
		}
	}
}

func TestNavLinksArePlainAnchors(t *testing.T) {
	out := navHTML(panel.ViewInfra)
	if strings.Contains(out, `hx-`) {
		t.Fatal("nav should not use HTMX attributes")
	}
	if !strings.Contains(out, `href="/s/infra"`) || !strings.Contains(out, `dash-nav__link--active`) {
		t.Fatal("nav should link to section paths with active class")
	}
}
