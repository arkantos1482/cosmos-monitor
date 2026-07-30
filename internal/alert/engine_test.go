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
		EVMRPCOk: true, HasEVMListening: true, EVMListening: true, EVMSynced: true,
		RPCProbeOK: 1, RPCProbeTotal: 1,
		HasPMTParams: true, PMTEnabled: true, Inflation: 3.5,
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
		// Use infra RAM pressure: warn (80) → bad (95)
		if step == 1 {
			r.MemPct = 80
		} else {
			r.MemPct = 95
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

func TestEngineAlertMessageDeltaOnly(t *testing.T) {
	sender := &stubSender{}
	step := 0
	load := func() model.Report {
		step++
		r := healthyReport()
		r.NodeRunning = false // steady infra finding
		if step >= 2 {
			r.EVMRPCOk = false // new finding on tick 2
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
		t.Fatalf("expected 2 alerts, got %d", len(sender.msgs))
	}
	if !strings.Contains(sender.msgs[0], "container stopped") {
		t.Fatalf("first alert should be infra: %q", sender.msgs[0])
	}
	if strings.Contains(sender.msgs[1], "container stopped") {
		t.Fatalf("second alert must be delta-only (no dump of active infra): %q", sender.msgs[1])
	}
	if !strings.Contains(sender.msgs[1], "RPC DOWN") {
		t.Fatalf("second alert should mention new RPC finding: %q", sender.msgs[1])
	}
}

func TestEngineRewardsBootstrapNoPage(t *testing.T) {
	sender := &stubSender{}
	load := func() model.Report {
		r := healthyReport()
		r.PMTEnabled = false // already disabled when monitor starts
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
	if len(sender.msgs) != 0 {
		t.Fatalf("steady-state PMT disabled must not page on bootstrap, got %d msgs: %v", len(sender.msgs), sender.msgs)
	}
}

func TestEngineRewardsEnableDisableTransition(t *testing.T) {
	sender := &stubSender{}
	step := 0
	load := func() model.Report {
		step++
		r := healthyReport()
		switch step {
		case 1:
			r.PMTEnabled = true
		case 2:
			r.PMTEnabled = false // transition → page
		case 3:
			r.PMTEnabled = false // steady → no page
		case 4:
			r.PMTEnabled = true // recovery
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
	for i := 0; i < 4; i++ {
		e.tick()
	}
	if len(sender.msgs) != 2 {
		t.Fatalf("expected disable alert + recovery, got %d: %v", len(sender.msgs), sender.msgs)
	}
	if !strings.Contains(sender.msgs[0], "PMT disabled") {
		t.Fatalf("expected disable alert, got %q", sender.msgs[0])
	}
	if !strings.Contains(sender.msgs[1], "resolved") || !strings.Contains(sender.msgs[1], "PMT disabled") {
		t.Fatalf("expected disable recovery, got %q", sender.msgs[1])
	}
}

func TestEngineRewardsPoolEmptyTransition(t *testing.T) {
	sender := &stubSender{}
	step := 0
	load := func() model.Report {
		step++
		r := healthyReport()
		r.PMTEnabled = true
		r.PMTPoolEmpty = step >= 2 && step < 4
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
	for i := 0; i < 4; i++ {
		e.tick()
	}
	if len(sender.msgs) != 2 {
		t.Fatalf("expected empty alert + refill recovery, got %d: %v", len(sender.msgs), sender.msgs)
	}
	if !strings.Contains(sender.msgs[0], "PMT not emitting") {
		t.Fatalf("expected pool-empty alert, got %q", sender.msgs[0])
	}
	if !strings.Contains(sender.msgs[1], "resolved") {
		t.Fatalf("expected refill recovery, got %q", sender.msgs[1])
	}
}

func TestEngineRewardsMissingParamsNoFalseDisable(t *testing.T) {
	sender := &stubSender{}
	step := 0
	load := func() model.Report {
		step++
		r := healthyReport()
		if step == 1 {
			r.HasPMTParams = false // REST down — must not page as disabled
			r.PMTEnabled = false
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
	for _, m := range sender.msgs {
		if strings.Contains(m, "PMT disabled") {
			t.Fatalf("must not alert PMT disabled when params missing: %q", m)
		}
	}
}

func TestEngineNoNotListeningWhenListeningUnknown(t *testing.T) {
	sender := &stubSender{}
	load := func() model.Report {
		r := healthyReport()
		r.EVMRPCOk = false
		r.HasEVMListening = false
		r.EVMListening = false // zero value must not imply not listening
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
	if len(sender.msgs) != 1 {
		t.Fatalf("expected RPC DOWN only, got %d: %v", len(sender.msgs), sender.msgs)
	}
	if strings.Contains(sender.msgs[0], "not listening") {
		t.Fatalf("failed blockNumber must not imply not listening: %q", sender.msgs[0])
	}
	if !strings.Contains(sender.msgs[0], "RPC DOWN") {
		t.Fatalf("expected RPC DOWN: %q", sender.msgs[0])
	}
}

func TestFormatAlertMessage(t *testing.T) {
	active := map[string]panel.Finding{
		findingID(testFinding("infra", "container_stopped", "bad", "container stopped")): testFinding("infra", "container_stopped", "bad", "container stopped"),
		findingID(testFinding("evm", "rpc_down", "bad", "RPC DOWN")):                     testFinding("evm", "rpc_down", "bad", "RPC DOWN"),
	}
	msg := formatAlertMessage("node4", active, model.Report{BlockHeight: "1842032", TimeSinceBlock: "12s"})
	for _, want := range []string{
		"🚨 PMT — node4",
		"Infrastructure · bad",
		"container stopped",
		"EVM JSON-RPC · bad",
		"RPC DOWN",
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
