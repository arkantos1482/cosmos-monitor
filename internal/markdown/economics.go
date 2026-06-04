package markdown

import (
	"fmt"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func writeEconomics(m *mdWriter, d model.Report) {
	// ── 5. ECONOMICS ─────────────────────────────────────────────────────────
	m.section("5. ECONOMICS")

	fmt.Fprintf(m.w, "_How money moves on this chain — tx fees, PMT pool rewards, and (if active) inflation accumulate in `fee_collector`, then `x/distribution` pays validators each block._\n\n")

	m.subsection("Overview")
	m.hint("Mermaid diagram (below). Coins: tx fees / inflation / PMT → `fee_collector` → `x/distribution`. `x/staking` supplies voting power (no coin flow). Payout split: community tax, then per-validator commission vs delegators. `goal bonded` → `GET /cosmos/mint/v1beta1/params`. Preview in VS Code/Obsidian or use `pmtop --web`.")
	writeEconomicsDiagram(m.w, d)
	if d.PMTEnabled {
		fmt.Fprintf(m.w, "_PMT pool funds per-block rewards via mint hook → `fee_collector` (see PMT Rewards table below)._\n\n")
	}

	m.subsection("Staking Pool")
	m.hint("`bond denom`, `unbonding time`, `max validators` → `GET /cosmos/staking/v1beta1/params`; `total supply` → `x/bank` supply; `bonded` / `not bonded` → `GET /cosmos/staking/v1beta1/pool`.")
	if d.BondDenom != "" {
		m.row("bond denom", d.BondDenom)
	}
	m.row("total supply", d.TotalSupply)
	m.row("bonded", fmt.Sprintf("%s  (%.2f%%, goal %.0f%%)", d.BondedAmt, d.BondedPct, d.GoalBonded))
	m.row("not bonded", d.NotBonded)
	if d.UnbondingTime != "" {
		m.row("unbonding time", d.UnbondingTime+"  _(time locked after unstaking)_")
	}
	if d.MaxValidators > 0 {
		m.row("max validators", fmt.Sprintf("%d", d.MaxValidators))
	}

	m.subsection("Slashing Params")
	m.hint("`signed blocks window`, `min signed`, slash fractions → `GET /cosmos/slashing/v1beta1/params`.")
	if d.SlashWindow != "" && d.SlashWindow != "0" {
		m.row("signed blocks window", d.SlashWindow+" blocks")
	}
	if d.MinSigned > 0 {
		m.row("min signed per window", fmt.Sprintf("%.1f%%  _(miss more → downtime slash risk)_", d.MinSigned))
	}
	if d.SlashDowntime != "" {
		dtStr := d.SlashDowntime
		if d.SlashDTInactive {
			dtStr += "  ⚠ inactive"
		}
		m.row("slash / downtime", dtStr)
	}
	if d.SlashDS != "" {
		dsStr := d.SlashDS
		if d.SlashDSInactive {
			dsStr += "  ⚠ inactive"
		}
		m.row("slash / double-sign", dsStr)
	}

	m.subsection("Staking & Inflation  (x/mint + x/staking)")
	m.hint("`inflation rate` → `GET /cosmos/mint/v1beta1/inflation`; `annual provisions` → `…/annual-provisions`; `goal bonded`, `blocks / year` → `…/params`; `unbonding time` → `x/staking` params.")
	inflationStr := fmt.Sprintf("%.2f%%", d.Inflation)
	if d.Inflation == 0 {
		inflationStr += "  ⚠ inactive"
	}
	m.row("inflation rate", inflationStr+"  _(extra tokens minted when active — rewards stakers)_")
	if d.AnnualProvisions != "" {
		m.row("annual provisions", d.AnnualProvisions+"  _(absolute new tokens/year if inflation active)_")
	}
	m.row("goal bonded", fmt.Sprintf("%.0f%%  _(target stake ratio — inflation adjusts toward this)_", d.GoalBonded))
	if d.BlocksPerYear != "" {
		m.row("blocks / year", d.BlocksPerYear)
	}
	if d.UnbondingTime != "" {
		m.row("unbonding time", d.UnbondingTime+"  _(tokens locked after you unstake)_")
	}

	m.subsection("Distribution  (x/distribution)")
	m.hint("`community tax` → `GET /cosmos/distribution/v1beta1/params`; `community pool` → `…/community_pool`; `unclaimed staking rewards` → sum of per-validator `…/validators/{valoper}/outstanding_rewards`.")
	m.row("community tax", d.CommunityTax+"  _(%% of block rewards → community pool, not validators)_")
	m.row("community pool", d.CommunityPool+"  _(governance-controlled treasury)_")
	if d.TotalOutstanding != "" {
		m.row("unclaimed staking rewards", d.TotalOutstanding+"  _(validators haven't withdrawn yet)_")
	}

	m.subsection("Fee market (x/feemarket)")
	m.hint("KaTeX math (`$$` blocks below) and Mermaid diagram use live values from feemarket REST, CometBFT `block_results`, and `eth_gasPrice`. Payout path is in Overview above.")
	writeFeemarketSection(m.w, d)

	m.row("model", "EIP-1559  _(base fee rises when blocks are full, falls when empty)_")
	if d.BaseFee != "" {
		m.row("current base fee", d.BaseFee)
	}
	if d.GasPrice != "" {
		m.row("current gas price", d.GasPrice+"  _(from JSON-RPC eth_gasPrice)_")
	}
	if d.MinGasPrice != "" {
		m.row("min gas price", d.MinGasPrice+"  _(chain-enforced floor)_")
	}
	if d.BlockGas != "" {
		m.row("gas used (last block)", d.BlockGas)
	}
	if d.Elasticity > 0 {
		m.row("block gas target", fmt.Sprintf("max_block_gas ÷ %d", d.Elasticity))
	}
	if d.AdjCap != "" {
		m.row("base fee max change", d.AdjCap)
	}
	if d.BaseFeeChangeDenominator > 0 {
		m.row("change denominator", fmt.Sprintf("%d", d.BaseFeeChangeDenominator))
	}
	noBaseFeeStr := report.BoolStr(d.NoBaseFee)
	if d.NoBaseFee {
		noBaseFeeStr += "  _(EIP-1559 enforcement disabled)_"
	} else {
		noBaseFeeStr += "  _(EIP-1559 active)_"
	}
	m.row("no_base_fee flag", noBaseFeeStr)

	m.subsection("PMT Rewards  (x/pmtrewards — custom)")
	m.hint("`status`, `reward rate`, pool address → `GET /cosmos/evm/pmtrewards/v1/params`; `pool balance` → `x/bank` balances for pool address; runway/emissions derived in pmtop.")
	m.row("status", mdPMTStatus(d))
	if d.PMTRate != "" {
		m.row("reward rate", d.PMTRate+"  _(extra tokens per block from PMT pool)_")
	}
	if d.PMTAnnual != "" {
		m.row("annual emissions", d.PMTAnnual)
	}
	if d.PMTDailyEmit != "" {
		m.row("daily emissions", d.PMTDailyEmit)
	}
	if d.PMTPoolEmpty {
		m.row("pool balance", "0  — pool empty, no PMT rewards distributing")
	} else if d.PMTBalance != "" {
		bal := d.PMTBalance
		if d.PMTRunway != "" {
			bal += "  (" + d.PMTRunway + ")"
		}
		m.row("pool balance", bal)
	}
	if d.PMTPoolAddress != "" {
		m.row("pool address", d.PMTPoolAddress)
	}
}
