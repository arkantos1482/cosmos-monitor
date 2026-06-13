package fetch

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ValidatorScope selects which staking validators to query.
type ValidatorScope int

const (
	ValidatorsNone ValidatorScope = iota
	ValidatorsBonded
	ValidatorsAll
)

// ChainRecipe declares which chain data a dashboard section needs.
type ChainRecipe struct {
	CometExtended    bool // net_info, mempool, block interval, RPC validator priorities
	ConsensusParams  bool
	ValidatorScope   ValidatorScope
	StakingPool      bool
	SigningInfos     bool
	Supply           bool
	MintData         bool // inflation + annual provisions
	CommunityPool    bool
	FeemarketLive    bool // base_fee, block_gas, block_results
	ModuleBalances   bool
	Governance       bool
	ValidatorRewards bool
	LocalStaking     bool // local delegations (balance enrichment is separate)
}

// ChainRecipeFull fetches everything needed for the overview page.
var ChainRecipeFull = ChainRecipe{
	CometExtended: true, ConsensusParams: true,
	ValidatorScope: ValidatorsAll, StakingPool: true, SigningInfos: true,
	Supply: true, MintData: true, CommunityPool: true, FeemarketLive: true,
	ModuleBalances: true, Governance: true, ValidatorRewards: true, LocalStaking: true,
}

// ChainRecipeNode is the Validator (/s/node) section.
var ChainRecipeNode = ChainRecipe{
	CometExtended: true, ValidatorScope: ValidatorsBonded,
	StakingPool: true, SigningInfos: true,
}

// ChainRecipeStaking is the Staking section.
var ChainRecipeStaking = ChainRecipe{
	CometExtended: true, ValidatorScope: ValidatorsAll,
	StakingPool: true, SigningInfos: true, Supply: true,
	ModuleBalances: true, LocalStaking: true,
}

// ChainRecipeSlashing is the Slashing section.
var ChainRecipeSlashing = ChainRecipe{
	ValidatorScope: ValidatorsAll, SigningInfos: true,
}

// ChainRecipeRewards is the Rewards section.
var ChainRecipeRewards = ChainRecipe{
	CometExtended: true, ValidatorScope: ValidatorsBonded,
	StakingPool: true, MintData: true, FeemarketLive: true,
}

// ChainRecipeDistribution is the Distribution section.
var ChainRecipeDistribution = ChainRecipe{
	ValidatorScope: ValidatorsAll, CommunityPool: true,
	ModuleBalances: true, ValidatorRewards: true,
}

// ChainRecipeFeemarket is the Fee market section.
var ChainRecipeFeemarket = ChainRecipe{
	ConsensusParams: true, FeemarketLive: true,
}

// ChainRecipeGovernance is the Governance section.
var ChainRecipeGovernance = ChainRecipe{
	Governance: true, ModuleBalances: true,
}

