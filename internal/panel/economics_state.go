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

func ecoTxFeesEffect(d model.Report) string {
	if d.LastBlockFees == "" {
		return "no parent-block fee data"
	}
	v, ok := economicsParseAmount(d.LastBlockFees)
	if !ok || v <= 0 {
		return "no tx fee income this block"
	}
	return "gas × base fee → fee_collector"
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

func trimFeeNote(s string) string {
	if i := strings.Index(s, "  _"); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return s
}

func orEcoDash(s string) string {
	if s == "" || s == "—" {
		return "—"
	}
	return s
}

func ecoDomainValueHTML(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "true":
		return `<span class="badge badge--ok">true</span>`
	case "false":
		return `<span class="badge badge--bad">false</span>`
	default:
		return softWrapHTML(v)
	}
}

func ecoDomainRow(b *strings.Builder, rowClass, param, value, effect string) {
	fmt.Fprintf(b, `<div%s><div class="eco-domain__param">%s</div><div class="eco-domain__value">%s</div><div class="eco-domain__effect">%s</div></div>`,
		rowClass, html.EscapeString(param), ecoDomainValueHTML(value), html.EscapeString(effect))
}

func ecoDomainDivider(b *strings.Builder) {
	b.WriteString(`<div class="eco-domain__divider">Governance params</div>`)
}

func ecoDomainHint(b *strings.Builder, hint string) {
	fmt.Fprintf(b, `<div class="eco-domain__hint"><code>%s</code></div>`, html.EscapeString(hint))
}

func economicsDomainCardsHTML(d model.Report, compact bool) string {
	var b strings.Builder
	b.WriteString(`<div class="eco-domains">`)
	b.WriteString(pmtRewardsCardHTML(d, compact))
	b.WriteString(inflationCardHTML(d, compact))
	b.WriteString(stakingCardHTML(d, compact))
	b.WriteString(txFeesCardHTML(d, compact))
	b.WriteString(`</div>`)
	return b.String()
}

func pmtRewardsCardHTML(d model.Report, compact bool) string {
	var b strings.Builder
	b.WriteString(`<div class="eco-domain eco-domain--pmtrewards">`)
	b.WriteString(`<h3 class="eco-domain__title">PMT Rewards <span class="eco-domain__subtitle">x/pmtrewards</span></h3>`)
	b.WriteString(`<div class="eco-domain__rows">`)

	enabledCls := ""
	if !d.PMTEnabled {
		enabledCls = ` class="eco-domain__row--inactive"`
	}
	ecoDomainRow(&b, enabledCls, "enabled", boolStr(d.PMTEnabled), ecoPMTEffect(d))

	rateCls := ""
	if !d.PMTEnabled {
		rateCls = ` class="eco-domain__row--inactive"`
	} else if d.PMTPoolEmpty || d.PMTRate == "" {
		rateCls = ` class="eco-domain__row--warn"`
	}
	ecoDomainRow(&b, rateCls, "reward / block", orEcoDash(d.PMTRate), ecoPMTRateEffect(d))

	poolCls := ""
	if !d.PMTEnabled {
		poolCls = ` class="eco-domain__row--inactive"`
	} else if d.PMTPoolEmpty {
		poolCls = ` class="eco-domain__row--warn"`
	}
	ecoDomainRow(&b, poolCls, "pool balance", orEcoDash(d.PMTBalance), ecoPoolEffect(d))

	if !compact && d.PMTAnnual != "" {
		annualCls := ""
		if !d.PMTEnabled {
			annualCls = ` class="eco-domain__row--inactive"`
		}
		ecoDomainRow(&b, annualCls, "annual emissions", d.PMTAnnual, "estimated yearly payout at current rate")
	}

	ecoDomainDivider(&b)
	if d.PMTPoolAddress != "" {
		ecoDomainRow(&b, "", "pool address", d.PMTPoolAddress, "funding source for block rewards")
	}
	ecoDomainHint(&b, "GET /cosmos/evm/pmtrewards/v1/params")

	b.WriteString(`</div></div>`)
	return b.String()
}

