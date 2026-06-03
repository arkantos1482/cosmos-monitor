package fetch

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ChainSnapshot holds all chain-level data.
type ChainSnapshot struct {
	NodeID     string
	Moniker    string
	AppVersion string
	ListenAddr    string
	RpcListenAddr string
	Network       string

	BlockHeight     int64
	LatestBlockTime time.Time
	BlockInterval   time.Duration
	CatchingUp      bool
	PeerCount       int
	PeerMonikers        []string
	MempoolTxs          int
	NextProposerMoniker string

	// This node's validator identity from /status (empty if full node).
	LocalConsensusAddr string
	LocalVotingPower   int64

	Validators []ValidatorInfo

	BondedTokens     string
	NotBondedTokens  string
	TotalSupply      string
	TotalSupplyDenom string
	Inflation        float64
	AnnualProvisions string
	CommunityPool    string

	BaseFee  string
	BlockGas uint64

	VotingProposals []ProposalInfo
	DepositProposals []ProposalInfo

	UpgradeName   string
	UpgradeHeight int64

	TokenPairs []TokenPairInfo

	IBCClientCount int

	Params ChainParams

	Err error
}

// ProposalTally holds the vote counts for a proposal.
type ProposalTally struct {
	Yes        string
	No         string
	Abstain    string
	NoWithVeto string
}

// ValidatorInfo holds per-validator data.
type ValidatorInfo struct {
	Moniker            string
	OperatorAddr       string
	ConsensusAddr      string
	NodeID             string // CometBFT peer ID (from /net_info or /status for this node)
	P2PDial            string // node_id@listen_addr when known from chain RPC
	P2PConnected       bool   // peer visible in this node's /net_info (or is this node)
	Status             string
	Jailed             bool
	Tombstoned         bool
	VotingPowerTokens  string
	VotingPowerPercent float64
	Commission         float64
	MissedBlocks           int64
	OutstandingRewards     string // formatted display string
	OutstandingRewardsAmt  string // raw base-denom amount for summing
	OutstandingRewardsDenom string
	CommissionEarned       string
	ProposerPriority       int64
}

// ProposalInfo holds governance proposal data.
type ProposalInfo struct {
	ID         uint64
	Title      string
	Status     string
	VotingEnd  time.Time
	DepositEnd time.Time
	Tally      ProposalTally
}

// TokenPairInfo holds ERC20 token pair data.
type TokenPairInfo struct {
	Denom     string
	ERC20Addr string
	Enabled   bool
}

// ChainParams holds chain parameters fetched once on launch.
type ChainParams struct {
	UnbondingTime      time.Duration
	MaxValidators      int
	BondDenom          string
	SignedBlocksWindow int64
	MinSignedPerWindow float64
	SlashFractionDowntime   string
	SlashFractionDoubleSign string
	BlocksPerYear      int64
	GoalBonded         float64
	CommunityTax       float64
	VotingPeriod       time.Duration
	Quorum             float64
	Threshold          float64
	VetoThreshold      float64
	EVMDenom           string
	MinGasPrice        float64
	Elasticity         int64
	NoBaseFee                bool
	BaseFeeChangeDenominator int64
	ERC20Enabled       bool
	ActiveStaticPrecompiles []string
	HistoryServeWindow      int64
	HardforkLondon          string
	HardforkShanghai        string
	HardforkCancun          string
	// pmtrewards module
	RewardPerBlockAmount      string
	RewardPerBlockDenom       string
	PMTRewardsEnabled         bool
	PMTRewardsPoolAddress     string
	PMTRewardsPoolBalanceAmt  string // raw base-denom amount
	PMTRewardsPoolBalanceDenom string
}

// --- RPC response types ---

type statusResp struct {
	Result struct {
		NodeInfo struct {
			ID         string `json:"id"`
			Moniker    string `json:"moniker"`
			Version    string `json:"version"`
			ListenAddr string `json:"listen_addr"`
			Network    string `json:"network"`
			Other      struct {
				RPCAddress string `json:"rpc_address"`
			} `json:"other"`
		} `json:"node_info"`
		SyncInfo struct {
			LatestBlockHeight string `json:"latest_block_height"`
			LatestBlockTime   string `json:"latest_block_time"`
			CatchingUp        bool   `json:"catching_up"`
		} `json:"sync_info"`
		ValidatorInfo struct {
			Address     string `json:"address"`
			VotingPower string `json:"voting_power"`
		} `json:"validator_info"`
	} `json:"result"`
}

