package panel

import (
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/feemarket"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func feemarketDomainCardsHTML(d model.Report) string {
	c := feemarket.LoadContext(d)
	return ecoDomainsWrap(feemarketModuleCardHTML(c, d))
}

func feemarketModuleCardHTML(c feemarket.Context, d model.Report) string {
	var b strings.Builder
	ecoDomainCardOpen(&b, "eco-domain--feemarket", "Fee market", "x/feemarket")

	verdict := map[string]string{
		"busy": "BUSY", "quiet": "QUIET", "balanced": "BALANCED", "unknown": "UNKNOWN",
	}[c.Verdict]
	ecoDomainRow(&b, "", "base fee", orEcoDash(d.BaseFee), "minimum gas price this block")
	ecoDomainRow(&b, "", "verdict", verdict, "parent-block demand vs target")
	if c.UtilPct != "" {
		ecoDomainRow(&b, "", "W vs target", c.UtilPct+" of target", "drives next base fee adjustment")
	}
	gasLine := formatGasAmount(c.GasUsed, d.ParentBlockResultsOK) + " used · " +
		formatGasAmount(c.Wanted, d.ParentBlockResultsOK) + " W"
	ecoDomainRow(&b, "", "parent gas", gasLine, "gas_used and W from parent block")
	if c.Badge.Label != "" {
		ecoDomainRow(&b, "", "next adjustment", c.Badge.Label, "base fee direction next block")
	}

	ecoDomainCardClose(&b)
	return b.String()
}

func feeChainParamsDomainHTML(c feemarket.Context, d model.Report) string {
	rows := []struct{ param, value, effect string }{
		{"no_base_fee", boolStr(d.NoBaseFee), "disables dynamic base fee when true"},
		{"enable_height", formatEnableHeight(c.EnableHeight), "block when EIP-1559 fee market turns on"},
		{"base_fee (param store)", orEcoDash(c.BaseFeeParam), "genesis / governance initial base fee"},
		{"base_fee_change_denominator", formatParamUint(c.DenomU), "caps per-block base fee delta"},
		{"elasticity_multiplier", formatParamInt(c.Elasticity), "target gas = max_gas / elasticity"},
		{"min_gas_price", minGasPriceL5(c), "network minimum gas price floor"},
		{"min_gas_multiplier", orEcoDash(c.MinGasMultiplier), "weights in-block gas accumulator for W"},
		{"max_gas", maxGasL5(c), "consensus block gas limit"},
		{"max_bytes", formatParamInt(c.MaxBlockBytes), "consensus max block size"},
		{"evm_denom", orEcoDash(c.Denom), "native token for EVM gas"},
		{"london_block", londonStatus(c), "EIP-1559 activation height"},
		{"min_unit_gas", "1 apmt", "smallest gas price increment"},
	}
	var b strings.Builder
	b.WriteString(`<div class="eco-domain eco-domain--fee-params">`)
	b.WriteString(`<div class="eco-domain__rows">`)
	for _, row := range rows {
		ecoDomainRow(&b, "", row.param, row.value, row.effect)
	}
	b.WriteString(`</div></div>`)
	return b.String()
}
