package panel

import (
	"fmt"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeEconomics(w Writer, d model.Report) {
	w.Section("2. ECONOMICS")
	writeEconomicsSummary(w, d, SummaryEmbedded)
	w.Em("Chain-wide tokenomics — block rewards flow through `fee_collector` and `x/distribution` to the community pool and validators.")

	writeEconomicsOverview(w, d)

	w.Subsection("Chain parameters (reference)")
	writeEconomicsReference(w, d)
}

func writeEconomicsReference(w Writer, d model.Report) {
	w.Subsection("Staking params")
	w.Hint("`bond denom`, `unbonding time`, `max validators` → REST GET /cosmos/staking/v1beta1/params; `total supply` → module x/bank supply.")
	if d.BondDenom != "" {
		w.Row("bond denom", d.BondDenom)
	}
	w.Row("total supply", d.TotalSupply)
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

	w.Subsection("Mint params")
	w.Hint("`annual provisions` → REST GET /cosmos/mint/v1beta1/annual-provisions; `goal bonded`, `blocks / year` → REST GET …/params.")
	if d.AnnualProvisions != "" {
		prov := d.AnnualProvisions + "  _(absolute new tokens/year if inflation active)_"
		if d.Inflation <= 0 {
			prov += "  ⚠ inactive"
		}
		w.Row("annual provisions", prov)
	}
	w.Row("goal bonded", fmt.Sprintf("%.0f%%  _(target stake ratio — inflation adjusts toward this)_", d.GoalBonded))
	if d.BlocksPerYear != "" {
		w.Row("blocks / year", d.BlocksPerYear)
	}

	w.Subsection("PMT Rewards  (x/pmtrewards)")
	w.Hint("`pool address` → REST GET /cosmos/evm/pmtrewards/v1/params; `per-block rate`, `pool balance`, `status` → domain cards above.")
	if d.PMTPoolAddress != "" {
		w.Row("pool address", d.PMTPoolAddress)
	}
	if d.PMTAnnual != "" {
		w.Row("annual emissions", d.PMTAnnual)
	}
}

func writeEVMSection(w Writer, d model.Report) {
	w.Section("3. EVM JSON-RPC")
	writeEVMSummary(w, d, SummaryEmbedded)
	w.Em("Wallet and dApp connectivity (`eth_*`, `net_*`, `txpool_*`) on this node's JSON-RPC.")
	writeEVMRPCSection(w, d)
	w.BlankLine()
}