// FetchChainRecipe loads chain snapshot data according to recipe.
// CometBFT /status is always fetched (block height, sync, local validator key).
func FetchChainRecipe(rpc, rest string, recipe ChainRecipe) ChainSnapshot {
	snap := ChainSnapshot{}

	var status statusResp
	if err := doJSON(rpc+"/status", &status); err != nil {
		snap.Err = fmt.Errorf("rpc status: %w", err)
		return snap
	}

	snap.NodeID = status.Result.NodeInfo.ID
	snap.Moniker = status.Result.NodeInfo.Moniker
	snap.AppVersion = status.Result.NodeInfo.Version
	snap.ListenAddr = status.Result.NodeInfo.ListenAddr
	snap.RpcListenAddr = status.Result.NodeInfo.Other.RPCAddress
	snap.Network = status.Result.NodeInfo.Network
	snap.BlockHeight = parseInt64(status.Result.SyncInfo.LatestBlockHeight)
	snap.CatchingUp = status.Result.SyncInfo.CatchingUp
	if t, err := time.Parse(time.RFC3339Nano, status.Result.SyncInfo.LatestBlockTime); err == nil {
		snap.LatestBlockTime = t
	}
	snap.LocalConsensusAddr = strings.ToLower(status.Result.ValidatorInfo.Address)
	snap.LocalVotingPower = parseInt64(status.Result.ValidatorInfo.VotingPower)
	snap.LocalP2PDial = formatP2PDial(snap.NodeID, snap.ListenAddr)

	peerByMoniker := map[string]struct {
		nodeID string
		listen string
	}{}
	rpcVals := map[string]int64{}

	if recipe.CometExtended {
		if recipe.ConsensusParams {
			var consensusParams consensusParamsResp
			if err := doJSON(rpc+"/consensus_params", &consensusParams); err == nil {
				snap.BlockGasLimit = parseInt64(consensusParams.Result.ConsensusParams.Block.MaxGas)
				snap.MaxBlockBytes = parseInt64(consensusParams.Result.ConsensusParams.Block.MaxBytes)
			}
		}

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

		fetchBlockInterval(rpc, &snap)

		var valSet validatorsResp
		if err := doJSON(rpc+"/validators?per_page=100", &valSet); err == nil {
			for _, v := range valSet.Result.Validators {
				rpcVals[strings.ToLower(v.Address)] = parseInt64(v.ProposerPriority)
			}
		}
	} else if recipe.ConsensusParams {
		var consensusParams consensusParamsResp
		if err := doJSON(rpc+"/consensus_params", &consensusParams); err == nil {
			snap.BlockGasLimit = parseInt64(consensusParams.Result.ConsensusParams.Block.MaxGas)
			snap.MaxBlockBytes = parseInt64(consensusParams.Result.ConsensusParams.Block.MaxBytes)
		}
	}

	allVals := fetchStakingValidators(rest, recipe.ValidatorScope)

	if recipe.StakingPool {
		var pool stakingPoolResp
		if err := doJSON(rest+"/cosmos/staking/v1beta1/pool", &pool); err == nil {
			snap.BondedTokens = pool.Pool.BondedTokens
			snap.NotBondedTokens = pool.Pool.NotBondedTokens
		}
	}

	sigInfoMap, consBech32Map := map[string]struct {
		missed     int64
		tombstoned bool
	}{}, map[string]string{}
	if recipe.SigningInfos {
		sigInfoMap, consBech32Map = fetchSigningInfos(rest)
	}

	if recipe.Supply {
		var supply supplyResp
		if err := doJSON(rest+"/cosmos/bank/v1beta1/supply", &supply); err == nil && len(supply.Supply) > 0 {
			snap.TotalSupply = supply.Supply[0].Amount
			snap.TotalSupplyDenom = supply.Supply[0].Denom
		}
	}

	if recipe.MintData {
		var inf inflationResp
		if err := doJSON(rest+"/cosmos/mint/v1beta1/inflation", &inf); err == nil {
			snap.Inflation = parseFloat(inf.Inflation)
		}
		var ap annualProvisionsResp
		if err := doJSON(rest+"/cosmos/mint/v1beta1/annual-provisions", &ap); err == nil {
			snap.AnnualProvisions = ap.AnnualProvisions
		}
	}

	if recipe.CommunityPool {
		var cp communityPoolResp
		if err := doJSON(rest+"/cosmos/distribution/v1beta1/community_pool", &cp); err == nil {
			snap.CommunityPool = formatCoins(cp.Pool, "")
		}
	}

	if recipe.FeemarketLive {
		fetchFeemarketLive(rpc, rest, &snap)
	}

	if recipe.ModuleBalances {
		preferDenom := snap.Params.BondDenom
		if preferDenom == "" {
			preferDenom = snap.TotalSupplyDenom
		}
		snap.ModuleBalances = FetchModuleBalances(rest, preferDenom)
	}

	if recipe.Governance {
		fetchGovernanceProposals(rest, &snap)
		fetchGovernanceExtras(rest, &snap)
	}

	valList := validatorsFromMap(allVals)

	if recipe.ValidatorRewards {
		attachValidatorRewards(rest, valList)
	}

	totalBonded := parseFloat(snap.BondedTokens)
	enrichValidatorList(valList, totalBonded, sigInfoMap, consBech32Map, rpcVals, peerByMoniker,
		snap.Moniker, snap.NodeID, snap.ListenAddr, recipe.SigningInfos, recipe.CometExtended)

	snap.Validators = valList

	if snap.LocalConsensusAddr != "" {
		hexLocal := strings.ToLower(snap.LocalConsensusAddr)
		if b, ok := consBech32Map[hexLocal]; ok {
			snap.LocalConsensusBech32 = b
		} else {
			snap.LocalConsensusBech32 = hexToBech32(Bech32PrefixCons, hexLocal)
		}
	}

	if recipe.LocalStaking {
		resolveLocalStaking(rest, &snap)
	}

	if recipe.CometExtended {
		snap.NextProposerMoniker = nextProposerMoniker(valList)
	}

	return snap
}

