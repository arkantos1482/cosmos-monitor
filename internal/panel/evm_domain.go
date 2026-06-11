package panel

import (
	"fmt"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func evmDomainCardsHTML(d model.Report) string {
	return ecoDomainsWrap(evmVMCardHTML(d), evmERC20CardHTML(d))
}

func evmVMCardHTML(d model.Report) string {
	var b strings.Builder
	ecoDomainCardOpen(&b, "eco-domain--vm", "EVM", "x/vm")

	syncLabel := "synced"
	if !d.EVMSynced {
		syncLabel = "syncing"
	}
	ecoDomainRow(&b, "", "block height", orEcoDash(d.EVMBlock), "latest eth_blockNumber")
	ecoDomainRow(&b, "", "sync", syncLabel, "eth_syncing status")
	if d.EVMChainID > 0 {
		ecoDomainRow(&b, "", "chain ID", fmt.Sprintf("%d", d.EVMChainID), "MetaMask network ID")
	}
	if d.Local.EVMAddr != "" {
		val := ecoBalanceAddrHTML("", d.Local.EVMAddr)
		ecoDomainRowHTML(&b, "", "validator EVM addr", val, "local validator account on EVM")
	}

	ecoDomainDivider(&b)
	if d.EVMDenom != "" {
		ecoDomainRow(&b, "", "evm_denom", d.EVMDenom, "native EVM coin denom")
	}
	if len(d.Precompiles) > 0 {
		ecoDomainRow(&b, "", "precompiles", strings.Join(d.Precompiles, ", "), "enabled native precompile addresses")
	}
	if d.HistoryWindow != "" {
		ecoDomainRow(&b, "", "history window", d.HistoryWindow, "blocks of state history retained")
	}
	writeHardforkRows(&b, d)

	ecoDomainCardClose(&b)
	return b.String()
}

func writeHardforkRows(b *strings.Builder, d model.Report) {
	for _, hf := range []struct{ name, height string }{
		{"london", d.HardforkLondon},
		{"shanghai", d.HardforkShanghai},
		{"cancun", d.HardforkCancun},
	} {
		if hf.height == "" {
			continue
		}
		label := hf.height
		if hf.height == "0" {
			label = "0 (genesis)"
		}
		ecoDomainRow(b, "", hf.name+"_block", label, "EVM hardfork activation height")
	}
}

func evmERC20CardHTML(d model.Report) string {
	var b strings.Builder
	ecoDomainCardOpen(&b, "eco-domain--erc20", "ERC20", "x/erc20")

	enabled := 0
	for _, tp := range d.TokenPairs {
		if tp.Enabled {
			enabled++
		}
	}
	ecoDomainRow(&b, "", "token pairs", fmt.Sprintf("%d", len(d.TokenPairs)), "see Governance for full list")
	ecoDomainRow(&b, "", "enabled pairs", fmt.Sprintf("%d", enabled), "pairs accepting conversions")

	ecoDomainDivider(&b)
	ecoDomainRow(&b, "", "enable_erc20", boolStr(d.ERC20Enabled), "module-wide ERC20 conversion toggle")

	ecoDomainCardClose(&b)
	return b.String()
}
