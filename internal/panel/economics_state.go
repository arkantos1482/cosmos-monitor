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

func economicsDistributionModuleAddr(d model.Report) string {
	return moduleAccountDisplayAddress(d, "distribution")
}

func stakingCardHTML(d model.Report, compact bool) string {
	var b strings.Builder
	b.WriteString(`<div class="eco-domain eco-domain--staking">`)
	b.WriteString(`<h3 class="eco-domain__title">Staking <span class="eco-domain__subtitle">x/staking</span></h3>`)
	b.WriteString(`<div class="eco-domain__rows">`)

	ecoDomainRow(&b, "", "bonded", fmt.Sprintf("%.2f%%", d.BondedPct), "share of supply staked")

	if !compact {
		if d.BondedAmt != "" {
			ecoDomainRow(&b, "", "bonded amt", d.BondedAmt, "actively securing chain")
		}
		if d.NotBonded != "" {
			ecoDomainRow(&b, "", "not bonded", d.NotBonded, "unbonded or unbonding")
		}
		if d.TotalSupply != "" {
			ecoDomainRow(&b, "", "total supply", d.TotalSupply, "bank module supply")
		}
	}

	writeStakingModuleAccountRows(&b, d)
	writeStakingParamRows(&b, d)

	b.WriteString(`</div></div>`)
	return b.String()
}

func slashingCardHTML(d model.Report, _ bool) string {
	var b strings.Builder
	b.WriteString(`<div class="eco-domain eco-domain--slashing">`)
	b.WriteString(`<h3 class="eco-domain__title">Slashing <span class="eco-domain__subtitle">x/slashing</span></h3>`)
	b.WriteString(`<div class="eco-domain__rows">`)

	writeSlashingParamRows(&b, d)
	writeSlashingPenaltyMatrix(&b, d)

	b.WriteString(`</div></div>`)
	return b.String()
}

func writeStakingModuleAccountRows(b *strings.Builder, d model.Report) {
	writeModuleAccountRow(b, d, "bonded_tokens_pool", "staked tokens escrow")
	writeModuleAccountRow(b, d, "not_bonded_tokens_pool", "unbonded / unbonding escrow")
}

func writeStakingParamRows(b *strings.Builder, d model.Report) {
	ecoDomainDivider(b)
	if d.BondDenom != "" {
		ecoDomainRow(b, "", "bond denom", d.BondDenom, "staking unit of account")
	}
	if d.UnbondingTime != "" {
		ecoDomainRow(b, "", "unbonding time", d.UnbondingTime, "time locked after unstaking")
	}
	if d.MaxValidators > 0 {
		ecoDomainRow(b, "", "max validators", fmt.Sprintf("%d", d.MaxValidators), "active set cap")
	}
}

func writeSlashingParamRows(b *strings.Builder, d model.Report) {
	if d.SlashWindow != "" && d.SlashWindow != "0" {
		ecoDomainRow(b, "", "signed blocks window", d.SlashWindow+" blocks", "downtime tracking window")
	}
	if d.MinSigned > 0 {
		ecoDomainRow(b, "", "min signed", fmt.Sprintf("%.1f%%", d.MinSigned), "miss more → downtime slash risk")
	}
	if d.SlashDowntime != "" {
		dtCls := ""
		dtEffect := "fraction slashed for downtime"
		if d.SlashDTInactive {
			dtCls = ` class="eco-domain__row--warn"`
			dtEffect = "inactive — downtime slash disabled"
		}
		ecoDomainRow(b, dtCls, "slash / downtime", d.SlashDowntime, dtEffect)
	}
	if d.SlashDS != "" {
		dsCls := ""
		dsEffect := "fraction slashed for double-sign"
		if d.SlashDSInactive {
			dsCls = ` class="eco-domain__row--warn"`
			dsEffect = "inactive — double-sign slash disabled"
		}
		ecoDomainRow(b, dsCls, "slash / double-sign", d.SlashDS, dsEffect)
	}
}

func economicsLedgerTableHTML(rows []EcoLedgerRow) string {
	headers := []string{"Step", "Where", "In this block", "Balance now", "Check"}
	var b strings.Builder
	b.WriteString(`<div class="table-scroll"><table class="data-table data-table--ledger">`)
	b.WriteString(`<thead><tr>`)
	for _, h := range headers {
		fmt.Fprintf(&b, `<th%s>%s</th>`, tableColumnClass(h), html.EscapeString(h))
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
			tdCls := tableColumnClass(headers[i])
			fmt.Fprintf(&b, `<td%s>%s</td>`, tdCls, formatValue(cell))
		}
		b.WriteString(`</tr>`)
	}
	b.WriteString(`</tbody></table></div>`)
	return b.String()
}
