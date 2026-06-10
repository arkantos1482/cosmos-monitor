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
		MemPct: 42, DiskPct: 55, Load1: 0.5,
	}
	out := BuildView(ViewHome, d)
	for _, want := range []string{
		`class="dash-status"`,
		`dash-home__group--chain`,
		`dash-home__group--node`,
		`class="dash-cards dash-cards--bento"`,
		`dash-card--span2`,
		`dash-card--infra`,
		`dash-card__gauges`,
		`mini-gauge`,
		`dash-card__footer`,
		`View section →`,
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
	d := model.Report{
		Moniker: "node1", Synced: true, BlockHeight: "1",
		Local: model.LocalValidator{IsValidator: true, Moniker: "node1", Status: "BOND_STATUS_BONDED"},
	}
	out := BuildView(ViewNode, d)
	if strings.Contains(out, `class="dash-status"`) {
		t.Fatal("node view should not include status strip")
	}
	if !strings.Contains(out, `class="dash-heading">2. VALIDATOR</h2>`) {
		t.Fatal("node view should only render validator section")
	}
	for _, sub := range []string{
		`class="id-board"`,
		"id-board__row--account",
		"id-board__row--operator",
		"id-board__row--consensus",
		"id-board__row--p2p",
		"Application (Cosmos SDK / ABCI state)",
		"CometBFT (consensus + networking)",
		"Staking", "Rewards", "Slashing",
		"Live state", "Proposer", "P2P &amp; RPC",
	} {
		if !strings.Contains(out, sub) {
			t.Fatalf("node view missing subsection %q", sub)
		}
	}
	if strings.Contains(out, `THIS VALIDATOR`) {
		t.Fatal("node view should not include legacy this-validator section")
	}
	if !strings.Contains(out, `dash-section--node`) {
		t.Fatal("node view should have section accent class")
	}
	if strings.Contains(out, `class="dash-heading">1. INFRASTRUCTURE</h2>`) {
		t.Fatal("node view should not include infrastructure")
	}
}

func TestStatusStripOnlyOnHome(t *testing.T) {
	d := model.Report{Moniker: "n", Synced: true, BlockHeight: "1", BaseFee: "1000"}
	out := BuildView(ViewHome, d)
	if !strings.Contains(out, `class="dash-status"`) {
		t.Fatal("home view should include status strip")
	}
	sections := []View{ViewInfra, ViewNode, ViewValidators, ViewEconomics, ViewFeemarket, ViewGovernance, ViewEVM}
	for _, v := range sections {
		out := BuildView(v, d)
		if strings.Contains(out, `class="dash-status"`) {
			t.Fatalf("view %s should not include status strip", v)
		}
	}
}
