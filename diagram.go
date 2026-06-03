package main

import (
	"fmt"
	"io"
	"strings"

	mcmd "github.com/AlexanderGrooff/mermaid-ascii/cmd"
	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
	"github.com/arkantos1482/cosmos-monitor/fetch"
)

// Diagram padding (mermaid-ascii): border = space inside each box; padX/padY = gap between nodes.
// Defaults are compact for TUI scrolling; override with -diagram-border / -diagram-padx / -diagram-pady.
var (
	diagramBorderPad = 0
	diagramPadX      = 2
	diagramPadY      = 2
)

func SetDiagramPadding(border, padX, padY int) {
	if border < 0 {
		border = 0
	}
	if padX < 0 {
		padX = 0
	}
	if padY < 0 {
		padY = 0
	}
	diagramBorderPad = border
	diagramPadX = padX
	diagramPadY = padY
}

func mermaidConfig(useAscii bool) (*diagram.Config, error) {
	return diagram.NewCLIConfig(useAscii, false, false, diagramBorderPad, diagramPadX, diagramPadY, "TD")
}

// renderMermaid converts Mermaid source to Unicode box-drawing text (terminal + web).
func renderMermaid(src string) (string, error) {
	cfg, err := mermaidConfig(false)
	if err != nil {
		return "", err
	}
	out, err := mcmd.RenderDiagram(src, cfg)
	if err == nil {
		return strings.TrimRight(out, "\n"), nil
	}
	cfg2, err2 := mermaidConfig(true)
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

func diagramDenom(d WebData) string {
	if d.EVMDenom != "" {
		return d.EVMDenom
	}
	if d.BondDenom != "" {
		return d.BondDenom
	}
	return "apmt"
}

func economicsOverviewMermaid(d WebData) string {
	comm := fmt.Sprintf("Community tax %s", d.CommunityTax)
	if d.CommunityTaxZero {
		comm = "Community tax 0%"
	}
	distLabel := "x/distribution BeginBlock"
	if d.BondedPct > 0 {
		if d.GoalBonded > 0 {
			distLabel = fmt.Sprintf("x/distribution · %.1f%% bonded (goal %.0f%%)", d.BondedPct, d.GoalBonded)
		} else {
			distLabel = fmt.Sprintf("x/distribution · %.1f%% bonded", d.BondedPct)
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "graph TD\n")
	fmt.Fprintf(&b, "  fees[%s]\n", mermaidLabel("Tx fees (ante / EVM)"))
	fmt.Fprintf(&b, "  fc[%s]\n", mermaidLabel("fee_collector module"))
	fmt.Fprintf(&b, "  dist[%s]\n", mermaidLabel(distLabel))
	fmt.Fprintf(&b, "  val[%s]\n", mermaidLabel("Validators (by stake %)"))
	fmt.Fprintf(&b, "  comm[%s]\n", mermaidLabel(comm))
	fmt.Fprintf(&b, "  op[%s]\n", mermaidLabel("commission → operator"))
	fmt.Fprintf(&b, "  del[%s]\n", mermaidLabel("remainder → delegators"))

	if d.Inflation > 0 {
		infl := fmt.Sprintf("Inflation %.2f%% (x/mint)", d.Inflation)
		fmt.Fprintf(&b, "  infl[%s]\n", mermaidLabel(infl))
	}
	if d.PMTEnabled {
		pmt := "PMT pool (x/pmtrewards)"
		if d.PMTRate != "" {
			pmt = "PMT pool · " + d.PMTRate
		}
		if d.PMTPoolEmpty {
			pmt += " — empty"
		}
		fmt.Fprintf(&b, "  pmtPool[%s]\n", mermaidLabel(pmt))
	}

	fmt.Fprintf(&b, "  fees --> fc\n")
	if d.Inflation > 0 {
		fmt.Fprintf(&b, "  infl --> fc\n")
	}
	if d.PMTEnabled {
		fmt.Fprintf(&b, "  pmtPool -->|mint BeginBlock hook| fc\n")
	}
	fmt.Fprintf(&b, "  fc --> dist\n")
	fmt.Fprintf(&b, "  dist --> val\n")
	fmt.Fprintf(&b, "  dist --> comm\n")
	fmt.Fprintf(&b, "  val --> op\n")
	fmt.Fprintf(&b, "  val --> del\n")
	return b.String()
}

func feemarketMechanicsMermaid(d WebData) string {
	denom := diagramDenom(d)

	gasUsed := "gas used (last block)"
	if d.BlockGas != "" {
		gasUsed = "gas used: " + d.BlockGas
	}
	gasTarget := "target = max_block_gas ÷ elasticity"
	if d.Elasticity > 0 {
		gasTarget = fmt.Sprintf("target = max_block_gas ÷ %d", d.Elasticity)
	}

	parentBF := "parent base fee"
	if d.BaseFee != "" {
		parentBF = "base fee: " + fetch.FormatFeeAmount(d.BaseFee, denom)
	}

	baseFeeLabel := "base fee (this block)"
	if d.NoBaseFee {
		baseFeeLabel = "base fee disabled (no_base_fee)"
	} else if d.BaseFee != "" {
		baseFeeLabel = "base fee: " + fetch.FormatFeeAmount(d.BaseFee, denom)
	}

	gasRPC := "eth_gasPrice"
	if d.GasPrice != "" {
		gasRPC = "gas price: " + fetch.FormatFeeAmount(d.GasPrice, denom)
	}

	var paramsParts []string
	if d.BaseFeeChangeDenominator > 0 {
		paramsParts = append(paramsParts, fmt.Sprintf("denom %d", d.BaseFeeChangeDenominator))
	}
	if d.Elasticity > 0 {
		paramsParts = append(paramsParts, fmt.Sprintf("elasticity %d", d.Elasticity))
	}
	if d.MinGasPrice != "" {
		paramsParts = append(paramsParts, "min_gas "+d.MinGasPrice)
	}
	if d.AdjCap != "" {
		paramsParts = append(paramsParts, "max Δ "+d.AdjCap)
	}
	paramsLabel := "feemarket params"
	if len(paramsParts) > 0 {
		paramsLabel = strings.Join(paramsParts, " · ")
	}

	var b strings.Builder
	fmt.Fprintf(&b, "graph TD\n")
	fmt.Fprintf(&b, "  gasUsed[%s]\n", mermaidLabel(gasUsed))
	fmt.Fprintf(&b, "  gasTarget[%s]\n", mermaidLabel(gasTarget))
	fmt.Fprintf(&b, "  parentBF[%s]\n", mermaidLabel(parentBF))
	fmt.Fprintf(&b, "  params[%s]\n", mermaidLabel(paramsLabel))
	fmt.Fprintf(&b, "  calc[%s]\n", mermaidLabel("BeginBlock: CalculateBaseFee"))
	fmt.Fprintf(&b, "  baseFee[%s]\n", mermaidLabel(baseFeeLabel))
	fmt.Fprintf(&b, "  ante[%s]\n", mermaidLabel("ante: VerifyFee + DeductFees"))
	fmt.Fprintf(&b, "  eff[%s]\n", mermaidLabel("effective price ≥ base fee"))
	fmt.Fprintf(&b, "  gasRPC[%s]\n", mermaidLabel(gasRPC))
	fmt.Fprintf(&b, "  endBlk[%s]\n", mermaidLabel("EndBlock: block_gas_wanted"))

	fmt.Fprintf(&b, "  gasUsed --> calc\n")
	fmt.Fprintf(&b, "  gasTarget --> calc\n")
	fmt.Fprintf(&b, "  parentBF --> calc\n")
	fmt.Fprintf(&b, "  params --> calc\n")
	fmt.Fprintf(&b, "  calc -->|gasUsed > target ⇒ fee ↑| baseFee\n")
	fmt.Fprintf(&b, "  baseFee --> eff\n")
	fmt.Fprintf(&b, "  baseFee --> gasRPC\n")
	fmt.Fprintf(&b, "  eff --> ante\n")
	fmt.Fprintf(&b, "  baseFee --> endBlk\n")
	return b.String()
}
