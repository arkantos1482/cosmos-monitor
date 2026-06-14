package panel

import (
	"fmt"
	"html"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func distributionParamsCardHTML(d model.Report) string {
	var b strings.Builder
	ecoDomainCardOpen(&b, "eco-domain--distribution", "Distribution", "x/distribution")

	ecoDomainDividerDist(&b, "Community treasury")
	poolCls := ""
	if d.CommunityPool == "" {
		poolCls = ` class="eco-domain__row--inactive"`
	}
	ecoDomainRow(&b, poolCls, "community pool", orEcoDash(d.CommunityPool),
		"governance-controlled reserve — funded by community tax and direct deposits")
	taxCls := ""
	if d.CommunityTaxZero {
		taxCls = ` class="eco-domain__row--inactive"`
	}
	ecoDomainRow(&b, taxCls, "community_tax", orEcoDash(d.CommunityTax),
		"fraction of block rewards diverted to community pool before validator split")

	ecoDomainDividerDist(&b, "Withdraw policy")
	withdrawEffect := "delegators may set a custom withdraw address"
	if !d.WithdrawAddrEnabled {
		withdrawEffect = "withdrawals go to delegator account address"
	}
	ecoDomainRow(&b, "", "withdraw_addr_enabled", boolStr(d.WithdrawAddrEnabled), withdrawEffect)

	ecoDomainCardClose(&b)
	return ecoDomainsWrap(b.String())
}

func distributionEscrowBlockHTML(d model.Report) string {
	bal := moduleAccountBalance(d, "distribution")
	addr := moduleAccountDisplayAddress(d, "distribution")
	effect, warn := distributionEscrowReconcile(d)
	if bal == "" && addr == "" && effect == "" {
		return ""
	}

	var b strings.Builder
	b.WriteString(`<div class="dist-escrow">`)
	b.WriteString(`<div class="dist-escrow__label">distribution escrow</div>`)
	if bal != "" || addr != "" {
		fmt.Fprintf(&b, `<div class="dist-escrow__acct">%s</div>`, ecoBalanceAddrHTML(orEcoDash(bal), addr))
	}
	if effect != "" {
		cls := "dist-escrow__note"
		if warn {
			cls += " dist-escrow__note--warn"
		}
		state := distributionUnclaimedTotal(d)
		bank := distributionModuleBalance(d)
		if state != "" && bank != "" && state != bank {
			fmt.Fprintf(&b, `<div class="%s">%s — bank %s, state tracks %s</div>`,
				cls, html.EscapeString(effect), html.EscapeString(bank), html.EscapeString(state))
		} else {
			fmt.Fprintf(&b, `<div class="%s">%s</div>`, cls, html.EscapeString(effect))
		}
	}
	b.WriteString(`</div>`)
	return b.String()
}

func ecoDomainDividerDist(b *strings.Builder, title string) {
	fmt.Fprintf(b, `<div class="eco-domain__divider">%s</div>`, html.EscapeString(title))
}

func distributionUnclaimedTotal(d model.Report) string {
	return economicsUnclaimedTotal(d)
}

func writeDistributionValidatorTable(w Writer, d model.Report) {
	if len(d.Validators) == 0 {
		return
	}
	rows := make([][]string, 0, len(d.Validators))
	for _, v := range d.Validators {
		total := validatorUnclaimedTotal(v)
		if total == "" {
			total = "—"
		}
		comm := v.CommissionEarned
		if comm == "" {
			comm = "—"
		}
		del := v.Outstanding
		if del == "" {
			del = "—"
		}
		rows = append(rows, []string{
			report.Truncate(v.Moniker, 14),
			identityCell(v.Operator),
			total,
			comm,
			del,
			fmt.Sprintf("%.1f%%", v.CommissionFloat),
		})
	}
	writeValidatorSetTable(w,
		[]string{"moniker", "operator", "total", "commission", "outstanding share", "comm. rate"},
		rows, d.Validators)
}
