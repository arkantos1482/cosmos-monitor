# pmtop — API Reference

All endpoints are polled from the node itself (localhost). Endpoints that return non-2xx or 501 are silently skipped; the UI shows "N/A" or "?" for the affected field.

---

## CometBFT RPC — `:26657` (GET, JSON)

| Endpoint | Used for | Key fields |
|----------|----------|------------|
| `GET /status` | node identity, height, sync | `result.node_info.{id,moniker,version}`, `result.sync_info.{latest_block_height,latest_block_time,catching_up}` |
| `GET /net_info` | peer count | `result.n_peers` |
| `GET /block` | latest block time | `result.block.header.{height,time}` |
| `GET /block?height={h-1}` | previous block time (compute interval) | same |
| `GET /validators?per_page=100` | validator set (proposer priority) | `result.validators[].{address,proposer_priority,pub_key.value}` |

`pub_key.value` is a base64-encoded raw ed25519 key. SHA256 the bytes → take first 20 bytes → hex = consensus address.

---

## Cosmos REST LCD — `:1317` (GET, JSON)

### Standard Cosmos SDK modules

| Endpoint | Used for | Key fields |
|----------|----------|------------|
| `GET /cosmos/staking/v1beta1/validators?status=BOND_STATUS_BONDED&pagination.limit=100` | bonded validators | `validators[].{operator_address,description.moniker,status,tokens,commission.commission_rates.rate,jailed,consensus_pubkey.key}` |
| `GET /cosmos/staking/v1beta1/validators?status=BOND_STATUS_UNBONDING&...` | unbonding validators | same |
| `GET /cosmos/staking/v1beta1/validators?status=BOND_STATUS_UNBONDED&...` | unbonded validators | same |
| `GET /cosmos/staking/v1beta1/pool` | bonded/not-bonded totals | `pool.{bonded_tokens,not_bonded_tokens}` |
| `GET /cosmos/staking/v1beta1/params` | staking params | `params.{unbonding_time,max_validators,bond_denom}` |
| `GET /cosmos/slashing/v1beta1/signing_infos?pagination.limit=100` | missed blocks, tombstone | `info[].{address,missed_blocks_counter,tombstoned}` |
| `GET /cosmos/slashing/v1beta1/params` | slashing params | `params.{signed_blocks_window,min_signed_per_window}` |
| `GET /cosmos/bank/v1beta1/supply` | total token supply | `supply[].{denom,amount}` |
| `GET /cosmos/mint/v1beta1/inflation` | current inflation rate | `inflation` (decimal string) |
| `GET /cosmos/mint/v1beta1/params` | mint params | `params.{blocks_per_year,goal_bonded}` |
| `GET /cosmos/distribution/v1beta1/community_pool` | community pool | `pool[].{denom,amount}` |
| `GET /cosmos/distribution/v1beta1/params` | community tax | `params.community_tax` |
| `GET /cosmos/distribution/v1beta1/validators/{valoper}/outstanding_rewards` | per-validator rewards | `rewards.rewards[].{denom,amount}` |
| `GET /cosmos/distribution/v1beta1/validators/{valoper}/commission` | per-validator commission | `commission.commission[].{denom,amount}` |
| `GET /cosmos/gov/v1beta1/proposals?proposal_status=2` | voting-period proposals | `proposals[].{proposal_id,content.title,status,voting_end_time,deposit_end_time}` |
| `GET /cosmos/gov/v1beta1/proposals?proposal_status=1` | deposit-period proposals | same |
| `GET /cosmos/gov/v1/proposals?proposal_status=2` | (v1 fallback) | `proposals[].{id,title,status,voting_end_time,deposit_end_time}` |
| `GET /cosmos/gov/v1beta1/params/voting` | voting period | `voting_params.voting_period` |
| `GET /cosmos/gov/v1beta1/params/tallying` | quorum, threshold | `tally_params.{quorum,threshold}` |
| `GET /cosmos/upgrade/v1beta1/current_plan` | pending upgrade | `plan.{name,height}` (null if none) |
| `GET /cosmos/auth/v1beta1/params` | auth params (unused currently) | `params.{max_memo_characters,tx_sig_limit}` |
| `GET /ibc/core/client/v1/client_states` | IBC client count | `client_states[]` (count only) |

