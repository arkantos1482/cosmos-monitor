package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/arkantos1482/cosmos-monitor/fetch"
)

const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiCyan   = "\033[36m"
	ansiWhite  = "\033[97m"
)

func clr(code, s string) string { return code + s + ansiReset }

// runwaySuffix returns a colored " (~N days left)" suffix for terminal display.
func runwaySuffix(d WebData) string {
	if d.PMTRunway == "" {
		return ""
	}
	if d.PMTRunwayLow {
		return fmt.Sprintf("  (%s%s%s)", ansiRed, d.PMTRunway, ansiReset)
	}
	return "  (" + d.PMTRunway + ")"
}

// pmtStatusLine builds the colored PMT rewards status string for the terminal.
func pmtStatusLine(d WebData) string {
	if !d.PMTEnabled {
		return clr(ansiDim, "disabled")
	}
	if d.PMTPoolEmpty {
		return clr(ansiRed, "ENABLED — pool EMPTY  (validators receive no PMT rewards)")
	}
	return clr(ansiGreen, "●") + " distributing  " + d.PMTRate + "   pool " + d.PMTBalance + runwaySuffix(d)
}

func printAll(w io.Writer, d WebData) {
	section    := func(name string)         { fmt.Fprintf(w, "\n%s\n", clr(ansiBold+ansiCyan, name)) }
	subsection := func(name string)         { fmt.Fprintf(w, "  %s\n", clr(ansiYellow, name)) }
	row        := func(label, value string) { fmt.Fprintf(w, "    %s%-22s%s  %s\n", ansiDim, label, ansiReset, value) }

	syncStr := clr(ansiGreen, "synced")
	if !d.Synced {
		syncStr = clr(ansiRed, "CATCHING UP")
	}

	// ── 1. INFRASTRUCTURE ────────────────────────────────────────────────────
	section("1. INFRASTRUCTURE")

	subsection("OS")
	row("load", fmt.Sprintf("%.2f / %.2f / %.2f  (1m 5m 15m)", d.Load1, d.Load5, d.Load15))
	row("ram",  fmt.Sprintf("%s / %s  (%d%%)", d.MemUsed, d.MemTotal, d.MemPct))
	row("disk", fmt.Sprintf("%s / %s  (%d%%)", d.DiskUsed, d.DiskTotal, d.DiskPct))

	subsection("Container")
	status := clr(ansiRed, "stopped")
	if d.NodeRunning {
		status = clr(ansiGreen, "running")
	}
	row("status",   status)
	row("cpu",      d.NodeCPU)
	row("ram",      fmt.Sprintf("%s / %s", d.NodeMemUsed, d.NodeMemTotal))
	row("restarts", fmt.Sprintf("%d", d.Restarts))
	if d.NodeUptime != "" {
		row("uptime", d.NodeUptime)
	}

	// ── 2. NODE ──────────────────────────────────────────────────────────────
	section("2. NODE")

	subsection("Identity")
	row("node ID", d.NodeID)
	row("moniker", d.Moniker)
	if d.AppVersion != "" {
		row("version", d.AppVersion)
	}

	subsection("Block")
	row("height", d.BlockHeight)
	if d.BlockInterval != "" {
		row("interval", d.BlockInterval)
	}
	if d.TimeSinceBlock != "" {
		row("time since last", d.TimeSinceBlock)
	}

	subsection("Sync")
	row("status", syncStr)
	if d.LatestBlockTime != "" {
		row("latest block time", d.LatestBlockTime)
	}

	subsection("Peers")
	row("cosmos peers", fmt.Sprintf("%d", d.PeerCount))
	row("evm peers",    fmt.Sprintf("%d", d.EVMPeerCount))

	// ── 3. VALIDATORS ────────────────────────────────────────────────────────
	section("3. VALIDATORS")

	fmt.Fprintf(w, "  %-20s  %6s  %10s  %10s  %14s  %14s  %8s  %10s  %s\n",
		"moniker", "vp%", "commission", "missed", "outstanding", "earned", "tombstoned", "status", "jailed")
	for _, v := range d.Validators {
		tombStr := "no"
		if v.Tombstoned {
			tombStr = clr(ansiRed, "YES")
		}
		jailed := ""
		if v.Jailed {
			jailed = clr(ansiRed, "JAILED")
		}
		outstanding := v.Outstanding
		if outstanding == "" {
			outstanding = "–"
		}
		earned := v.Earned
		if earned == "" {
			earned = "–"
		}
		fmt.Fprintf(w, "  %-20s  %5.1f%%  %9.1f%%  %10d  %14s  %14s  %8s  %10s  %s\n",
			truncate(v.Moniker, 20),
			v.VPFloat,
			v.CommissionFloat,
			v.Missed,
			outstanding,
			earned,
			tombStr,
			v.Status,
			jailed,
		)
	}

	subsection("Summary")
	row("tombstoned",      fmt.Sprintf("%d", d.TombstonedCount))
	row("below threshold", fmt.Sprintf("%d", d.BelowThreshold))

	// ── 4. ECONOMICS ─────────────────────────────────────────────────────────
	section("4. ECONOMICS")

	subsection("x/staking")
	row("total supply",   d.TotalSupply)
	row("bonded",         fmt.Sprintf("%s  (%.2f%%, goal %.0f%%)", d.BondedAmt, d.BondedPct, d.GoalBonded))
	row("not bonded",     d.NotBonded)
	if d.UnbondingTime != "" {
		row("unbonding time", d.UnbondingTime)
	}
	if d.MaxValidators > 0 {
		row("max validators", fmt.Sprintf("%d", d.MaxValidators))
	}
	if d.BlocksPerYear != "" {
		row("blocks / year", d.BlocksPerYear)
	}

	subsection("x/mint")
	inflationStr := fmt.Sprintf("%.2f%%", d.Inflation)
	if d.Inflation == 0 {
		inflationStr += "  " + clr(ansiYellow, "⚠ inactive") + "  (module present — can be activated via governance)"
	}
	row("inflation", inflationStr)
	if d.BlocksPerYear != "" {
		row("blocks / year", d.BlocksPerYear)
	}

	subsection("x/distribution")
	row("community tax",  d.CommunityTax)
	row("community pool", d.CommunityPool)

	subsection("x/pmtrewards  (custom)")
	row("status", pmtStatusLine(d))
	if d.PMTRate != "" {
		row("rate", d.PMTRate)
		if d.PMTAnnual != "" {
			row("annual emissions", d.PMTAnnual)
		}
	}
	if d.PMTPoolEmpty {
		row("pool balance", clr(ansiRed, "0 PMT  — pool empty"))
	} else if d.PMTBalance != "" {
		row("pool balance", d.PMTBalance+runwaySuffix(d))
	}
	if d.PMTPoolAddress != "" {
		row("pool address", d.PMTPoolAddress)
	}

	subsection("x/slashing  (economic impact)")
	if d.SlashDowntime != "" {
		dtStr := d.SlashDowntime
		if d.SlashDTInactive {
			dtStr += "  " + clr(ansiYellow, "⚠ inactive") + "  (downtime slashing disabled — can be enabled via governance)"
		} else {
			dtStr += "  " + clr(ansiGreen, "active") + "  (validators lose this % of stake for downtime)"
		}
		row("slash / downtime", dtStr)
	}
	if d.SlashDS != "" {
		dsStr := d.SlashDS
		if d.SlashDSInactive {
			dsStr += "  " + clr(ansiYellow, "⚠ inactive") + "  (double-sign slashing disabled — can be enabled via governance)"
		} else {
			dsStr += "  " + clr(ansiGreen, "active") + "  (validators lose this % of stake for double-sign)"
		}
		row("slash / double-sign", dsStr)
	}

	if d.PMTInsights {
		subsection("Insights")
		if d.PMTPoolEmpty {
			row("pool runway",     clr(ansiRed, "EMPTY")+"  (no PMT rewards distributing)")
			row("daily emissions", d.PMTDailyEmit)
		} else {
			if d.PMTRunwayLow {
				row("pool runway", clr(ansiRed, d.PMTRunwayDays))
			} else {
				row("pool runway", d.PMTRunwayDays)
			}
			row("daily emissions", d.PMTDailyEmit)
			if d.PMTPerValDay != "" {
				row("per validator/day", d.PMTPerValDay)
			}
		}
		if d.PMTRevFlow != "" {
			row("revenue flow", d.PMTRevFlow)
			fmt.Fprintf(w, "    %-22s    ├─ %.0f%%  commission → validator\n", "", d.PMTCommPct)
			fmt.Fprintf(w, "    %-22s    └─ %.0f%%  → delegators (pro-rata)\n", "", d.PMTDelegPct)
		}
	}

	subsection("Validator Earnings  (unclaimed)")
	if d.TotalOutstanding != "" {
		row("total outstanding", d.TotalOutstanding)
	}
	row("commission rate", fmt.Sprintf("%.0f%%  of staking rewards", d.CommissionRate))

	// ── 5. EVM ────────────────────────────────────────────────────────────────
	section("5. EVM")

	evSyncStr := clr(ansiGreen, "synced")
	if !d.EVMSynced {
		evSyncStr = clr(ansiYellow, "syncing")
	}

	subsection("Identity")
	row("chain ID", fmt.Sprintf("%d", d.EVMChainID))
	if d.EVMDenom != "" {
		row("denom", d.EVMDenom)
	}
	if d.EVMClient != "" {
		row("client", d.EVMClient)
	}

	subsection("Health")
	if d.EVMRPCOk {
		row("rpc", clr(ansiGreen, "ok")+"  (eth_blockNumber responded)")
	} else {
		row("rpc", clr(ansiRed, "error"))
	}
	listeningStr := clr(ansiGreen, "yes")
	if !d.EVMListening {
		listeningStr = clr(ansiYellow, "no")
	}
	row("net listening", listeningStr)
	if d.EVMBlockAge != "" {
		ageStr := d.EVMBlockAge
		if d.EVMBlockAgeErr {
			ageStr = clr(ansiRed, ageStr) + "  ⚠ stalled"
		} else if d.EVMBlockAgeWarn {
			ageStr = clr(ansiYellow, ageStr) + "  ⚠ slow"
		}
		row("last block age", ageStr)
	}
	row("sync",   evSyncStr)
	row("txpool", fmt.Sprintf("pending %d   queued %d", d.PendingTx, d.QueuedTx))

	subsection("Block")
	row("block", d.EVMBlock)
	row("sync",  evSyncStr)

	subsection("Gas")
	if d.BaseFee != "" {
		row("base fee", d.BaseFee+" wei")
	}
	if d.GasPrice != "" {
		row("gas price", d.GasPrice)
	}
	if d.MinGasPrice != "" {
		row("min gas price", d.MinGasPrice)
	}

	subsection("Fee Market Mechanics")
	row("model", "EIP-1559  (dynamic base fee — adjusts to target block utilization)")
	if d.BaseFee != "" {
		row("base fee", d.BaseFee+" wei")
		if d.AdjCap != "" {
			row("adjustment cap", d.AdjCap)
		}
	}
	if d.Elasticity > 0 {
		row("block target", fmt.Sprintf("max_gas ÷ %d  (rises if > 1/%d full, falls if < 1/%d full)", d.Elasticity, d.Elasticity, d.Elasticity))
	}
	noBaseFeeStr := boolStr(d.NoBaseFee) + "  (fee enforcement "
	if d.NoBaseFee {
		noBaseFeeStr += clr(ansiYellow, "disabled")
	} else {
		noBaseFeeStr += clr(ansiGreen, "active")
	}
	noBaseFeeStr += ")"
	row("no_base_fee", noBaseFeeStr)
	feeDestStr := "validators"
	if d.CommunityTaxZero {
		feeDestStr += "  (0% community tax — fees are not burned)"
	}
	row("fee destination", feeDestStr)

	if len(d.Precompiles) > 0 {
		subsection("Precompiles")
		row("active", fmt.Sprintf("%d", len(d.Precompiles)))
		for _, pc := range d.Precompiles {
			fmt.Fprintf(w, "    %s\n", pc)
		}
	}

	subsection("Config")
	if d.HistoryWindow != "" {
		row("history serve window", d.HistoryWindow)
	}
	row("ERC20 enabled", boolStr(d.ERC20Enabled))
	if d.HardforkLondon != "" {
		row("London height", d.HardforkLondon)
	}
	if d.HardforkShanghai != "" {
		row("Shanghai height", d.HardforkShanghai)
	}
	if d.HardforkCancun != "" {
		row("Cancun height", d.HardforkCancun)
	}

	// ── 6. CHAIN ──────────────────────────────────────────────────────────────
	section("6. CHAIN")

	subsection("Governance")
	row("voting period", d.VotingPeriod)
	row("quorum",        fmt.Sprintf("%.1f%%", d.Quorum))
	row("threshold",     fmt.Sprintf("%.1f%%", d.Threshold))
	if d.VetoThreshold > 0 {
		row("veto threshold", fmt.Sprintf("%.1f%%", d.VetoThreshold))
	}

	if len(d.Proposals) > 0 {
		subsection("Active Proposals (voting period)")
		for _, pr := range d.Proposals {
			fmt.Fprintf(w, "  #%-4d  %-40s  ends %s\n", pr.ID, truncate(pr.Title, 40), pr.End)
			if pr.HasTally {
				fmt.Fprintf(w, "         yes %-14s  no %-14s  abstain %-14s  veto %s\n",
					pr.TallyYes, pr.TallyNo, pr.TallyAbstain, pr.TallyVeto)
			}
		}
	}

	if len(d.DepositProposals) > 0 {
		subsection("Deposit-Period Proposals")
		for _, pr := range d.DepositProposals {
			fmt.Fprintf(w, "  #%-4d  %-40s  deposit ends %s\n", pr.ID, truncate(pr.Title, 40), pr.End)
		}
	}

	if len(d.Proposals)+len(d.DepositProposals) == 0 {
		fmt.Fprintln(w, "  none active")
	}

	subsection("Slashing")
	if d.SlashWindow != "" && d.SlashWindow != "0" {
		row("window", d.SlashWindow+" blocks")
	}
	if d.MinSigned > 0 {
		row("min signed", fmt.Sprintf("%.1f%%", d.MinSigned))
	}
	if d.SlashDowntime != "" {
		row("slash / downtime", d.SlashDowntime)
	}
	if d.SlashDS != "" {
		row("slash / double sign", d.SlashDS)
	}

	subsection("Upgrade")
	if d.UpgradeName == "" {
		row("pending", "none")
	} else {
		row("name",          d.UpgradeName)
		row("target height", d.UpgradeHeight)
		if d.BlocksLeft != "" {
			row("blocks remaining", d.BlocksLeft)
		}
	}

	subsection("IBC")
	row("active clients", fmt.Sprintf("%d", d.IBCClients))

	subsection(fmt.Sprintf("Token Pairs  (%d)", len(d.TokenPairs)))
	if len(d.TokenPairs) == 0 {
		fmt.Fprintln(w, "    none registered")
	}
	for _, tp := range d.TokenPairs {
		enabled := "yes"
		if !tp.Enabled {
			enabled = "no"
		}
		fmt.Fprintf(w, "    %-30s  %-42s  %s\n", tp.Denom, tp.ERC20, enabled)
	}

	fmt.Fprintln(w)
}

