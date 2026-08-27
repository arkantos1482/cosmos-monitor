(function () {
  const app = document.getElementById("delegate-app");
  if (!app) return;

  const chainId = Number(app.dataset.chainId || "290290");
  const chainHex = "0x" + chainId.toString(16);
  const precompile = app.dataset.precompile;
  const rpc = app.dataset.rpc;
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
    presetWrap: document.getElementById("delegate-preset-wrap"),
    preset: document.getElementById("delegate-preset-valoper"),
    customWrap: document.getElementById("delegate-custom-wrap"),
    custom: document.getElementById("delegate-valoper-custom"),
    amount: document.getElementById("delegate-amount"),
    submit: document.getElementById("delegate-submit"),
    error: document.getElementById("delegate-error"),
    tx: document.getElementById("delegate-tx"),
  };

  let provider;
  let signer;

  function showError(msg) {
    el.error.hidden = !msg;
    el.error.textContent = msg || "";
  }

  function showTx(hash) {
    el.tx.hidden = !hash;
    el.tx.textContent = hash ? "tx " + hash : "";
  }

  function valoper() {
    if (el.select.value === "custom") {
      return (el.custom.value || "").trim();
    }
    return el.select.value;
  }

  function syncCustom() {
    const custom = el.select.value === "custom";
    if (el.picker) {
      el.picker.classList.toggle("delegate-picker--custom", custom);
    }
    if (el.presetWrap) {
      el.presetWrap.hidden = custom;
    }
    el.customWrap.hidden = !custom;
    if (!custom && el.preset) {
      el.preset.textContent = el.select.value;
    }
    if (custom && el.custom) {
      el.custom.focus();
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

  async function connect() {
    showError("");
    if (typeof ethers === "undefined") {
      showError("ethers failed to load (CDN). Reload the page.");
      return;
    }
    if (!window.ethereum) {
      showError("No injected wallet. Install MetaMask or Rabby.");
      return;
    }
    try {
      await ensureChain();
      provider = new ethers.BrowserProvider(window.ethereum, "any");
      await provider.send("eth_requestAccounts", []);
      signer = await provider.getSigner();
      el.connect.textContent = "Switch / reconnect";
      await refresh();
    } catch (e) {
      showError(e && e.message ? e.message : String(e));
    }
  }

  async function delegate() {
    showError("");
    showTx("");
    if (!signer) {
      showError("Connect a wallet first.");
      return;
    }
    const v = valoper();
    if (!v.startsWith("cosmosvaloper1") || v.startsWith("0x")) {
      showError("Validator must be a cosmosvaloper1… string, not 0x.");
      return;
    }
    const amt = (el.amount.value || "").trim();
    if (!amt) {
      showError("Enter an amount in PMT.");
      return;
    }
    let wei;
    try {
      wei = ethers.parseEther(amt);
    } catch (e) {
      showError("Amount is not a valid PMT number (18 decimals).");
      return;
    }
    if (wei <= 0n) {
      showError("Amount must be greater than 0.");
      return;
    }
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
      await tx.wait();
      await refresh();
    } catch (e) {
      const msg = e && e.shortMessage ? e.shortMessage : (e && e.message ? e.message : String(e));
      showError(msg);
    }
  }

  el.connect.addEventListener("click", connect);
  el.submit.addEventListener("click", delegate);
  el.select.addEventListener("change", function () {
    syncCustom();
    refresh();
  });
  el.custom.addEventListener("change", refresh);
  if (window.ethereum) {
    window.ethereum.on("accountsChanged", function () { connect(); });
    window.ethereum.on("chainChanged", function () { connect(); });
  }
  syncCustom();
})();
