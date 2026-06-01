# evmd query — Full Feature Inventory

Sourced by running `docker compose exec evmd evmd query <module> --help` on node4.
Use this list to audit pmtmon coverage (what is monitored vs. what exists).

---

## How to run

```bash
# From node4
cd /home/ubuntu/evmd-docker
docker compose exec evmd evmd query <module> <subcommand> --home /data
```

---

## Modules and their query commands

### auth
| Command | What it returns |
|---------|----------------|
| `account <address>` | Account detail by address |
| `account-info <address>` | Common account info (sequence, number) |
| `accounts` | All accounts (paginated) |
| `address-by-acc-num <num>` | Address for an account number |
| `address-bytes-to-string` | Bytes → bech32 conversion — **not in scope** (general utility) |
| `address-string-to-bytes` | bech32 → bytes conversion — **not in scope** (general utility) |
| `bech32-prefix` | Chain bech32 prefix — **not in scope** (general utility) |
| `module-account <name>` | Single module account info |
| `module-accounts` | All module accounts |
| `params` | Auth module params (memo limit, sig limit) |

### authz — not in scope
| Command | What it returns |
|---------|----------------|
| `grants <granter> <grantee>` | Grants for a granter-grantee pair |
| `grants-by-grantee <grantee>` | All grants a grantee has received |
| `grants-by-granter <granter>` | All grants a granter has issued |

### bank
| Command | What it returns |
|---------|----------------|
| `balance <address> <denom>` | Single denom balance for an address |
| `balances <address>` | All balances for an address |
| `denom-metadata <denom>` | Coin metadata for a denom |
| `denom-metadata-by-query-string` | Coin metadata (query string variant) |
| `denom-owners <denom>` | All addresses holding a denom |
| `denom-owners-by-query` | denom-owners (query string variant) |
| `denoms-metadata` | Metadata for all registered denoms |
| `params` | Bank params (send enabled defaults) |
| `send-enabled` | Per-denom send-enabled entries |
| `spendable-balance <address> <denom>` | Spendable balance for one denom |
| `spendable-balances <address>` | All spendable balances for an address |
| `total-supply` | Total supply of all coins |
| `total-supply-of <denom>` | Total supply of one denom |

### consensus
| Command | What it returns |
|---------|----------------|
| `params` | Consensus parameters (block, evidence, validator) |
| `comet block-by-height <height>` | Committed block at a height |
| `comet block-latest` | Latest committed block |
| `comet node-info` | Current node info |
| `comet syncing` | Node sync status |
| `comet validator-set` | Latest validator set |
| `comet validator-set-by-height <height>` | Validator set at a height |

### distribution
| Command | What it returns |
|---------|----------------|
| `commission <valoper>` | Validator accumulated commission |
| `community-pool` | Coins in the community pool |
| `delegator-validators <delegator>` | Validators a delegator is delegated to |
| `delegator-withdraw-address <delegator>` | Withdraw address for a delegator |
| `params` | Distribution params (community tax, withdraw addr enabled) |
| `rewards <delegator>` | All delegation rewards for a delegator |
| `rewards-by-validator <delegator> <valoper>` | Rewards from a specific validator |
| `slashes <valoper> <start> <end>` | Slash events for a validator |
| `validator-distribution-info <valoper>` | Full distribution info for a validator |
| `validator-outstanding-rewards <valoper>` | Outstanding un-withdrawn rewards |

### erc20
| Command | What it returns |
|---------|----------------|
| `params` | ERC20 module params (enable_erc20) |
| `token-pair <token>` | Single registered ERC20↔Cosmos token pair |
| `token-pairs` | All registered token pairs |

### evidence — not in scope

### evm
| Command | What it returns |
|---------|----------------|
| `0x-to-bech32 <0x>` | bech32 address for a 0x address — **not in scope** (general utility) |
| `account <address>` | EVM account info (balance, code hash, nonce) |
| `balance-bank <0x> <denom>` | Bank balance for a 0x address |
| `balance-erc20 <0x> <erc20>` | ERC20 token balance for a 0x address |
| `bech32-to-0x <bech32>` | 0x address for a bech32 address — **not in scope** (general utility) |
| `code <address>` | Contract bytecode — **not in scope** (general utility) |
| `config` | EVM chain config (hardfork heights) |
| `params` | EVM module params (denom, EIPs, precompiles) |
| `storage <address> <key>` | Contract storage slot value — **not in scope** (general utility) |

