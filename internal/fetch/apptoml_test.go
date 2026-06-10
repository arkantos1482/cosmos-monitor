package fetch

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFetchAppTomlGasConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.toml")
	content := `minimum-gas-prices = "0apmt"
max-tx-gas-wanted = 0

[evm]
min-tip = 0

[evm.mempool]
price-limit = 1
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("APPTOML_PATH", path)

	cfg := FetchAppTomlGasConfig()
	if !cfg.OK {
		t.Fatal("expected parsed config")
	}
	if cfg.MinGasPrices != "0apmt" {
		t.Fatalf("min gas prices: %q", cfg.MinGasPrices)
	}
	if cfg.EVMMinTip != "0" || cfg.MempoolPriceLimit != "1" || cfg.MaxTxGasWanted != "0" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
}
