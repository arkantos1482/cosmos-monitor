package report

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func Build(chain fetch.ChainSnapshot, ev fetch.EVMSnapshot, sys fetch.SystemSnapshot, docker fetch.DockerSnapshot, evmHTTPEndpoint string, status model.StatusAvailability) model.Report {
	p := chain.Params
	d := model.Report{}

	d.Moniker = chain.Moniker
	d.HasChainStatus = status.ChainOK
	d.HasEVMPeers = status.EVMOK
	d.HasNodeStatus = status.DockerOK
	d.Synced = !chain.CatchingUp
	d.BlockHeight = FormatInt(chain.BlockHeight)
	d.TimeUTC = time.Now().UTC().Format("15:04:05") + " UTC"
	d.PeerCount = chain.PeerCount
	d.EVMPeerCount = ev.PeerCount

	d.NodeID = chain.NodeID
	d.AppVersion = chain.AppVersion
	d.ListenAddr = chain.ListenAddr
	d.RpcListenAddr = chain.RpcListenAddr
	d.Network = chain.Network
	d.LocalConsensusAddr = chain.LocalConsensusAddr
	if chain.LocalVotingPower > 0 {
		d.LocalVotingPower = FormatInt(chain.LocalVotingPower)
	}
	d.PeerMonikers = chain.PeerMonikers
	d.MempoolTxs = chain.MempoolTxs
	d.NextProposer = chain.NextProposerMoniker
	if chain.BlockInterval > 0 {
		d.BlockInterval = FormatDur(chain.BlockInterval)
	}
	if !chain.LatestBlockTime.IsZero() {
		d.TimeSinceBlock = FormatDur(time.Since(chain.LatestBlockTime))
		d.LatestBlockTime = chain.LatestBlockTime.UTC().Format("2006-01-02 15:04:05 UTC")
	}

	d.Load1, d.Load5, d.Load15 = sys.LoadAvg1, sys.LoadAvg5, sys.LoadAvg15
	memUsed := uint64(0)
	if sys.MemTotal > sys.MemAvail {
		memUsed = sys.MemTotal - sys.MemAvail
	}
	d.MemUsed = FormatBytes(memUsed)
	d.MemTotal = FormatBytes(sys.MemTotal)
	d.MemPct = int(float64(memUsed) / float64(max(sys.MemTotal, uint64(1))) * 100)
	d.DiskUsed = FormatBytes(sys.DiskUsed)
	d.DiskTotal = FormatBytes(sys.DiskTotal)
	d.DiskPct = int(float64(sys.DiskUsed) / float64(max(sys.DiskTotal, uint64(1))) * 100)

	d.NodeRunning = docker.Running
	d.NodeCPU = fmt.Sprintf("%.1f%%", docker.CPUPercent)
	d.NodeMemUsed = FormatBytes(docker.MemUsage)
	d.NodeMemTotal = FormatBytes(docker.MemLimit)
	d.Restarts = docker.RestartCount
	if !docker.StartedAt.IsZero() {
		d.NodeUptime = FormatDurFull(time.Since(docker.StartedAt))
	}

	maxMissed := int64(0)
	if p.SignedBlocksWindow > 0 {
		maxMissed = int64(float64(p.SignedBlocksWindow) * (1 - p.MinSignedPerWindow))
	}

	vals := make([]fetch.ValidatorInfo, len(chain.Validators))
	copy(vals, chain.Validators)
	sort.Slice(vals, func(i, j int) bool {
		return strings.ToLower(vals[i].Moniker) < strings.ToLower(vals[j].Moniker)
	})

	localAddr := strings.ToLower(chain.LocalConsensusAddr)
	var localVal *fetch.ValidatorInfo
	tombCount, belowThresh, jailedCount, bondedCount := 0, 0, 0, 0

	for i := range chain.Validators {
		v := &chain.Validators[i]
		if v.Tombstoned {
			tombCount++
		}
		if v.Jailed {
			jailedCount++
		}
		if v.Status == "BONDED" {
			bondedCount++
		}
		if v.Status == "BONDED" && !v.Tombstoned && maxMissed > 0 && v.MissedBlocks > maxMissed {
			belowThresh++
		}
		if localAddr != "" && strings.EqualFold(v.ConsensusAddr, localAddr) {
			localVal = v
		}
	}
	if localVal == nil && chain.Moniker != "" {
		for i := range chain.Validators {
			if chain.Validators[i].Moniker == chain.Moniker {
				localVal = &chain.Validators[i]
				break
			}
		}
	}

	d.TombstonedCount = tombCount
	d.BelowThreshold = belowThresh
	d.JailedCount = jailedCount
	d.BondedCount = bondedCount

	for _, v := range vals {
		isLocal := localVal != nil && v.OperatorAddr == localVal.OperatorAddr
		p2p := v.P2PDial
		if p2p == "" {
			p2p = "—"
		}
		d.Validators = append(d.Validators, model.Validator{
			Moniker:         v.Moniker,
			Operator:        v.OperatorAddr,
			NodeID:          v.NodeID,
			ConsensusAddr:   strings.ToUpper(v.ConsensusAddr),
			ConsensusBech32: v.ConsensusBech32,
			P2PDial:         p2p,
			P2PConnected:    v.P2PConnected,
			VPFloat:         v.VotingPowerPercent,
			CommissionFloat: v.Commission * 100,
			Missed:          v.MissedBlocks,
			MissedHigh:      maxMissed > 0 && v.MissedBlocks > maxMissed,
			Status:          strings.ToLower(v.Status),
			Jailed:          v.Jailed,
			Tombstoned:      v.Tombstoned,
			IsLocal:         isLocal,
		})
	}

	d.Local = buildLocalValidator(chain, localVal, maxMissed)

	denom := p.BondDenom
	if denom == "" {
		denom = chain.TotalSupplyDenom
	}
	d.BondDenom = denom
	bondedF, _ := fetch.NormalizeCoin(chain.BondedTokens, denom)
	totalF, _ := fetch.NormalizeCoin(chain.TotalSupply, chain.TotalSupplyDenom)
	if totalF > 0 {
		d.BondedPct = bondedF / totalF * 100
	}
	d.TotalSupply = fetch.FormatCoin(chain.TotalSupply, chain.TotalSupplyDenom)
	d.BondedAmt = fetch.FormatCoin(chain.BondedTokens, denom)
	d.GoalBonded = p.GoalBonded * 100
	d.NotBonded = fetch.FormatCoin(chain.NotBondedTokens, denom)
	d.UnbondingTime = FormatDurFull(p.UnbondingTime)
	if p.MaxValidators > 0 {
		d.MaxValidators = int64(p.MaxValidators)
	}
	d.Inflation = chain.Inflation * 100
	if chain.AnnualProvisions != "" {
		d.AnnualProvisions = fetch.FormatCoin(chain.AnnualProvisions, denom)
	}
	d.CommunityPool = chain.CommunityPool
	if p.BlocksPerYear > 0 {
		d.BlocksPerYear = FormatInt(p.BlocksPerYear)
	}

	d.CommunityTaxPct = p.CommunityTax * 100
	d.CommunityTaxZero = p.CommunityTax == 0
	if p.CommunityTax == 0 {
		d.CommunityTax = "0%"
	} else {
		d.CommunityTax = fmt.Sprintf("%.2f%%", d.CommunityTaxPct)
	}

	d.PMTEnabled = p.PMTRewardsEnabled
	d.PMTPoolEmpty = p.PMTRewardsPoolBalanceAmt == "" || p.PMTRewardsPoolBalanceAmt == "0"
	if p.RewardPerBlockAmount != "" {
		d.PMTRate = fetch.FormatCoin(p.RewardPerBlockAmount, p.RewardPerBlockDenom) + "/block"
		if p.BlocksPerYear > 0 {
			rewardF, _ := fetch.NormalizeCoin(p.RewardPerBlockAmount, p.RewardPerBlockDenom)
			_, dispDenom := fetch.NormalizeCoin("0", p.RewardPerBlockDenom)
			d.PMTAnnual = "~" + fetch.FormatAmountUnit(rewardF*float64(p.BlocksPerYear), dispDenom) + "/year"
		}
		if chain.BlockInterval > 0 {
			rewardF, _ := fetch.NormalizeCoin(p.RewardPerBlockAmount, p.RewardPerBlockDenom)
			_, dispDenom := fetch.NormalizeCoin("0", p.RewardPerBlockDenom)
			blocksPerDay := 86400.0 / chain.BlockInterval.Seconds()
			d.PMTDailyEmit = "~" + fetch.FormatAmountUnit(rewardF*blocksPerDay, dispDenom) + "/day"
		}
	}
	if !d.PMTPoolEmpty {
		d.PMTBalance = fetch.FormatCoin(p.PMTRewardsPoolBalanceAmt, p.PMTRewardsPoolBalanceDenom)
		poolF, _ := fetch.NormalizeCoin(p.PMTRewardsPoolBalanceAmt, p.PMTRewardsPoolBalanceDenom)
		if chain.BlockInterval > 0 && p.RewardPerBlockAmount != "" {
			rewardF, _ := fetch.NormalizeCoin(p.RewardPerBlockAmount, p.RewardPerBlockDenom)
			blocksPerDay := 86400.0 / chain.BlockInterval.Seconds()
			if daily := rewardF * blocksPerDay; daily > 0 {
				d.PMTRunway = fmt.Sprintf("~%.0f days left", poolF/daily)
			}
		}
	}
	d.PMTPoolAddress = p.PMTRewardsPoolAddress

	var totalOutF, totalCommF float64
	var outDenom string
	for _, v := range chain.Validators {
		if v.OutstandingRewardsAmt != "" {
			f, dd := fetch.NormalizeCoin(v.OutstandingRewardsAmt, v.OutstandingRewardsDenom)
			totalOutF += f
			outDenom = dd
		}
		if v.CommissionEarnedAmt != "" {
			f, dd := fetch.NormalizeCoin(v.CommissionEarnedAmt, v.CommissionEarnedDenom)
			totalCommF += f
			if outDenom == "" {
				outDenom = dd
			}
		}
	}
	if outDenom != "" {
		if totalOutF > 0 || totalCommF > 0 {
			d.TotalOutstanding = fetch.FormatAmountUnit(totalOutF+totalCommF, outDenom) +
				fmt.Sprintf("  across %d validators", len(chain.Validators))
		}
		if totalOutF > 0 {
			d.UnclaimedDelegator = fetch.FormatAmountUnit(totalOutF, outDenom)
		}
		if totalCommF > 0 {
			d.UnclaimedCommission = fetch.FormatAmountUnit(totalCommF, outDenom)
		}
	}

	for _, mod := range chain.ModuleBalances {
		if mod.Name == "" {
			continue
		}
		bal := "0"
		if mod.Amount != "" {
			bal = fetch.FormatCoin(mod.Amount, mod.Denom)
		}
		role := moduleAccountRole(mod.Name)
		d.ModuleAccounts = append(d.ModuleAccounts, model.ModuleAccountRow{
			Name:    mod.Name,
			Address: mod.Address,
			Balance: bal,
			Role:    role,
		})
	}
	if chain.LastBlockFeeRaw != "" {
		lastFeeDenom := denom
		if p.EVMDenom != "" {
			lastFeeDenom = p.EVMDenom
		}
		d.LastBlockFees = fetch.FormatFeeAmount(chain.LastBlockFeeRaw, lastFeeDenom) + "  _(parent block gas × base fee)_"
	}
	if chain.Inflation > 0 && chain.AnnualProvisions != "" && p.BlocksPerYear > 0 {
		provF, _ := fetch.NormalizeCoin(chain.AnnualProvisions, denom)
		_, dispDenom := fetch.NormalizeCoin("0", denom)
		perBlock := provF / float64(p.BlocksPerYear)
		d.InflationPerBlock = fetch.FormatAmountUnit(perBlock, dispDenom) + "/block"
		if chain.BlockInterval > 0 {
			blocksPerDay := 86400.0 / chain.BlockInterval.Seconds()
			d.InflationPerDay = "~" + fetch.FormatAmountUnit(perBlock*blocksPerDay, dispDenom) + "/day"
		}
	}

	d.SlashWindow = FormatInt(p.SignedBlocksWindow)
	d.MinSigned = p.MinSignedPerWindow * 100
	if p.SignedBlocksWindow > 0 {
		d.SlashMaxMissed = int64(float64(p.SignedBlocksWindow) * (1 - p.MinSignedPerWindow))
	}
	d.DowntimeJail = FormatDurFull(p.DowntimeJailDuration)
	d.SlashDowntime, d.SlashDTInactive = slashFraction(p.SlashFractionDowntime)
	d.SlashDS, d.SlashDSInactive = slashFraction(p.SlashFractionDoubleSign)

	feeDenom := p.EVMDenom
	if feeDenom == "" {
		feeDenom = denom
	}
	if chain.BaseFee != "" {
		d.BaseFeeRaw = chain.BaseFee
		d.BaseFee = fetch.FormatFeeAmount(chain.BaseFee, feeDenom)
	}
	if chain.BlockGas > 0 {
		d.BlockGas = FormatInt(int64(chain.BlockGas))
	}
	switch {
	case chain.BlockGasLimit < 0:
		// CometBFT -1 = unlimited; feemarket uses MaxUint64 in CalcGasBaseFee.
		d.BlockGasLimit = ^uint64(0)
	case chain.BlockGasLimit > 0:
		d.BlockGasLimit = uint64(chain.BlockGasLimit)
	}
	d.ParentBlockGasUsed = chain.ParentBlockGasUsed
	d.ParentBlockGasWanted = chain.ParentBlockGasWanted
	d.ParentBlockResultsOK = chain.ParentBlockResultsOK
	if d.ParentBlockGasWanted == 0 && chain.BlockGas > 0 {
		d.ParentBlockGasWanted = chain.BlockGas
	}
	d.BaseFeeChangeDenominator = p.BaseFeeChangeDenominator
	if p.MinGasPriceRaw != "" {
		d.MinGasPriceRaw = p.MinGasPriceRaw
	}
	if p.MinGasPrice > 0 {
		d.MinGasPrice = fetch.FormatAmountUnit(p.MinGasPrice, denom)
	}
	if p.MinGasMultiplier > 0 {
		d.MinGasMultiplier = fmt.Sprintf("%.4g", p.MinGasMultiplier)
	}
	d.NoBaseFee = p.NoBaseFee
	d.Elasticity = p.Elasticity
	d.EnableHeight = p.EnableHeight
	if p.BaseFeeParam != "" {
		d.BaseFeeParam = fetch.FormatFeeAmount(p.BaseFeeParam, feeDenom)
	}
	if chain.MaxBlockBytes > 0 {
		d.MaxBlockBytes = chain.MaxBlockBytes
	}
	appCfg := fetch.FetchAppTomlGasConfig()
	d.NodeAppTomlPath = appCfg.Path
	if appCfg.MinGasPrices != "" {
		d.NodeMinGasPrices = appCfg.MinGasPrices
	}
	if appCfg.EVMMinTip != "" {
		d.NodeEVMMinTip = appCfg.EVMMinTip
	}
	if appCfg.MempoolPriceLimit != "" {
		d.NodeMempoolPriceLimit = appCfg.MempoolPriceLimit
	}
	if appCfg.MaxTxGasWanted != "" {
		d.NodeMaxTxGasWanted = appCfg.MaxTxGasWanted
	}

	d.EVMHTTPEndpoint = evmHTTPEndpoint
	d.EVMWSEndpoint = EVMWSEndpoint(evmHTTPEndpoint)
	d.JSONRPCAPIs = DefaultJSONRPCAPIs
	d.TxpoolGlobalSlots = DefaultTxpoolGlobalSlots
	d.TxpoolGlobalQueue = DefaultTxpoolGlobalQueue
	d.EVMChainID = ev.ChainID
	if p.EVMDenom != "" {
		d.EVMDenom = p.EVMDenom
	}
	d.EVMClient = ev.ClientVersion
	d.EVMRPCOk = ev.Err == nil
	d.EVMListening = ev.NetListening
	d.EVMSynced = !ev.Syncing
	d.EVMBlock = FormatInt(int64(ev.BlockNumber))
	d.PendingTx = ev.PendingTx
	d.QueuedTx = ev.QueuedTx
	if ev.EVMBlockTimestamp > 0 {
		age := time.Since(time.Unix(int64(ev.EVMBlockTimestamp), 0))
		d.EVMBlockAge = fmt.Sprintf("%.1fs", age.Seconds())
		d.EVMBlockAgeWarn = age > 30*time.Second && age <= 2*time.Minute
		d.EVMBlockAgeErr = age > 2*time.Minute
	}
	d.Precompiles = p.ActiveStaticPrecompiles
	if p.HistoryServeWindow > 0 {
		d.HistoryWindow = FormatInt(p.HistoryServeWindow)
	}
	d.HardforkLondon = p.HardforkLondon
	d.HardforkShanghai = p.HardforkShanghai
	d.HardforkCancun = p.HardforkCancun
	d.ERC20Enabled = p.ERC20Enabled

	d.RPCProbeTotal = len(ev.Probes)
	for _, probe := range ev.Probes {
		if probe.OK {
			d.RPCProbeOK++
		}
		d.RPCProbes = append(d.RPCProbes, model.RPCProbe{
			Method:   probe.Method,
			OK:       probe.OK,
			Latency:  fmt.Sprintf("%.0fms", float64(probe.Latency)/float64(time.Millisecond)),
			Error:    probe.Error,
			Request:  probe.Request,
			Response: probe.Response,
		})
	}

	d.VotingPeriod = FormatDurFull(p.VotingPeriod)
	d.Quorum = p.Quorum * 100
	d.Threshold = p.Threshold * 100
	d.VetoThreshold = p.VetoThreshold * 100
	for _, pr := range chain.VotingProposals {
		d.Proposals = append(d.Proposals, model.Proposal{
			ID:           uint64(pr.ID),
			Title:        pr.Title,
			End:          pr.VotingEnd.Format("2006-01-02"),
			TallyYes:     pr.Tally.Yes,
			TallyNo:      pr.Tally.No,
			TallyAbstain: pr.Tally.Abstain,
			TallyVeto:    pr.Tally.NoWithVeto,
			HasTally:     pr.Tally.Yes != "" || pr.Tally.No != "" || pr.Tally.Abstain != "" || pr.Tally.NoWithVeto != "",
		})
	}
	for _, pr := range chain.DepositProposals {
		d.DepositProposals = append(d.DepositProposals, model.Proposal{
			ID:    uint64(pr.ID),
			Title: pr.Title,
			End:   pr.DepositEnd.Format("2006-01-02"),
		})
	}

	d.UpgradeName = chain.UpgradeName
	if chain.UpgradeHeight > 0 {
		d.UpgradeHeight = FormatInt(chain.UpgradeHeight)
		if chain.BlockHeight > 0 && chain.UpgradeHeight > chain.BlockHeight {
			d.BlocksLeft = FormatInt(chain.UpgradeHeight - chain.BlockHeight)
		}
	}

	d.IBCClients = chain.IBCClientCount
	for _, tp := range chain.TokenPairs {
		d.TokenPairs = append(d.TokenPairs, model.TokenPair{
			Denom:   tp.Denom,
			ERC20:   tp.ERC20Addr,
			Enabled: tp.Enabled,
		})
	}

	return d
}

