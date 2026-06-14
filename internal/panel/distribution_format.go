package panel

import (
	"fmt"
	"html"
	"math"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

type unclaimedStack struct {
	Total            string
	Outstanding      string
	Commission       string
	OutstandingLabel string
	CommissionLabel  string
	OutstandingClaim string
	CommissionClaim  string
}

func unclaimedStackFromLocal(lv model.LocalValidator) unclaimedStack {
	return unclaimedStack{
		Total:            localUnclaimedTotal(lv),
		Outstanding:      lv.Outstanding,
		Commission:       lv.CommissionEarned,
		OutstandingLabel: "delegator share",
		CommissionLabel:  "your commission",
	}
}

func unclaimedStackFromLocalDetailed(lv model.LocalValidator) unclaimedStack {
	s := unclaimedStackFromLocal(lv)
	s.OutstandingClaim = "MsgWithdrawDelegatorReward"
	s.CommissionClaim = "MsgWithdrawValidatorCommission"
	return s
}

func unclaimedStackFromNetwork(d model.Report) unclaimedStack {
	return unclaimedStack{
		Total:            distributionUnclaimedTotal(d),
		Outstanding:      d.UnclaimedDelegator,
		Commission:       d.UnclaimedCommission,
		OutstandingLabel: "delegator share",
		CommissionLabel:  "operator commission",
	}
}

func unclaimedStackFromNetworkDetailed(d model.Report) unclaimedStack {
	s := unclaimedStackFromNetwork(d)
	s.OutstandingClaim = "MsgWithdrawDelegatorReward"
	s.CommissionClaim = "MsgWithdrawValidatorCommission"
	return s
}

func networkUnclaimedBreakdownHTML(d model.Report) string {
	return unclaimedStackHTML(unclaimedStackFromNetworkDetailed(d))
}

func (s unclaimedStack) empty() bool {
	return strings.TrimSpace(s.Total) == "" &&
		strings.TrimSpace(s.Outstanding) == "" &&
		strings.TrimSpace(s.Commission) == ""
}

func unclaimedStackHTML(s unclaimedStack) string {
	if s.empty() {
		return ""
	}
	total := strings.TrimSpace(s.Total)
	if total == "" {
		total = sumUnclaimedAmounts(s.Outstanding, s.Commission, model.Report{})
	}
	out := strings.TrimSpace(s.Outstanding)
	comm := strings.TrimSpace(s.Commission)
	outLabel := strings.TrimSpace(s.OutstandingLabel)
	if outLabel == "" {
		outLabel = "delegator share"
	}
	commLabel := strings.TrimSpace(s.CommissionLabel)
	if commLabel == "" {
		commLabel = "commission"
	}

	var b strings.Builder
	b.WriteString(`<div class="unclaimed-stack">`)
	fmt.Fprintf(&b,
		`<div class="unclaimed-stack__head"><span class="unclaimed-stack__head-label">unclaimed total</span>`+
			`<span class="unclaimed-stack__head-val">%s</span></div>`,
		html.EscapeString(total))

	if out != "" || comm != "" {
		b.WriteString(`<div class="unclaimed-stack__equation">`)
		if out != "" {
			b.WriteString(unclaimedStackPart(outLabel, out, s.OutstandingClaim))
		}
		if out != "" && comm != "" {
			b.WriteString(`<span class="unclaimed-stack__op" aria-hidden="true">+</span>`)
		}
		if comm != "" {
			b.WriteString(unclaimedStackPart(commLabel, comm, s.CommissionClaim))
		}
		b.WriteString(`</div>`)
	}
	b.WriteString(`</div>`)
	return b.String()
}

func unclaimedStackPart(label, amount, claim string) string {
	var b strings.Builder
	b.WriteString(`<div class="unclaimed-stack__part">`)
	fmt.Fprintf(&b, `<span class="unclaimed-stack__part-label">%s</span>`, html.EscapeString(label))
	fmt.Fprintf(&b, `<span class="unclaimed-stack__part-val">%s</span>`, html.EscapeString(amount))
	if claim != "" {
		fmt.Fprintf(&b, `<span class="unclaimed-stack__part-hint">%s</span>`, html.EscapeString(claim))
	}
	b.WriteString(`</div>`)
	return b.String()
}

func localUnclaimedBreakdownHTML(lv model.LocalValidator) string {
	return unclaimedStackHTML(unclaimedStackFromLocalDetailed(lv))
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
