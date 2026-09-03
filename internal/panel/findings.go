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
// Rewards findings are included for current PMT state; the alert engine
// seeds them on first healthy fetch so steady-state does not page.
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

// FindingsWithoutRewards is Findings minus rewards — used by the alert engine
// so rewards can be seeded / transition-tracked separately.
func FindingsWithoutRewards(d model.Report) []Finding {
	var out []Finding
	infraFindings(d, &out)
	evmFindings(d, &out)
	nodeFindings(d, &out)
	stakingFindings(d, &out)
	slashingFindings(d, &out)
	feemarketFindings(d, &out)
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
	if tone := infraMeterTone(d.MemPct); tone != "" {
		appendFinding(out, "infra", "host_ram", tone, fmt.Sprintf("host RAM %d%%", d.MemPct))
	}
	if tone := infraMeterTone(s.chainDiskPct); tone != "" {
		appendFinding(out, "infra", "disk", tone, fmt.Sprintf("%s %d%%", s.chainDiskLabel, s.chainDiskPct))
	}
}

func evmFindings(d model.Report, out *[]Finding) {
	// Telegram/RPC alert signal = liveness only (eth_blockNumber).
	// Degraded probes, syncing, and block age stay UI-only.
	if !d.EVMRPCOk {
		appendFinding(out, "evm", "rpc_down", "bad", "RPC DOWN")
	}
	if d.HasEVMListening && !d.EVMListening {
		appendFinding(out, "evm", "not_listening", "warn", "not listening")
	}
	if strings.TrimSpace(d.JSONRPCAPIs) == "" {
		appendFinding(out, "evm", "jsonrpc_apis_unknown", "warn", "JSON-RPC APIs unknown (app.toml)")
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

// RewardsStateFindings returns current PMT bad-state findings for alert tracking.
// Callers must only page uncommon transitions (enable↔disable, pool empty↔refill);
// steady-state inflation-off / still-empty / still-disabled must not re-page.
// Missing params (HasPMTParams false) yields no findings — fetch fail ≠ chain bad.
func RewardsStateFindings(d model.Report) []Finding {
	var out []Finding
	rewardsFindings(d, &out)
	return out
}

func rewardsFindings(d model.Report, out *[]Finding) {
	if !d.HasPMTParams {
		return
	}
	if !d.PMTEnabled {
		appendFinding(out, "rewards", "pmt_disabled", "bad", "PMT disabled")
		return
	}
	if d.PMTPoolEmpty {
		appendFinding(out, "rewards", "pmt_not_emitting", "warn", "PMT not emitting")
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
