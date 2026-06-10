package panel

import (
	"fmt"

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
	w.Em("This node — application staking, rewards, and slashing; CometBFT consensus state and P2P identity.")

	if lv.IsValidator {
		writeNodeApplication(w, d, lv)
	} else {
		w.Layer("Application (Cosmos SDK / ABCI state)")
		w.Subsection("Identity")
		w.Hint("`role`, `moniker` → CometBFT GET /status; role derived when consensus address is absent from x/staking.")
		w.Row("role", lv.SigningStatus)
		w.Row("moniker", d.Moniker)
	}

	writeNodeCometBFT(w, d, lv, syncStr)
}

func writeNodeApplication(w Writer, d model.Report, lv model.LocalValidator) {
	w.Layer("Application (Cosmos SDK / ABCI state)")

	w.Subsection("Identity")
	w.Hint("`operator`, `account` → x/staking + GET /cosmos/evm/vm/v1/validator_account/{cons_address}; `moniker` → x/staking description.")
	if lv.OperatorAddr != "" {
		w.Row("operator", lv.OperatorAddr+"  _(valoper — staking queries, distribution txs)_")
	}
	if lv.AccountAddr != "" {
		w.Row("account", formatBech32EVM(lv.AccountAddr, lv.EVMAddr)+"  _(self-delegation / operator wallet)_")
	} else {
		w.Row("account", "–  _(validator_account API unavailable)_")
	}
	w.Row("moniker", lv.Moniker)

	w.Subsection("Staking")
	w.Hint("`status`, `jailed`, `voting power`, `commission` → REST GET /cosmos/staking/v1beta1/validators.")
	w.Row("status", lv.Status)
	if lv.Jailed {
		w.Row("jailed", "yes")
	}
	if lv.Tombstoned {
		w.Row("tombstoned", "YES")
	}
	w.Row("voting power", fmt.Sprintf("%s  (%.1f%% of bonded stake)", lv.VotingPower, lv.VPPercent))
	w.Row("commission", fmt.Sprintf("%.1f%%  _(validator cut of delegator rewards)_", lv.Commission))

	w.Subsection("Rewards")
	w.Hint("`outstanding rewards`, `commission earned` → REST GET /cosmos/distribution/v1beta1/validators/{valoper}/outstanding_rewards, …/commission.")
	if lv.Outstanding != "" {
		w.Row("outstanding rewards", lv.Outstanding+"  _(total unclaimed — x/distribution)_")
	} else {
		w.Row("outstanding rewards", "–")
	}
	if lv.CommissionEarned != "" {
		w.Row("commission earned", lv.CommissionEarned+"  _(unclaimed validator commission)_")
	} else {
		w.Row("commission earned", "–")
	}

	w.Subsection("Slashing")
	w.Hint("`signing health`, `missed / window` → REST GET /cosmos/slashing/v1beta1/signing_infos + params.")
	w.Row("signing health", lv.SigningStatus)
	if d.SlashWindow != "" && d.SlashWindow != "0" {
		w.Row("missed / window", fmt.Sprintf("%d / %s blocks  (max allowed: %d)", lv.Missed, d.SlashWindow, lv.MaxMissed))
	}
}

func writeNodeCometBFT(w Writer, d model.Report, lv model.LocalValidator, syncStr string) {
	w.Layer("CometBFT (consensus + networking)")

	w.Subsection("Live state")
	w.Hint("`sync`, `height`, `last block`, `interval` → CometBFT GET /status, GET /block; `mempool` → GET /num_unconfirmed_txs.")
	w.Row("sync", syncStr)
	w.Row("height", d.BlockHeight)
	if d.BlockInterval != "" {
		w.Row("interval", d.BlockInterval)
	}
	if d.TimeSinceBlock != "" {
		w.Row("last block", d.TimeSinceBlock)
	}
	w.Row("mempool", fmt.Sprintf("%d pending", d.MempoolTxs))

	if lv.IsValidator || d.LocalConsensusAddr != "" {
		w.Subsection("Validator key")
		w.Hint("`consensus` → x/slashing signing_infos or derived from staking pubkey; `voting power` → CometBFT GET /status validator_info; `proposer priority` → GET /validators.")
		if lv.ConsensusBech32 != "" {
			w.Row("consensus", lv.ConsensusBech32+"  _(valcons — signs blocks)_")
		} else if lv.ConsensusAddr != "" {
			w.Row("consensus", "–  _(valcons unavailable)_")
		}
		if d.LocalVotingPower != "" {
			w.Row("voting power", d.LocalVotingPower+"  _(consensus units — `/status` validator_info)_")
		}
		w.Row("proposer priority", report.FormatInt(lv.ProposerPriority))
		if lv.IsNextProposer {
			w.Row("proposer", "**next block proposer**")
		} else if d.NextProposer != "" {
			w.Row("proposer", "not next  _(next: "+d.NextProposer+")_")
		}
	}

	w.Subsection("P2P & RPC")
	w.Hint("`node ID`, `p2p listen`, `p2p dial`, `rpc listen` → CometBFT GET /status (node_info); dial is node_id@listen_addr.")
	if d.NodeID != "" {
		w.Row("node ID", d.NodeID+"  _(CometBFT peer ID)_")
	}
	p2pDial := lv.P2PDial
	if p2pDial == "" {
		p2pDial = formatP2PDial(d.NodeID, d.ListenAddr)
	}
	if p2pDial != "" {
		w.Row("p2p dial", p2pDial+"  _(peer dial string)_")
	}
	if d.ListenAddr != "" {
		w.Row("p2p listen", d.ListenAddr+"  _(advertised from `/status`)_")
	}
	if d.RpcListenAddr != "" {
		w.Row("rpc listen", d.RpcListenAddr)
	}
	if d.AppVersion != "" {
		w.Row("version", d.AppVersion)
	}
	if d.Network != "" {
		w.Row("chain ID", d.Network)
	}
}

func formatBech32EVM(bech, evm string) string {
	if bech == "" {
		return "–"
	}
	if evm != "" {
		return bech + "  _(EVM: " + evm + ")_"
	}
	return bech
}

func formatP2PDial(nodeID, listen string) string {
	if nodeID == "" || listen == "" {
		return ""
	}
	addr := listen
	if len(addr) > 6 && addr[:6] == "tcp://" {
		addr = addr[6:]
	}
	return nodeID + "@" + addr
}