func printEssentials(w io.Writer, d WebData) {
	sec := func(name, summary string) {
		if summary != "" {
			fmt.Fprintf(w, "\n%s %s   %s\n", clr(ansiDim, "──"), clr(ansiBold+ansiCyan, name), summary)
		} else {
			fmt.Fprintf(w, "\n%s %s\n", clr(ansiDim, "──"), clr(ansiBold+ansiCyan, name))
		}
	}
	line := func(parts ...string) { fmt.Fprintf(w, "  %s\n", strings.Join(parts, "   ")) }
	kv   := func(k, v string) string { return clr(ansiDim, k) + " " + v }

	// ── header
	syncStr := clr(ansiGreen, "● synced")
	if !d.Synced {
		syncStr = clr(ansiRed, "● CATCHING UP")
	}
	fmt.Fprintf(w, "%s  %s  %s  %s UTC\n",
		clr(ansiBold+ansiWhite, "pmtop  "+d.Moniker),
		syncStr, d.BlockHeight, time.Now().UTC().Format("15:04:05"))

	// ── SYSTEM
	sec("SYSTEM", "")
	nodeStatus := clr(ansiRed, "stopped")
	if d.NodeRunning {
		nodeStatus = clr(ansiGreen, "● running")
	}
	uptimeStr := ""
	if d.NodeUptime != "" {
		uptimeStr = kv("up", d.NodeUptime)
	}
	line(nodeStatus,
		kv("cpu", d.NodeCPU),
		kv("ram", fmt.Sprintf("%d%%", d.MemPct)),
		kv("disk", fmt.Sprintf("%d%%", d.DiskPct)),
		kv("restarts", fmt.Sprintf("%d", d.Restarts)),
		uptimeStr)
	line(kv("load", fmt.Sprintf("%.2f/%.2f/%.2f", d.Load1, d.Load5, d.Load15)),
		kv("ram", fmt.Sprintf("%s/%s", d.MemUsed, d.MemTotal)),
		kv("disk", fmt.Sprintf("%s/%s", d.DiskUsed, d.DiskTotal)),
		kv("height", d.BlockHeight),
		kv("interval", d.BlockInterval),
		kv("peers", fmt.Sprintf("%d/%d", d.PeerCount, d.EVMPeerCount)))

	// ── VALIDATORS
	bondedCount, jailedCount := 0, 0
	for _, v := range d.Validators {
		if v.Status == "bonded" {
			bondedCount++
		}
		if v.Jailed {
			jailedCount++
		}
	}
	totalVals := len(d.Validators)
	bondedStr := clr(ansiGreen, fmt.Sprintf("%d/%d bonded", bondedCount, totalVals))
	if bondedCount < totalVals {
		bondedStr = clr(ansiRed, fmt.Sprintf("%d/%d bonded", bondedCount, totalVals))
	}
	jailedStr := fmt.Sprintf("%d jailed", jailedCount)
	if jailedCount > 0 {
		jailedStr = clr(ansiRed, jailedStr)
	}
	belowStr := fmt.Sprintf("%d below threshold", d.BelowThreshold)
	if d.BelowThreshold > 0 {
		belowStr = clr(ansiYellow, belowStr)
	}
	sec("VALIDATORS", bondedStr+"   "+jailedStr+"   "+belowStr)

	for _, v := range d.Validators {
		flags := ""
		if v.Jailed {
			flags += "  " + clr(ansiRed, "JAILED")
		}
		if v.Tombstoned {
			flags += "  " + clr(ansiRed, "TOMBSTONED")
		}
		missedStr := fmt.Sprintf("%d missed", v.Missed)
		if v.MissedHigh {
			missedStr = clr(ansiRed, missedStr)
		}
		earned := v.Outstanding
		if earned == "" {
			earned = "–"
		}
		fmt.Fprintf(w, "  %-16s  %5.1f%%  %-14s  %s  earned %s%s\n",
			truncate(v.Moniker, 16), v.VPFloat,
			missedStr, v.Status, earned, flags)
	}

	// ── ECONOMICS
	sec("ECONOMICS", "")
	line(kv("supply", d.TotalSupply),
		kv("bonded", fmt.Sprintf("%s  %.1f%%  (goal %.0f%%)", d.BondedAmt, d.BondedPct, d.GoalBonded)))
	line(kv("rewards", pmtStatusLine(d)))
	fmt.Fprintln(w)
}

