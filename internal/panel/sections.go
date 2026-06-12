package panel

import "github.com/arkantos1482/cosmos-monitor/internal/model"

// View identifies a dashboard page (home or one of seven sections).
type View string

const (
	ViewHome      View = "home"
	ViewInfra     View = "infra"
	ViewNode      View = "node"
	ViewStaking   View = "staking"
	ViewRewards   View = "rewards"
	ViewEconomics View = "economics"
	ViewFeemarket View = "feemarket"
	ViewGovernance View = "governance"
	ViewEVM       View = "evm"
)

// NavScope groups sidebar and home cards into chain-wide vs this-node views.
type NavScope string

const (
	NavScopeChain NavScope = "chain"
	NavScopeNode  NavScope = "node"
)

// NavScopeLabel is the human heading for a nav/home group.
func NavScopeLabel(s NavScope) string {
	switch s {
	case NavScopeChain:
		return "Chain"
	case NavScopeNode:
		return "This node"
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

// Nav lists home plus monitoring sections (this node group, then chain).
var Nav = []NavItem{
	{ViewHome, "Overview", "/", ""},
	{ViewInfra, "Infrastructure", "/s/infra", NavScopeNode},
	{ViewNode, "Validator", "/s/node", NavScopeNode},
	{ViewEVM, "EVM JSON-RPC", "/s/evm", NavScopeNode},
	{ViewStaking, "Staking", "/s/staking", NavScopeChain},
	{ViewRewards, "Rewards", "/s/rewards", NavScopeChain},
	{ViewEconomics, "Economics", "/s/economics", NavScopeChain},
	{ViewFeemarket, "Fee market", "/s/feemarket", NavScopeChain},
	{ViewGovernance, "Governance", "/s/governance", NavScopeChain},
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
	case ViewHome, ViewInfra, ViewNode, ViewStaking, ViewRewards, ViewEconomics, ViewFeemarket, ViewGovernance, ViewEVM:
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
	case ViewRewards:
		writeRewards(w, d)
	case ViewEconomics:
		writeEconomics(w, d)
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
