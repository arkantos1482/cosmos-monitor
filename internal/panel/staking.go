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
	if badges := localBadges(d); len(badges) > 0 {
		writeSummaryBadges(w, "staking-summary__badges", badges...)
	}
	if d.JailedCount > 0 || d.BelowThreshold > 0 {
		w.WriteHTML(fmt.Sprintf(
			`<p class="staking-summary__alert">⚠ %d jailed · %d below min signed</p>`,
			d.JailedCount, d.BelowThreshold))
	}
	w.WriteHTML(`</div>`)
}

func writeStaking(w Writer, d model.Report) {
	lv := d.Local

	w.Section("1. STAKING")
	writeStakingSummary(w, d, SummaryEmbedded)
	w.Em("This validator — staking then slashing. Chain — pool and validator stake, then slashing params and per-validator signing.")

	w.Layer("This validator")
	if lv.IsValidator {
		writeStakingLocalStake(w, lv)
		writeStakingLocalSlashing(w, d, lv)
	} else {
		w.Subsection("Role")
		w.Hint("`role` → CometBFT GET /status; derived when consensus address is absent from x/staking.")
		w.Row("role", lv.SigningStatus)
	}

	w.Layer("Chain")
	writeStakingChainStake(w, d)
	writeStakingChainSlashing(w, d)
	w.Hint(stakingSourcesHint())
	w.BlankLine()
}

func writeStakingLocalStake(w Writer, lv model.LocalValidator) {
	w.Subsection("Staking")
	w.Hint("`status`, `voting power`, `commission` → REST GET /cosmos/staking/v1beta1/validators.")
	w.Row("status", lv.Status)
	w.Row("voting power", fmt.Sprintf("%s  (%.1f%% of bonded stake)", lv.VotingPower, lv.VPPercent))
	w.Row("commission", fmt.Sprintf("%.1f%%  _(validator cut of delegator rewards)_", lv.Commission))
}

func writeStakingLocalSlashing(w Writer, d model.Report, lv model.LocalValidator) {
	w.Subsection("Slashing")
	w.Hint("`jailed`, `tombstoned`, `signing health`, `missed / window` → x/staking validators + REST GET /cosmos/slashing/v1beta1/signing_infos and params.")
	if lv.Jailed {
		w.Row("jailed", "yes")
	}
	if lv.Tombstoned {
		w.Row("tombstoned", "YES")
	}
	w.Row("signing health", lv.SigningStatus)
	if d.SlashWindow != "" && d.SlashWindow != "0" {
		w.Row("missed / window", fmt.Sprintf("%d / %s blocks  (max allowed: %d)", lv.Missed, d.SlashWindow, lv.MaxMissed))
	}
}

func writeStakingChainStake(w Writer, d model.Report) {
	w.Subsection("Staking")
	w.WriteHTML(stakingCardHTML(d, false))
	writeValidatorStakeTable(w, d)
}

func writeStakingChainSlashing(w Writer, d model.Report) {
	w.Subsection("Slashing")
	w.WriteHTML(slashingCardHTML(d, false))
	writeValidatorSlashingTable(w, d)
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
			valLocalMark(v),
		})
	}
	w.Table([]string{"moniker", "vp%", "commission", "status", "local"}, stakeRows)
}

func writeValidatorSlashingTable(w Writer, d model.Report) {
	w.Hint("`missed`, `tombstoned` → REST GET /cosmos/slashing/v1beta1/signing_infos; `jailed` → module x/staking validators; `health` → derived (missed vs min_signed_per_window from slashing params).")
	secRows := make([][]string, 0, len(d.Validators))
	for _, v := range d.Validators {
		missed := fmt.Sprintf("%d", v.Missed)
		health := validatorSlashingHealth(v, &missed)
		jailed, tomb := "", ""
		if v.Jailed {
			jailed = "yes"
		}
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

func validatorSlashingHealth(v model.Validator, missed *string) string {
	switch {
	case v.Tombstoned:
		return "tombstoned"
	case v.Jailed:
		return "jailed"
	case v.MissedHigh:
		*missed += " ⚠"
		return "⚠ below min signed"
	case v.Missed > 0:
		return "ok (some misses)"
	default:
		return "ok"
	}
}

func stakingSourcesHint() string {
	return "`status`, `voting power`, `commission` → REST GET /cosmos/staking/v1beta1/validators; " +
		"`bonded`, `bond denom`, `unbonding time`, `max validators` → REST GET /cosmos/staking/v1beta1/pool, /cosmos/staking/v1beta1/params; " +
		"`signed blocks window`, `min signed`, `slash fractions` → REST GET /cosmos/slashing/v1beta1/params; " +
		"`module account balances` → REST GET /cosmos/bank/v1beta1/balances/{address}; " +
		"`module account addresses` → REST GET /cosmos/auth/v1beta1/module_accounts."
}
