package fetch

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ChainSnapshot holds all chain-level data.
type ChainSnapshot struct {
	NodeID        string
	Moniker       string
	AppVersion    string
	ListenAddr    string
	RpcListenAddr string
	Network       string

	BlockHeight         int64
	LatestBlockTime     time.Time
	BlockInterval       time.Duration
	CatchingUp          bool
	PeerCount           int
	PeerMonikers        []string
	MempoolTxs          int
	NextProposerMoniker string

	// This node's validator identity from /status (empty if full node).
	LocalConsensusAddr   string
	LocalConsensusBech32 string
	LocalAccountAddr        string
	LocalAccountLiquidAmt   string
	LocalAccountLiquidDenom string
	LocalDelegations        []DelegationInfo
	LocalP2PDial       string
	LocalVotingPower          int64

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
	// BlockGasLimit is consensus block.max_gas (-1 = unlimited, 0 = unknown).
	BlockGasLimit int64
	// MaxBlockBytes is consensus block.max_bytes.
	MaxBlockBytes int64

	// Parent block (H-1) from CometBFT block_results.
	ParentBlockGasUsed   uint64
	ParentBlockGasWanted uint64 // stored wanted (block_gas event or REST fallback)
	ParentBlockResultsOK bool
	ParentBaseFeeEvent   string // fee_market base_fee from begin_block at H (optional)

	VotingProposals  []ProposalInfo
	DepositProposals []ProposalInfo

	UpgradeName   string
	UpgradeHeight int64

	TokenPairs []TokenPairInfo

	IBCClientCount int

	ModuleBalances []ModuleBalanceInfo
	// LastBlockFeeRaw is parent block gas_used × base_fee (atto), when both are known.
	LastBlockFeeRaw string

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
	Moniker                 string
	OperatorAddr            string
	ConsensusAddr           string
	ConsensusBech32         string
	AccountAddr             string
	NodeID                  string // CometBFT peer ID (from /net_info or /status for this node)
	P2PDial                 string // node_id@listen_addr when known from chain RPC
	P2PConnected            bool   // peer visible in this node's /net_info (or is this node)
	Status                  string
	Jailed                  bool
	Tombstoned              bool
	VotingPowerTokens       string
	VotingPowerPercent      float64
	Commission              float64
	MissedBlocks            int64
	OutstandingRewards      string // formatted display string
	OutstandingRewardsAmt   string // raw base-denom amount for summing
	OutstandingRewardsDenom string
	CommissionEarned        string
	CommissionEarnedAmt     string
	CommissionEarnedDenom   string
	ProposerPriority        int64
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
	UnbondingTime            time.Duration
	MaxValidators            int
	BondDenom                string
	SignedBlocksWindow       int64
	MinSignedPerWindow       float64
	DowntimeJailDuration     time.Duration
	SlashFractionDowntime    string
	SlashFractionDoubleSign  string
	BlocksPerYear            int64
	GoalBonded               float64
	CommunityTax             float64
	WithdrawAddrEnabled      bool
	VotingPeriod             time.Duration
	Quorum                   float64
	Threshold                float64
	VetoThreshold            float64
	EVMDenom                 string
	EVMDenomName             string // bank metadata name (MetaMask network label)
	EVMDenomSymbol           string // bank metadata symbol (MetaMask currency symbol)
	EVMDenomDecimals         uint32 // display-denom exponent (MetaMask decimals)
	MinGasPrice              float64
	Elasticity               int64
	NoBaseFee                bool
	BaseFeeChangeDenominator int64
	MinGasMultiplier         float64
	MinGasPriceRaw           string
	EnableHeight             int64
	BaseFeeParam             string
	ERC20Enabled             bool
	ActiveStaticPrecompiles  []string
	HistoryServeWindow       int64
	HardforkLondon           string
	HardforkShanghai         string
	HardforkCancun           string
	// pmtrewards module
	RewardPerBlockAmount       string
	RewardPerBlockDenom        string
	PMTRewardsEnabled          bool
	PMTRewardsPoolAddress      string
	PMTRewardsPoolBalanceAmt   string // raw base-denom amount
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

type consensusParamsResp struct {
	Result struct {
		ConsensusParams struct {
			Block struct {
				MaxGas   string `json:"max_gas"`
				MaxBytes string `json:"max_bytes"`
			} `json:"block"`
		} `json:"consensus_params"`
	} `json:"result"`
}

// --- REST response types ---

type stakingValidatorsResp struct {
	Validators []struct {
		OperatorAddress string `json:"operator_address"`
		Description     struct {
			Moniker string `json:"moniker"`
		} `json:"description"`
		Status     string `json:"status"`
		Tokens     string `json:"tokens"`
		Commission struct {
			CommissionRates struct {
				Rate string `json:"rate"`
			} `json:"commission_rates"`
		} `json:"commission"`
		Jailed          bool `json:"jailed"`
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
		Address             string `json:"address"`
		MissedBlocksCounter string `json:"missed_blocks_counter"`
		Tombstoned          bool   `json:"tombstoned"`
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
		Denom        string `json:"denom"`
		Erc20Address string `json:"erc20_address"`
		Enabled      bool   `json:"enabled"`
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
		DowntimeJailDuration    string `json:"downtime_jail_duration"`
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
		BaseFee                  string `json:"base_fee"`
		BaseFeeChangeDenominator int64  `json:"base_fee_change_denominator"`
		EnableHeight             string `json:"enable_height"`
		MinGasPrice              string `json:"min_gas_price"`
		MinGasMultiplier         string `json:"min_gas_multiplier"`
		ElasticityMultiplier     int64  `json:"elasticity_multiplier"`
	} `json:"params"`
}

type evmParamsResp struct {
	Params struct {
		EvmDenom                string   `json:"evm_denom"`
		ActiveStaticPrecompiles []string `json:"active_static_precompiles"`
		HistoryServeWindow      string   `json:"history_serve_window"`
	} `json:"params"`
}

type bankDenomUnit struct {
	Denom    string `json:"denom"`
	Exponent uint32 `json:"exponent"`
}

type bankDenomMetadataResp struct {
	Metadata struct {
		Name       string          `json:"name"`
		Symbol     string          `json:"symbol"`
		Display    string          `json:"display"`
		Base       string          `json:"base"`
		DenomUnits []bankDenomUnit `json:"denom_units"`
	} `json:"metadata"`
}

type erc20ParamsResp struct {
	Params struct {
		EnableErc20 bool `json:"enable_erc20"`
	} `json:"params"`
}

type distributionParamsResp struct {
	Params struct {
		CommunityTax        string `json:"community_tax"`
		WithdrawAddrEnabled bool   `json:"withdraw_addr_enabled"`
	} `json:"params"`
}

type pmtRewardsParamsResp struct {
	Params struct {
		Enabled        bool `json:"enabled"`
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

// FormatFeeAmount formats base fee or gas price for compact display.
// Handles integer atto/wei strings, long decimal zeros from REST, and pre-formatted values.
func FormatFeeAmount(raw, denom string) string {
	if raw == "" || raw == "0" {
		return "0"
	}
	if strings.Contains(raw, " ") {
		return raw
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return raw
	}
	if v == 0 {
		return "0"
	}
	if !strings.Contains(raw, ".") {
		displayVal, _ := coinToDisplay(raw, denom)
		if denom != "" && displayVal > 0 && displayVal < 1e-6 {
			return raw + " " + denom
		}
		return FormatCoin(raw, denom)
	}
	if v >= 1 {
		return FormatCoin(fmt.Sprintf("%.0f", v), denom)
	}
	if denom != "" {
		return FormatAmountUnit(v, displayDenom(denom))
	}
	return FormatAmount(v)
}

// FormatCoin converts a raw Cosmos amount string + on-chain denom to a
// human-readable display string, e.g. "400000000000000000000000000" + "apmt"
// → "400.00M PMT".
func FormatCoin(rawAmount, denom string) string {
	displayVal, displayDenom := coinToDisplay(rawAmount, denom)
	if displayDenom == "" {
		return FormatAmount(displayVal)
	}
	return FormatAmountUnit(displayVal, displayDenom)
}

// NormalizeCoin returns the display float value and display denom for a raw
// Cosmos coin, useful when the caller needs the numeric value for calculations.
func NormalizeCoin(rawAmount, denom string) (float64, string) {
	return coinToDisplay(rawAmount, denom)
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

// FetchEVMWalletParams loads VM denom settings and bank metadata for the EVM panel.
func FetchEVMWalletParams(rest string) ChainParams {
	p := ChainParams{}
	var ep evmParamsResp
	if err := doJSON(rest+"/cosmos/evm/vm/v1/params", &ep); err == nil {
		p.EVMDenom = ep.Params.EvmDenom
		fetchEVMDenomMetadata(rest, &p)
	}
	return p
}

func fetchEVMDenomMetadata(rest string, p *ChainParams) {
	if p.EVMDenom == "" {
		return
	}
	var meta bankDenomMetadataResp
	if err := doJSON(rest+"/cosmos/bank/v1beta1/denoms_metadata/"+p.EVMDenom, &meta); err == nil {
		p.EVMDenomName = meta.Metadata.Name
		p.EVMDenomSymbol = meta.Metadata.Symbol
		p.EVMDenomDecimals = evmDisplayDecimals(meta.Metadata.Display, meta.Metadata.DenomUnits)
	}
}

func evmDisplayDecimals(display string, units []bankDenomUnit) uint32 {
	if display != "" {
		for _, u := range units {
			if u.Denom == display {
				return u.Exponent
			}
		}
	}
	var max uint32
	for _, u := range units {
		if u.Exponent > max {
			max = u.Exponent
		}
	}
	return max
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
		p.DowntimeJailDuration = parseDuration(slp.Params.DowntimeJailDuration)
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
		p.MinGasPriceRaw = fmp.Params.MinGasPrice
		p.MinGasPrice = parseFloat(fmp.Params.MinGasPrice)
		p.MinGasMultiplier = parseFloat(fmp.Params.MinGasMultiplier)
		p.Elasticity = fmp.Params.ElasticityMultiplier
		p.NoBaseFee = fmp.Params.NoBaseFee
		p.BaseFeeChangeDenominator = fmp.Params.BaseFeeChangeDenominator
		p.BaseFeeParam = fmp.Params.BaseFee
		p.EnableHeight = parseInt64(fmp.Params.EnableHeight)
	}

	var ep evmParamsResp
	if err := doJSON(rest+"/cosmos/evm/vm/v1/params", &ep); err == nil {
		p.EVMDenom = ep.Params.EvmDenom
		p.ActiveStaticPrecompiles = ep.Params.ActiveStaticPrecompiles
		p.HistoryServeWindow = parseInt64(ep.Params.HistoryServeWindow)
		fetchEVMDenomMetadata(rest, &p)
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
		p.WithdrawAddrEnabled = dp.Params.WithdrawAddrEnabled
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