### cosmos-evm custom modules

> **These replaced the old evmos/ethermint routes which return 501 on this chain.**

#### x/feemarket — `/cosmos/evm/feemarket/v1/`

| Endpoint | Used for | Key fields |
|----------|----------|------------|
| `GET /cosmos/evm/feemarket/v1/base_fee` | EIP-1559 base fee of last block | `base_fee` (decimal string) |
| `GET /cosmos/evm/feemarket/v1/block_gas` | stored W (block gas wanted); fallback when `block_results` lacks event | `gas` (int64 string) |
| `GET /cosmos/evm/feemarket/v1/params` | feemarket params | `params.{no_base_fee,base_fee_change_denominator,elasticity_multiplier,base_fee,min_gas_price,min_gas_multiplier}` |

#### x/vm (EVM) — `/cosmos/evm/vm/v1/`

| Endpoint | Used for | Key fields |
|----------|----------|------------|
| `GET /cosmos/evm/vm/v1/params` | VM/EVM module params | `params.{evm_denom,extra_eips,active_static_precompiles,history_serve_window}` |
| `GET /cosmos/evm/vm/v1/base_fee` | base fee (same as feemarket but checks London hardfork status) | `base_fee` (math.Int) |
| `GET /cosmos/evm/vm/v1/min_gas_price` | global min gas price (converted to 18 decimal EVM precision) | `min_gas_price` (math.Int) |
| `GET /cosmos/evm/vm/v1/config` | EVM chain config (hardfork heights, etc.) | `config` (ChainConfig object) |
| `GET /cosmos/evm/vm/v1/account/{address}` | Ethereum account (EVM-facing view) | `balance`, `code_hash`, `nonce` |
| `GET /cosmos/evm/vm/v1/cosmos_account/{address}` | Cosmos account info for an Ethereum address | `cosmos_address`, `sequence`, `account_number` |
| `GET /cosmos/evm/vm/v1/validator_account/{cons_address}` | Cosmos account info for a validator consensus address | `account_address`, `sequence`, `account_number` |
| `GET /cosmos/evm/vm/v1/balances/{address}` | EVM denom balance | `balance` |
| `GET /cosmos/evm/vm/v1/storage/{address}/{key}` | contract storage slot value | `value` (hex hash) |
| `GET /cosmos/evm/vm/v1/codes/{address}` | contract bytecode | `code` (bytes) |
| `GET /cosmos/evm/vm/v1/eth_call` | simulate an Ethereum call (debug/read-only) | `hash`, `logs`, `ret`, `vm_error`, `gas_used` |
| `GET /cosmos/evm/vm/v1/estimate_gas` | estimate gas for a transaction | `gas`, `ret`, `vm_error` |
| `GET /cosmos/evm/vm/v1/trace_tx` | debug_traceTransaction | `data` (serialized bytes) |
| `GET /cosmos/evm/vm/v1/trace_block` | debug_traceBlockByNumber / debug_traceBlockByHash | `data` (serialized bytes) |
| `GET /cosmos/evm/vm/v1/trace_call` | debug_traceCall | `data` (serialized bytes) |

#### x/precisebank — `/cosmos/evm/precisebank/v1/`

> Tracks sub-integer (fractional) token amounts for EVM ↔ Cosmos denomination conversion at 18-decimal precision.

| Endpoint | Used for | Key fields |
|----------|----------|------------|
| `GET /cosmos/evm/precisebank/v1/remainder` | amount held in reserve but not yet in circulation (fractional remainder) | `remainder` (Coin: `denom`, `amount`) |
| `GET /cosmos/evm/precisebank/v1/fractional_balance/{address}` | fractional (sub-integer) balance of an address | `fractional_balance` (Coin: `denom`, `amount`) |