func fetchBlockInterval(rpc string, snap *ChainSnapshot) {
	var latestBlock, prevBlock blockResp
	if err := doJSON(rpc+"/block", &latestBlock); err != nil {
		return
	}
	prevHeight := parseInt64(latestBlock.Result.Block.Header.Height) - 1
	if prevHeight <= 0 {
		return
	}
	if err := doJSON(fmt.Sprintf("%s/block?height=%d", rpc, prevHeight), &prevBlock); err != nil {
		return
	}
	t1, e1 := time.Parse(time.RFC3339Nano, latestBlock.Result.Block.Header.Time)
	t2, e2 := time.Parse(time.RFC3339Nano, prevBlock.Result.Block.Header.Time)
	if e1 == nil && e2 == nil {
		snap.BlockInterval = t1.Sub(t2)
	}
}

func fetchStakingValidators(rest string, scope ValidatorScope) map[string]ValidatorInfo {
	allVals := map[string]ValidatorInfo{}
	if scope == ValidatorsNone || rest == "" {
		return allVals
	}
	statuses := []string{"BOND_STATUS_BONDED", "BOND_STATUS_UNBONDING", "BOND_STATUS_UNBONDED"}
	if scope == ValidatorsBonded {
		statuses = []string{"BOND_STATUS_BONDED"}
	}
	for _, status := range statuses {
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
	return allVals
}

func fetchSigningInfos(rest string) (map[string]struct {
	missed     int64
	tombstoned bool
}, map[string]string) {
	sigInfoMap := map[string]struct {
		missed     int64
		tombstoned bool
	}{}
	consBech32Map := map[string]string{}
	var sigInfos signingInfosResp
	if err := doJSON(rest+"/cosmos/slashing/v1beta1/signing_infos?pagination.limit=100", &sigInfos); err != nil {
		return sigInfoMap, consBech32Map
	}
	for _, si := range sigInfos.Info {
		hexAddr := bech32ToHex(si.Address)
		if hexAddr == "" {
			hexAddr = strings.ToLower(si.Address)
		}
		if hexAddr != "" && si.Address != "" {
			consBech32Map[hexAddr] = si.Address
		}
		sigInfoMap[hexAddr] = struct {
			missed     int64
			tombstoned bool
		}{parseInt64(si.MissedBlocksCounter), si.Tombstoned}
	}
	return sigInfoMap, consBech32Map
}

func fetchFeemarketLive(rpc, rest string, snap *ChainSnapshot) {
	var bf baseFeeResp
	if err := doJSON(rest+"/cosmos/evm/feemarket/v1/base_fee", &bf); err == nil {
		snap.BaseFee = bf.BaseFee
	}
	var bg blockGasResp
	if err := doJSON(rest+"/cosmos/evm/feemarket/v1/block_gas", &bg); err == nil {
		snap.BlockGas, _ = strconv.ParseUint(bg.Gas, 10, 64)
	}
	if snap.BlockHeight > 0 {
		parent := FetchBlockResults(rpc, snap.BlockHeight-1)
		if parent.OK {
			snap.ParentBlockResultsOK = true
			snap.ParentBlockGasUsed = parent.GasUsedSum
			snap.ParentBlockGasWanted = parent.BlockGasWanted
		}
		cur := FetchBlockResults(rpc, snap.BlockHeight)
		if cur.OK && cur.BaseFeeEvent != "" {
			snap.ParentBaseFeeEvent = cur.BaseFeeEvent
		}
	}
	if snap.ParentBlockGasWanted == 0 && snap.BlockGas > 0 {
		snap.ParentBlockGasWanted = snap.BlockGas
	}
	if snap.ParentBlockGasUsed > 0 && snap.BaseFee != "" {
		baseFeeF := parseFloat(snap.BaseFee)
		if baseFeeF > 0 {
			fee := baseFeeF * float64(snap.ParentBlockGasUsed)
			snap.LastBlockFeeRaw = fmt.Sprintf("%.0f", fee)
		}
	}
}

func fetchGovernanceProposals(rest string, snap *ChainSnapshot) {
	var votingProps, depositProps proposalsResp
	doJSON(fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals?proposal_status=2", rest), &votingProps)
	doJSON(fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals?proposal_status=1", rest), &depositProps)
	if len(votingProps.Proposals)+len(depositProps.Proposals) == 0 {
		doJSON(fmt.Sprintf("%s/cosmos/gov/v1/proposals?proposal_status=2", rest), &votingProps)
		doJSON(fmt.Sprintf("%s/cosmos/gov/v1/proposals?proposal_status=1", rest), &depositProps)
	}
	for _, p := range votingProps.Proposals {
		snap.VotingProposals = append(snap.VotingProposals, parseProposal(p))
	}
	for _, p := range depositProps.Proposals {
		snap.DepositProposals = append(snap.DepositProposals, parseProposal(p))
	}
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
						Yes: tr.Tally.Yes, No: tr.Tally.No,
						Abstain: tr.Tally.Abstain, NoWithVeto: tr.Tally.NoWithVeto,
					}
				}
			}(i, vp.ID)
		}
		twg.Wait()
		for i := range snap.VotingProposals {
			snap.VotingProposals[i].Tally = tallies[i]
		}
	}
}

