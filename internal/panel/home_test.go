package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestBuildHomeSummaryCards(t *testing.T) {
	d := model.Report{
		Moniker: "node1", Synced: true, BlockHeight: "100",
		NodeRunning: true, BondedCount: 4, PMTEnabled: true,
		EVMChainID: 290290, EVMSynced: true,
	}
	out := BuildView(ViewHome, d)
	for _, want := range []string{
		`class="dash-cards"`,
		`href="/s/infra"`,
		`href="/s/evm"`,
		"Infrastructure",
		"EVM JSON-RPC",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("home view missing %q", want)
		}
	}
}

func TestBuildViewSingleSection(t *testing.T) {
	d := model.Report{Moniker: "node1", Synced: true, BlockHeight: "1"}
	out := BuildView(ViewNode, d)
	if !strings.Contains(out, `class="dash-heading">2. NODE</h2>`) {
		t.Fatal("node view should only render node section")
	}
	if strings.Contains(out, `class="dash-heading">1. INFRASTRUCTURE</h2>`) {
		t.Fatal("node view should not include infrastructure")
	}
}