type netInfoResp struct {
	Result struct {
		NPeers string `json:"n_peers"`
		Peers  []struct {
			NodeInfo struct {
				ID         string `json:"id"`
				Moniker    string `json:"moniker"`
				ListenAddr string `json:"listen_addr"`
			} `json:"node_info"`
		} `json:"peers"`
	} `json:"result"`
}

type numUnconfirmedTxsResp struct {
	Result struct {
		NTxs string `json:"n_txs"`
	} `json:"result"`
}

type validatorsResp struct {
	Result struct {
		Validators []struct {
			Address          string `json:"address"`
			ProposerPriority string `json:"proposer_priority"`
			PubKey           struct {
				Value string `json:"value"` // base64 raw key bytes
			} `json:"pub_key"`
		} `json:"validators"`
	} `json:"result"`
}

type blockResp struct {
	Result struct {
		Block struct {
			Header struct {
				Height string `json:"height"`
				Time   string `json:"time"`
			} `json:"header"`
		} `json:"block"`
	} `json:"result"`
}

// --- REST response types ---

type stakingValidatorsResp struct {
	Validators []struct {
		OperatorAddress string `json:"operator_address"`
		Description     struct {
			Moniker string `json:"moniker"`
		} `json:"description"`
		Status string `json:"status"`
		Tokens string `json:"tokens"`
		Commission struct {
			CommissionRates struct {
				Rate string `json:"rate"`
			} `json:"commission_rates"`
		} `json:"commission"`
		Jailed         bool `json:"jailed"`
		ConsensusPubkey struct {
			Key string `json:"key"` // base64 raw pubkey bytes
		} `json:"consensus_pubkey"`
	} `json:"validators"`
}

type stakingPoolResp struct {
	Pool struct {
		BondedTokens    string `json:"bonded_tokens"`
		NotBondedTokens string `json:"not_bonded_tokens"`
	} `json:"pool"`
}

type signingInfosResp struct {
	Info []struct {
		Address            string `json:"address"`
		MissedBlocksCounter string `json:"missed_blocks_counter"`
		Tombstoned         bool   `json:"tombstoned"`
	} `json:"info"`
}

type supplyResp struct {
	Supply []struct {
		Denom  string `json:"denom"`
		Amount string `json:"amount"`
	} `json:"supply"`
}

type inflationResp struct {
	Inflation string `json:"inflation"`
}

type annualProvisionsResp struct {
	AnnualProvisions string `json:"annual_provisions"`
}

type communityPoolResp struct {
	Pool []struct {
		Denom  string `json:"denom"`
		Amount string `json:"amount"`
	} `json:"pool"`
}

type baseFeeResp struct {
	BaseFee string `json:"base_fee"`
}

type blockGasResp struct {
	Gas string `json:"gas"`
}

type proposalsResp struct {
	Proposals []struct {
		ProposalID string `json:"proposal_id"` // v1beta1
		ID         string `json:"id"`          // v1
		Title      string `json:"title"`       // v1
		Content    struct {
			Title string `json:"title"`
		} `json:"content"` // v1beta1
		Status         string `json:"status"`
		VotingEndTime  string `json:"voting_end_time"`
		DepositEndTime string `json:"deposit_end_time"`
	} `json:"proposals"`
}

type proposalTallyResp struct {
	Tally struct {
		Yes        string `json:"yes"`
		No         string `json:"no"`
		Abstain    string `json:"abstain"`
		NoWithVeto string `json:"no_with_veto"`
	} `json:"tally"`
}

type upgradePlanResp struct {
	Plan *struct {
		Name   string `json:"name"`
		Height string `json:"height"`
	} `json:"plan"`
}

type tokenPairsResp struct {
	TokenPairs []struct {
		Denom    string `json:"denom"`
		Erc20Address string `json:"erc20_address"`
		Enabled  bool   `json:"enabled"`
	} `json:"token_pairs"`
}

type ibcClientsResp struct {
	ClientStates []struct{} `json:"client_states"`
}

type validatorRewardsResp struct {
	Rewards struct {
		Rewards []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"rewards"`
	} `json:"rewards"`
}

type validatorCommissionResp struct {
	Commission struct {
		Commission []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"commission"`
	} `json:"commission"`
}

// --- params response types ---

type stakingParamsResp struct {
	Params struct {
		UnbondingTime string `json:"unbonding_time"`
		MaxValidators int    `json:"max_validators"`
		BondDenom     string `json:"bond_denom"`
	} `json:"params"`
}

