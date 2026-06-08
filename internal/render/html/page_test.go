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
		`Economics`,
		`<p>body</p>`,
		`id="data"`,
		`hx-get="/s/economics"`,
		`hx-trigger="every 5s"`,
		`hx-swap="innerHTML show:none scroll:none settle:none"`,
		`hx-boost="true"`,
		`htmx.org`,
		`htmx:afterSwap`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("page missing %q", want)
		}
	}
	for _, bad := range []string{
		`syncNavActive`,
		`scheduleAutoRefresh`,
		`location.reload`,
		`sessionStorage`,
		`snapshotDashState`,
		`restoreDashState`,
		`setInterval`,
		`/fragment`,
		`hx-target="#data"`,
	} {
		if strings.Contains(out, bad) {
			t.Fatalf("page should not contain %q", bad)
		}
	}
}

func TestNavLinksPlainHref(t *testing.T) {
	out := navHTML(panel.ViewInfra)
	if !strings.Contains(out, `href="/s/infra"`) || !strings.Contains(out, `dash-nav__link--active`) {
		t.Fatal("nav should mark active section with plain href links")
	}
	if strings.Contains(out, `hx-get=`) || strings.Contains(out, `hx-target=`) {
		t.Fatal("nav links should rely on body hx-boost, not per-link HTMX attrs")
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
