package panel

import (
	"fmt"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func writeNode(w Writer, d model.Report) {
	lv := d.Local
	syncStr := "synced"
	if !d.Synced {
		syncStr = "CATCHING UP"
	}

	w.Section("2. VALIDATOR")
	w.Em("This machine's validator — staking and rewards, block signing, operator identity, live consensus, and node process.")

	if lv.IsValidator {
		w.Subsection("Staking")
		w.Hint("`status`, `jailed`, `voting power`, `commission` → module x/staking validators (matched via CometBFT GET /status validator_info consensus address); `outstanding rewards`, `commission earned` → REST GET /cosmos/distribution/v1beta1/validators/{valoper}/outstanding_rewards, …/commission.")
		w.Row("status", lv.Status)
		if lv.Jailed {
			w.Row("jailed", "yes")
		}
		if lv.Tombstoned {
			w.Row("tombstoned", "YES")
		}
		w.Row("voting power", fmt.Sprintf("%s  (%.1f%% of bonded stake)", lv.VotingPower, lv.VPPercent))
		w.Row("commission", fmt.Sprintf("%.1f%%  _(your cut of delegator rewards)_", lv.Commission))
		if lv.Outstanding != "" {
			w.Row("outstanding rewards", lv.Outstanding+"  _(total rewards not yet withdrawn — x/distribution)_")
		} else {
			w.Row("outstanding rewards", "–")
		}
		if lv.CommissionEarned != "" {
			w.Row("commission earned", lv.CommissionEarned+"  _(validator commission, unclaimed — x/distribution)_")
		} else {
			w.Row("commission earned", "–")
		}

		w.Subsection("Block Signing")
		w.Hint("`signing health`, `missed / window` → REST GET /cosmos/slashing/v1beta1/signing_infos + params; `proposer`, `proposer priority` → CometBFT GET /validators.")
		w.Row("signing health", lv.SigningStatus)
		if d.SlashWindow != "" && d.SlashWindow != "0" {
			w.Row("missed / window", fmt.Sprintf("%d / %s blocks  (max allowed: %d)", lv.Missed, d.SlashWindow, lv.MaxMissed))
		}
		if lv.IsNextProposer {
			w.Row("proposer", "**next block proposer**")
		} else if d.NextProposer != "" {
			w.Row("proposer", "not next  _(next: "+d.NextProposer+")_")
		}
		if lv.ProposerPriority != 0 {
			w.Row("proposer priority", report.FormatInt(lv.ProposerPriority))
		}

		w.Subsection("Operator")
		w.Hint("`operator address`, `moniker` → module x/staking (matched via CometBFT GET /status validator_info consensus address).")
		if lv.OperatorAddr != "" {
			w.Row("operator address", lv.OperatorAddr+"  _(staking / rewards — `evmd query/distribution` use this)_")
		}
		w.Row("moniker", lv.Moniker)
	} else {
		w.Subsection("Operator")
		w.Hint("`role`, `moniker` → CometBFT GET /status (node_info; validator_info when present); `role` derived when consensus address is not in bonded x/staking set.")
		w.Row("role", lv.SigningStatus)
		w.Row("moniker", d.Moniker)
	}

	w.Subsection("Consensus")
	w.Hint("`sync`, `height`, `last block`, `interval` → CometBFT GET /status, GET /block (H−1 via ?height=); `consensus address`, `voting power` → CometBFT GET /status validator_info; `mempool` → CometBFT GET /num_unconfirmed_txs.")
	w.Row("sync", syncStr)
	w.Row("height", d.BlockHeight)
	if d.BlockInterval != "" {
		w.Row("interval", d.BlockInterval)
	}
	if d.TimeSinceBlock != "" {
		w.Row("last block", d.TimeSinceBlock)
	}
	if d.LocalConsensusAddr != "" {
		w.Row("consensus address", strings.ToUpper(d.LocalConsensusAddr)+"  _(hex; signs blocks — `/status` validator_info)_")
	}
	if d.LocalVotingPower != "" {
		w.Row("voting power", d.LocalVotingPower+"  _(consensus units — `/status` validator_info)_")
	}
	w.Row("mempool", fmt.Sprintf("%d pending", d.MempoolTxs))

	w.Subsection("Node")
	w.Hint("`node ID`, `version`, `chain ID`, `p2p listen`, `rpc listen` → CometBFT GET /status (node_info only; sync_info and validator_info in Consensus).")
	if d.NodeID != "" {
		w.Row("node ID", d.NodeID+"  _(CometBFT P2P peer ID)_")
	}
	if d.AppVersion != "" {
		w.Row("version", d.AppVersion)
	}
	if d.Network != "" {
		w.Row("chain ID", d.Network)
	}
	if d.ListenAddr != "" {
		w.Row("p2p listen", d.ListenAddr+"  _(advertised dial address from `/status`)_")
	}
	if d.RpcListenAddr != "" {
		w.Row("rpc listen", d.RpcListenAddr)
	}
}