func fetchGovernanceExtras(rest string, snap *ChainSnapshot) {
	var upgradePlan upgradePlanResp
	if err := doJSON(rest+"/cosmos/upgrade/v1beta1/current_plan", &upgradePlan); err == nil && upgradePlan.Plan != nil {
		snap.UpgradeName = upgradePlan.Plan.Name
		snap.UpgradeHeight = parseInt64(upgradePlan.Plan.Height)
	}
	var tp tokenPairsResp
	if err := doJSON(rest+"/cosmos/evm/erc20/v1/token_pairs", &tp); err == nil {
		for _, pair := range tp.TokenPairs {
			snap.TokenPairs = append(snap.TokenPairs, TokenPairInfo{
				Denom: pair.Denom, ERC20Addr: pair.Erc20Address, Enabled: pair.Enabled,
			})
		}
	}
	var ibcClients ibcClientsResp
	if err := doJSON(rest+"/ibc/core/client/v1/client_states", &ibcClients); err == nil {
		snap.IBCClientCount = len(ibcClients.ClientStates)
	}
}

func validatorsFromMap(allVals map[string]ValidatorInfo) []ValidatorInfo {
	valList := make([]ValidatorInfo, 0, len(allVals))
	for _, v := range allVals {
		valList = append(valList, v)
	}
	return valList
}

