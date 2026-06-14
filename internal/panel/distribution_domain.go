package panel

import (
	"fmt"
	"html"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func distributionCardHTML(d model.Report) string {
	var b strings.Builder
	ecoDomainCardOpen(&b, "eco-domain--distribution", "Distribution", "x/distribution")

	ecoDomainDividerDist(&b, "Unclaimed rewards")
	writeDistributionUnclaimedRows(&b, d)

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

	ecoDomainDividerDist(&b, "Distribution escrow")
	writeModuleAccountRow(&b, d, "distribution",
		"bank balance — holds all unclaimed delegator and operator rewards until withdrawn")
	writeDistributionEscrowReconcileRow(&b, d)

	withdrawEffect := "delegators may set a custom withdraw address"
	if !d.WithdrawAddrEnabled {
		withdrawEffect = "withdrawals go to delegator account address"
	}
	ecoDomainRow(&b, "", "withdraw_addr_enabled", boolStr(d.WithdrawAddrEnabled), withdrawEffect)

	ecoDomainCardClose(&b)
	return b.String()
}

func writeDistributionEscrowReconcileRow(b *strings.Builder, d model.Report) {
	effect, warn := distributionEscrowReconcile(d)
	if effect == "" {
		return
	}
	cls := ""
	if warn {
		cls = ` class="eco-domain__row--warn"`
	}
	state := distributionUnclaimedTotal(d)
	bank := distributionModuleBalance(d)
	val := orEcoDash(bank)
	if state != "" && bank != "" && state != bank {
		val = fmt.Sprintf("%s → state tracks %s", bank, state)
	}
	ecoDomainRow(b, cls, "escrow check", val, effect)
}

func ecoDomainDividerDist(b *strings.Builder, title string) {
	fmt.Fprintf(b, `<div class="eco-domain__divider">%s</div>`, html.EscapeString(title))
}

func writeDistributionUnclaimedRows(b *strings.Builder, d model.Report) {
	del := d.UnclaimedDelegator
	comm := d.UnclaimedCommission
	total := distributionUnclaimedTotal(d)
	if del == "" && comm == "" {
		ecoDomainRow(b, ` class="eco-domain__row--inactive"`, "total unclaimed", "—",
			"no rewards waiting to be withdrawn network-wide")
		return
	}
	if total != "" {
		ecoDomainRow(b, "", "total unclaimed", total,
			"delegator share + operator commission — all escrowed until someone claims")
	}
	if del != "" {
		ecoDomainRow(b, "", "for delegators", del,
			"sum of per-validator delegator shares — MsgWithdrawDelegatorReward")
	}
	if comm != "" {
		ecoDomainRow(b, "", "for operators", comm,
			"sum of validator commission balances — MsgWithdrawValidatorCommission")
	}
}

func distributionUnclaimedTotal(d model.Report) string {
	return economicsUnclaimedTotal(d)
}

func distributionDomainCardsHTML(d model.Report) string {
	return ecoDomainsWrap(distributionCardHTML(d))
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
