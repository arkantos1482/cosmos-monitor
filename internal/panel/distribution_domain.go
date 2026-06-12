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

	taxCls := ""
	if d.CommunityTaxZero {
		taxCls = ` class="eco-domain__row--inactive"`
	}
	ecoDomainRow(&b, taxCls, "community_tax", orEcoDash(d.CommunityTax),
		"fraction of block rewards sent to community pool before validator split")

	withdrawEffect := "delegators may set a custom withdraw address"
	if !d.WithdrawAddrEnabled {
		withdrawEffect = "withdrawals go to delegator account address"
	}
	ecoDomainRow(&b, "", "withdraw_addr_enabled", boolStr(d.WithdrawAddrEnabled), withdrawEffect)

	poolCls := ""
	if d.CommunityPool == "" {
		poolCls = ` class="eco-domain__row--inactive"`
	}
	ecoDomainRow(&b, poolCls, "community pool", orEcoDash(d.CommunityPool),
		"accumulated tax and direct funding — spendable via governance")

	ecoDomainDividerDist(&b, "Module accounts")
	writeModuleAccountRow(&b, d, "fee_collector",
		"receives fees and minted rewards each block; cleared in BeginBlock")
	writeModuleAccountRow(&b, d, "distribution",
		"module escrow — typically ~0 after payout")

	ecoDomainDividerDist(&b, "Outstanding (network)")
	writeDistributionUnclaimedRows(&b, d)

	ecoDomainCardClose(&b)
	return b.String()
}

func ecoDomainDividerDist(b *strings.Builder, title string) {
	fmt.Fprintf(b, `<div class="eco-domain__divider">%s</div>`, html.EscapeString(title))
}

func writeDistributionUnclaimedRows(b *strings.Builder, d model.Report) {
	del := d.UnclaimedDelegator
	comm := d.UnclaimedCommission
	if del == "" && comm == "" {
		ecoDomainRow(b, ` class="eco-domain__row--inactive"`, "outstanding rewards", "—",
			"no unclaimed delegator rewards across validators")
		return
	}
	if del != "" {
		ecoDomainRow(b, "", "delegator rewards", del,
			"summed outstanding_rewards — claim with MsgWithdrawDelegatorReward")
	}
	if comm != "" {
		ecoDomainRow(b, "", "validator commission", comm,
			"summed accumulated commission — claim with MsgWithdrawValidatorCommission")
	}
	total := distributionUnclaimedTotal(d)
	if total != "" && del != "" && comm != "" {
		ecoDomainRow(b, "", "total outstanding", total, "delegator share + validator commission")
	}
}

func distributionUnclaimedTotal(d model.Report) string {
	del := d.UnclaimedDelegator
	comm := d.UnclaimedCommission
	if del != "" && comm != "" {
		return economicsUnclaimedTotal(d)
	}
	if del != "" {
		return del
	}
	return comm
}

func distributionFeeCollectorStatus(d model.Report) string {
	check := economicsFeeCollectorCheck(d)
	if check == "cleared" {
		return "cleared"
	}
	if check == "not cleared?" {
		return "pending"
	}
	bal := FeeCollectorBalance(d)
	if bal == "" {
		return "—"
	}
	return bal
}

func distributionFeeCollectorTone(d model.Report) string {
	switch economicsFeeCollectorCheck(d) {
	case "cleared":
		return "ok"
	case "not cleared?":
		return "warn"
	default:
		return ""
	}
}

func distributionDomainCardsHTML(d model.Report) string {
	return ecoDomainsWrap(distributionCardHTML(d))
}

func writeDistributionValidatorTable(w Writer, d model.Report) {
	if len(d.Validators) == 0 {
		return
	}
	w.Hint("`outstanding`, `commission` → REST GET /cosmos/distribution/v1beta1/validators/{valoper}/outstanding_rewards, …/commission.")
	rows := make([][]string, 0, len(d.Validators))
	for _, v := range d.Validators {
		out := v.Outstanding
		if out == "" {
			out = "—"
		}
		comm := v.CommissionEarned
		if comm == "" {
			comm = "—"
		}
		rows = append(rows, []string{
			report.Truncate(v.Moniker, 14),
			identityCell(v.Operator),
			out,
			comm,
			fmt.Sprintf("%.1f%%", v.CommissionFloat),
		})
	}
	writeValidatorSetTable(w,
		[]string{"moniker", "operator", "outstanding", "commission", "rate"},
		rows, d.Validators)
}
