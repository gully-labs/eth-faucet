<script>
  import { onMount, onDestroy } from 'svelte';
  import { getAddress } from '@ethersproject/address';

  // GullyChain testnet (gully-labs fork). The faucet logic, flags and API below are upstream's; only the page changed.
  const CHAIN = {
    id: 485590,
    hex: '0x768d6',
    name: 'GullyChain Testnet',
    rpc: 'https://testnet-rpc.gullychain.xyz',
    explorer: 'https://testnet-explorer.gullychain.xyz',
  };
  const LINKS = [
    { name: 'App', href: 'https://app.gullychain.xyz' },
    { name: 'Bridge', href: 'https://testnet-bridge.gullychain.xyz/' },
    { name: 'Docs', href: 'https://docs.gullychain.xyz/' },
    { name: 'Explorer', href: CHAIN.explorer + '/' },
  ];
  const SOURCE = 'https://github.com/gully-labs/eth-faucet';

  let input = null;
  let faucetInfo = {
    account: '0x0000000000000000000000000000000000000000',
    network: 'testnet',
    payout: 1,
    symbol: 'ETH',
    hcaptcha_sitekey: '',
  };

  let mounted = false;
  let hcaptchaLoaded = false;
  let isLoading = false;
  let widgetID = null;

  // Use unique function name to avoid global variable conflicts
  const HCAPTCHA_CALLBACK_NAME = 'hcaptchaOnLoad_' + Date.now();

  // ---- toasts (replaces bulma-toast; same call shape: toast({ message, type })) ----
  let toasts = [];
  let toastSeq = 0;
  function toast({ message, type }) {
    const id = ++toastSeq;
    toasts = [...toasts, { id, message, type: type || 'is-info' }];
    setTimeout(() => dismiss(id), type === 'is-success' ? 12000 : 7000);
  }
  function dismiss(id) {
    toasts = toasts.filter((t) => t.id !== id);
  }
  function txLink(message) {
    const m = /Txhash:\s*(0x[0-9a-fA-F]{64})/.exec(message || '');
    return m ? CHAIN.explorer + '/tx/' + m[1] : null;
  }

  // ---- live block height from the public RPC (read-only, eth_blockNumber every 6 s) ----
  let blockNumber = null;
  let blockBumped = false;
  let blockTimer = null;
  async function pollBlock() {
    try {
      const res = await fetch(CHAIN.rpc, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ jsonrpc: '2.0', id: 1, method: 'eth_blockNumber', params: [] }),
      });
      const data = await res.json();
      const n = parseInt(data?.result, 16);
      if (Number.isFinite(n) && n !== blockNumber) {
        blockNumber = n;
        blockBumped = true;
        setTimeout(() => (blockBumped = false), 900);
      }
    } catch (error) {
      console.error('Block height poll failed:', error);
    }
  }

  // ---- add the chain to an injected wallet (EIP-3085) ----
  let addingChain = false;
  async function addToWallet() {
    const eth = typeof window !== 'undefined' ? window.ethereum : null;
    if (!eth || typeof eth.request !== 'function') {
      toast({ message: 'No wallet found. Install a browser wallet, or add the network by hand with the settings below.', type: 'is-warning' });
      return;
    }
    addingChain = true;
    try {
      await eth.request({
        method: 'wallet_addEthereumChain',
        params: [
          {
            chainId: CHAIN.hex,
            chainName: CHAIN.name,
            nativeCurrency: { name: 'Ether', symbol: 'ETH', decimals: 18 },
            rpcUrls: [CHAIN.rpc],
            blockExplorerUrls: [CHAIN.explorer],
          },
        ],
      });
      toast({ message: `${CHAIN.name} added to your wallet.`, type: 'is-success' });
    } catch (error) {
      if (error?.code !== 4001) {
        toast({ message: error?.message || 'The wallet refused the network.', type: 'is-warning' });
      }
    } finally {
      addingChain = false;
    }
  }

  async function copyAccount() {
    try {
      await navigator.clipboard.writeText(faucetInfo.account);
      toast({ message: 'Faucet address copied.', type: 'is-info' });
    } catch {
      toast({ message: 'Could not copy. Select the address and copy it by hand.', type: 'is-warning' });
    }
  }

  onMount(async () => {
    pollBlock();
    blockTimer = setInterval(pollBlock, 6000);
    try {
      const res = await fetch('/api/info');
      if (!res.ok) {
        throw new Error('Failed to fetch faucet info');
      }
      faucetInfo = await res.json();
      mounted = true;
    } catch (error) {
      console.error('Failed to load faucet info:', error);
      toast({
        message: 'Failed to load faucet information. Please refresh the page.',
        type: 'is-danger',
      });
    }
  });
  onDestroy(() => {
    if (blockTimer) clearInterval(blockTimer);
  });

  // Register hCaptcha callback function
  if (typeof window !== 'undefined') {
    window[HCAPTCHA_CALLBACK_NAME] = () => {
      hcaptchaLoaded = true;
    };
  }

  $: document.title = `GullyChain Testnet Faucet · ${faucetInfo.payout} ${faucetInfo.symbol} a day`;

  $: if (
    mounted &&
    hcaptchaLoaded &&
    widgetID === null &&
    faucetInfo.hcaptcha_sitekey &&
    window.hcaptcha
  ) {
    try {
      widgetID = window.hcaptcha.render('hcaptcha', {
        sitekey: faucetInfo.hcaptcha_sitekey,
      });
    } catch (error) {
      console.error('Failed to render hCaptcha:', error);
    }
  }

  async function handleRequest() {
    if (isLoading) return;

    let address = input?.trim();
    if (!address) {
      toast({
        message: 'Please enter an address or ENS name',
        type: 'is-warning',
      });
      return;
    }

    isLoading = true;

    try {
      // Handle ENS name resolution
      if (address.endsWith('.eth')) {
        try {
          const res = await fetch(`https://api.ensdata.net/${address}`, {
            method: 'GET',
            headers: { Accept: 'application/json' },
          });

          if (!res.ok) {
            throw new Error(`HTTP ${res.status}: ${res.statusText}`);
          }

          const data = await res.json();
          const resolvedAddress = data?.address;

          if (!resolvedAddress || !resolvedAddress.startsWith('0x')) {
            toast({
              message: 'Invalid ENS name or no address found',
              type: 'is-warning',
            });
            return;
          }

          address = resolvedAddress;
        } catch (error) {
          const errorMessage = error?.message || 'Failed to resolve ENS name';
          toast({ message: errorMessage, type: 'is-warning' });
          return;
        }
      }

      // Validate and normalize Ethereum address
      try {
        address = getAddress(address);
      } catch {
        toast({ message: 'Invalid Ethereum address', type: 'is-warning' });
        return;
      }

      // Prepare request headers
      const headers = {
        'Content-Type': 'application/json',
      };

      // Execute hCaptcha verification
      if (hcaptchaLoaded && widgetID !== null) {
        try {
          const { response } = await window.hcaptcha.execute(widgetID, {
            async: true,
          });
          headers['h-captcha-response'] = response;
        } catch (error) {
          console.error('hCaptcha execution failed:', error);
          toast({
            message: 'Verification failed. Please try again.',
            type: 'is-warning',
          });
          return;
        }
      }

      // Send request
      const res = await fetch('/api/claim', {
        method: 'POST',
        headers,
        body: JSON.stringify({ address }),
      });

      let data;
      try {
        data = await res.json();
      } catch (error) {
        console.error('Failed to parse response:', error);
        toast({
          message: 'Invalid response from server. Please try again.',
          type: 'is-danger',
        });
        return;
      }

      if (res.ok) {
        const message = data?.msg || 'Transaction successful';
        toast({ message, type: 'is-success' });
        input = '';
      } else {
        const errorMessage = data?.msg || 'Request failed';
        toast({ message: errorMessage, type: 'is-danger' });
      }
    } catch (error) {
      console.error('Request failed:', error);
      toast({
        message:
          error?.message || 'An unexpected error occurred. Please try again.',
        type: 'is-danger',
      });
    } finally {
      isLoading = false;
    }
  }

  function shortAddr(a) {
    return a ? a.slice(0, 6) + '…' + a.slice(-4) : '';
  }
  const fmtBlock = (n) => (n == null ? '—' : n.toLocaleString('en-US'));
