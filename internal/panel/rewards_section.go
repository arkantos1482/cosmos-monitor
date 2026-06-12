package panel

import (
	"fmt"
	"html"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeRewardsDistributionSummaryRows(w Writer, d model.Report) {
	if d.CommunityPool != "" {
		w.WriteHTML(fmt.Sprintf(`<div class="eco-summary__row">Community pool: %s</div>`, html.EscapeString(d.CommunityPool)))
	}
	if d.CommunityTax != "" {
		w.WriteHTML(fmt.Sprintf(`<div class="eco-summary__row">Community tax: %s</div>`, html.EscapeString(d.CommunityTax)))
	}
	if total := economicsUnclaimedTotal(d); total != "" {
		w.WriteHTML(fmt.Sprintf(`<div class="eco-summary__row">Unclaimed: %s</div>`, html.EscapeString(total)))
	}
	if bal := FeeCollectorBalance(d); bal != "" {
		w.WriteHTML(fmt.Sprintf(`<div class="eco-summary__row">fee_collector: %s</div>`, html.EscapeString(bal)))
	}
}

func writeRewardsDistribution(w Writer, d model.Report) {
	w.Subsection("Distribution")
	writeRewardsDistributionModule(w, d)
	writeRewardsUnclaimedBalances(w, d)
	writeRewardsCommunityTax(w, d)
}

func writeRewardsDistributionModule(w Writer, d model.Report) {
	bal := distributionModuleBalance(d)
	addr := economicsDistributionModuleAddr(d)
	if bal == "" && addr == "" {
		return
	}
	html := economicsDistItemsHTML([]economicsDistItem{{
		param:   "distribution escrow",
		balance: orEcoDash(bal),
		addr:    addr,
		effect:  "x/distribution module escrow (often ~0 after BeginBlock payout)",
	}})
	w.WriteHTML(html)
}

func writeRewardsCommunityTax(w Writer, d model.Report) {
	if d.CommunityTax == "" && d.CommunityPool == "" {
		return
	}
	effect := ecoTaxEffect(d)
	if d.CommunityPool != "" {
		effect += " · pool " + d.CommunityPool
	}
	rowCls := ""
	if d.CommunityTaxZero {
		rowCls = "eco-domain__row--inactive"
	}
	w.WriteHTML(economicsDistItemsHTML([]economicsDistItem{{
		param:    "community tax",
		balance:  orEcoDash(d.CommunityTax),
		addr:     economicsDistributionModuleAddr(d),
		effect:   effect,
		rowClass: rowCls,
	}}))
}

func writeRewardsUnclaimedBalances(w Writer, d model.Report) {
	del := economicsUnclaimedDelegator(d)
	comm := economicsUnclaimedCommission(d)
	total := economicsUnclaimedTotal(d)
	if del == "" && comm == "" && total == "" {
		return
	}
	w.Subsection("Unclaimed rewards")
	addr := economicsDistributionModuleAddr(d)
	var items []economicsDistItem
	if del != "" {
		items = append(items, economicsDistItem{
			param:   "delegator share",
			balance: del,
			addr:    addr,
			effect:  "summed outstanding_rewards across validators",
		})
	}
	if comm != "" {
		items = append(items, economicsDistItem{
			param:   "validator commission",
			balance: comm,
			addr:    addr,
			effect:  "summed validator commission across validators",
		})
	}
	if total != "" {
		effect := "delegator share + validator commission"
		if _, suffix := splitOutstandingSuffix(d.TotalOutstanding); suffix != "" {
			effect += " · " + suffix
		}
		if check := economicsUnclaimedCheck(d); check != "" && check != "—" {
			effect += " · " + check
		}
		items = append(items, economicsDistItem{
			param:   "total outstanding",
			balance: total,
			addr:    addr,
			effect:  effect,
		})
	}
	w.WriteHTML(economicsDistItemsHTML(items))
}

func writeRewardsLedger(w Writer, d model.Report) {
	rows := economicsLedgerRows(d)
	if len(rows) == 0 {
		return
	}
	w.Subsection("Block reward ledger")
	if intro := rewardsLedgerIntro(d); intro != "" {
		w.Em(intro)
	}
	w.WriteHTML(economicsLedgerTableHTML(rows))
}

func rewardsLedgerIntro(d model.Report) string {
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
