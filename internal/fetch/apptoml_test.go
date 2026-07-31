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

[evm]
max-tx-gas-wanted = 0
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

func TestFetchAppTomlJSONRPCAPIs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.toml")
	content := `[json-rpc]
api = "eth,txpool,net,web3"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("APPTOML_PATH", path)

	cfg := FetchAppTomlGasConfig()
	if cfg.JSONRPCAPIs != "eth,txpool,net,web3" {
		t.Fatalf("json-rpc api: %q", cfg.JSONRPCAPIs)
	}
	if !cfg.OK {
		t.Fatal("expected OK when json-rpc api is set")
	}
}

func TestFetchAppTomlTxpoolLimits(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.toml")
	content := `[evm.mempool]
global-slots = 5120
global-queue = 1024
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("APPTOML_PATH", path)

	cfg := FetchAppTomlGasConfig()
	if cfg.TxpoolGlobalSlots != "5120" || cfg.TxpoolGlobalQueue != "1024" {
		t.Fatalf("txpool limits: %+v", cfg)
	}
}
