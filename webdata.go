package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/arkantos1482/cosmos-monitor/fetch"
)

// WebData holds all pre-formatted data for dashboard rendering.
type WebData struct {
	Moniker      string
	Synced       bool
	BlockHeight  string
	TimeUTC      string
	PeerCount    int
	EVMPeerCount uint64

	// node
	NodeID          string
	AppVersion      string
	BlockInterval   string
	TimeSinceBlock  string
	LatestBlockTime string
	ListenAddr           string
	RpcListenAddr        string
	Network              string
	LocalConsensusAddr   string
	LocalVotingPower     string
	PeerMonikers         []string
	MempoolTxs      int
	NextProposer    string

	// system — OS
	Load1, Load5, Load15 float64
	MemUsed, MemTotal    string
	MemPct               int
	DiskUsed, DiskTotal  string
	DiskPct              int

	// system — container
	NodeRunning  bool
	NodeCPU      string
	NodeMemUsed  string
	NodeMemTotal string
	NodeUptime   string
	Restarts     int

	// validator set (network)
	Validators      []WebValidator
	BondedCount     int
	JailedCount     int
	TombstonedCount int
	BelowThreshold  int

	// local validator (this node)
	Local WebLocalValidator

	// economics
	BondDenom         string
	TotalSupply       string
	BondedAmt         string
	BondedPct         float64
	GoalBonded        float64
	NotBonded         string
	UnbondingTime     string
	MaxValidators     int64
	Inflation         float64
	AnnualProvisions  string
	CommunityPool     string
	CommunityTax      string
	CommunityTaxZero  bool
	CommunityTaxPct   float64
	BlocksPerYear     string
	TotalOutstanding  string

	// PMT rewards
	PMTEnabled     bool
	PMTPoolEmpty   bool
	PMTRate        string
	PMTBalance     string
	PMTRunway      string
	PMTAnnual      string
	PMTPoolAddress string
	PMTDailyEmit   string

	// fee market (shown in economics)
	BaseFee         string
	BaseFeeRaw      string
	BlockGas        string
	BlockGasLimit   uint64 // consensus max_gas; 0 when unknown or unlimited
	GasPrice        string
	MinGasPrice     string
	MinGasPriceRaw  string
	MinGasMultiplier string
	AdjCap          string
	NoBaseFee       bool
	Elasticity      int64
	BaseFeeChangeDenominator int64
	ParentBlockGasUsed       uint64
	ParentBlockGasWanted     uint64
	ParentBlockResultsOK     bool

	// slashing (validator set section)
	SlashWindow     string
	MinSigned       float64
	SlashDowntime   string
	SlashDTInactive bool
	SlashDS         string
	SlashDSInactive bool

	// EVM JSON-RPC
	EVMEndpoint     string
	EVMChainID      uint64
	EVMDenom        string
	EVMClient       string
	EVMRPCOk        bool
	EVMListening    bool
	EVMBlockAge     string
	EVMBlockAgeWarn bool
	EVMBlockAgeErr  bool
	EVMSynced       bool
	EVMBlock        string
	PendingTx       uint64
	QueuedTx        uint64
	RPCProbes       []WebRPCProbe
	RPCProbeOK      int
	RPCProbeTotal   int
	Precompiles     []string
	HistoryWindow   string
	HardforkLondon  string
	HardforkShanghai string
	HardforkCancun  string
	ERC20Enabled    bool

	// governance
	VotingPeriod     string
	Quorum           float64
	Threshold        float64
	VetoThreshold    float64
	Proposals        []WebProposal
	DepositProposals []WebProposal
	UpgradeName      string
	UpgradeHeight    string
	BlocksLeft       string
	IBCClients       int
	TokenPairs       []WebTokenPair
}

type WebValidator struct {
	Moniker         string
	Operator        string
	NodeID          string
	ConsensusAddr   string
	P2PDial         string
	P2PConnected    bool
	VPFloat         float64
	CommissionFloat float64
	Missed          int64
	MissedHigh      bool
	Status          string
	Jailed          bool
	Tombstoned      bool
	IsLocal         bool
}

