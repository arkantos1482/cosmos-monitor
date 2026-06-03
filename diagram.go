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

func joinLabel(parts ...string) string {
	var out []string
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, " · ")
}

func economicsFeesLabel(d WebData) string {
	label := "Tx fees (ante / EVM)"
	if d.MempoolTxs > 0 {
		label = joinLabel(label, fmt.Sprintf("mempool %d", d.MempoolTxs))
	}
	if d.PendingTx > 0 {
		label = joinLabel(label, fmt.Sprintf("evm pending %d", d.PendingTx))
	}
	return label
}

func economicsFCLabel(d WebData) string {
	label := "fee_collector"
	if d.TotalOutstanding != "" {
		return joinLabel(label, "outstanding "+d.TotalOutstanding)
	}
	return joinLabel(label, "cleared each BeginBlock")
}

func economicsValLabel(d WebData) string {
	label := "Validators"
	if d.BondedCount > 0 {
		label = fmt.Sprintf("%d validators", d.BondedCount)
	}
	if d.BondedAmt != "" {
		label = joinLabel(label, d.BondedAmt+" bonded")
	} else if d.BondedPct > 0 {
		label = joinLabel(label, fmt.Sprintf("%.1f%% stake", d.BondedPct))
	}
	return label
}

func economicsCommLabel(d WebData) string {
	comm := fmt.Sprintf("community tax %s", d.CommunityTax)
	if d.CommunityTaxZero {
		comm = "community tax 0%"
	}
	if d.CommunityPool != "" {
		return joinLabel(comm, "pool "+d.CommunityPool)
	}
	return comm
}

func economicsOpLabel(d WebData) string {
	if d.Local.IsValidator && d.Local.Commission > 0 {
		return fmt.Sprintf("commission %.1f%% → operator", d.Local.Commission)
	}
	if n := len(d.Validators); n > 0 {
		sum := 0.0
		for _, v := range d.Validators {
			sum += v.CommissionFloat
		}
		return fmt.Sprintf("commission ~%.1f%% → operator", sum/float64(n))
	}
	return "commission → operator"
}

func economicsDelLabel(d WebData) string {
	if d.TotalOutstanding != "" {
		return joinLabel("delegators", "outstanding "+d.TotalOutstanding)
	}
	return "remainder → delegators"
}

func economicsPMTPoolLabel(d WebData) string {
	pmt := "PMT pool (x/pmtrewards)"
	if d.PMTRate != "" {
		pmt = joinLabel(pmt, d.PMTRate)
	}
	if d.PMTBalance != "" {
		pmt = joinLabel(pmt, d.PMTBalance)
		if d.PMTRunway != "" {
			pmt = joinLabel(pmt, d.PMTRunway)
		}
	} else if d.PMTPoolEmpty {
		pmt += " — empty"
	}
	return pmt
}

