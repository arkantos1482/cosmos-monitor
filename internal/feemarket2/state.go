package feemarket2

import (
	"fmt"
	"strconv"
	"strings"

	sdkmath "cosmossdk.io/math"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

// Adjustment describes how the next block's base fee will move.
type Adjustment int

const (
	AdjStable Adjustment = iota
	AdjRising
	AdjFalling
	AdjDisabled
	AdjPending
)

func (a Adjustment) Label() string {
	switch a {
	case AdjRising:
		return "rising"
	case AdjFalling:
		return "falling"
	case AdjDisabled:
		return "fixed pricing"
	case AdjPending:
		return "not active yet"
	default:
		return "stable"
	}
}

func (a Adjustment) BadgeClass() string {
	switch a {
	case AdjRising:
		return "badge--bad"
	case AdjFalling, AdjStable:
		return "badge--ok"
	case AdjDisabled, AdjPending:
		return "badge--warn"
	default:
		return ""
	}
}

// State is a module-centric snapshot derived from chain queries in model.Report.
type State struct {
	Height       int64
	Denom        string
	BaseFee      string
	BaseFeeRaw   string
	MinGasPrice  string
	MinGasMult   string

	NoBaseFee    bool
	EnableHeight int64
	EIP1559On    bool
	Mode         string

	GasWanted uint64
	GasUsed   uint64
	GasLimit  uint64
	GasTarget uint64
	UtilPct   int

	Elasticity      int64
	ChangeDenom     int64
	BaseFeeParam    string
	LastBlockFees   string

	Adj            Adjustment
	ProjectedRaw   string
	HasProjection  bool

	NodeMinGas     string
	NodeEVMTip     string
	NodePriceLimit string
	NodeMaxGasWant string
}

// LoadState builds fee-market state from a dashboard report.
func LoadState(d model.Report) State {
	s := State{
		Height:          parseInt64(d.BlockHeight),
		Denom:           pickDenom(d),
		BaseFee:         d.BaseFee,
		BaseFeeRaw:      d.BaseFeeRaw,
		MinGasPrice:     d.MinGasPrice,
		MinGasMult:      d.MinGasMultiplier,
		NoBaseFee:       d.NoBaseFee,
		EnableHeight:    d.EnableHeight,
		Elasticity:      d.Elasticity,
		ChangeDenom:     d.BaseFeeChangeDenominator,
		BaseFeeParam:    d.BaseFeeParam,
		LastBlockFees:   d.LastBlockFees,
		GasWanted:       parseUint(d.BlockGas),
		GasUsed:         d.ParentBlockGasUsed,
		GasLimit:        d.BlockGasLimit,
		NodeMinGas:      d.NodeMinGasPrices,
		NodeEVMTip:      d.NodeEVMMinTip,
		NodePriceLimit:  d.NodeMempoolPriceLimit,
		NodeMaxGasWant:  d.NodeMaxTxGasWanted,
	}
	if s.GasUsed == 0 && d.ParentBlockGasWanted > 0 {
		s.GasUsed = d.ParentBlockGasWanted
	}
	if s.Elasticity > 0 && s.GasLimit > 0 {
		s.GasTarget = s.GasLimit / uint64(s.Elasticity)
	}
	if s.GasTarget > 0 && s.GasWanted > 0 {
		s.UtilPct = int(s.GasWanted * 100 / s.GasTarget)
		if s.UtilPct > 100 {
			s.UtilPct = 100
		}
	}
	s.EIP1559On = !s.NoBaseFee && s.Height >= s.EnableHeight
	s.Mode = modeLabel(s)
	s.Adj, s.ProjectedRaw, s.HasProjection = projectAdjustment(s)
	return s
}

func modeLabel(s State) string {
	if s.NoBaseFee {
		return "Fixed (no_base_fee)"
	}
	if s.Height < s.EnableHeight {
		return fmt.Sprintf("EIP-1559 from block %d", s.EnableHeight)
	}
	return "EIP-1559 active"
}

func projectAdjustment(s State) (Adjustment, string, bool) {
	if s.NoBaseFee {
		return AdjDisabled, "", false
	}
	if s.Height < s.EnableHeight {
		return AdjPending, "", false
	}
	if s.GasTarget == 0 || s.ChangeDenom == 0 {
		return AdjStable, "", false
	}
	parent, ok := parseDec(s.BaseFeeRaw)
	if !ok {
		parent, ok = parseDec(s.BaseFeeParam)
	}
	if !ok {
		return AdjStable, "", false
	}
	gas := s.GasWanted
	if gas == 0 {
		gas = s.GasUsed
	}
	minPrice, _ := parseDec(s.MinGasPrice)
	next := CalcGasBaseFee(gas, s.GasTarget, uint64(s.ChangeDenom), parent, MinUnitGas, minPrice)
	if next.Equal(parent) {
		return AdjStable, next.String(), true
	}
	if next.GT(parent) {
		return AdjRising, next.String(), true
	}
	return AdjFalling, next.String(), true
}

// TransferCost estimates a 21k-gas EVM transfer at the current base fee.
func TransferCost(baseFeeRaw, denom string) string {
	raw := strings.TrimSpace(baseFeeRaw)
	if raw == "" || raw == "0" {
		return "—"
	}
	fee, err := strconv.ParseFloat(raw, 64)
	if err != nil || fee <= 0 {
		return "—"
	}
	total := fee * 21000
	return fetch.FormatCoin(fmt.Sprintf("%.0f", total), denom)
}

func pickDenom(d model.Report) string {
	if d.EVMDenom != "" {
		return d.EVMDenom
	}
	return d.BondDenom
}

func parseInt64(s string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return n
}

func parseUint(s string) uint64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	n, _ := strconv.ParseUint(s, 10, 64)
	return n
}

func parseDec(s string) (sdkmath.LegacyDec, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return sdkmath.LegacyDec{}, false
	}
	d, err := sdkmath.LegacyNewDecFromStr(s)
	if err != nil {
		return sdkmath.LegacyDec{}, false
	}
	return d, true
}
