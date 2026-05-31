package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/arkantos1482/cosmos-monitor/fetch"
)

// out is the print target; replaced with an \r\n-translating writer in raw mode.
var out io.Writer = os.Stdout

func printAll(chain fetch.ChainSnapshot, ev fetch.EVMSnapshot, sys fetch.SystemSnapshot, docker fetch.DockerSnapshot) {
	p := chain.Params

	// ── header ───────────────────────────────────────────────────────────────
	syncStr := "synced"
	if chain.CatchingUp {
		syncStr = "CATCHING UP"
	}
	fmt.Fprintf(out, "pmtop  %s  %s  height %s  %s UTC\n\n",
		chain.Moniker, syncStr, fmtInt(chain.BlockHeight), time.Now().UTC().Format("15:04:05"))

	// ── node ─────────────────────────────────────────────────────────────────
	section("NODE")
	row("node ID",  chain.NodeID)
	row("version",  chain.AppVersion)
	row("height",   fmtInt(chain.BlockHeight))
	row("block",    fmtDur(chain.BlockInterval))
	row("sync",     syncStr)
	row("peers",    fmt.Sprintf("%d (comet)  /  %d (evm)", chain.PeerCount, ev.PeerCount))

	// ── system ───────────────────────────────────────────────────────────────
	section("SYSTEM")
	memUsed := uint64(0)
	if sys.MemTotal > sys.MemAvail {
		memUsed = sys.MemTotal - sys.MemAvail
	}
	row("load", fmt.Sprintf("%.2f  %.2f  %.2f  (1m 5m 15m)", sys.LoadAvg1, sys.LoadAvg5, sys.LoadAvg15))
	row("ram",  fmt.Sprintf("%s / %s  (%s%%)", fmtBytes(memUsed), fmtBytes(sys.MemTotal), fmtPct(memUsed, sys.MemTotal)))
	row("disk", fmt.Sprintf("%s / %s  (%s%%)", fmtBytes(sys.DiskUsed), fmtBytes(sys.DiskTotal), fmtPct(sys.DiskUsed, sys.DiskTotal)))

	// ── container ─────────────────────────────────────────────────────────────
	section("CONTAINER")
	status := "stopped"
	if docker.Running {
		status = "running"
	}
	row("status",   status)
	row("cpu",      fmt.Sprintf("%.1f%%", docker.CPUPercent))
	row("ram",      fmt.Sprintf("%s / %s", fmtBytes(docker.MemUsage), fmtBytes(docker.MemLimit)))
	row("restarts", fmt.Sprintf("%d", docker.RestartCount))
	if !docker.StartedAt.IsZero() {
		row("uptime", fmtUptime(docker.StartedAt))
	}

	// ── staking ───────────────────────────────────────────────────────────────
	section("STAKING")
	denom := p.BondDenom
	if denom == "" {
		denom = chain.TotalSupplyDenom
	}
	bondedDisp := fetch.FormatCoin(chain.BondedTokens, denom)
	totalDisp  := fetch.FormatCoin(chain.TotalSupply, chain.TotalSupplyDenom)
	bondedF, _ := fetch.NormalizeCoin(chain.BondedTokens, denom)
	totalF, _  := fetch.NormalizeCoin(chain.TotalSupply, chain.TotalSupplyDenom)
	bondPct := 0.0
	if totalF > 0 {
		bondPct = bondedF / totalF * 100
	}
	row("supply",     totalDisp)
	row("bonded",     fmt.Sprintf("%s  %.1f%%  (goal %.1f%%)", bondedDisp, bondPct, p.GoalBonded*100))
	row("not bonded", fetch.FormatCoin(chain.NotBondedTokens, denom))
	row("inflation",  fmt.Sprintf("%.4f%%", chain.Inflation*100))
	row("unbonding",  fmtDurFull(p.UnbondingTime))
	row("max vals",   fmt.Sprintf("%d", p.MaxValidators))
	row("denom",      denom)
	if p.BlocksPerYear > 0 {
		row("blocks/yr", fmtInt(p.BlocksPerYear))
	}

	// ── distribution ──────────────────────────────────────────────────────────
	section("DISTRIBUTION")
	row("community pool", chain.CommunityPool)
	row("community tax",  fmt.Sprintf("%.4f%%", p.CommunityTax*100))

	// ── pmt rewards ───────────────────────────────────────────────────────────
	section("PMT REWARDS")
	row("enabled",    boolStr(p.PMTRewardsEnabled))
	row("reward/blk", fetch.FormatCoin(p.RewardPerBlockAmount, p.RewardPerBlockDenom))
	if p.PMTRewardsPoolAddress != "" {
		row("pool", p.PMTRewardsPoolAddress)
	}

	// ── feemarket ─────────────────────────────────────────────────────────────
	section("FEEMARKET")
	row("base fee",    chain.BaseFee+" wei")
	row("block gas",   fmtInt(int64(chain.BlockGas)))
	row("min gas",     fmt.Sprintf("%.9f %s", p.MinGasPrice, denom))
	row("elasticity",  fmt.Sprintf("%d", p.Elasticity))
	if p.BaseFeeChangeDenominator > 0 {
		row("change denom", fmt.Sprintf("%d", p.BaseFeeChangeDenominator))
	}
	row("no_base_fee", boolStr(p.NoBaseFee))

	// ── evm ───────────────────────────────────────────────────────────────────
	section("EVM")
	evSyncStr := "synced"
	if ev.Syncing {
		evSyncStr = "syncing"
	}
	row("chain ID",       fmt.Sprintf("%d", ev.ChainID))
	row("denom",          p.EVMDenom)
	row("block",          fmtInt(int64(ev.BlockNumber)))
	row("sync",           evSyncStr)
	row("gas price",      ev.GasPrice)
	row("txpool",         fmt.Sprintf("%d pending  /  %d queued", ev.PendingTx, ev.QueuedTx))
	row("erc20",          boolStr(p.ERC20Enabled))
	if p.HistoryServeWindow > 0 {
		row("history window", fmtInt(p.HistoryServeWindow))
	}
	if len(p.ActiveStaticPrecompiles) > 0 {
		row("precompiles", fmt.Sprintf("%d active", len(p.ActiveStaticPrecompiles)))
		for _, pc := range p.ActiveStaticPrecompiles {
			fmt.Fprintf(out, "    %s\n", pc)
		}
	}

	// ── token pairs ───────────────────────────────────────────────────────────
	section(fmt.Sprintf("TOKEN PAIRS  (%d)", len(chain.TokenPairs)))
	if len(chain.TokenPairs) == 0 {
		fmt.Fprintln(out, "  none registered")
	}
	for _, tp := range chain.TokenPairs {
		enabled := ""
		if !tp.Enabled {
			enabled = "  [disabled]"
		}
		fmt.Fprintf(out, "  %-30s  %s%s\n", tp.Denom, tp.ERC20Addr, enabled)
	}

	// ── ibc ───────────────────────────────────────────────────────────────────
	section("IBC")
	row("clients", fmt.Sprintf("%d", chain.IBCClientCount))

	// ── precisebank ───────────────────────────────────────────────────────────
	section("PRECISEBANK")
	if chain.PrecisebankRemainder != "" {
		row("remainder", chain.PrecisebankRemainder)
	} else {
		fmt.Fprintln(out, "  n/a")
	}

	// ── slashing ──────────────────────────────────────────────────────────────
	section("SLASHING")
	row("window",     fmt.Sprintf("%s blocks", fmtInt(p.SignedBlocksWindow)))
	row("min signed", fmt.Sprintf("%.1f%%", p.MinSignedPerWindow*100))
	tombCount := 0
	for _, v := range chain.Validators {
		if v.Tombstoned {
			tombCount++
		}
	}
	row("tombstoned", fmt.Sprintf("%d", tombCount))

	// ── governance ────────────────────────────────────────────────────────────
	section("GOVERNANCE")
	row("voting period", fmtDurFull(p.VotingPeriod))
	row("quorum",        fmt.Sprintf("%.1f%%", p.Quorum*100))
	row("threshold",     fmt.Sprintf("%.1f%%", p.Threshold*100))
	if len(chain.Proposals) == 0 {
		row("proposals", "none active")
	} else {
		for _, pr := range chain.Proposals {
			fmt.Fprintf(out, "  #%d  %-40s  %s  ends %s\n",
				pr.ID, pr.Title, pr.Status, pr.VotingEnd.Format("2006-01-02"))
		}
	}

	// ── upgrade ───────────────────────────────────────────────────────────────
	section("UPGRADE")
	if chain.UpgradeName == "" {
		row("pending", "none")
	} else {
		row("name",   chain.UpgradeName)
		row("height", fmtInt(chain.UpgradeHeight))
		if chain.BlockHeight > 0 && chain.UpgradeHeight > chain.BlockHeight {
			row("blocks left", fmtInt(chain.UpgradeHeight-chain.BlockHeight))
		}
	}

	// ── validators ────────────────────────────────────────────────────────────
	section("VALIDATORS")
	vals := make([]fetch.ValidatorInfo, len(chain.Validators))
	copy(vals, chain.Validators)
	sort.Slice(vals, func(i, j int) bool {
		return vals[i].VotingPowerPercent > vals[j].VotingPowerPercent
	})
	fmt.Fprintf(out, "  %-20s  %6s  %10s  %8s  %10s  %14s  %14s  %s\n",
		"moniker", "vp%", "commission", "missed", "tombstoned", "outstanding", "earned", "status")
	for _, v := range vals {
		tombStr := "no"
		if v.Tombstoned {
			tombStr = "YES"
		}
		st := strings.ToLower(v.Status)
		if v.Jailed {
			st += " JAILED"
		}
		outstanding := v.OutstandingRewards
		if outstanding == "" {
			outstanding = "–"
		}
		earned := v.CommissionEarned
		if earned == "" {
			earned = "–"
		}
		fmt.Fprintf(out, "  %-20s  %5.1f%%  %9.1f%%  %8d  %10s  %14s  %14s  %s\n",
			truncate(v.Moniker, 20),
			v.VotingPowerPercent,
			v.Commission*100,
			v.MissedBlocks,
			tombStr,
			outstanding,
			earned,
			st,
		)
	}

	// ── footer ────────────────────────────────────────────────────────────────
	fmt.Fprintln(out, )
	fmt.Fprintln(out, "r  refresh    q  quit")
}

// ── helpers ──────────────────────────────────────────────────────────────────

func section(name string) {
	fmt.Fprintf(out, "\n%s\n", name)
}

func row(label, value string) {
	fmt.Fprintf(out, "  %-18s  %s\n", label, value)
}

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
