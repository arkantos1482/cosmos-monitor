package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func writeSlashingSummary(w Writer, d model.Report, mode SummaryMode) {
	lv := d.Local
	summaryWrapStart(w, mode, "slashing")

	if mode == SummaryOverviewClickable {
		writeSlashingCompactSummary(w, d, lv)
	} else {
		writeSlashingEmbeddedSummary(w, d, lv)
	}

	summaryWrapEnd(w, mode)
}

func writeSlashingCompactSummary(w Writer, d model.Report, lv model.LocalValidator) {
	w.WriteHTML(`<div class="slashing-summary slashing-summary--compact">`)
	if lv.IsValidator && lv.SigningStatus != "" {
		w.WriteHTML(fmt.Sprintf(
			`<div class="slashing-summary__row">%s</div>`,
			html.EscapeString(lv.SigningStatus)))
		if d.SlashWindow != "" && d.SlashWindow != "0" {
			w.WriteHTML(fmt.Sprintf(
				`<div class="slashing-summary__row">%d / %s missed</div>`,
				lv.Missed, html.EscapeString(d.SlashWindow)))
		}
	}
	if d.JailedCount > 0 || d.BelowThreshold > 0 {
		w.WriteHTML(fmt.Sprintf(
			`<div class="slashing-summary__row slashing-summary__row--warn">%d jailed · %d below min signed</div>`,
			d.JailedCount, d.BelowThreshold))
	} else if d.SlashWindow != "" && d.SlashWindow != "0" {
		w.WriteHTML(fmt.Sprintf(
			`<div class="slashing-summary__row">window %s · min signed %.0f%%</div>`,
			html.EscapeString(d.SlashWindow), d.MinSigned))
	}
	w.WriteHTML(`</div>`)
}

func writeSlashingEmbeddedSummary(w Writer, d model.Report, lv model.LocalValidator) {
	w.WriteHTML(`<div class="slashing-summary">`)
	if lv.IsValidator {
		w.WriteHTML(`<div class="slashing-summary__local">`)
		writeSummaryBadges(w, "slashing-summary__badges", localBadges(d)...)
		if lv.SigningStatus != "" {
			w.WriteHTML(fmt.Sprintf(
				`<span class="slashing-summary__health">%s</span>`,
				html.EscapeString(lv.SigningStatus)))
		}
		w.WriteHTML(`</div>`)
	}
	writeSlashingNetworkSummaryBody(w, d)
	w.WriteHTML(`</div>`)
}

func writeSlashingNetworkSummaryBody(w Writer, d model.Report) {
	if d.JailedCount == 0 && d.BelowThreshold == 0 {
		return
	}
	w.WriteHTML(`<div class="val-summary val-summary--slashing">`)
	var alerts []string
	if d.JailedCount > 0 {
		alerts = append(alerts, fmt.Sprintf("%d jailed", d.JailedCount))
	}
	if d.BelowThreshold > 0 {
		alerts = append(alerts, fmt.Sprintf("%d below min signed", d.BelowThreshold))
	}
	for _, alert := range alerts {
		w.WriteHTML(fmt.Sprintf(`<p class="val-summary__alert">⚠ %s</p>`, html.EscapeString(alert)))
	}
	w.WriteHTML(`</div>`)
}

func writeSlashing(w Writer, d model.Report) {
	lv := d.Local

	w.Section("2. SLASHING")
	writeSlashingSummary(w, d, SummaryEmbedded)
	w.Em("Signing health for this validator, network slashing parameters, and per-validator missed-block status.")

	w.Subsection("This validator")
	if lv.IsValidator {
		writeSlashingLocal(w, d, lv)
	} else {
		w.Hint("`role` → CometBFT GET /status; derived when consensus address is absent from x/staking.")
		w.Row("role", lv.SigningStatus)
	}
	w.Subsection("Network-wide")
	w.WriteHTML(slashingCardHTML(d, false))
	writeValidatorSlashingTable(w, d)

	w.Hint(slashingSourcesHint())
	w.BlankLine()
}

func writeSlashingLocal(w Writer, d model.Report, lv model.LocalValidator) {
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

func slashingSourcesHint() string {
	return "`jailed` → REST GET /cosmos/staking/v1beta1/validators; " +
		"`missed`, `tombstoned` → REST GET /cosmos/slashing/v1beta1/signing_infos; " +
		"`signed blocks window`, `min signed`, `slash fractions` → REST GET /cosmos/slashing/v1beta1/params."
}
