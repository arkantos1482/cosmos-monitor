package panel

import (
	"fmt"
	"html"
	"math"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func localUnclaimedBreakdownHTML(lv model.LocalValidator) string {
	del := strings.TrimSpace(lv.Outstanding)
	comm := strings.TrimSpace(lv.CommissionEarned)
	if del == "" && comm == "" {
		return ""
	}
	total := localUnclaimedTotal(lv)
	var b strings.Builder
	b.WriteString(`<div class="unclaimed-breakdown unclaimed-breakdown--horizontal">`)
	if total != "" {
		b.WriteString(unclaimedBreakdownCard("total", total, "not yet withdrawn", true))
	}
	if del != "" {
		b.WriteString(unclaimedBreakdownCard("delegator share", del, "MsgWithdrawDelegatorReward", false))
	}
	if comm != "" {
		b.WriteString(unclaimedBreakdownCard("your commission", comm, "MsgWithdrawValidatorCommission", false))
	}
	b.WriteString(`</div>`)
	return b.String()
}

func unclaimedBreakdownCard(label, amount, hint string, hero bool) string {
	cls := "unclaimed-breakdown__card"
	if hero {
		cls += " unclaimed-breakdown__card--hero"
	}
	return fmt.Sprintf(
		`<div class="%s"><span class="unclaimed-breakdown__card-label">%s</span>`+
			`<span class="unclaimed-breakdown__card-val">%s</span>`+
			`<span class="unclaimed-breakdown__card-hint">%s</span></div>`,
		cls, html.EscapeString(label), html.EscapeString(amount), html.EscapeString(hint))
}

func validatorUnclaimedTotal(v model.Validator) string {
	return sumUnclaimedAmounts(v.Outstanding, v.CommissionEarned, model.Report{})
}

func localUnclaimedTotal(lv model.LocalValidator) string {
	return sumUnclaimedAmounts(lv.Outstanding, lv.CommissionEarned, model.Report{})
}

func sumUnclaimedAmounts(outstanding, commission string, d model.Report) string {
	del := strings.TrimSpace(outstanding)
	comm := strings.TrimSpace(commission)
	if del != "" && comm != "" {
		delF, ok1 := economicsParseAmount(del)
		commF, ok2 := economicsParseAmount(comm)
		if ok1 && ok2 {
			unit := amountUnit(del)
			return fetch.FormatAmountUnit(delF+commF, unit)
		}
	}
	if del != "" {
		return del
	}
	return comm
}

func amountUnit(s string) string {
	parts := strings.Fields(s)
	if len(parts) >= 2 {
		return parts[1]
	}
	return economicsDenom(model.Report{})
}

func distributionEscrowReconcile(d model.Report) (effect string, warn bool) {
	bank := distributionModuleBalance(d)
	state := distributionUnclaimedTotal(d)
	if bank == "" || state == "" {
		return "", false
	}
	bankF, ok1 := economicsParseAmount(bank)
	stateF, ok2 := economicsParseAmount(state)
	if !ok1 || !ok2 {
		return "", false
	}
	diff := stateF - bankF
	if math.Abs(diff) < 1e-9 {
		return "distribution escrow bank balance matches tracked unclaimed total", false
	}
	unit := amountUnit(state)
	gap := fetch.FormatAmountUnit(math.Abs(diff), unit)
	if diff > 0 {
		return fmt.Sprintf(
			"bank is %s short of state total — delegator and operator shares both escrow in this module until withdrawn; gap often clears at the next BeginBlock",
			gap), true
	}
	return fmt.Sprintf("bank is %s above tracked unclaimed total", gap), true
}
