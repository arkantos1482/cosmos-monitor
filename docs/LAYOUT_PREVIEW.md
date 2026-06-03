# pmtop layout preview (node4 sample)

**Status:** Implemented per decisions below.

| Decision | Choice |
|----------|--------|
| §3 title | **3. VALIDATOR SET** (unchanged) |
| §4 title | **4. THIS VALIDATOR** |
| P2P dial source | Chain RPC only (`/status`, `/net_info`) — no inventory/manifest |
| Connected peers | Under §3 as **P2P on this node** (`/net_info`) |

---

## Proposed output (example — node4)

Below is what the terminal / web markdown would look like after the change.  
Values in _italics_ are illustrative; bonded table rows would come from live REST/RPC.

---

# 1. INFRASTRUCTURE

_(unchanged)_

## OS

- **load**: 0.12 / 0.15 / 0.18  (1m 5m 15m)
- **ram**: 4.2 GB / 16.0 GB  (26%)
- **disk**: 120 GB / 500 GB  (24%)

## Container

- **status**: running
- **cpu**: 12%
- **ram**: 2.1 GB / 8.0 GB
- **restarts**: 0
- **uptime**: 12d 4h

---

# 2. NODE

_About **this machine's** CometBFT process — identity, listen addresses, and **this node's** consensus participation._

## Node

- **moniker**: node4
- **node ID**: `3381ddd6b06ec766400d3bdbddcfaaa2305f4984` _(CometBFT P2P peer ID)_
- **version**: 0.38.17
- **chain ID**: pmt
- **p2p listen**: `ec2-34-203-36-91.compute-1.amazonaws.com:26656` _(advertised dial address)_
- **rpc listen**: `tcp://0.0.0.0:26657` _(from `/status` node_info.other.rpc_address)_

## Consensus

_Block production view for **this** process (from `/status` sync_info + validator_info)._

- **sync**: synced
- **height**: 465,387
- **interval**: 5.01s
- **last block**: 4s ago
- **consensus address**: `870DA29817F423D56B6E1C81697E2CEA25662D9C` _(hex; signs blocks)_
- **voting power**: 100,000,000 _(consensus units from `/status`)_
- **mempool**: 0 pending

_(Removed from §2: connected peer monikers — moved to §3 as network-wide context.)_

---

# 3. CHAIN STATUS

_All validators on **pmt** — how to dial them on P2P, stake, signing health, and chain-wide staking/slashing params._

## Validators

| moniker | operator | p2p dial | node ID | consensus | vp% | commission | missed | status | jailed | local |
|---------|----------|----------|---------|-----------|-----|------------|--------|--------|--------|-------|
| node1 | cosmosvaloper1akk…zp4 | `7c90c689…@ec2-18-206-235-38…:26656` | 7c90c689… | 31AEC3D5… | 25.0% | 10.0% | 0 | BONDED | | |
| node2 | cosmosvaloper1r2d…jk0a | `d8e831a5…@ec2-3-88-102-95…:26656` | d8e831a5… | B022E73E… | 25.0% | 10.0% | 0 | BONDED | | |
| node3 | cosmosvaloper15hr…ej3e | `b00a33cb…@ec2-98-88-84-176…:26656` | b00a33cb… | 74F9309C… | 25.0% | 10.0% | 0 | BONDED | | |
| node4 | cosmosvaloper1vmr…pdxh | `3381ddd6…@ec2-34-203-36-91…:26656` | 3381ddd6… | 870DA298… | 25.0% | 10.0% | 0 | BONDED | | **this node** |

**p2p dial** = `node_id@external_host:26656` from deploy inventory (`tools/ops/deploy/inventory.ini` `external_address`), keyed by moniker ↔ manifest `node_id`.

**connected** _(optional badge on row)_: whether this validator's node ID appears in **this** node's `/net_info` peers (node4 example: 3 peers — node1, node2, node3).

## Connected peers (this node)

- **count**: 3
- **monikers**: node1, node2, node3  
  _(what §2 "P2P" showed today — kept here as "who we are actually peered with", separate from static dial strings)_

## Summary

- **bonded**: 4
- **jailed**: 0
- **tombstoned**: 0
- **below min signed**: 0
- **next proposer**: node2

## Staking pool

_(unchanged — bond denom, supply, bonded %, unbonding time, max validators)_

## Slashing params

_(unchanged — signed blocks window, min signed %, slash fractions)_

---

# 4. THIS VALIDATOR

_Staking & rewards for **this machine's** operator — matched via consensus address from §2._

## Operator

- **operator address**: `cosmosvaloper1vmr9wxpldngnh0tvpr8h2pk2aycts3v7z8pdxh` _(primary identity for CLI / rewards)_
- **moniker**: node4

## Staking

- **status**: BONDED
- **voting power**: 100,000,000,000,000,000,000,000,000 apmt  (25.0% of bonded stake)
- **commission**: 10.0%

## Block signing

- **signing health**: OK
- **missed / window**: 0 / 10,000 blocks  (max allowed: 500)
- **proposer**: not next  _(next: node2)_

## Unclaimed rewards

- **outstanding rewards**: …
- **commission earned**: …

_(Removed from §4: node ID, consensus address — already under §2 Node / Consensus.)_

---

# 5–7. ECONOMICS, GOVERNANCE, EVM JSON-RPC

_(no change in this pass)_

---

## Implementation notes (for later)

| Field | Source |
|-------|--------|
| Node / listen / rpc | `GET /status` |
| Consensus block stats | `GET /status` sync_info + block interval RPC |
| Consensus addr / power | `GET /status` validator_info |
| Validator table (stake, missed) | existing REST + slashing |
| **p2p dial per validator** | static map: moniker → `node_id` (manifest) + host:port (inventory `external_address`); ship as embedded JSON or `-validators-config` flag |
| Connected peer monikers | existing `net_info` (stays under §3) |

**Makefile:** add something like `make snapshot` (capture full pane, e.g. `-S -500`) so `deploy` smoke tests include §2–§4 — optional follow-up.

---

## Please confirm

1. Section title **"3. CHAIN STATUS"** vs **"3. NETWORK"** vs **"3. VALIDATOR SET"** (with subtitle)?
2. Keep **Connected peers (this node)** under §3, or drop if the per-validator dial table is enough?
3. **§4** title stay **THIS VALIDATOR** or shorten to **VALIDATOR**?
4. OK to load P2P dial strings from deploy `inventory.ini` + `manifest.json` (read-only at startup), or prefer fetching only from chain APIs?
