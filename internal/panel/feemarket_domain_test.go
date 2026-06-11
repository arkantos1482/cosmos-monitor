package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestFeemarketDomainCards(t *testing.T) {
	d := model.Report{
		BlockHeight: "100", BaseFee: "7 apmt", BaseFeeRaw: "7000",
		ParentBlockGasWanted: 21000, ParentBlockGasUsed: 18000,
		BlockGasLimit: 30_000_000, Elasticity: 2, EVMDenom: "apmt",
		BaseFeeChangeDenominator: 8, MinGasMultiplier: "0.5",
		ParentBlockResultsOK: true, EVMChainID: 290290,
		ModuleAccounts: []model.ModuleAccountRow{
			{Name: "fee_collector", Balance: "1 PMT", Address: "cosmos1fee"},
		},
	}
	chunk := feemarketChunk(t, Build(d))
	cards := chunk
	if end := strings.Index(chunk, `class="fee-nav"`); end > 0 {
		cards = chunk[:end]
	}
	for _, want := range []string{
		`eco-domain--feemarket`,
		"base fee",
		"parent gas",
		`class="eco-domain__effect"`,
	} {
		if !strings.Contains(cards, want) {
			t.Fatalf("fee market missing %q", want)
		}
	}
	for _, gone := range []string{
		`eco-domain--vm`,
		"fee_collector",
		"elasticity_multiplier",
		`eco-domain__divider">Governance params`,
	} {
		if strings.Contains(cards, gone) {
			t.Fatalf("fee market domain card should be live-state only, found %q", gone)
		}
	}
}

func TestFeemarketL5ParamEffects(t *testing.T) {
	d := model.Report{
		BlockHeight: "1", BaseFeeRaw: "7",
		BlockGasLimit: ^uint64(0), Elasticity: 2,
		BaseFeeChangeDenominator: 8, EVMDenom: "apmt",
	}
	chunk := feemarketChunk(t, Build(d))
	l5 := chunk[strings.Index(chunk, `id="fee-L5"`):]
	if !strings.Contains(l5, `eco-domain--fee-params`) {
		t.Fatal("L5 should use domain param rows with effect column")
	}
	if strings.Contains(l5, `<table class="fee-table"><thead><tr><th>Parameter</th>`) {
		t.Fatal("L5 should not use flat param table without effect column")
	}
	if !strings.Contains(l5, "elasticity_multiplier") {
		t.Fatal("L5 should list governance params")
	}
}
