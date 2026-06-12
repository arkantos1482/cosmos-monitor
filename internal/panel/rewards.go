package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeRewards(w Writer, d model.Report) {
	w.Section("4. REWARDS")
	writeEmbeddedSectionIntro(w, "Block emission from the PMT rewards pool (x/pmtrewards) and SDK inflation (x/mint); both land in fee_collector before x/distribution splits.")
	writeRewardsSummary(w, d, SummaryEmbedded)

	if d.Local.IsValidator {
		w.Subsection("This validator")
		writeRewardsLocal(w, d)
	}

	w.Subsection("Network-wide")
	w.WriteHTML(rewardsDomainCardsHTML(d))

	w.Subsection("Emission sources")
	w.Hint("Per-block amounts from module queries; x/pmtrewards transfers run inside x/mint BeginBlock via evmd's custom MintFn.")
	w.WriteHTML(rewardsEmissionTableHTML(d))

	w.Hint(rewardsSourcesHint())
	w.BlankLine()
}

func writeRewardsSummary(w Writer, d model.Report, mode SummaryMode) {
	summaryWrapStart(w, mode, "rewards")
	writeRewardsSummaryBody(w, d)
	summaryWrapEnd(w, mode)
}

func writeRewardsSummaryBody(w Writer, d model.Report) {
	w.WriteHTML(`<div class="rewards-summary">`)
	w.WriteHTML(`<div class="rewards-summary__kpis">`)

	label, val, tone := rewardsSummaryPMT(d)
	writeRewardsSummaryKPI(w, label, val, tone)

	label, val, tone = rewardsSummaryInflation(d)
	writeRewardsSummaryKPI(w, label, val, tone)

	if emit := rewardsEmissionPerBlock(d); emit != "—" {
		writeRewardsSummaryKPI(w, "combined emission", emit, "ok")
	}
	if d.PMTEnabled && !d.PMTPoolEmpty && d.PMTBalance != "" {
		writeRewardsSummaryKPI(w, "reward pool", d.PMTBalance, "")
		if d.PMTRunway != "" {
			writeRewardsSummaryKPI(w, "pool runway", d.PMTRunway, "")
		}
	}

	w.WriteHTML(`</div></div>`)
}

func writeRewardsSummaryKPI(w Writer, label, value, tone string) {
	if value == "" {
		return
	}
	valCls := "rewards-summary__kpi-val"
	if tone != "" {
		valCls += " rewards-summary__kpi-val--" + tone
	}
	w.WriteHTML(fmt.Sprintf(
		`<div class="rewards-summary__kpi"><span class="rewards-summary__kpi-label">%s</span>`+
			`<span class="%s">%s</span></div>`,
		html.EscapeString(label), valCls, html.EscapeString(value)))
}

func writeRewardsLocal(w Writer, d model.Report) {
	lv := d.Local
	w.Hint("`per-block` → estimated from combined emission × VP% × commission (x/distribution applies VP weighting on chain).")
	if op, del, _, ok := localValidatorPerBlockRewards(d); ok {
		w.Row("per-block commission", op+fmt.Sprintf("  (%.2f%% VP · %.1f%% commission)", lv.VPPercent, lv.Commission))
		w.Row("per-block delegators", del)
	} else if emit := rewardsEmissionPerBlock(d); emit != "—" {
		w.Row("per-block emission", "—  _(no VP or no active emission)_")
	}
}

func rewardsSourcesHint() string {
	return "`enabled`, `reward_per_block`, `pool_address` → REST GET /cosmos/evm/pmtrewards/v1/params; " +
		"`pool balance` → REST GET /cosmos/bank/v1beta1/balances/{pool_address}; " +
		"`inflation`, `annual_provisions` → REST GET /cosmos/mint/v1beta1/inflation, /cosmos/mint/v1beta1/annual-provisions; " +
		"`goal_bonded`, `blocks_per_year` → REST GET /cosmos/mint/v1beta1/params; " +
		"`bonded %` → REST GET /cosmos/staking/v1beta1/pool; " +
		"`per-block estimates` → derived (annual_provisions ÷ blocks_per_year, reward_per_block, block interval)."
}
