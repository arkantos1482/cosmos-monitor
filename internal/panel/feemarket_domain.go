package panel

import (
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/feemarket"
	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func feemarketDomainCardsHTML(s feemarket.State, d model.Report) string {
	return ecoDomainsWrap(
		fmModuleCardHTML(s, d),
		fmConsensusCardHTML(s, d),
		fmNodePolicyCardHTML(s),
	)
}

func fmModuleCardHTML(s feemarket.State, d model.Report) string {
	var b strings.Builder
	ecoDomainCardOpen(&b, "eco-domain--feemarket", "x/feemarket params", "on-chain module store")
	ecoDomainRow(&b, "", "no_base_fee", boolStr(d.NoBaseFee), "disables dynamic base fee when true")
	ecoDomainRow(&b, "", "enable_height", formatParamInt(s.EnableHeight), "block height EIP-1559 activates")
	ecoDomainRow(&b, "", "base_fee", orEcoDash(s.BaseFeeParam), "stored genesis / governance floor")
	ecoDomainRow(&b, "", "base_fee_change_denominator", formatParamInt(s.ChangeDenom), "max per-block delta as 1/N of base fee")
	ecoDomainRow(&b, "", "elasticity_multiplier", formatParamInt(s.Elasticity), "gas target = block limit ÷ this")
	ecoDomainRow(&b, "", "min_gas_price", orEcoDash(s.MinGasPrice), "floor when base fee decreases")
	ecoDomainRow(&b, "", "min_gas_multiplier", orEcoDash(s.MinGasMult), "EndBlock: W ≥ gas_used and W ≥ wanted×multiplier")
	ecoDomainCardClose(&b)
	return b.String()
}

func fmConsensusCardHTML(s feemarket.State, d model.Report) string {
	var b strings.Builder
	ecoDomainCardOpen(&b, "eco-domain--feemarket-consensus", "Consensus limits", "CometBFT block caps")
	limit := "unlimited"
	if s.GasLimit > 0 {
		limit = formatUint(s.GasLimit) + " gas"
	}
	ecoDomainRow(&b, "", "max_gas", limit, "block gas meter ceiling")
	target := "—"
	if s.GasTarget > 0 {
		target = formatUint(s.GasTarget) + " gas"
	}
	ecoDomainRow(&b, "", "gas target", target, "limit ÷ elasticity_multiplier")
	if d.MaxBlockBytes > 0 {
		ecoDomainRow(&b, "", "max_block_bytes", formatParamInt(d.MaxBlockBytes), "serialized block size cap")
	}
	ecoDomainCardClose(&b)
	return b.String()
}

func fmNodePolicyCardHTML(s feemarket.State) string {
	var b strings.Builder
	ecoDomainCardOpen(&b, "eco-domain--feemarket-node", "Node policy", "local app.toml / mempool")
	ecoDomainRow(&b, policyRowClass(s.NodeMinGas), "minimum-gas-prices", orEcoDash(s.NodeMinGas), "CometBFT CheckTx minimum")
	ecoDomainRow(&b, policyRowClass(s.NodeEVMTip), "evm.min-tip", orEcoDash(s.NodeEVMTip), "EVM ante min effective gas price")
	ecoDomainRow(&b, policyRowClass(s.NodePriceLimit), "mempool.price-bump", orEcoDash(s.NodePriceLimit), "replacement tx price bump %")
	ecoDomainRow(&b, policyRowClass(s.NodeMaxGasWant), "max-tx-gas-wanted", orEcoDash(s.NodeMaxGasWant), "per-tx gas wanted cap in mempool")
	ecoDomainCardClose(&b)
	return b.String()
}

func policyRowClass(v string) string {
	if strings.TrimSpace(v) == "" || v == "—" {
		return "eco-domain__row--inactive"
	}
	return ""
}

func fmProjectedFee(s feemarket.State) string {
	if !s.HasProjection {
		return "—"
	}
	return fetch.FormatFeeAmount(s.ProjectedRaw, s.Denom)
}
