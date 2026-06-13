package html

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/panel"
)

func TestFullPageHTMXShell(t *testing.T) {
	status := panel.RenderStatusStrip(model.Report{Moniker: "node1", Synced: true, BlockHeight: "1"})
	out := FullPage("node1", panel.ViewRewards, status, "<p>body</p>")
	for _, want := range []string{
		`id="dash-status"`,
		`id="dash-nav"`,
		`dash-nav__link--active`,
		`dash-nav__link--rewards`,
		`dash-nav__icon`,
		`href="/s/rewards"`,
		`Rewards`,
		`<p>body</p>`,
		`class="dash-content"`,
		`id="data"`,
		`hx-get="/s/rewards"`,
		`hx-trigger="every 5s"`,
		`hx-swap="innerMorph settle:0.15s focus-scroll:false"`,
		`htmx.org@4.0.0-beta4`,
		`htmx.config.implicitInheritance = true`,
		`hx-boost="true"`,
		`htmx.org`,
		`--accent-rewards`,
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
		`dash-nav__link--economics`,
		`href="/s/economics"`,
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
	if !strings.Contains(out, `dash-nav__link--infra`) {
		t.Fatal("nav should include section accent class")
	}
	if strings.Contains(out, `hx-get=`) || strings.Contains(out, `hx-target=`) {
		t.Fatal("nav links should rely on body hx-boost, not per-link HTMX attrs")
	}
}

func TestNavGroupOrder(t *testing.T) {
	out := navHTML(panel.ViewHome)
	runtimeIdx := strings.Index(out, `>Runtime</p>`)
	validatorIdx := strings.Index(out, `>Validator</p>`)
	economicsIdx := strings.Index(out, `>Economics</p>`)
	governanceIdx := strings.Index(out, `>Governance</p>`)
	if runtimeIdx < 0 || validatorIdx < 0 || economicsIdx < 0 || governanceIdx < 0 ||
		runtimeIdx > validatorIdx || validatorIdx > economicsIdx || economicsIdx > governanceIdx {
		t.Fatal("nav should list Runtime → Validator → Economics → Governance groups")
	}
}

func TestDataURL(t *testing.T) {
	if dataURL(panel.ViewHome) != "/" {
		t.Fatal("home data URL should be /")
	}
	if dataURL(panel.ViewRewards) != "/s/rewards" {
		t.Fatal("rewards data URL should be /s/rewards")
	}
}
