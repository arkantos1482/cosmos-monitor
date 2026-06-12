package panel

import (
	"fmt"
	"html"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

// EcoLedgerRow is one row in the block reward ledger with optional inactive styling.
type EcoLedgerRow struct {
	Cells    []string
	Inactive bool
	Warn     bool
}

func economicsHasRewardSource(d model.Report) bool {
	if d.PMTEnabled && d.PMTRate != "" && !d.PMTPoolEmpty {
		return true
	}
	if d.Inflation > 0 && d.InflationPerBlock != "" {
		return true
	}
	if v, ok := economicsParseAmount(d.LastBlockFees); ok && v > 0 {
		return true
	}
	return false
}

func ecoPMTEffect(d model.Report) string {
	if !d.PMTEnabled {
		return "module disabled — no PMT block rewards"
	}
	if d.PMTPoolEmpty {
		return "enabled but pool empty — nothing to distribute"
	}
	if d.PMTRate == "" {
		return "enabled — rate unknown"
	}
	return "distributing to fee_collector each block"
}

func ecoPoolEffect(d model.Report) string {
	if !d.PMTEnabled {
		return "—"
	}
	if d.PMTPoolEmpty {
		return "empty — rewards stop"
	}
	if d.PMTRunway != "" {
		return "funded · " + d.PMTRunway
	}
	return "funded"
}

func ecoPMTRateEffect(d model.Report) string {
	if !d.PMTEnabled {
		return "inactive"
	}
	if d.PMTRate == "" {
		return "no rate configured"
	}
	if d.PMTPoolEmpty {
		return "rate set but pool cannot pay"
	}
	return "active per-block emission"
}

func ecoInflationEffect(d model.Report) string {
	if d.Inflation <= 0 {
		return "inactive — x/mint not minting"
	}
	if d.InflationPerBlock != "" {
		return "active — mints each block"
	}
	return "rate set — per-block amount unavailable"
}

func ecoAnnualProvEffect(d model.Report) string {
	if d.Inflation <= 0 {
		return "inactive"
	}
	if d.AnnualProvisions == "" {
		return "—"
	}
	return "absolute mint budget / year"
}

func ecoTaxEffect(d model.Report) string {
	if d.CommunityTaxZero {
		return "0% — community pool gets no cut"
	}
	if !economicsHasRewardSource(d) {
		return "tax configured but no rewards flowing"
	}
	return "skims % of block rewards → community pool"
}

type ecoCardStatus struct {
	cardMod    string
	badgeClass string
	label      string
}

func ecoPMTCardStatus(d model.Report) ecoCardStatus {
	if !d.PMTEnabled {
		return ecoCardStatus{"eco-domain--inactive", "badge--bad", "inactive"}
	}
	if d.PMTPoolEmpty {
		return ecoCardStatus{"eco-domain--ineffective", "badge--warn", "ineffective"}
	}
	return ecoCardStatus{"eco-domain--active", "badge--ok", "active"}
}

func ecoInflationCardStatus(d model.Report) ecoCardStatus {
	if d.Inflation <= 0 {
		return ecoCardStatus{"eco-domain--inactive", "badge--bad", "inactive"}
	}
	return ecoCardStatus{"eco-domain--active", "badge--ok", "active"}
}

func ecoDomainTitleHTML(name, subtitle, badgeClass, statusLabel string) string {
	var b strings.Builder
	b.WriteString(`<h3 class="eco-domain__title">`)
	b.WriteString(html.EscapeString(name))
	fmt.Fprintf(&b, ` <span class="eco-domain__subtitle">%s</span>`, html.EscapeString(subtitle))
	if statusLabel != "" {
		fmt.Fprintf(&b, ` <span class="eco-domain__status badge %s">%s</span>`,
			html.EscapeString(badgeClass), html.EscapeString(statusLabel))
	}
	b.WriteString(`</h3>`)
	return b.String()
}

func ecoBondedVsGoalEffect(d model.Report) string {
	if d.BondedPct > d.GoalBonded {
		return fmt.Sprintf("%.1f%% over goal — inflation decreases", d.BondedPct-d.GoalBonded)
	}
	if d.BondedPct < d.GoalBonded {
		return fmt.Sprintf("%.1f%% under goal — inflation increases", d.GoalBonded-d.BondedPct)
	}
	return "at goal — inflation stable"
}

type economicsDistItem struct {
	param, balance, addr, effect, rowClass string
}

func economicsDistItemsHTML(items []economicsDistItem) string {
	if len(items) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(`<div class="eco-dist"><div class="eco-domain__rows">`)
	for _, item := range items {
		val := ecoBalanceAddrHTML(item.balance, item.addr)
		ecoDomainRowHTML(&b, item.rowClass, item.param, val, item.effect)
	}
	b.WriteString(`</div></div>`)
	return b.String()
}

func economicsDistributionModuleAddr(d model.Report) string {
	return moduleAccountDisplayAddress(d, "distribution")
}

func pmtRewardsCardHTML(d model.Report, compact bool) string {
	return pmtRewardsCardHTMLTitled(d, compact, true)
}

func pmtRewardsCardHTMLTitled(d model.Report, compact, titled bool) string {
	var b strings.Builder
	st := ecoPMTCardStatus(d)
	fmt.Fprintf(&b, `<div class="eco-domain eco-domain--pmtrewards %s">`, st.cardMod)
	if titled {
		b.WriteString(ecoDomainTitleHTML("PMT Rewards", "x/pmtrewards", st.badgeClass, st.label))
	}
	b.WriteString(`<div class="eco-domain__rows">`)
	writePMTRewardsRows(&b, d, compact)
	b.WriteString(`</div></div>`)
	return b.String()
}

func writePMTRewardsRows(b *strings.Builder, d model.Report, compact bool) {

	enabledCls := ""
	if !d.PMTEnabled {
		enabledCls = ` class="eco-domain__row--inactive"`
	}
	ecoDomainRow(b, enabledCls, "enabled", boolStr(d.PMTEnabled), ecoPMTEffect(d))

	rateCls := ""
	if !d.PMTEnabled {
		rateCls = ` class="eco-domain__row--inactive"`
	} else if d.PMTPoolEmpty || d.PMTRate == "" {
		rateCls = ` class="eco-domain__row--warn"`
	}
	ecoDomainRow(b, rateCls, "reward / block", orEcoDash(d.PMTRate), ecoPMTRateEffect(d))

	poolCls := ""
	if !d.PMTEnabled {
		poolCls = ` class="eco-domain__row--inactive"`
	} else if d.PMTPoolEmpty {
		poolCls = ` class="eco-domain__row--warn"`
	}
	poolVal := ecoBalanceAddrHTML(orEcoDash(d.PMTBalance), displayAddress(d.PMTPoolAddress))
	ecoDomainRowHTML(b, poolCls, "reward pool", poolVal, ecoPoolEffect(d))

	if !compact && d.PMTAnnual != "" {
		annualCls := ""
		if !d.PMTEnabled {
			annualCls = ` class="eco-domain__row--inactive"`
		}
		ecoDomainRow(b, annualCls, "annual emissions", d.PMTAnnual, "estimated yearly payout at current rate")
	}
}

func inflationCardHTML(d model.Report, compact bool) string {
	return inflationCardHTMLTitled(d, compact, true)
}

func inflationCardHTMLTitled(d model.Report, compact, titled bool) string {
	var b strings.Builder
	st := ecoInflationCardStatus(d)
	fmt.Fprintf(&b, `<div class="eco-domain eco-domain--inflation %s">`, st.cardMod)
	if titled {
		b.WriteString(ecoDomainTitleHTML("Inflation", "x/mint", st.badgeClass, st.label))
	}
	b.WriteString(`<div class="eco-domain__rows">`)
	writeInflationRows(&b, d, compact)
	b.WriteString(`</div></div>`)
	return b.String()
}

func writeInflationRows(b *strings.Builder, d model.Report, compact bool) {

	inflCls := ""
	if d.Inflation <= 0 {
		inflCls = ` class="eco-domain__row--inactive"`
	}
	inflVal := fmt.Sprintf("%.2f%%", d.Inflation)
	if d.Inflation > 0 && d.InflationPerBlock != "" && !compact {
		inflVal += " (" + d.InflationPerBlock + "/block)"
	}
	ecoDomainRow(b, inflCls, "inflation", inflVal, ecoInflationEffect(d))

	provCls := ""
	if d.Inflation <= 0 || d.AnnualProvisions == "" {
		provCls = ` class="eco-domain__row--inactive"`
	}
	ecoDomainRow(b, provCls, "annual provisions", orEcoDash(d.AnnualProvisions), ecoAnnualProvEffect(d))

	ecoDomainRow(b, "", "bonded vs goal",
		fmt.Sprintf("%.2f%% vs %.0f%%", d.BondedPct, d.GoalBonded),
		ecoBondedVsGoalEffect(d))

	ecoDomainDivider(b)
	ecoDomainRow(b, "", "goal bonded", fmt.Sprintf("%.0f%%", d.GoalBonded), "target stake ratio for inflation")
	if d.BlocksPerYear != "" {
		ecoDomainRow(b, "", "blocks / year", d.BlocksPerYear, "mint schedule denominator")
	}
}

func stakingCardHTML(d model.Report, compact bool) string {
	var b strings.Builder
	b.WriteString(`<div class="eco-domain eco-domain--staking">`)
	b.WriteString(`<h3 class="eco-domain__title">Staking <span class="eco-domain__subtitle">x/staking</span></h3>`)
	b.WriteString(`<div class="eco-domain__rows">`)

	ecoDomainRow(&b, "", "bonded", fmt.Sprintf("%.2f%%", d.BondedPct), "share of supply staked")

	if !compact {
		if d.BondedAmt != "" {
			ecoDomainRow(&b, "", "bonded amt", d.BondedAmt, "actively securing chain")
		}
		if d.NotBonded != "" {
			ecoDomainRow(&b, "", "not bonded", d.NotBonded, "unbonded or unbonding")
		}
		if d.TotalSupply != "" {
			ecoDomainRow(&b, "", "total supply", d.TotalSupply, "bank module supply")
		}
	}

	writeStakingModuleAccountRows(&b, d)
	writeStakingParamRows(&b, d)

	b.WriteString(`</div></div>`)
	return b.String()
}

func slashingCardHTML(d model.Report, _ bool) string {
	var b strings.Builder
	b.WriteString(`<div class="eco-domain eco-domain--slashing">`)
	b.WriteString(`<h3 class="eco-domain__title">Slashing <span class="eco-domain__subtitle">x/slashing</span></h3>`)
	b.WriteString(`<div class="eco-domain__rows">`)

	writeSlashingParamRows(&b, d)
	writeSlashingPenaltyMatrix(&b, d)

	b.WriteString(`</div></div>`)
	return b.String()
}

func writeStakingModuleAccountRows(b *strings.Builder, d model.Report) {
	writeModuleAccountRow(b, d, "bonded_tokens_pool", "staked tokens escrow")
	writeModuleAccountRow(b, d, "not_bonded_tokens_pool", "unbonded / unbonding escrow")
}

func writeStakingParamRows(b *strings.Builder, d model.Report) {
	ecoDomainDivider(b)
	if d.BondDenom != "" {
		ecoDomainRow(b, "", "bond denom", d.BondDenom, "staking unit of account")
	}
	if d.UnbondingTime != "" {
		ecoDomainRow(b, "", "unbonding time", d.UnbondingTime, "time locked after unstaking")
	}
	if d.MaxValidators > 0 {
		ecoDomainRow(b, "", "max validators", fmt.Sprintf("%d", d.MaxValidators), "active set cap")
	}
}

func writeSlashingParamRows(b *strings.Builder, d model.Report) {
	if d.SlashWindow != "" && d.SlashWindow != "0" {
		ecoDomainRow(b, "", "signed blocks window", d.SlashWindow+" blocks", "downtime tracking window")
	}
	if d.MinSigned > 0 {
		ecoDomainRow(b, "", "min signed", fmt.Sprintf("%.1f%%", d.MinSigned), "miss more → downtime slash risk")
	}
	if d.SlashDowntime != "" {
		dtCls := ""
		dtEffect := "fraction slashed for downtime"
		if d.SlashDTInactive {
			dtCls = ` class="eco-domain__row--warn"`
			dtEffect = "inactive — downtime slash disabled"
		}
		ecoDomainRow(b, dtCls, "slash / downtime", d.SlashDowntime, dtEffect)
	}
	if d.SlashDS != "" {
		dsCls := ""
		dsEffect := "fraction slashed for double-sign"
		if d.SlashDSInactive {
			dsCls = ` class="eco-domain__row--warn"`
			dsEffect = "inactive — double-sign slash disabled"
		}
		ecoDomainRow(b, dsCls, "slash / double-sign", d.SlashDS, dsEffect)
	}
}

func economicsLedgerTableHTML(rows []EcoLedgerRow) string {
	headers := []string{"Step", "Where", "In this block", "Balance now", "Check"}
	var b strings.Builder
	b.WriteString(`<div class="table-scroll"><table class="data-table data-table--ledger">`)
	b.WriteString(`<thead><tr>`)
	for _, h := range headers {
		fmt.Fprintf(&b, `<th%s>%s</th>`, tableColumnClass(h), html.EscapeString(h))
	}
	b.WriteString(`</tr></thead><tbody>`)
	for _, row := range rows {
		trCls := ""
		switch {
		case row.Inactive:
			trCls = ` class="eco-row--inactive"`
		case row.Warn:
			trCls = ` class="eco-row--warn"`
		}
		fmt.Fprintf(&b, `<tr%s>`, trCls)
		for i, cell := range row.Cells {
			if i == 0 {
				step := html.EscapeString(strings.TrimSpace(cell))
				fmt.Fprintf(&b, `<td class="data-table__step" data-step="%s">%s</td>`, step, step)
				continue
			}
			tdCls := tableColumnClass(headers[i])
			fmt.Fprintf(&b, `<td%s>%s</td>`, tdCls, formatValue(cell))
		}
		b.WriteString(`</tr>`)
	}
	b.WriteString(`</tbody></table></div>`)
	return b.String()
}
