package panel

import (
	"fmt"
	"html"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func writeNodeSummary(w Writer, d model.Report, mode SummaryMode) {
	syncStr := "synced"
	syncKind := "ok"
	if !d.Synced {
		syncStr = "catching up"
		syncKind = "warn"
	}

	summaryWrapStart(w, mode, "node")
	w.WriteHTML(`<div class="node-summary">`)
	w.WriteHTML(`<div class="node-summary__header">`)
	w.WriteHTML(fmt.Sprintf(`<span class="node-summary__moniker">%s</span>`, html.EscapeString(d.Moniker)))
	writeSummaryBadges(w, "node-summary__badges", summaryBadge{syncStr, syncKind})
	w.WriteHTML(`</div>`)
	w.WriteHTML(`<div class="node-summary__grid">`)
	for _, row := range nodeSummaryRows(d) {
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

func nodeSummaryRows(d model.Report) []struct{ label, val string } {
	blockVal := d.BlockHeight
	if d.TimeSinceBlock != "" {
		if blockVal != "" {
			blockVal += " · " + d.TimeSinceBlock
		} else {
			blockVal = d.TimeSinceBlock
		}
	}
	rows := []struct{ label, val string }{
		{"block", blockVal},
		{"mempool", fmt.Sprintf("%d pending", d.MempoolTxs)},
	}
	if d.BlockInterval != "" {
		rows = append(rows, struct{ label, val string }{"interval", d.BlockInterval})
	}
	if d.HasNodeStatus || d.PeerCount > 0 {
		rows = append(rows, struct{ label, val string }{"peers", fmt.Sprintf("%d P2P", d.PeerCount)})
	}
	return rows
}

func writeNode(w Writer, d model.Report) {
	lv := d.Local
	syncStr := "synced"
	if !d.Synced {
		syncStr = "CATCHING UP"
	}

	w.Section("2. VALIDATOR")
	writeEmbeddedSectionIntro(w, "Live CometBFT sync, block timing, mempool, and P2P peers for this node, plus the full validator-set dial table.")
	writeNodeSummary(w, d, SummaryEmbedded)

	writeNodeCometBFT(w, d, lv, syncStr)
	writeValidatorP2PNetwork(w, d)
	writeSectionSources(w, ViewNode, d)
	w.BlankLine()
}

func writeNodeCometBFT(w Writer, d model.Report, lv model.LocalValidator, syncStr string) {
	w.Layer("CometBFT (consensus + networking)")

	w.Subsection("Live state")
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
		if d.LocalVotingPower != "" {
			w.Row("voting power", d.LocalVotingPower+"  _(consensus units — `/status` validator_info)_")
		}
		w.Row("proposer priority", report.FormatInt(lv.ProposerPriority))
		if lv.IsNextProposer {
			w.Row("proposer", "**next block proposer**")
		} else if d.NextProposer != "" {
			w.Row("proposer", "not next  _(next: "+d.NextProposer+")_")
		}
		if cons := localConsensusBech32(lv); cons != "" {
			w.Row("consensus", cons+"  _(consensus pubkey — bech32)_")
		}
	}

	w.Subsection("P2P & RPC")
	w.Row("cosmos peers", fmt.Sprintf("%d  _(CometBFT P2P connections)_", d.PeerCount))
	if d.NodeID != "" {
		w.Row("node ID", strings.ToLower(d.NodeID)+"  _(CometBFT P2P peer ID — hex)_")
	}
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
