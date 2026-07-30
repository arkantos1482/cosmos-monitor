package alert

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/panel"
)

// LoadFunc fetches a dashboard report for evaluation.
type LoadFunc func() model.Report

// Engine polls findings and sends Telegram alerts with cooldown and recovery.
type Engine struct {
	cfg    Config
	load   LoadFunc
	sender Sender

	active   map[string]panel.Finding
	lastFire map[string]time.Time

	// rewardsSeeded is set after the first tick with HasPMTParams so a
	// pre-existing disabled/empty pool does not page as a "transition".
	rewardsSeeded bool
}

func NewEngine(cfg Config, load LoadFunc, sender Sender) *Engine {
	return &Engine{
		cfg:      cfg,
		load:     load,
		sender:   sender,
		active:   map[string]panel.Finding{},
		lastFire: map[string]time.Time{},
	}
}

// Run blocks, polling at cfg.Interval until the process exits.
func (e *Engine) Run() {
	if !e.cfg.Enabled() && !e.cfg.DryRun {
		log.Println("alert: disabled (set PMTOP_TELEGRAM_TOKEN and PMTOP_TELEGRAM_CHAT_ID)")
		return
	}
	log.Printf("alert: polling every %s (cooldown %s, node %s)", e.cfg.Interval, e.cfg.Cooldown, e.nodeLabel())
	ticker := time.NewTicker(e.cfg.Interval)
	defer ticker.Stop()
	e.tick()
	for range ticker.C {
		e.tick()
	}
}

func (e *Engine) tick() {
	d := e.load()
	now := time.Now()
	current := e.indexCurrentFindings(d)

	var delta []panel.Finding
	var recoveries []panel.Finding

	for id, f := range current {
		prev, was := e.active[id]
		if !was {
			if e.cooldownReady(id, now) {
				delta = append(delta, f)
				e.lastFire[id] = now
			}
			continue
		}
		if severityRank(f.Severity) > severityRank(prev.Severity) {
			if e.cooldownReady(id, now) {
				delta = append(delta, f)
				e.lastFire[id] = now
			}
		}
	}

	for id, prev := range e.active {
		if _, ok := current[id]; ok {
			continue
		}
		if e.cooldownReady(id, now) {
			recoveries = append(recoveries, prev)
			e.lastFire[id] = now
		}
	}

	e.active = current

	if len(delta) > 0 {
		msg := formatAlertMessage(e.nodeLabel(), indexFindings(delta), d)
		e.send(msg, "alert")
	}
	for _, f := range recoveries {
		msg := formatRecoveryMessage(e.nodeLabel(), f)
		e.send(msg, "recovery")
	}
}

// indexCurrentFindings builds the active finding set for this tick.
// Rewards findings are seeded on the first healthy PMT params fetch so
// steady-state disabled/empty does not page; later enable↔disable and
// pool empty↔refill still fire as new/cleared findings.
func (e *Engine) indexCurrentFindings(d model.Report) map[string]panel.Finding {
	current := indexFindings(panel.FindingsWithoutRewards(d))

	if !d.HasPMTParams {
		// Fetch fail ≠ chain recovered — keep prior rewards findings sticky.
		for id, f := range e.active {
			if f.Section == "rewards" {
				current[id] = f
			}
		}
		return current
	}

	rewards := panel.RewardsStateFindings(d)
	if !e.rewardsSeeded {
		for _, f := range rewards {
			id := findingID(f)
			e.active[id] = f
		}
		e.rewardsSeeded = true
	}
	for _, f := range rewards {
		current[findingID(f)] = f
	}
	return current
}

func (e *Engine) cooldownReady(id string, now time.Time) bool {
	last, ok := e.lastFire[id]
	if !ok {
		return true
	}
	return now.Sub(last) >= e.cfg.Cooldown
}

func (e *Engine) send(text, kind string) {
	if e.cfg.DryRun {
		log.Printf("alert dry-run (%s):\n%s", kind, text)
		return
	}
	if e.sender == nil {
		return
	}
	if err := e.sender.Send(text); err != nil {
		log.Printf("alert: %s send failed: %v", kind, err)
	}
}

