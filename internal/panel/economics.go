package panel

import (
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeEconomics(w Writer, d model.Report) {
	w.Section("3. ECONOMICS")
	writeEconomicsSummary(w, d, SummaryEmbedded)
	w.Em("Chain-wide distribution — how block rewards route through `fee_collector` and `x/distribution` to the community pool and validators. Reward sources → § Rewards.")

	writeEconomicsOverview(w, d)
	w.Hint(economicsSourcesHint())
	w.BlankLine()
}

func economicsSourcesHint() string {
	return "`community tax`, `community pool` → REST GET /cosmos/distribution/v1beta1/params, /cosmos/distribution/v1beta1/community_pool; " +
		"`unclaimed delegator`, `unclaimed commission` → REST GET /cosmos/distribution/v1beta1/validators/{valoper}/outstanding_rewards, …/commission (summed across validators); " +
		"`module account balances` → REST GET /cosmos/bank/v1beta1/balances/{address}; " +
		"`module account addresses` → REST GET /cosmos/auth/v1beta1/module_accounts; " +
		"`fee_collector cleared`, `unclaimed check` → derived (x/bank balances, outstanding sums)."
}

func writeEVMSection(w Writer, d model.Report) {
	w.Section("3. EVM JSON-RPC")
	writeEVMSummary(w, d, SummaryEmbedded)
	w.Em("Wallet and dApp connectivity (`eth_*`, `net_*`, `txpool_*`) on this node's JSON-RPC.")
	writeEVMRPCSection(w, d)
	w.BlankLine()
}