func buildLocalValidator(chain fetch.ChainSnapshot, v *fetch.ValidatorInfo, maxMissed int64) model.LocalValidator {
	lv := model.LocalValidator{
		Moniker:         chain.Moniker,
		NodeID:          chain.NodeID,
		ConsensusAddr:   chain.LocalConsensusAddr,
		ConsensusBech32: chain.LocalConsensusBech32,
		AccountAddr:     chain.LocalAccountAddr,
		P2PDial:         chain.LocalP2PDial,
	}
	if lv.AccountAddr != "" {
		lv.EVMAddr = fetch.AccBech32ToEVM(lv.AccountAddr)
	}
	if v == nil {
		if chain.LocalVotingPower > 0 || chain.LocalConsensusAddr != "" {
			lv.IsValidator = true
			lv.VotingPower = FormatInt(chain.LocalVotingPower)
			lv.SigningStatus = "validator key present — not matched to staking API"
		} else {
			lv.SigningStatus = "this node is not a validator (full node / observer)"
		}
		lv.Delegations = buildDelegationRows(chain.LocalDelegations, lv.AccountAddr)
		return lv
	}

	lv.IsValidator = true
	lv.OperatorAddr = v.OperatorAddr
	if lv.AccountAddr == "" {
		lv.AccountAddr = fetch.ValOperToAcc(v.OperatorAddr)
		if lv.AccountAddr != "" {
			lv.EVMAddr = fetch.AccBech32ToEVM(lv.AccountAddr)
		}
	}
	lv.Delegations = buildDelegationRows(chain.LocalDelegations, lv.AccountAddr)
	if lv.ConsensusAddr == "" {
		lv.ConsensusAddr = v.ConsensusAddr
	}
	if lv.ConsensusBech32 == "" {
		lv.ConsensusBech32 = v.ConsensusBech32
	}
	if lv.P2PDial == "" {
		lv.P2PDial = v.P2PDial
	}
	lv.VotingPower = fetch.FormatCoin(v.VotingPowerTokens, chain.Params.BondDenom)
	lv.VPPercent = v.VotingPowerPercent
	lv.Commission = v.Commission * 100
	lv.Status = strings.ToLower(v.Status)
	lv.Jailed = v.Jailed
	lv.Tombstoned = v.Tombstoned
	lv.Missed = v.MissedBlocks
	lv.MaxMissed = maxMissed
	lv.MissedHigh = maxMissed > 0 && v.MissedBlocks > maxMissed
	lv.ProposerPriority = v.ProposerPriority
	lv.IsNextProposer = v.Moniker == chain.NextProposerMoniker
	lv.Outstanding = v.OutstandingRewards
	lv.CommissionEarned = v.CommissionEarned

	switch {
	case lv.Tombstoned:
		lv.SigningStatus = "TOMBSTONED — permanently removed from validator set"
	case lv.Jailed:
		lv.SigningStatus = "JAILED — not signing blocks"
	case lv.MissedHigh:
		lv.SigningStatus = fmt.Sprintf("⚠ below min signed  (%d missed, max allowed %d in window)", lv.Missed, maxMissed)
	case lv.Missed > 0:
		lv.SigningStatus = fmt.Sprintf("ok  (%d missed in current window)", lv.Missed)
	default:
		lv.SigningStatus = "ok  (no missed blocks in current window)"
	}
	return lv
}

