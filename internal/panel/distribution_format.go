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
	b.WriteString(`<div class="unclaimed-breakdown">`)
	if total != "" {
		fmt.Fprintf(&b, `<div class="unclaimed-breakdown__total"><span class="unclaimed-breakdown__amount">%s</span>`+
			`<span class="unclaimed-breakdown__hint">total not yet withdrawn</span></div>`,
			html.EscapeString(total))
	}
	b.WriteString(`<ul class="unclaimed-breakdown__parts">`)
	if del != "" {
		b.WriteString(unclaimedPartLI("delegator share", del, "claim with MsgWithdrawDelegatorReward"))
	}
	if comm != "" {
		b.WriteString(unclaimedPartLI("your commission", comm, "claim with MsgWithdrawValidatorCommission"))
	}
	b.WriteString(`</ul></div>`)
	return b.String()
}

func unclaimedPartLI(label, amount, claim string) string {
	return fmt.Sprintf(
		`<li><span class="unclaimed-breakdown__part-label">%s</span>`+
			`<span class="unclaimed-breakdown__part-val">%s</span>`+
			`<span class="unclaimed-breakdown__part-claim">%s</span></li>`,
		html.EscapeString(label), html.EscapeString(amount), html.EscapeString(claim))
}

func localUnclaimedTotal(lv model.LocalValidator) string {
	del := strings.TrimSpace(lv.Outstanding)
	comm := strings.TrimSpace(lv.CommissionEarned)
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
