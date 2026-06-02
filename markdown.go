package main

import (
	"fmt"
	"strings"
)

func buildMarkdown(d WebData) string {
	var b strings.Builder
	w := &b

	section    := func(name string)         { fmt.Fprintf(w, "\n# %s\n\n", name) }
	subsection := func(name string)         { fmt.Fprintf(w, "\n## %s\n\n", name) }
	row        := func(label, value string) { fmt.Fprintf(w, "- **%s**: %s\n", label, value) }

	syncStr := "synced"
	if !d.Synced {
		syncStr = "CATCHING UP"
	}

	// ── 1. INFRASTRUCTURE ────────────────────────────────────────────────────
	section("1. INFRASTRUCTURE")

	subsection("OS")
	row("load", fmt.Sprintf("%.2f / %.2f / %.2f  (1m 5m 15m)", d.Load1, d.Load5, d.Load15))
	row("ram",  fmt.Sprintf("%s / %s  (%d%%)", d.MemUsed, d.MemTotal, d.MemPct))
	row("disk", fmt.Sprintf("%s / %s  (%d%%)", d.DiskUsed, d.DiskTotal, d.DiskPct))

	subsection("Container")
	nodeStatus := "stopped"
	if d.NodeRunning {
		nodeStatus = "running"
	}
	row("status",   nodeStatus)
	row("cpu",      d.NodeCPU)
	row("ram",      fmt.Sprintf("%s / %s", d.NodeMemUsed, d.NodeMemTotal))
	row("restarts", fmt.Sprintf("%d", d.Restarts))
	if d.NodeUptime != "" {
		row("uptime", d.NodeUptime)
	}

	// ── 2. NODE ──────────────────────────────────────────────────────────────
	section("2. NODE")

	subsection("Identity")
	row("moniker", d.Moniker)
	if d.AppVersion != "" {
		row("version", d.AppVersion)
	}
	if d.ListenAddr != "" {
		row("p2p", d.ListenAddr)
	}
	if d.Network != "" {
		row("chain", d.Network)
	}

	subsection("Block")
	row("height", d.BlockHeight)
	if d.BlockInterval != "" {
		row("interval", d.BlockInterval)
	}
	if d.TimeSinceBlock != "" {
		row("last block", d.TimeSinceBlock+" ago")
	}
	row("status", syncStr)

	subsection("P2P")
	if len(d.PeerMonikers) > 0 {
		row("peers", strings.Join(d.PeerMonikers, "  "))
	} else {
		row("peers", fmt.Sprintf("%d", d.PeerCount))
	}
	row("mempool", fmt.Sprintf("%d pending", d.MempoolTxs))


	// ── 3. VALIDATORS ────────────────────────────────────────────────────────
	section("3. VALIDATORS")

	fmt.Fprintf(w, "| moniker | vp%% | commission | missed | outstanding | earned | tombstoned | status | jailed |\n")
	fmt.Fprintf(w, "|---------|------|------------|--------|-------------|--------|------------|--------|--------|\n")
	for _, v := range d.Validators {
		tombStr := "no"
		if v.Tombstoned {
			tombStr = "YES"
		}
		jailed := ""
		if v.Jailed {
			jailed = "JAILED"
		}
		outstanding := v.Outstanding
		if outstanding == "" {
			outstanding = "–"
		}
		earned := v.Earned
		if earned == "" {
			earned = "–"
		}
		fmt.Fprintf(w, "| %s | %.1f%% | %.1f%% | %d | %s | %s | %s | %s | %s |\n",
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

	subsection("Pool")
	row("total supply", d.TotalSupply)
	row("bonded",       fmt.Sprintf("%s  (%.2f%%, goal %.0f%%)", d.BondedAmt, d.BondedPct, d.GoalBonded))
	row("not bonded",   d.NotBonded)
	if d.UnbondingTime != "" {
		row("unbonding time", d.UnbondingTime)
	}
	if d.MaxValidators > 0 {
		row("max validators", fmt.Sprintf("%d", d.MaxValidators))
	}

	subsection("Slashing")
	if d.SlashWindow != "" && d.SlashWindow != "0" {
		row("window", d.SlashWindow+" blocks")
	}
	if d.MinSigned > 0 {
		row("min signed", fmt.Sprintf("%.1f%%", d.MinSigned))
	}
	if d.SlashDowntime != "" {
		dtStr := d.SlashDowntime
		if d.SlashDTInactive {
			dtStr += "  ⚠ inactive"
		} else {
			dtStr += "  active"
		}
		row("slash / downtime", dtStr)
	}
	if d.SlashDS != "" {
		dsStr := d.SlashDS
		if d.SlashDSInactive {
			dsStr += "  ⚠ inactive"
		} else {
			dsStr += "  active"
		}
		row("slash / double-sign", dsStr)
	}

	// ── 4. ECONOMICS ─────────────────────────────────────────────────────────
	section("4. ECONOMICS")

	subsection("x/mint")
	inflationStr := fmt.Sprintf("%.2f%%", d.Inflation)
	if d.Inflation == 0 {
		inflationStr += "  ⚠ inactive  (module present — can be activated via governance)"
	}
	row("inflation", inflationStr)
	if d.BlocksPerYear != "" {
		row("blocks / year", d.BlocksPerYear)
	}

	subsection("x/distribution")
	row("community tax",  d.CommunityTax)
	row("community pool", d.CommunityPool)

	subsection("x/pmtrewards  (custom)")
	row("status", mdPMTStatus(d))
	if d.PMTRate != "" {
		row("rate", d.PMTRate)
		if d.PMTAnnual != "" {
			row("annual emissions", d.PMTAnnual)
		}
	}
	if d.PMTPoolEmpty {
		row("pool balance", "0 PMT  — pool empty")
	} else if d.PMTBalance != "" {
		mdRunway := ""
		if d.PMTRunway != "" {
			mdRunway = "  (" + d.PMTRunway + ")"
		}
		row("pool balance", d.PMTBalance+mdRunway)
	}
	if d.PMTPoolAddress != "" {
		row("pool address", d.PMTPoolAddress)
	}

	if d.PMTInsights {
		subsection("Insights")
		if d.PMTPoolEmpty {
			row("pool runway",     "EMPTY  (no PMT rewards distributing)")
			row("daily emissions", d.PMTDailyEmit)
		} else {
			row("pool runway",     d.PMTRunwayDays)
			row("daily emissions", d.PMTDailyEmit)
			if d.PMTPerValDay != "" {
				row("per validator/day", d.PMTPerValDay)
			}
		}
		if d.PMTRevFlow != "" {
			fmt.Fprintf(w, "\n```\n")
			fmt.Fprintf(w, "%s\n", d.PMTRevFlow)
			fmt.Fprintf(w, "    ├─ %.0f%%  commission → validator\n", d.PMTCommPct)
			fmt.Fprintf(w, "    └─ %.0f%%  → delegators (pro-rata)\n", d.PMTDelegPct)
			fmt.Fprintf(w, "```\n\n")
		}
	}

	subsection("Validator Earnings  (unclaimed)")
	if d.TotalOutstanding != "" {
		row("total outstanding", d.TotalOutstanding)
	}
	row("commission rate", fmt.Sprintf("%.0f%%  of staking rewards", d.CommissionRate))

	// ── 5. GOVERNANCE ────────────────────────────────────────────────────────
	section("5. GOVERNANCE")

	subsection("Voting")
	row("voting period", d.VotingPeriod)
	row("quorum",        fmt.Sprintf("%.1f%%", d.Quorum))
	row("threshold",     fmt.Sprintf("%.1f%%", d.Threshold))
	if d.VetoThreshold > 0 {
		row("veto threshold", fmt.Sprintf("%.1f%%", d.VetoThreshold))
	}

	if len(d.Proposals) > 0 {
		subsection("Active Proposals (voting period)")
		for _, pr := range d.Proposals {
			fmt.Fprintf(w, "- **#%d** %s  _(ends %s)_\n", pr.ID, truncate(pr.Title, 40), pr.End)
			if pr.HasTally {
				fmt.Fprintf(w, "  - yes %s  no %s  abstain %s  veto %s\n",
					pr.TallyYes, pr.TallyNo, pr.TallyAbstain, pr.TallyVeto)
			}
		}
		fmt.Fprintln(w)
	}

	if len(d.DepositProposals) > 0 {
		subsection("Deposit-Period Proposals")
		for _, pr := range d.DepositProposals {
			fmt.Fprintf(w, "- **#%d** %s  _(deposit ends %s)_\n", pr.ID, truncate(pr.Title, 40), pr.End)
		}
		fmt.Fprintln(w)
	}

	if len(d.Proposals)+len(d.DepositProposals) == 0 {
		fmt.Fprintf(w, "none active\n\n")
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
		fmt.Fprintf(w, "none registered\n\n")
	}
	for _, tp := range d.TokenPairs {
		enabled := "yes"
		if !tp.Enabled {
			enabled = "no"
		}
		fmt.Fprintf(w, "- `%s`  `%s`  enabled: %s\n", tp.Denom, tp.ERC20, enabled)
	}

	// ── 6. EVM ────────────────────────────────────────────────────────────────
	section("6. EVM")

	evSyncStr := "synced"
	if !d.EVMSynced {
		evSyncStr = "syncing"
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
		row("rpc", "ok  (eth_blockNumber responded)")
	} else {
		row("rpc", "error")
	}
	listeningStr := "yes"
	if !d.EVMListening {
		listeningStr = "no"
	}
	row("net listening", listeningStr)
	row("peers", fmt.Sprintf("%d", d.EVMPeerCount))
	if d.EVMBlockAge != "" {
		ageStr := d.EVMBlockAge
		if d.EVMBlockAgeErr {
			ageStr += "  ⚠ stalled"
		} else if d.EVMBlockAgeWarn {
			ageStr += "  ⚠ slow"
		}
		row("last block age", ageStr)
	}
	row("sync",   evSyncStr)
	row("txpool", fmt.Sprintf("pending %d   queued %d", d.PendingTx, d.QueuedTx))

	subsection("Block")
	row("block", d.EVMBlock)

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
		noBaseFeeStr += "disabled)"
	} else {
		noBaseFeeStr += "active)"
	}
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
			fmt.Fprintf(w, "- %s\n", pc)
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

	fmt.Fprintln(w)
	return b.String()
}

// mdPMTStatus returns a plain-text PMT status string for Markdown output.
func mdPMTStatus(d WebData) string {
	if !d.PMTEnabled {
		return "disabled"
	}
	if d.PMTPoolEmpty {
		return "ENABLED — pool EMPTY  (validators receive no PMT rewards)"
	}
	suffix := ""
	if d.PMTRunway != "" {
		suffix = "  (" + d.PMTRunway + ")"
	}
	return "distributing  " + d.PMTRate + "   pool " + d.PMTBalance + suffix
}
