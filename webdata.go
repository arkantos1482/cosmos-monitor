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
	Moniker     string
	Synced      bool
	BlockHeight string
	TimeUTC     string
	PeerCount   int
	EVMPeerCount uint64

	// system — OS
	Load1, Load5, Load15 float64
	MemUsed, MemTotal    string
	MemPct               int
	DiskUsed, DiskTotal  string
	DiskPct              int

	// system — container
	NodeRunning bool
	NodeCPU     string
	NodeMemUsed string
	NodeMemTotal string
	NodeUptime  string
	Restarts    int

	// validators
	Validators []WebValidator

	// economics — staking
	TotalSupply   string
	BondedAmt     string
	BondedPct     float64
	GoalBonded    float64
	NotBonded     string
	Inflation     float64
	CommunityPool string
	BlocksPerYear string

	// economics — PMT rewards
	PMTEnabled   bool
	PMTPoolEmpty bool
	PMTStatus    string
	PMTRate      string
	PMTBalance   string
	PMTRunway    string
	PMTAnnual    string

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
	EVMChainID     uint64
	EVMClient      string
	EVMRPCOk       bool
	EVMListening   bool
	EVMBlockAge    string
	EVMBlockAgeWarn bool
	EVMBlockAgeErr  bool
	EVMSynced      bool
	EVMBlock       string
	BaseFee        string
	GasPrice       string
	AdjCap         string
	NoBaseFee      bool
	Elasticity     int64
	ERC20Enabled   bool
	PendingTx      uint64
	QueuedTx       uint64

	// chain — governance
	VotingPeriod  string
	Quorum        float64
	Threshold     float64
	VetoThreshold float64
	Proposals     []WebProposal

	// chain — slashing (window/threshold — subset for CHAIN section)
	ChainSlashDT string
	ChainSlashDS string

	// chain — upgrade
	UpgradeName   string
	UpgradeHeight string
	BlocksLeft    string

	// chain — IBC + token pairs
	IBCClients int
	TokenPairs []WebTokenPair
}

type WebValidator struct {
	Moniker     string
	VP          string
	Commission  string
	Missed      int64
	Outstanding string
	Earned      string
	Status      string
	Jailed      bool
	Tombstoned  bool
	MissedHigh  bool
}

