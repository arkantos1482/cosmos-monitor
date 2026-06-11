package panel

import (
	"fmt"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func governanceDomainCardsHTML(d model.Report) string {
	return ecoDomainsWrap(
		govModuleCardHTML(d),
		upgradeModuleCardHTML(d),
		ibcModuleCardHTML(d),
		erc20ModuleCardHTML(d),
	)
}

func govModuleCardHTML(d model.Report) string {
	var b strings.Builder
	ecoDomainCardOpen(&b, "eco-domain--gov", "Governance", "x/gov")

	ecoDomainRow(&b, "", "voting proposals", fmt.Sprintf("%d", len(d.Proposals)), "active on-chain votes")
	ecoDomainRow(&b, "", "deposit period", fmt.Sprintf("%d", len(d.DepositProposals)), "awaiting min deposit")

	writeModuleAccountRow(&b, d, "gov", "proposal deposit escrow")

	ecoDomainDivider(&b)
	if d.VotingPeriod != "" {
		ecoDomainRow(&b, "", "voting period", d.VotingPeriod, "max time in voting stage")
	}
	if d.Quorum > 0 {
		ecoDomainRow(&b, "", "quorum", fmt.Sprintf("%.1f%%", d.Quorum), "min turnout for proposal to pass")
	}
	if d.Threshold > 0 {
		ecoDomainRow(&b, "", "threshold", fmt.Sprintf("%.1f%%", d.Threshold), "min yes share of participating votes")
	}
	if d.VetoThreshold > 0 {
		ecoDomainRow(&b, "", "veto threshold", fmt.Sprintf("%.1f%%", d.VetoThreshold), "no_with_veto share to reject")
	}

	ecoDomainCardClose(&b)
	return b.String()
}

func upgradeModuleCardHTML(d model.Report) string {
	var b strings.Builder
	ecoDomainCardOpen(&b, "eco-domain--upgrade", "Upgrade", "x/upgrade")

	if d.UpgradeName == "" || d.UpgradeName == "none" {
		ecoDomainRow(&b, "", "pending plan", "none", "no scheduled chain upgrade")
	} else {
		ecoDomainRow(&b, "", "plan", d.UpgradeName, "scheduled upgrade name")
		ecoDomainRow(&b, "", "target height", d.UpgradeHeight, "block height for upgrade")
		if d.BlocksLeft != "" {
			ecoDomainRow(&b, "", "blocks left", d.BlocksLeft, "remaining blocks until upgrade")
		}
	}

	ecoDomainCardClose(&b)
	return b.String()
}

func ibcModuleCardHTML(d model.Report) string {
	var b strings.Builder
	ecoDomainCardOpen(&b, "eco-domain--ibc", "IBC", "x/ibc")
	ecoDomainRow(&b, "", "active clients", fmt.Sprintf("%d", d.IBCClients), "light clients tracked on chain")
	ecoDomainCardClose(&b)
	return b.String()
}

func erc20ModuleCardHTML(d model.Report) string {
	var b strings.Builder
	ecoDomainCardOpen(&b, "eco-domain--erc20", "ERC20", "x/erc20")

	enabled := 0
	for _, tp := range d.TokenPairs {
		if tp.Enabled {
			enabled++
		}
	}
	ecoDomainRow(&b, "", "token pairs", fmt.Sprintf("%d", len(d.TokenPairs)), "registered Cosmos↔EVM mappings")
	ecoDomainRow(&b, "", "enabled pairs", fmt.Sprintf("%d", enabled), "pairs accepting conversions")

	ecoDomainDivider(&b)
	ecoDomainRow(&b, "", "enable_erc20", boolStr(d.ERC20Enabled), "module-wide ERC20 conversion toggle")

	ecoDomainCardClose(&b)
	return b.String()
}

func governanceTokenPairsHTML(pairs []model.TokenPair) string {
	if len(pairs) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(`<div class="eco-dist"><div class="eco-domain__rows">`)
	for _, tp := range pairs {
		effect := "conversion disabled"
		rowCls := ` class="eco-domain__row--inactive"`
		if tp.Enabled {
			effect = "Cosmos denom ↔ ERC20 contract"
			rowCls = ""
		}
		val := ecoBalanceAddrHTML(tp.Denom, tp.ERC20)
		ecoDomainRowHTML(&b, rowCls, tp.Denom, val, effect)
	}
	b.WriteString(`</div></div>`)
	return b.String()
}
