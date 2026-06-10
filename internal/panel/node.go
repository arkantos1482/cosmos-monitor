package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/feemarket"
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
	var badges []summaryBadge
	badges = append(badges, summaryBadge{syncStr, syncKind})
	badges = append(badges, localBadges(d)...)
	writeSummaryBadges(w, "node-summary__badges", badges...)
	w.WriteHTML(`</div>`)
	w.WriteHTML(`<div class="node-summary__grid">`)
	for _, row := range []struct{ label, val string }{
		{"height", d.BlockHeight + " · " + d.TimeSinceBlock},
		{"voting power", nodeVPRow(lv)},
		{"next proposer", proposer},
		{"peers", fmt.Sprintf("%d cosmos · %d evm", d.PeerCount, d.EVMPeerCount)},
		{"signing", lv.SigningStatus},
	} {
		if row.val == "" || row.val == " · " {
			continue
		}
		w.WriteHTML(fmt.Sprintf(
			`<div class="node-summary__cell"><span class="node-summary__label">%s</span>`+
				`<span class="node-summary__val">%s</span></div>`,
			html.EscapeString(row.label), html.EscapeString(row.val)))
	}
	w.WriteHTML(`</div></div>`)
	summaryWrapEnd(w, mode)
}

func nodeVPRow(lv model.LocalValidator) string {
	if !lv.IsValidator {
		return lv.SigningStatus
	}
	return fmt.Sprintf("%.1f%% · %s · %.1f%% commission", lv.VPPercent, lv.Status, lv.Commission)
}

func writeNode(w Writer, d model.Report) {
	lv := d.Local
	syncStr := "synced"
	if !d.Synced {
		syncStr = "CATCHING UP"
	}

	writeNodeSummary(w, d, SummaryEmbedded)
	w.Section("2. VALIDATOR")
	w.Em("This validator on this node — identities, application staking, and CometBFT live state.")

	writeIdentityBoard(w, d, lv)

	if lv.IsValidator {
		writeNodeApplication(w, d, lv)
	} else {
		w.Layer("Application (Cosmos SDK / ABCI state)")
		w.Subsection("Role")
		w.Hint("`role` → CometBFT GET /status; derived when consensus address is absent from x/staking.")
		w.Row("role", lv.SigningStatus)
	}

	writeNodeCometBFT(w, d, lv, syncStr)
	writeNodeFeeAcceptance(w, d)
}

func writeNodeFeeAcceptance(w Writer, d model.Report) {
	c := feemarket.LoadContext(d)
	if c.NodeMinGasPrices == "" && c.NodeEVMMinTip == "" && c.NodeMempoolPriceLimit == "" &&
		c.NodeMaxTxGasWanted == "" && c.NodeAppTomlPath == "" {
		return
	}
	w.Subsection("Fee acceptance (app.toml)")
	w.Hint("`minimum-gas-prices`, `evm.min-tip`, `evm.mempool.price-limit`, `evm.max-tx-gas-wanted` → local app.toml (APPTOML_PATH or ~/.evmd/config/app.toml). Chain fee params live in § Fee market.")
	for _, row := range nodeFeeAcceptanceRows(c) {
		w.Row(row[0], row[1])
	}
}

func nodeFeeAcceptanceRows(c feemarket.Context) [][]string {
	rows := [][]string{
		{"minimum-gas-prices", orDash(c.NodeMinGasPrices)},
		{"evm.min-tip", orDash(c.NodeEVMMinTip)},
		{"evm.mempool.price-limit", orDash(c.NodeMempoolPriceLimit)},
		{"evm.max-tx-gas-wanted", orDash(c.NodeMaxTxGasWanted)},
	}
	if c.NodeAppTomlPath != "" {
		rows = append(rows, []string{"config path", c.NodeAppTomlPath})
	}
	return rows
}

func writeNodeApplication(w Writer, d model.Report, lv model.LocalValidator) {
	w.Layer("Application (Cosmos SDK / ABCI state)")

	w.Subsection("Staking")
	w.Hint("`status`, `jailed`, `voting power`, `commission` → REST GET /cosmos/staking/v1beta1/validators.")
	w.Row("status", lv.Status)
	if lv.Jailed {
		w.Row("jailed", "yes")
	}
	if lv.Tombstoned {
		w.Row("tombstoned", "YES")
	}
	w.Row("voting power", fmt.Sprintf("%s  (%.1f%% of bonded stake)", lv.VotingPower, lv.VPPercent))
	w.Row("commission", fmt.Sprintf("%.1f%%  _(validator cut of delegator rewards)_", lv.Commission))

	w.Subsection("Rewards")
	w.Hint("`outstanding rewards`, `commission earned` → REST GET /cosmos/distribution/v1beta1/validators/{valoper}/outstanding_rewards, …/commission; `per-block` → derived (network reward flow × VP% × commission).")
	if lv.Outstanding != "" {
		w.Row("outstanding rewards", lv.Outstanding+"  _(total unclaimed — x/distribution)_")
	} else {
		w.Row("outstanding rewards", "–")
	}
	if lv.CommissionEarned != "" {
		w.Row("commission earned", lv.CommissionEarned+"  _(unclaimed validator commission)_")
	} else {
		w.Row("commission earned", "–")
	}
	if op, del, _, ok := localValidatorPerBlockRewards(d); ok {
		w.Row("per-block commission", op+fmt.Sprintf("  (%.2f%% VP · %.2f%% commission)", lv.VPPercent, lv.Commission))
		w.Row("per-block delegators", del)
	}

	w.Subsection("Slashing")
	w.Hint("`signing health`, `missed / window` → REST GET /cosmos/slashing/v1beta1/signing_infos + params.")
	w.Row("signing health", lv.SigningStatus)
	if d.SlashWindow != "" && d.SlashWindow != "0" {
		w.Row("missed / window", fmt.Sprintf("%d / %s blocks  (max allowed: %d)", lv.Missed, d.SlashWindow, lv.MaxMissed))
	}
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
	w.Hint("`p2p listen`, `p2p dial`, `rpc listen` → CometBFT GET /status (node_info; dial is node_id@listen_addr).")
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
