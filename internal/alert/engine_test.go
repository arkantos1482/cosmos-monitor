package alert

import (
	"strings"
	"testing"
	"time"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/panel"
)

type stubSender struct {
	msgs []string
}

func (s *stubSender) Send(text string) error {
	s.msgs = append(s.msgs, text)
	return nil
}

func healthyReport() model.Report {
	return model.Report{
		NodeRunning: true, Synced: true,
		EVMRPCOk: true, EVMListening: true, EVMSynced: true,
		RPCProbeOK: 1, RPCProbeTotal: 1,
		PMTEnabled: true, Inflation: 3.5,
	}
}

func testFinding(section, key, severity, msg string) panel.Finding {
	return panel.Finding{Section: section, Key: key, Severity: severity, Message: msg}
}

func TestEngineCooldownSuppressesRepeatAlert(t *testing.T) {
	sender := &stubSender{}
	load := func() model.Report {
		r := healthyReport()
		r.NodeRunning = false
		return r
	}
	cfg := Config{
		NodeName: "node4",
		Interval: time.Second,
		Cooldown: time.Hour,
		Token:    "tok",
		ChatID:   "123",
	}
	e := NewEngine(cfg, load, sender)
	e.tick()
	e.tick()
	if len(sender.msgs) != 1 {
		t.Fatalf("expected 1 alert after cooldown, got %d", len(sender.msgs))
	}
}

func TestEngineRecoveryAfterClear(t *testing.T) {
	sender := &stubSender{}
	step := 0
	load := func() model.Report {
		step++
		r := healthyReport()
		if step == 1 {
			r.NodeRunning = false
		}
		return r
	}
	cfg := Config{
		NodeName: "node4",
		Interval: time.Second,
		Cooldown: 0,
		Token:    "tok",
		ChatID:   "123",
	}
	e := NewEngine(cfg, load, sender)
	e.tick()
	e.tick()
	if len(sender.msgs) != 2 {
		t.Fatalf("expected alert + recovery, got %d messages", len(sender.msgs))
	}
	if !strings.Contains(sender.msgs[1], "resolved") {
		t.Fatalf("expected recovery message, got %q", sender.msgs[1])
	}
	if !strings.Contains(sender.msgs[1], "container stopped") {
		t.Fatalf("recovery should name cleared finding, got %q", sender.msgs[1])
	}
}

func TestEngineWorsenedSeverityRefires(t *testing.T) {
	sender := &stubSender{}
	step := 0
	load := func() model.Report {
		step++
		r := healthyReport()
		r.EVMBlockAge = "45s"
		if step >= 2 {
			r.EVMBlockAgeErr = true
			r.EVMBlockAge = "2m"
		} else {
			r.EVMBlockAgeWarn = true
		}
		return r
	}
	cfg := Config{
		NodeName: "node4",
		Interval: time.Second,
		Cooldown: 0,
		Token:    "tok",
		ChatID:   "123",
	}
	e := NewEngine(cfg, load, sender)
	e.tick()
	e.tick()
	if len(sender.msgs) != 2 {
		t.Fatalf("expected refire on warn→bad, got %d messages", len(sender.msgs))
	}
}

func TestFormatAlertMessage(t *testing.T) {
	active := map[string]panel.Finding{
		findingID(testFinding("infra", "container_stopped", "bad", "container stopped")): testFinding("infra", "container_stopped", "bad", "container stopped"),
		findingID(testFinding("evm", "rpc_degraded", "warn", "RPC DEGRADED (2/6 probes failing)")): testFinding("evm", "rpc_degraded", "warn", "RPC DEGRADED (2/6 probes failing)"),
	}
	msg := formatAlertMessage("node4", active, model.Report{BlockHeight: "1842032", TimeSinceBlock: "12s"})
	for _, want := range []string{
		"🚨 PMT — node4",
		"Infrastructure · bad",
		"container stopped",
		"EVM JSON-RPC · warn",
		"RPC DEGRADED",
		"Height 1842032 · 12s ago",
	} {
		if !strings.Contains(msg, want) {
			t.Fatalf("alert message missing %q:\n%s", want, msg)
		}
	}
}

func TestFormatRecoveryMessage(t *testing.T) {
	msg := formatRecoveryMessage("node4", testFinding("infra", "container_stopped", "bad", "container stopped"))
	if !strings.Contains(msg, "✅ PMT — node4 resolved") {
		t.Fatalf("unexpected recovery header: %q", msg)
	}
	if !strings.Contains(msg, "Infrastructure · container stopped") {
		t.Fatalf("unexpected recovery body: %q", msg)
	}
}

func TestEscapeHTML(t *testing.T) {
	if got := escapeHTML(`a<b&c>`); got != `a&lt;b&amp;c&gt;` {
		t.Fatalf("escapeHTML: got %q", got)
	}
}
