package html

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/panel"
)

func TestFullPageNavAndFragmentURL(t *testing.T) {
	out := FullPage("node1", panel.ViewEconomics, "<p>body</p>")
	for _, want := range []string{
		`id="dash-view"`,
		`value="economics"`,
		`id="dash-refresh"`,
		`hx-include="#dash-view"`,
		`hx-push-url="/s/economics"`,
		`id="dash-nav"`,
		`dash-nav__link--active`,
		`href="/s/economics"`,
		`Economics`,
		`<p>body</p>`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("page missing %q", want)
		}
	}
	if strings.Contains(out, `id="data" hx-get=`) {
		t.Fatal("poll trigger should not live on #data")
	}
}

func TestWrapFragmentOOB(t *testing.T) {
	out := WrapFragment(panel.ViewInfra, `<section>infra</section>`)
	for _, want := range []string{
		`<section>infra</section>`,
		`id="dash-view"`,
		`value="infra"`,
		`hx-swap-oob="true"`,
		`id="dash-nav"`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("page missing %q", want)
		}
	}
}