func economicsOverviewMermaid(d WebData) string {
	distLabel := "x/distribution BeginBlock"
	if d.BondedPct > 0 {
		if d.GoalBonded > 0 {
			distLabel = fmt.Sprintf("x/distribution · %.1f%% bonded (goal %.0f%%)", d.BondedPct, d.GoalBonded)
		} else {
			distLabel = fmt.Sprintf("x/distribution · %.1f%% bonded", d.BondedPct)
		}
	}
	if d.TotalOutstanding != "" {
		distLabel = joinLabel(distLabel, "outstanding "+d.TotalOutstanding)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "graph TD\n")
	fmt.Fprintf(&b, "  fees[%s]\n", mermaidLabel(economicsFeesLabel(d)))
	fmt.Fprintf(&b, "  fc[%s]\n", mermaidLabel(economicsFCLabel(d)))
	fmt.Fprintf(&b, "  dist[%s]\n", mermaidLabel(distLabel))
	fmt.Fprintf(&b, "  val[%s]\n", mermaidLabel(economicsValLabel(d)))
	fmt.Fprintf(&b, "  comm[%s]\n", mermaidLabel(economicsCommLabel(d)))
	fmt.Fprintf(&b, "  op[%s]\n", mermaidLabel(economicsOpLabel(d)))
	fmt.Fprintf(&b, "  del[%s]\n", mermaidLabel(economicsDelLabel(d)))

	if d.Inflation > 0 {
		infl := fmt.Sprintf("Inflation %.2f%% (x/mint)", d.Inflation)
		fmt.Fprintf(&b, "  infl[%s]\n", mermaidLabel(infl))
	}
	if d.PMTEnabled {
		fmt.Fprintf(&b, "  pmtPool[%s]\n", mermaidLabel(economicsPMTPoolLabel(d)))
	}
	if d.Inflation > 0 && d.AnnualProvisions != "" {
		infl := fmt.Sprintf("Inflation %.2f%%", d.Inflation)
		fmt.Fprintf(&b, "  infl[%s]\n", mermaidLabel(joinLabel(infl, d.AnnualProvisions+"/yr")))
	} else if d.Inflation > 0 {
		fmt.Fprintf(&b, "  infl[%s]\n", mermaidLabel(fmt.Sprintf("Inflation %.2f%% (x/mint)", d.Inflation)))
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
	fmt.Fprintf(&b, "  op --> del\n")
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
		parentBF = "parent base fee: " + fetch.FormatFeeAmount(d.BaseFee, denom)
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

	var paramsTune []string
	if d.BaseFeeChangeDenominator > 0 {
		paramsTune = append(paramsTune, fmt.Sprintf("change denom %d", d.BaseFeeChangeDenominator))
	}
	if d.Elasticity > 0 {
		paramsTune = append(paramsTune, fmt.Sprintf("elasticity %d", d.Elasticity))
	}
	paramsTuneLabel := "feemarket tuning"
	if len(paramsTune) > 0 {
		paramsTuneLabel = strings.Join(paramsTune, " · ")
	}

	var paramsFloor []string
	if d.MinGasPrice != "" {
		paramsFloor = append(paramsFloor, "min_gas "+d.MinGasPrice)
	}
	if d.AdjCap != "" {
		paramsFloor = append(paramsFloor, "max Δ "+d.AdjCap)
	}
	paramsFloorLabel := "feemarket floors"
	if len(paramsFloor) > 0 {
		paramsFloorLabel = strings.Join(paramsFloor, " · ")
	}

	var b strings.Builder
	fmt.Fprintf(&b, "graph TD\n")
	fmt.Fprintf(&b, "  gasUsed[%s]\n", mermaidLabel(gasUsed))
	fmt.Fprintf(&b, "  gasTarget[%s]\n", mermaidLabel(gasTarget))
	fmt.Fprintf(&b, "  parentBF[%s]\n", mermaidLabel(parentBF))
	fmt.Fprintf(&b, "  paramsTune[%s]\n", mermaidLabel(paramsTuneLabel))
	fmt.Fprintf(&b, "  paramsFloor[%s]\n", mermaidLabel(paramsFloorLabel))
	fmt.Fprintf(&b, "  calc[%s]\n", mermaidLabel("BeginBlock: CalculateBaseFee"))
	fmt.Fprintf(&b, "  baseFee[%s]\n", mermaidLabel(baseFeeLabel))
	fmt.Fprintf(&b, "  eff[%s]\n", mermaidLabel("effective price ≥ base fee"))
	fmt.Fprintf(&b, "  ante[%s]\n", mermaidLabel("ante: VerifyFee + DeductFees"))
	fmt.Fprintf(&b, "  gasRPC[%s]\n", mermaidLabel(gasRPC))
	fmt.Fprintf(&b, "  endBlk[%s]\n", mermaidLabel("EndBlock: block_gas_wanted"))

	// Vertical spine (avoids wide fan-in to calc / baseFee).
	fmt.Fprintf(&b, "  gasUsed --> gasTarget\n")
	fmt.Fprintf(&b, "  gasTarget --> parentBF\n")
	fmt.Fprintf(&b, "  parentBF --> paramsTune\n")
	fmt.Fprintf(&b, "  paramsTune --> paramsFloor\n")
	fmt.Fprintf(&b, "  paramsFloor --> calc\n")
	fmt.Fprintf(&b, "  calc -->|gasUsed > target ⇒ fee ↑| baseFee\n")
	fmt.Fprintf(&b, "  baseFee --> eff\n")
	fmt.Fprintf(&b, "  eff --> ante\n")
	fmt.Fprintf(&b, "  baseFee --> gasRPC\n")
	fmt.Fprintf(&b, "  ante --> endBlk\n")
	return b.String()
}