### feegrant — not in scope

### feemarket
| Command | What it returns |
|---------|----------------|
| `base-fee` | EIP-1559 base fee at a given (or latest) block |
| `block-gas` | Gas used at a given (or latest) block |
| `params` | Feemarket params (no_base_fee, elasticity, change denominator, min gas) |

### gov
| Command | What it returns |
|---------|----------------|
| `constitution` | Current chain constitution text |
| `deposit <proposal-id> <depositor>` | Single deposit detail |
| `deposits <proposal-id>` | All deposits on a proposal |
| `params` | Governance params (voting period, quorum, threshold, deposit) |
| `proposal <id>` | Single proposal details |
| `proposals` | All proposals (filterable by status) |
| `tally <id>` | Current tally for a proposal |
| `vote <proposal-id> <voter>` | Single vote detail |
| `votes <proposal-id>` | All votes on a proposal |

### ibc — not in scope (all: client, channel, connection)

### ibc-transfer — not in scope

### mint
| Command | What it returns |
|---------|----------------|
| `annual-provisions` | Current annual provisions |
| `inflation` | Current inflation rate |
| `params` | Mint params (blocks/year, goal bonded, inflation bounds) |

### pmtrewards
| Command | What it returns |
|---------|----------------|
| `params` | PMT rewards params (enabled, reward/block, pool address) |

### precisebank — not in scope

### slashing
| Command | What it returns |
|---------|----------------|
| `params` | Slashing params (window, min signed, slash fractions) |
| `signing-info <cons-address>` | Signing info for a single validator |
| `signing-infos` | Signing info for all validators |

### staking
| Command | What it returns |
|---------|----------------|
| `delegation <delegator> <valoper>` | Single delegation record |
| `delegations <delegator>` | All delegations by a delegator |
| `delegations-to <valoper>` | All delegations to a validator |
| `delegator-validator <delegator> <valoper>` | Validator info for a delegation pair |
| `delegator-validators <delegator>` | All validators a delegator is delegated to |
| `historical-info <height>` | Historical staking info at a height |
| `params` | Staking params (unbonding time, max validators, bond denom) |
| `pool` | Bonded / not-bonded pool totals |
| `redelegation <delegator> <src> <dst>` | Redelegation record |
| `unbonding-delegation <delegator> <valoper>` | Unbonding delegation record |
| `unbonding-delegations <delegator>` | All unbonding delegations for a delegator |
| `unbonding-delegations-from <valoper>` | All unbonding delegations from a validator |
| `validator <valoper>` | Single validator info |
| `validators` | All validators (all statuses) |

### upgrade
| Command | What it returns |
|---------|----------------|
| `applied <upgrade-name>` | Block height where an upgrade was applied |
| `authority` | Upgrade authority address |
| `module-versions` | Consensus versions of all modules |
| `plan` | Current pending upgrade plan (name + height) |

---

## Top-level / CometBFT queries (not under a module)

| Command | What it returns |
|---------|----------------|
| `block` | Committed block by height, hash, or event |
| `block-results` | Block results (events, tx results) by height |
| `comet-validator-set` | Full CometBFT validator set at given height |
| `tx <hash>` | Single transaction by hash |
| `txs` | Paginated transactions matching a set of events |

---

## Not in scope for PMT mon

- **authz** — tx signing delegation, not ops state
- **feegrant** — fee allowances between accounts, not ops state
- **evidence** — misbehaviour records, not ops state
- **ibc** (client, channel, connection, ibc-transfer) — no active IBC on PMT currently
- **precisebank** — internal fractional accounting, not surfaced in ops
- **general utility endpoints** — address format conversions (`0x-to-bech32`, `bech32-to-0x`, `address-bytes-to-string`, etc.), contract bytecode/storage reads (`code`, `storage`), bech32 prefix
