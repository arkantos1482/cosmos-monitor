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
			`<div class="staking-summary__row">%s · %.1f%% VP · %.1f%% commission</div>`,
			html.EscapeString(lv.Status), lv.VPPercent, lv.Commission))
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
	if lv.IsValidator {
		w.WriteHTML(`<div class="staking-summary__local">`)
		writeSummaryBadges(w, "staking-summary__badges", localBadges(d)...)
		w.WriteHTML(fmt.Sprintf(
			`<span class="staking-summary__vp">%.1f%% VP · %s · %.1f%% commission</span>`,
			lv.VPPercent, html.EscapeString(lv.Status), lv.Commission))
		w.WriteHTML(`</div>`)
	}
	writeStakingChainSummaryBody(w, d)
	w.WriteHTML(`</div>`)
}

func writeStaking(w Writer, d model.Report) {
	lv := d.Local

	w.Section("1. STAKING")
	writeStakingSummary(w, d, SummaryEmbedded)
	w.Em("This validator stake and commission, then network bonded pool and stake table.")

	w.Subsection("This validator")
	if lv.IsValidator {
		writeStakingLocalStake(w, lv)
	} else {
		w.Hint("`role` → CometBFT GET /status; derived when consensus address is absent from x/staking.")
		w.Row("role", lv.SigningStatus)
	}
	w.Subsection("Network-wide")
	w.WriteHTML(stakingCardHTML(d, false))
	writeValidatorStakeTable(w, d)

	w.Hint(stakingSourcesHint())
	w.BlankLine()
}

func writeStakingLocalStake(w Writer, lv model.LocalValidator) {
	w.Hint("`status`, `voting power`, `commission` → REST GET /cosmos/staking/v1beta1/validators.")
	w.Row("status", lv.Status)
	w.Row("voting power", fmt.Sprintf("%s  (%.1f%% of bonded stake)", lv.VotingPower, lv.VPPercent))
	w.Row("commission", fmt.Sprintf("%.1f%%  _(validator cut of delegator rewards)_", lv.Commission))
}

func writeValidatorStakeTable(w Writer, d model.Report) {
	w.Hint("`vp%`, `commission`, `status` → REST GET /cosmos/staking/v1beta1/validators (bonded, unbonding, unbonded).")
	stakeRows := make([][]string, 0, len(d.Validators))
	for _, v := range d.Validators {
		stakeRows = append(stakeRows, []string{
			report.Truncate(v.Moniker, 14),
			fmt.Sprintf("%.1f%%", v.VPFloat),
			fmt.Sprintf("%.1f%%", v.CommissionFloat),
			v.Status,
		})
	}
	writeValidatorSetTable(w, []string{"moniker", "vp%", "commission", "status"}, stakeRows, d.Validators)
}

func stakingSourcesHint() string {
	return "`status`, `voting power`, `commission` → REST GET /cosmos/staking/v1beta1/validators; " +
		"`bonded`, `bond denom`, `unbonding time`, `max validators` → REST GET /cosmos/staking/v1beta1/pool, /cosmos/staking/v1beta1/params; " +
		"`module account balances` → REST GET /cosmos/bank/v1beta1/balances/{address}; " +
		"`module account addresses` → REST GET /cosmos/auth/v1beta1/module_accounts."
}
