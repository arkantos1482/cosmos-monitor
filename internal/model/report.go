package model

// StatusAvailability records which status-strip data sources succeeded.
type StatusAvailability struct {
	ChainOK  bool
	EVMOK    bool
	DockerOK bool
}

// Report holds pre-formatted dashboard data for HTML panel output.
type Report struct {
	Moniker      string
	Synced       bool
	BlockHeight  string
	TimeUTC      string
	PeerCount    int
	EVMPeerCount uint64
	// Status-strip source availability (false → render "—" instead of zero/false defaults).
	HasChainStatus bool
	HasEVMPeers    bool
	HasNodeStatus  bool

	NodeID             string
	AppVersion         string
	BlockInterval      string
	TimeSinceBlock     string
	LatestBlockTime    string
	ListenAddr         string
	RpcListenAddr      string
	Network            string
	LocalConsensusAddr string
	LocalVotingPower   string
	PeerMonikers       []string
	MempoolTxs         int
	NextProposer       string

	Load1, Load5, Load15 float64
	MemUsed, MemTotal    string
	MemPct               int
	DiskUsed, DiskTotal  string
	DiskPct              int

	NodeRunning  bool
	NodeCPU      string
	NodeMemUsed  string
	NodeMemTotal string
	NodeUptime   string
	Restarts     int

	Validators      []Validator
	BondedCount     int
	JailedCount     int
	TombstonedCount int
	BelowThreshold  int

	Local LocalValidator

	BondDenom        string
	TotalSupply      string
	BondedAmt        string
	BondedPct        float64
	GoalBonded       float64
	NotBonded        string
	UnbondingTime    string
	MaxValidators    int64
	Inflation        float64
	AnnualProvisions string
	CommunityPool    string
	CommunityTax     string
	CommunityTaxZero bool
	CommunityTaxPct  float64
	BlocksPerYear    string
	TotalOutstanding    string
	UnclaimedDelegator  string // sum of validator outstanding_rewards (delegator share)
	UnclaimedCommission string // sum of accumulated validator commission

	ModuleAccounts   []ModuleAccountRow
	LastBlockFees    string // parent block gas_used × base_fee (estimate)
	InflationPerBlock string
	InflationPerDay   string

	PMTEnabled     bool
	PMTPoolEmpty   bool
	PMTRate        string
	PMTBalance     string
	PMTRunway      string
	PMTAnnual      string
	PMTPoolAddress string
	PMTDailyEmit   string

	BaseFee                  string
	BaseFeeRaw               string
	BlockGas                 string
	BlockGasLimit            uint64
	MinGasPrice              string
	MinGasPriceRaw           string
	MinGasMultiplier         string
	NoBaseFee                bool
	EnableHeight             int64
	BaseFeeParam             string
	MaxBlockBytes            int64
	NodeMinGasPrices         string
	NodeEVMMinTip            string
	NodeMempoolPriceLimit    string
	NodeMaxTxGasWanted       string
	NodeAppTomlPath          string
	Elasticity               int64
	BaseFeeChangeDenominator int64
	ParentBlockGasUsed       uint64
	ParentBlockGasWanted     uint64
	ParentBlockResultsOK     bool

	SlashWindow     string
	MinSigned       float64
	SlashDowntime   string
	SlashDTInactive bool
	SlashDS         string
	SlashDSInactive bool

	EVMHTTPEndpoint   string
	EVMWSEndpoint     string
	JSONRPCAPIs       string
	TxpoolGlobalSlots uint64
	TxpoolGlobalQueue uint64
	EVMChainID        uint64
	EVMDenom          string
	EVMClient         string
	EVMRPCOk          bool
	EVMListening      bool
	EVMBlockAge       string
	EVMBlockAgeWarn   bool
	EVMBlockAgeErr    bool
	EVMSynced         bool
	EVMBlock          string
	PendingTx         uint64
	QueuedTx          uint64
	RPCProbes         []RPCProbe
	RPCProbeOK        int
	RPCProbeTotal     int
	Precompiles       []string
	HistoryWindow     string
	HardforkLondon    string
	HardforkShanghai  string
	HardforkCancun    string
	ERC20Enabled      bool

	VotingPeriod     string
	Quorum           float64
	Threshold        float64
	VetoThreshold    float64
	Proposals        []Proposal
	DepositProposals []Proposal
	UpgradeName      string
	UpgradeHeight    string
	BlocksLeft       string
	IBCClients       int
	TokenPairs       []TokenPair
}

type Validator struct {
	Moniker         string
	Operator        string
	NodeID          string
	ConsensusAddr   string
	ConsensusBech32 string
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

type LocalValidator struct {
	IsValidator      bool
	Moniker          string
	NodeID           string
	ConsensusAddr    string
	ConsensusBech32  string
	AccountAddr      string
	EVMAddr          string
	P2PDial          string
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

type RPCProbe struct {
	Method   string
	OK       bool
	Latency  string
	Error    string
	Request  string
	Response string
}

type Proposal struct {
	ID           uint64
	Title        string
	End          string
	TallyYes     string
	TallyNo      string
	TallyAbstain string
	TallyVeto    string
	HasTally     bool
}

type TokenPair struct {
	Denom   string
	ERC20   string
	Enabled bool
}

// ModuleAccountRow is a module account balance for the economics overview tables.
type ModuleAccountRow struct {
	Name    string
	Address string
	Balance string
	Role    string
}
