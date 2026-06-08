package html

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/panel"
)

func TestFullPageHTMXShell(t *testing.T) {
	out := FullPage("node1", panel.ViewEconomics, "<p>body</p>")
	for _, want := range []string{
		`id="dash-nav"`,
		`dash-nav__link--active`,
		`href="/s/economics"`,
		`hx-get="/s/economics"`,
		`Economics`,
		`<p>body</p>`,
		`id="data"`,
		`hx-trigger="every 5s"`,
		`hx-swap="innerHTML show:none scroll:none settle:none"`,
		`htmx.org`,
		`htmx:afterSwap`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("page missing %q", want)
		}
	}
	for _, bad := range []string{
		`scheduleAutoRefresh`,
		`location.reload`,
		`sessionStorage`,
		`snapshotDashState`,
		`restoreDashState`,
		`setInterval`,
		`/fragment`,
	} {
		if strings.Contains(out, bad) {
			t.Fatalf("page should not contain %q", bad)
		}
	}
}

func TestNavLinksUseHTMX(t *testing.T) {
	out := navHTML(panel.ViewInfra)
	if !strings.Contains(out, `hx-get="/s/infra"`) || !strings.Contains(out, `dash-nav__link--active`) {
		t.Fatal("nav should use HTMX partial navigation with active class")
	}
}

func TestDataURL(t *testing.T) {
	if dataURL(panel.ViewHome) != "/" {
		t.Fatal("home data URL should be /")
	}
	if dataURL(panel.ViewEconomics) != "/s/economics" {
		t.Fatal("economics data URL should be /s/economics")
	}
}
