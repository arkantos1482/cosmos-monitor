package markdown

import (
	"fmt"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeNode(m *mdWriter, d model.Report) {
	syncStr := "synced"
	if !d.Synced {
		syncStr = "CATCHING UP"
	}

	m.section("2. NODE")

	fmt.Fprintf(m.w, "_This machine's CometBFT process — identity, listen addresses, and consensus view._\n\n")

	m.subsection("Node")
	m.hint("`moniker`, `node ID`, `version`, `chain ID`, `p2p listen`, `rpc listen` → CometBFT RPC `GET /status` (`node_info`, `sync_info`, `validator_info` not used here).")
	m.row("moniker", d.Moniker)
	if d.NodeID != "" {
		m.row("node ID", d.NodeID+"  _(CometBFT P2P peer ID)_")
	}
	if d.AppVersion != "" {
		m.row("version", d.AppVersion)
	}
	if d.Network != "" {
		m.row("chain ID", d.Network)
	}
	if d.ListenAddr != "" {
		m.row("p2p listen", d.ListenAddr+"  _(advertised dial address from `/status`)_")
	}
	if d.RpcListenAddr != "" {
		m.row("rpc listen", d.RpcListenAddr)
	}

	m.subsection("Consensus")
	m.hint("`sync`, `height`, `last block`, `interval` → `/status` + `/block` (and `/block?height=h-1`); `consensus address`, `voting power` → `/status` `validator_info`; `mempool` → `/num_unconfirmed_txs`.")
	m.row("sync", syncStr)
	m.row("height", d.BlockHeight)
	if d.BlockInterval != "" {
		m.row("interval", d.BlockInterval)
	}
	if d.TimeSinceBlock != "" {
		m.row("last block", d.TimeSinceBlock)
	}
	if d.LocalConsensusAddr != "" {
		m.row("consensus address", strings.ToUpper(d.LocalConsensusAddr)+"  _(hex; signs blocks — `/status` validator_info)_")
	}
	if d.LocalVotingPower != "" {
		m.row("voting power", d.LocalVotingPower+"  _(consensus units — `/status` validator_info)_")
	}
	m.row("mempool", fmt.Sprintf("%d pending", d.MempoolTxs))
}
