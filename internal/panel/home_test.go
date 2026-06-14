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
		UnclaimedDelegator:  "0.006 PMT",
		UnclaimedCommission: "0.0006 PMT",
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
		`rewards-summary`,
		`class="fm-summary"`,
		`class="gov-summary"`,
		`class="infra-summary"`,
		`class="node-summary"`,
		`class="evm-summary"`,
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
	if strings.Contains(out, `dash-overview__group-title`) {
		t.Fatal("overview should not render group headings; each card has its own title")
	}
	runtimeIdx := strings.Index(out, `dash-overview__group--runtime`)
	validatorIdx := strings.Index(out, `dash-overview__group--validator`)
	economicsIdx := strings.Index(out, `dash-overview__group--economics`)
	governanceIdx := strings.Index(out, `dash-overview__group--governance`)
	if runtimeIdx < 0 || validatorIdx < 0 || economicsIdx < 0 || governanceIdx < 0 ||
		runtimeIdx > validatorIdx || validatorIdx > economicsIdx || economicsIdx > governanceIdx {
		t.Fatal("overview should show Runtime → Validator → Economics → Governance groups")
	}
	
	feeCard := strings.Index(out, `dash-overview__card--feemarket`)
	rewardsCard := strings.Index(out, `dash-overview__card--rewards`)
	distCard := strings.Index(out, `dash-overview__card--distribution`)
	if feeCard < 0 || rewardsCard < 0 || distCard < 0 {
		t.Fatal("overview should include fee market, rewards, and distribution cards")
	}
	if feeCard > rewardsCard || rewardsCard > distCard {
		t.Fatal("economics overview cards should be ordered fee market → rewards → distribution")
	}
	if !strings.Contains(out[feeCard:rewardsCard], `fm-summary`) {
		t.Fatal("fee market overview card should include fee summary")
	}
	if !strings.Contains(out[rewardsCard:distCard], `rewards-summary`) {
		t.Fatal("rewards overview card should show rewards summary KPIs")
	}
	if !strings.Contains(out[distCard:], `unclaimed-stack__head-label">unclaimed total`) {
		t.Fatal("distribution overview card should include unclaimed total summary")
	}
}

