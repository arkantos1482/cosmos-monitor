package panel

import (
	"fmt"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeEconomics(w Writer, d model.Report) {
	w.Section("5. ECONOMICS")
	w.Em("Block rewards: PMT pool and inflation mint into `fee_collector`; tx fees join there; `x/distribution` splits each BeginBlock to the community pool and validators.")

	writeEconomicsOverview(w, d)

	w.Subsection("Chain parameters (reference)")
	writeEconomicsReference(w, d)
}

func writeEconomicsReference(w Writer, d model.Report) {
	w.Subsection("Staking pool")
	w.Hint("`bond denom`, `unbonding time`, `max validators` → REST GET /cosmos/staking/v1beta1/params; `total supply` → module x/bank supply; `bonded`, `not bonded` → REST GET …/pool.")
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

	w.Subsection("Slashing params")
	w.Hint("`signed blocks window`, `min signed`, slash fractions → REST GET /cosmos/slashing/v1beta1/params.")
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

	w.Subsection("Mint / inflation params")
	w.Hint("`annual provisions` → REST GET /cosmos/mint/v1beta1/annual-provisions; `goal bonded`, `blocks / year` → REST GET …/params.")
	if d.AnnualProvisions != "" {
		w.Row("annual provisions", d.AnnualProvisions+"  _(absolute new tokens/year if inflation active)_")
	}
	w.Row("goal bonded", fmt.Sprintf("%.0f%%  _(target stake ratio — inflation adjusts toward this)_", d.GoalBonded))
	if d.BlocksPerYear != "" {
		w.Row("blocks / year", d.BlocksPerYear)
	}

	w.Subsection("Distribution params")
	w.Hint("`community tax` → REST GET /cosmos/distribution/v1beta1/params; live balances, per-block split → ledger.")
	if d.CommunityTax != "" {
		w.Row("community tax", d.CommunityTax+"  _(%% of block rewards → community pool)_")
	}

	w.Subsection("Fee market (x/feemarket)")
	writeFeemarketSection(w, d)

	w.Subsection("PMT Rewards  (x/pmtrewards)")
	w.Hint("`status`, `pool address` → REST GET /cosmos/evm/pmtrewards/v1/params; `per-block rate`, `pool balance` → ledger (Block reward ledger above).")
	w.Row("status", pmtStatus(d))
	if d.PMTAnnual != "" {
		w.Row("annual emissions", d.PMTAnnual)
	}
	if d.PMTPoolAddress != "" {
		w.Row("pool address", d.PMTPoolAddress)
	}
	for _, m := range d.ModuleAccounts {
		if m.Address != "" {
			w.Row(m.Name+" address", m.Address)
		}
	}
}

func writeEVMSection(w Writer, d model.Report) {
	w.Section("7. EVM JSON-RPC")
	w.Em("Wallet and dApp connectivity (`eth_*`, `net_*`, `txpool_*`) on this node's JSON-RPC.")
	writeEVMRPCSection(w, d)
	w.BlankLine()
}
