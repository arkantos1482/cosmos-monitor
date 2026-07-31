package fetchall

import (
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/panel"
)

func TestNeedsAppTomlIncludesEVM(t *testing.T) {
	if !needsAppToml(panel.ViewEVM) {
		t.Fatal("EVM view must fetch app.toml for JSON-RPC APIs and txpool limits")
	}
	if !needsAppToml(panel.ViewHome) || !needsAppToml(panel.ViewFeemarket) {
		t.Fatal("home and feemarket should still fetch app.toml")
	}
	if needsAppToml(panel.ViewStaking) {
		t.Fatal("staking should not fetch app.toml")
	}
}
