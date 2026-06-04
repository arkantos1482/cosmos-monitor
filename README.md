# pmtop

Live operations dashboard for PMT / Cosmos EVM nodes.

## Output

`pmtop` emits a single **portable Markdown** document:

- ` ```mermaid ` diagrams (economics, feemarket)
- `$$` LaTeX blocks (EIP-1559 fee math)
- GFM tables, lists, and fenced `text` / `bash` for RPC probe logs

### Terminal (default)

Prints **raw Markdown** to stdout (good for agents and piping). Diagrams and math appear as source fences, not rendered graphics.

Keys: `r` refresh, `q` quit.

### Web UI

```bash
pmtop --web :7777
```

Open `http://localhost:7777`. The server converts Markdown to HTML and runs **Mermaid.js** + **KaTeX** in the browser (HTMX refresh every 5s).

### VS Code / Obsidian

Copy terminal output into a `.md` file and preview with extensions that support Mermaid and math (e.g. Markdown Preview Mermaid Support, Markdown+Math).

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-rpc` | `http://localhost:26657` | CometBFT RPC |
| `-rest` | `http://localhost:1317` | Cosmos REST |
| `-evm` | `http://localhost:8545` | EVM JSON-RPC |
| `-container` | `evmd-node` | Docker container name |
| `-web` | _(empty)_ | Web listen address (e.g. `:7777`) |
