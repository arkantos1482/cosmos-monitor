package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestInfraLayout(t *testing.T) {
	d := model.Report{
		NodeRunning: true, MemPct: 72, DiskPct: 45,
		Load1: 1.2, Load5: 0.9, Load15: 0.7,
		MemUsed: "11 GiB", MemTotal: "16 GiB",
		DiskUsed: "120 GiB", DiskTotal: "256 GiB",
		NodeCPU: "12.3%", NodeMemUsed: "2.1 GiB", NodeMemTotal: "4 GiB",
		Restarts: 2, NodeUptime: "3d 4h",
	}
	out := BuildView(ViewInfra, d)

	for _, want := range []string{
		`class="infra-summary__top"`,
		`class="infra-summary__kpis"`,
		`class="data-table infra-compare"`,
		`Host vs container`,
		`evmd-node`,
		`badge--ok">running</span>`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("infra view missing %q", want)
		}
	}
	for _, gone := range []string{
		`infra-summary__row`,
		`infra-summary__load`,
		`<h3 class="dash-subheading">OS</h3>`,
		`<h3 class="dash-subheading">Container</h3>`,
	} {
		if strings.Contains(out, gone) {
			t.Fatalf("infra view should not contain stale %q", gone)
		}
	}
}
