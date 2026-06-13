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
	writeSlashingSummaryBody(w, d, lv)
	summaryWrapEnd(w, mode)
}

func writeSlashingSummaryBody(w Writer, d model.Report, lv model.LocalValidator) {
	w.WriteHTML(`<div class="slashing-summary">`)
	if badges := localBadges(d); len(badges) > 0 {
		writeSummaryBadges(w, "slashing-summary__badges", badges...)
	}
	if lv.IsValidator {
		w.WriteHTML(`<div class="slashing-summary__kpis">`)
		writeSlashingSummaryKPI(w, "signing health", slashingHealthShort(lv), slashingHealthTone(lv))
		if d.SlashWindow != "" && d.SlashWindow != "0" {
			writeSlashingSummaryKPI(w, "missed / window",
				fmt.Sprintf("%d / %s", lv.Missed, d.SlashWindow), slashingMissedTone(lv))
			if lv.MaxMissed > 0 {
				writeSlashingSummaryKPI(w, "headroom",
					fmt.Sprintf("%d blocks", slashingHeadroom(lv)), slashingHeadroomTone(lv))
			}
		}
		w.WriteHTML(`</div>`)
	} else if lv.SigningStatus != "" {
		w.WriteHTML(fmt.Sprintf(
			`<p class="slashing-summary__role">%s</p>`,
			html.EscapeString(lv.SigningStatus)))
	}
	writeSlashingNetworkKPIs(w, d)
	w.WriteHTML(`</div>`)
}

func writeSlashingSummaryKPI(w Writer, label, value, tone string) {
	if value == "" {
		return
	}
	valCls := "slashing-summary__kpi-val"
	if tone != "" {
		valCls += " slashing-summary__kpi-val--" + tone
	}
	w.WriteHTML(fmt.Sprintf(
		`<div class="slashing-summary__kpi"><span class="slashing-summary__kpi-label">%s</span>`+
			`<span class="%s">%s</span></div>`,
		html.EscapeString(label), valCls, html.EscapeString(value)))
}

func writeSlashingNetworkKPIs(w Writer, d model.Report) {
	w.WriteHTML(`<div class="slashing-summary__kpis slashing-summary__kpis--network">`)
	if d.JailedCount > 0 {
		writeSlashingSummaryKPI(w, "jailed", fmt.Sprintf("%d", d.JailedCount), "bad")
	}
	if d.BelowThreshold > 0 {
		writeSlashingSummaryKPI(w, "below min signed", fmt.Sprintf("%d", d.BelowThreshold), "warn")
	}
	if d.TombstonedCount > 0 {
		writeSlashingSummaryKPI(w, "tombstoned", fmt.Sprintf("%d", d.TombstonedCount), "bad")
	}
	if d.SlashWindow != "" && d.SlashWindow != "0" {
		writeSlashingSummaryKPI(w, "window", d.SlashWindow+" blocks", "")
		if d.MinSigned > 0 {
			writeSlashingSummaryKPI(w, "min signed", fmt.Sprintf("%.0f%%", d.MinSigned), "")
		}
		if d.SlashMaxMissed > 0 {
			writeSlashingSummaryKPI(w, "max missed", report.FormatInt(d.SlashMaxMissed), "")
		}
	}
	w.WriteHTML(`</div>`)
}

func slashingHealthShort(lv model.LocalValidator) string {
	switch {
	case lv.Tombstoned:
		return "tombstoned"
	case lv.Jailed:
		return "jailed"
	case lv.MissedHigh:
		return fmt.Sprintf("below min signed (%d missed)", lv.Missed)
	case lv.Missed > 0:
		return fmt.Sprintf("ok · %d missed", lv.Missed)
	default:
		return "ok · no misses"
	}
}

func slashingHeadroom(lv model.LocalValidator) int64 {
	if lv.MaxMissed <= 0 {
		return 0
	}
	remaining := lv.MaxMissed - lv.Missed
	if remaining < 0 {
		return 0
	}
	return remaining
}

func slashingHealthTone(lv model.LocalValidator) string {
	switch {
	case lv.Tombstoned, lv.Jailed:
		return "bad"
	case lv.MissedHigh:
		return "warn"
	default:
		return "ok"
	}
}

func slashingMissedTone(lv model.LocalValidator) string {
	if lv.MissedHigh {
		return "warn"
	}
	return ""
}

func slashingHeadroomTone(lv model.LocalValidator) string {
	if lv.MaxMissed <= 0 {
		return ""
	}
	pct := float64(slashingHeadroom(lv)) / float64(lv.MaxMissed) * 100
	switch {
	case pct <= 10:
		return "bad"
	case pct <= 25:
		return "warn"
	default:
		return "ok"
	}
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
		w.Row("role", lv.SigningStatus)
	}
	w.Subsection("Network-wide")
	w.WriteHTML(slashingCardHTML(d, false))
	writeValidatorSlashingTable(w, d)

	writeSectionSources(w, ViewSlashing, d)
	w.BlankLine()
}

func writeSlashingLocal(w Writer, d model.Report, lv model.LocalValidator) {
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

// writeSlashingPenaltyMatrix renders a compact infraction → penalty map per Cosmos SDK x/slashing + x/evidence.
func writeSlashingPenaltyMatrix(b *strings.Builder, d model.Report) {
	dtTrigger := downtimeTriggerLabel(d)
	dsTrigger := "conflicting votes (equivocation)"

	b.WriteString(`<div class="slashing-penalties">`)
	b.WriteString(`<div class="eco-domain__divider">Penalties <span class="eco-domain__subtitle">chain parameters</span></div>`)
	b.WriteString(`<p class="slashing-penalties__note">Configured slash fractions and jail rules — not live slashing events.</p>`)
	b.WriteString(`<div class="table-scroll"><table class="data-table data-table--penalties"><thead><tr>`)
	penaltyHeaders := []string{"infraction", "slash stake", "jail", "tombstone"}
	for _, h := range penaltyHeaders {
		fmt.Fprintf(b, `<th%s>%s</th>`, tableColumnClass(h), html.EscapeString(h))
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
	fmt.Fprintf(b, `</td><td class="data-table__num">%s</td><td class="data-table__center">%s</td><td class="data-table__center">%s</td></tr>`,
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
	return fmt.Sprintf(`<span class="penalty-tag penalty-tag--param">%s</span>`, html.EscapeString(amount))
}

func jailPenaltyCell(detail string, severe bool) string {
	if detail == "" {
		return penaltyNoCell()
	}
	cls := "penalty-tag penalty-tag--param"
	if severe {
		cls += " penalty-tag--emph"
	}
	return fmt.Sprintf(`<span class="%s">%s</span>`, cls, html.EscapeString(detail))
}

func penaltyYesCell() string {
	return `<span class="penalty-tag penalty-tag--param" title="permanent">yes</span>`
}

func penaltyNoCell() string {
	return `<span class="penalty-tag penalty-tag--off">—</span>`
}
