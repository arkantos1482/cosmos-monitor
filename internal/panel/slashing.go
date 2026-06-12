package panel

import (
	"fmt"
	"html"
	"strings"

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
	if d.JailedCount == 0 && d.BelowThreshold == 0 && d.SlashWindow != "" && d.SlashWindow != "0" {
		w.WriteHTML(fmt.Sprintf(
			`<div class="slashing-summary__row">window %s · min signed %.0f%%</div>`,
			html.EscapeString(d.SlashWindow), d.MinSigned))
	}
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
	writeEmbeddedSectionIntro(w, "Local missed blocks and jail/tombstone status; network slashing parameters and per-validator security table.")
	writeSlashingSummary(w, d, SummaryEmbedded)

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
		})
	}
	writeValidatorSetTable(w, []string{"moniker", "missed", "jailed", "tombstoned", "health"}, secRows, d.Validators)
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

// writeSlashingPenaltyMatrix renders a compact infraction → penalty map per Cosmos SDK x/slashing + x/evidence.
func writeSlashingPenaltyMatrix(b *strings.Builder, d model.Report) {
	dtTrigger := downtimeTriggerLabel(d)
	dsTrigger := "conflicting votes (equivocation)"

	b.WriteString(`<div class="slashing-penalties">`)
	b.WriteString(`<div class="eco-domain__divider">Penalties</div>`)
	b.WriteString(`<div class="table-scroll"><table class="data-table data-table--penalties"><thead><tr>`)
	for _, h := range []string{"infraction", "slash stake", "jail", "tombstone"} {
		fmt.Fprintf(b, `<th>%s</th>`, html.EscapeString(h))
	}
	b.WriteString(`</tr></thead><tbody>`)

	writeSlashingPenaltyRow(b, "downtime", dtTrigger,
		slashPenaltyCell(d.SlashDowntime, d.SlashDTInactive),
		jailPenaltyCell(d.DowntimeJail, false),
		penaltyNoCell())
	writeSlashingPenaltyRow(b, "double-sign", dsTrigger,
		slashPenaltyCell(d.SlashDS, d.SlashDSInactive),
		jailPenaltyCell("permanent", true),
		penaltyYesCell())

	b.WriteString(`</tbody></table></div></div>`)
}

func writeSlashingPenaltyRow(b *strings.Builder, name, trigger, slash, jail, tomb string) {
	fmt.Fprintf(b, `<tr><td class="slashing-penalties__infraction"><span class="slashing-penalties__name">%s</span>`,
		html.EscapeString(name))
	if trigger != "" {
		fmt.Fprintf(b, `<span class="slashing-penalties__trigger">%s</span>`, html.EscapeString(trigger))
	}
	fmt.Fprintf(b, `</td><td class="data-table__num">%s</td><td class="data-table__num">%s</td><td class="data-table__num">%s</td></tr>`,
		slash, jail, tomb)
}

func downtimeTriggerLabel(d model.Report) string {
	if d.SlashWindow == "" || d.SlashWindow == "0" || d.SlashMaxMissed <= 0 {
		return "missed blocks in window"
	}
	return fmt.Sprintf("miss > %s / %s window", report.FormatInt(d.SlashMaxMissed), d.SlashWindow)
}

func slashPenaltyCell(amount string, inactive bool) string {
	if inactive || amount == "" {
		return `<span class="penalty-tag penalty-tag--off">off</span>`
	}
	return fmt.Sprintf(`<span class="penalty-tag penalty-tag--slash">%s</span>`, html.EscapeString(amount))
}

func jailPenaltyCell(detail string, severe bool) string {
	if detail == "" {
		return penaltyNoCell()
	}
	cls := "penalty-tag penalty-tag--jail"
	if severe {
		cls += " penalty-tag--severe"
	}
	return fmt.Sprintf(`<span class="%s">%s</span>`, cls, html.EscapeString(detail))
}

func penaltyYesCell() string {
	return `<span class="penalty-tag penalty-tag--yes" title="permanent">yes</span>`
}

func penaltyNoCell() string {
	return `<span class="penalty-tag penalty-tag--off">—</span>`
}
