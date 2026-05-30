# pmtop — implementation handover

Single-node Cosmos SDK + EVM chain monitoring TUI.
One command on the node, full picture at a glance — across three lenses:
**tech** (is the infrastructure healthy?) · **business** (is the chain active?) · **economics** (what does the value picture look like?)

---

## Layout philosophy

The UI mirrors how the chain is actually architected — stack layers, not invented categories.
Anyone who works with a Cosmos-EVM chain professionally already thinks in these layers.
New modules or metrics always have an obvious home.

### Persistent header — the sentinel

One bar, always on screen. Answers the single most urgent question: **is anything on fire?**
If every value here is green, the operator can relax and read the panels below.
If anything is red, it jumps out before they read anything else.

Contains: block height · block interval · sync status · peer count · container status · CPU · RAM · disk

### Four panels — the stack layers

| Panel | Layer | Question |
|-------|-------|----------|
| **SYSTEM** | OS + Docker | Is the machine healthy? |
| **CONSENSUS** | CometBFT | Is the network agreeing? |
| **APPLICATION** | Cosmos SDK modules | What is the chain's state? |
| **EVM** | EVM + feemarket + ERC20 | What is the EVM layer doing? |

Arranged as a 2×2 grid. Each panel is independent — a failure fetching one doesn't affect the others.

### Validator table — full-width, scrollable

Dense per-validator detail. Spans application-layer data (voting power, missed blocks, jailed status)
and economic data (commission rate, outstanding rewards, commission earned).
Too much for a panel cell — gets its own section.

### Governance strip — event-driven, bottom

Governance proposals and upgrade plan. Near-static when nothing is happening.
High urgency when a vote is closing or an upgrade is imminent.
Lives at the bottom so it doesn't crowd the always-relevant panels above.

---

The **fetch layer** is organized by *data source* (transport concern).
The **UI layer** is organized by *stack layer* (meaning concern).
These are separate — never mix them.

---

## Locked decisions

| Decision | Choice | Reason |
|----------|--------|--------|
| Where it runs | **On the node** (ssh in, run `pmtop`) | direct procfs + Docker socket access |
| Refresh | **Manual only** — fetch on launch + `r` to re-fetch | no ticker, no background goroutines between fetches |
| Config | **CLI flags only** | no config file; interval/multi-node/alerts are a future layer |
| Transport | **Plain HTTP everywhere** | one pattern for all sources; Docker socket is also HTTP over Unix socket |
| Language | **Go** | single binary, no runtime deps, fits chain tooling |
| TUI | **Bubbletea + Lipgloss + Bubbles** | Charm.sh ecosystem; tick→fetch→render maps perfectly to manual refresh model |

---

## CLI flags

```
pmtop [flags]

--rpc        string   CometBFT RPC endpoint   (default: http://localhost:26657)
--rest       string   Cosmos REST/LCD endpoint (default: http://localhost:1317)
--evm        string   EVM JSON-RPC endpoint    (default: http://localhost:8545)
--container  string   Docker container name    (default: evmd-node)
```

All flags have the defaults above, so bare `pmtop` works on a standard node.

---

## App lifecycle

```
parse flags
  → build Config
  → start bubbletea
    → on Init:    fire fetchAll cmd
    → on r key:   fire fetchAll cmd
    → on q / ctrl+c: quit
    → fetchAll returns snapshotMsg
    → model updates, view re-renders
    → wait for next keypress
```

`fetchAll` fires all four domain fetchers **concurrently** via `tea.Batch`.
Each domain returns independently — if one fails the others still display.

---

## File structure

