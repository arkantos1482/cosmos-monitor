package panel

import (
	"fmt"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/feemarket"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

// Finding is a single warn/bad condition for alerting (mirrors section summary badges).
type Finding struct {
	Section  string // infra, evm, node, ...
	Key      string // stable dedup key
	Severity string // warn | bad
	Message  string
}

// Findings evaluates all dashboard sections for warn/bad conditions.
// Does not use the global status strip (status.go).
func Findings(d model.Report) []Finding {
	var out []Finding
	infraFindings(d, &out)
	evmFindings(d, &out)
	nodeFindings(d, &out)
	stakingFindings(d, &out)
	slashingFindings(d, &out)
	feemarketFindings(d, &out)
	rewardsFindings(d, &out)
	distributionFindings(d, &out)
	return out
}

func appendFinding(out *[]Finding, section, key, severity, msg string) {
	if severity == "" || severity == "ok" {
		return
	}
	*out = append(*out, Finding{
		Section:  section,
		Key:      key,
		Severity: severity,
		Message:  msg,
	})
}

func severityFromBadgeClass(cls string) string {
	switch cls {
	case "badge--bad":
		return "bad"
	case "badge--warn":
		return "warn"
	default:
		return ""
	}
}

func infraFindings(d model.Report, out *[]Finding) {
	s := loadInfraState(d)
	if !d.NodeRunning {
		appendFinding(out, "infra", "container_stopped", "bad", "container stopped")
	}
	if d.NodeOOMKilled {
		appendFinding(out, "infra", "oom_killed", "bad", "OOM killed")
	}
	if d.Restarts >= 3 {
		appendFinding(out, "infra", "restarts", "warn", fmt.Sprintf("%d restarts", d.Restarts))
	}
	if tone := infraMeterTone(d.MemPct); tone != "" {
		appendFinding(out, "infra", "host_ram", tone, fmt.Sprintf("host RAM %d%%", d.MemPct))
	}
	if tone := infraMeterTone(s.chainDiskPct); tone != "" {
		appendFinding(out, "infra", "disk", tone, fmt.Sprintf("%s %d%%", s.chainDiskLabel, s.chainDiskPct))
	}
	if tone := infraMeterTone(d.NodeMemPct); tone != "" {
		appendFinding(out, "infra", "container_mem", tone, fmt.Sprintf("container RAM %d%%", d.NodeMemPct))
	}
}

func evmFindings(d model.Report, out *[]Finding) {
	switch evmRPCOverallStatus(d) {
	case "DOWN":
		appendFinding(out, "evm", "rpc_down", "bad", "RPC DOWN")
	case "DEGRADED":
		fail := d.RPCProbeTotal - d.RPCProbeOK
		appendFinding(out, "evm", "rpc_degraded", "warn",
			fmt.Sprintf("RPC DEGRADED (%d/%d probes failing)", fail, d.RPCProbeTotal))
	}
	if !d.EVMListening {
		appendFinding(out, "evm", "not_listening", "warn", "not listening")
	}
	if !d.EVMSynced {
		appendFinding(out, "evm", "syncing", "warn", "syncing")
	}
	if d.EVMBlockAge != "" {
		_, tone := evmBlockAgeKPI(d)
		switch tone {
		case "bad":
			appendFinding(out, "evm", "block_age", "bad", fmt.Sprintf("block age STALE (%s)", d.EVMBlockAge))
		case "warn":
			appendFinding(out, "evm", "block_age", "warn", fmt.Sprintf("block age SLOW (%s)", d.EVMBlockAge))
		}
	}
}

func nodeFindings(d model.Report, out *[]Finding) {
	if !d.Synced {
		appendFinding(out, "node", "catching_up", "warn", "catching up")
	}
}

func stakingFindings(d model.Report, out *[]Finding) {
	for _, b := range localBadges(d) {
		appendFinding(out, "staking", badgeKey(b.text), b.kind, b.text)
	}
}

func slashingFindings(d model.Report, out *[]Finding) {
	lv := d.Local
	for _, b := range localBadges(d) {
		appendFinding(out, "slashing", badgeKey(b.text), b.kind, b.text)
	}
	if lv.IsValidator && lv.MaxMissed > 0 {
		if tone := slashingHeadroomTone(lv); tone != "" {
			appendFinding(out, "slashing", "headroom", tone,
				fmt.Sprintf("headroom low (%d blocks)", slashingHeadroom(lv)))
		}
	}
	if d.JailedCount > 0 {
		appendFinding(out, "slashing", "network_jailed", "bad",
			fmt.Sprintf("%d jailed", d.JailedCount))
	}
	if d.BelowThreshold > 0 {
		appendFinding(out, "slashing", "below_min_signed", "warn",
			fmt.Sprintf("%d below min signed", d.BelowThreshold))
	}
	if d.TombstonedCount > 0 {
		appendFinding(out, "slashing", "network_tombstoned", "bad",
			fmt.Sprintf("%d tombstoned", d.TombstonedCount))
	}
}

func feemarketFindings(d model.Report, out *[]Finding) {
	s := feemarket.LoadState(d)
	if sev := severityFromBadgeClass(s.Adj.BadgeClass()); sev != "" {
		appendFinding(out, "feemarket", "adjustment", sev, "base fee "+s.Adj.Label())
	}
}

func rewardsFindings(d model.Report, out *[]Finding) {
	for _, b := range rewardsSummaryBadges(d) {
		appendFinding(out, "rewards", badgeKey(b.text), b.kind, b.text)
	}
	if d.Inflation <= 0 {
		appendFinding(out, "rewards", "inflation_off", "bad", "inflation off")
	}
}

func distributionFindings(d model.Report, out *[]Finding) {
	effect, warn := distributionEscrowReconcile(d)
	if warn {
		appendFinding(out, "distribution", "escrow_mismatch", "warn", effect)
	}
}

func badgeKey(text string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(text)), " ", "_")
}
