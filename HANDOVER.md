# pmtop — implementation handover

Single-node Cosmos SDK + EVM chain monitoring tool.
Run it on the node (`~/pmtop`), get a full picture across system, chain, EVM, and economics.

---

## Output structure

Sections printed top to bottom:

```
NODE          — node ID, version, height, block interval, sync, peer counts
SYSTEM        — load averages, RAM, disk
CONTAINER     — status, CPU%, RAM, restarts, uptime
STAKING       — supply, bonded%, inflation, unbonding, denom
DISTRIBUTION  — community pool, community tax
PMT REWARDS   — enabled, reward/block, pool address
FEEMARKET     — base fee, block gas, min gas, elasticity, change denominator
EVM           — chain ID, denom, block, sync, gas price, txpool, erc20, precompiles
TOKEN PAIRS   — registered ERC20↔Cosmos bridges
IBC           — client count
PRECISEBANK   — fractional remainder
SLASHING      — sign window, min signed, tombstoned count
GOVERNANCE    — voting period, quorum, threshold, active proposals
UPGRADE       — pending upgrade name + height
VALIDATORS    — per-validator: vp%, commission, missed, tombstoned, rewards, earned, status
```

---

## Locked decisions

| Decision | Choice | Reason |
|----------|--------|--------|
| Where it runs | **On the node** (`ssh -t node4 ~/pmtop`) | direct /proc + Docker socket access |
| Output | **Plain text to stdout** | no TUI framework, terminal scrolls naturally |
| Refresh | **Manual only** — `r` to re-fetch, `q` to quit | no ticker |
| Config | **CLI flags only** | no config file |
| Transport | **Plain HTTP everywhere** | one pattern; Docker socket is also HTTP over Unix |
| Language | **Go** | single binary, no runtime deps |
| Terminal | **raw mode via `golang.org/x/term`** | single-keypress r/q without Enter |

---

## CLI flags

```
pmtop [flags]

-rpc        string   CometBFT RPC endpoint   (default: http://localhost:26657)
-rest       string   Cosmos REST/LCD endpoint (default: http://localhost:1317)
-evm        string   EVM JSON-RPC endpoint    (default: http://localhost:8545)
-container  string   Docker container name    (default: evmd-node)
```

---

## File structure

```
tools/ops/pmtop/
├── main.go        parse flags → concurrent fetch → print loop (r/q)
├── print.go       printAll() — all sections as plain fmt.Printf
│
└── fetch/         pure functions, zero state, organised by data source
    ├── client.go  shared http.Client + doJSON / doDockerJSON
    ├── chain.go   CometBFT :26657 + Cosmos REST :1317 — all chain data + params
    ├── evm.go     EVM JSON-RPC :8545
    ├── system.go  /proc/loadavg, /proc/meminfo, syscall.Statfs
    └── docker.go  Docker socket → container stats + inspect
```

**Invariant**: `fetch/` = pure functions, no state, organised by data source.
`print.go` = pure render from snapshot. Adding a new metric: 1 field in the right
`fetch/*.go` struct, 1 fetch call, 1 `row()` line in `print.go`. Nothing else changes.

---

## Dev loop

Edit in `tools/ops/pmtop/`, then from that directory:

```bash
git add -A && git commit -m "..." && make push-deploy
```

`make push-deploy` pushes and SSHes into node4 to pull, rebuild, and smoke-test.
To run interactively on node4:

```bash
ssh -t -i ~/.ssh/pmt-nodes.pem ubuntu@<node4-host> '~/pmtop'
```

Or from `tools/ops/deploy/`:

```bash
make pmtop-node4
```
