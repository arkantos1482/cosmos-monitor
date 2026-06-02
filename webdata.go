package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/arkantos1482/cosmos-monitor/fetch"
)

// WebData holds all pre-formatted data for dashboard.html rendering.
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
	ListenAddr      string
	Network         string
	PeerMonikers    []string
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

	// validators
	Validators      []WebValidator
	TombstonedCount int
	BelowThreshold  int

	// economics — staking
	TotalSupply   string
	BondedAmt     string
	BondedPct     float64
	GoalBonded    float64
	NotBonded     string
	UnbondingTime string
	MaxValidators int64
	Inflation     float64
	CommunityPool string
	CommunityTax  string
	BlocksPerYear string

	// economics — distribution
	CommunityTaxZero bool

	// economics — PMT rewards
	PMTEnabled     bool
	PMTPoolEmpty   bool
	PMTRate        string
	PMTBalance     string
	PMTRunway      string
	PMTAnnual      string
	PMTPoolAddress string
	// PMT insights
	PMTInsights   bool
	PMTRunwayDays string
	PMTDailyEmit  string
	PMTPerValDay  string
	PMTRevFlow    string
	PMTCommPct    float64
	PMTDelegPct   float64

	// economics — slashing
	SlashWindow     string
	MinSigned       float64
	SlashDowntime   string
	SlashDTInactive bool
	SlashDS         string
	SlashDSInactive bool

	// economics — validator earnings
	TotalOutstanding string
	CommissionRate   float64

	// EVM
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
	BaseFee         string
	GasPrice        string
	MinGasPrice     string
	AdjCap          string
	NoBaseFee       bool
	Elasticity      int64
	ERC20Enabled    bool
	PendingTx       uint64
	QueuedTx        uint64
	Precompiles     []string

	// EVM config
	HistoryWindow   string
	HardforkLondon  string
	HardforkShanghai string
	HardforkCancun  string

	// chain — governance
	VotingPeriod    string
	Quorum          float64
	Threshold       float64
	VetoThreshold   float64
	Proposals       []WebProposal
	DepositProposals []WebProposal

	// chain — upgrade
	UpgradeName   string
	UpgradeHeight string
	BlocksLeft    string

	// chain — IBC + token pairs
	IBCClients int
	TokenPairs []WebTokenPair
}

