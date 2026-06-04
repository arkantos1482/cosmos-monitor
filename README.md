# pmtop

Live operations dashboard for PMT / Cosmos EVM nodes.

## Architecture

```
fetch → report.Build → model.Report → panel.Build (HTML fragment)
                                              └─ web: HTMX shell + panel HTML
```

## Output

`pmtop` renders a structured **HTML panel** (sections, tables, Mermaid diagrams, KaTeX math):

- `<div class="mermaid">` diagrams (economics, feemarket)
- `<div class="math-display">` LaTeX blocks (EIP-1559 fee math)
- HTML tables, lists, and `<pre>` blocks for RPC probe logs

### Web UI (default)

```bash
pmtop                 # serves http://localhost:7777
pmtop -web :8080      # custom listen address
```

Open the URL in a browser. HTMX refresh every 5s; Mermaid.js + KaTeX in the browser.

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
