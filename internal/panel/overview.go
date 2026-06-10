package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeOverview(w Writer, d model.Report) {
	w.WriteHTML(`<div class="dash-overview">`)
	w.WriteHTML(`<p class="dash-overview__lead">Live snapshot — scroll for all sections. Refreshes every 5s.</p>`)

	writeOverviewGroup(w, d, NavScopeChain, []struct {
		slug string
		fn   func(Writer, model.Report, SummaryMode)
	}{
		{"validators", writeValidatorsSummary},
		{"economics", writeEconomicsSummary},
		{"feemarket", writeFeemarketSummary},
		{"governance", writeGovernanceSummary},
	})
	writeOverviewGroup(w, d, NavScopeNode, []struct {
		slug string
		fn   func(Writer, model.Report, SummaryMode)
	}{
		{"infra", writeInfraSummary},
		{"node", writeNodeSummary},
		{"evm", writeEVMSummary},
	})

	w.WriteHTML(`</div>`)
}

func writeOverviewGroup(w Writer, d model.Report, scope NavScope, items []struct {
	slug string
	fn   func(Writer, model.Report, SummaryMode)
}) {
	label := NavScopeLabel(scope)
	if label == "" {
		return
	}
	w.WriteHTML(fmt.Sprintf(`<div class="dash-overview__group dash-overview__group--%s">`, html.EscapeString(string(scope))))
	w.WriteHTML(fmt.Sprintf(`<h2 class="dash-overview__group-title">%s</h2>`, html.EscapeString(label)))
	w.WriteHTML(`<div class="dash-overview__stack">`)
	for _, item := range items {
		item.fn(w, d, SummaryOverviewClickable)
	}
	w.WriteHTML(`</div></div>`)
}