#### x/pmtrewards — `/cosmos/evm/pmtrewards/v1/`

> PMT-specific rewards module parameters.

| Endpoint | Used for | Key fields |
|----------|----------|------------|
| `GET /cosmos/evm/pmtrewards/v1/params` | PMT rewards module params | `params` (Params object) |

#### x/erc20 — `/cosmos/evm/erc20/v1/`

| Endpoint | Used for | Key fields |
|----------|----------|------------|
| `GET /cosmos/evm/erc20/v1/token_pairs` | registered ERC20↔Cosmos bridges | `token_pairs[].{denom,erc20_address,enabled}`, `pagination` |
| `GET /cosmos/evm/erc20/v1/token_pairs/{token}` | single token pair | `token_pair.{denom,erc20_address,enabled}` |
| `GET /cosmos/evm/erc20/v1/params` | ERC20 module params | `params.{enable_erc20,...}` |

---

## EVM JSON-RPC — `:8545` (POST, JSON-RPC 2.0)

All calls use `POST` with body `{"jsonrpc":"2.0","method":"<method>","params":[],"id":1}`.

| Method | Used for | Result type |
|--------|----------|-------------|
| `eth_blockNumber` | latest EVM block (sanity / EVM sync) | hex string |
| `eth_chainId` | EVM chain ID | hex string |
| `eth_syncing` | sync status | `false` or object |
| `eth_gasPrice` | current gas price (wei) | hex string |
| `txpool_status` | pending/queued tx counts | `{pending: hex, queued: hex}` |
| `net_peerCount` | EVM peer count | hex string |

---

## Docker socket — `unix:///var/run/docker.sock` (GET, HTTP)

Base URL: `http://localhost` (dialed over the socket).

| Endpoint | Used for | Key fields |
|----------|----------|------------|
| `GET /containers/{name}/stats?stream=false` | CPU%, RAM usage/limit | `cpu_stats.cpu_usage.total_usage`, `precpu_stats.cpu_usage.total_usage`, `cpu_stats.system_cpu_usage`, `precpu_stats.system_cpu_usage`, `cpu_stats.online_cpus`, `memory_stats.{usage,limit}` |
| `GET /containers/{name}/json` | running state, restarts, uptime | `State.{Running,StartedAt}`, `RestartCount` |

CPU% formula:
```
cpuDelta = cpu_stats.cpu_usage.total_usage - precpu_stats.cpu_usage.total_usage
sysDelta  = cpu_stats.system_cpu_usage      - precpu_stats.system_cpu_usage
cpuPercent = (cpuDelta / sysDelta) * numCPUs * 100
```

---

## /proc filesystem — local reads (no HTTP)

| File | Used for | Fields parsed |
|------|----------|---------------|
| `/proc/loadavg` | load averages | columns 1, 2, 3 (1m, 5m, 15m) |
| `/proc/meminfo` | RAM | `MemTotal`, `MemAvailable` (kB) |

Disk: `syscall.Statfs("/")` → `Blocks`, `Bfree`, `Bsize` → total and used bytes.

---

## Broken / 501 routes on this chain (do not use)

These were evmos-era routes that cosmos-evm does not implement:

| Old route | Replaced by |
|-----------|-------------|
| `/ethermint/feemarket/v1/base_fee` | `/cosmos/evm/feemarket/v1/base_fee` |
| `/ethermint/feemarket/v1/block_gas` | `/cosmos/evm/feemarket/v1/block_gas` |
| `/ethermint/feemarket/v1/params` | `/cosmos/evm/feemarket/v1/params` |
| `/ethermint/evm/v1/params` | `/cosmos/evm/vm/v1/params` |
| `/evmos/erc20/v1/token_pairs` | `/cosmos/evm/erc20/v1/token_pairs` |
| `/evmos/erc20/v1/params` | `/cosmos/evm/erc20/v1/params` |