```
tools/ops/pmtop/
├── main.go               parse flags → Config → tea.NewProgram → Run
│
├── fetch/                organized by DATA SOURCE (transport concern)
│   ├── client.go         shared http.Client + doJSON(url, target) helper
│   │                     + newDockerClient() (Unix socket http.Client)
│   ├── chain.go          all CometBFT :26657 + Cosmos REST :1317 calls
│   ├── evm.go            all EVM JSON-RPC :8545 calls
│   ├── system.go         /proc/loadavg, /proc/meminfo, df syscall
│   └── docker.go         Docker socket → stats + inspect
│
├── snapshot.go           all domain types: ChainSnapshot, EVMSnapshot,
│                         SystemSnapshot, DockerSnapshot, Snapshot (root)
│
└── ui/                   organized by STACK LAYER (meaning concern)
    ├── model.go          bubbletea Model struct + Init()
    ├── update.go         Update() — handle keyMsg + snapshotMsg + errMsg
    ├── view.go           View() — assembles header + 2×2 grid + table + strip
    ├── header.go         renderHeader()      — node pulse sentinel bar
    ├── system.go         renderSystem()      — SYSTEM panel: OS + Docker
    ├── consensus.go      renderConsensus()   — CONSENSUS panel: CometBFT
    ├── application.go    renderApplication() — APPLICATION panel: SDK modules
    ├── evm.go            renderEVM()         — EVM panel: feemarket + ERC20
    ├── validators.go     renderValidators()  — full-width validator table
    └── governance.go     renderGovernance()  — proposals + upgrade strip
```

**Invariant**: `fetch/` = pure functions, zero state. Organized by data source.
`ui/` = pure render from snapshot. Organized by stack layer.
Adding a new metric: 1 field in `snapshot.go`, 1 call in the right `fetch/*.go`, 1 line in the right `ui/*.go`. Nothing else changes.

---

## Data types (snapshot.go)

```go
type Config struct {
    RPC       string // --rpc
    REST      string // --rest
    EVM       string // --evm
    Container string // --container
}

type ChainSnapshot struct {
    // identity
    NodeID    string
    Moniker   string
    AppVersion string

    // live state
    BlockHeight     int64
    LatestBlockTime time.Time
    BlockInterval   time.Duration // current block - previous block timestamp
    CatchingUp      bool
    PeerCount       int

    // validators
    Validators []ValidatorInfo

    // staking / economics
    BondedTokens    sdk.Int // or string
    NotBondedTokens sdk.Int
    TotalSupply     sdk.Coin
    Inflation       sdk.Dec
    CommunityPool   sdk.DecCoins
    RewardPerBlock  sdk.Dec // pmtrewards param

    // feemarket
    BaseFee  sdk.Int
    BlockGas uint64

    // governance
    Proposals []ProposalInfo

    // upgrade
    UpgradeName   string // empty = no pending upgrade
    UpgradeHeight int64

    // erc20
    TokenPairs []TokenPairInfo

    // ibc
    IBCClientCount int

    // params (fetched once, on launch only)
    Params ChainParams

    Err error
}

type ValidatorInfo struct {
    Moniker            string
    OperatorAddr       string // valoper
    ConsensusAddr      string // valcons
    Status             string // BONDED / UNBONDING / UNBONDED
    Jailed             bool
    Tombstoned         bool
    VotingPowerTokens  sdk.Int
    VotingPowerPercent sdk.Dec
    Commission         sdk.Dec
    MissedBlocks       int64
    OutstandingRewards sdk.DecCoins
    CommissionEarned   sdk.DecCoins
    ProposerPriority   int64
}

type ProposalInfo struct {
    ID          uint64
    Title       string
    Status      string // VOTING_PERIOD / DEPOSIT_PERIOD / PASSED / REJECTED
    VotingEnd   time.Time
    DepositEnd  time.Time
}

type TokenPairInfo struct {
    Denom       string
    ERC20Addr   string
    Enabled     bool
}

type ChainParams struct {
    // staking
    UnbondingTime  time.Duration
    MaxValidators  int
    BondDenom      string
    // slashing
    SignedBlocksWindow int64
    MinSignedPerWindow sdk.Dec
    // mint
    BlocksPerYear int64
    GoalBonded    sdk.Dec
    // gov
    VotingPeriod  time.Duration
    Quorum        sdk.Dec
    Threshold     sdk.Dec
    // evm
    EVMDenom      string
    // feemarket
    MinGasPrice   sdk.Dec
    Elasticity    int64
    // erc20
    ERC20Enabled  bool
}

type EVMSnapshot struct {
    BlockNumber uint64
    ChainID     uint64
    Syncing     bool
    GasPrice    string
    PendingTx   uint64
    QueuedTx    uint64
    PeerCount   uint64
    Err         error
}

type SystemSnapshot struct {
    LoadAvg1  float64
    LoadAvg5  float64
    LoadAvg15 float64
    MemTotal  uint64
    MemAvail  uint64
    DiskTotal uint64
    DiskUsed  uint64
    Err       error
}

type DockerSnapshot struct {
    Running      bool
    CPUPercent   float64
    MemUsage     uint64
    MemLimit     uint64
    RestartCount int
    StartedAt    time.Time
    Err          error
}

type Snapshot struct {
    Chain  ChainSnapshot
    EVM    EVMSnapshot
    System SystemSnapshot
    Docker DockerSnapshot
    FetchedAt time.Time
}
```

