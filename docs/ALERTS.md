# pmtop Telegram Alerts

pmtop can send Telegram messages when dashboard **sections** show warn or bad conditions. The global status strip is not evaluated separately. Each node that runs `-alert` pages independently (no cross-node dedupe).

## Setup

1. Create a Telegram bot via [@BotFather](https://t.me/BotFather) and note the bot token.
2. Add the bot to a chat or channel and obtain the chat ID.
3. Store credentials in `tools/ops/deploy/secrets.yml` (gitignored):

```yaml
telegram_bot_token: "123456:ABC..."
telegram_chat_id: "-1001234567890"
```

4. Start pmtop with `-alert` (or use `make remote-start-alert` on node4).

## Flags and environment

| Flag | Env fallback | Default | Purpose |
|------|--------------|---------|---------|
| `-alert` | — | off | Enable alert polling |
| `-alert-interval` | `PMTOP_ALERT_INTERVAL` | `30s` | Poll interval (independent of 5s HTMX UI poll) |
| `-alert-dry-run` | — | off | Log messages without POSTing to Telegram |
| `-node-name` | `PMTOP_NODE_NAME` | hostname | Label in alert text |
| — | `PMTOP_TELEGRAM_TOKEN` | — | Bot token (required unless dry-run) |
| — | `PMTOP_TELEGRAM_CHAT_ID` | — | Target chat (required unless dry-run) |
| — | `PMTOP_ALERT_COOLDOWN` | `15m` | Per-finding cooldown before repeat/recovery |

Web UI (`-web :7777`) and alerts can run together. Set `-web ""` for alert-only mode.

## Alert policy

| Section | What pages Telegram | Notes |
|---------|---------------------|-------|
| Infrastructure | container stopped, OOM, restarts ≥3, disk/RAM pressure | as-today |
| EVM JSON-RPC | **liveness only**: `eth_blockNumber` failure (RPC DOWN) after **2 consecutive alert ticks** (~60s); `net_listening` false **only if that RPC returned a value** | Degraded probes, syncing, block age stay **UI-only**. Failed blockNumber must not imply “not listening”. Single-poll blips do not page. |
| Validator | catching up | as-today |
| Staking | local jailed, tombstoned, missed blocks high | as-today |
| Slashing | local + network slashing KPIs, headroom low | as-today |
| Fee market | base fee rising, disabled/pending | as-today |
| Rewards | **transitions only**: enable↔disable (when params fetch healthy); pool empty↔refill | No steady-state “still disabled / still empty / inflation off”. Missing PMT params (REST off) is **not** “PMT disabled”. |
| Distribution | escrow bank ≠ tracked unclaimed total | as-today |

Evaluation lives in `internal/panel/findings.go` (plus rewards seeding in `internal/alert/engine.go`).

## Message format

Active alert — **only newly fired or worsened findings this tick** (not a dump of all active):

```
🚨 PMT — node4

Infrastructure · bad
  • container stopped

Height 1842032 · 12s ago
```

Recovery when a condition clears (per finding):

```
✅ PMT — node4 resolved

Infrastructure · container stopped
```

Cooldown (default 15m) applies per section+key for both alerts and recoveries. While recovery is cooldown-blocked, the finding stays **sticky** in the engine so a brief clear cannot re-alert after cooldown (avoids repeating RPC DOWN with no ✅). Expect up to 4× duplicate messages across nodes when the same chain event is seen on every validator.

## node4 dev workflow

From `tools/ops/pmtop/`:

```bash
# Build + dev UI on :7778 (existing)
make remote-dev-release

# Optional dev alerts on :7778 (prod Docker :7777 uses Ansible)
make remote-start-alert

# Dry-run alerts without Telegram POST
make remote-start-alert REMOTE_PMTOP_FLAGS="-alert-dry-run"

# Combine with other flags
make remote-start-alert REMOTE_PMTOP_FLAGS="-show-sources"
```

`REMOTE_PMTOP_FLAGS` is appended to the remote `pmtop` command for any `remote-start*` target.

Local dry-run:

```bash
go run ./cmd/pmtop -alert -alert-dry-run -web ""
```

## Tests

```bash
go test ./internal/alert/... ./internal/panel/... ./internal/fetch/... -run 'Findings|Engine|Format|FetchEVM'
```
