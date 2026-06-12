package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func writeNodeSummary(w Writer, d model.Report, mode SummaryMode) {
	lv := d.Local
	syncStr := "synced"
	syncKind := "ok"
	if !d.Synced {
		syncStr = "catching up"
		syncKind = "warn"
	}
	proposer := "no"
	if lv.IsNextProposer {
		proposer = "yes — next block"
	}

	summaryWrapStart(w, mode, "node")
	w.WriteHTML(`<div class="node-summary">`)
	w.WriteHTML(`<div class="node-summary__header">`)
	w.WriteHTML(fmt.Sprintf(`<span class="node-summary__moniker">%s</span>`, html.EscapeString(d.Moniker)))
	writeSummaryBadges(w, "node-summary__badges", summaryBadge{syncStr, syncKind})
	w.WriteHTML(`</div>`)
	w.WriteHTML(`<div class="node-summary__grid">`)
	for _, row := range []struct{ label, val string }{
		{"height", d.BlockHeight + " · " + d.TimeSinceBlock},
		{"next proposer", proposer},
		{"peers", fmt.Sprintf("%d cosmos · %d evm", d.PeerCount, d.EVMPeerCount)},
	} {
		if row.val == "" || row.val == " · " {
			continue
		}
		w.WriteHTML(fmt.Sprintf(
			`<div class="node-summary__cell"><span class="node-summary__label">%s</span>`+
				`<span class="node-summary__val">%s</span></div>`,
			html.EscapeString(row.label), html.EscapeString(row.val)))
	}
	w.WriteHTML(`</div>`)
	if len(d.Validators) > 0 {
		writeValidatorP2PSummaryBody(w, d)
	}
	w.WriteHTML(`</div>`)
	summaryWrapEnd(w, mode)
}

func writeNode(w Writer, d model.Report) {
	lv := d.Local
	syncStr := "synced"
	if !d.Synced {
		syncStr = "CATCHING UP"
	}

	w.Section("2. VALIDATOR")
	writeNodeSummary(w, d, SummaryEmbedded)
	w.Em("This node — CometBFT consensus and P2P live state, plus validator-set dial identities. Stake and operator addresses → § Staking. Signing health → § Slashing.")

	writeValidatorIdentityBoard(w, d, lv)

	writeNodeCometBFT(w, d, lv, syncStr)
	writeValidatorP2PNetwork(w, d)
	w.BlankLine()
}

func writeNodeCometBFT(w Writer, d model.Report, lv model.LocalValidator, syncStr string) {
	w.Layer("CometBFT (consensus + networking)")

	w.Subsection("Live state")
	w.Hint("`sync`, `height`, `last block`, `interval` → CometBFT GET /status, GET /block; `mempool` → GET /num_unconfirmed_txs.")
	w.Row("sync", syncStr)
	w.Row("height", d.BlockHeight)
	if d.BlockInterval != "" {
		w.Row("interval", d.BlockInterval)
	}
	if d.TimeSinceBlock != "" {
		w.Row("last block", d.TimeSinceBlock)
	}
	w.Row("mempool", fmt.Sprintf("%d pending", d.MempoolTxs))

	if lv.IsValidator || d.LocalConsensusAddr != "" {
		w.Subsection("Proposer")
		w.Hint("`voting power` → CometBFT GET /status validator_info; `proposer priority` → GET /validators.")
		if d.LocalVotingPower != "" {
			w.Row("voting power", d.LocalVotingPower+"  _(consensus units — `/status` validator_info)_")
		}
		w.Row("proposer priority", report.FormatInt(lv.ProposerPriority))
		if lv.IsNextProposer {
			w.Row("proposer", "**next block proposer**")
		} else if d.NextProposer != "" {
			w.Row("proposer", "not next  _(next: "+d.NextProposer+")_")
		}
	}

	w.Subsection("P2P & RPC")
	w.Hint("`cosmos peers` → CometBFT GET /net_info; `evm peers` → JSON-RPC net_peerCount; `p2p listen`, `p2p dial`, `rpc listen` → CometBFT GET /status (node_info; dial is node_id@listen_addr).")
	w.Row("cosmos peers", fmt.Sprintf("%d  _(CometBFT P2P connections)_", d.PeerCount))
	w.Row("evm peers", fmt.Sprintf("%d  _(JSON-RPC net_peerCount — often 0 on validators)_", d.EVMPeerCount))
	p2pDial := lv.P2PDial
	if p2pDial == "" {
		p2pDial = formatP2PDial(d.NodeID, d.ListenAddr)
	}
	if p2pDial != "" {
		w.Row("p2p dial", p2pDial+"  _(peer dial string)_")
	}
	if d.ListenAddr != "" {
		w.Row("p2p listen", d.ListenAddr+"  _(advertised from `/status`)_")
	}
	if d.RpcListenAddr != "" {
		w.Row("rpc listen", d.RpcListenAddr)
	}
	if d.AppVersion != "" {
		w.Row("version", d.AppVersion)
	}
	if d.Network != "" {
		w.Row("chain ID", d.Network)
	}
}

func formatP2PDial(nodeID, listen string) string {
	if nodeID == "" || listen == "" {
		return ""
	}
	addr := listen
	if len(addr) > 6 && addr[:6] == "tcp://" {
		addr = addr[6:]
	}
	return nodeID + "@" + addr
}
