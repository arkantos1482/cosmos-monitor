package panel

import "github.com/arkantos1482/cosmos-monitor/internal/model"

// View identifies a dashboard page (home or one of seven sections).
type View string

const (
	ViewHome       View = "home"
	ViewInfra      View = "infra"
	ViewNode       View = "node"
	ViewValidators View = "validators"
	ViewEconomics  View = "economics"
	ViewFeemarket  View = "feemarket"
	ViewGovernance View = "governance"
	ViewEVM        View = "evm"
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

// Nav lists home plus monitoring sections (chain group, then this node).
var Nav = []NavItem{
	{ViewHome, "Overview", "/", ""},
	{ViewValidators, "Validator set", "/s/validators", NavScopeChain},
	{ViewEconomics, "Economics", "/s/economics", NavScopeChain},
	{ViewFeemarket, "Fee market", "/s/feemarket", NavScopeChain},
	{ViewGovernance, "Governance", "/s/governance", NavScopeChain},
	{ViewInfra, "Infrastructure", "/s/infra", NavScopeNode},
	{ViewNode, "Validator", "/s/node", NavScopeNode},
	{ViewEVM, "EVM JSON-RPC", "/s/evm", NavScopeNode},
}

// ParseView maps a URL segment or query value to a View. Unknown values become ViewHome.
func ParseView(s string) View {
	switch View(s) {
	case "local": // legacy path — merged into Validator
		return ViewNode
	case ViewHome, ViewInfra, ViewNode, ViewValidators, ViewEconomics, ViewFeemarket, ViewGovernance, ViewEVM:
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
	case ViewValidators:
		writeValidators(w, d)
	case ViewEconomics:
		writeEconomics(w, d)
	case ViewFeemarket:
		writeFeemarket(w, d)
	case ViewGovernance:
		writeGovernance(w, d)
	case ViewEVM:
		writeEVMSection(w, d)
	default:
		writeStatusStrip(w, d)
		writeHome(w, d)
	}
}