type WebProposal struct {
	ID    uint64
	Title string
	End   string
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

	// ── system — OS ──────────────────────────────────────────────────────────
	d.Load1, d.Load5, d.Load15 = sys.LoadAvg1, sys.LoadAvg5, sys.LoadAvg15
	memUsed := uint64(0)
	if sys.MemTotal > sys.MemAvail {
		memUsed = sys.MemTotal - sys.MemAvail
	}
	d.MemUsed = fmtBytes(memUsed)
	d.MemTotal = fmtBytes(sys.MemTotal)
	d.MemPct = int(float64(memUsed) / float64(max64(sys.MemTotal, 1)) * 100)
	d.DiskUsed = fmtBytes(sys.DiskUsed)
	d.DiskTotal = fmtBytes(sys.DiskTotal)
	d.DiskPct = int(float64(sys.DiskUsed) / float64(max64(sys.DiskTotal, 1)) * 100)

	// ── system — container ────────────────────────────────────────────────────
	d.NodeRunning = docker.Running
	d.NodeCPU = fmt.Sprintf("%.1f%%", docker.CPUPercent)
	d.NodeMemUsed = fmtBytes(docker.MemUsage)
	d.NodeMemTotal = fmtBytes(docker.MemLimit)
	d.Restarts = docker.RestartCount
	if !docker.StartedAt.IsZero() {
		d.NodeUptime = fmtUptime(docker.StartedAt)
	}

	// ── validators ────────────────────────────────────────────────────────────
	maxMissed := int64(0)
	if p.SignedBlocksWindow > 0 {
		maxMissed = int64(float64(p.SignedBlocksWindow) * (1 - p.MinSignedPerWindow))
	}
	vals := make([]fetch.ValidatorInfo, len(chain.Validators))
	copy(vals, chain.Validators)
	sort.Slice(vals, func(i, j int) bool { return vals[i].VotingPowerPercent > vals[j].VotingPowerPercent })
	for _, v := range vals {
		wv := WebValidator{
			Moniker:    v.Moniker,
			VP:         fmt.Sprintf("%.1f%%", v.VotingPowerPercent),
			Commission: fmt.Sprintf("%.1f%%", v.Commission*100),
			Missed:     v.MissedBlocks,
			Status:     strings.ToLower(v.Status),
			Jailed:     v.Jailed,
			Tombstoned: v.Tombstoned,
			MissedHigh: maxMissed > 0 && v.MissedBlocks > maxMissed,
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
	d.Inflation = chain.Inflation * 100
	d.CommunityPool = chain.CommunityPool
	if p.BlocksPerYear > 0 {
		d.BlocksPerYear = fmtInt(p.BlocksPerYear)
	}

	// ── economics — PMT rewards ────────────────────────────────────────────────
	d.PMTEnabled = p.PMTRewardsEnabled
	d.PMTPoolEmpty = p.PMTRewardsPoolBalanceAmt == "" || p.PMTRewardsPoolBalanceAmt == "0"
	d.PMTStatus = pmtRewardsStatusPlain(p)
	if p.RewardPerBlockAmount != "" {
		d.PMTRate = fetch.FormatCoin(p.RewardPerBlockAmount, p.RewardPerBlockDenom) + "/block"
		if p.BlocksPerYear > 0 {
			rewardF, _ := fetch.NormalizeCoin(p.RewardPerBlockAmount, p.RewardPerBlockDenom)
			_, dispDenom := fetch.NormalizeCoin("0", p.RewardPerBlockDenom)
			d.PMTAnnual = fmt.Sprintf("~%.0f %s/year", rewardF*float64(p.BlocksPerYear), dispDenom)
		}
	}
	if !d.PMTPoolEmpty {
		d.PMTBalance = fetch.FormatCoin(p.PMTRewardsPoolBalanceAmt, p.PMTRewardsPoolBalanceDenom)
		d.PMTRunway = poolRunwayPlain(p, chain.BlockInterval)
	}

	// ── economics — slashing ──────────────────────────────────────────────────
	d.SlashWindow = fmtInt(p.SignedBlocksWindow)
	d.MinSigned = p.MinSignedPerWindow * 100
	if p.SlashFractionDowntime != "" {
		d.SlashDowntime = fmtFraction(p.SlashFractionDowntime)
		dtF := 0.0
		fmt.Sscanf(p.SlashFractionDowntime, "%f", &dtF)
		d.SlashDTInactive = dtF == 0
	}
	if p.SlashFractionDoubleSign != "" {
		d.SlashDS = fmtFraction(p.SlashFractionDoubleSign)
		dsF := 0.0
		fmt.Sscanf(p.SlashFractionDoubleSign, "%f", &dsF)
		d.SlashDSInactive = dsF == 0
	}

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
		d.TotalOutstanding = fmt.Sprintf("%.6f %s", totalOutF, outDenom)
	}
	if len(chain.Validators) > 0 {
		d.CommissionRate = chain.Validators[0].Commission * 100
	}

	// ── EVM ───────────────────────────────────────────────────────────────────
	d.EVMChainID = ev.ChainID
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
	d.NoBaseFee = p.NoBaseFee
	d.Elasticity = p.Elasticity
	d.ERC20Enabled = p.ERC20Enabled
	if chain.BaseFee != "" && p.BaseFeeChangeDenominator > 0 {
		baseFeeF := 0.0
		fmt.Sscanf(chain.BaseFee, "%f", &baseFeeF)
		if baseFeeF > 0 {
			cap := baseFeeF / float64(p.BaseFeeChangeDenominator)
			d.AdjCap = fmt.Sprintf("±%g wei/block  (base_fee ÷ %d)", cap, p.BaseFeeChangeDenominator)
		}
	}

	// ── chain — governance ────────────────────────────────────────────────────
	d.VotingPeriod = fmtDurFull(p.VotingPeriod)
	d.Quorum = p.Quorum * 100
	d.Threshold = p.Threshold * 100
	d.VetoThreshold = p.VetoThreshold * 100
	for _, pr := range chain.VotingProposals {
		d.Proposals = append(d.Proposals, WebProposal{
			ID:    uint64(pr.ID),
			Title: pr.Title,
			End:   pr.VotingEnd.Format("2006-01-02"),
		})
	}

	// ── chain — slashing (subset for CHAIN section) ───────────────────────────
	d.ChainSlashDT = d.SlashDowntime
	d.ChainSlashDS = d.SlashDS

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

func pmtRewardsStatusPlain(p fetch.ChainParams) string {
	if !p.PMTRewardsEnabled {
		return "disabled"
	}
	if p.PMTRewardsPoolBalanceAmt == "" || p.PMTRewardsPoolBalanceAmt == "0" {
		return "ENABLED — pool EMPTY"
	}
	return "distributing"
}

func poolRunwayPlain(p fetch.ChainParams, blockInterval time.Duration) string {
	if blockInterval <= 0 || p.RewardPerBlockAmount == "" || p.PMTRewardsPoolBalanceAmt == "" {
		return ""
	}
	poolF, _ := fetch.NormalizeCoin(p.PMTRewardsPoolBalanceAmt, p.PMTRewardsPoolBalanceDenom)
	rewardF, _ := fetch.NormalizeCoin(p.RewardPerBlockAmount, p.RewardPerBlockDenom)
	if rewardF <= 0 || poolF <= 0 {
		return ""
	}
	blocksPerDay := 86400.0 / blockInterval.Seconds()
	days := poolF / (rewardF * blocksPerDay)
	return fmt.Sprintf("~%.0f days left", days)
}

func meterColor(pct int) string {
	if pct > 90 {
		return "var(--red)"
	}
	if pct > 75 {
		return "var(--yellow)"
	}
	return "var(--green)"
}

func max64(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}
