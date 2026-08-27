# Handoff: add a Delegate page to pmtop (wallet staking UI)

**Created:** 2026-08-26  
**Parent workspace:** `/home/abbas/workspace/cosmos-evm-old`  
**Implement in:** `tools/ops/pmtop` (git submodule → https://github.com/arkantos1482/cosmos-monitor)  
**Do not use:** retired `pmt-monitor` / `pmt-dashboard` (archived under `tools/ops/deploy/legacy/`). “PMT mon” here means **pmtop** (Go + HTMX).

---

## Goal

Add a **new dashboard page** where a user connects MetaMask (or any injected EVM wallet) and **delegates native PMT** via the staking precompile.

Ease staking: dropdown of **node1–node4 valopers**, or paste **any** `cosmosvaloper1…`. One **Delegate** click = one tx. Do **not** prescribe amounts. Do **not** add “evenly across 4.” No Remix.

This is for the **client restake** after they received unlocked tokens (handover Phase E). The four operator validators are this network’s nodes. The user becomes a **Delegator**, not a validator.

**Do not replace** the existing read-only **Staking** monitor at `/s/staking` (bonded set, VP, this-validator KPIs). Add a **separate** page.

**Suggested route:** `/delegate` (not under `/s/` — that shell 5s-polls `#data`). **No sidebar item.** Discoverability is a button on **Overview** plus the link on `/s/staking`. Links to `/delegate` must `hx-boost="false"` so wallet scripts load.

---

## Product (v1)

Must have:

1. Connect / switch wallet; add-or-switch chain **290290**.
2. Show connected address + native PMT balance.
3. Validator control: dropdown of the four operator `cosmosvaloper1…` strings, plus **Other** / paste any valoper. Suggest **addresses**, not amounts.
4. Amount input in **PMT** (18 decimals) — user-chosen; no recommended split.
5. **Delegate** button → **one** `delegate(...)` tx, value `0`. Repeat for another validator if they want. No batch / no 4-tx helper.
6. Read `delegation(user, valoper)` for the selected valoper (and show presets if cheap).
7. No numeric leftover / “leave 1 PMT” guard. Brief hints covering what can go wrong (gas, wrong chain, valoper vs `0x`, full-balance, etc.).
8. Clear errors (wrong chain, `msg.sender` mismatch, insufficient funds).

v1 nice-to-have (do if cheap): tx hash links / copy; “switch to PMT” CTA.

**Out of scope for v1:** `undelegate` UI, rewards (`0x…0801`), `createValidator`, backend signing, Cosmos CLI.

---

## Chain / contract (copy these)

| | |
|--|--|
| EVM chain ID | `290290` (`0x46df2`) |
| Cosmos chain ID | `pmt` — **never** pass this to `wallet_addEthereumChain` |
| Native token | PMT, **18 decimals** (not 6) |
| Precompile | `0x0000000000000000000000000000000000000800` |
| Wallet RPC (browser) | Public **HTTPS** from [`public-endpoints.md`](../../deploy/docs/public-endpoints.md). Not `localhost`, not `http://ec2-…:8545`. |

**Primary wallet RPC:** `https://node1.pmtchain.com`  
Fallbacks: `https://node2.pmtchain.com`, `https://node3.pmtchain.com`, `https://node4.pmtchain.com`.

Public pmtop (same file): `https://node-admin-1.pmtchain.com/` … `node-admin-4`.

pmtop **server** `-evm` stays `http://localhost:8545` on the node (read-only probes). The **page JS** must use a **public HTTPS** RPC for `wallet_addEthereumChain` (`-evm-public`, default node1 URL above). Do not assume `window.location.hostname:8545`.

**Dev vs public:** `:7778` is operator fast iteration (`make remote-dev-release`, no Docker image pull). `:7777` is prod Docker. Tunnels still work; they are not the client-facing URLs.

**ABI**

```js
[
  "function delegate(address delegatorAddress, string validatorAddress, uint256 amount) returns (bool)",
  "function undelegate(address delegatorAddress, string validatorAddress, uint256 amount) returns (int64 completionTime)",
  "function delegation(address delegatorAddress, string validatorAddress) view returns (uint256 shares, tuple(string denom, uint256 amount) balance)"
]
```

**Write**

```ts
await staking.delegate(connectedAddress, valoperString, amountWei); // NO { value }
```

Rules that will break the page if ignored:

- `delegatorAddress` **===** signer (`contract.Caller()`). Any other address reverts.
- `validatorAddress` is the **bech32 valoper string**, not `0x`.
- Amount is **wei** (`parseEther`). Do not attach native value to the tx.
- Do not `eth_sendTransaction` to `0x…0800` as a plain transfer.

**Validators (presets; also allow paste)**

| Label | valoper |
|-------|---------|
| node1 | `cosmosvaloper1akkvh0ahmve830rj4mhkdnqs49kzw23cl98zp4` |
| node2 | `cosmosvaloper1r2dqta25pxj8av9grlfxvnfje006papu0tjk0a` |
| node3 | `cosmosvaloper15hr4x4rfj0y82puk74xegugn5s5clphzcfej3e` |
| node4 | `cosmosvaloper1vmr9wxpldngnh0tvpr8h2pk2aycts3v7z8pdxh` |

Docs already in parent repo:

- `tools/ops/deploy/docs/public-endpoints.md` — HTTPS RPC and public pmtop (source of truth)
- `tools/ops/deploy/docs/staking-frontend-brief.md` — ABI / call rules / preset valopers (not a Remix/snippet product spec)
- `tools/ops/deploy/docs/client-restake-guide.md` — client MetaMask steps (no Remix, no 4-tx split)
- Official: https://docs.cosmos.network/evm/latest/documentation/smart-contracts/precompiles/staking

Gas: `estimateGas`, fallback limit **500_000**. Chain base fee is tiny; wallet 1 gwei is fine.

---

## Architecture of pmtop (do not fight it)

```
fetchall.LoadFor(view) → report.Build → model.Report → panel.BuildView (HTML fragment)
                                                              └─ html.FullPage + HTMX
```

Key files:

| File | Role |
|------|------|
| `internal/panel/sections.go` | `View`, `Nav`, `ParseView`, `writeView` |
| `internal/panel/overview.go` | Home cards — **do not** add Delegate as a 5s-polled overview card unless the card is a static link only |
| `internal/panel/staking.go` | Existing **monitor** staking page — leave it |
| `internal/render/html/server.go` | `/` and `/s/{slug}` |
| `internal/render/html/page.go` | nav icons, `dataURL`, `FullPage` |
| `internal/render/html/templates/layout.html` | shell; `#data` **polls every 5s** (`hx-trigger="every 5s"`) |
| `internal/render/html/static/theme.css` | styles |
| `cmd/pmtop/main.go` | flags, `html.Start` |
| `internal/panel/home_test.go`, `internal/render/html/page_test.go` | nav/href tests you **will** break if you add a nav item |

There is **no npm**, no SPA, almost no JS except HTMX from unpkg. Keep it that way: **embed or CDN one small `delegate.js`** (ethers v6 from CDN is OK). No Next.js, no React unless you have a very strong reason (you don’t).

**Server never signs.** No operator mnemonics, no faucet keys, no `evmd tx` from this page.

### Critical HTMX pitfall

`layout.html` morphs `#data` every 5 seconds on `/s/*`. A wallet form **inside `#data` will lose connection state**.

**Required:** serve Delegate on **`/delegate`** with **no** `hx-trigger` / `hx-get` on `#data`. Do not special-case poll on `/s/delegate`. Optional: redirect `/s/delegate` → `/delegate`.

Status strip on this page is static (no 5s claim in the header).

Boost nav (`<body hx-boost>`) is fine: full page load on section change re-inits JS. After adding a script tag, ensure it runs on full page load (and on boost: `htmx:load` or put the script in the shell so it isn’t destroyed).

Alerts (`-alert`) score **sections**. Do not map Delegate into warn/bad unless you add an explicit skip.

---

## Implementation sketch (agent should follow)

1. `panel.ViewDelegate = "delegate"`. Do **not** add a sidebar Nav item. Overview CTA + link from `/s/staking`, both `hx-boost="false"`.
2. `ParseView` / `writeView` / `navSlug` / `navIcons` / `chainRecipeFor`: Delegate uses a **minimal** fetch (status bar only).
3. `writeDelegate(w, d)`: server-render chrome (lead, dropdown of four valopers + other, precompile, chain id, public RPC). Empty “staked” filled by JS.
4. Embed `delegate.js` + styles in `theme.css`. Look like pmtop.
5. Flag `-evm-public` default `https://node1.pmtchain.com` → `html.Start` / `data-rpc`.
6. Tests:
   - nav contains `href="/delegate"`
   - fragment contains precompile, all four valopers, chain id `290290`
   - FullPage for Delegate does **not** include `hx-trigger="every 5s"` on `#data`
   - FullPage for another `/s/` view still polls
   - `go test ./...`
7. Local: `make build && make test`. This pass: `make remote-dev-release` + `make tunnel-dev` → `http://localhost:7778/delegate`. Prod Docker `:7777` / public `node-admin-*` only when asked.
8. Browser-verify with **`agent-browser`**: open `/delegate`, snapshot, confirm Connect + four validators. Wait 6s, snapshot again (form must not wipe). Full MetaMask may need a human.

Prod deploy (`playbooks/5-deploy-pmtop.yml`, `:7777`) only after user asks. Ansible must **never** restart `evmd-node` (`--no-deps pmtop`).

---

## Git / submodule

Code lives in the **cosmos-monitor** repo (submodule at `tools/ops/pmtop`).

Typical flow:

```bash
cd tools/ops/pmtop
# implement, go test ./...
git checkout -b delegate-page   # or commit on current branch per that repo
git push
# parent repo: bump submodule pointer only when asked
```

`make remote-dev-release` **pushes the pmtop repo** and rebuilds on node4 `:7778`. Confirm which remote/branch that Makefile uses before force-pushing.

Do **not** print mnemonics from `tools/ops/deploy/nodes/manifest.json`.

---

## Copy (UI)

Short lead: *Connect MetaMask and delegate native PMT to a validator. Pick node1–node4, or choose Another validator and paste a valoper. Unbonding later takes 21 days.*

---

## Prompt to paste in a new session

```text
Implement a Delegate page on pmtop from this handoff:
  tools/ops/pmtop/docs/HANDOFF-delegate-page.md
  (also tools/ops/deploy/docs/staking-frontend-brief.md)

Workspace: /home/abbas/workspace/cosmos-evm-old
App: tools/ops/pmtop (Go+HTMX submodule arkantos1482/cosmos-monitor). NOT pmt-monitor/Next.js.

Add NEW Delegate page (Overview button, not a sidebar item). Do NOT replace read-only /s/staking.
One MetaMask delegate() per click; dropdown of four operator valopers + paste any valoper.
Do not prescribe amounts. No Remix. No 4-tx even split.
Precompile 0x…0800, chain 290290, 18 decimals. Server never signs.
No 5s HTMX poll on this page. Public HTTPS RPC (public-endpoints.md), not localhost.
```