---

## fetch/client.go

```go
// shared HTTP client (5s timeout)
var httpClient = &http.Client{Timeout: 5 * time.Second}

// doJSON fetches url and decodes JSON into target
func doJSON(url string, target any) error

// newDockerClient returns an http.Client that dials /var/run/docker.sock
func newDockerClient() *http.Client
```

---

## fetch/chain.go — full endpoint list

All functions take `cfg Config` and return typed structs + error.

### CometBFT RPC (:26657) — plain JSON over HTTP GET

| Endpoint | Returns |
|----------|---------|
| `GET /status` | node_id, moniker, app_version, block_height, latest_block_time, catching_up |
| `GET /net_info` | n_peers, peers list |
| `GET /validators?height=0&per_page=100` | validator set with proposer_priority, pub_key (→ valcons) |
| `GET /block` | latest block: height, time |
| `GET /block?height={latest-1}` | previous block: time (used to compute block interval) |

### Cosmos REST LCD (:1317) — REST/JSON

**Live state:**

| Endpoint | Returns |
|----------|---------|
| `GET /cosmos/staking/v1beta1/validators?status=BOND_STATUS_BONDED&pagination.limit=100` | validators: operator_address, description.moniker, status, tokens, commission.commission_rates.rate, jailed |
| `GET /cosmos/staking/v1beta1/validators?status=BOND_STATUS_UNBONDING&pagination.limit=100` | (same fields, merge with bonded list) |
| `GET /cosmos/staking/v1beta1/validators?status=BOND_STATUS_UNBONDED&pagination.limit=100` | (same) |
| `GET /cosmos/staking/v1beta1/pool` | bonded_tokens, not_bonded_tokens |
| `GET /cosmos/slashing/v1beta1/signing_infos?pagination.limit=100` | valcons_address, missed_blocks_counter, tombstoned |
| `GET /cosmos/distribution/v1beta1/validators/{valoper}/outstanding_rewards` | per validator: rewards (DecCoins) |
| `GET /cosmos/distribution/v1beta1/validators/{valoper}/commission` | per validator: commission (DecCoins) |
| `GET /cosmos/bank/v1beta1/supply` | total supply (all denoms) |
| `GET /cosmos/mint/v1beta1/inflation` | current inflation rate |
| `GET /cosmos/distribution/v1beta1/community_pool` | community pool balance |
| `GET /ethermint/feemarket/v1/base_fee` | current EIP-1559 base fee |
| `GET /ethermint/feemarket/v1/block_gas` | gas used in last block |
| `GET /cosmos/gov/v1beta1/proposals?proposal_status=2` | voting period proposals |
| `GET /cosmos/gov/v1beta1/proposals?proposal_status=1` | deposit period proposals |
| `GET /cosmos/upgrade/v1beta1/current_plan` | pending upgrade name + height |
| `GET /evmos/erc20/v1/token_pairs` | registered cosmos↔evm token bridges |
| `GET /ibc/core/client/v1/client_states` | IBC client count |

**Params (fetch once on launch, not on every `r`):**

| Endpoint | Returns |
|----------|---------|
| `GET /cosmos/staking/v1beta1/params` | unbonding_time, max_validators, bond_denom |
| `GET /cosmos/slashing/v1beta1/params` | signed_blocks_window, min_signed_per_window, slash_fraction_double_sign |
| `GET /cosmos/mint/v1beta1/params` | blocks_per_year, goal_bonded |
| `GET /cosmos/distribution/v1beta1/params` | community_tax |
| `GET /cosmos/gov/v1beta1/params/voting` | voting_period |
| `GET /cosmos/gov/v1beta1/params/tallying` | quorum, threshold |
| `GET /ethermint/feemarket/v1/params` | min_gas_price, elasticity_multiplier, base_fee_change_denominator |
| `GET /ethermint/evm/v1/params` | evm_denom |
| `GET /cosmos/auth/v1beta1/params` | max_memo_characters, tx_sig_limit |

