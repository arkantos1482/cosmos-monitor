package html

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/panel"
)

func TestFullPageNavAndFragmentURL(t *testing.T) {
	out := FullPage("node1", panel.ViewEconomics, "<p>body</p>")
	for _, want := range []string{
		`hx-get="/fragment?view=economics"`,
		`hx-push-url="/s/economics"`,
		`function syncDashView`,
		`function viewFromPath`,
		`dash-nav__link--active`,
		`href="/s/economics"`,
		`Economics`,
		`<p>body</p>`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("page missing %q", want)
		}
	}
}
