package panel

import (
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeEconomics(w Writer, d model.Report) {
	w.Section("2. ECONOMICS")
	writeEconomicsSummary(w, d, SummaryEmbedded)
	w.Em("Chain-wide tokenomics — block rewards flow through `fee_collector` and `x/distribution` to the community pool and validators.")

	writeEconomicsOverview(w, d)
	w.Hint(economicsSourcesHint())
	w.BlankLine()
}

func economicsSourcesHint() string {
	return "`PMT rewards` → REST GET /cosmos/evm/pmtrewards/v1/params; " +
		"`inflation`, `annual provisions` → REST GET /cosmos/mint/v1beta1/inflation, /cosmos/mint/v1beta1/annual-provisions; " +
		"`blocks / year`, mint params → REST GET /cosmos/mint/v1beta1/params; " +
		"`bonded`, `bond denom`, `unbonding time`, `max validators` → REST GET /cosmos/staking/v1beta1/pool, /cosmos/staking/v1beta1/params; " +
		"`signed blocks window`, `min signed`, `slash fractions` → REST GET /cosmos/slashing/v1beta1/params; " +
		"`community tax`, `community pool` → REST GET /cosmos/distribution/v1beta1/params, /cosmos/distribution/v1beta1/community_pool; " +
		"`unclaimed delegator`, `unclaimed commission` → REST GET /cosmos/distribution/v1beta1/validators/{valoper}/outstanding_rewards, …/commission (summed across validators); " +
		"`module account balances` → REST GET /cosmos/bank/v1beta1/balances/{address}; " +
		"`module account addresses` → REST GET /cosmos/auth/v1beta1/module_accounts; " +
		"`ledger per-block amounts` → derived (PMT rate, mint inflation/block, parent-block fees); " +
		"`fee_collector cleared`, `unclaimed check` → derived (x/bank balances, outstanding sums)."
}

func writeEVMSection(w Writer, d model.Report) {
	w.Section("3. EVM JSON-RPC")
	writeEVMSummary(w, d, SummaryEmbedded)
	w.Em("Wallet and dApp connectivity (`eth_*`, `net_*`, `txpool_*`) on this node's JSON-RPC.")
	writeEVMRPCSection(w, d)
	w.BlankLine()
}
