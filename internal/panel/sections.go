package panel

import "github.com/arkantos1482/cosmos-monitor/internal/model"

// View identifies a dashboard page (home or one of seven sections).
type View string

const (
	ViewHome          View = "home"
	ViewInfra         View = "infra"
	ViewNode          View = "node"
	ViewValidators    View = "validators"
	ViewLocalValidator View = "local"
	ViewEconomics     View = "economics"
	ViewGovernance    View = "governance"
	ViewEVM           View = "evm"
)

// NavItem is a sidebar link target.
type NavItem struct {
	View  View
	Label string
	Path  string
}

// Nav lists home plus the seven monitoring sections (display order).
var Nav = []NavItem{
	{ViewHome, "Overview", "/"},
	{ViewInfra, "Infrastructure", "/s/infra"},
	{ViewNode, "Node", "/s/node"},
	{ViewValidators, "Validator set", "/s/validators"},
	{ViewLocalValidator, "This validator", "/s/local"},
	{ViewEconomics, "Economics", "/s/economics"},
	{ViewGovernance, "Governance", "/s/governance"},
	{ViewEVM, "EVM JSON-RPC", "/s/evm"},
}

// ParseView maps a URL segment or query value to a View. Unknown values become ViewHome.
func ParseView(s string) View {
	switch View(s) {
	case ViewHome, ViewInfra, ViewNode, ViewValidators, ViewLocalValidator, ViewEconomics, ViewGovernance, ViewEVM:
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
	case ViewLocalValidator:
		writeLocalValidator(w, d)
	case ViewEconomics:
		writeEconomics(w, d)
	case ViewGovernance:
		writeGovernance(w, d)
	case ViewEVM:
		writeEVMSection(w, d)
	default:
		writeHome(w, d)
	}
}