func (e *Engine) nodeLabel() string {
	if e.cfg.NodeName != "" {
		return e.cfg.NodeName
	}
	return "pmtop"
}

func indexFindings(fs []panel.Finding) map[string]panel.Finding {
	m := make(map[string]panel.Finding, len(fs))
	for _, f := range fs {
		m[findingID(f)] = f
	}
	return m
}

func findingID(f panel.Finding) string {
	return f.Section + "\x00" + f.Key
}

func severityRank(s string) int {
	switch s {
	case "bad":
		return 2
	case "warn":
		return 1
	default:
		return 0
	}
}

var sectionTitles = map[string]string{
	"infra":        "Infrastructure",
	"evm":          "EVM JSON-RPC",
	"node":         "Validator",
	"staking":      "Staking",
	"slashing":     "Slashing",
	"feemarket":    "Fee market",
	"rewards":      "Rewards",
	"distribution": "Distribution",
}

func formatAlertMessage(node string, active map[string]panel.Finding, d model.Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "🚨 PMT — %s\n\n", escapeHTML(node))

	sections := groupedFindings(active)
	for _, sec := range sections {
		title := sectionTitles[sec]
		if title == "" {
			title = sec
		}
		worst := worstSeverity(active, sec)
		fmt.Fprintf(&b, "%s · %s\n", escapeHTML(title), escapeHTML(worst))
		for _, f := range sortedSectionFindings(active, sec) {
			fmt.Fprintf(&b, "  • %s\n", escapeHTML(f.Message))
		}
		b.WriteByte('\n')
	}

	if d.BlockHeight != "" || d.TimeSinceBlock != "" {
		height := strings.TrimSpace(d.BlockHeight)
		ago := strings.TrimSpace(d.TimeSinceBlock)
		switch {
		case height != "" && ago != "":
			fmt.Fprintf(&b, "Height %s · %s ago", escapeHTML(height), escapeHTML(ago))
		case height != "":
			fmt.Fprintf(&b, "Height %s", escapeHTML(height))
		case ago != "":
			fmt.Fprintf(&b, "Last block %s ago", escapeHTML(ago))
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func formatRecoveryMessage(node string, f panel.Finding) string {
	title := sectionTitles[f.Section]
	if title == "" {
		title = f.Section
	}
	return fmt.Sprintf("✅ PMT — %s resolved\n\n%s · %s",
		escapeHTML(node), escapeHTML(title), escapeHTML(f.Message))
}

func groupedFindings(active map[string]panel.Finding) []string {
	seen := map[string]bool{}
	var out []string
	for _, f := range active {
		if seen[f.Section] {
			continue
		}
		seen[f.Section] = true
		out = append(out, f.Section)
	}
	sort.Slice(out, func(i, j int) bool {
		ri, rj := sectionOrder(out[i]), sectionOrder(out[j])
		if ri != rj {
			return ri < rj
		}
		return out[i] < out[j]
	})
	return out
}

func sectionOrder(sec string) int {
	order := []string{"infra", "evm", "node", "staking", "slashing", "feemarket", "rewards", "distribution"}
	for i, s := range order {
		if s == sec {
			return i
		}
	}
	return len(order)
}

func sortedSectionFindings(active map[string]panel.Finding, sec string) []panel.Finding {
	var out []panel.Finding
	for _, f := range active {
		if f.Section == sec {
			out = append(out, f)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Severity != out[j].Severity {
			return severityRank(out[i].Severity) > severityRank(out[j].Severity)
		}
		return out[i].Message < out[j].Message
	})
	return out
}

func worstSeverity(active map[string]panel.Finding, sec string) string {
	worst := ""
	for _, f := range active {
		if f.Section != sec {
			continue
		}
		if severityRank(f.Severity) > severityRank(worst) {
			worst = f.Severity
		}
	}
	return worst
}
