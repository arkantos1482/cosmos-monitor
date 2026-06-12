package panel

import "github.com/arkantos1482/cosmos-monitor/internal/model"

// View identifies a dashboard page (home or one of the chain/node sections).
type View string

const (
	ViewHome       View = "home"
	ViewInfra      View = "infra"
	ViewNode       View = "node"
	ViewStaking    View = "staking"
	ViewSlashing   View = "slashing"
	ViewRewards       View = "rewards"
	ViewDistribution  View = "distribution"
	ViewFeemarket     View = "feemarket"
	ViewGovernance View = "governance"
	ViewEVM        View = "evm"
)

// NavScope groups sidebar and home cards by operational concern.
type NavScope string

const (
	NavScopeRuntime    NavScope = "runtime"
	NavScopeValidator  NavScope = "validator"
	NavScopeEconomics  NavScope = "economics"
	NavScopeGovernance NavScope = "governance"
)

// NavScopeLabel is the human heading for a nav/home group.
func NavScopeLabel(s NavScope) string {
	switch s {
	case NavScopeRuntime:
		return "Runtime"
	case NavScopeValidator:
		return "Validator"
	case NavScopeEconomics:
		return "Economics"
	case NavScopeGovernance:
		return "Governance"
	default:
		return ""
	}
}

// NavItem is a sidebar link target.
type NavItem struct {
	View  View
	Label string
	Path  string
	Scope NavScope // empty for Overview only
}

// Nav lists home plus monitoring sections (runtime → validator → economics → governance).
var Nav = []NavItem{
	{ViewHome, "Overview", "/", ""},
	{ViewInfra, "Infrastructure", "/s/infra", NavScopeRuntime},
	{ViewEVM, "EVM JSON-RPC", "/s/evm", NavScopeRuntime},
	{ViewNode, "Validator", "/s/node", NavScopeValidator},
	{ViewStaking, "Staking", "/s/staking", NavScopeEconomics},
	{ViewSlashing, "Slashing", "/s/slashing", NavScopeEconomics},
	{ViewRewards, "Rewards", "/s/rewards", NavScopeEconomics},
	{ViewDistribution, "Distribution", "/s/distribution", NavScopeEconomics},
	{ViewFeemarket, "Fee market", "/s/feemarket", NavScopeEconomics},
	{ViewGovernance, "Governance", "/s/governance", NavScopeGovernance},
}

// NavLabelForSlug returns the sidebar label for a section slug (e.g. "validators").
func NavLabelForSlug(slug string) string {
	path := "/s/" + slug
	for _, item := range Nav {
		if item.Path == path {
			return item.Label
		}
	}
	return ""
}

// ParseView maps a URL segment or query value to a View. Unknown values become ViewHome.
func ParseView(s string) View {
	switch View(s) {
	case "local", "validators": // legacy paths — merged into Validator
		return ViewNode
	case "economics": // legacy path — merged into rewards
		return ViewRewards
	case ViewHome, ViewInfra, ViewNode, ViewStaking, ViewSlashing, ViewRewards, ViewDistribution, ViewFeemarket, ViewGovernance, ViewEVM:
		return View(s)
	default:
		return ViewHome
	}
}

func writeView(w Writer, v View, d model.Report) {
	switch v {
	case ViewInfra:
		writeInfra(w, d)
	case ViewNode:
		writeNode(w, d)
	case ViewStaking:
		writeStaking(w, d)
	case ViewSlashing:
		writeSlashing(w, d)
	case ViewRewards:
		writeRewards(w, d)
	case ViewDistribution:
		writeDistribution(w, d)
	case ViewFeemarket:
		writeFeemarket(w, d)
	case ViewGovernance:
		writeGovernance(w, d)
	case ViewEVM:
		writeEVMSection(w, d)
	default:
		writeOverview(w, d)
	}
}