> Note: outstanding_rewards and commission require one request **per validator**.
> Fire them concurrently via a worker pool (e.g., semaphore of 10) to avoid hammering the node.

---

## fetch/evm.go — full endpoint list

All calls to EVM JSON-RPC `:8545` use POST with `Content-Type: application/json`.

```json
{"jsonrpc":"2.0","method":"<method>","params":[],"id":1}
```

| Method | Returns |
|--------|---------|
| `eth_blockNumber` | latest EVM block number (hex) |
| `eth_chainId` | chain ID (hex) |
| `eth_syncing` | false if synced, object if syncing |
| `eth_gasPrice` | current gas price (hex wei) |
| `txpool_status` | pending (hex), queued (hex) |
| `net_peerCount` | peer count (hex) |

---

## fetch/system.go

| Source | Parses |
|--------|--------|
| `/proc/loadavg` | load averages: 1m, 5m, 15m |
| `/proc/meminfo` | MemTotal, MemAvailable |
| `syscall.Statfs("/")` | disk total blocks, free blocks, block size → total + used bytes |

No external commands. Pure Go file reads + syscall.

---

## fetch/docker.go

Uses `newDockerClient()` (HTTP over `/var/run/docker.sock`). Base URL is `http://localhost`.

| Endpoint | Returns |
|----------|---------|
| `GET /containers/{name}/stats?stream=false` | cpu_stats, precpu_stats, memory_stats → compute CPU% |
| `GET /containers/{name}/json` | State.Running, RestartCount, State.StartedAt |

CPU% formula:
```
cpuDelta = cpu_stats.cpu_usage.total_usage - precpu_stats.cpu_usage.total_usage
sysDelta  = cpu_stats.system_cpu_usage - precpu_stats.system_cpu_usage
cpuPercent = (cpuDelta / sysDelta) * numCPUs * 100
```

---

## ui/model.go

```go
type Model struct {
    config    Config
    snapshot  *Snapshot   // nil until first fetch completes
    loading   bool        // true while fetchAll in flight
    params    ChainParams // fetched once, survives re-fetches
    paramsLoaded bool
}

// messages
type snapshotMsg Snapshot
type paramsMsg   ChainParams
type errMsg      struct{ err error }
```

`Init()` returns `tea.Batch(fetchParams(cfg), fetchAll(cfg))`.

On `r` key: returns `fetchAll(cfg)` cmd (params are already loaded, not re-fetched).

---

## ui/view.go — layout