func inflationCardHTML(d model.Report, compact bool) string {
	var b strings.Builder
	b.WriteString(`<div class="eco-domain eco-domain--inflation">`)
	b.WriteString(`<h3 class="eco-domain__title">Inflation <span class="eco-domain__subtitle">x/mint</span></h3>`)
	b.WriteString(`<div class="eco-domain__rows">`)

	inflCls := ""
	if d.Inflation <= 0 {
		inflCls = ` class="eco-domain__row--inactive"`
	}
	inflVal := fmt.Sprintf("%.2f%%", d.Inflation)
	if d.Inflation > 0 && d.InflationPerBlock != "" && !compact {
		inflVal += " (" + d.InflationPerBlock + "/block)"
	}
	ecoDomainRow(&b, inflCls, "inflation", inflVal, ecoInflationEffect(d))

	provCls := ""
	if d.Inflation <= 0 || d.AnnualProvisions == "" {
		provCls = ` class="eco-domain__row--inactive"`
	}
	ecoDomainRow(&b, provCls, "annual provisions", orEcoDash(d.AnnualProvisions), ecoAnnualProvEffect(d))

	ecoDomainRow(&b, "", "bonded vs goal",
		fmt.Sprintf("%.2f%% vs %.0f%%", d.BondedPct, d.GoalBonded),
		ecoBondedVsGoalEffect(d))

	ecoDomainDivider(&b)
	ecoDomainRow(&b, "", "goal bonded", fmt.Sprintf("%.0f%%", d.GoalBonded), "target stake ratio for inflation")
	if d.BlocksPerYear != "" {
		ecoDomainRow(&b, "", "blocks / year", d.BlocksPerYear, "mint schedule denominator")
	}
	ecoDomainHint(&b, "GET /cosmos/mint/v1beta1/params")

	b.WriteString(`</div></div>`)
	return b.String()
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

	ecoDomainDivider(&b)
	if d.BondDenom != "" {
		ecoDomainRow(&b, "", "bond denom", d.BondDenom, "staking unit of account")
	}
	if d.UnbondingTime != "" {
		ecoDomainRow(&b, "", "unbonding time", d.UnbondingTime, "time locked after unstaking")
	}
	if d.MaxValidators > 0 {
		ecoDomainRow(&b, "", "max validators", fmt.Sprintf("%d", d.MaxValidators), "active set cap")
	}
	if d.SlashWindow != "" && d.SlashWindow != "0" {
		ecoDomainRow(&b, "", "signed blocks window", d.SlashWindow+" blocks", "downtime tracking window")
	}
	if d.MinSigned > 0 {
		ecoDomainRow(&b, "", "min signed", fmt.Sprintf("%.1f%%", d.MinSigned), "miss more → downtime slash risk")
	}
	if d.SlashDowntime != "" {
		dtCls := ""
		dtEffect := "fraction slashed for downtime"
		if d.SlashDTInactive {
			dtCls = ` class="eco-domain__row--warn"`
			dtEffect = "inactive — downtime slash disabled"
		}
		ecoDomainRow(&b, dtCls, "slash / downtime", d.SlashDowntime, dtEffect)
	}
	if d.SlashDS != "" {
		dsCls := ""
		dsEffect := "fraction slashed for double-sign"
		if d.SlashDSInactive {
			dsCls = ` class="eco-domain__row--warn"`
			dsEffect = "inactive — double-sign slash disabled"
		}
		ecoDomainRow(&b, dsCls, "slash / double-sign", d.SlashDS, dsEffect)
	}
	ecoDomainHint(&b, "GET /cosmos/staking/v1beta1/params · GET /cosmos/slashing/v1beta1/params")

	b.WriteString(`</div></div>`)
	return b.String()
}

func txFeesCardHTML(d model.Report, compact bool) string {
	var b strings.Builder
	b.WriteString(`<div class="eco-domain eco-domain--txfees">`)
	b.WriteString(`<h3 class="eco-domain__title">Tx Fees</h3>`)
	b.WriteString(`<div class="eco-domain__rows">`)

	feesCls := ""
	if d.LastBlockFees == "" {
		feesCls = ` class="eco-domain__row--inactive"`
	}
	ecoDomainRow(&b, feesCls, "last block fees", orEcoDash(trimFeeNote(d.LastBlockFees)), ecoTxFeesEffect(d))

	if !compact {
		b.WriteString(`<div><div class="eco-domain__param">fee market</div><div class="eco-domain__value">—</div><div class="eco-domain__effect">See <a href="/s/feemarket">Fee Market</a> for base fee and elasticity</div></div>`)
	}

	b.WriteString(`</div></div>`)
	return b.String()
}

func economicsLedgerTableHTML(rows []EcoLedgerRow) string {
	headers := []string{"Step", "Where", "In this block", "Balance now", "Check"}
	var b strings.Builder
	b.WriteString(`<div class="table-scroll"><table class="data-table data-table--ledger">`)
	b.WriteString(`<thead><tr>`)
	for i, h := range headers {
		thCls := ""
		if i > 0 {
			thCls = ` class="data-table__num"`
		}
		fmt.Fprintf(&b, `<th%s>%s</th>`, thCls, html.EscapeString(h))
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
			tdCls := ""
			if i > 0 {
				tdCls = ` class="data-table__num"`
			}
			fmt.Fprintf(&b, `<td%s>%s</td>`, tdCls, formatValue(cell))
		}
		b.WriteString(`</tr>`)
	}
	b.WriteString(`</tbody></table></div>`)
	return b.String()
}
