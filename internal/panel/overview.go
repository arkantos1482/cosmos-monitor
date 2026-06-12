package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeOverview(w Writer, d model.Report) {
	w.WriteHTML(`<div class="dash-overview">`)
	w.WriteHTML(`<p class="dash-overview__lead">Live snapshot — scroll for all sections. Refreshes every 5s.</p>`)

	writeOverviewGroup(w, d, NavScopeRuntime, []struct {
		slug string
		fn   func(Writer, model.Report, SummaryMode)
	}{
		{"infra", writeInfraSummary},
		{"evm", writeEVMSummary},
	})
	writeOverviewGroup(w, d, NavScopeValidator, []struct {
		slug string
		fn   func(Writer, model.Report, SummaryMode)
	}{
		{"node", writeNodeSummary},
	})
	writeOverviewGroup(w, d, NavScopeEconomics, []struct {
		slug string
		fn   func(Writer, model.Report, SummaryMode)
	}{
		{"staking", writeStakingSummary},
		{"slashing", writeSlashingSummary},
		{"feemarket", writeFeemarketSummary},
		{"rewards", writeRewardsSummary},
		{"distribution", writeDistributionSummary},
	})
	writeOverviewGroup(w, d, NavScopeGovernance, []struct {
		slug string
		fn   func(Writer, model.Report, SummaryMode)
	}{
		{"governance", writeGovernanceSummary},
	})

	w.WriteHTML(`</div>`)
}

func writeOverviewGroup(w Writer, d model.Report, scope NavScope, items []struct {
	slug string
	fn   func(Writer, model.Report, SummaryMode)
}) {
	if NavScopeLabel(scope) == "" {
		return
	}
	w.WriteHTML(fmt.Sprintf(`<div class="dash-overview__group dash-overview__group--%s">`, html.EscapeString(string(scope))))
	w.WriteHTML(`<div class="dash-overview__stack">`)
	for _, item := range items {
		item.fn(w, d, SummaryOverviewClickable)
	}
	w.WriteHTML(`</div></div>`)
}
