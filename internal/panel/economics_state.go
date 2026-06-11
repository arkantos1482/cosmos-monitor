package panel

import (
	"fmt"
	"html"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

// EcoLedgerRow is one row in the block reward ledger with optional inactive styling.
type EcoLedgerRow struct {
	Cells    []string
	Inactive bool
	Warn     bool
}

func economicsHasRewardSource(d model.Report) bool {
	if d.PMTEnabled && d.PMTRate != "" && !d.PMTPoolEmpty {
		return true
	}
	if d.Inflation > 0 && d.InflationPerBlock != "" {
		return true
	}
	if v, ok := economicsParseAmount(d.LastBlockFees); ok && v > 0 {
		return true
	}
	return false
}

func economicsSummaryBadges(d model.Report) []summaryBadge {
	b := []summaryBadge{pmtPoolBadge(d)}
	if d.Inflation <= 0 {
		b = append(b, summaryBadge{"inflation off", "bad"})
	} else {
		b = append(b, summaryBadge{fmt.Sprintf("inflation %.2f%%", d.Inflation), "ok"})
	}
	if d.CommunityTaxZero {
		b = append(b, summaryBadge{"community tax 0%", "warn"})
	} else if d.CommunityTax != "" {
		b = append(b, summaryBadge{"community tax " + d.CommunityTax, "ok"})
	}
	if !economicsHasRewardSource(d) {
		b = append(b, summaryBadge{"no reward inflow", "bad"})
	}
	return b
}

type ecoFlag struct {
	param    string
	value    string
	effect   string
	inactive bool
	warn     bool
}

func economicsFlags(d model.Report) []ecoFlag {
	flags := []ecoFlag{
		{
			param:  "pmtrewards.enabled",
			value:  boolStr(d.PMTEnabled),
			effect: ecoPMTEffect(d),
			inactive: !d.PMTEnabled,
		},
		{
			param: "pmtrewards.pool_balance",
			value: orEcoDash(d.PMTBalance),
			effect: ecoPoolEffect(d),
			inactive: !d.PMTEnabled,
			warn:     d.PMTEnabled && d.PMTPoolEmpty,
		},
		{
			param:    "pmtrewards.reward_per_block",
			value:    orEcoDash(d.PMTRate),
			effect:   ecoPMTRateEffect(d),
			inactive: !d.PMTEnabled || d.PMTRate == "",
			warn:     d.PMTEnabled && !d.PMTPoolEmpty && d.PMTRate == "",
		},
		{
			param:    "mint.inflation",
			value:    fmt.Sprintf("%.2f%%", d.Inflation),
			effect:   ecoInflationEffect(d),
			inactive: d.Inflation <= 0,
		},
		{
			param:    "mint.annual_provisions",
			value:    orEcoDash(d.AnnualProvisions),
			effect:   ecoAnnualProvEffect(d),
			inactive: d.Inflation <= 0 || d.AnnualProvisions == "",
		},
		{
			param:    "distribution.community_tax",
			value:    orEcoDash(d.CommunityTax),
			effect:   ecoTaxEffect(d),
			inactive: d.CommunityTaxZero,
			warn:     !d.CommunityTaxZero && !economicsHasRewardSource(d),
		},
		{
			param:    "tx_fees.last_block",
			value:    orEcoDash(trimFeeNote(d.LastBlockFees)),
			effect:   ecoTxFeesEffect(d),
			inactive: d.LastBlockFees == "",
		},
		{
			param:  "reward_in.total_per_block",
			value:  orEcoDash(RewardInPerBlockTotal(d)),
			effect: ecoRewardInEffect(d),
			inactive: !economicsHasRewardSource(d),
		},
		{
			param:  "staking.bonded_ratio",
			value:  fmt.Sprintf("%.2f%% (goal %.0f%%)", d.BondedPct, d.GoalBonded),
			effect: "inflation adjusts toward goal bonded",
		},
	}
	return flags
}

func ecoPMTEffect(d model.Report) string {
	if !d.PMTEnabled {
		return "module disabled — no PMT block rewards"
	}
	if d.PMTPoolEmpty {
		return "enabled but pool empty — nothing to distribute"
	}
	if d.PMTRate == "" {
		return "enabled — rate unknown"
	}
	return "distributing to fee_collector each block"
}

func ecoPoolEffect(d model.Report) string {
	if !d.PMTEnabled {
		return "—"
	}
	if d.PMTPoolEmpty {
		return "empty — rewards stop"
	}
	if d.PMTRunway != "" {
		return "funded · " + d.PMTRunway
	}
	return "funded"
}

func ecoPMTRateEffect(d model.Report) string {
	if !d.PMTEnabled {
		return "inactive"
	}
	if d.PMTRate == "" {
		return "no rate configured"
	}
	if d.PMTPoolEmpty {
		return "rate set but pool cannot pay"
	}
	return "active per-block emission"
}

func ecoInflationEffect(d model.Report) string {
	if d.Inflation <= 0 {
		return "inactive — x/mint not minting"
	}
	if d.InflationPerBlock != "" {
		return "active — mints each block"
	}
	return "rate set — per-block amount unavailable"
}

func ecoAnnualProvEffect(d model.Report) string {
	if d.Inflation <= 0 {
		return "inactive"
	}
	if d.AnnualProvisions == "" {
		return "—"
	}
	return "absolute mint budget / year"
}

func ecoTaxEffect(d model.Report) string {
	if d.CommunityTaxZero {
		return "0% — community pool gets no cut"
	}
	if !economicsHasRewardSource(d) {
		return "tax configured but no rewards flowing"
	}
	return "skims % of block rewards → community pool"
}

func ecoTxFeesEffect(d model.Report) string {
	if d.LastBlockFees == "" {
		return "no parent-block fee data"
	}
	v, ok := economicsParseAmount(d.LastBlockFees)
	if !ok || v <= 0 {
		return "no tx fee income this block"
	}
	return "gas × base fee → fee_collector"
}

func ecoRewardInEffect(d model.Report) string {
	if !economicsHasRewardSource(d) {
		return "nothing entering fee_collector"
	}
	return "sum of active reward sources"
}

func trimFeeNote(s string) string {
	if i := strings.Index(s, "  _"); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return s
}

func orEcoDash(s string) string {
	if s == "" || s == "—" {
		return "—"
	}
	return s
}

func economicsFlagsTableHTML(flags []ecoFlag) string {
	var b strings.Builder
	b.WriteString(`<div class="eco-flags" id="eco-flags">`)
	b.WriteString(`<table class="eco-flags__table"><thead><tr>`)
	for _, h := range []string{"Parameter", "Value", "Effect"} {
		fmt.Fprintf(&b, `<th>%s</th>`, html.EscapeString(h))
	}
	b.WriteString(`</tr></thead><tbody>`)
	for _, f := range flags {
		trCls := ""
		switch {
		case f.inactive:
			trCls = ` class="eco-flags__row--inactive"`
		case f.warn:
			trCls = ` class="eco-flags__row--warn"`
		}
		fmt.Fprintf(&b, `<tr%s>`, trCls)
		fmt.Fprintf(&b, `<td class="eco-flags__param"><code>%s</code></td>`, html.EscapeString(f.param))
		b.WriteString(`<td class="eco-flags__val">`)
		b.WriteString(ecoFlagValueHTML(f.value))
		b.WriteString(`</td><td class="eco-flags__effect">`)
		b.WriteString(ecoFlagEffectHTML(f.effect, f.inactive, f.warn))
		b.WriteString(`</td></tr>`)
	}
	b.WriteString(`</tbody></table></div>`)
	return b.String()
}

func ecoFlagValueHTML(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "true":
		return `<span class="badge badge--ok">true</span>`
	case "false":
		return `<span class="badge badge--bad">false</span>`
	default:
		return softWrapHTML(v)
	}
}

