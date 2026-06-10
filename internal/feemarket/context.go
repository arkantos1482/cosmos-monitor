package feemarket

import (
	"fmt"
	"strconv"
	"strings"

	"cosmossdk.io/math"
	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

// Badge holds fee-market status label and CSS class.
type Badge struct {
	Label string
	Class string
}

// Context is the loaded fee-market state for panel rendering.
type Context struct {
	CurrentBlock string
	ParentBlock  string
	CurrentHeight int64
	ParentHeight  int64

	Wanted            uint64
	GasUsed           uint64
	Target            uint64
	HasTarget         bool
	UnlimitedBlockGas bool

	BaseFee    string
	BaseFeeRaw string
	Denom      string

	Badge        Badge
	NoBaseFee    bool
	FeesDisabled bool

	UtilPct    string
	LoadBarPct float64

	Verdict      string
	NextAdj      string
	L1Footnote   string
	WRelation    string
	Verify       string
	DecreaseStep math.LegacyDec

	MinGasMultiplier string
	MinGasPrice      string
	MinGasPriceRaw   string
	EnableHeight     int64
	BaseFeeParam     string
	MaxBlockBytes    int64
	BlockGasLimit    uint64
	Elasticity       int64
	DenomU           uint64

	NodeMinGasPrices      string
	NodeEVMMinTip         string
	NodeMempoolPriceLimit string
	NodeMaxTxGasWanted    string
	NodeAppTomlPath       string

	HardforkLondon string
	BlockInterval  string
}

func LoadContext(d model.Report) Context {
	c := Context{
		CurrentBlock:          currentBlock(d),
		ParentBlock:           parentBlock(d),
		CurrentHeight:         parseHeight(d.BlockHeight),
		BaseFee:               d.BaseFee,
		BaseFeeRaw:            d.BaseFeeRaw,
		Denom:                 denom(d),
		NoBaseFee:             d.NoBaseFee,
		MinGasMultiplier:      d.MinGasMultiplier,
		MinGasPrice:           d.MinGasPrice,
		MinGasPriceRaw:        d.MinGasPriceRaw,
		EnableHeight:          d.EnableHeight,
		BaseFeeParam:          d.BaseFeeParam,
		MaxBlockBytes:         d.MaxBlockBytes,
		BlockGasLimit:         d.BlockGasLimit,
		Elasticity:            d.Elasticity,
		NodeMinGasPrices:      d.NodeMinGasPrices,
		NodeEVMMinTip:         d.NodeEVMMinTip,
		NodeMempoolPriceLimit: d.NodeMempoolPriceLimit,
		NodeMaxTxGasWanted:    d.NodeMaxTxGasWanted,
		NodeAppTomlPath:       d.NodeAppTomlPath,
		HardforkLondon:        d.HardforkLondon,
		BlockInterval:         d.BlockInterval,
	}
	if c.CurrentHeight > 0 {
		c.ParentHeight = c.CurrentHeight - 1
	}
	c.Wanted = storedWanted(d)
	c.GasUsed = d.ParentBlockGasUsed
	c.Target, c.HasTarget = gasTarget(d)
	c.UnlimitedBlockGas = d.BlockGasLimit == ^uint64(0)
	if d.BaseFeeChangeDenominator > 0 {
		c.DenomU = uint64(d.BaseFeeChangeDenominator)
	}

	if c.EnableHeight > 0 && c.CurrentHeight > 0 && c.CurrentHeight < c.EnableHeight {
		c.FeesDisabled = true
	}

	c.fillUtilization()
	c.fillBadge()
	c.fillVerify()
	return c
}

func denom(d model.Report) string {
	if d.EVMDenom != "" {
		return d.EVMDenom
	}
	if d.BondDenom != "" {
		return d.BondDenom
	}
	return "apmt"
}

func currentBlock(d model.Report) string {
	if d.BlockHeight != "" {
		return d.BlockHeight
	}
	return "—"
}

func parentBlock(d model.Report) string {
	h := parseHeight(d.BlockHeight)
	if h > 0 {
		return report.FormatInt(h - 1)
	}
	return "—"
}

func parseHeight(s string) int64 {
	s = strings.ReplaceAll(s, ",", "")
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func storedWanted(d model.Report) uint64 {
	if d.ParentBlockGasWanted > 0 {
		return d.ParentBlockGasWanted
	}
	return parseUint(d.BlockGas)
}

func parseUint(s string) uint64 {
	s = strings.ReplaceAll(s, ",", "")
	n, _ := strconv.ParseUint(s, 10, 64)
	return n
}

func gasTarget(d model.Report) (uint64, bool) {
	if d.BlockGasLimit > 0 && d.Elasticity > 0 {
		return d.BlockGasLimit / uint64(d.Elasticity), true
	}
	return 0, false
}

func FormatUint(n uint64) string {
	if n == ^uint64(0) {
		return MaxUint64Label
	}
	return report.FormatInt(int64(n))
}

func (c *Context) TargetDisplay() string {
	if !c.HasTarget {
		return "—"
	}
	if c.UnlimitedBlockGas {
		return fmt.Sprintf("%s ÷ %d (sentinel)", MaxUint64Label, c.Elasticity)
	}
	return FormatUint(c.Target)
}

func (c *Context) fillUtilization() {
	if c.UnlimitedBlockGas || !c.HasTarget || c.Target == 0 {
		return
	}
	pct := float64(c.Wanted) / float64(c.Target) * 100
	c.UtilPct = fmt.Sprintf("%.2f%%", pct)
	c.LoadBarPct = pct
	if c.LoadBarPct > 100 {
		c.LoadBarPct = 100
	}
}

func (c *Context) fillVerdict() {
	if !c.HasTarget {
		c.Verdict = "unknown"
		return
	}
	switch {
	case c.Wanted > c.Target:
		c.Verdict = "busy"
	case c.Wanted < c.Target:
		c.Verdict = "quiet"
	default:
		c.Verdict = "balanced"
	}
}

func (c *Context) fillBadge() {
	c.fillVerdict()

	if c.FeesDisabled {
		c.Badge = Badge{Label: "FEES DISABLED", Class: "disabled"}
		c.NextAdj = "—"
		c.L1Footnote = fmt.Sprintf("Base fee adjustment activates at block %s (enable_height).", report.FormatInt(c.EnableHeight))
		return
	}

	if c.NoBaseFee {
		c.Badge = Badge{Label: "FIXED PRICING", Class: "stable"}
		c.NextAdj = "—"
		c.L1Footnote = "Fixed minimum gas price — congestion does not change the base fee."
		return
	}

	if !c.HasTarget {
		c.Badge = Badge{Label: "UNKNOWN", Class: "stable"}
		c.NextAdj = "?"
		return
	}

	switch {
	case c.Wanted > c.Target:
		c.Badge = Badge{Label: "RISING", Class: "rising"}
		c.NextAdj = "↑"
		c.L1Footnote = fmt.Sprintf("Demand above capacity on block %s → base fee rises at BeginBlock of block %s.", c.ParentBlock, c.CurrentBlock)
	case c.Wanted == c.Target:
		c.Badge = Badge{Label: "STABLE", Class: "stable"}
		c.NextAdj = "="
		c.L1Footnote = fmt.Sprintf("Demand matched capacity on block %s → base fee holds at BeginBlock of block %s.", c.ParentBlock, c.CurrentBlock)
	default:
		current, ok := ParseLegacyDec(c.BaseFeeRaw)
		minGasPrice := math.LegacyZeroDec()
		if mp, okMP := ParseLegacyDec(c.MinGasPriceRaw); okMP {
			minGasPrice = mp
		}
		atFloor := false
		if ok {
			c.DecreaseStep = DecreaseStep(c.Wanted, c.Target, c.DenomU, current)
			newBase := CalcGasBaseFee(c.Wanted, c.Target, c.DenomU, current, MinUnitGas, minGasPrice)
			actualDelta := current.Sub(newBase)
			maxStep := math.LegacyZeroDec()
			if c.DenomU > 0 {
				maxStep = current.QuoInt(math.NewIntFromUint64(c.DenomU)).TruncateDec()
			}
			atFloor = actualDelta.TruncateDec().IsZero() || maxStep.IsZero() || !current.GT(minGasPrice)
		}
		if atFloor {
			c.Badge = Badge{Label: "AT FLOOR", Class: "floor"}
			c.NextAdj = "hold"
			c.L1Footnote = floorFootnote(*c, current, minGasPrice)
		} else {
			c.Badge = Badge{Label: "FALLING", Class: "falling"}
			c.NextAdj = "↓"
			c.L1Footnote = fmt.Sprintf("Demand below capacity on block %s → base fee falls at BeginBlock of block %s.", c.ParentBlock, c.CurrentBlock)
		}
	}
}

func floorFootnote(c Context, current, minGasPrice math.LegacyDec) string {
	var parts []string
	if c.DenomU > 0 && current.IsPositive() {
		rawStep := current.QuoInt(math.NewIntFromUint64(c.DenomU))
		if rawStep.TruncateDec().IsZero() {
			parts = append(parts, fmt.Sprintf("precision floor — max decrease step base÷%d = %s at current base.", c.DenomU, fetch.FormatFeeStep(rawStep, c.Denom)))
		}
	}
	if minGasPrice.IsPositive() && !current.GT(minGasPrice) {
		parts = append(parts, fmt.Sprintf("min_gas_price = %s (governance floor binding).", c.MinGasPrice))
	} else if c.MinGasPrice != "" {
		parts = append(parts, fmt.Sprintf("min_gas_price = %s (no governance floor binding).", c.MinGasPrice))
	} else {
		parts = append(parts, "min_gas_price = 0 (no governance floor binding).")
	}
	return strings.Join(parts, " ")
}

func (c *Context) fillVerify() {
	if c.NoBaseFee || !c.HasTarget || c.DenomU == 0 {
		return
	}
	current, ok := ParseLegacyDec(c.BaseFeeRaw)
	if !ok {
		return
	}
	minGasPrice := math.LegacyZeroDec()
	if mp, ok := ParseLegacyDec(c.MinGasPriceRaw); ok {
		minGasPrice = mp
	}
	c.Verify = VerifyMatch(current, c.Wanted, c.Target, c.DenomU, minGasPrice)
}

func (c *Context) WGasUsedRelation() string {
	if c.WRelation != "" {
		return c.WRelation
	}
	if c.Wanted == c.GasUsed {
		return "W = gas_used"
	}
	if c.Wanted > c.GasUsed {
		return "W higher — in-block gas accumulator × min_gas_multiplier exceeded gas_used"
	}
	return "W lower than gas_used (unusual)"
}

func (c *Context) HomeCardLine() string {
	switch {
	case c.FeesDisabled:
		return fmt.Sprintf("fees activate at block %s", report.FormatInt(c.EnableHeight))
	case c.Badge.Label == "AT FLOOR":
		return "at precision floor"
	case c.UnlimitedBlockGas:
		return "target sentinel"
	case c.UtilPct != "":
		return "util " + c.UtilPct + " of target"
	default:
		return "demand vs target"
	}
}
