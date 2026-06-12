package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestBuildOverviewStack(t *testing.T) {
	d := model.Report{
		Moniker: "node1", Synced: true, BlockHeight: "100",
		NodeRunning: true, BondedCount: 4, PMTEnabled: true,
		EVMChainID: 290290, EVMSynced: true, EVMRPCOk: true,
		MemPct: 42, DiskPct: 55, Load1: 0.5,
		RPCProbeOK: 4, RPCProbeTotal: 4,
		RPCProbes: []model.RPCProbe{{Method: "eth_blockNumber", OK: true}},
		PMTRate: "0.1 PMT/block",
		CommunityTax: "2%", 
		CommunityPool: "0.5 PMT",
		ModuleAccounts: []model.ModuleAccountRow{
			{Name: "fee_collector", Balance: "0 PMT", Role: "fees"},
		},
	}
	out := BuildView(ViewHome, d)
	for _, want := range []string{
		`class="dash-overview"`,
		`dash-overview__group--runtime`,
		`dash-overview__group--validator`,
		`dash-overview__group--economics`,
		`dash-overview__group--governance`,
		`dash-overview__stack`,
		`class="dash-overview__card-title">Infrastructure</p>`,
		`class="dash-overview__card-title">Staking</p>`,
		`dash-overview__card--staking`,
		`dash-overview__card--slashing`,
		`dash-overview__card--rewards`,
		`dash-overview__card--distribution`,
		`dash-overview__card--feemarket`,
		`dash-overview__card--governance`,
		`dash-overview__card--infra`,
		`dash-overview__card--node`,
		`dash-overview__card--evm`,
		`href="/s/infra"`,
		`href="/s/evm"`,
		`staking-summary`,
		`eco-summary`,
		`class="fee-summary"`,
		`class="gov-summary"`,
		`class="infra-summary"`,
		`class="node-summary"`,
		`class="evm-summary"`,
		`eco-summary`,
		`mini-gauge`,
		`evm-summary__probe`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("overview missing %q", want)
		}
	}
	if strings.Contains(out, `class="dash-status"`) {
		t.Fatal("overview view should not include status strip in main content")
	}
	if strings.Contains(out, `dash-cards--bento`) {
		t.Fatal("overview should not use legacy bento cards")
	}
	if strings.Contains(out, `dash-overview__footer`) || strings.Contains(out, `View section →`) {
		t.Fatal("overview cards should not include footer CTA bar")
	}
	runtimeIdx := strings.Index(out, `dash-overview__group--runtime`)
	validatorIdx := strings.Index(out, `dash-overview__group--validator`)
	economicsIdx := strings.Index(out, `dash-overview__group--economics`)
	governanceIdx := strings.Index(out, `dash-overview__group--governance`)
	if runtimeIdx < 0 || validatorIdx < 0 || economicsIdx < 0 || governanceIdx < 0 ||
		runtimeIdx > validatorIdx || validatorIdx > economicsIdx || economicsIdx > governanceIdx {
		t.Fatal("overview should show Runtime → Validator → Economics → Governance groups")
	}
	
	distCard := strings.Index(out, `dash-overview__card--distribution`)
	feeCard := strings.Index(out, `dash-overview__card--feemarket`)
	rewardsCard := strings.Index(out, `dash-overview__card--rewards`)
	if distCard < 0 || feeCard < 0 || rewardsCard < 0 {
		t.Fatal("overview should include distribution, fee market, and rewards cards")
	}
	if distCard > feeCard || feeCard > rewardsCard {
		t.Fatal("economics overview cards should be ordered distribution → fee market → rewards")
	}
	if !strings.Contains(out[distCard:feeCard], "Community pool:") {
		t.Fatal("distribution overview card should include community pool summary")
	}
	if !strings.Contains(out[feeCard:rewardsCard], `fee-summary`) {
		t.Fatal("fee market overview card should include fee summary")
	}
	if !strings.Contains(out[rewardsCard:], `eco-summary--compact`) {
		t.Fatal("rewards overview card should show compact PMT/inflation summary")
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
	if !strings.Contains(out, `class="node-summary"`) {
		t.Fatal("node view should include embedded summary")
	}
	if !strings.Contains(out, `class="dash-heading">2. VALIDATOR</h2>`) {
		t.Fatal("node view should only render validator section")
	}
	for _, sub := range []string{
		`class="id-board"`,
		"CometBFT (consensus + networking)",
		`class="dash-layer__title">Validator set</h3>`,
	} {
		if !strings.Contains(out, sub) {
			t.Fatalf("node view missing subsection %q", sub)
		}
	}
	if strings.Contains(out, `dash-overview__card`) {
		t.Fatal("section view summary should be embedded, not overview card link")
	}
	idx := strings.Index(out, `class="dash-heading">2. VALIDATOR</h2>`)
	sumIdx := strings.Index(out, `class="node-summary"`)
	if idx < 0 || sumIdx < 0 || sumIdx < idx {
		t.Fatal("node summary should appear after section heading")
	}
}

