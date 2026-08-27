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
		`class="dash-page"`,
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
		`class="dash-page dash-page--wallet"`,
	} {
		if strings.Contains(out, bad) {
			t.Fatalf("page should not contain %q", bad)
		}
	}
}

func TestNavOmitsDelegate(t *testing.T) {
	out := navHTML(panel.ViewHome)
	if strings.Contains(out, `href="/delegate"`) || strings.Contains(out, `dash-nav__link--delegate`) {
		t.Fatal("sidebar should not list Delegate; entry is the overview CTA")
	}
	if strings.Contains(out, `href="/s/delegate"`) {
		t.Fatal("Delegate must not live under /s/")
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
	if !strings.Contains(out, `class="dash-nav__more"`) || !strings.Contains(out, `class="dash-nav__sections"`) {
		t.Fatal("nav should wrap section links in a compact Sections disclosure")
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
	if dataURL(panel.ViewDelegate) != "/delegate" {
		t.Fatal("delegate data URL should be /delegate")
	}
}

func TestFullPageDelegateHasNoPoll(t *testing.T) {
	status := panel.RenderStatusStrip(model.Report{Moniker: "node1", Synced: true, BlockHeight: "1"})
	out := FullPage("node1", panel.ViewDelegate, status, panel.BuildView(panel.ViewDelegate, model.Report{}))
	for _, want := range []string{
		`id="delegate-app"`,
		panel.StakingPrecompile,
		`290290`,
		`wallet`,
		`ethers@6`,
		`pmtop — node1 · Delegate`,
		`class="dash-page dash-page--wallet"`,
		`max-width: 1024px`,
		`max-width: 860px`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("delegate page missing %q", want)
		}
	}
	if strings.Contains(out, `max-width: 1200px`) {
		t.Fatal("layout must not collapse the fleet sidebar at 1200px")
	}
	if strings.Contains(out, `hx-trigger="every 5s"`) {
		t.Fatal("delegate page must not poll #data every 5s")
	}
	if strings.Contains(out, `live · 5s`) {
		t.Fatal("delegate page must not claim a 5s live refresh")
	}
}