func TestOverviewReusesSectionSummaries(t *testing.T) {
	d := model.Report{
		Moniker: "node1", Synced: true, BlockHeight: "100",
		BondedPct: 100, BondedCount: 4, BondedAmt: "4M PMT",
		SlashWindow: "10000", MinSigned: 95, JailedCount: 1,
		Local: model.LocalValidator{
			IsValidator: true, VPPercent: 25, Status: "BOND_STATUS_BONDED",
			Commission: 10, Missed: 2, MaxMissed: 50,
		},
		CommunityTax: "2%", CommunityPool: "0.5 PMT",
		UnclaimedDelegator: "0.006 PMT", UnclaimedCommission: "0.0006 PMT",
	}
	overview := BuildView(ViewHome, d)
	for _, tc := range []struct {
		section View
		markers []string
	}{
		{ViewStaking, []string{
			`staking-summary__kpi-label">voting power`,
			`staking-summary__kpi-label">bonded`,
		}},
		{ViewSlashing, []string{
			`slashing-summary__kpi-label">signing health`,
			`slashing-summary__kpi-label">jailed`,
		}},
		{ViewDistribution, []string{
			`unclaimed-stack__head-label">unclaimed total`,
		}},
	} {
		section := BuildView(tc.section, d)
		for _, marker := range tc.markers {
			if !strings.Contains(section, marker) {
				t.Fatalf("section %s missing summary marker %q", tc.section, marker)
			}
			if !strings.Contains(overview, marker) {
				t.Fatalf("overview should reuse section %s summary marker %q", tc.section, marker)
			}
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
	if !strings.Contains(out, `class="node-summary"`) {
		t.Fatal("node view should include embedded summary")
	}
	if !strings.Contains(out, `class="dash-heading">2. VALIDATOR</h2>`) {
		t.Fatal("node view should only render validator section")
	}
	for _, sub := range []string{
		`class="dash-subheading">Proposer</h3>`,
		"cosmos peers",
		"CometBFT (consensus + networking)",
		`class="dash-layer__title">Validator set</h3>`,
	} {
		if !strings.Contains(out, sub) {
			t.Fatalf("node view missing subsection %q", sub)
		}
	}
	for _, gone := range []string{
		`class="id-board"`,
		`id-board__row--consensus`,
		`id-board__row--p2p`,
		`id-board__row--account`,
		`id-board__row--operator`,
		"evm peers",
		`node-summary__label">voting power`,
		`node-summary__label">signing`,
	} {
		if strings.Contains(out, gone) {
			t.Fatalf("node view should not include %q", gone)
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

func TestOverviewDataSourcesProvenance(t *testing.T) {
	d := model.Report{
		Moniker: "node1", Synced: true, BlockHeight: "100",
		NodeRunning: true, BondedCount: 4, PMTEnabled: true,
		Exchanges: []model.SourceExchange{
			{
				Kind: "http", Method: "GET",
				URL: "http://localhost:26657/status", Request: "(none)",
				Response: `{"result":{}}`, OK: true, Latency: "1ms",
			},
			{
				Kind: "http", Method: "GET",
				URL: "http://localhost:1317/cosmos/distribution/v1beta1/params",
				Request: "(none)",
				Response: `{"params":{}}`, OK: true, Latency: "2ms",
			},
		},
	}
	out := BuildViewWithOptions(ViewHome, d, Options{ShowSources: true})
	if !strings.Contains(out, `class="dash-sources"`) {
		t.Fatal("overview should include data sources footer when enabled")
	}
	if !strings.Contains(out, `/status`) || !strings.Contains(out, `distribution/v1beta1/params`) {
		t.Fatal("overview data sources should include all traced endpoints")
	}
	if !strings.Contains(out, `id="dash-sources-overview"`) {
		t.Fatal("overview data sources need stable id for hx-preserve across refresh")
	}
	governanceIdx := strings.Index(out, `dash-overview__group--governance`)
	sourcesIdx := strings.Index(out, `class="dash-sources"`)
	if governanceIdx < 0 || sourcesIdx < 0 || sourcesIdx < governanceIdx {
		t.Fatal("overview data sources should render after overview content")
	}
	outHidden := BuildView(ViewHome, d)
	if strings.Contains(outHidden, `class="dash-sources"`) {
		t.Fatal("overview should hide data sources by default")
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
		Exchanges: []model.SourceExchange{{
			Kind: "http", Method: "GET",
			URL: "http://localhost:26657/status", Request: "(none)",
			Response: `{"result":{}}`, OK: true, Latency: "1ms",
		}},
	}
	out := BuildViewWithOptions(ViewNode, d, Options{ShowSources: true})
	if !strings.Contains(out, `class="dash-sources"`) {
		t.Fatal("validator section should include data sources footer when enabled")
	}
	if !strings.Contains(out, `dash-sources__tag">req`) {
		t.Fatal("validator data sources should show raw request")
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
		{ViewStaking, `staking-summary`, "At a glance"},
		{ViewSlashing, `slashing-summary`, "At a glance"},
		{ViewRewards, `rewards-summary`, "At a glance"},
		{ViewDistribution, `dist-summary`, "At a glance"},
		{ViewFeemarket, `class="fm-summary"`, "At a glance"},
		{ViewGovernance, `class="gov-summary"`, "At a glance"},
		{ViewInfra, `class="infra-summary"`, "At a glance"},
		{ViewNode, `class="node-summary"`, "At a glance"},
		{ViewEVM, `evm-summary__probe`, "At a glance"},
	} {
		out := BuildView(tc.view, d)
		if !strings.Contains(out, tc.want) {
			t.Fatalf("view %s missing summary marker %q", tc.view, tc.want)
		}
		if !strings.Contains(out, `dash-section__summary-card__title">Summary</h3>`) {
			t.Fatalf("view %s missing summary card title", tc.view)
		}
		if !strings.Contains(out, `class="dash-section__summary-card"`) {
			t.Fatalf("view %s missing summary card wrapper", tc.view)
		}
		if tc.gone != "" && strings.Contains(out, tc.gone) {
			t.Fatalf("view %s should not contain stale %q", tc.view, tc.gone)
		}
		if !sectionIntroBeforeSummary(out) {
			t.Fatalf("view %s lead and summary title should precede summary card", tc.view)
		}
		if !summaryAfterHeading(out) {
			t.Fatalf("view %s summary should follow section heading", tc.view)
		}
	}
}

func sectionIntroBeforeSummary(out string) bool {
	lead := strings.Index(out, `class="dash-callout dash-callout--note note"`)
	title := strings.Index(out, `dash-section__summary-card__title">Summary</h3>`)
	summary := strings.Index(out, `-summary`)
	if lead < 0 || title < 0 || summary < 0 {
		return false
	}
	return lead < title && title < summary
}

func summaryAfterHeading(out string) bool {
	heading := strings.Index(out, `class="dash-heading"`)
	summary := strings.Index(out, `-summary`)
	if heading < 0 || summary < 0 {
		return heading < 0 && summary < 0
	}
	return summary > heading
}
