(function () {
  const app = document.getElementById("delegate-app");
  if (!app) return;

  const chainId = Number(app.dataset.chainId || "290290");
  const chainHex = "0x" + chainId.toString(16);
  const precompile = app.dataset.precompile;
  const rpc = app.dataset.rpc;
  const explorer = (app.dataset.explorer || "https://pmtscan.com").replace(/\/$/, "");
  const fallbacks = (app.dataset.rpcFallbacks || "").split(",").map((s) => s.trim()).filter(Boolean);
  const abi = [
    "function delegate(address delegatorAddress, string validatorAddress, uint256 amount) returns (bool)",
    "function delegation(address delegatorAddress, string validatorAddress) view returns (uint256 shares, tuple(string denom, uint256 amount) balance)",
  ];

  const el = {
    addr: document.getElementById("delegate-address"),
    bal: document.getElementById("delegate-balance"),
    staked: document.getElementById("delegate-staked"),
    chain: document.getElementById("delegate-chain"),
    connect: document.getElementById("delegate-connect"),
    picker: document.getElementById("delegate-picker"),
    select: document.getElementById("delegate-valoper-select"),
    valoper: document.getElementById("delegate-valoper"),
    amount: document.getElementById("delegate-amount"),
    submit: document.getElementById("delegate-submit"),
    error: document.getElementById("delegate-error"),
    errorSubmit: document.getElementById("delegate-error-submit"),
    status: document.getElementById("delegate-status"),
    statusConnect: document.getElementById("delegate-status-connect"),
    tx: document.getElementById("delegate-tx"),
  };

  let provider;
  let signer;
  let inflight = false;
  let connected = false;

  function showError(msg, near) {
    const text = msg || "";
    function setSlot(node, value) {
      if (!node) return;
      node.hidden = !value;
      node.textContent = value;
    }
    if (!text) {
      setSlot(el.error, "");
      setSlot(el.errorSubmit, "");
      return;
    }
    const connectSlot = near !== "submit";
    setSlot(el.error, connectSlot ? text : "");
    setSlot(el.errorSubmit, connectSlot ? "" : text);
    const target = connectSlot ? el.error : el.errorSubmit;
    if (target && target.scrollIntoView) {
      target.scrollIntoView({ block: "nearest", behavior: "smooth" });
    }
  }

  function setStatus(node, kind, text) {
    if (!node) return;
    node.hidden = !text;
    node.textContent = text || "";
    node.classList.toggle("delegate-app__status-msg--pending", kind === "pending");
    node.classList.toggle("delegate-app__status-msg--ok", kind === "ok");
  }

  function showTx(hash) {
    el.tx.hidden = !hash;
    el.tx.textContent = hash ? "tx " + hash : "";
  }

  function connectLabel() {
    return connected ? "Switch / reconnect" : "Connect MetaMask";
  }

  function setInflight(on) {
    inflight = on;
    el.connect.disabled = on;
    el.submit.disabled = on;
    el.connect.setAttribute("aria-busy", on ? "true" : "false");
    el.submit.setAttribute("aria-busy", on ? "true" : "false");
    if (!on) {
      el.connect.textContent = connectLabel();
      el.submit.textContent = "Delegate";
    }
  }

  function valoper() {
    return (el.valoper.value || "").trim();
  }

  function fillValoperFromSelect() {
    const custom = el.select.value === "custom";
    if (el.picker) {
      el.picker.classList.toggle("delegate-picker--custom", custom);
    }
    if (custom) {
      el.valoper.disabled = false;
      el.valoper.value = "";
      el.valoper.focus();
      return;
    }
    el.valoper.value = el.select.value;
    el.valoper.disabled = true;
  }

  function matchSelectToValoper() {
    if (el.valoper.disabled) return;
    const v = valoper();
    let found = "custom";
    for (let i = 0; i < el.select.options.length; i++) {
      const opt = el.select.options[i];
      if (opt.value !== "custom" && opt.value === v) {
        found = opt.value;
        break;
      }
    }
    el.select.value = found;
    if (found !== "custom") {
      fillValoperFromSelect();
    } else if (el.picker) {
      el.picker.classList.toggle("delegate-picker--custom", true);
    }
  }

  async function ensureChain() {
    const eth = window.ethereum;
    if (!eth) throw new Error("No injected wallet. Install MetaMask or Rabby.");
    const current = await eth.request({ method: "eth_chainId" });
    if (current.toLowerCase() === chainHex.toLowerCase()) return;
    try {
      await eth.request({
        method: "wallet_switchEthereumChain",
        params: [{ chainId: chainHex }],
      });
    } catch (err) {
      if (err && (err.code === 4902 || err.code === -32603)) {
        await eth.request({
          method: "wallet_addEthereumChain",
          params: [{
            chainId: chainHex,
            chainName: "PMT",
            nativeCurrency: { name: "PMT", symbol: "PMT", decimals: 18 },
            rpcUrls: [rpc].concat(fallbacks),
            blockExplorerUrls: [explorer],
          }],
        });
      } else {
        throw err;
      }
    }
  }

  async function refresh() {
    if (!signer || typeof ethers === "undefined") return;
    const addr = await signer.getAddress();
    el.addr.textContent = addr;
    const net = await provider.getNetwork();
    const id = Number(net.chainId);
    el.chain.textContent = id === chainId ? String(id) : id + " (want " + chainId + ")";
    if (id !== chainId) {
      showError("Wrong chain " + id + ". Switch to 290290.");
      return;
    }
    const wei = await provider.getBalance(addr);
    el.bal.textContent = ethers.formatEther(wei) + " PMT";
    const v = valoper();
    if (!v.startsWith("cosmosvaloper1")) {
      el.staked.textContent = "—";
      return;
    }
    const c = new ethers.Contract(precompile, abi, provider);
    try {
      const d = await c.delegation(addr, v);
      el.staked.textContent = ethers.formatEther(d.balance.amount) + " PMT";
    } catch (e) {
      el.staked.textContent = "unread";
    }
  }

  async function attachSigner() {
    provider = new ethers.BrowserProvider(window.ethereum, "any");
    signer = await provider.getSigner();
    connected = true;
    el.connect.textContent = connectLabel();
  }

  // Silent: already permitted + already on 290290. Do not switch/add chain here
  // (that would pop MetaMask on every visit while the user is on another network).
  async function trySilentConnect() {
    if (typeof ethers === "undefined" || !window.ethereum) return;
    try {
      const accounts = await window.ethereum.request({ method: "eth_accounts" });
      if (!accounts || !accounts.length) return;
      const current = await window.ethereum.request({ method: "eth_chainId" });
      if ((current || "").toLowerCase() !== chainHex.toLowerCase()) return;
      await attachSigner();
      setStatus(el.statusConnect, "ok", "Connected.");
      await refresh();
    } catch (e) {
      // Stay disconnected; Connect still works.
    }
  }

  async function connect() {
    if (inflight) return;
    showError("");
    setStatus(el.statusConnect, "", "");
    if (typeof ethers === "undefined") {
      showError("ethers failed to load (CDN). Reload the page.");
      return;
    }
    if (!window.ethereum) {
      showError("No injected wallet. Install MetaMask or Rabby.");
      return;
    }
    setInflight(true);
    el.connect.textContent = "Connecting…";
    setStatus(el.statusConnect, "pending", "Waiting for wallet…");
    try {
      await ensureChain();
      await window.ethereum.request({ method: "eth_requestAccounts" });
      await attachSigner();
      setStatus(el.statusConnect, "ok", "Connected.");
      await refresh();
    } catch (e) {
      setStatus(el.statusConnect, "", "");
      showError(e && e.message ? e.message : String(e));
    } finally {
      setInflight(false);
    }
  }

  async function delegate() {
    if (inflight) return;
    showError("");
    showTx("");
    setStatus(el.status, "", "");
    if (!signer) {
      showError("Connect a wallet first.", "submit");
      return;
    }
    const v = valoper();
    if (!v.startsWith("cosmosvaloper1") || v.startsWith("0x")) {
      showError("Validator must be a cosmosvaloper1… string, not 0x.", "submit");
      return;
    }
    const amt = (el.amount.value || "").trim();
    if (!amt) {
      showError("Enter an amount in PMT.", "submit");
      return;
    }
    let wei;
    try {
      wei = ethers.parseEther(amt);
    } catch (e) {
      showError("Amount is not a valid PMT number.", "submit");
      return;
    }
    if (wei <= 0n) {
      showError("Amount must be greater than 0.", "submit");
      return;
    }
    setInflight(true);
    el.submit.textContent = "Waiting for wallet…";
    setStatus(el.status, "pending", "Approve the transaction in your wallet. Do not click Delegate again.");
    try {
      await ensureChain();
      const addr = await signer.getAddress();
      const c = new ethers.Contract(precompile, abi, signer);
      let gas;
      try {
        gas = await c.delegate.estimateGas(addr, v, wei);
      } catch (e) {
        gas = 500000n;
      }
      const tx = await c.delegate(addr, v, wei, { gasLimit: gas });
      showTx(tx.hash);
      el.submit.textContent = "Confirming…";
      setStatus(el.status, "pending", "Submitted. Waiting for confirmation…");
      await tx.wait();
      setStatus(el.status, "ok", "Delegated. Balance and stake below are refreshed.");
      await refresh();
    } catch (e) {
      setStatus(el.status, "", "");
      const msg = e && e.shortMessage ? e.shortMessage : (e && e.message ? e.message : String(e));
      showError(msg, "submit");
    } finally {
      setInflight(false);
    }
  }

  el.connect.addEventListener("click", connect);
  el.submit.addEventListener("click", delegate);
  el.select.addEventListener("change", function () {
    fillValoperFromSelect();
    refresh();
  });
  el.valoper.addEventListener("input", matchSelectToValoper);
  el.valoper.addEventListener("change", refresh);
  if (window.ethereum) {
    window.ethereum.on("accountsChanged", function () { if (!inflight) connect(); });
    window.ethereum.on("chainChanged", function () { if (!inflight) connect(); });
  }
  fillValoperFromSelect();
  trySilentConnect();
})();
