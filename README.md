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
- **Live updates**: `#data` polls its URL every 5s via HTMX (`innerHTML` swap, no scroll jump). Mermaid/KaTeX re-init after poll swaps.

## Output

`pmtop` renders a structured **HTML panel** using dashboard components (cards, stat grids, badges, data tables) plus Mermaid and KaTeX:

- `<section class="dash-section">` / stat grids / status badges
- `<div class="mermaid">` diagrams (economics, feemarket)
- `<div class="math-panel">` with per-line `<div class="math-line">` KaTeX (EIP-1559 fee math)
- Scrollable data tables and code blocks for RPC probe logs

### Web UI (default)

```bash
pmtop                 # serves http://localhost:7777
pmtop -web :8080      # custom listen address
```

Open the URL in a browser. Section navigation uses HTMX boost (full SSR page per section); `#data` polls every 5s for live metrics. Mermaid.js + KaTeX re-init after poll swaps.

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

## Makefile (node4 ops)

Remote targets SSH to node4 (`tools/ops/pmtop/Makefile`). Typical flow: `make push-deploy`, then `make start` and `make tunnel` to open http://localhost:7777.

| Target | Description |
|--------|-------------|
| `run` | Foreground web UI on node4 (`~/pmtop`, port 7777) |
| `start` / `stop` | Background tmux session on node4 |
| `tunnel` | Forward node4 :7777 to localhost |
| `deploy` | Pull, build, smoke-test `--dump` on node4 |
| `logs` / `status` / `evmd` / `shell` | Validator logs, RPC status, `evmd` CLI, container shell |