```
┌─ pmtop · {moniker} · {node_id[:8]} ──────────────── {HH:MM:SS} UTC ─┐
│  {height}  ·  {n}s/blk  ·  {SYNCED ✓ | CATCHING UP ✗}  ·  {n} peers │
│  container {running ✓ | stopped ✗}  ·  CPU {n}%  ·  RAM {n}%  ·  Disk {n}% │
├─────────────────────────────────┬────────────────────────────────────┤
│ SYSTEM                          │ CONSENSUS                          │
│                                 │                                    │
│ CPU       45%                   │ Height    1,234,567                │
│ RAM       8.2 / 16 GB    51%    │ Block     5.18s                    │
│ Disk      234 / 500 GB   46%    │ Sync      SYNCED ✓                 │
│ Load      1.2  0.9  0.8         │ Peers     12                       │
│                                 │                                    │
│ Container  evmd-node            │ Validators                         │
│  Status    running ✓            │  15 bonded · 0 jailed · 0 tombstoned│
│  Restarts  0                    │  next proposer: alice              │
│  Uptime    3d 14h               │                                    │
│  CPU       12%                  │ Signing window   10,000 blocks     │
│  RAM       2.1 / 8 GB           │ Min signed       5%                │
├─────────────────────────────────┼────────────────────────────────────┤
│ APPLICATION                     │ EVM                                │
│                                 │                                    │
│ Supply     1,000,000,000 PMT    │ Chain ID   290290                  │
│ Bonded     670,000,000   67%    │ Sync       synced ✓                │
│            ███████░░░           │                                    │
│ Inflation  0.00%                │ txpool     3 pending / 0 queued    │
│ Reward/blk 0.1 PMT              │ Base fee   0 apmt                  │
│ Community  50,000 PMT           │ Gas used   2.1M / 44M  ████░       │
│ Comm. tax  0%                   │ Gas price  0                       │
│ Goal bond  67%                  │                                    │
│ Unbonding  504h                 │ ERC20 pairs  3 registered          │
│                                 │ IBC clients  0                     │
├─────────────────────────────────┴────────────────────────────────────┤
│ VALIDATORS                                                            │
│ moniker      vp%   commission  missed   outstanding    earned  status │
│ alice        8.2%  5.0%             0   12.3 PMT    1.1 PMT   ● BON  │
│ bob          7.9%  3.0%            47    9.1 PMT    0.8 PMT   ● BON  │
│ charlie ⚠    6.1%  10.0%          210    7.8 PMT    0.7 PMT   ● BON  │
├──────────────────────────────────────────────────────────────────────┤
│ GOVERNANCE                          UPGRADE                           │
│ #3 SomeProp   VOTING   ends 2d 14h  none pending                     │
│ #4 ParamChng  DEPOSIT  ends 5d 01h                                   │
└──────────────────────────────────────────────────────────────────────┘
 r:refresh  q:quit
```

### Render function responsibilities

| File | Function | Draws |
|------|----------|-------|
| `header.go` | `renderHeader(s Snapshot)` | sentinel bar: height, block time, sync, peers, container, CPU%, RAM%, Disk% |
| `system.go` | `renderSystem(s Snapshot)` | OS: CPU, RAM, disk, load avg · Docker: status, restarts, uptime, container CPU+RAM |
| `consensus.go` | `renderConsensus(s Snapshot)` | block height, interval, sync, peers, validator counts, next proposer, slashing window params |
| `application.go` | `renderApplication(s Snapshot)` | supply, bonded ratio + bar, inflation, reward/blk, community pool, community tax, goal bonded, unbonding time |
| `evm.go` | `renderEVM(s Snapshot)` | chainId, EVM sync, txpool, base fee, gas used + bar, gas price, ERC20 pairs, IBC clients |
| `validators.go` | `renderValidators(s Snapshot)` | scrollable table: vp%, commission, missed blocks, outstanding rewards, commission earned, status |
| `governance.go` | `renderGovernance(s Snapshot)` | active proposals with status + countdown · upgrade plan name + target height |
| `view.go` | `View()` | assembles all of the above; lipgloss layout for 2×2 grid + table + strip |

### Color rules

| Condition | Element | Color |
|-----------|---------|-------|
| `catching_up: true` | header sync status | red |
| `catching_up: true` | consensus sync line | red |
| container not running | header + system panel | red |
| validator `jailed: true` | validator row | red |
| `missed_blocks > 500` | validator row | red |
| `missed_blocks > 100` | validator row | yellow |
| `tombstoned: true` | validator row | red + strikethrough |
| disk > 95% | disk line in header + system | red |
| disk > 85% | disk line in header + system | yellow |
| txpool pending > 50 | txpool line | yellow |
| upgrade plan present | governance strip upgrade | yellow |
| fetch error for a panel | that panel's content | dim + "unavailable: {err}" |

---

## Build

```bash
cd tools/ops/pmtop
go build -o pmtop .

# run on node (all defaults)
./pmtop

# run pointing at non-default ports
./pmtop --rpc http://localhost:36657 --container mynode
```

---

## What "done" looks like

- [ ] `pmtop` compiles to a single binary
- [ ] Launch → full dashboard renders within 3s on a live node
- [ ] Every data point from the query catalog above is visible somewhere in the layout
- [ ] `r` re-fetches all live data (params stay cached)
- [ ] Domain failure (e.g., Docker socket missing) shows "unavailable" in that panel; others still render
- [ ] `q` exits cleanly

---

## Explicitly out of scope (future layer)

- Auto-refresh ticker
- Time-series / sparklines
- Multi-node comparison
- Alerts / notifications
- Config file
