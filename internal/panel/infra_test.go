package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestInfraContent(t *testing.T) {
	d := model.Report{
		NumCPU: 4,
		MemPct: 72, MemUsed: "11 GiB", MemTotal: "16 GiB", MemAvail: "4.5 GiB",
		DiskPct: 45, DiskUsed: "120 GiB", DiskTotal: "256 GiB", DiskAvail: "136 GiB",
		HasChainDataDisk: true, DataDiskUsed: "88 GiB", DataDiskTotal: "200 GiB", DataDiskPct: 44,
		Load1: 1.2, Load5: 0.9, Load15: 0.7,
		SwapUsed: "128 MiB", SwapTotal: "2 GiB",
	}
	out := BuildView(ViewInfra, d)

	for _, want := range []string{
		`class="dash-subheading">Host resources</h3>`,
		`class="infra-meter"`,
		`chain data`,
		`4 CPUs`,
		`88 GiB used of 200 GiB`,
		`class="infra-summary__gauges"`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("infra view missing %q", want)
		}
	}
	for _, gone := range []string{
		`class="dash-subheading">Container</h3>`,
		`/home/ubuntu/.evmd`,
		`.evmd`,
		`eco-domain--infra`,
		`pmt-blockchain:latest`,
		`started at`,
		`class="dash-layer__title">Host</h3>`,
		`Host vs container`,
		`infra-compare`,
		`>root disk<`,
	} {
		if strings.Contains(out, gone) {
			t.Fatalf("infra view should not contain %q", gone)
		}
	}
}

func TestInfraSummaryUsesChainDataGauge(t *testing.T) {
	d := model.Report{
		MemPct:           10,
		DiskPct:          90,
		HasChainDataDisk: true,
		DataDiskPct:      55,
		Load1:            0.4,
		NumCPU:           2,
	}
	out := BuildView(ViewInfra, d)
	if !strings.Contains(out, `>chain data<`) {
		t.Fatal("summary should label disk gauge as chain data when DATA_PATH is set")
	}
}

func TestInfraMeterTone(t *testing.T) {
	if infraMeterTone(50) != "" {
		t.Fatal("expected no tone below 75%")
	}
	if infraMeterTone(80) != "warn" {
		t.Fatal("expected warn at 80%")
	}
	if infraMeterTone(95) != "bad" {
		t.Fatal("expected bad at 95%")
	}
}