type slashingParamsResp struct {
	Params struct {
		SignedBlocksWindow      string `json:"signed_blocks_window"`
		MinSignedPerWindow      string `json:"min_signed_per_window"`
		SlashFractionDowntime   string `json:"slash_fraction_downtime"`
		SlashFractionDoubleSign string `json:"slash_fraction_double_sign"`
	} `json:"params"`
}

type mintParamsResp struct {
	Params struct {
		BlocksPerYear string `json:"blocks_per_year"`
		GoalBonded    string `json:"goal_bonded"`
	} `json:"params"`
}

type govVotingParamsResp struct {
	VotingParams struct {
		VotingPeriod string `json:"voting_period"`
	} `json:"voting_params"`
}

type govTallyParamsResp struct {
	TallyParams struct {
		Quorum        string `json:"quorum"`
		Threshold     string `json:"threshold"`
		VetoThreshold string `json:"veto_threshold"`
	} `json:"tally_params"`
}

type feemarketParamsResp struct {
	Params struct {
		NoBaseFee                bool   `json:"no_base_fee"`
		BaseFeeChangeDenominator int64  `json:"base_fee_change_denominator"`
		MinGasPrice              string `json:"min_gas_price"`
		ElasticityMultiplier     int64  `json:"elasticity_multiplier"`
	} `json:"params"`
}

type evmParamsResp struct {
	Params struct {
		EvmDenom                string   `json:"evm_denom"`
		ActiveStaticPrecompiles []string `json:"active_static_precompiles"`
		HistoryServeWindow      int64    `json:"history_serve_window"`
	} `json:"params"`
}

type erc20ParamsResp struct {
	Params struct {
		EnableErc20 bool `json:"enable_erc20"`
	} `json:"params"`
}

type distributionParamsResp struct {
	Params struct {
		CommunityTax string `json:"community_tax"`
	} `json:"params"`
}