type WebLocalValidator struct {
	IsValidator      bool
	Moniker          string
	NodeID           string
	ConsensusAddr    string
	OperatorAddr     string
	VotingPower      string
	VPPercent        float64
	Commission       float64
	Status           string
	Jailed           bool
	Tombstoned       bool
	Missed           int64
	MaxMissed        int64
	MissedHigh       bool
	SigningStatus    string
	IsNextProposer   bool
	ProposerPriority int64
	Outstanding      string
	CommissionEarned string
}

type WebRPCProbe struct {
	Method   string
	OK       bool
	Latency  string
	Error    string
	Request  string
	Response string
}

type WebProposal struct {
	ID           uint64
	Title        string
	End          string
	TallyYes     string
	TallyNo      string
	TallyAbstain string
	TallyVeto    string
	HasTally     bool
}

type WebTokenPair struct {
	Denom   string
	ERC20   string
	Enabled bool
}

func buildWebData(chain fetch.ChainSnapshot, ev fetch.EVMSnapshot, sys fetch.SystemSnapshot, docker fetch.DockerSnapshot) WebData {
	p := chain.Params
	d := WebData{}

	d.Moniker = chain.Moniker
	d.Synced = !chain.CatchingUp
	d.BlockHeight = fmtInt(chain.BlockHeight)
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
		d.LocalVotingPower = fmtInt(chain.LocalVotingPower)
	}
	d.PeerMonikers = chain.PeerMonikers
	d.MempoolTxs = chain.MempoolTxs
	d.NextProposer = chain.NextProposerMoniker
	if chain.BlockInterval > 0 {
		d.BlockInterval = fmtDur(chain.BlockInterval)
	}
	if !chain.LatestBlockTime.IsZero() {
		d.TimeSinceBlock = fmtDur(time.Since(chain.LatestBlockTime))
		d.LatestBlockTime = chain.LatestBlockTime.UTC().Format("2006-01-02 15:04:05 UTC")
	}

	d.Load1, d.Load5, d.Load15 = sys.LoadAvg1, sys.LoadAvg5, sys.LoadAvg15
	memUsed := uint64(0)
	if sys.MemTotal > sys.MemAvail {
		memUsed = sys.MemTotal - sys.MemAvail
	}
	d.MemUsed = fmtBytes(memUsed)
	d.MemTotal = fmtBytes(sys.MemTotal)
	d.MemPct = int(float64(memUsed) / float64(max(sys.MemTotal, uint64(1))) * 100)
	d.DiskUsed = fmtBytes(sys.DiskUsed)
	d.DiskTotal = fmtBytes(sys.DiskTotal)
	d.DiskPct = int(float64(sys.DiskUsed) / float64(max(sys.DiskTotal, uint64(1))) * 100)

	d.NodeRunning = docker.Running
	d.NodeCPU = fmt.Sprintf("%.1f%%", docker.CPUPercent)
	d.NodeMemUsed = fmtBytes(docker.MemUsage)
	d.NodeMemTotal = fmtBytes(docker.MemLimit)
	d.Restarts = docker.RestartCount
	if !docker.StartedAt.IsZero() {
		d.NodeUptime = fmtDurFull(time.Since(docker.StartedAt))
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
		d.Validators = append(d.Validators, WebValidator{
			Moniker:         v.Moniker,
			Operator:        v.OperatorAddr,
			NodeID:          v.NodeID,
			ConsensusAddr:   strings.ToUpper(v.ConsensusAddr),
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
	d.UnbondingTime = fmtDurFull(p.UnbondingTime)
	if p.MaxValidators > 0 {
		d.MaxValidators = int64(p.MaxValidators)
	}
	d.Inflation = chain.Inflation * 100
	if chain.AnnualProvisions != "" {
		d.AnnualProvisions = fetch.FormatCoin(chain.AnnualProvisions, denom)
	}
	d.CommunityPool = chain.CommunityPool
	if p.BlocksPerYear > 0 {
		d.BlocksPerYear = fmtInt(p.BlocksPerYear)
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

	var totalOutF float64
	var outDenom string
	for _, v := range chain.Validators {
		if v.OutstandingRewardsAmt != "" {
			f, dd := fetch.NormalizeCoin(v.OutstandingRewardsAmt, v.OutstandingRewardsDenom)
			totalOutF += f
			outDenom = dd
		}
	}
	if outDenom != "" {
		d.TotalOutstanding = fetch.FormatAmountUnit(totalOutF, outDenom) +
			fmt.Sprintf("  across %d validators", len(chain.Validators))
	}

	d.SlashWindow = fmtInt(p.SignedBlocksWindow)
	d.MinSigned = p.MinSignedPerWindow * 100
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
		d.BlockGas = fmtInt(int64(chain.BlockGas))
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
	d.GasPrice = ev.GasPrice
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
	if chain.BaseFee != "" && p.BaseFeeChangeDenominator > 0 {
		baseFeeF, _ := strconv.ParseFloat(chain.BaseFee, 64)
		if baseFeeF > 0 {
			cap := baseFeeF / float64(p.BaseFeeChangeDenominator)
			_, capDenom := fetch.NormalizeCoin("0", feeDenom)
			d.AdjCap = "±" + fetch.FormatAmountUnit(cap, capDenom) + "/block" +
				fmt.Sprintf("  (base_fee ÷ %d)", p.BaseFeeChangeDenominator)
		}
	}

	d.EVMChainID = ev.ChainID
	if p.EVMDenom != "" {
		d.EVMDenom = p.EVMDenom
	}
	d.EVMClient = ev.ClientVersion
	d.EVMRPCOk = ev.Err == nil
	d.EVMListening = ev.NetListening
	d.EVMSynced = !ev.Syncing
	d.EVMBlock = fmtInt(int64(ev.BlockNumber))
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
		d.HistoryWindow = fmtInt(p.HistoryServeWindow)
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
		d.RPCProbes = append(d.RPCProbes, WebRPCProbe{
			Method:   probe.Method,
			OK:       probe.OK,
			Latency:  fmt.Sprintf("%.0fms", float64(probe.Latency)/float64(time.Millisecond)),
			Error:    probe.Error,
			Request:  fetch.TruncateJSON(probe.Request, 120),
			Response: fetch.TruncateJSON(probe.Response, 180),
		})
	}

	d.VotingPeriod = fmtDurFull(p.VotingPeriod)
	d.Quorum = p.Quorum * 100
	d.Threshold = p.Threshold * 100
	d.VetoThreshold = p.VetoThreshold * 100
	for _, pr := range chain.VotingProposals {
		d.Proposals = append(d.Proposals, WebProposal{
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
		d.DepositProposals = append(d.DepositProposals, WebProposal{
			ID:    uint64(pr.ID),
			Title: pr.Title,
			End:   pr.DepositEnd.Format("2006-01-02"),
		})
	}

	d.UpgradeName = chain.UpgradeName
	if chain.UpgradeHeight > 0 {
		d.UpgradeHeight = fmtInt(chain.UpgradeHeight)
		if chain.BlockHeight > 0 && chain.UpgradeHeight > chain.BlockHeight {
			d.BlocksLeft = fmtInt(chain.UpgradeHeight - chain.BlockHeight)
		}
	}

	d.IBCClients = chain.IBCClientCount
	for _, tp := range chain.TokenPairs {
		d.TokenPairs = append(d.TokenPairs, WebTokenPair{
			Denom:   tp.Denom,
			ERC20:   tp.ERC20Addr,
			Enabled: tp.Enabled,
		})
	}

	return d
}

func buildLocalValidator(chain fetch.ChainSnapshot, v *fetch.ValidatorInfo, maxMissed int64) WebLocalValidator {
	lv := WebLocalValidator{
		Moniker:       chain.Moniker,
		NodeID:        chain.NodeID,
		ConsensusAddr: chain.LocalConsensusAddr,
	}
	if v == nil {
		if chain.LocalVotingPower > 0 || chain.LocalConsensusAddr != "" {
			lv.IsValidator = true
			lv.VotingPower = fmtInt(chain.LocalVotingPower)
			lv.SigningStatus = "validator key present — not matched to staking API"
		} else {
			lv.SigningStatus = "this node is not a validator (full node / observer)"
		}
		return lv
	}

	lv.IsValidator = true
	lv.OperatorAddr = v.OperatorAddr
	if lv.ConsensusAddr == "" {
		lv.ConsensusAddr = v.ConsensusAddr
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

func slashFraction(raw string) (string, bool) {
	if raw == "" {
		return "", false
	}
	v := 0.0
	fmt.Sscanf(raw, "%f", &v)
	return fmtFraction(raw), v == 0
}
