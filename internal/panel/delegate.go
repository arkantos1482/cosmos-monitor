package panel

import (
	"fmt"
	"html"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

const (
	StakingPrecompile = "0x0000000000000000000000000000000000000800"
	DefaultPublicEVM  = "https://node1.pmtchain.com"
	EVMChainID        = 290290
)

// OperatorValidator is a contractor-run validator offered as a Delegate dropdown preset.
type OperatorValidator struct {
	Label   string
	Valoper string
}

// OperatorValidators are the four live PMT operators. Delegators may also paste any valoper.
var OperatorValidators = []OperatorValidator{
	{Label: "node1", Valoper: "cosmosvaloper1akkvh0ahmve830rj4mhkdnqs49kzw23cl98zp4"},
	{Label: "node2", Valoper: "cosmosvaloper1r2dqta25pxj8av9grlfxvnfje006papu0tjk0a"},
	{Label: "node3", Valoper: "cosmosvaloper15hr4x4rfj0y82puk74xegugn5s5clphzcfej3e"},
	{Label: "node4", Valoper: "cosmosvaloper1vmr9wxpldngnh0tvpr8h2pk2aycts3v7z8pdxh"},
}

func publicEVM(w Writer) string {
	if dw, ok := w.(*docWriter); ok {
		if s := strings.TrimSpace(dw.opts.PublicEVM); s != "" {
			return s
		}
	}
	return DefaultPublicEVM
}

func writeDelegate(w Writer, d model.Report) {
	_ = d
	rpc := html.EscapeString(publicEVM(w))
	w.Section("DELEGATE")
	writeSectionLead(w, "Connect MetaMask and delegate native PMT to a validator. Pick node1–node4, or choose Another validator and paste a valoper. Unbonding later takes 21 days.")

	w.WriteHTML(`<div id="delegate-app" class="delegate-app"` +
		fmt.Sprintf(` data-chain-id="%d" data-precompile="%s" data-rpc="%s"`, EVMChainID, StakingPrecompile, rpc) +
		` data-rpc-fallbacks="https://node2.pmtchain.com,https://node3.pmtchain.com,https://node4.pmtchain.com">`)

	w.WriteHTML(`<div class="delegate-app__status">`)
	w.WriteHTML(`<p class="delegate-app__row"><span class="delegate-app__label">Wallet</span> <span id="delegate-address" class="delegate-app__mono">not connected</span></p>`)
	w.WriteHTML(`<p class="delegate-app__row"><span class="delegate-app__label">PMT</span> <span id="delegate-balance" class="delegate-app__mono">—</span></p>`)
	w.WriteHTML(`<p class="delegate-app__row"><span class="delegate-app__label">Delegation</span> <span id="delegate-staked" class="delegate-app__mono">—</span></p>`)
	w.WriteHTML(`<p class="delegate-app__row"><span class="delegate-app__label">Chain</span> <span id="delegate-chain" class="delegate-app__mono">need 290290</span></p>`)
	w.WriteHTML(`</div>`)

	w.WriteHTML(`<div class="delegate-app__actions">`)
	w.WriteHTML(`<button type="button" class="delegate-btn" id="delegate-connect">Connect MetaMask</button>`)
	w.WriteHTML(`</div>`)
	w.WriteHTML(`<p id="delegate-error" class="delegate-app__error" hidden role="alert"></p>`)

	w.WriteHTML(`<fieldset class="delegate-picker" id="delegate-picker">`)
	w.WriteHTML(`<legend>Validator</legend>`)
	w.WriteHTML(`<p class="delegate-picker__hint">Pick a node to fill the valoper. Choose Another validator to paste a different one.</p>`)
	w.WriteHTML(`<label class="delegate-field"><span>Choose</span>`)
	w.WriteHTML(`<select id="delegate-valoper-select" class="delegate-input">`)
	for _, v := range OperatorValidators {
		w.WriteHTML(fmt.Sprintf(`<option value="%s">%s</option>`,
			html.EscapeString(v.Valoper), html.EscapeString(v.Label)))
	}
	w.WriteHTML(`<option value="custom">Another validator</option>`)
	w.WriteHTML(`</select></label>`)
	w.WriteHTML(`<label class="delegate-field"><span>Valoper</span>`)
	w.WriteHTML(fmt.Sprintf(`<input id="delegate-valoper" class="delegate-input" type="text" spellcheck="false" placeholder="cosmosvaloper1…" autocomplete="off" disabled value="%s"/>`,
		html.EscapeString(OperatorValidators[0].Valoper)))
	w.WriteHTML(`</label>`)
	w.WriteHTML(`</fieldset>`)

	w.WriteHTML(`<label class="delegate-field"><span>Amount (PMT)</span>`)
	w.WriteHTML(`<input id="delegate-amount" class="delegate-input" type="text" inputmode="decimal" placeholder="you choose" autocomplete="off"/>`)
	w.WriteHTML(`</label>`)

	w.WriteHTML(`<div class="delegate-app__actions">`)
	w.WriteHTML(`<button type="button" class="delegate-btn delegate-btn--primary" id="delegate-submit">Delegate</button>`)
	w.WriteHTML(`</div>`)
	w.WriteHTML(`<p id="delegate-error-submit" class="delegate-app__error" hidden role="alert"></p>`)
	w.WriteHTML(`<p id="delegate-tx" class="delegate-app__tx" hidden></p>`)

	w.WriteHTML(`<ul class="delegate-hints">`)
	w.WriteHTML(`<li>Need an injected wallet (MetaMask / Rabby). This page never holds keys.</li>`)
	w.WriteHTML(`<li>Switch to chain <code>290290</code> (PMT).</li>`)
	w.WriteHTML(`<li>Connect the account that actually holds the PMT you want to bond.</li>`)
	w.WriteHTML(`<li>Validator is a <code>cosmosvaloper1…</code> string, not a <code>0x</code> operator address.</li>`)
	w.WriteHTML(`<li>A pasted valoper that is jailed, unbonded, or mistyped will revert.</li>`)
	w.WriteHTML(`<li>Leave liquid PMT for gas. Bonding the entire balance often fails the tx.</li>`)
	w.WriteHTML(`<li>The tx value is 0. Do not send PMT to <code>0x…0800</code> as a transfer.</li>`)
	w.WriteHTML(`<li>Signer must be the delegator (<code>msg.sender</code>). Wrong account reverts.</li>`)
	w.WriteHTML(`<li>One click = one tx. Repeat to bond to another validator.</li>`)
	w.WriteHTML(`<li>Unbonding later (not this page) takes 21 days. This page does not undelegate.</li>`)
	w.WriteHTML(`</ul>`)
	w.WriteHTML(`</div>`)
}