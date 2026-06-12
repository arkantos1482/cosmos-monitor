package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/feemarket"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeFeemarketSummary(w Writer, d model.Report, mode SummaryMode) {
	c := feemarket.LoadContext(d)
	transfer := feemarket.TransferCost(c.BaseFeeRaw, c.Denom)
	baseFee := d.BaseFee
	if baseFee == "" {
		baseFee = "—"
	}
	verdict := map[string]string{
		"busy": "BUSY", "quiet": "QUIET", "balanced": "BALANCED", "unknown": "UNKNOWN",
	}[c.Verdict]
	gasLine := formatGasAmount(c.GasUsed, d.ParentBlockResultsOK) + " used · " +
		formatGasAmount(c.Wanted, d.ParentBlockResultsOK) + " W"

	summaryWrapStart(w, mode, "feemarket")
	w.WriteHTML(`<div class="fee-summary">`)
	w.WriteHTML(`<div class="fee-summary__top">`)
	badgeCls := c.Badge.Class
	if badgeCls == "" {
		badgeCls = "stable"
	}
	w.WriteHTML(fmt.Sprintf(`<span class="fee-badge fee-badge--%s">%s</span>`,
		html.EscapeString(badgeCls), html.EscapeString(c.Badge.Label)))
	w.WriteHTML(fmt.Sprintf(`<span class="fee-summary__base">%s</span>`, html.EscapeString(baseFee)))
	w.WriteHTML(`</div>`)
	w.WriteHTML(`<div class="fee-summary__meta">`)
	w.WriteHTML(fmt.Sprintf(`<span class="fee-summary__pill">%s</span>`, html.EscapeString(verdict)))
	if c.UtilPct != "" {
		w.WriteHTML(fmt.Sprintf(`<span class="fee-summary__pill">%s of target</span>`, html.EscapeString(c.UtilPct)))
	}
	w.WriteHTML(fmt.Sprintf(`<span class="fee-summary__pill">transfer %s</span>`, html.EscapeString(transfer)))
	w.WriteHTML(`</div>`)
	w.WriteHTML(fmt.Sprintf(`<p class="fee-summary__gas">%s</p>`, html.EscapeString(gasLine)))
	w.WriteHTML(`</div>`)
	summaryWrapEnd(w, mode)
}

func writeFeemarket(w Writer, d model.Report) {
	c := feemarket.LoadContext(d)
	w.Section("5. FEE MARKET")
	writeFeemarketSummary(w, d, SummaryEmbedded)
	w.Em("Chain-wide EIP-1559 fee market — live base fee, demand, and governance parameters.")
	writeFeemarketFeeAcceptance(w, d)
	writeFeemarketPage(w, d)
	w.Hint(feemarketSourcesHint(c))
	w.BlankLine()
}

func writeFeemarketFeeAcceptance(w Writer, d model.Report) {
	c := feemarket.LoadContext(d)
	if c.NodeMinGasPrices == "" && c.NodeEVMMinTip == "" && c.NodeMempoolPriceLimit == "" &&
		c.NodeMaxTxGasWanted == "" && c.NodeAppTomlPath == "" {
		return
	}
	w.Subsection("Local policy (app.toml)")
	w.Hint("`minimum-gas-prices`, `evm.min-tip`, `evm.mempool.price-limit`, `evm.max-tx-gas-wanted` → local app.toml (APPTOML_PATH or ~/.evmd/config/app.toml). This node's local policy — network-wide fee params are below.")
	for _, row := range feeAcceptanceRows(c) {
		w.Row(row[0], row[1])
	}
}

func feeAcceptanceRows(c feemarket.Context) [][]string {
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

func feemarketSourcesHint(c feemarket.Context) string {
	appToml := "local app.toml (APPTOML_PATH or ~/.evmd/config/app.toml)"
	return fmt.Sprintf(
		"`head height` → CometBFT GET /status; "+
			"`max_gas`, `max_bytes` → CometBFT GET /consensus_params; "+
			"`gas_used`, `W` → CometBFT GET /block_results?height=%s; "+
			"`base_fee` (BeginBlock) → CometBFT GET /block_results?height=%s; "+
			"`block interval` → CometBFT GET /block; "+
			"`base_fee` → REST GET /cosmos/evm/feemarket/v1/base_fee; "+
			"`W` (fallback) → REST GET /cosmos/evm/feemarket/v1/block_gas; "+
			"`no_base_fee`, `elasticity`, `min_gas_*`, … → REST GET /cosmos/evm/feemarket/v1/params; "+
			"`evm_denom` → REST GET /cosmos/evm/vm/v1/params; "+
			"`london_block` → REST GET /cosmos/evm/vm/v1/config; "+
			"local policy (app.toml) → %s (subsection above).",
		c.ParentBlock, c.CurrentBlock, appToml,
	)
}
