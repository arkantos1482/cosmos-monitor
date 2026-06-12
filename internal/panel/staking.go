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
	if d.JailedCount > 0 || d.BelowThreshold > 0 {
		w.WriteHTML(fmt.Sprintf(
			`<div class="staking-summary__row staking-summary__row--warn">%d jailed · %d below min signed</div>`,
			d.JailedCount, d.BelowThreshold))
	}
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
	w.WriteHTML(stakingChainDomainsHTML(d, true))
	w.WriteHTML(`</div>`)
}

func writeStaking(w Writer, d model.Report) {
	lv := d.Local

	w.Section("1. STAKING")
	writeStakingSummary(w, d, SummaryEmbedded)
	w.Em("This validator's stake and signing health first, then chain-wide pool, parameters, and validator tables.")

	w.Layer("This validator")
	if lv.IsValidator {
		writeStakingLocalState(w, d, lv)
	} else {
		w.Subsection("Role")
		w.Hint("`role` → CometBFT GET /status; derived when consensus address is absent from x/staking.")
		w.Row("role", lv.SigningStatus)
	}

	w.Layer("Chain")
	w.WriteHTML(stakingChainDomainsHTML(d, false))
	writeValidatorStakeTable(w, d)
	writeValidatorSlashingTable(w, d)
	w.Hint(stakingSourcesHint())
	w.BlankLine()
}

func writeStakingLocalState(w Writer, d model.Report, lv model.LocalValidator) {
	w.Subsection("Stake")
	w.Hint("`status`, `jailed`, `voting power`, `commission` → REST GET /cosmos/staking/v1beta1/validators.")
	w.Row("status", lv.Status)
	if lv.Jailed {
		w.Row("jailed", "yes")
	}
	if lv.Tombstoned {
		w.Row("tombstoned", "YES")
	}
	w.Row("voting power", fmt.Sprintf("%s  (%.1f%% of bonded stake)", lv.VotingPower, lv.VPPercent))
	w.Row("commission", fmt.Sprintf("%.1f%%  _(validator cut of delegator rewards)_", lv.Commission))

	w.Subsection("Signing")
	w.Hint("`signing health`, `missed / window` → REST GET /cosmos/slashing/v1beta1/signing_infos + params.")
	w.Row("signing health", lv.SigningStatus)
	if d.SlashWindow != "" && d.SlashWindow != "0" {
		w.Row("missed / window", fmt.Sprintf("%d / %s blocks  (max allowed: %d)", lv.Missed, d.SlashWindow, lv.MaxMissed))
	}
}

func writeValidatorStakeTable(w Writer, d model.Report) {
	w.Subsection("Validator stake")
	w.Hint("`vp%`, `commission`, `status` → REST GET /cosmos/staking/v1beta1/validators (bonded, unbonding, unbonded).")
	stakeRows := make([][]string, 0, len(d.Validators))
	for _, v := range d.Validators {
		stakeRows = append(stakeRows, []string{
			report.Truncate(v.Moniker, 14),
			fmt.Sprintf("%.1f%%", v.VPFloat),
			fmt.Sprintf("%.1f%%", v.CommissionFloat),
			v.Status,
			valLocalMark(v),
		})
	}
	w.Table([]string{"moniker", "vp%", "commission", "status", "local"}, stakeRows)
}

func writeValidatorSlashingTable(w Writer, d model.Report) {
	w.Subsection("Validator slashing")
	w.Hint("`missed`, `tombstoned` → REST GET /cosmos/slashing/v1beta1/signing_infos; `jailed` → module x/staking validators; `health` → derived (missed vs min_signed_per_window from slashing params).")
	secRows := make([][]string, 0, len(d.Validators))
	for _, v := range d.Validators {
		missed := fmt.Sprintf("%d", v.Missed)
		health := "ok"
		if v.Tombstoned {
			health = "tombstoned"
		} else if v.Jailed {
			health = "jailed"
		} else if v.MissedHigh {
			health = "⚠ below min signed"
			missed += " ⚠"
		} else if v.Missed > 0 {
			health = "ok (some misses)"
		}
		jailed := ""
		if v.Jailed {
			jailed = "yes"
		}
		tomb := ""
		if v.Tombstoned {
			tomb = "yes"
		}
		secRows = append(secRows, []string{
			report.Truncate(v.Moniker, 14),
			missed,
			jailed,
			tomb,
			health,
			valLocalMark(v),
		})
	}
	w.Table([]string{"moniker", "missed", "jailed", "tombstoned", "health", "local"}, secRows)
}

func stakingChainDomainsHTML(d model.Report, compact bool) string {
	return ecoDomainsWrap(stakingCardHTML(d, compact), slashingCardHTML(d, compact))
}

func stakingSourcesHint() string {
	return "`status`, `voting power`, `commission`, `jailed` → REST GET /cosmos/staking/v1beta1/validators; " +
		"`bonded`, `bond denom`, `unbonding time`, `max validators` → REST GET /cosmos/staking/v1beta1/pool, /cosmos/staking/v1beta1/params; " +
		"`signed blocks window`, `min signed`, `slash fractions` → REST GET /cosmos/slashing/v1beta1/params; " +
		"`module account balances` → REST GET /cosmos/bank/v1beta1/balances/{address}; " +
		"`module account addresses` → REST GET /cosmos/auth/v1beta1/module_accounts."
}
