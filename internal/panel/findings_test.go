package panel

import (
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func findingKeys(fs []Finding) map[string]string {
	m := make(map[string]string, len(fs))
	for _, f := range fs {
		m[f.Section+"/"+f.Key] = f.Severity
	}
	return m
}

func hasFinding(t *testing.T, fs []Finding, section, key, severity string) {
	t.Helper()
	for _, f := range fs {
		if f.Section == section && f.Key == key {
			if f.Severity != severity {
				t.Fatalf("finding %s/%s: want severity %q, got %q", section, key, severity, f.Severity)
			}
			return
		}
	}
	t.Fatalf("missing finding %s/%s (%s)", section, key, severity)
}

func TestFindingsInfra(t *testing.T) {
	fs := Findings(model.Report{
		MemPct: 80, DiskPct: 92,
	})
	hasFinding(t, fs, "infra", "host_ram", "warn")
	hasFinding(t, fs, "infra", "disk", "bad")
}

func TestFindingsInfraChainDataDisk(t *testing.T) {
	fs := Findings(model.Report{
		HasChainDataDisk: true, DataDiskPct: 78, DiskPct: 10,
	})
	hasFinding(t, fs, "infra", "disk", "warn")
}

func TestFindingsEVM(t *testing.T) {
	fs := Findings(model.Report{
		EVMRPCOk: false, HasEVMListening: true, EVMListening: false, EVMSynced: false,
		EVMBlockAge: "2m", EVMBlockAgeErr: true,
		RPCProbeOK: 4, RPCProbeTotal: 6,
	})
	hasFinding(t, fs, "evm", "rpc_down", "bad")
	hasFinding(t, fs, "evm", "not_listening", "warn")
	for _, f := range fs {
		if f.Section == "evm" && (f.Key == "rpc_degraded" || f.Key == "syncing" || f.Key == "block_age") {
			t.Fatalf("UI-only EVM condition must not alert: %+v", f)
		}
	}

	fs = Findings(model.Report{
		EVMRPCOk: true, HasEVMListening: true, EVMListening: true, EVMSynced: true,
		EVMBlockAge: "45s", EVMBlockAgeWarn: true,
		RPCProbeOK: 4, RPCProbeTotal: 6, JSONRPCAPIs: "eth,net,web3",
	})
	for _, f := range fs {
		if f.Section == "evm" {
			t.Fatalf("degraded/slow must not alert when liveness OK: %+v", f)
		}
	}

	fs = Findings(model.Report{
		EVMRPCOk: false, HasEVMListening: false, EVMListening: false,
	})
	hasFinding(t, fs, "evm", "rpc_down", "bad")
	for _, f := range fs {
		if f.Key == "not_listening" {
			t.Fatalf("listening unknown must not alert not_listening: %+v", f)
		}
	}
}

func TestFindingsEVMJSONRPCAPIsUnknown(t *testing.T) {
	fs := Findings(model.Report{
		EVMRPCOk: true, HasEVMListening: true, EVMListening: true, EVMSynced: true,
	})
	hasFinding(t, fs, "evm", "jsonrpc_apis_unknown", "warn")
}

func TestFindingsNode(t *testing.T) {
	fs := Findings(model.Report{Synced: false})
	hasFinding(t, fs, "node", "catching_up", "warn")
}

func TestFindingsStaking(t *testing.T) {
	fs := Findings(model.Report{
		Local: model.LocalValidator{IsValidator: true, Jailed: true, MissedHigh: true},
	})
	hasFinding(t, fs, "staking", "jailed", "bad")
	hasFinding(t, fs, "staking", "missed_blocks_high", "warn")
}

func TestFindingsSlashing(t *testing.T) {
	fs := Findings(model.Report{
		JailedCount: 1, BelowThreshold: 2, TombstonedCount: 1,
		Local: model.LocalValidator{IsValidator: true, Missed: 4800, MaxMissed: 5000},
	})
	hasFinding(t, fs, "slashing", "network_jailed", "bad")
	hasFinding(t, fs, "slashing", "below_min_signed", "warn")
	hasFinding(t, fs, "slashing", "network_tombstoned", "bad")
	hasFinding(t, fs, "slashing", "headroom", "bad")
}

func TestFindingsRewards(t *testing.T) {
	fs := Findings(model.Report{HasPMTParams: true, PMTEnabled: false, Inflation: 0})
	hasFinding(t, fs, "rewards", "pmt_disabled", "bad")
	for _, f := range fs {
		if f.Key == "inflation_off" {
			t.Fatalf("inflation off must not page: %+v", f)
		}
	}

	fs = Findings(model.Report{HasPMTParams: true, PMTEnabled: true, PMTPoolEmpty: true})
	hasFinding(t, fs, "rewards", "pmt_not_emitting", "warn")

	fs = Findings(model.Report{Inflation: 3.5}) // PMT params unknown
	for _, f := range fs {
		if f.Key == "pmt_disabled" || f.Key == "pmt_not_emitting" || f.Key == "inflation_off" {
			t.Fatalf("must not alert PMT/inflation when params unknown: %+v", f)
		}
	}
}

func TestFindingsDistribution(t *testing.T) {
	d := model.Report{
		ModuleAccounts: []model.ModuleAccountRow{{
			Name:    "distribution",
			Balance: "100 PMT",
		}},
		UnclaimedDelegator:  "80 PMT",
		UnclaimedCommission: "30 PMT",
	}
	fs := Findings(d)
	hasFinding(t, fs, "distribution", "escrow_mismatch", "warn")
}

func TestFindingsHealthyEmpty(t *testing.T) {
	fs := Findings(model.Report{
		Synced: true,
		EVMRPCOk: true, HasEVMListening: true, EVMListening: true, EVMSynced: true,
		JSONRPCAPIs: "eth,txpool,net,web3",
		HasPMTParams: true, PMTEnabled: true, Inflation: 3.5,
	})
	for _, f := range fs {
		t.Fatalf("unexpected finding on healthy report: %+v", f)
	}
}
