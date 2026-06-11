package panel

import (
	"fmt"
	"html"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeEconomicsSummary(w Writer, d model.Report, mode SummaryMode) {
	summaryWrapStart(w, mode, "economics")
	
	if mode == SummaryOverviewClickable {
		// Compact variant for home overview card
		writeEconomicsCompactSummary(w, d)
	} else {
		// Full domain cards for economics page (SummaryEmbedded)
		w.WriteHTML(`<div class="eco-summary">`)
		w.WriteHTML(economicsDomainCardsHTML(d, false))
		w.WriteHTML(`</div>`)
	}
	
	summaryWrapEnd(w, mode)
}

func writeEconomicsCompactSummary(w Writer, d model.Report) {
	w.WriteHTML(`<div class="eco-summary eco-summary--compact">`)

	pmtStatus := "disabled"
	if d.PMTEnabled {
		switch {
		case d.PMTPoolEmpty:
			pmtStatus = "pool empty"
		case d.PMTRate != "":
			pmtStatus = d.PMTRate
		default:
			pmtStatus = "enabled"
		}
	}
	w.WriteHTML(fmt.Sprintf(`<div class="eco-summary__row">PMT: %s</div>`, html.EscapeString(pmtStatus)))

	inflStatus := "off"
	if d.Inflation > 0 {
		inflStatus = fmt.Sprintf("%.2f%%", d.Inflation)
	}
	w.WriteHTML(fmt.Sprintf(`<div class="eco-summary__row">Inflation: %s</div>`, html.EscapeString(inflStatus)))

	w.WriteHTML(fmt.Sprintf(`<div class="eco-summary__row">Staking: %.2f%% bonded</div>`, d.BondedPct))

	w.WriteHTML(`</div>`)
}

func writeEconomicsKPIRows(w Writer, d model.Report) {
	if total := RewardInPerBlockTotal(d); total != "—" {
		w.Row("reward in / block", total)
	}
	if bal := FeeCollectorBalance(d); bal != "" {
		check := economicsFeeCollectorCheck(d)
		val := bal
		if check == "cleared" {
			val = bal + "  _(cleared each block)_"
		} else if check == "not cleared?" {
			val = bal + "  _(stuck — check distribution)_"
		}
		w.Row("fee_collector", val)
	}
	if d.CommunityPool != "" {
		w.Row("community pool", d.CommunityPool)
	}
	if del := economicsUnclaimedDelegator(d); del != "" {
		w.Row("unclaimed delegator", del)
	}
	if comm := economicsUnclaimedCommission(d); comm != "" {
		w.Row("unclaimed commission", comm)
	}
	if d.PMTEnabled && d.PMTBalance != "" {
		val := d.PMTBalance
		if d.PMTRunway != "" {
			val += "  (" + d.PMTRunway + ")"
		}
		w.Row("PMT pool", val)
	}
}

func writeEconomicsOverview(w Writer, d model.Report) {
	writeEconomicsLedger(w, d)
	w.Subsection("Distribution")
	writeEconomicsDistributionModule(w, d)
	writeEconomicsUnclaimedBalances(w, d)
	writeEconomicsCommunityTax(w, d)
}

func writeEconomicsDistributionModule(w Writer, d model.Report) {
	bal := distributionModuleBalance(d)
	addr := moduleAccountDisplayAddress(d, "distribution")
	if bal == "" && addr == "" {
		return
	}
	val := orEcoDash(bal)
	if addr != "" {
		if val != "—" {
			val += " @ " + addr
		} else {
			val = addr
		}
	}
	w.Row("distribution escrow", val)
}

func writeEconomicsCommunityTax(w Writer, d model.Report) {
	val := orEcoDash(d.CommunityTax)
	effect := ecoTaxEffect(d)
	if d.CommunityTaxZero {
		val += "  _(0% — community pool gets no cut)_"
	} else {
		val += "  _(" + effect + ")_"
	}
	w.Row("community tax", val)
}


func writeEconomicsUnclaimedBalances(w Writer, d model.Report) {
	del := economicsUnclaimedDelegator(d)
	comm := economicsUnclaimedCommission(d)
	if del == "" && comm == "" {
		return
	}
	w.Subsection("Unclaimed rewards")
	if del != "" {
		w.Row("delegator share", del)
	}
	if comm != "" {
		w.Row("validator commission", comm)
	}
	if total := economicsUnclaimedTotal(d); total != "" {
		val := total
		if _, suffix := splitOutstandingSuffix(d.TotalOutstanding); suffix != "" {
			val += "  _(" + suffix + ")_"
		}
		if check := economicsUnclaimedCheck(d); check != "" && check != "—" {
			val += "  _[" + check + "]_"
		}
		w.Row("total outstanding", val)
	}
}

func writeEconomicsLedger(w Writer, d model.Report) {
	rows := economicsLedgerRows(d)
	if len(rows) == 0 {
		return
	}
	w.Subsection("Block reward ledger")
	if intro := economicsLedgerIntro(d); intro != "" {
		w.Em(intro)
	}
	w.WriteHTML(economicsLedgerTableHTML(rows))
}

func economicsLedgerIntro(d model.Report) string {
	addr := moduleAccountDisplayAddress(d, "fee_collector")
	bal := FeeCollectorBalance(d)
	if addr == "" && bal == "" {
		return ""
	}
	var parts []string
	if addr != "" {
		parts = append(parts, "fee_collector "+addr)
	}
	if bal != "" {
		parts = append(parts, "balance "+bal)
	}
	return strings.Join(parts, " · ")
}

