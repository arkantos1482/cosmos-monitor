# cosmos-monitor — Dev & Bugfix Guide

## Repo

- GitHub: https://github.com/arkantos1482/cosmos-monitor
- Local source: `tools/ops/pmtop/` inside `cosmos-evm-old`
- Node4 source: `~/cosmos-monitor/` (cloned from GitHub)

## Dev workflow

Edit locally, commit, push, then pull and rebuild on node4:

```bash
# 1. Make changes locally (in cosmos-evm-old/tools/ops/pmtop/)
# 2. Commit and push
git add -A && git commit -m "fix: description" && git push

# 3. Pull, build, and run on node4 in one command
ssh -t node4 'cd ~/cosmos-monitor && git pull && go build -o ~/pmtop . && ~/pmtop'
```

Go is installed at `/usr/local/go/bin/go` on node4 and is already on `$PATH` via `~/.bashrc`.
The SSH alias `node4` is configured in `~/.ssh/config`.

## Source map

```
main.go               — flag parsing, tea.NewProgram
fetch/client.go       — shared http.Client, doJSON, newDockerClient
fetch/chain.go        — CometBFT RPC + Cosmos REST; FetchChain + FetchParams
fetch/evm.go          — EVM JSON-RPC (eth_blockNumber, eth_chainId, etc.)
fetch/system.go       — /proc/loadavg, /proc/meminfo, syscall.Statfs
fetch/docker.go       — Docker socket stats + inspect; CPU% formula
ui/model.go           — bubbletea Model, Config, message types
ui/update.go          — Update(); fetchAllCmd (4 concurrent goroutines)
ui/view.go            — View(); assembles header + 2×2 grid + validator table + governance strip
ui/header.go          — renderHeader(): sentinel bar
ui/system.go          — renderSystem(): SYSTEM panel
ui/consensus.go       — renderConsensus(): CONSENSUS panel
ui/validators.go      — renderValidators(): full-width table
ui/evm.go             — renderEVM(): EVM panel
ui/application.go     — renderApplication(): APPLICATION panel
ui/governance.go      — renderGovernance(): proposals + upgrade strip
ui/style.go           — lipgloss styles, colour helpers
```

## Data sources on node4

All confirmed reachable from the host (not inside Docker):

| Source | Address |
|--------|---------|
| CometBFT RPC | `localhost:26657` |
| Cosmos REST | `localhost:1317` |
| EVM JSON-RPC | `localhost:8545` |
| Docker socket | `/var/run/docker.sock` |

## Known issues / to-do

- **Refresh loop**: the TUI currently has no auto-refresh ticker — it only fetches on startup and on `r` keypress. Add a `tea.Tick` every 5s in `Init()`.
- **Validator missed blocks threshold**: the warning threshold (>100) and error threshold (>500) are hardcoded in `ui/validators.go`; consider deriving from `SignedBlocksWindow` param.
- **Block gas display**: `snap.BlockGas` is fetched but not shown in any panel — either display it in CONSENSUS or drop the field.
- **EVM syncing field**: `eth_syncing` returns `false` when synced and a struct when syncing; the current code casts both correctly but worth verifying on a catching-up node.
- **Terminal width edge cases**: panels can overflow or wrap on narrow terminals; minimum width guard is not enforced.

## Node4 quick reference

```bash
# Connect
ssh node4

# Run monitor
~/pmtop

# Rebuild after pull
cd ~/cosmos-monitor && git pull && go build -o ~/pmtop . && ~/pmtop

# Check validator container
docker logs -f evmd-node

# REST API
curl localhost:1317/cosmos/staking/v1beta1/validators?status=BOND_STATUS_BONDED

# RPC
curl localhost:26657/status
```
