package main

import (
	"fmt"
	"io"
	"sort"
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

func printAll(w io.Writer, chain fetch.ChainSnapshot, ev fetch.EVMSnapshot, sys fetch.SystemSnapshot, docker fetch.DockerSnapshot) {
	section    := func(name string)         { fmt.Fprintf(w, "\n%s\n", clr(ansiBold+ansiCyan, name)) }
	subsection := func(name string)         { fmt.Fprintf(w, "  %s\n", clr(ansiYellow, name)) }
	row        := func(label, value string) { fmt.Fprintf(w, "    %s%-20s%s  %s\n", ansiDim, label, ansiReset, value) }

	p := chain.Params

	syncStr := clr(ansiGreen, "synced")
	if chain.CatchingUp {
		syncStr = clr(ansiRed, "CATCHING UP")
	}
	fmt.Fprintf(w, "%s  %s  height %s  %s UTC\n",
		clr(ansiBold+ansiWhite, "pmtop  "+chain.Moniker),
		syncStr, fmtInt(chain.BlockHeight), time.Now().UTC().Format("15:04:05"))

	// ── 1. HEALTH ────────────────────────────────────────────────────────────
	section("1. HEALTH")

	subsection("OS")
	memUsed := uint64(0)
	if sys.MemTotal > sys.MemAvail {
		memUsed = sys.MemTotal - sys.MemAvail
	}
	row("load", fmt.Sprintf("%.2f / %.2f / %.2f  (1m 5m 15m)", sys.LoadAvg1, sys.LoadAvg5, sys.LoadAvg15))
	row("ram",  fmt.Sprintf("%s / %s  (%s%%)", fmtBytes(memUsed), fmtBytes(sys.MemTotal), fmtPct(memUsed, sys.MemTotal)))
	row("disk", fmt.Sprintf("%s / %s  (%s%%)", fmtBytes(sys.DiskUsed), fmtBytes(sys.DiskTotal), fmtPct(sys.DiskUsed, sys.DiskTotal)))

	subsection("Container")
	status := clr(ansiRed, "stopped")
	if docker.Running {
		status = clr(ansiGreen, "running")
	}
	row("status",   status)
	row("cpu",      fmt.Sprintf("%.1f%%", docker.CPUPercent))
	row("ram",      fmt.Sprintf("%s / %s", fmtBytes(docker.MemUsage), fmtBytes(docker.MemLimit)))
	row("restarts", fmt.Sprintf("%d", docker.RestartCount))
	if !docker.StartedAt.IsZero() {
		row("uptime", fmtUptime(docker.StartedAt))
	}

	// ── 2. NODE ───────────────────────────────────────────────────────────────
	section("2. NODE")

	subsection("Identity")
	row("node ID", chain.NodeID)
	row("moniker", chain.Moniker)
	row("version", chain.AppVersion)

	subsection("Block")
	row("height",   fmtInt(chain.BlockHeight))
	row("interval", fmtDur(chain.BlockInterval))
	if !chain.LatestBlockTime.IsZero() {
		row("time since last", fmtDur(time.Since(chain.LatestBlockTime)))
	}

	subsection("Sync")
	row("status", syncStr)
	if !chain.LatestBlockTime.IsZero() {
		row("latest block time", chain.LatestBlockTime.UTC().Format("2006-01-02 15:04:05 UTC"))
	}

	subsection("Peers")
	row("cosmos peers", fmt.Sprintf("%d", chain.PeerCount))
	row("evm peers",    fmt.Sprintf("%d", ev.PeerCount))

	// ── 3. TOKENOMICS ─────────────────────────────────────────────────────────
	section("3. TOKENOMICS")

	denom := p.BondDenom
	if denom == "" {
		denom = chain.TotalSupplyDenom
	}
	bondedF, _ := fetch.NormalizeCoin(chain.BondedTokens, denom)
	totalF, _  := fetch.NormalizeCoin(chain.TotalSupply, chain.TotalSupplyDenom)
	bondPct := 0.0
	if totalF > 0 {
		bondPct = bondedF / totalF * 100
	}

	subsection("Supply")
	row("total supply", fetch.FormatCoin(chain.TotalSupply, chain.TotalSupplyDenom))
	row("bonded",       fetch.FormatCoin(chain.BondedTokens, denom))
	row("not bonded",   fetch.FormatCoin(chain.NotBondedTokens, denom))
	row("bonded %",     fmt.Sprintf("%.2f%%", bondPct))

	subsection("Inflation")
	row("rate",        fmt.Sprintf("%.4f%%", chain.Inflation*100))
	row("goal bonded", fmt.Sprintf("%.1f%%", p.GoalBonded*100))
	if p.BlocksPerYear > 0 {
		row("blocks / year", fmtInt(p.BlocksPerYear))
	}
	if chain.AnnualProvisions != "" && chain.AnnualProvisions != "0" {
		row("annual provisions", fetch.FormatCoin(chain.AnnualProvisions, denom))
	}

	subsection("Staking Params")
	row("unbonding time", fmtDurFull(p.UnbondingTime))
	if p.MaxValidators > 0 {
		row("max validators", fmt.Sprintf("%d", p.MaxValidators))
	}
	if denom != "" {
		row("bond denom", denom)
	}

	subsection("Distribution")
	row("community pool", chain.CommunityPool)
	row("community tax",  fmt.Sprintf("%.4f%%", p.CommunityTax*100))

	subsection("PMT Rewards")
	row("enabled", boolStr(p.PMTRewardsEnabled))
	if p.RewardPerBlockAmount != "" && p.RewardPerBlockDenom != "" {
		row("reward / block", fetch.FormatCoin(p.RewardPerBlockAmount, p.RewardPerBlockDenom))
	}
	if p.PMTRewardsPoolAddress != "" {
		row("pool address", p.PMTRewardsPoolAddress)
	}

	// ── 4. EVM ────────────────────────────────────────────────────────────────
	section("4. EVM")

	evSyncStr := clr(ansiGreen, "synced")
	if ev.Syncing {
		evSyncStr = clr(ansiYellow, "syncing")
	}

	subsection("Identity")
	row("chain ID", fmt.Sprintf("%d", ev.ChainID))
	if p.EVMDenom != "" {
		row("denom", p.EVMDenom)
	}

	subsection("Block")
	row("block", fmtInt(int64(ev.BlockNumber)))
	row("sync",  evSyncStr)

	subsection("Gas")
	if chain.BaseFee != "" {
		row("base fee", chain.BaseFee+" wei")
	}
	if ev.GasPrice != "" {
		row("gas price", ev.GasPrice)
	}
	if p.MinGasPrice > 0 {
		row("min gas price", fmt.Sprintf("%.9f %s", p.MinGasPrice, denom))
	}

	subsection("Fee Market")
	if p.Elasticity > 0 {
		row("elasticity", fmt.Sprintf("%d", p.Elasticity))
	}
	if p.BaseFeeChangeDenominator > 0 {
		row("change denominator", fmt.Sprintf("%d", p.BaseFeeChangeDenominator))
	}
	row("no_base_fee", boolStr(p.NoBaseFee))

	subsection("Txpool")
	row("pending", fmt.Sprintf("%d", ev.PendingTx))
	row("queued",  fmt.Sprintf("%d", ev.QueuedTx))

	if len(p.ActiveStaticPrecompiles) > 0 {
		subsection("Precompiles")
		row("active", fmt.Sprintf("%d", len(p.ActiveStaticPrecompiles)))
		for _, pc := range p.ActiveStaticPrecompiles {
			fmt.Fprintf(w, "    %s\n", pc)
		}
	}

	subsection("Config")
	if p.HistoryServeWindow > 0 {
		row("history serve window", fmtInt(p.HistoryServeWindow))
	}
	row("ERC20 enabled", boolStr(p.ERC20Enabled))
	if p.HardforkLondon != "" {
		row("London height", p.HardforkLondon)
	}
	if p.HardforkShanghai != "" {
		row("Shanghai height", p.HardforkShanghai)
	}
	if p.HardforkCancun != "" {
		row("Cancun height", p.HardforkCancun)
	}

	// ── 5. TOKEN PAIRS ────────────────────────────────────────────────────────
	section(fmt.Sprintf("5. TOKEN PAIRS  (%d)", len(chain.TokenPairs)))
	if len(chain.TokenPairs) == 0 {
		fmt.Fprintln(w, "  none registered")
	}
	for _, tp := range chain.TokenPairs {
		enabled := "yes"
		if !tp.Enabled {
			enabled = "no"
		}
		fmt.Fprintf(w, "  %-30s  %-42s  %s\n", tp.Denom, tp.ERC20Addr, enabled)
	}

	// ── 6. VALIDATORS ─────────────────────────────────────────────────────────
	section("6. VALIDATORS")
	vals := make([]fetch.ValidatorInfo, len(chain.Validators))
	copy(vals, chain.Validators)
	sort.Slice(vals, func(i, j int) bool {
		return vals[i].VotingPowerPercent > vals[j].VotingPowerPercent
	})
	fmt.Fprintf(w, "  %-20s  %6s  %10s  %10s  %14s  %14s  %8s  %10s  %s\n",
		"moniker", "vp%", "commission", "missed", "outstanding", "earned", "tombstoned", "status", "jailed")
	for _, v := range vals {
		tombStr := "no"
		if v.Tombstoned {
			tombStr = clr(ansiRed, "YES")
		}
		st := strings.ToLower(v.Status)
		jailed := ""
		if v.Jailed {
			jailed = clr(ansiRed, "JAILED")
		}
		outstanding := v.OutstandingRewards
		if outstanding == "" {
			outstanding = "–"
		}
		earned := v.CommissionEarned
		if earned == "" {
			earned = "–"
		}
		fmt.Fprintf(w, "  %-20s  %5.1f%%  %9.1f%%  %10d  %14s  %14s  %8s  %10s  %s\n",
			truncate(v.Moniker, 20),
			v.VotingPowerPercent,
			v.Commission*100,
			v.MissedBlocks,
			outstanding,
			earned,
			tombStr,
			st,
			jailed,
		)
	}

	// ── 7. SECURITY ───────────────────────────────────────────────────────────
	section("7. SECURITY")

	subsection("Slashing Params")
	if p.SignedBlocksWindow > 0 {
		row("window", fmt.Sprintf("%s blocks", fmtInt(p.SignedBlocksWindow)))
	}
	if p.MinSignedPerWindow > 0 {
		row("min signed", fmt.Sprintf("%.1f%%", p.MinSignedPerWindow*100))
	}
	if p.SlashFractionDowntime != "" {
		row("slash / downtime", fmtFraction(p.SlashFractionDowntime))
	}
	if p.SlashFractionDoubleSign != "" {
		row("slash / double sign", fmtFraction(p.SlashFractionDoubleSign))
	}

	subsection("Summary")
	tombCount := 0
	belowThreshold := 0
	for _, v := range chain.Validators {
		if v.Tombstoned {
			tombCount++
		}
		if v.Status == "BONDED" && !v.Tombstoned && p.SignedBlocksWindow > 0 {
			maxMissed := int64(float64(p.SignedBlocksWindow) * (1 - p.MinSignedPerWindow))
			if v.MissedBlocks > maxMissed {
				belowThreshold++
			}
		}
	}
	row("tombstoned",      fmt.Sprintf("%d", tombCount))
	row("below threshold", fmt.Sprintf("%d", belowThreshold))

	// ── 8. GOVERNANCE ─────────────────────────────────────────────────────────
	section("8. GOVERNANCE")

	subsection("Params")
	row("voting period", fmtDurFull(p.VotingPeriod))
	row("quorum",        fmt.Sprintf("%.1f%%", p.Quorum*100))
	row("threshold",     fmt.Sprintf("%.1f%%", p.Threshold*100))
	if p.VetoThreshold > 0 {
		row("veto threshold", fmt.Sprintf("%.1f%%", p.VetoThreshold*100))
	}

	if len(chain.VotingProposals) > 0 {
		subsection("Active Proposals (voting period)")
		for _, pr := range chain.VotingProposals {
			fmt.Fprintf(w, "  #%-4d  %-40s  ends %s\n",
				pr.ID, truncate(pr.Title, 40), pr.VotingEnd.Format("2006-01-02"))
			t := pr.Tally
			if t.Yes != "" || t.No != "" || t.Abstain != "" || t.NoWithVeto != "" {
				fmt.Fprintf(w, "         yes %-14s  no %-14s  abstain %-14s  veto %s\n",
					t.Yes, t.No, t.Abstain, t.NoWithVeto)
			}
		}
	}

	if len(chain.DepositProposals) > 0 {
		subsection("Deposit-Period Proposals")
		for _, pr := range chain.DepositProposals {
			fmt.Fprintf(w, "  #%-4d  %-40s  deposit ends %s\n",
				pr.ID, truncate(pr.Title, 40), pr.DepositEnd.Format("2006-01-02"))
		}
	}

	if len(chain.VotingProposals)+len(chain.DepositProposals) == 0 {
		fmt.Fprintln(w, "  none active")
	}

	// ── 9. UPGRADE ────────────────────────────────────────────────────────────
	section("9. UPGRADE")
	if chain.UpgradeName == "" {
		row("pending", "none")
	} else {
		row("name",   chain.UpgradeName)
		row("target height", fmtInt(chain.UpgradeHeight))
		if chain.BlockHeight > 0 && chain.UpgradeHeight > chain.BlockHeight {
			row("blocks remaining", fmtInt(chain.UpgradeHeight-chain.BlockHeight))
		}
	}

	// ── 10. IBC ───────────────────────────────────────────────────────────────
	section("10. IBC")
	row("active clients", fmt.Sprintf("%d", chain.IBCClientCount))

	fmt.Fprintln(w)
}

func printEssentials(w io.Writer, chain fetch.ChainSnapshot, ev fetch.EVMSnapshot, sys fetch.SystemSnapshot, docker fetch.DockerSnapshot) {
	row     := func(label, value string) { fmt.Fprintf(w, "  %s%-12s%s  %s\n", ansiDim, label, ansiReset, value) }
	section := func(name string)         { fmt.Fprintf(w, "\n%s\n", clr(ansiBold+ansiCyan, name)) }

	syncStr := clr(ansiGreen, "synced")
	if chain.CatchingUp {
		syncStr = clr(ansiRed, "CATCHING UP")
	}
	fmt.Fprintf(w, "%s  %s  height %s  %s UTC\n",
		clr(ansiBold+ansiWhite, "pmtop  "+chain.Moniker),
		syncStr, fmtInt(chain.BlockHeight), time.Now().UTC().Format("15:04:05"))

	section("HEALTH")
	memUsed := uint64(0)
	if sys.MemTotal > sys.MemAvail {
		memUsed = sys.MemTotal - sys.MemAvail
	}
	row("load", fmt.Sprintf("%.2f / %.2f / %.2f", sys.LoadAvg1, sys.LoadAvg5, sys.LoadAvg15))
	row("ram",  fmt.Sprintf("%s / %s  (%s%%)", fmtBytes(memUsed), fmtBytes(sys.MemTotal), fmtPct(memUsed, sys.MemTotal)))
	row("disk", fmt.Sprintf("%s / %s  (%s%%)", fmtBytes(sys.DiskUsed), fmtBytes(sys.DiskTotal), fmtPct(sys.DiskUsed, sys.DiskTotal)))
	nodeStatus := clr(ansiRed, "stopped")
	if docker.Running {
		nodeStatus = clr(ansiGreen, "running")
	}
	nodeInfo := fmt.Sprintf("%s  %.1f%% CPU", nodeStatus, docker.CPUPercent)
	if !docker.StartedAt.IsZero() {
		nodeInfo += "  up " + fmtUptime(docker.StartedAt)
	}
	row("node", nodeInfo)

	section("CHAIN")
	row("height",   fmtInt(chain.BlockHeight))
	row("interval", fmtDur(chain.BlockInterval))
	if !chain.LatestBlockTime.IsZero() {
		row("lag", fmtDur(time.Since(chain.LatestBlockTime)))
	}
	row("peers", fmt.Sprintf("%d cosmos  %d evm", chain.PeerCount, ev.PeerCount))

	section("VALIDATORS")
	vals := make([]fetch.ValidatorInfo, len(chain.Validators))
	copy(vals, chain.Validators)
	sort.Slice(vals, func(i, j int) bool { return vals[i].VotingPowerPercent > vals[j].VotingPowerPercent })
	for _, v := range vals {
		flags := ""
		if v.Jailed {
			flags += "  " + clr(ansiRed, "JAILED")
		}
		if v.Tombstoned {
			flags += "  " + clr(ansiRed, "TOMBSTONED")
		}
		fmt.Fprintf(w, "  %-16s  %5.1f%%  %d missed  %s%s\n",
			truncate(v.Moniker, 16), v.VotingPowerPercent,
			v.MissedBlocks, strings.ToLower(v.Status), flags)
	}
	fmt.Fprintln(w)
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

// fmtFraction formats a slash fraction string (decimal like "0.000100000000000000")
// as a human-readable percentage.
func fmtFraction(s string) string {
	v := 0.0
	fmt.Sscanf(s, "%f", &v)
	return fmt.Sprintf("%.4f%%", v*100)
}
