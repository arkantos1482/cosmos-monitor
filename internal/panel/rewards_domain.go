package panel

import (
	"fmt"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

type rewardsCardStatus struct {
	modClass   string
	badgeClass string
	label      string
}

func pmtRewardsCardStatus(d model.Report) rewardsCardStatus {
	if !d.PMTEnabled {
		return rewardsCardStatus{"eco-domain--inactive", "badge--bad", "disabled"}
	}
	if d.PMTPoolEmpty {
		return rewardsCardStatus{"eco-domain--ineffective", "badge--warn", "pool empty"}
	}
	return rewardsCardStatus{"eco-domain--active", "badge--ok", "emitting"}
}

func mintCardStatus(d model.Report) rewardsCardStatus {
	if d.Inflation <= 0 {
		return rewardsCardStatus{"eco-domain--inactive", "badge--bad", "off"}
	}
	return rewardsCardStatus{"eco-domain--active", "badge--ok", "minting"}
}

func rewardsEmissionPerBlock(d model.Report) string {
	total, unit, parts := rewardsEmissionAmounts(d)
	if parts == 0 || total <= 0 {
		return "—"
	}
	return fetch.FormatAmountUnit(total, unit) + "/block"
}

func rewardsEmissionAmounts(d model.Report) (total float64, unit string, parts int) {
	unit = economicsDenom(d)
	if v, u, ok := economicsPMTPerBlock(d); ok {
		total += v
		unit = u
		parts++
	}
	if d.InflationPerBlock != "" {
		if v, ok := economicsParseAmount(d.InflationPerBlock); ok && v > 0 {
			total += v
			parts++
		}
	}
	return total, unit, parts
}

func rewardsSummaryPMT(d model.Report) (label, value, tone string) {
	if !d.PMTEnabled {
		return "PMT emission", "disabled", ""
	}
	if d.PMTPoolEmpty {
		return "PMT emission", "pool empty", "warn"
	}
	if d.PMTRate != "" {
		return "PMT emission", d.PMTRate, "ok"
	}
	return "PMT emission", "enabled", "ok"
}

func rewardsSummaryInflation(d model.Report) (label, value, tone string) {
	if d.Inflation <= 0 {
		return "inflation", "off", ""
	}
	val := fmt.Sprintf("%.2f%%", d.Inflation)
	if d.InflationPerBlock != "" {
		val += " · " + d.InflationPerBlock
	}
	return "inflation", val, "ok"
}

func rewardsDomainCardsHTML(d model.Report) string {
	return ecoDomainsWrap(pmtRewardsDomainCard(d), mintInflationDomainCard(d))
}

func pmtRewardsDomainCard(d model.Report) string {
	st := pmtRewardsCardStatus(d)
	var b strings.Builder
	fmt.Fprintf(&b, `<div class="eco-domain eco-domain--pmtrewards %s">`, st.modClass)
	ecoDomainCardTitle(&b, "PMT Rewards", "x/pmtrewards", st.badgeClass, st.label)
	b.WriteString(`<div class="eco-domain__rows">`)

	enabledCls := ""
	if !d.PMTEnabled {
		enabledCls = ` class="eco-domain__row--inactive"`
	}
	ecoDomainRow(&b, enabledCls, "enabled", boolStr(d.PMTEnabled),
		"governance toggle — when false, no pool transfers run")

	rateCls := ""
	rateEffect := "fixed per-block payout from params"
	if !d.PMTEnabled {
		rateCls = ` class="eco-domain__row--inactive"`
		rateEffect = "inactive"
	} else if d.PMTPoolEmpty || d.PMTRate == "" {
		rateCls = ` class="eco-domain__row--warn"`
		if d.PMTPoolEmpty {
			rateEffect = "configured but pool cannot fund transfers"
		}
	}
	ecoDomainRow(&b, rateCls, "reward_per_block", orEcoDash(d.PMTRate), rateEffect)

	poolCls := ""
	poolEffect := "bank account debited each block (up to reward_per_block)"
	if !d.PMTEnabled {
		poolCls = ` class="eco-domain__row--inactive"`
		poolEffect = "—"
	} else if d.PMTPoolEmpty {
		poolCls = ` class="eco-domain__row--warn"`
		poolEffect = "empty — BeginBlock transfer skipped"
	} else if d.PMTRunway != "" {
		poolEffect = "funded · " + d.PMTRunway + " at current rate"
	}
	poolVal := ecoBalanceAddrHTML(orEcoDash(d.PMTBalance), displayAddress(d.PMTPoolAddress))
	ecoDomainRowHTML(&b, poolCls, "pool_address", poolVal, poolEffect)

	if d.PMTAnnual != "" {
		annualCls := ""
		if !d.PMTEnabled {
			annualCls = ` class="eco-domain__row--inactive"`
		}
		ecoDomainRow(&b, annualCls, "annual (est.)", d.PMTAnnual,
			"reward_per_block × blocks_per_year")
	}
	if d.PMTDailyEmit != "" && d.PMTEnabled {
		ecoDomainRow(&b, "", "daily (est.)", d.PMTDailyEmit,
			"reward_per_block × blocks/day from observed block time")
	}

	ecoDomainDividerDist(&b, "BeginBlock path")
	ecoDomainRow(&b, "", "hook", "x/mint MintFn",
		"evmd registers NewPMTRewardMintFn — not a separate BeginBlock in x/pmtrewards")
	ecoDomainRow(&b, "", "transfer", "pool → fee_collector",
		"SendCoinsFromAccountToModule; partial payout if pool balance < reward_per_block")

	ecoDomainCardClose(&b)
	return b.String()
}

func mintInflationDomainCard(d model.Report) string {
	st := mintCardStatus(d)
	var b strings.Builder
	fmt.Fprintf(&b, `<div class="eco-domain eco-domain--inflation %s">`, st.modClass)
	ecoDomainCardTitle(&b, "Inflation", "x/mint", st.badgeClass, st.label)
	b.WriteString(`<div class="eco-domain__rows">`)

	inflCls := ""
	inflEffect := "current annualized mint rate (adjusted each block from bonded ratio)"
	if d.Inflation <= 0 {
		inflCls = ` class="eco-domain__row--inactive"`
		inflEffect = "mint module not issuing new supply"
	}
	ecoDomainRow(&b, inflCls, "inflation", fmt.Sprintf("%.2f%%", d.Inflation), inflEffect)

	provCls := ""
	provEffect := "tokens minted to fee_collector over one year at current rate"
	if d.Inflation <= 0 || d.AnnualProvisions == "" {
		provCls = ` class="eco-domain__row--inactive"`
		if d.Inflation <= 0 {
			provEffect = "inactive"
		}
	}
	ecoDomainRow(&b, provCls, "annual_provisions", orEcoDash(d.AnnualProvisions), provEffect)

	perBlockCls := ""
	perBlockEffect := "annual_provisions ÷ blocks_per_year"
	if d.Inflation <= 0 || d.InflationPerBlock == "" {
		perBlockCls = ` class="eco-domain__row--inactive"`
		if d.Inflation <= 0 {
			perBlockEffect = "inactive"
		}
	}
	ecoDomainRow(&b, perBlockCls, "per block", orEcoDash(d.InflationPerBlock), perBlockEffect)

	if d.InflationPerDay != "" && d.Inflation > 0 {
		ecoDomainRow(&b, "", "per day (est.)", d.InflationPerDay,
			"per-block mint × blocks/day from observed block time")
	}

	ecoDomainDividerDist(&b, "Inflation params")
	ecoDomainRow(&b, "", "goal_bonded", fmt.Sprintf("%.0f%%", d.GoalBonded),
		"target bonded ratio — mint raises/lowers inflation when stake drifts")
	ecoDomainRow(&b, "", "bonded now", fmt.Sprintf("%.2f%%", d.BondedPct),
		mintBondedVsGoalEffect(d))
	if d.BlocksPerYear != "" {
		ecoDomainRow(&b, "", "blocks_per_year", d.BlocksPerYear,
			"denominator for per-block inflation and PMT annual estimates")
	}

	ecoDomainDividerDist(&b, "BeginBlock path")
	ecoDomainRow(&b, "", "mint", "new supply → fee_collector",
		"standard SDK inflation mint before distribution BeginBlock")
	writeModuleAccountRow(&b, d, "fee_collector",
		"receives PMT pool transfers + inflation mint; cleared by x/distribution")

	ecoDomainCardClose(&b)
	return b.String()
}

func mintBondedVsGoalEffect(d model.Report) string {
	if d.GoalBonded <= 0 {
		return "bonded stake share of total supply"
	}
	if d.BondedPct > d.GoalBonded {
		return fmt.Sprintf("%.1f%% above goal — inflation trends down", d.BondedPct-d.GoalBonded)
	}
	if d.BondedPct < d.GoalBonded {
		return fmt.Sprintf("%.1f%% below goal — inflation trends up", d.GoalBonded-d.BondedPct)
	}
	return "at goal — inflation stable"
}

func rewardsEmissionTableHTML(d model.Report) string {
	headers := []string{"source", "per block", "destination", "status"}
	rows := [][]string{
		rewardsEmissionRowPMT(d),
		rewardsEmissionRowMint(d),
	}
	var b strings.Builder
	b.WriteString(`<div class="table-scroll"><table class="data-table data-table--emission"><thead><tr>`)
	for _, h := range headers {
		fmt.Fprintf(&b, `<th%s>%s</th>`, tableColumnClass(h), h)
	}
	b.WriteString(`</tr></thead><tbody>`)
	for _, row := range rows {
		trCls := ""
		if row[3] == "inactive" || row[3] == "disabled" {
			trCls = ` class="eco-row--inactive"`
		} else if row[3] == "pool empty" {
			trCls = ` class="eco-row--warn"`
		}
		fmt.Fprintf(&b, `<tr%s>`, trCls)
		for i, cell := range row {
			if i == 3 {
				fmt.Fprintf(&b, `<td class="data-table__center">%s</td>`, formatValue(cell))
				continue
			}
			fmt.Fprintf(&b, `<td%s>%s</td>`, tableColumnClass(headers[i]), formatValue(cell))
		}
		b.WriteString(`</tr>`)
	}
	b.WriteString(`</tbody></table></div>`)
	return b.String()
}

func rewardsEmissionRowPMT(d model.Report) []string {
	rate := orEcoDash(d.PMTRate)
	status := "inactive"
	if !d.PMTEnabled {
		status = "disabled"
	} else if d.PMTPoolEmpty {
		status = "pool empty"
	} else if d.PMTRate != "" {
		status = "active"
	}
	return []string{"x/pmtrewards (pool)", rate, "fee_collector", status}
}

func rewardsEmissionRowMint(d model.Report) []string {
	perBlock := orEcoDash(d.InflationPerBlock)
	status := "inactive"
	if d.Inflation > 0 && d.InflationPerBlock != "" {
		status = "active"
	} else if d.Inflation > 0 {
		status = "rate only"
	}
	return []string{"x/mint (inflation)", perBlock, "fee_collector", status}
}