type pmtRewardsParamsResp struct {
	Params struct {
		Enabled        bool   `json:"enabled"`
		RewardPerBlock struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"reward_per_block"`
		PoolAddress string `json:"pool_address"`
	} `json:"params"`
}

func parseDuration(s string) time.Duration {
	// Modern Cosmos SDK returns Go duration strings like "1814400s", "24h0m0s".
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}
	// Older versions returned nanoseconds as a plain integer.
	if ns, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Duration(ns)
	}
	return 0
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func parseInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

// convertDenom normalises a raw Cosmos coin amount to a human display unit.
// Recognises standard SI prefixes:
//
//	a (atto  = 10⁻¹⁸) → "apmt" becomes "PMT", divided by 1e18
//	n (nano  = 10⁻⁹)
//	u (micro = 10⁻⁶) → "uatom" becomes "ATOM", divided by 1e6
//	m (milli = 10⁻³)
func convertDenom(v float64, denom string) (float64, string) {
	if len(denom) == 0 {
		return v, denom
	}
	switch denom[0] {
	case 'a':
		return v / 1e18, strings.ToUpper(denom[1:])
	case 'n':
		return v / 1e9, strings.ToUpper(denom[1:])
	case 'u':
		return v / 1e6, strings.ToUpper(denom[1:])
	case 'm':
		return v / 1e3, strings.ToUpper(denom[1:])
	}
	return v, strings.ToUpper(denom)
}

// formatDisplayAmount renders a post-conversion float as a compact human string.
func formatDisplayAmount(v float64) string {
	switch {
	case v >= 1e12:
		return fmt.Sprintf("%.2fT", v/1e12)
	case v >= 1e9:
		return fmt.Sprintf("%.2fB", v/1e9)
	case v >= 1e6:
		return fmt.Sprintf("%.2fM", v/1e6)
	case v >= 1e3:
		return fmt.Sprintf("%.2fK", v/1e3)
	case v >= 0.001:
		return fmt.Sprintf("%.4f", v)
	case v >= 1e-7:
		// Fixed-point is easier to read in narrow columns than scientific notation.
		return fmt.Sprintf("%.6f", v)
	case v > 0:
		return fmt.Sprintf("%.2e", v)
	default:
		return "0"
	}
}

// FormatCoin converts a raw Cosmos amount string + on-chain denom to a
// human-readable display string, e.g. "400000000000000000000000000" + "apmt"
// → "400.00M PMT".
func FormatCoin(rawAmount, denom string) string {
	v, _ := strconv.ParseFloat(rawAmount, 64)
	displayVal, displayDenom := convertDenom(v, denom)
	if displayDenom == "" {
		return formatDisplayAmount(displayVal)
	}
	return formatDisplayAmount(displayVal) + " " + displayDenom
}

// NormalizeCoin returns the display float value and display denom for a raw
// Cosmos coin, useful when the caller needs the numeric value for calculations.
func NormalizeCoin(rawAmount, denom string) (float64, string) {
	v, _ := strconv.ParseFloat(rawAmount, 64)
	return convertDenom(v, denom)
}

func formatCoins(coins []struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}, preferDenom string) string {
	if len(coins) == 0 {
		return "0"
	}
	for _, c := range coins {
		if c.Denom == preferDenom || preferDenom == "" {
			return FormatCoin(c.Amount, c.Denom)
		}
	}
	return FormatCoin(coins[0].Amount, coins[0].Denom)
}

const bech32Charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

// bech32ToHex decodes a bech32 address and returns the raw bytes as a lowercase hex string.
// Checksum is not verified — addresses come from trusted API responses.
func bech32ToHex(bech string) string {
	lower := strings.ToLower(bech)
	pos := strings.LastIndex(lower, "1")
	if pos < 1 || pos+7 > len(lower) {
		return ""
	}
	// strip HRP, separator, and 6-char checksum
	encoded := lower[pos+1 : len(lower)-6]
	bits5 := make([]byte, len(encoded))
	for i, c := range encoded {
		idx := strings.IndexByte(bech32Charset, byte(c))
		if idx < 0 {
			return ""
		}
		bits5[i] = byte(idx)
	}
	// convert 5-bit groups to 8-bit bytes
	acc, bits := 0, 0
	var result []byte
	for _, b := range bits5 {
		acc = acc<<5 | int(b)
		bits += 5
		if bits >= 8 {
			bits -= 8
			result = append(result, byte(acc>>bits))
			acc &= (1 << bits) - 1
		}
	}
	return fmt.Sprintf("%x", result)
}

// consAddrFromPubkey computes the consensus address (20-byte SHA256 prefix, hex-encoded)
// from a base64-encoded raw ed25519 public key as returned by the Cosmos staking API.
func consAddrFromPubkey(pubkeyBase64 string) string {
	b, err := base64.StdEncoding.DecodeString(pubkeyBase64)
	if err != nil || len(b) == 0 {
		return ""
	}
	h := sha256.Sum256(b)
	return fmt.Sprintf("%x", h[:20])
}

// parseProposal converts a raw proposal entry from the proposals API into a ProposalInfo.
func parseProposal(p struct {
	ProposalID string `json:"proposal_id"`
	ID         string `json:"id"`
	Title      string `json:"title"`
	Content    struct {
		Title string `json:"title"`
	} `json:"content"`
	Status         string `json:"status"`
	VotingEndTime  string `json:"voting_end_time"`
	DepositEndTime string `json:"deposit_end_time"`
}) ProposalInfo {
	idStr := p.ProposalID
	if idStr == "" {
		idStr = p.ID
	}
	id, _ := strconv.ParseUint(idStr, 10, 64)
	title := p.Content.Title
	if title == "" {
		title = p.Title
	}
	info := ProposalInfo{
		ID:     id,
		Title:  title,
		Status: p.Status,
	}
	info.VotingEnd, _ = time.Parse(time.RFC3339Nano, p.VotingEndTime)
	info.DepositEnd, _ = time.Parse(time.RFC3339Nano, p.DepositEndTime)
	return info
}

func formatP2PDial(nodeID, listen string) string {
	if nodeID == "" || listen == "" {
		return ""
	}
	return nodeID + "@" + strings.TrimPrefix(listen, "tcp://")
}

func applyValidatorP2P(v *ValidatorInfo, localMoniker, localNodeID, localListen string, peers map[string]struct {
	nodeID string
	listen string
}) {
	if v.Moniker == localMoniker {
		v.NodeID = localNodeID
		v.P2PDial = formatP2PDial(localNodeID, localListen)
		v.P2PConnected = localNodeID != ""
		return
	}
	if p, ok := peers[v.Moniker]; ok {
		v.NodeID = p.nodeID
		v.P2PDial = formatP2PDial(p.nodeID, p.listen)
		v.P2PConnected = true
	}
}

// FetchChain fetches all chain data from CometBFT RPC and Cosmos REST.
func FetchChain(rpc, rest string) ChainSnapshot {
	snap := ChainSnapshot{}

	// --- CometBFT RPC ---
	var status statusResp
	if err := doJSON(rpc+"/status", &status); err != nil {
		snap.Err = fmt.Errorf("rpc status: %w", err)
		return snap
	}

	snap.NodeID     = status.Result.NodeInfo.ID
	snap.Moniker    = status.Result.NodeInfo.Moniker
	snap.AppVersion = status.Result.NodeInfo.Version
	snap.ListenAddr = status.Result.NodeInfo.ListenAddr
	snap.RpcListenAddr = status.Result.NodeInfo.Other.RPCAddress
	snap.Network = status.Result.NodeInfo.Network
	snap.BlockHeight = parseInt64(status.Result.SyncInfo.LatestBlockHeight)
	snap.CatchingUp  = status.Result.SyncInfo.CatchingUp
	if t, err := time.Parse(time.RFC3339Nano, status.Result.SyncInfo.LatestBlockTime); err == nil {
		snap.LatestBlockTime = t
	}
	snap.LocalConsensusAddr = strings.ToLower(status.Result.ValidatorInfo.Address)
	snap.LocalVotingPower = parseInt64(status.Result.ValidatorInfo.VotingPower)

	peerByMoniker := map[string]struct {
		nodeID string
		listen string
	}{}
	var netInfo netInfoResp
	if err := doJSON(rpc+"/net_info", &netInfo); err == nil {
		snap.PeerCount, _ = strconv.Atoi(netInfo.Result.NPeers)
		for _, p := range netInfo.Result.Peers {
			if p.NodeInfo.Moniker != "" {
				snap.PeerMonikers = append(snap.PeerMonikers, p.NodeInfo.Moniker)
				peerByMoniker[p.NodeInfo.Moniker] = struct {
					nodeID string
					listen string
				}{nodeID: p.NodeInfo.ID, listen: p.NodeInfo.ListenAddr}
			}
		}
	}

	var unconfirmed numUnconfirmedTxsResp
	if err := doJSON(rpc+"/num_unconfirmed_txs", &unconfirmed); err == nil {
		snap.MempoolTxs, _ = strconv.Atoi(unconfirmed.Result.NTxs)
	}

	// block interval: latest - previous
	var latestBlock, prevBlock blockResp
	if err := doJSON(rpc+"/block", &latestBlock); err == nil {
		prevHeight := parseInt64(latestBlock.Result.Block.Header.Height) - 1
		if prevHeight > 0 {
			if err := doJSON(fmt.Sprintf("%s/block?height=%d", rpc, prevHeight), &prevBlock); err == nil {
				t1, e1 := time.Parse(time.RFC3339Nano, latestBlock.Result.Block.Header.Time)
				t2, e2 := time.Parse(time.RFC3339Nano, prevBlock.Result.Block.Header.Time)
				if e1 == nil && e2 == nil {
					snap.BlockInterval = t1.Sub(t2)
				}
			}
		}
	}

	// Validator set from RPC (for proposer priority); keyed by lowercase hex consensus address
	var valSet validatorsResp
	rpcVals := map[string]int64{}
	if err := doJSON(rpc+"/validators?per_page=100", &valSet); err == nil {
		for _, v := range valSet.Result.Validators {
			rpcVals[strings.ToLower(v.Address)] = parseInt64(v.ProposerPriority)
		}
	}

	// --- Cosmos REST ---

	// Staking validators (all statuses)
	allVals := map[string]ValidatorInfo{}
	for _, status := range []string{"BOND_STATUS_BONDED", "BOND_STATUS_UNBONDING", "BOND_STATUS_UNBONDED"} {
		var r stakingValidatorsResp
		if err := doJSON(fmt.Sprintf("%s/cosmos/staking/v1beta1/validators?status=%s&pagination.limit=100", rest, status), &r); err == nil {
			for _, v := range r.Validators {
				statusLabel := strings.TrimPrefix(v.Status, "BOND_STATUS_")
				info := ValidatorInfo{
					Moniker:           v.Description.Moniker,
					OperatorAddr:      v.OperatorAddress,
					ConsensusAddr:     consAddrFromPubkey(v.ConsensusPubkey.Key),
					Status:            statusLabel,
					Jailed:            v.Jailed,
					VotingPowerTokens: v.Tokens,
					Commission:        parseFloat(v.Commission.CommissionRates.Rate),
				}
				allVals[v.OperatorAddress] = info
			}
		}
	}

	// staking pool
	var pool stakingPoolResp
	if err := doJSON(rest+"/cosmos/staking/v1beta1/pool", &pool); err == nil {
		snap.BondedTokens = pool.Pool.BondedTokens
		snap.NotBondedTokens = pool.Pool.NotBondedTokens
	}

	// slashing signing infos — keyed by lowercase hex consensus address (decoded from bech32)
	sigInfoMap := map[string]struct {
		missed     int64
		tombstoned bool
	}{}
	var sigInfos signingInfosResp
	if err := doJSON(rest+"/cosmos/slashing/v1beta1/signing_infos?pagination.limit=100", &sigInfos); err == nil {
		for _, si := range sigInfos.Info {
			hexAddr := bech32ToHex(si.Address)
			if hexAddr == "" {
				hexAddr = strings.ToLower(si.Address) // fallback: already hex
			}
			sigInfoMap[hexAddr] = struct {
				missed     int64
				tombstoned bool
			}{parseInt64(si.MissedBlocksCounter), si.Tombstoned}
		}
	}

	// total supply
	var supply supplyResp
	if err := doJSON(rest+"/cosmos/bank/v1beta1/supply", &supply); err == nil && len(supply.Supply) > 0 {
		snap.TotalSupply = supply.Supply[0].Amount
		snap.TotalSupplyDenom = supply.Supply[0].Denom
	}

	// inflation
	var inf inflationResp
	if err := doJSON(rest+"/cosmos/mint/v1beta1/inflation", &inf); err == nil {
		snap.Inflation = parseFloat(inf.Inflation)
	}

	// annual provisions
	var ap annualProvisionsResp
	if err := doJSON(rest+"/cosmos/mint/v1beta1/annual-provisions", &ap); err == nil {
		snap.AnnualProvisions = ap.AnnualProvisions
	}

	// community pool
	var cp communityPoolResp
	if err := doJSON(rest+"/cosmos/distribution/v1beta1/community_pool", &cp); err == nil {
		snap.CommunityPool = formatCoins(cp.Pool, "")
	}

	// base fee
	var bf baseFeeResp
	if err := doJSON(rest+"/cosmos/evm/feemarket/v1/base_fee", &bf); err == nil {
		snap.BaseFee = bf.BaseFee
	}

	// block gas
	var bg blockGasResp
	if err := doJSON(rest+"/cosmos/evm/feemarket/v1/block_gas", &bg); err == nil {
		snap.BlockGas, _ = strconv.ParseUint(bg.Gas, 10, 64)
	}

	// governance proposals — try v1beta1 first, fall back to v1
	var votingProps, depositProps proposalsResp
	doJSON(fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals?proposal_status=2", rest), &votingProps)
	doJSON(fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals?proposal_status=1", rest), &depositProps)
	if len(votingProps.Proposals)+len(depositProps.Proposals) == 0 {
		doJSON(fmt.Sprintf("%s/cosmos/gov/v1/proposals?proposal_status=2", rest), &votingProps)
		doJSON(fmt.Sprintf("%s/cosmos/gov/v1/proposals?proposal_status=1", rest), &depositProps)
	}

	// parse voting proposals
	for _, p := range votingProps.Proposals {
		snap.VotingProposals = append(snap.VotingProposals, parseProposal(p))
	}
	// parse deposit proposals
	for _, p := range depositProps.Proposals {
		snap.DepositProposals = append(snap.DepositProposals, parseProposal(p))
	}

	// fetch tallies for voting-period proposals concurrently
	if len(snap.VotingProposals) > 0 {
		tallies := make([]ProposalTally, len(snap.VotingProposals))
		var twg sync.WaitGroup
		for i, vp := range snap.VotingProposals {
			twg.Add(1)
			go func(idx int, id uint64) {
				defer twg.Done()
				var tr proposalTallyResp
				tallyURL := fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals/%d/tally", rest, id)
				if err := doJSON(tallyURL, &tr); err != nil {
					tallyURL = fmt.Sprintf("%s/cosmos/gov/v1/proposals/%d/tally", rest, id)
					_ = doJSON(tallyURL, &tr)
				}
				if tr.Tally.Yes != "" || tr.Tally.No != "" || tr.Tally.Abstain != "" || tr.Tally.NoWithVeto != "" {
					tallies[idx] = ProposalTally{
						Yes:        tr.Tally.Yes,
						No:         tr.Tally.No,
						Abstain:    tr.Tally.Abstain,
						NoWithVeto: tr.Tally.NoWithVeto,
					}
				}
			}(i, vp.ID)
		}
		twg.Wait()
		for i := range snap.VotingProposals {
			snap.VotingProposals[i].Tally = tallies[i]
		}
	}

	// upgrade plan
	var upgradePlan upgradePlanResp
	if err := doJSON(rest+"/cosmos/upgrade/v1beta1/current_plan", &upgradePlan); err == nil && upgradePlan.Plan != nil {
		snap.UpgradeName = upgradePlan.Plan.Name
		snap.UpgradeHeight = parseInt64(upgradePlan.Plan.Height)
	}

	// token pairs
	var tp tokenPairsResp
	if err := doJSON(rest+"/cosmos/evm/erc20/v1/token_pairs", &tp); err == nil {
		for _, pair := range tp.TokenPairs {
			snap.TokenPairs = append(snap.TokenPairs, TokenPairInfo{
				Denom:     pair.Denom,
				ERC20Addr: pair.Erc20Address,
				Enabled:   pair.Enabled,
			})
		}
	}

	// IBC clients
	var ibcClients ibcClientsResp
	if err := doJSON(rest+"/ibc/core/client/v1/client_states", &ibcClients); err == nil {
		snap.IBCClientCount = len(ibcClients.ClientStates)
	}

	// per-validator rewards and commission (concurrent, max 10 workers)
	type valResult struct {
		valoper      string
		rewards      string
		rewardsAmt   string
		rewardsDenom string
		commEarned   string
	}
	valList := make([]ValidatorInfo, 0, len(allVals))
	for _, v := range allVals {
		valList = append(valList, v)
	}

	results := make([]valResult, len(valList))
	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup
	for i, v := range valList {
		wg.Add(1)
		go func(idx int, val ValidatorInfo) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			res := valResult{valoper: val.OperatorAddr}

			var rewards validatorRewardsResp
			if err := doJSON(fmt.Sprintf("%s/cosmos/distribution/v1beta1/validators/%s/outstanding_rewards", rest, val.OperatorAddr), &rewards); err == nil {
				res.rewards = formatCoins(rewards.Rewards.Rewards, "")
				if len(rewards.Rewards.Rewards) > 0 {
					res.rewardsAmt = rewards.Rewards.Rewards[0].Amount
					res.rewardsDenom = rewards.Rewards.Rewards[0].Denom
				}
			}

			var comm validatorCommissionResp
			if err := doJSON(fmt.Sprintf("%s/cosmos/distribution/v1beta1/validators/%s/commission", rest, val.OperatorAddr), &comm); err == nil {
				res.commEarned = formatCoins(comm.Commission.Commission, "")
			}

			results[idx] = res
		}(i, v)
	}
	wg.Wait()

	rewardMap := map[string]valResult{}
	for _, r := range results {
		rewardMap[r.valoper] = r
	}

	// compute total bonded for VP%
	totalBonded := parseFloat(snap.BondedTokens)

	for i, v := range valList {
		if r, ok := rewardMap[v.OperatorAddr]; ok {
			valList[i].OutstandingRewards = r.rewards
			valList[i].OutstandingRewardsAmt = r.rewardsAmt
			valList[i].OutstandingRewardsDenom = r.rewardsDenom
			valList[i].CommissionEarned = r.commEarned
		}
		// VP%
		tokens := parseFloat(v.VotingPowerTokens)
		if totalBonded > 0 {
			valList[i].VotingPowerPercent = tokens / totalBonded * 100
		}
		// slashing info
		if si, ok := sigInfoMap[strings.ToLower(v.ConsensusAddr)]; ok {
			valList[i].MissedBlocks = si.missed
			valList[i].Tombstoned = si.tombstoned
		}
		// proposer priority from RPC
		if pp, ok := rpcVals[strings.ToLower(v.ConsensusAddr)]; ok {
			valList[i].ProposerPriority = pp
		}
		applyValidatorP2P(&valList[i], snap.Moniker, snap.NodeID, snap.ListenAddr, peerByMoniker)
	}

	snap.Validators = valList

	// next proposer: validator with the highest proposer priority
	first := true
	var maxPriority int64
	for _, v := range valList {
		if first || v.ProposerPriority > maxPriority {
			maxPriority = v.ProposerPriority
			snap.NextProposerMoniker = v.Moniker
			first = false
		}
	}

	return snap
}

// FetchParams fetches chain params (called once on launch).
func FetchParams(rest string) ChainParams {
	p := ChainParams{}

	var sp stakingParamsResp
	if err := doJSON(rest+"/cosmos/staking/v1beta1/params", &sp); err == nil {
		p.UnbondingTime = parseDuration(sp.Params.UnbondingTime)
		p.MaxValidators = sp.Params.MaxValidators
		p.BondDenom = sp.Params.BondDenom
	}

	var slp slashingParamsResp
	if err := doJSON(rest+"/cosmos/slashing/v1beta1/params", &slp); err == nil {
		p.SignedBlocksWindow = parseInt64(slp.Params.SignedBlocksWindow)
		p.MinSignedPerWindow = parseFloat(slp.Params.MinSignedPerWindow)
		p.SlashFractionDowntime = slp.Params.SlashFractionDowntime
		p.SlashFractionDoubleSign = slp.Params.SlashFractionDoubleSign
	}

	var mp mintParamsResp
	if err := doJSON(rest+"/cosmos/mint/v1beta1/params", &mp); err == nil {
		p.BlocksPerYear = parseInt64(mp.Params.BlocksPerYear)
		p.GoalBonded = parseFloat(mp.Params.GoalBonded)
	}

	var gvp govVotingParamsResp
	if err := doJSON(rest+"/cosmos/gov/v1beta1/params/voting", &gvp); err == nil {
		p.VotingPeriod = parseDuration(gvp.VotingParams.VotingPeriod)
	}

	var gtp govTallyParamsResp
	if err := doJSON(rest+"/cosmos/gov/v1beta1/params/tallying", &gtp); err == nil {
		p.Quorum = parseFloat(gtp.TallyParams.Quorum)
		p.Threshold = parseFloat(gtp.TallyParams.Threshold)
		p.VetoThreshold = parseFloat(gtp.TallyParams.VetoThreshold)
	}

	var fmp feemarketParamsResp
	if err := doJSON(rest+"/cosmos/evm/feemarket/v1/params", &fmp); err == nil {
		p.MinGasPrice = parseFloat(fmp.Params.MinGasPrice)
		p.Elasticity = fmp.Params.ElasticityMultiplier
		p.NoBaseFee = fmp.Params.NoBaseFee
		p.BaseFeeChangeDenominator = fmp.Params.BaseFeeChangeDenominator
	}

	var ep evmParamsResp
	if err := doJSON(rest+"/cosmos/evm/vm/v1/params", &ep); err == nil {
		p.EVMDenom = ep.Params.EvmDenom
		p.ActiveStaticPrecompiles = ep.Params.ActiveStaticPrecompiles
		p.HistoryServeWindow = ep.Params.HistoryServeWindow
	}

	// EVM chain config for hardfork heights
	var evmConfigRaw struct {
		Config map[string]json.RawMessage `json:"config"`
	}
	if err := doJSON(rest+"/cosmos/evm/vm/v1/config", &evmConfigRaw); err == nil {
		p.HardforkLondon = rawToString(evmConfigRaw.Config["london_block"])
		p.HardforkShanghai = rawToString(evmConfigRaw.Config["shanghai_time"])
		p.HardforkCancun = rawToString(evmConfigRaw.Config["cancun_time"])
	}

	var erc20p erc20ParamsResp
	if err := doJSON(rest+"/cosmos/evm/erc20/v1/params", &erc20p); err == nil {
		p.ERC20Enabled = erc20p.Params.EnableErc20
	}

	var dp distributionParamsResp
	if err := doJSON(rest+"/cosmos/distribution/v1beta1/params", &dp); err == nil {
		p.CommunityTax = parseFloat(dp.Params.CommunityTax)
	}

	var pmtr pmtRewardsParamsResp
	if err := doJSON(rest+"/cosmos/evm/pmtrewards/v1/params", &pmtr); err == nil {
		p.RewardPerBlockAmount = pmtr.Params.RewardPerBlock.Amount
		p.RewardPerBlockDenom = pmtr.Params.RewardPerBlock.Denom
		p.PMTRewardsEnabled = pmtr.Params.Enabled
		p.PMTRewardsPoolAddress = pmtr.Params.PoolAddress
	}

	if p.PMTRewardsPoolAddress != "" {
		var poolBal struct {
			Balances []struct {
				Denom  string `json:"denom"`
				Amount string `json:"amount"`
			} `json:"balances"`
		}
		if err := doJSON(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", rest, p.PMTRewardsPoolAddress), &poolBal); err == nil {
			for _, b := range poolBal.Balances {
				if p.PMTRewardsPoolBalanceAmt == "" || b.Denom == p.BondDenom {
					p.PMTRewardsPoolBalanceAmt = b.Amount
					p.PMTRewardsPoolBalanceDenom = b.Denom
				}
			}
		}
	}

	return p
}

// rawToString converts a json.RawMessage to a trimmed string, stripping quotes.
// Handles both JSON strings ("123") and JSON numbers (123).
func rawToString(raw json.RawMessage) string {
	if raw == nil {
		return ""
	}
	s := strings.Trim(string(raw), `"`)
	if s == "null" || s == "" {
		return ""
	}
	return s
}
