package panel

import (
	"fmt"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeNode(w Writer, d model.Report) {
	syncStr := "synced"
	if !d.Synced {
		syncStr = "CATCHING UP"
	}

	w.Section("2. NODE")
	w.Em("This machine's CometBFT process — identity, listen addresses, and consensus view.")

	w.Subsection("Node")
	w.Hint("`moniker`, `node ID`, `version`, `chain ID`, `p2p listen`, `rpc listen` → CometBFT GET /status (node_info only; sync_info and validator_info in Consensus).")
	w.Row("moniker", d.Moniker)
	if d.NodeID != "" {
		w.Row("node ID", d.NodeID+"  _(CometBFT P2P peer ID)_")
	}
	if d.AppVersion != "" {
		w.Row("version", d.AppVersion)
	}
	if d.Network != "" {
		w.Row("chain ID", d.Network)
	}
	if d.ListenAddr != "" {
		w.Row("p2p listen", d.ListenAddr+"  _(advertised dial address from `/status`)_")
	}
	if d.RpcListenAddr != "" {
		w.Row("rpc listen", d.RpcListenAddr)
	}

	w.Subsection("Consensus")
	w.Hint("`sync`, `height`, `last block`, `interval` → CometBFT GET /status, GET /block (H−1 via ?height=); `consensus address`, `voting power` → CometBFT GET /status validator_info; `mempool` → CometBFT GET /num_unconfirmed_txs.")
	w.Row("sync", syncStr)
	w.Row("height", d.BlockHeight)
	if d.BlockInterval != "" {
		w.Row("interval", d.BlockInterval)
	}
	if d.TimeSinceBlock != "" {
		w.Row("last block", d.TimeSinceBlock)
	}
	if d.LocalConsensusAddr != "" {
		w.Row("consensus address", strings.ToUpper(d.LocalConsensusAddr)+"  _(hex; signs blocks — `/status` validator_info)_")
	}
	if d.LocalVotingPower != "" {
		w.Row("voting power", d.LocalVotingPower+"  _(consensus units — `/status` validator_info)_")
	}
	w.Row("mempool", fmt.Sprintf("%d pending", d.MempoolTxs))
}
