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
	row("ram", fmt.Sprintf("%s / %s  (%d%%)", d.MemUsed, d.MemTotal, d.MemPct))
	row("disk", fmt.Sprintf("%s / %s  (%d%%)", d.DiskUsed, d.DiskTotal, d.DiskPct))

	subsection("Container")
	nodeStatus := "stopped"
	if d.NodeRunning {
		nodeStatus = "running"
	}
	row("status", nodeStatus)
	row("cpu", d.NodeCPU)
	row("ram", fmt.Sprintf("%s / %s", d.NodeMemUsed, d.NodeMemTotal))
	row("restarts", fmt.Sprintf("%d", d.Restarts))
	if d.NodeUptime != "" {
		row("uptime", d.NodeUptime)
	}

	// ── 2. NODE ──────────────────────────────────────────────────────────────
	section("2. NODE")

	fmt.Fprintf(w, "_This machine's CometBFT process — identity, listen addresses, and consensus view._\n\n")

	subsection("Node")
	row("moniker", d.Moniker)
	if d.NodeID != "" {
		row("node ID", d.NodeID+"  _(CometBFT P2P peer ID)_")
	}
	if d.AppVersion != "" {
		row("version", d.AppVersion)
	}
	if d.Network != "" {
		row("chain ID", d.Network)
	}
	if d.ListenAddr != "" {
		row("p2p listen", d.ListenAddr+"  _(advertised dial address from `/status`)_")
	}
	if d.RpcListenAddr != "" {
		row("rpc listen", d.RpcListenAddr)
	}

	subsection("Consensus")
	row("sync", syncStr)
	row("height", d.BlockHeight)
	if d.BlockInterval != "" {
		row("interval", d.BlockInterval)
	}
	if d.TimeSinceBlock != "" {
		row("last block", d.TimeSinceBlock)
	}
	if d.LocalConsensusAddr != "" {
		row("consensus address", strings.ToUpper(d.LocalConsensusAddr)+"  _(hex; signs blocks — `/status` validator_info)_")
	}
	if d.LocalVotingPower != "" {
		row("voting power", d.LocalVotingPower+"  _(consensus units — `/status` validator_info)_")
	}
	row("mempool", fmt.Sprintf("%d pending", d.MempoolTxs))

	// ── 3. VALIDATOR SET ─────────────────────────────────────────────────────
	section("3. VALIDATOR SET")

	fmt.Fprintf(w, "_All validators on the chain — P2P dial from `/status` + `/net_info`, stake, signing health, and params._\n\n")
	fmt.Fprintf(w, "_P2P dial is `node_id@listen_addr` when this node sees the peer in `/net_info`, or from `/status` for this node. Other validators show — until peered._\n\n")

	fmt.Fprintf(w, "| moniker | operator | p2p dial | node ID | consensus | vp%% | commission | missed | status | jailed | local |\n")
	fmt.Fprintf(w, "|---------|----------|----------|---------|-----------|-----|------------|--------|--------|--------|-------|\n")
	for _, v := range d.Validators {
		missed := fmt.Sprintf("%d", v.Missed)
		if v.MissedHigh {
			missed += " ⚠"
		}
		jailed := ""
		if v.Jailed {
			jailed = "yes"
		}
		local := ""
		if v.IsLocal {
			local = "**this node**"
		}
		tomb := ""
		if v.Tombstoned {
			tomb = "  tombstoned"
		}
		p2p := v.P2PDial
		if p2p == "" || p2p == "—" {
			p2p = "—"
		}
		fmt.Fprintf(w, "| %s | %s | %s | %s | %s | %.1f%% | %.1f%% | %s | %s%s | %s | %s |\n",
			truncate(v.Moniker, 12),
			v.Operator,
			p2p,
			v.NodeID,
			v.ConsensusAddr,
			v.VPFloat,
			v.CommissionFloat,
			missed,
			v.Status,
			tomb,
			jailed,
			local,
		)
	}

	subsection("P2P on this node")
	if len(d.PeerMonikers) > 0 {
		row("connected", fmt.Sprintf("%d — %s", d.PeerCount, strings.Join(d.PeerMonikers, ", "))+"  _(from `/net_info`)_")
	} else {
		row("connected", fmt.Sprintf("%d peers  _(from `/net_info`)_", d.PeerCount))
	}

	subsection("Summary")
	row("bonded", fmt.Sprintf("%d", d.BondedCount))
	row("jailed", fmt.Sprintf("%d", d.JailedCount))
	row("tombstoned", fmt.Sprintf("%d", d.TombstonedCount))
	row("below min signed", fmt.Sprintf("%d", d.BelowThreshold))
	if d.NextProposer != "" {
		row("next proposer", d.NextProposer)
	}

	subsection("Staking Pool")
	if d.BondDenom != "" {
		row("bond denom", d.BondDenom)
	}
	row("total supply", d.TotalSupply)
	row("bonded", fmt.Sprintf("%s  (%.2f%%, goal %.0f%%)", d.BondedAmt, d.BondedPct, d.GoalBonded))
	row("not bonded", d.NotBonded)
	if d.UnbondingTime != "" {
		row("unbonding time", d.UnbondingTime+"  _(time locked after unstaking)_")
	}
	if d.MaxValidators > 0 {
		row("max validators", fmt.Sprintf("%d", d.MaxValidators))
	}

	subsection("Slashing Params")
	if d.SlashWindow != "" && d.SlashWindow != "0" {
		row("signed blocks window", d.SlashWindow+" blocks")
	}
	if d.MinSigned > 0 {
		row("min signed per window", fmt.Sprintf("%.1f%%  _(miss more → downtime slash risk)_", d.MinSigned))
	}
	if d.SlashDowntime != "" {
		dtStr := d.SlashDowntime
		if d.SlashDTInactive {
			dtStr += "  ⚠ inactive"
		}
		row("slash / downtime", dtStr)
	}
	if d.SlashDS != "" {
		dsStr := d.SlashDS
		if d.SlashDSInactive {
			dsStr += "  ⚠ inactive"
		}
		row("slash / double-sign", dsStr)
	}

	// ── 4. THIS VALIDATOR ──────────────────────────────────────────────────────
	section("4. THIS VALIDATOR")

	fmt.Fprintf(w, "_Staking and rewards for this machine's validator — matched via `/status` consensus address. Node identity is in §2._\n\n")

	lv := d.Local
	if !lv.IsValidator {
		row("role", lv.SigningStatus)
		row("moniker", d.Moniker)
	} else {
		subsection("Operator")
		if lv.OperatorAddr != "" {
			row("operator address", lv.OperatorAddr+"  _(staking / rewards — `evmd query/distribution` use this)_")
		}
		row("moniker", lv.Moniker)

		subsection("Staking")
		row("status", lv.Status)
		if lv.Jailed {
			row("jailed", "yes")
		}
		if lv.Tombstoned {
			row("tombstoned", "YES")
		}
		row("voting power", fmt.Sprintf("%s  (%.1f%% of bonded stake)", lv.VotingPower, lv.VPPercent))
		row("commission", fmt.Sprintf("%.1f%%  _(your cut of delegator rewards)_", lv.Commission))

		subsection("Block Signing")
		row("signing health", lv.SigningStatus)
		if d.SlashWindow != "" && d.SlashWindow != "0" {
			row("missed / window", fmt.Sprintf("%d / %s blocks  (max allowed: %d)", lv.Missed, d.SlashWindow, lv.MaxMissed))
		}
		if lv.IsNextProposer {
			row("proposer", "**next block proposer**")
		} else if d.NextProposer != "" {
			row("proposer", "not next  _(next: "+d.NextProposer+")_")
		}
		if lv.ProposerPriority != 0 {
			row("proposer priority", fmtInt(lv.ProposerPriority))
		}

		subsection("Unclaimed Rewards")
		if lv.Outstanding != "" {
			row("outstanding rewards", lv.Outstanding+"  _(total rewards not yet withdrawn)_")
		} else {
			row("outstanding rewards", "–")
		}
		if lv.CommissionEarned != "" {
			row("commission earned", lv.CommissionEarned+"  _(your validator commission, unclaimed)_")
		} else {
			row("commission earned", "–")
		}
	}

	// ── 5. ECONOMICS ─────────────────────────────────────────────────────────
	section("5. ECONOMICS")

	fmt.Fprintf(w, "_How money moves on this chain — staking, inflation, fees, and rewards._\n\n")

	subsection("Overview")
	fmt.Fprintf(w, "```\n")
	fmt.Fprintf(w, "                         WHERE VALUE COMES FROM\n")
	fmt.Fprintf(w, "  ┌─────────────────────┐              ┌──────────────────────┐\n")
	fmt.Fprintf(w, "  │  Tx fees (EVM/Cosmos)│              │  Inflation (x/mint)   │\n")
	fmt.Fprintf(w, "  │  paid by users       │              │  new tokens / block   │\n")
	fmt.Fprintf(w, "  └──────────┬──────────┘              └──────────┬───────────┘\n")
	fmt.Fprintf(w, "             │                                    │\n")
	fmt.Fprintf(w, "             └────────────────┬───────────────────┘\n")
	fmt.Fprintf(w, "                              ▼\n")
	fmt.Fprintf(w, "                    ┌─────────────────────┐\n")
	fmt.Fprintf(w, "                    │  Block reward pool   │\n")
	fmt.Fprintf(w, "                    │  (x/distribution)    │\n")
	fmt.Fprintf(w, "                    └──────────┬──────────┘\n")
	fmt.Fprintf(w, "           ┌───────────────────┼────────────────────┐\n")
	fmt.Fprintf(w, "           ▼                   ▼                    ▼\n")
	fmt.Fprintf(w, "   ┌───────────────┐   ┌──────────────┐   ┌─────────────────┐\n")
	fmt.Fprintf(w, "   │  Validators    │   │ Community    │   │ PMT pool        │\n")
	fmt.Fprintf(w, "   │  (by stake %)  │   │ pool (tax)   │   │ (x/pmtrewards)  │\n")
	fmt.Fprintf(w, "   └───────┬───────┘   └──────────────┘   └─────────────────┘\n")
	fmt.Fprintf(w, "           │\n")
	fmt.Fprintf(w, "     commission %% → operator\n")
	fmt.Fprintf(w, "     remainder    → delegators\n")
	fmt.Fprintf(w, "```\n\n")

	row("bond denom", d.BondDenom)
	row("total supply", d.TotalSupply)
	row("bonded stake", fmt.Sprintf("%s  (%.1f%% of supply)", d.BondedAmt, d.BondedPct))
	row("not bonded", d.NotBonded)

	subsection("Staking & Inflation  (x/mint + x/staking)")
	inflationStr := fmt.Sprintf("%.2f%%", d.Inflation)
	if d.Inflation == 0 {
		inflationStr += "  ⚠ inactive"
	}
	row("inflation rate", inflationStr+"  _(extra tokens minted when active — rewards stakers)_")
	if d.AnnualProvisions != "" {
		row("annual provisions", d.AnnualProvisions+"  _(absolute new tokens/year if inflation active)_")
	}
	row("goal bonded", fmt.Sprintf("%.0f%%  _(target stake ratio — inflation adjusts toward this)_", d.GoalBonded))
	if d.BlocksPerYear != "" {
		row("blocks / year", d.BlocksPerYear)
	}
	if d.UnbondingTime != "" {
		row("unbonding time", d.UnbondingTime+"  _(tokens locked after you unstake)_")
	}

	subsection("Distribution  (x/distribution)")
	row("community tax", d.CommunityTax+"  _(%% of block rewards → community pool, not validators)_")
	row("community pool", d.CommunityPool+"  _(governance-controlled treasury)_")
	if d.TotalOutstanding != "" {
		row("unclaimed staking rewards", d.TotalOutstanding+"  _(validators haven't withdrawn yet)_")
	}

	subsection("Fee Structure & Flow")
	fmt.Fprintf(w, "_Every EVM transaction pays gas. Fees are collected on-chain and routed to validators (and optionally the community pool)._\n\n")
	fmt.Fprintf(w, "```\n")
	fmt.Fprintf(w, "  User submits EVM tx\n")
	fmt.Fprintf(w, "       │\n")
	fmt.Fprintf(w, "       ▼ pays gas = gas_used × effective_gas_price\n")
	fmt.Fprintf(w, "  ┌─────────────────────────────────────────┐\n")
	fmt.Fprintf(w, "  │  EIP-1559 fee market (x/feemarket)       │\n")
	fmt.Fprintf(w, "  │  base_fee (burned/adjusted) + priority   │\n")
	fmt.Fprintf(w, "  └──────────────────┬──────────────────────┘\n")
	fmt.Fprintf(w, "                     ▼\n")
	if d.CommunityTaxZero {
		fmt.Fprintf(w, "  100%% of collected fees → validators (pro-rata by stake)\n")
	} else {
		fmt.Fprintf(w, "  ┌─────────────────────────────────────────┐\n")
		fmt.Fprintf(w, "  │  x/distribution splits block income:     │\n")
		fmt.Fprintf(w, "  │  • %.1f%% → community pool               │\n", d.CommunityTaxPct)
		fmt.Fprintf(w, "  │  • %.1f%% → validators + delegators        │\n", 100-d.CommunityTaxPct)
		fmt.Fprintf(w, "  └─────────────────────────────────────────┘\n")
	}
	if d.PMTEnabled && d.PMTRate != "" {
		fmt.Fprintf(w, "  + PMT pool adds %s from x/pmtrewards\n", d.PMTRate)
	}
	fmt.Fprintf(w, "```\n\n")

	row("model", "EIP-1559  _(base fee rises when blocks are full, falls when empty)_")
	if d.BaseFee != "" {
		row("current base fee", d.BaseFee+" wei")
	}
	if d.GasPrice != "" {
		row("current gas price", d.GasPrice+"  _(from JSON-RPC eth_gasPrice)_")
	}
	if d.MinGasPrice != "" {
		row("min gas price", d.MinGasPrice+"  _(chain-enforced floor)_")
	}
	if d.BlockGas != "" {
		row("gas used (last block)", d.BlockGas)
	}
	if d.Elasticity > 0 {
		row("block gas target", fmt.Sprintf("max_block_gas ÷ %d", d.Elasticity))
	}
	if d.AdjCap != "" {
		row("base fee max change", d.AdjCap)
	}
	if d.BaseFeeChangeDenominator > 0 {
		row("change denominator", fmt.Sprintf("%d", d.BaseFeeChangeDenominator))
	}
	noBaseFeeStr := boolStr(d.NoBaseFee)
	if d.NoBaseFee {
		noBaseFeeStr += "  _(EIP-1559 enforcement disabled)_"
	} else {
		noBaseFeeStr += "  _(EIP-1559 active)_"
	}
	row("no_base_fee flag", noBaseFeeStr)

	subsection("PMT Rewards  (x/pmtrewards — custom)")
	row("status", mdPMTStatus(d))
	if d.PMTRate != "" {
		row("reward rate", d.PMTRate+"  _(extra tokens per block from PMT pool)_")
	}
	if d.PMTAnnual != "" {
		row("annual emissions", d.PMTAnnual)
	}
	if d.PMTDailyEmit != "" {
		row("daily emissions", d.PMTDailyEmit)
	}
	if d.PMTPoolEmpty {
		row("pool balance", "0  — pool empty, no PMT rewards distributing")
	} else if d.PMTBalance != "" {
		bal := d.PMTBalance
		if d.PMTRunway != "" {
			bal += "  (" + d.PMTRunway + ")"
		}
		row("pool balance", bal)
	}
	if d.PMTPoolAddress != "" {
		row("pool address", d.PMTPoolAddress)
	}

	// ── 6. GOVERNANCE ────────────────────────────────────────────────────────
	section("6. GOVERNANCE")

	if len(d.Proposals) > 0 {
		subsection(fmt.Sprintf("Active Proposals  (%d)", len(d.Proposals)))
		for _, pr := range d.Proposals {
			fmt.Fprintf(w, "- **#%d** %s  _(voting ends %s)_\n", pr.ID, truncate(pr.Title, 40), pr.End)
			if pr.HasTally {
				fmt.Fprintf(w, "  - yes %s  no %s  abstain %s  veto %s\n",
					pr.TallyYes, pr.TallyNo, pr.TallyAbstain, pr.TallyVeto)
			}
		}
		fmt.Fprintln(w)
	}

	if len(d.DepositProposals) > 0 {
		subsection(fmt.Sprintf("Deposit-Period Proposals  (%d)", len(d.DepositProposals)))
		for _, pr := range d.DepositProposals {
			fmt.Fprintf(w, "- **#%d** %s  _(deposit ends %s)_\n", pr.ID, truncate(pr.Title, 40), pr.End)
		}
		fmt.Fprintln(w)
	}

	if len(d.Proposals)+len(d.DepositProposals) == 0 {
		fmt.Fprintf(w, "_No active proposals._\n\n")
	}

	subsection("Voting Params")
	row("voting period", d.VotingPeriod)
	row("quorum", fmt.Sprintf("%.1f%%", d.Quorum))
	row("threshold", fmt.Sprintf("%.1f%%", d.Threshold))
	if d.VetoThreshold > 0 {
		row("veto threshold", fmt.Sprintf("%.1f%%", d.VetoThreshold))
	}

	subsection("Upgrade")
	if d.UpgradeName == "" {
		row("pending", "none")
	} else {
		row("name", d.UpgradeName)
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

	// ── 7. EVM JSON-RPC ────────────────────────────────────────────────────────
	section("7. EVM JSON-RPC")

	fmt.Fprintf(w, "_Health and metrics from the EVM JSON-RPC endpoint (`eth_*`, `net_*`, `txpool_*`)._\n\n")

	evSyncStr := "synced"
	if !d.EVMSynced {
		evSyncStr = "syncing"
	}

	subsection("Endpoint Health")
	okStr := "error"
	if d.EVMRPCOk {
		okStr = fmt.Sprintf("ok  (%d/%d methods responded)", d.RPCProbeOK, d.RPCProbeTotal)
	}
	row("overall", okStr)
	listeningStr := "no"
	if d.EVMListening {
		listeningStr = "yes"
	}
	row("net_listening", listeningStr)
	row("peers", fmt.Sprintf("%d  _(net_peerCount)_", d.EVMPeerCount))
	row("sync", evSyncStr+"  _(eth_syncing)_")
	if d.EVMClient != "" {
		row("client", d.EVMClient+"  _(web3_clientVersion)_")
	}

	subsection("Live Metrics")
	row("chain ID", fmt.Sprintf("%d  _(eth_chainId)_", d.EVMChainID))
	if d.EVMDenom != "" {
		row("denom", d.EVMDenom)
	}
	row("block height", d.EVMBlock+"  _(eth_blockNumber)_")
	if d.EVMBlockAge != "" {
		ageStr := d.EVMBlockAge
		if d.EVMBlockAgeErr {
			ageStr += "  ⚠ stalled"
		} else if d.EVMBlockAgeWarn {
			ageStr += "  ⚠ slow"
		}
		row("last block age", ageStr+"  _(eth_getBlockByNumber)_")
	}
	row("txpool", fmt.Sprintf("pending %d   queued %d  _(txpool_status)_", d.PendingTx, d.QueuedTx))

	subsection("Method Probes")
	fmt.Fprintf(w, "| method | status | latency | error |\n")
	fmt.Fprintf(w, "|--------|--------|---------|-------|\n")
	for _, p := range d.RPCProbes {
		status := "ok"
		errStr := ""
		if !p.OK {
			status = "FAIL"
			errStr = truncate(p.Error, 30)
		}
		fmt.Fprintf(w, "| `%s` | %s | %s | %s |\n", p.Method, status, p.Latency, errStr)
	}
	fmt.Fprintln(w)

	subsection("Raw JSON-RPC Samples")
	for _, p := range d.RPCProbes {
		status := "ok"
		if !p.OK {
			status = "FAIL"
		}
		fmt.Fprintf(w, "**%s** (%s, %s)\n\n", p.Method, status, p.Latency)
		fmt.Fprintf(w, "```json\n%s\n→\n%s\n```\n\n", p.Request, p.Response)
	}

	subsection("EVM Config")
	if len(d.Precompiles) > 0 {
		row("precompiles", fmt.Sprintf("%d active", len(d.Precompiles)))
		for _, pc := range d.Precompiles {
			fmt.Fprintf(w, "  - `%s`\n", pc)
		}
	}
	if d.HistoryWindow != "" {
		row("history serve window", d.HistoryWindow+" blocks")
	}
	row("ERC20 module", boolStr(d.ERC20Enabled))
	if d.HardforkLondon != "" {
		row("London", "height "+d.HardforkLondon)
	}
	if d.HardforkShanghai != "" {
		row("Shanghai", "time "+d.HardforkShanghai)
	}
	if d.HardforkCancun != "" {
		row("Cancun", "time "+d.HardforkCancun)
	}

	fmt.Fprintln(w)
	return b.String()
}

func mdPMTStatus(d WebData) string {
	if !d.PMTEnabled {
		return "disabled"
	}
	if d.PMTPoolEmpty {
		return "enabled — pool empty  (no PMT rewards distributing)"
	}
	suffix := ""
	if d.PMTRunway != "" {
		suffix = "  (" + d.PMTRunway + ")"
	}
	return "distributing  " + d.PMTRate + "   pool " + d.PMTBalance + suffix
}
