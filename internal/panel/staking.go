package panel

import (
	"fmt"
	"html"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func writeStakingSummary(w Writer, d model.Report, mode SummaryMode) {
	lv := d.Local
	summaryWrapStart(w, mode, "staking")
	writeStakingSummaryBody(w, d, lv)
	summaryWrapEnd(w, mode)
}

func writeStakingSummaryBody(w Writer, d model.Report, lv model.LocalValidator) {
	w.WriteHTML(`<div class="staking-summary">`)
	if badges := localBadges(d); len(badges) > 0 {
		writeSummaryBadges(w, "staking-summary__badges", badges...)
	}
	if lv.IsValidator {
		w.WriteHTML(`<div class="staking-summary__kpis">`)
		writeStakingSummaryKPI(w, "voting power", fmt.Sprintf("%.1f%%", lv.VPPercent), "")
		writeStakingSummaryKPI(w, "status", lv.Status, "")
		writeStakingSummaryKPI(w, "commission", fmt.Sprintf("%.1f%%", lv.Commission), "")
		if lv.VotingPower != "" {
			writeStakingSummaryKPI(w, "bonded stake", lv.VotingPower, "")
		}
		w.WriteHTML(`</div>`)
	} else if lv.SigningStatus != "" {
		w.WriteHTML(fmt.Sprintf(
			`<p class="staking-summary__role">%s</p>`,
			html.EscapeString(lv.SigningStatus)))
	}
	w.WriteHTML(`<div class="staking-summary__kpis staking-summary__kpis--network">`)
	if d.BondedCount > 0 {
		writeStakingSummaryKPI(w, "active set", fmt.Sprintf("%d", d.BondedCount), "")
	}
	if d.BondedPct > 0 {
		writeStakingSummaryKPI(w, "bonded", fmt.Sprintf("%.1f%%", d.BondedPct), "")
	}
	if unhealthy := stakingUnhealthyCount(d); unhealthy > 0 {
		writeStakingSummaryKPI(w, "unhealthy", fmt.Sprintf("%d", unhealthy), "warn")
	}
	if d.BondedAmt != "" {
		writeStakingSummaryKPI(w, "total staked", d.BondedAmt, "")
	}
	w.WriteHTML(`</div>`)
	writeStakingValidatorChips(w, d)
	w.WriteHTML(`</div>`)
}

func stakingUnhealthyCount(d model.Report) int {
	return d.JailedCount + d.BelowThreshold + d.TombstonedCount
}

func writeStakingValidatorChips(w Writer, d model.Report) {
	if len(d.Validators) == 0 {
		return
	}
	w.WriteHTML(`<p class="staking-summary__heading">Validator set</p>`)
	w.WriteHTML(`<div class="staking-summary__chips">`)
	for _, v := range d.Validators {
		cls := stakingChipClass(v)
		if v.IsLocal {
			cls += " staking-summary__chip--local"
		}
		vp := ""
		if v.VPFloat > 0 {
			vp = fmt.Sprintf(` <span class="staking-summary__chip-vp">%.0f%%</span>`, v.VPFloat)
		}
		w.WriteHTML(fmt.Sprintf(
			`<span class="staking-summary__chip%s" title="%s">%s%s</span>`,
			cls, html.EscapeString(stakingChipTitle(v)),
			html.EscapeString(report.Truncate(v.Moniker, 14)), vp))
	}
	w.WriteHTML(`</div>`)
}

func stakingChipClass(v model.Validator) string {
	if v.Jailed || v.Tombstoned || v.MissedHigh {
		return " staking-summary__chip--warn"
	}
	return ""
}

func stakingChipTitle(v model.Validator) string {
	if v.Jailed {
		return "jailed"
	}
	if v.Tombstoned {
		return "tombstoned"
	}
	return v.Status
}

func writeStakingSummaryKPI(w Writer, label, value, tone string) {
	if value == "" {
		return
	}
	valCls := "staking-summary__kpi-val"
	if tone != "" {
		valCls += " staking-summary__kpi-val--" + tone
	}
	w.WriteHTML(fmt.Sprintf(
		`<div class="staking-summary__kpi"><span class="staking-summary__kpi-label">%s</span>`+
			`<span class="%s">%s</span></div>`,
		html.EscapeString(label), valCls, html.EscapeString(value)))
}

func writeStaking(w Writer, d model.Report) {
	lv := d.Local

	w.Section("1. STAKING")
	writeEmbeddedSectionIntro(w, "Local validator accounts and stake; network bonded pool, module balances, and validator set.")
	writeStakingSummary(w, d, SummaryEmbedded)

	w.Subsection("This validator")
	if lv.IsValidator {
		writeStakingLocal(w, lv)
		writeStakingDelegators(w, lv, d.BondDenom)
	} else {
		w.Row("role", lv.SigningStatus)
	}
	w.Subsection("Network-wide")
	w.WriteHTML(stakingCardHTML(d, false))
	writeValidatorStakingTable(w, d)

	writeSectionSources(w, ViewStaking, d)
	w.BlankLine()
}

func writeStakingLocal(w Writer, lv model.LocalValidator) {
	if moniker := strings.TrimSpace(lv.Moniker); moniker != "" {
		w.Row("moniker", moniker)
	}
	if idHTML := validatorStakingIdentityHTML(lv); idHTML != "" {
		w.WriteHTML(idHTML)
	}
	if lv.Status != "" {
		w.Row("status", lv.Status)
	}
	if lv.Jailed {
		w.Row("jailed", "yes")
	}
	if lv.Tombstoned {
		w.Row("tombstoned", "yes")
	}
	if vpHTML := votingPowerHTML(lv.VotingPower, lv.VPPercent); vpHTML != "" {
		bar := ""
		if lv.VPPercent > 0 {
			bar = fmt.Sprintf(`<div class="kpi-bar"><div class="kpi-bar__fill" style="width:%.1f%%"></div></div>`, lv.VPPercent)
		}
		w.RowHTML("voting power", vpHTML, bar)
	}
	if lv.Commission > 0 {
		w.Row("commission", fmt.Sprintf("%.1f%%", lv.Commission))
	}
	if lv.LiquidBalance != "" {
		w.Row("liquid balance", lv.LiquidBalance+"  _(bank — spendable, excl. bonded)_")
	}
	if lv.DelegatorCount > 0 {
		w.Row("delegators", fmt.Sprintf("%d", lv.DelegatorCount))
	}
}

func writeValidatorStakingTable(w Writer, d model.Report) {
	rows := make([][]string, 0, len(d.Validators))
	for _, v := range d.Validators {
		rows = append(rows, []string{
			report.Truncate(v.Moniker, 14),
			identityCell(v.Operator),
			fmt.Sprintf("%.1f%%", v.VPFloat),
			fmt.Sprintf("%.1f%%", v.CommissionFloat),
			v.Status,
		})
	}
	writeValidatorSetTable(w, []string{"moniker", "operator", "vp%", "commission", "status"}, rows, d.Validators)
}
