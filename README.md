# pmtop

Live operations dashboard for PMT / Cosmos EVM nodes.

## Architecture

```
fetch → report.Build → model.Report → markdown.Build (canonical MD)
                                              ├─ terminal: raw | glamour
                                              └─ html: goldmark + KaTeX + Mermaid
```

## Output

`pmtop` emits a single **portable Markdown** document:

- ` ```mermaid ` diagrams (economics, feemarket)
- `$$` LaTeX blocks (EIP-1559 fee math)
- GFM tables, lists, and fenced `text` / `bash` for RPC probe logs

### Terminal (default)

Interactive TUI prints Markdown (`--render raw`, default) or styled GFM (`--render glamour`). Keys: `r` refresh, `q` quit.

### Dump (CI / agents)

```bash
pmtop --dump                          # canonical markdown once, exit
pmtop --dump --format html            # HTML fragment (same as web body)
pmtop --dump --render glamour         # styled terminal GFM
```

### Web UI

```bash
pmtop --web :7777
```

Open `http://localhost:7777`. HTMX refresh every 5s; Mermaid.js + KaTeX in the browser.

### VS Code / Obsidian

`pmtop --dump > /tmp/pmt.md` and preview with Mermaid + math extensions.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-rpc` | `http://localhost:26657` | CometBFT RPC |
| `-rest` | `http://localhost:1317` | Cosmos REST |
| `-evm` | `http://localhost:8545` | EVM JSON-RPC |
| `-container` | `evmd-node` | Docker container name |
| `-web` | _(empty)_ | Web listen address (e.g. `:7777`) |
| `-dump` | `false` | Fetch once, print, exit |
| `-format` | `md` | With `-dump`: `md` or `html` |
| `-render` | `raw` | Terminal: `raw` or `glamour` |

## Build

```bash
go build -o pmtop ./cmd/pmtop
go test ./...
```