type WebValidator struct {
	Moniker         string
	VP              string
	VPFloat         float64
	Commission      string
	CommissionFloat float64
	Missed          int64
	Outstanding     string
	Earned          string
	Status          string
	Jailed          bool
	Tombstoned      bool
	MissedHigh      bool
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

	// ── header ──────────────────────────────────────────────────────────────
	d.Moniker = chain.Moniker
	d.Synced = !chain.CatchingUp
	d.BlockHeight = fmtInt(chain.BlockHeight)
	d.TimeUTC = time.Now().UTC().Format("15:04:05") + " UTC"
	d.PeerCount = chain.PeerCount
	d.EVMPeerCount = ev.PeerCount

	// ── node ─────────────────────────────────────────────────────────────────
	d.NodeID      = chain.NodeID
	d.AppVersion  = chain.AppVersion
	d.ListenAddr  = chain.ListenAddr
	d.Network     = chain.Network
	d.PeerMonikers = chain.PeerMonikers
	d.MempoolTxs  = chain.MempoolTxs
	d.NextProposer = chain.NextProposerMoniker
	if chain.BlockInterval > 0 {
		d.BlockInterval = fmtDur(chain.BlockInterval)
	}
	if !chain.LatestBlockTime.IsZero() {
		d.TimeSinceBlock = fmtDur(time.Since(chain.LatestBlockTime))
		d.LatestBlockTime = chain.LatestBlockTime.UTC().Format("2006-01-02 15:04:05 UTC")
	}

	// ── system — OS ──────────────────────────────────────────────────────────
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

	// ── system — container ────────────────────────────────────────────────────
	d.NodeRunning = docker.Running
	d.NodeCPU = fmt.Sprintf("%.1f%%", docker.CPUPercent)
	d.NodeMemUsed = fmtBytes(docker.MemUsage)
	d.NodeMemTotal = fmtBytes(docker.MemLimit)
	d.Restarts = docker.RestartCount
	if !docker.StartedAt.IsZero() {
		d.NodeUptime = fmtDurFull(time.Since(docker.StartedAt))
	}

	// ── validators ────────────────────────────────────────────────────────────
	maxMissed := int64(0)
	if p.SignedBlocksWindow > 0 {
		maxMissed = int64(float64(p.SignedBlocksWindow) * (1 - p.MinSignedPerWindow))
	}
	vals := make([]fetch.ValidatorInfo, len(chain.Validators))
	copy(vals, chain.Validators)
	sort.Slice(vals, func(i, j int) bool { return vals[i].VotingPowerPercent > vals[j].VotingPowerPercent })
	tombCount, belowThresh := 0, 0
	for _, v := range chain.Validators {
		if v.Tombstoned {
			tombCount++
		}
		if v.Status == "BONDED" && !v.Tombstoned && maxMissed > 0 && v.MissedBlocks > maxMissed {
			belowThresh++
		}
	}
	d.TombstonedCount = tombCount
	d.BelowThreshold = belowThresh

	for _, v := range vals {
		wv := WebValidator{
			Moniker:         v.Moniker,
			VP:              fmt.Sprintf("%.1f%%", v.VotingPowerPercent),
			VPFloat:         v.VotingPowerPercent,
			Commission:      fmt.Sprintf("%.1f%%", v.Commission*100),
			CommissionFloat: v.Commission * 100,
			Missed:          v.MissedBlocks,
			Status:          strings.ToLower(v.Status),
			Jailed:          v.Jailed,
			Tombstoned:      v.Tombstoned,
			MissedHigh:      maxMissed > 0 && v.MissedBlocks > maxMissed,
		}
		if v.OutstandingRewards != "" {
			wv.Outstanding = v.OutstandingRewards
		}
		if v.CommissionEarned != "" {
			wv.Earned = v.CommissionEarned
		}
		d.Validators = append(d.Validators, wv)
	}

	// ── economics — staking ───────────────────────────────────────────────────
	denom := p.BondDenom
	if denom == "" {
		denom = chain.TotalSupplyDenom
	}
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
	d.CommunityPool = chain.CommunityPool
	if p.BlocksPerYear > 0 {
		d.BlocksPerYear = fmtInt(p.BlocksPerYear)
	}

	// ── economics — distribution ──────────────────────────────────────────────
	d.CommunityTaxZero = p.CommunityTax == 0
	if p.CommunityTax == 0 {
		d.CommunityTax = "0%  → 100% of tx fees flow to validators"
	} else {
		d.CommunityTax = fmt.Sprintf("%.2f%%", p.CommunityTax*100)
	}

	// ── economics — PMT rewards ────────────────────────────────────────────────
	d.PMTEnabled = p.PMTRewardsEnabled
	d.PMTPoolEmpty = p.PMTRewardsPoolBalanceAmt == "" || p.PMTRewardsPoolBalanceAmt == "0"
	if p.RewardPerBlockAmount != "" {
		d.PMTRate = fetch.FormatCoin(p.RewardPerBlockAmount, p.RewardPerBlockDenom) + "/block"
		if p.BlocksPerYear > 0 {
			rewardF, _ := fetch.NormalizeCoin(p.RewardPerBlockAmount, p.RewardPerBlockDenom)
			_, dispDenom := fetch.NormalizeCoin("0", p.RewardPerBlockDenom)
			d.PMTAnnual = fmt.Sprintf("~%.0f %s/year  (%s blocks × %s)",
				rewardF*float64(p.BlocksPerYear), dispDenom,
				fmtInt(p.BlocksPerYear), fetch.FormatCoin(p.RewardPerBlockAmount, p.RewardPerBlockDenom))
		}
	}
	if !d.PMTPoolEmpty {
		d.PMTBalance = fetch.FormatCoin(p.PMTRewardsPoolBalanceAmt, p.PMTRewardsPoolBalanceDenom)
	}
	d.PMTPoolAddress = p.PMTRewardsPoolAddress

	if p.PMTRewardsEnabled && p.RewardPerBlockAmount != "" && chain.BlockInterval > 0 {
		d.PMTInsights = true
		rewardF, _ := fetch.NormalizeCoin(p.RewardPerBlockAmount, p.RewardPerBlockDenom)
		_, dispDenom := fetch.NormalizeCoin("0", p.RewardPerBlockDenom)
		blocksPerDay := 86400.0 / chain.BlockInterval.Seconds()
		dailyPMT := rewardF * blocksPerDay

		if d.PMTPoolEmpty {
			d.PMTRunwayDays = "EMPTY"
			d.PMTDailyEmit = fmt.Sprintf("0 %s/day  (pool empty)", dispDenom)
		} else {
			poolF, _ := fetch.NormalizeCoin(p.PMTRewardsPoolBalanceAmt, p.PMTRewardsPoolBalanceDenom)
			if dailyPMT > 0 {
				runwayDays := poolF / dailyPMT
				d.PMTRunwayDays = fmt.Sprintf("~%.0f days  (%.2f %s ÷ %.0f/day)", runwayDays, poolF, dispDenom, dailyPMT)
				d.PMTRunway = fmt.Sprintf("~%.0f days left", runwayDays)
			}
			d.PMTDailyEmit = fmt.Sprintf("~%.0f %s/day  (%.4f/block × ~%.0f blocks/day)", dailyPMT, dispDenom, rewardF, blocksPerDay)
		}
		if len(chain.Validators) > 0 {
			perVal := dailyPMT * (chain.Validators[0].VotingPowerPercent / 100)
			d.PMTPerValDay = fmt.Sprintf("~%.0f %s  (%.1f%% VP × %.0f/day)",
				perVal, dispDenom, chain.Validators[0].VotingPowerPercent, dailyPMT)
		}
		rateStr := fetch.FormatCoin(p.RewardPerBlockAmount, p.RewardPerBlockDenom)
		d.PMTRevFlow = fmt.Sprintf("[tx fees] + [pool %s/block] → %d validators", rateStr, max(len(chain.Validators), 1))
		if len(chain.Validators) > 0 {
			d.PMTCommPct = chain.Validators[0].Commission * 100
			d.PMTDelegPct = 100 - d.PMTCommPct
		}
	}

	// ── economics — slashing ──────────────────────────────────────────────────
	d.SlashWindow = fmtInt(p.SignedBlocksWindow)
	d.MinSigned = p.MinSignedPerWindow * 100
	d.SlashDowntime, d.SlashDTInactive = slashFraction(p.SlashFractionDowntime)
	d.SlashDS, d.SlashDSInactive = slashFraction(p.SlashFractionDoubleSign)

	// ── economics — validator earnings ────────────────────────────────────────
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
		d.TotalOutstanding = fmt.Sprintf("%.6f %s  across %d validators", totalOutF, outDenom, len(chain.Validators))
	}
	if len(chain.Validators) > 0 {
		d.CommissionRate = chain.Validators[0].Commission * 100
	}

	// ── EVM ───────────────────────────────────────────────────────────────────
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
	d.BaseFee = chain.BaseFee
	d.GasPrice = ev.GasPrice
	if p.MinGasPrice > 0 {
		d.MinGasPrice = fmt.Sprintf("%.9f %s", p.MinGasPrice, denom)
	}
	d.NoBaseFee = p.NoBaseFee
	d.Elasticity = p.Elasticity
	d.ERC20Enabled = p.ERC20Enabled
	if chain.BaseFee != "" && p.BaseFeeChangeDenominator > 0 {
		baseFeeF := 0.0
		fmt.Sscanf(chain.BaseFee, "%f", &baseFeeF)
		if baseFeeF > 0 {
			cap := baseFeeF / float64(p.BaseFeeChangeDenominator)
			d.AdjCap = fmt.Sprintf("±(base_fee ÷ %d) = ±%g wei/block  (max change per block)", p.BaseFeeChangeDenominator, cap)
		}
	}
	d.Precompiles = p.ActiveStaticPrecompiles

	// EVM config
	if p.HistoryServeWindow > 0 {
		d.HistoryWindow = fmtInt(p.HistoryServeWindow)
	}
	d.HardforkLondon = p.HardforkLondon
	d.HardforkShanghai = p.HardforkShanghai
	d.HardforkCancun = p.HardforkCancun

	// ── chain — governance ────────────────────────────────────────────────────
	d.VotingPeriod = fmtDurFull(p.VotingPeriod)
	d.Quorum = p.Quorum * 100
	d.Threshold = p.Threshold * 100
	d.VetoThreshold = p.VetoThreshold * 100
	for _, pr := range chain.VotingProposals {
		wp := WebProposal{
			ID:           uint64(pr.ID),
			Title:        pr.Title,
			End:          pr.VotingEnd.Format("2006-01-02"),
			TallyYes:     pr.Tally.Yes,
			TallyNo:      pr.Tally.No,
			TallyAbstain: pr.Tally.Abstain,
			TallyVeto:    pr.Tally.NoWithVeto,
			HasTally:     pr.Tally.Yes != "" || pr.Tally.No != "" || pr.Tally.Abstain != "" || pr.Tally.NoWithVeto != "",
		}
		d.Proposals = append(d.Proposals, wp)
	}
	for _, pr := range chain.DepositProposals {
		d.DepositProposals = append(d.DepositProposals, WebProposal{
			ID:    uint64(pr.ID),
			Title: pr.Title,
			End:   pr.DepositEnd.Format("2006-01-02"),
		})
	}

	// ── chain — upgrade ────────────────────────────────────────────────────────
	d.UpgradeName = chain.UpgradeName
	if chain.UpgradeHeight > 0 {
		d.UpgradeHeight = fmtInt(chain.UpgradeHeight)
		if chain.BlockHeight > 0 && chain.UpgradeHeight > chain.BlockHeight {
			d.BlocksLeft = fmtInt(chain.UpgradeHeight - chain.BlockHeight)
		}
	}

	// ── chain — IBC + token pairs ─────────────────────────────────────────────
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


func slashFraction(raw string) (string, bool) {
	if raw == "" {
		return "", false
	}
	v := 0.0
	fmt.Sscanf(raw, "%f", &v)
	return fmtFraction(raw), v == 0
}
