package markdown

import (
	"fmt"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeEVMSection(m *mdWriter, d model.Report) {
	// ── 7. EVM JSON-RPC ────────────────────────────────────────────────────────
	m.section("7. EVM JSON-RPC")

	fmt.Fprintf(m.w, "_Wallet and dApp connectivity (`eth_*`, `net_*`, `txpool_*`) on this node's JSON-RPC._\n\n")
	writeEVMRPCSection(m.w, d)
	fmt.Fprintln(m.w)
}