func printDashboard(w io.Writer, chain fetch.ChainSnapshot, ev fetch.EVMSnapshot, sys fetch.SystemSnapshot, docker fetch.DockerSnapshot) {
	d := buildWebData(chain, ev, sys, docker)
	printEssentials(w, d)
	printAll(w, d)
}

// ── helpers ──────────────────────────────────────────────────────────────────

func fmtInt(n int64) string {
	if n <= 0 {
		return fmt.Sprintf("%d", n)
	}
	s := fmt.Sprintf("%d", n)
	out := make([]byte, 0, len(s)+4)
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(c))
	}
	return string(out)
}

func fmtBytes(b uint64) string {
	const (KB = 1024; MB = 1024 * KB; GB = 1024 * MB)
	switch {
	case b >= GB:
		return fmt.Sprintf("%.1f GB", float64(b)/GB)
	case b >= MB:
		return fmt.Sprintf("%.1f MB", float64(b)/MB)
	case b >= KB:
		return fmt.Sprintf("%.1f KB", float64(b)/KB)
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func fmtPct(used, total uint64) string {
	if total == 0 {
		return "0"
	}
	return fmt.Sprintf("%.0f", float64(used)/float64(total)*100)
}

func fmtDur(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func fmtDurFull(d time.Duration) string {
	if d == 0 {
		return "0"
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}

func fmtUptime(t time.Time) string {
	return fmtDurFull(time.Since(t))
}

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func fmtFraction(s string) string {
	v := 0.0
	fmt.Sscanf(s, "%f", &v)
	return fmt.Sprintf("%.4f%%", v*100)
}
