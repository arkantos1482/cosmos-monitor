package main

import (
	"fmt"
	"io"
	"strings"

	mcmd "github.com/AlexanderGrooff/mermaid-ascii/cmd"
	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
)

// renderMermaid converts Mermaid source to Unicode box-drawing text (terminal + web).
func renderMermaid(src string) (string, error) {
	cfg, err := diagram.NewCLIConfig(false, false, false, 1, 4, 4, "TD")
	if err != nil {
		return "", err
	}
	out, err := mcmd.RenderDiagram(src, cfg)
	if err == nil {
		return strings.TrimRight(out, "\n"), nil
	}
	cfg2, err2 := diagram.NewCLIConfig(true, false, false, 1, 4, 4, "TD")
	if err2 != nil {
		return "", err
	}
	out, err = mcmd.RenderDiagram(src, cfg2)
	return strings.TrimRight(out, "\n"), err
}

func writeDiagram(w io.Writer, mermaid string) {
	out, err := renderMermaid(mermaid)
	if err != nil {
		fmt.Fprintf(w, "_diagram render failed: %v_\n\n", err)
		return
	}
	fmt.Fprintf(w, "```text\n%s\n```\n\n", out)
}

func mermaidLabel(s string) string {
	s = strings.ReplaceAll(s, `"`, `'`)
	s = strings.ReplaceAll(s, "\n", " ")
	return `"` + s + `"`
}

func economicsOverviewMermaid(d WebData) string {
	infl := fmt.Sprintf("Inflation %.2f%%", d.Inflation)
	if d.Inflation == 0 {
		infl = "Inflation OFF"
	}
	comm := fmt.Sprintf("Community tax %s", d.CommunityTax)
	if d.CommunityTaxZero {
		comm = "Community tax 0%"
	}
	distLabel := "Block rewards (x/distribution)"
	if d.BondedPct > 0 {
		if d.GoalBonded > 0 {
			distLabel = fmt.Sprintf("x/distribution · bonded %.1f%% (goal %.0f%%)", d.BondedPct, d.GoalBonded)
		} else {
			distLabel = fmt.Sprintf("x/distribution · bonded %.1f%%", d.BondedPct)
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "graph TD\n")
	fmt.Fprintf(&b, "  fees[%s]\n", mermaidLabel("Tx fees (EVM/Cosmos)"))
	fmt.Fprintf(&b, "  mint[%s]\n", mermaidLabel(infl))
	fmt.Fprintf(&b, "  dist[%s]\n", mermaidLabel(distLabel))
	fmt.Fprintf(&b, "  val[%s]\n", mermaidLabel("Validators (by stake %)"))
	fmt.Fprintf(&b, "  comm[%s]\n", mermaidLabel(comm))
	fmt.Fprintf(&b, "  op[%s]\n", mermaidLabel("commission → operator"))
	fmt.Fprintf(&b, "  del[%s]\n", mermaidLabel("remainder → delegators"))
	if d.PMTEnabled {
		pmt := "PMT pool (x/pmtrewards)"
		if d.PMTRate != "" {
			pmt = "PMT " + d.PMTRate + "/block"
		}
		if d.PMTPoolEmpty {
			pmt += " — empty"
		}
		fmt.Fprintf(&b, "  pmt[%s]\n", mermaidLabel(pmt))
	}
	fmt.Fprintf(&b, "  fees --> dist\n")
	fmt.Fprintf(&b, "  mint --> dist\n")
	fmt.Fprintf(&b, "  dist --> val\n")
	fmt.Fprintf(&b, "  dist --> comm\n")
	if d.PMTEnabled {
		fmt.Fprintf(&b, "  dist --> pmt\n")
	}
	fmt.Fprintf(&b, "  val --> op\n")
	fmt.Fprintf(&b, "  val --> del\n")
	return b.String()
}

func feeFlowMermaid(d WebData) string {
	fm := "EIP-1559 (x/feemarket)"
	if d.NoBaseFee {
		fm = "feemarket (no_base_fee)"
	}

	var splitLabel string
	if d.CommunityTaxZero {
		splitLabel = "100% → validators (pro-rata stake)"
	} else {
		splitLabel = fmt.Sprintf("%.0f%% community / %.0f%% validators",
			d.CommunityTaxPct, 100-d.CommunityTaxPct)
	}

	fmLabel := fm
	if d.BaseFee != "" || d.GasPrice != "" {
		if d.BaseFee != "" && d.GasPrice != "" {
			fmLabel = fm + " · base " + d.BaseFee + " · gas " + d.GasPrice
		} else if d.BaseFee != "" {
			fmLabel = fm + " · base " + d.BaseFee
		} else {
			fmLabel = fm + " · gas " + d.GasPrice
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "graph TD\n")
	fmt.Fprintf(&b, "  user[%s]\n", mermaidLabel("User EVM tx"))
	fmt.Fprintf(&b, "  fm[%s]\n", mermaidLabel(fmLabel))
	fmt.Fprintf(&b, "  split[%s]\n", mermaidLabel("x/distribution: "+splitLabel))
	fmt.Fprintf(&b, "  val[%s]\n", mermaidLabel("Validators + delegators"))
	if !d.CommunityTaxZero {
		fmt.Fprintf(&b, "  comm[%s]\n", mermaidLabel("Community pool"))
	}
	if d.PMTEnabled && d.PMTRate != "" {
		fmt.Fprintf(&b, "  pmt[%s]\n", mermaidLabel("PMT +"+d.PMTRate+"/block"))
	}
	fmt.Fprintf(&b, "  user -->|gas used x price| fm\n")
	fmt.Fprintf(&b, "  fm --> split\n")
	fmt.Fprintf(&b, "  split --> val\n")
	if !d.CommunityTaxZero {
		fmt.Fprintf(&b, "  split --> comm\n")
	}
	if d.PMTEnabled && d.PMTRate != "" {
		fmt.Fprintf(&b, "  pmt --> val\n")
	}
	return b.String()
}

func pmtRewardsMermaid(d WebData) string {
	if !d.PMTEnabled {
		return ""
	}
	pool := "PMT pool"
	if d.PMTBalance != "" {
		pool = "Pool " + d.PMTBalance
		if d.PMTRunway != "" {
			pool += " (" + d.PMTRunway + ")"
		}
	} else if d.PMTPoolEmpty {
		pool = "Pool empty"
	}
	rate := "per-block emission"
	if d.PMTRate != "" {
		rate = d.PMTRate + "/block"
	}
	if d.PMTPoolEmpty {
		rate = rate + " — pool empty"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "graph LR\n")
	fmt.Fprintf(&b, "  pool[%s]\n", mermaidLabel(pool))
	fmt.Fprintf(&b, "  emit[%s]\n", mermaidLabel(rate))
	fmt.Fprintf(&b, "  val[%s]\n", mermaidLabel("Validators (stake %)"))
	fmt.Fprintf(&b, "  pool --> emit\n")
	fmt.Fprintf(&b, "  emit --> val\n")
	return b.String()
}
