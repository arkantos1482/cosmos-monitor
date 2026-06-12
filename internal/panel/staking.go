package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func writeStakingSummary(w Writer, d model.Report, mode SummaryMode) {
	lv := d.Local
	summaryWrapStart(w, mode, "staking")

	if mode == SummaryOverviewClickable {
		writeStakingCompactSummary(w, d, lv)
	} else {
		writeStakingEmbeddedSummary(w, d, lv)
	}

	summaryWrapEnd(w, mode)
}

func writeStakingCompactSummary(w Writer, d model.Report, lv model.LocalValidator) {
	w.WriteHTML(`<div class="staking-summary staking-summary--compact">`)
	if lv.IsValidator {
		w.WriteHTML(fmt.Sprintf(
			`<div class="staking-summary__row">%.1f%% VP · %s · %.1f%% commission</div>`,
			lv.VPPercent, html.EscapeString(lv.Status), lv.Commission))
	} else if lv.SigningStatus != "" {
		w.WriteHTML(fmt.Sprintf(`<div class="staking-summary__row">%s</div>`, html.EscapeString(lv.SigningStatus)))
	}
	w.WriteHTML(fmt.Sprintf(
		`<div class="staking-summary__row">%.2f%% bonded · %d active</div>`,
		d.BondedPct, d.BondedCount))
	w.WriteHTML(`</div>`)
}

func writeStakingEmbeddedSummary(w Writer, d model.Report, lv model.LocalValidator) {
	w.WriteHTML(`<div class="staking-summary">`)
	if badges := localBadges(d); len(badges) > 0 {
		writeSummaryBadges(w, "staking-summary__badges", badges...)
	}
	if lv.IsValidator {
		w.WriteHTML(`<div class="staking-summary__kpis">`)
		writeStakingSummaryKPI(w, "voting power", fmt.Sprintf("%.1f%%", lv.VPPercent))
		writeStakingSummaryKPI(w, "status", lv.Status)
		writeStakingSummaryKPI(w, "commission", fmt.Sprintf("%.1f%%", lv.Commission))
		if lv.VotingPower != "" {
			writeStakingSummaryKPI(w, "bonded stake", lv.VotingPower)
		}
		w.WriteHTML(`</div>`)
	} else if lv.SigningStatus != "" {
		w.WriteHTML(fmt.Sprintf(
			`<p class="staking-summary__role">%s</p>`,
			html.EscapeString(lv.SigningStatus)))
	}
	w.WriteHTML(`<div class="staking-summary__kpis staking-summary__kpis--network">`)
	writeStakingSummaryKPI(w, "network bonded", fmt.Sprintf("%.2f%%", d.BondedPct))
	writeStakingSummaryKPI(w, "active set", fmt.Sprintf("%d", d.BondedCount))
	if d.JailedCount > 0 {
		writeStakingSummaryKPI(w, "jailed", fmt.Sprintf("%d", d.JailedCount))
	}
	if d.BondedAmt != "" {
		writeStakingSummaryKPI(w, "total staked", d.BondedAmt)
	}
	w.WriteHTML(`</div></div>`)
}

func writeStakingSummaryKPI(w Writer, label, value string) {
	w.WriteHTML(fmt.Sprintf(
		`<div class="staking-summary__kpi"><span class="staking-summary__kpi-label">%s</span>`+
			`<span class="staking-summary__kpi-val">%s</span></div>`,
		html.EscapeString(label), html.EscapeString(value)))
}

func writeStaking(w Writer, d model.Report) {
	lv := d.Local

	w.Section("1. STAKING")
	writeEmbeddedSectionIntro(w, "Local validator accounts and stake; network bonded pool, module balances, and validator set.")
	writeStakingSummary(w, d, SummaryEmbedded)

	w.Subsection("This validator")
	if lv.IsValidator {
		writeStakingDelegators(w, lv)
	} else {
		w.Hint("`role` → CometBFT GET /status; derived when consensus address is absent from x/staking.")
		w.Row("role", lv.SigningStatus)
	}
	w.Subsection("Network-wide")
	w.WriteHTML(stakingCardHTML(d, false))
	writeValidatorStakingTable(w, d)

	w.Hint(stakingSourcesHint())
	w.BlankLine()
}

func writeValidatorStakingTable(w Writer, d model.Report) {
	w.Hint("`operator`, `vp%`, `commission`, `status` → REST GET /cosmos/staking/v1beta1/validators.")
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

func stakingSourcesHint() string {
	return "`status`, `voting power`, `commission` → REST GET /cosmos/staking/v1beta1/validators; " +
		"`bonded`, `bond denom`, `unbonding time`, `max validators` → REST GET /cosmos/staking/v1beta1/pool, /cosmos/staking/v1beta1/params; " +
		"`module account balances` → REST GET /cosmos/bank/v1beta1/balances/{address}; " +
		"`module account addresses` → REST GET /cosmos/auth/v1beta1/module_accounts."
}
