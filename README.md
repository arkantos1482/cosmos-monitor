# pmtop

Live operations dashboard for PMT / Cosmos EVM nodes.

## Architecture

```
fetch → report.Build → model.Report → panel.Build (HTML fragment)
                                              ├─ terminal: plain text (htmlToPlain)
                                              └─ web: HTMX shell + panel HTML
```

## Output

`pmtop` renders a structured **HTML panel** (sections, tables, Mermaid diagrams, KaTeX math):

- `<div class="mermaid">` diagrams (economics, feemarket)
- `<div class="math-display">` LaTeX blocks (EIP-1559 fee math)
- HTML tables, lists, and `<pre>` blocks for RPC probe logs

### Terminal (default)

Interactive TUI prints plain text derived from the same panel (all sections preserved). Keys: `r` refresh, `q` quit.

### Dump (CI / agents)

```bash
pmtop --dump                          # plain text once, exit
pmtop --dump --format html            # HTML fragment (same as web body)
```

### Web UI

```bash
pmtop --web :7777
```

Open `http://localhost:7777`. HTMX refresh every 5s; Mermaid.js + KaTeX in the browser.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-rpc` | `http://localhost:26657` | CometBFT RPC |
| `-rest` | `http://localhost:1317` | Cosmos REST |
| `-evm` | `http://localhost:8545` | EVM JSON-RPC |
| `-container` | `evmd-node` | Docker container name |
| `-web` | _(empty)_ | Web listen address (e.g. `:7777`) |
| `-dump` | `false` | Fetch once, print, exit |
| `-format` | `plain` | With `-dump`: `plain` or `html` |

## Build

```bash
go build -o pmtop ./cmd/pmtop
go test ./...
```