func attachValidatorRewards(rest string, valList []ValidatorInfo) {
	type valResult struct {
		valoper       string
		rewards       string
		rewardsAmt    string
		rewardsDenom  string
		commEarned    string
		commEarnedAmt string
		commEarnedDen string
	}
	results := make([]valResult, len(valList))
	sem := make(chan struct{}, 10)
	var rwg sync.WaitGroup
	for i, v := range valList {
		rwg.Add(1)
		go func(idx int, val ValidatorInfo) {
			defer rwg.Done()
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
				if len(comm.Commission.Commission) > 0 {
					res.commEarnedAmt = comm.Commission.Commission[0].Amount
					res.commEarnedDen = comm.Commission.Commission[0].Denom
				}
			}
			results[idx] = res
		}(i, v)
	}
	rwg.Wait()
	rewardMap := map[string]valResult{}
	for _, r := range results {
		rewardMap[r.valoper] = r
	}
	for i, v := range valList {
		if r, ok := rewardMap[v.OperatorAddr]; ok {
			valList[i].OutstandingRewards = r.rewards
			valList[i].OutstandingRewardsAmt = r.rewardsAmt
			valList[i].OutstandingRewardsDenom = r.rewardsDenom
			valList[i].CommissionEarned = r.commEarned
			valList[i].CommissionEarnedAmt = r.commEarnedAmt
			valList[i].CommissionEarnedDenom = r.commEarnedDen
		}
	}
}

func enrichValidatorList(valList []ValidatorInfo, totalBonded float64,
	sigInfoMap map[string]struct {
		missed     int64
		tombstoned bool
	},
	consBech32Map map[string]string, rpcVals map[string]int64,
	peerByMoniker map[string]struct {
		nodeID string
		listen string
	},
	localMoniker, localNodeID, localListen string,
	applySigning, applyP2P bool,
) {
	for i, v := range valList {
		tokens := parseFloat(v.VotingPowerTokens)
		if totalBonded > 0 {
			valList[i].VotingPowerPercent = tokens / totalBonded * 100
		}
		hexCons := strings.ToLower(v.ConsensusAddr)
		if b, ok := consBech32Map[hexCons]; ok {
			valList[i].ConsensusBech32 = b
		} else if hexCons != "" {
			valList[i].ConsensusBech32 = hexToBech32(Bech32PrefixCons, hexCons)
		}
		if applySigning {
			if si, ok := sigInfoMap[hexCons]; ok {
				valList[i].MissedBlocks = si.missed
				valList[i].Tombstoned = si.tombstoned
			}
		}
		if pp, ok := rpcVals[hexCons]; ok {
			valList[i].ProposerPriority = pp
		}
		if applyP2P {
			applyValidatorP2P(&valList[i], localMoniker, localNodeID, localListen, peerByMoniker)
		}
	}
}

func resolveLocalStaking(rest string, snap *ChainSnapshot) {
	localOperator := findLocalOperator(*snap)
	if localOperator == "" {
		return
	}
	snap.LocalDelegations = FetchValidatorDelegations(rest, localOperator)
	snap.LocalAccountAddr = ValOperToAcc(localOperator)
}

func findLocalOperator(snap ChainSnapshot) string {
	if snap.LocalConsensusAddr != "" {
		for _, v := range snap.Validators {
			if strings.EqualFold(v.ConsensusAddr, snap.LocalConsensusAddr) {
				return v.OperatorAddr
			}
		}
	}
	if snap.Moniker != "" {
		for _, v := range snap.Validators {
			if v.Moniker == snap.Moniker {
				return v.OperatorAddr
			}
		}
	}
	return ""
}

func nextProposerMoniker(valList []ValidatorInfo) string {
	first := true
	var maxPriority int64
	var moniker string
	for _, v := range valList {
		if first || v.ProposerPriority > maxPriority {
			maxPriority = v.ProposerPriority
			moniker = v.Moniker
			first = false
		}
	}
	return moniker
}
