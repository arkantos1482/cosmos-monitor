package panel

import (
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeEconomics(w Writer, d model.Report) {
	w.Section("2. ECONOMICS")
	writeEconomicsSummary(w, d, SummaryEmbedded)
	w.Em("Chain-wide tokenomics — block rewards flow through `fee_collector` and `x/distribution` to the community pool and validators.")

	writeEconomicsOverview(w, d)
}

func writeEVMSection(w Writer, d model.Report) {
	w.Section("3. EVM JSON-RPC")
	writeEVMSummary(w, d, SummaryEmbedded)
	w.Em("Wallet and dApp connectivity (`eth_*`, `net_*`, `txpool_*`) on this node's JSON-RPC.")
	writeEVMRPCSection(w, d)
	w.BlankLine()
}
