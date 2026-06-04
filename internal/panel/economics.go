package panel

import (
	"fmt"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func writeEconomics(w Writer, d model.Report) {
	w.Section("5. ECONOMICS")
	w.Em("How money moves on this chain — tx fees, PMT pool rewards, and (if active) inflation accumulate in `fee_collector`, then `x/distribution` pays validators each block.")

	w.Subsection("Overview")
	w.Hint("Mermaid diagram (below). Coins: tx fees / inflation / PMT → `fee_collector` → `x/distribution`. `x/staking` supplies voting power (no coin flow). Payout split: community tax, then per-validator commission vs delegators. `goal bonded` → `GET /cosmos/mint/v1beta1/params`. Preview in VS Code/Obsidian or use `pmtop --web`.")
	writeEconomicsDiagram(w, d)
	if d.PMTEnabled {
		w.Em("PMT pool funds per-block rewards via mint hook → `fee_collector` (see PMT Rewards table below).")
	}

	w.Subsection("Staking Pool")
	w.Hint("`bond denom`, `unbonding time`, `max validators` → `GET /cosmos/staking/v1beta1/params`; `total supply` → `x/bank` supply; `bonded` / `not bonded` → `GET /cosmos/staking/v1beta1/pool`.")
	if d.BondDenom != "" {
		w.Row("bond denom", d.BondDenom)
	}
	w.Row("total supply", d.TotalSupply)
	w.Row("bonded", fmt.Sprintf("%s  (%.2f%%, goal %.0f%%)", d.BondedAmt, d.BondedPct, d.GoalBonded))
	w.Row("not bonded", d.NotBonded)
	if d.UnbondingTime != "" {
		w.Row("unbonding time", d.UnbondingTime+"  _(time locked after unstaking)_")
	}
	if d.MaxValidators > 0 {
		w.Row("max validators", fmt.Sprintf("%d", d.MaxValidators))
	}

	w.Subsection("Slashing Params")
	w.Hint("`signed blocks window`, `min signed`, slash fractions → `GET /cosmos/slashing/v1beta1/params`.")
	if d.SlashWindow != "" && d.SlashWindow != "0" {
		w.Row("signed blocks window", d.SlashWindow+" blocks")
	}
	if d.MinSigned > 0 {
		w.Row("min signed per window", fmt.Sprintf("%.1f%%  _(miss more → downtime slash risk)_", d.MinSigned))
	}
	if d.SlashDowntime != "" {
		dtStr := d.SlashDowntime
		if d.SlashDTInactive {
			dtStr += "  ⚠ inactive"
		}
		w.Row("slash / downtime", dtStr)
	}
	if d.SlashDS != "" {
		dsStr := d.SlashDS
		if d.SlashDSInactive {
			dsStr += "  ⚠ inactive"
		}
		w.Row("slash / double-sign", dsStr)
	}

	w.Subsection("Staking & Inflation  (x/mint + x/staking)")
	w.Hint("`inflation rate` → `GET /cosmos/mint/v1beta1/inflation`; `annual provisions` → `…/annual-provisions`; `goal bonded`, `blocks / year` → `…/params`; `unbonding time` → `x/staking` params.")
	inflationStr := fmt.Sprintf("%.2f%%", d.Inflation)
	if d.Inflation == 0 {
		inflationStr += "  ⚠ inactive"
	}
	w.Row("inflation rate", inflationStr+"  _(extra tokens minted when active — rewards stakers)_")
	if d.AnnualProvisions != "" {
		w.Row("annual provisions", d.AnnualProvisions+"  _(absolute new tokens/year if inflation active)_")
	}
	w.Row("goal bonded", fmt.Sprintf("%.0f%%  _(target stake ratio — inflation adjusts toward this)_", d.GoalBonded))
	if d.BlocksPerYear != "" {
		w.Row("blocks / year", d.BlocksPerYear)
	}
	if d.UnbondingTime != "" {
		w.Row("unbonding time", d.UnbondingTime+"  _(tokens locked after you unstake)_")
	}

	w.Subsection("Distribution  (x/distribution)")
	w.Hint("`community tax` → `GET /cosmos/distribution/v1beta1/params`; `community pool` → `…/community_pool`; `unclaimed staking rewards` → sum of per-validator `…/validators/{valoper}/outstanding_rewards`.")
	w.Row("community tax", d.CommunityTax+"  _(%% of block rewards → community pool, not validators)_")
	w.Row("community pool", d.CommunityPool+"  _(governance-controlled treasury)_")
	if d.TotalOutstanding != "" {
		w.Row("unclaimed staking rewards", d.TotalOutstanding+"  _(validators haven't withdrawn yet)_")
	}

	w.Subsection("Fee market (x/feemarket)")
	w.Hint("Live status, receipt walkthrough, and wallet vs chain view from feemarket REST, CometBFT block_results, and eth_gasPrice. Payout path is in the Overview diagram above.")
	writeFeemarketSection(w, d)

	w.Row("model", "EIP-1559  _(base fee rises when blocks are full, falls when empty)_")
	if d.BaseFee != "" {
		w.Row("current base fee", d.BaseFee)
	}
	if d.GasPrice != "" {
		w.Row("current gas price", d.GasPrice+"  _(from JSON-RPC eth_gasPrice)_")
	}
	if d.MinGasPrice != "" {
		w.Row("min gas price", d.MinGasPrice+"  _(chain-enforced floor)_")
	}
	if d.BlockGas != "" {
		w.Row("gas used (last block)", d.BlockGas)
	}
	if d.Elasticity > 0 {
		w.Row("block gas target", fmt.Sprintf("max_block_gas ÷ %d", d.Elasticity))
	}
	if d.AdjCap != "" {
		w.Row("base fee max change", d.AdjCap)
	}
	if d.BaseFeeChangeDenominator > 0 {
		w.Row("change denominator", fmt.Sprintf("%d", d.BaseFeeChangeDenominator))
	}
	noBaseFeeStr := report.BoolStr(d.NoBaseFee)
	if d.NoBaseFee {
		noBaseFeeStr += "  _(EIP-1559 enforcement disabled)_"
	} else {
		noBaseFeeStr += "  _(EIP-1559 active)_"
	}
	w.Row("no_base_fee flag", noBaseFeeStr)

	w.Subsection("PMT Rewards  (x/pmtrewards — custom)")
	w.Hint("`status`, `reward rate`, pool address → `GET /cosmos/evm/pmtrewards/v1/params`; `pool balance` → `x/bank` balances for pool address; runway/emissions derived in pmtop.")
	w.Row("status", pmtStatus(d))
	if d.PMTRate != "" {
		w.Row("reward rate", d.PMTRate+"  _(extra tokens per block from PMT pool)_")
	}
	if d.PMTAnnual != "" {
		w.Row("annual emissions", d.PMTAnnual)
	}
	if d.PMTDailyEmit != "" {
		w.Row("daily emissions", d.PMTDailyEmit)
	}
	if d.PMTPoolEmpty {
		w.Row("pool balance", "0  — pool empty, no PMT rewards distributing")
	} else if d.PMTBalance != "" {
		bal := d.PMTBalance
		if d.PMTRunway != "" {
			bal += "  (" + d.PMTRunway + ")"
		}
		w.Row("pool balance", bal)
	}
	if d.PMTPoolAddress != "" {
		w.Row("pool address", d.PMTPoolAddress)
	}
}

func writeEVMSection(w Writer, d model.Report) {
	w.Section("7. EVM JSON-RPC")
	w.Em("Wallet and dApp connectivity (`eth_*`, `net_*`, `txpool_*`) on this node's JSON-RPC.")
	writeEVMRPCSection(w, d)
	w.BlankLine()
}