func ecoFlagEffectHTML(effect string, inactive, warn bool) string {
	cls := ""
	switch {
	case inactive:
		cls = ` class="eco-flags__effect--inactive"`
	case warn:
		cls = ` class="eco-flags__effect--warn"`
	}
	return fmt.Sprintf(`<span%s>%s</span>`, cls, inlineHTML(effect))
}

func economicsLedgerTableHTML(rows []EcoLedgerRow) string {
	headers := []string{"Step", "Where", "In this block", "Balance now", "Check"}
	var b strings.Builder
	b.WriteString(`<div class="table-scroll"><table class="data-table data-table--ledger">`)
	b.WriteString(`<thead><tr>`)
	for i, h := range headers {
		thCls := ""
		if i > 0 {
			thCls = ` class="data-table__num"`
		}
		fmt.Fprintf(&b, `<th%s>%s</th>`, thCls, html.EscapeString(h))
	}
	b.WriteString(`</tr></thead><tbody>`)
	for _, row := range rows {
		trCls := ""
		switch {
		case row.Inactive:
			trCls = ` class="eco-row--inactive"`
		case row.Warn:
			trCls = ` class="eco-row--warn"`
		}
		fmt.Fprintf(&b, `<tr%s>`, trCls)
		for i, cell := range row.Cells {
			if i == 0 {
				step := html.EscapeString(strings.TrimSpace(cell))
				fmt.Fprintf(&b, `<td class="data-table__step" data-step="%s">%s</td>`, step, step)
				continue
			}
			tdCls := ""
			if i > 0 {
				tdCls = ` class="data-table__num"`
			}
			fmt.Fprintf(&b, `<td%s>%s</td>`, tdCls, formatValue(cell))
		}
		b.WriteString(`</tr>`)
	}
	b.WriteString(`</tbody></table></div>`)
	return b.String()
}
