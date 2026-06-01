# pmtop — Information Architecture

Display order top to bottom. Items marked `[gap]` are not yet fetched and need new fetch code.

---

## 1. Health

### OS
| Field | Detail |
|-------|--------|
| load | 1m / 5m / 15m averages |
| ram | used / total |
| disk | used / total |

### Container
| Field | Detail |
|-------|--------|
| status | running / stopped |
| cpu | % usage |
| ram | used / limit |
| restarts | count |
| uptime | duration since StartedAt |

---

## 2. Node

### Identity
| Field | Detail |
|-------|--------|
| node ID | CometBFT peer ID |
| moniker | human name |
| version | app version string |

### Block
| Field | Detail |
|-------|--------|
| height | latest block height |
| interval | time between last two blocks |
| time since last block | now − latest block time |

### Sync
| Field | Detail |
|-------|--------|
| status | synced / CATCHING UP |
| latest block time | UTC timestamp |

### Peers
| Field | Detail |
|-------|--------|
| cosmos peers | CometBFT p2p count |
| evm peers | EVM JSON-RPC net_peerCount |

---

## 3. Tokenomics

### Supply
| Field | Detail |
|-------|--------|
| total supply | all tokens in existence |
| bonded | tokens staked |
| not bonded | tokens not staked |
| bonded % | bonded / total |

### Inflation
| Field | Detail |
|-------|--------|
| rate | current inflation % |
| goal bonded % | target bonded ratio |
| blocks / year | used for inflation calc |
| annual provisions | absolute new tokens / year `[gap]` |

### Staking Params
| Field | Detail |
|-------|--------|
| unbonding time | duration |
| max validators | ceiling on active set |
| bond denom | on-chain staking denom |

### Distribution
| Field | Detail |
|-------|--------|
| community pool | balance |
| community tax | % of block rewards |

### PMT Rewards
| Field | Detail |
|-------|--------|
| enabled | yes / no |
| reward / block | amount + denom |
| pool address | source address |

---

## 4. EVM

### Identity
| Field | Detail |
|-------|--------|
| chain ID | EVM numeric chain ID |
| denom | EVM-facing denom (e.g. apmt) |

### Block
| Field | Detail |
|-------|--------|
| block | latest EVM block number |
| sync | synced / syncing |

### Gas
| Field | Detail |
|-------|--------|
| base fee | EIP-1559 base fee (wei) |
| gas price | current gas price |
| min gas price | floor from feemarket params |

### Fee Market
| Field | Detail |
|-------|--------|
| elasticity | elasticity multiplier |
| change denominator | base fee change denominator |
| no_base_fee | flag (disables EIP-1559) |

### Txpool
| Field | Detail |
|-------|--------|
| pending | pending tx count |
| queued | queued tx count |

### Precompiles
| Field | Detail |
|-------|--------|
| active | count of active static precompiles |
| addresses | list of precompile addresses |

### Config
| Field | Detail |
|-------|--------|
| history serve window | blocks of EVM history served |
| ERC20 enabled | yes / no |
| hardfork heights | London, Shanghai, etc. activation heights `[gap]` |

---

## 5. Token Pairs

Registered ERC20 ↔ Cosmos bridges, one row each:

| Field | Detail |
|-------|--------|
| denom | Cosmos denom |
| ERC20 address | 0x contract address |
| enabled | yes / no |

---

## 6. Validators

Sorted by voting power % descending. One row per validator:

| Field | Detail |
|-------|--------|
| moniker | display name |
| VP % | share of total bonded power |
| commission | rate % |
| earned | accumulated commission (undistributed) |
| outstanding | total undistributed rewards |
| missed | missed blocks counter |
| tombstoned | yes / YES (highlighted) |
| status | bonded / unbonding / unbonded |
| jailed | appended to status if true |

---

## 7. Security

### Slashing Params
| Field | Detail |
|-------|--------|
| window | signed blocks window (blocks) |
| min signed | minimum signed % |
| slash fraction downtime | `[gap]` |
| slash fraction double sign | `[gap]` |

### Summary
| Field | Detail |
|-------|--------|
| tombstoned count | validators with tombstoned = true |
| below threshold | validators currently below min signed % `[gap]` |

---

## 8. Governance

### Params
| Field | Detail |
|-------|--------|
| voting period | duration |
| quorum | minimum participation % |
| threshold | pass threshold % |
| veto threshold | veto threshold % `[gap]` |

### Active Proposals (voting period)
One row per proposal:

| Field | Detail |
|-------|--------|
| ID | proposal ID |
| title | proposal title |
| status | VOTING_PERIOD / DEPOSIT_PERIOD |
| voting ends | date |
| tally | yes / no / veto / abstain counts `[gap]` |

### Deposit-Period Proposals
Same fields, deposit_end shown instead of voting_end.

---

## 9. Upgrade

| Field | Detail |
|-------|--------|
| name | upgrade name |
| target height | block height it triggers at |
| blocks remaining | target − current height |

No pending upgrade → show "none".

---

## 10. IBC

| Field | Detail |
|-------|--------|
| active clients | count of registered IBC light clients |
