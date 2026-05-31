# PMT Tokenomics Reference

## Token

| Field | Value |
|-------|-------|
| Name | PMT |
| Base denom | `apmt` (atto-PMT, 1 PMT = 10¹⁸ apmt) |
| Total supply | 2,000,000,000 PMT (fixed) |
| Inflation | 0% — mint module is present but inactive |

---

## Staking

| Field | Value |
|-------|-------|
| Bonded | 400M PMT (20% of supply) |
| Goal bonded | 67% |
| Validators | 4 (equal stake, 25% VP each) |
| Commission | 10% per validator |
| Unbonding time | 21 days |
| Bond denom | `apmt` |

---

## Fee Distribution Model

The chain uses the standard Cosmos `x/distribution` module with these parameters:

| Field | Value | Implication |
|-------|-------|-------------|
| Community tax | 0% | No community pool cut |
| Community pool | 0 PMT | Never accumulates |

**Result:** 100% of all transaction fees flow to validators and their delegators,
split proportional to voting power. Validators take their commission (10%) first;
the remainder goes to delegators.

### Per-validator fee revenue

With 4 equal validators, each receives 25% of each block's tx fees.
Currently tx volume is minimal, so outstanding fee rewards per validator
are in the micro-PMT range.

---

## PMT Rewards (custom `x/pmtrewards` module)

This is a non-standard module that distributes a fixed amount of PMT per block
from a dedicated pool account, independent of tx volume or inflation.

| Field | Value |
|-------|-------|
| Rate | 0.1 PMT / block |
| Annual rate | ~631,152 PMT/year (at 6,311,520 blocks/year) |
| Daily rate | ~1,662 PMT/day (at ~5.2s block time) |
| Source | Pool account: `cosmos13zwnumfrgpplakmr5n9xmuw7sqp2tr3ppfzs6k` |
| Distribution | Proportional to validator voting power |
| Commission | Validator commission (10%) applies to PMT rewards too |

### States

| State | Condition | Meaning |
|-------|-----------|---------|
| **disabled** | `enabled = false` | Module turned off by governance |
| **ENABLED — pool EMPTY** | `enabled = true`, pool balance = 0 | Configured but no PMT to distribute — validators receive **zero** PMT rewards |
| **distributing** | `enabled = true`, pool balance > 0 | Actively paying out; validators receive PMT each block |

### Pool lifetime estimate

```
days_remaining = pool_balance / (0.1 PMT/block × blocks_per_day)
blocks_per_day ≈ 86400s / block_interval
```

Example: 1M PMT in pool → ~602 days at 5.2s block time.

---

## Validator Economics Summary

A validator's total revenue per block:

```
revenue = (tx_fees_in_block × VP%) + (pmt_reward_per_block × VP%)
validator_cut = revenue × commission_rate
delegator_cut = revenue × (1 - commission_rate)
```

**Outstanding rewards** (`distribution.outstanding_rewards`): accumulated but unclaimed
fees + PMT rewards. Validators claim these on-demand via `MsgWithdrawValidatorCommission`
/ `MsgWithdrawDelegatorReward`.

**Commission earned** (`distribution.validators/{addr}/commission`): the validator's
own commission portion, separate from delegator rewards.

When the PMT pool is empty, outstanding rewards reflect **tx fees only**.

---

## Alert Conditions

| Condition | Severity | What to check |
|-----------|----------|---------------|
| PMT pool empty | 🔴 HIGH | Fund `cosmos13zwnumfrgpplakmr5n9xmuw7sqp2tr3ppfzs6k` |
| Pool < 30 days runway | 🟡 WARN | Plan a top-up governance tx |
| Validator jailed | 🔴 HIGH | Check signing, restart node if needed |
| Validator tombstoned | 🔴 CRITICAL | Permanent — must replace validator |
| MissedBlocks > threshold | 🟡 WARN | Network/node instability |
| Container stopped | 🔴 HIGH | `docker start evmd-node` on the relevant node |
| CatchingUp = true | 🟡 WARN | Node is replaying blocks, not yet live |
| Bonded < 4/4 | 🔴 HIGH | One or more validators unbonded or jailed |
