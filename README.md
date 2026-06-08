# pmtop

Live operations dashboard for PMT / Cosmos EVM nodes.

## Architecture

```
fetchall.LoadFor(view) → report.Build → model.Report → panel.BuildView (HTML fragment)
                                                              └─ web: html/template shell + HTMX
```

- **View-scoped fetch**: each page (`/` or `/s/{slug}`) fetches only the data that section needs, with a short (~4s) per-view snapshot cache so polls and boost navigations do not hammer RPC.
- **Dual-mode HTTP**: same URL serves a full HTML document (direct load or `HX-Boosted` nav) or a fragment when `HX-Request` is set without `HX-Boosted` (5s poll on `#data`).
- **Boost navigation**: `<body hx-boost>` handles section links and home cards; the server renders the full shell with the correct active nav. No client-side nav sync hacks.
- **Live updates**: `#data` polls its URL every 5s via HTMX (`innerHTML` swap, no scroll jump).

## Output

`pmtop` renders a structured **HTML panel** using dashboard components (cards, stat grids, badges, data tables):

- `<section class="dash-section">` / stat grids / status badges
- `<pre class="fee-formula">` and code blocks for EIP-1559 fee math
- Scrollable data tables for RPC probe logs

### Web UI (default)

```bash
pmtop                 # serves http://localhost:7777
pmtop -web :8080      # custom listen address
```

Open the URL in a browser. Section navigation uses HTMX boost (full SSR page per section); `#data` polls every 5s for live metrics.

### Dump (CI / agents)

```bash
pmtop --dump          # HTML fragment to stdout, then exit
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-rpc` | `http://localhost:26657` | CometBFT RPC |
| `-rest` | `http://localhost:1317` | Cosmos REST |
| `-evm` | `http://localhost:8545` | EVM JSON-RPC |
| `-container` | `evmd-node` | Docker container name |
| `-web` | `:7777` | Web listen address; empty disables (requires `-dump`) |
| `-dump` | `false` | Fetch once, print HTML fragment, exit |

## Submodule (parent repo)

This directory is a git submodule in the cosmos-evm parent repo:

```bash
# From parent repo root
git submodule update --init tools/ops/pmtop
```

Remote: https://github.com/arkantos1482/cosmos-monitor

**Dev loop:** commit → `make remote-dev-release` (push + reload on node4).

## Build

```bash
go build -o pmtop ./cmd/pmtop
go test ./...
```

Or from this directory:

```bash
make build    # ./pmtop
make test
make serve    # web UI on http://localhost:7777
make dump     # HTML fragment to stdout (one shot)
```

## Makefile (remote ops)

`NODE4_HOST` is node4 today (change at top of `Makefile` to point elsewhere). Fleet SSH uses `SSH_USER` / `SSH_KEY`. Run `make help`.

**Remote pmtop:** `remote-pull`, `remote-build`, `remote-smoke`, `remote-stop`, `remote-start`, `remote-verify`, …

**Remote validator** (docker/evmd, not pmtop): `remote-logs`, `remote-status`, `remote-evmd`, `remote-shell`

**Integration** (atomics only, one layer):

| Target | Atomics |
|--------|---------|
| `remote-deploy` | remote-pull + remote-build + remote-smoke |
| `remote-restart` | remote-stop + remote-start |
| `remote-reload` | remote-pull + remote-build + remote-smoke + remote-stop + remote-start + remote-verify |
| `remote-dev-release` | push + remote-pull + remote-build + remote-smoke + remote-stop + remote-start + remote-verify |

Typical flow: `make remote-dev-release`, then `make tunnel` → http://localhost:7777