</script>

<svelte:head>
  {#if mounted && faucetInfo.hcaptcha_sitekey}
    <script
      src="https://hcaptcha.com/1/api.js?onload={HCAPTCHA_CALLBACK_NAME}&render=explicit"
      async
      defer
    ></script>
  {/if}
</svelte:head>

<a class="skip" href="#claim">Skip to the faucet</a>

<header class="top">
  <div class="wrap">
    <a class="brand" href="/" aria-label="GullyChain testnet faucet">
      <img src="/icon-96.png" width="36" height="36" alt="" />
      <span class="wordmark"><span class="gully">Gully</span><span class="chain">Chain</span></span>
      <span class="badge">Testnet faucet</span>
    </a>
    <nav class="nav" aria-label="GullyChain">
      {#each LINKS as l}
        <a href={l.href} rel="noopener">{l.name}</a>
      {/each}
    </nav>
    <div class="ticker" class:bump={blockBumped} aria-live="polite" title="Latest block from {CHAIN.rpc}">
      <span class="dot" aria-hidden="true"></span>
      <span class="lbl">Block</span>
      <span class="num">{fmtBlock(blockNumber)}</span>
    </div>
  </div>
</header>

<main>
  <section class="hero">
    <div class="mascot" aria-hidden="true">
      <img
        src="/mascot-rapid-900.webp"
        srcset="/mascot-rapid-520.webp 520w, /mascot-rapid-900.webp 900w"
        sizes="(max-width: 900px) 520px, 900px"
        width="900"
        height="900"
        alt=""
        fetchpriority="high"
      />
      <div class="fade"></div>
    </div>

    <div class="wrap hero-grid">
      <div class="copy">
        <p class="kicker">GullyChain · chain id <span class="num">{CHAIN.id}</span></p>
        <h1>Testnet ETH, <span class="hl">{faucetInfo.payout}&nbsp;{faucetInfo.symbol}</span> a day</h1>
        <p class="lead">
          One claim per address and per IP every 24 hours. Testnet ETH has no value; use it to try the
          <a href={LINKS[0].href} rel="noopener">GullyChain app</a>, deploy contracts and test the bridge.
        </p>

        <div id="hcaptcha" data-size="invisible"></div>
        <form id="claim" class="card claim" onsubmit={(e) => { e.preventDefault(); if (!isLoading) handleRequest(); }}>
          <label class="field-label" for="address">Your address or ENS name</label>
          <div class="field">
            <input
              id="address"
              bind:value={input}
              class="input num"
              type="text"
              autocomplete="off"
              spellcheck="false"
              placeholder="0x… or name.eth"
              onkeydown={(e) => {
                if (e.key === 'Enter' && !isLoading) {
                  e.preventDefault();
                  handleRequest();
                }
              }}
            />
            <button type="submit" class="btn primary" disabled={isLoading} aria-busy={isLoading}>
              {#if isLoading}
                <span class="spinner" aria-hidden="true"></span>Sending
              {:else}
                Request {faucetInfo.payout} {faucetInfo.symbol}
              {/if}
            </button>
          </div>
          <p class="hint">
            Serving from
            <button type="button" class="addr num" onclick={copyAccount} title="Copy the faucet address">{faucetInfo.account}</button>
            {#if faucetInfo.hcaptcha_sitekey}
              · protected by hCaptcha
            {/if}
          </p>
        </form>

        <div class="actions">
          <button type="button" class="btn secondary" onclick={addToWallet} disabled={addingChain}>
            Add GullyChain Testnet to wallet
          </button>
          <a class="btn ghost" href={LINKS[1].href} rel="noopener">Bridge from Arbitrum Sepolia</a>
        </div>
      </div>
    </div>
  </section>

  <section class="wrap details">
    <div class="card network">
      <p class="section-label">Network settings</p>
      <dl>
        <div><dt>Network name</dt><dd>{CHAIN.name}</dd></div>
        <div><dt>Chain ID</dt><dd class="num">{CHAIN.id}</dd></div>
        <div><dt>RPC URL</dt><dd class="num">{CHAIN.rpc}</dd></div>
        <div><dt>Currency</dt><dd class="num">ETH</dd></div>
        <div><dt>Explorer</dt><dd class="num"><a href={CHAIN.explorer} rel="noopener">{CHAIN.explorer.replace('https://', '')}</a></dd></div>
        <div><dt>Latest block</dt><dd class="num">{fmtBlock(blockNumber)}</dd></div>
      </dl>
    </div>
    <div class="card howto">
      <p class="section-label">How it works</p>
      <ol>
        <li>Add the network to your wallet, or copy the settings on the left.</li>
        <li>Paste your address above and press <b>Request</b>. The faucet sends {faucetInfo.payout} {faucetInfo.symbol} from <span class="num">{shortAddr(faucetInfo.account)}</span>.</li>
        <li>Come back tomorrow: one claim per address and per IP every 24 hours.</li>
      </ol>
      <p class="muted">
        Need more? Bridge Arbitrum Sepolia ETH through the <a href={LINKS[1].href} rel="noopener">testnet bridge</a>,
        or read the <a href={LINKS[2].href} rel="noopener">docs</a>.
      </p>
    </div>
  </section>
</main>

<footer class="foot">
  <div class="wrap">
    <span class="wordmark small"><span class="gully">Gully</span><span class="chain">Chain</span></span>
    <span class="tagline">Your Profile Is Your Wallet</span>
    <span class="muted">Testnet only. Tokens have no value. <a href={SOURCE} rel="noopener">Source</a> (MIT, a fork of chainflag/eth-faucet).</span>
  </div>
</footer>

<div class="toasts" aria-live="polite" aria-atomic="false">
  {#each toasts as t (t.id)}
    <div class="toast {t.type}" role="status">
      <span class="msg">
        {t.message}
        {#if txLink(t.message)}
          · <a href={txLink(t.message)} rel="noopener">View on explorer</a>
        {/if}
      </span>
      <button type="button" class="x" onclick={() => dismiss(t.id)} aria-label="Dismiss">×</button>
    </div>
  {/each}
</div>

<style>
  .wrap {
    max-width: 1180px;
    margin: 0 auto;
    padding: 0 clamp(16px, 4vw, 40px);
  }
  .skip {
    position: absolute;
    left: -999px;
    top: 8px;
    background: var(--gc-cyan);
    color: #000;
    padding: 8px 14px;
    border-radius: var(--gc-radius-pill);
    font-weight: 700;
    z-index: 100;
  }
  .skip:focus {
    left: 16px;
  }

  /* ---- header ---- */
  .top {
    position: sticky;
    top: 0;
    z-index: 20;
    background: rgba(0, 0, 0, 0.82);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border-bottom: 1px solid var(--gc-line);
  }
  .top .wrap {
    display: flex;
    align-items: center;
    gap: 16px;
    min-height: 64px;
    flex-wrap: wrap;
  }
  .brand {
    display: flex;
    align-items: center;
    gap: 10px;
    color: var(--gc-text);
    margin-right: auto;
  }
  .brand img {
    width: 36px;
    height: 36px;
  }
  .wordmark {
    font-family: var(--f-wordmark);
    font-weight: 800;
    letter-spacing: -0.02em;
    font-size: 24px;
    line-height: 1;
    white-space: nowrap;
  }
  .wordmark.small {
    font-size: 20px;
  }
  .wordmark .gully {
    color: #ffffff;
  }
  .wordmark .chain {
    background: var(--gc-chain-gradient);
    -webkit-background-clip: text;
    background-clip: text;
    color: transparent;
  }
  .badge {
    font-family: var(--f-badge);
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-size: 14px;
    line-height: 1;
    display: inline-flex;
    align-items: center;
    height: 28px;
    padding: 0 12px;
    border-radius: var(--gc-radius-pill);
    background: var(--gc-royal);
    color: #ffffff;
    margin-left: 4px;
  }
  .nav {
    display: flex;
    gap: 4px;
  }
  .nav a {
    color: var(--gc-muted);
    font-weight: 600;
    font-size: 15px;
    padding: 8px 10px;
    border-radius: var(--gc-radius-pill);
  }
  .nav a:hover {
    color: var(--gc-text);
    background: var(--gc-night);
  }
  .ticker {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    height: 36px;
    padding: 0 14px 0 12px;
    border: 1px solid var(--gc-line);
    background: var(--gc-night);
    border-radius: var(--gc-radius-pill);
    font-size: 14px;
    color: var(--gc-muted);
    transition: border-color 0.3s;
  }
  .ticker.bump {
    border-color: var(--gc-cyan);
  }
  .ticker .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--gc-cyan);
    animation: pulse 2s infinite;
  }
  .ticker .lbl {
    font-family: var(--f-badge);
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-size: 13px;
  }
  .ticker .num {
    color: var(--gc-text);
  }
  .num {
    font-family: var(--f-num);
    font-weight: 500;
    font-variant-numeric: tabular-nums;
  }
  @keyframes pulse {
    50% {
      opacity: 0.35;
    }
  }

  /* ---- hero: text left over solid black, mascot right, faded with the kit gradient ---- */
  .hero {
    position: relative;
    background: var(--gc-black);
    overflow: hidden;
    padding: clamp(40px, 7vw, 88px) 0 clamp(32px, 5vw, 64px);
  }
  .mascot {
    position: absolute;
    inset: 0 0 0 auto;
    width: min(62vw, 820px);
    pointer-events: none;
  }
  .mascot img {
    position: absolute;
    right: -4%;
    top: 50%;
    transform: translateY(-50%);
    width: 100%;
    height: auto;
    max-width: none;
    /* soften the pose's own dark backdrop into the page black (the mascot itself is untouched) */
    -webkit-mask-image: radial-gradient(ellipse 56% 58% at 54% 48%, #000 28%, transparent 74%);
    mask-image: radial-gradient(ellipse 56% 58% at 54% 48%, #000 28%, transparent 74%);
  }
  .mascot .fade {
    position: absolute;
    inset: 0;
    background: linear-gradient(90deg, #000 40%, rgba(0, 0, 0, 0.6) 58%, transparent 80%);
  }
  .hero-grid {
    position: relative;
    display: grid;
    grid-template-columns: minmax(0, 620px) minmax(0, 1fr);
  }
  .kicker {
    margin: 0 0 12px;
    font-family: var(--f-badge);
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-size: 15px;
    color: var(--gc-cyan);
  }
  h1 {
    font-family: var(--f-wordmark);
    font-weight: 800;
    letter-spacing: -0.02em;
    font-size: clamp(36px, 5.4vw, 64px);
    line-height: 1.02;
    margin: 0 0 18px;
    color: #ffffff;
  }
  h1 .hl {
    color: var(--gc-cyan);
    font-variant-numeric: tabular-nums;
  }
  .lead {
    color: var(--gc-muted);
    font-size: 17px;
    max-width: 540px;
    margin: 0 0 26px;
  }
  .lead a {
    color: var(--gc-text);
    border-bottom: 1px solid var(--gc-cyan);
  }

  /* ---- card + form ---- */
  .card {
    background: var(--gc-night);
    border: 1px solid var(--gc-line);
    border-radius: var(--gc-radius-card);
    padding: 24px;
  }
  .claim {
    margin: 0;
    max-width: 620px;
  }
  .field-label {
    display: block;
    font-family: var(--f-badge);
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-size: 13px;
    color: var(--gc-muted);
    margin-bottom: 10px;
  }
  .field {
    display: flex;
    gap: 10px;
  }
  .input {
    flex: 1 1 auto;
    min-width: 0;
    min-height: 48px;
    padding: 0 18px;
    border-radius: var(--gc-radius-pill);
    border: 1px solid var(--gc-line);
    background: var(--gc-black);
    color: var(--gc-text);
    font-size: 15px;
    outline: none;
    transition: border-color 0.15s;
  }
  .input::placeholder {
    color: var(--gc-muted);
    opacity: 0.8;
  }
  .input:focus {
    border-color: var(--gc-cyan);
  }
  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    min-height: 44px;
    padding: 0 22px;
    border-radius: var(--gc-radius-pill);
    border: 0;
    font-family: var(--f-ui);
    font-weight: 700;
    font-size: 15px;
    cursor: pointer;
    white-space: nowrap;
    transition: filter 0.15s, transform 0.15s;
  }
  .btn:hover:not(:disabled) {
    filter: brightness(1.08);
  }
  .btn:active:not(:disabled) {
    transform: translateY(1px);
  }
  .btn:disabled {
    cursor: default;
    opacity: 0.7;
  }
  .btn.primary {
    background: var(--gc-cyan);
    color: #000000;
  }
  .btn.secondary {
    background: var(--gc-royal);
    color: #ffffff;
  }
  .btn.ghost {
    background: transparent;
    color: var(--gc-text);
    border: 1px solid var(--gc-line);
  }
  .btn.ghost:hover {
    border-color: var(--gc-cyan);
  }
  .spinner {
    width: 16px;
    height: 16px;
    border-radius: 50%;
    border: 2px solid rgba(0, 0, 0, 0.25);
    border-top-color: #000;
    animation: spin 0.8s linear infinite;
  }
  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
  .hint {
    margin: 12px 0 0;
    font-size: 13.5px;
    color: var(--gc-muted);
    overflow-wrap: anywhere;
  }
  .addr {
    background: none;
    border: 0;
    padding: 0;
    color: var(--gc-text);
    font-size: 13.5px;
    cursor: copy;
    border-bottom: 1px dashed var(--gc-line);
    overflow-wrap: anywhere;
    word-break: break-all;
    text-align: left;
  }
  .addr:hover {
    border-bottom-color: var(--gc-cyan);
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    margin-top: 18px;
  }

  /* ---- details ---- */
  .details {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
    padding-top: 8px;
    padding-bottom: 48px;
  }
  .section-label {
    margin: 0 0 14px;
    font-family: var(--f-badge);
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-size: 14px;
    color: var(--gc-cyan);
  }
  dl {
    margin: 0;
    display: grid;
    gap: 0;
  }
  dl div {
    display: grid;
    grid-template-columns: 130px 1fr;
    gap: 12px;
    padding: 9px 0;
    border-top: 1px solid var(--gc-line);
    font-size: 14.5px;
  }
  dt {
    color: var(--gc-muted);
  }
  dd {
    margin: 0;
    color: var(--gc-text);
    overflow-wrap: anywhere;
  }
  ol {
    margin: 0 0 14px;
    padding-left: 20px;
    color: var(--gc-text);
    font-size: 15px;
  }
  ol li {
    margin-bottom: 8px;
  }
  .muted {
    color: var(--gc-muted);
    font-size: 14px;
    margin: 0;
  }

  /* ---- footer ---- */
  .foot {
    border-top: 1px solid var(--gc-line);
    padding: 24px 0 32px;
  }
  .foot .wrap {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 10px 18px;
  }
  .tagline {
    font-family: var(--f-tagline);
    font-weight: 500;
    letter-spacing: 0.2em;
    font-size: 11px;
    color: #ffffff;
    text-transform: none;
  }

  /* ---- toasts ---- */
  .toasts {
    position: fixed;
    left: 50%;
    bottom: 20px;
    transform: translateX(-50%);
    display: grid;
    gap: 8px;
    z-index: 50;
    width: min(560px, calc(100vw - 32px));
  }
  .toast {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    padding: 12px 14px 12px 16px;
    border-radius: 14px;
    border: 1px solid var(--gc-line);
    background: var(--gc-night);
    color: var(--gc-text);
    font-size: 14.5px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.45);
    animation: rise 0.25s ease-out;
  }
  .toast .msg {
    flex: 1;
    overflow-wrap: anywhere;
  }
  .toast.is-success {
    border-color: var(--gc-cyan);
  }
  .toast.is-warning {
    border-color: var(--gc-royal);
  }
  .toast.is-danger {
    border-color: var(--gc-down);
  }
  .toast .x {
    background: none;
    border: 0;
    color: var(--gc-muted);
    font-size: 20px;
    line-height: 1;
    cursor: pointer;
    padding: 0 2px;
  }
  .toast .x:hover {
    color: var(--gc-text);
  }
  @keyframes rise {
    from {
      opacity: 0;
      transform: translateY(8px);
    }
  }

  @media (max-width: 900px) {
    .hero-grid {
      grid-template-columns: minmax(0, 1fr);
    }
    .mascot {
      width: 80vw;
      opacity: 0.55;
    }
    .mascot img {
      right: -30%;
    }
    .details {
      grid-template-columns: 1fr;
    }
    .nav {
      order: 3;
      width: 100%;
      margin: 0 -10px 8px;
    }
  }
  @media (max-width: 560px) {
    .field {
      flex-direction: column;
    }
    .btn {
      width: 100%;
      white-space: normal;
    }
    .badge {
      display: none;
    }
    dl div {
      grid-template-columns: 1fr;
      gap: 2px;
    }
    .ticker {
      height: 32px;
      font-size: 13px;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .ticker .dot,
    .toast {
      animation: none;
    }
  }
</style>
