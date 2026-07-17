# pmtop Telegram Alerts

pmtop can send Telegram messages when dashboard **sections** show warn or bad conditions — the same logic as each section's summary card badges. The global status strip is not evaluated separately.

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

## Evaluated sections

| Section | Examples |
|---------|----------|
| Infrastructure | container stopped, OOM, restarts ≥3, disk/RAM pressure |
| EVM JSON-RPC | RPC down/degraded, block age slow/stale, not listening, syncing |
| Validator | catching up |
| Staking | local jailed, tombstoned, missed blocks high |
| Slashing | local + network slashing KPIs, headroom low |
| Fee market | base fee rising, disabled/pending |
| Rewards | PMT disabled/not emitting, inflation off |
| Distribution | escrow bank ≠ tracked unclaimed total |

Evaluation lives in `internal/panel/findings.go` and reuses existing domain helpers.

## Message format

Active alert (one batched message per tick when something new or worsened fires):

```
🚨 PMT — node4

Infrastructure · bad
  • container stopped

Height 1842032 · 12s ago
```

Recovery when a condition clears:

```
✅ PMT — node4 resolved

Infrastructure · container stopped
```

Cooldown (default 15m) applies per section+key for both alerts and recoveries.

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
go test ./internal/alert/... ./internal/panel/... -run 'Findings|Engine|Format'
```
