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
	
	// One-line health per domain (3 rows max)
	rewardStatus := "no inflow"
	if economicsHasRewardSource(d) {
		if total := RewardInPerBlockTotal(d); total != "—" {
			rewardStatus = total + "/block"
		} else {
			rewardStatus = "active"
		}
	}
	w.WriteHTML(fmt.Sprintf(`<div class="eco-summary__row">Rewards in: %s</div>`, rewardStatus))
	
	distStatus := "0% tax → pool empty"
	if !d.CommunityTaxZero && d.CommunityTax != "" {
		poolDisplay := d.CommunityPool
		if poolDisplay == "" {
			poolDisplay = "0 PMT"
		}
		distStatus = fmt.Sprintf("%s tax → pool %s", d.CommunityTax, poolDisplay)
	}
	w.WriteHTML(fmt.Sprintf(`<div class="eco-summary__row">Distribution: %s</div>`, distStatus))
	
	w.WriteHTML(fmt.Sprintf(`<div class="eco-summary__row">Staking: %.2f%% bonded (goal %.0f%%)</div>`, 
		d.BondedPct, d.GoalBonded))
	
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
	writeEconomicsModuleAccountsTable(w, d)
	writeEconomicsFlagsPanel(w, d)
}


func writeEconomicsModuleAccountsTable(w Writer, d model.Report) {
	if len(d.ModuleAccounts) == 0 {
		return
	}
	w.Subsection("Module accounts")
	w.Hint("Live balances from x/bank, addresses from x/auth module accounts, ledger step mapping shows where each appears in the reward flow.")
	
	// Build ledger step mapping
	stepMap := map[string]string{
		"fee_collector":          "4",
		"distribution":           "6", 
		"bonded_tokens_pool":     "—",
		"not_bonded_tokens_pool": "—",
	}
	
	w.WriteHTML(writeEconomicsModuleAccountsTableHTML(d.ModuleAccounts, stepMap))
}

func writeEconomicsModuleAccountsTableHTML(accounts []model.ModuleAccountRow, stepMap map[string]string) string {
	headers := []string{"Module", "Balance", "Address", "Ledger step", "Role"}
	var b strings.Builder
	b.WriteString(`<div class="eco-module-accounts"><div class="table-scroll">`)
	b.WriteString(`<table class="data-table"><thead><tr>`)
	for i, h := range headers {
		thCls := ""
		if i == 1 || i == 3 { // Balance and Ledger step columns
			thCls = ` class="data-table__num"`
		}
		fmt.Fprintf(&b, `<th%s>%s</th>`, thCls, html.EscapeString(h))
	}
	b.WriteString(`</tr></thead><tbody>`)
	
	for _, acc := range accounts {
		step, hasStep := stepMap[acc.Name]
		if !hasStep {
			step = "—"
		}
		
		b.WriteString(`<tr>`)
		fmt.Fprintf(&b, `<td class="eco-module-accounts__module"><code>%s</code></td>`, html.EscapeString(acc.Name))
		fmt.Fprintf(&b, `<td class="eco-module-accounts__balance data-table__num">%s</td>`, html.EscapeString(acc.Balance))
		fmt.Fprintf(&b, `<td class="eco-module-accounts__address">%s</td>`, html.EscapeString(acc.Address))
		fmt.Fprintf(&b, `<td class="eco-module-accounts__ledger data-table__num">%s</td>`, html.EscapeString(step))
		fmt.Fprintf(&b, `<td class="eco-module-accounts__role">%s</td>`, html.EscapeString(acc.Role))
		b.WriteString(`</tr>`)
	}
	
	b.WriteString(`</tbody></table></div></div>`)
	return b.String()
}

func writeEconomicsLedger(w Writer, d model.Report) {
	rows := economicsLedgerRows(d)
	if len(rows) == 0 {
		return
	}
	w.Subsection("Block reward ledger")
	w.Hint("`In this block` → derived (per-block mint, PMT, fees); `Balance now` → module x/bank balances; `Check` → derived (fee_collector cleared, pool drift); `reward flow` → derived (BeginBlock via fee_collector and x/distribution, see table). Grey/red rows are inactive on this chain.")
	w.WriteHTML(economicsLedgerTableHTML(rows))
}

func writeEconomicsFlagsPanel(w Writer, d model.Report) {
	flags := economicsFlags(d)
	if len(flags) == 0 {
		return
	}
	w.Subsection("Advanced parameters (reward flow)")
	w.Em("Governance and module knobs that determine whether each ledger step is active. Red rows are ineffective on this chain right now.")
	w.WriteHTML(economicsFlagsTableHTML(flags))
}