func TestNodeSectionDataSourcesProvenance(t *testing.T) {
	d := model.Report{
		Moniker: "node1", Synced: true, BlockHeight: "1", NodeID: "abc",
		ListenAddr: "tcp://0.0.0.0:26656", RpcListenAddr: "tcp://0.0.0.0:26657",
		Local: model.LocalValidator{
			IsValidator: true, Moniker: "node1", Status: "BOND_STATUS_BONDED",
			VotingPower: "100", VPPercent: 25, Commission: 10,
		},
	}
	out := BuildViewWithOptions(ViewNode, d, Options{ShowSources: true})
	if !strings.Contains(out, `class="dash-sources"`) {
		t.Fatal("validator section should include data sources footer when enabled")
	}
	outHidden := BuildView(ViewNode, d)
	if strings.Contains(outHidden, `class="dash-sources"`) {
		t.Fatal("validator section should hide data sources by default")
	}
}

func TestStatusStripNotInBuildView(t *testing.T) {
	d := model.Report{Moniker: "n", Synced: true, BlockHeight: "1", BaseFee: "1000"}
	out := BuildView(ViewHome, d)
	if strings.Contains(out, `id="dash-status"`) {
		t.Fatal("BuildView should not include global status bar")
	}
	strip := RenderStatusStrip(d)
	if !strings.Contains(strip, `id="dash-status"`) {
		t.Fatal("RenderStatusStrip should include dash-status id")
	}
	oob := BuildStatusOOB(d)
	if !strings.Contains(oob, `hx-swap-oob="true"`) {
		t.Fatal("BuildStatusOOB should include hx-swap-oob")
	}
}

func TestSectionSummariesEmbedded(t *testing.T) {
	d := model.Report{
		Moniker: "n", Synced: true, BlockHeight: "1", BondedCount: 4,
		PMTEnabled: true, PMTRate: "0.1 PMT/block",
		EVMRPCOk: true, EVMSynced: true, RPCProbeOK: 1, RPCProbeTotal: 1,
		RPCProbes: []model.RPCProbe{{Method: "eth_chainId", OK: true}},
		CommunityTax: "2%",
		CommunityPool: "0.5 PMT",
		ModuleAccounts: []model.ModuleAccountRow{
			{Name: "fee_collector", Balance: "1 PMT", Role: "fees"},
		},
	}
	for _, tc := range []struct {
		view View
		want string
		gone string
	}{
		{ViewStaking, `staking-summary`, `class="dash-subheading">Summary</h3>`},
		{ViewSlashing, `slashing-summary`, `class="dash-subheading">Summary</h3>`},
		{ViewRewards, `eco-domains`, "At a glance"},
		{ViewDistribution, `dist-summary`, "At a glance"},
		{ViewFeemarket, `class="fee-summary"`, ""},
		{ViewGovernance, `class="gov-summary"`, ""},
		{ViewInfra, `class="infra-summary"`, ""},
		{ViewNode, `class="node-summary"`, `class="dash-subheading">Summary</h3>`},
		{ViewEVM, `evm-summary__probe`, "<strong>RPC:"},
	} {
		out := BuildView(tc.view, d)
		if !strings.Contains(out, tc.want) {
			t.Fatalf("view %s missing summary marker %q", tc.view, tc.want)
		}
		if tc.gone != "" && strings.Contains(out, tc.gone) {
			t.Fatalf("view %s should not contain duplicate %q", tc.view, tc.gone)
		}
		if !summaryAfterHeading(out) {
			t.Fatalf("view %s summary should follow section heading", tc.view)
		}
	}
}

func summaryAfterHeading(out string) bool {
	heading := strings.Index(out, `class="dash-heading"`)
	summary := strings.Index(out, `-summary`)
	if heading < 0 || summary < 0 {
		return heading < 0 && summary < 0
	}
	return summary > heading
}