func buildDelegationRows(delegations []fetch.DelegationInfo, localAccount string) []model.DelegationRow {
	if len(delegations) == 0 {
		return nil
	}
	rows := make([]model.DelegationRow, 0, len(delegations))
	for _, d := range delegations {
		rows = append(rows, model.DelegationRow{
			Delegator: d.DelegatorAddr,
			EVMAddr:   fetch.AccBech32ToEVM(d.DelegatorAddr),
			Balance:   fetch.FormatCoin(d.BalanceAmt, d.BalanceDenom),
			IsLocal:   localAccount != "" && d.DelegatorAddr == localAccount,
		})
	}
	return rows
}

func moduleAccountRole(name string) string {
	for _, spec := range []struct{ name, role string }{
		{"fee_collector", "Fees + minted rewards land here each block, then distribution clears"},
		{"distribution", "x/distribution module escrow (often ~0 after BeginBlock payout)"},
		{"bonded_tokens_pool", "Staked tokens (locked; matches staking pool bonded)"},
		{"not_bonded_tokens_pool", "Unbonding / unbonded stake in staking pool"},
		{"gov", "Proposal deposit escrow until voting or refund"},
	} {
		if spec.name == name {
			return spec.role
		}
	}
	return ""
}

func slashFraction(raw string) (string, bool) {
	if raw == "" {
		return "", false
	}
	v := 0.0
	fmt.Sscanf(raw, "%f", &v)
	return FormatFraction(raw), v == 0
}
