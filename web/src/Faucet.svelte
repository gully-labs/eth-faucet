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
    { name: 'App', href: 'https://app.gullychain.xyz', icon: 'grid' },
    { name: 'Bridge', href: 'https://testnet-bridge.gullychain.xyz/', icon: 'arrows' },
    { name: 'Docs', href: 'https://docs.gullychain.xyz/', icon: 'book' },
    { name: 'Explorer', href: CHAIN.explorer + '/', icon: 'search' },
  ];
  const SOURCE = 'https://github.com/gully-labs/eth-faucet';

  // Lucide icons (ISC, lucide.dev): 24-unit paths, drawn at 20px, stroke 1.75
  const ICONS = {
    grid: '<rect width="7" height="7" x="3" y="3" rx="1"/><rect width="7" height="7" x="14" y="3" rx="1"/><rect width="7" height="7" x="14" y="14" rx="1"/><rect width="7" height="7" x="3" y="14" rx="1"/>',
    arrows: '<path d="M8 3 4 7l4 4"/><path d="M4 7h16"/><path d="m16 21 4-4-4-4"/><path d="M20 17H4"/>',
    book: '<path d="M12 7v14"/><path d="M3 18a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1h5a4 4 0 0 1 4 4 4 4 0 0 1 4-4h5a1 1 0 0 1 1 1v13a1 1 0 0 1-1 1h-6a3 3 0 0 0-3 3 3 3 0 0 0-3-3z"/>',
    search: '<circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/>',
    droplet: '<path d="M12 22a7 7 0 0 0 7-7c0-2-1-3.9-3-5.5s-3.5-4-4-6.5c-.5 2.5-2 4.9-4 6.5C6 11.1 5 13 5 15a7 7 0 0 0 7 7z"/>',
    wallet: '<path d="M19 7V4a1 1 0 0 0-1-1H5a2 2 0 0 0 0 4h15a1 1 0 0 1 1 1v4h-3a2 2 0 0 0 0 4h3a1 1 0 0 0 1-1v-2a1 1 0 0 0-1-1"/><path d="M3 5v14a2 2 0 0 0 2 2h15a1 1 0 0 0 1-1v-4"/>',
    plus: '<circle cx="12" cy="12" r="10"/><path d="M8 12h8"/><path d="M12 8v8"/>',
    check: '<circle cx="12" cy="12" r="10"/><path d="m9 12 2 2 4-4"/>',
    alert: '<path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3"/><path d="M12 9v4"/><path d="M12 17h.01"/>',
    xcircle: '<circle cx="12" cy="12" r="10"/><path d="m15 9-6 6"/><path d="m9 9 6 6"/>',
    info: '<circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/>',
    copy: '<rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/>',
    arrow: '<path d="M5 12h14"/><path d="m12 5 7 7-7 7"/>',
    x: '<path d="M18 6 6 18"/><path d="m6 6 12 12"/>',
    boxes: '<path d="M2.97 12.92A2 2 0 0 0 2 14.63v3.24a2 2 0 0 0 .97 1.71l3 1.8a2 2 0 0 0 2.06 0L12 19v-5.5l-5-3-4.03 2.42Z"/><path d="m7 16.5-4.74-2.85"/><path d="m7 16.5 5-3"/><path d="M7 16.5v5.17"/><path d="M12 13.5V19l3.97 2.38a2 2 0 0 0 2.06 0l3-1.8a2 2 0 0 0 .97-1.71v-3.24a2 2 0 0 0-.97-1.71L17 10.5l-5 3Z"/><path d="m17 16.5-5-3"/><path d="m17 16.5 4.74-2.85"/><path d="M17 16.5v5.17"/><path d="M7.97 4.42A2 2 0 0 0 7 6.13v4.37l5 3 5-3V6.13a2 2 0 0 0-.97-1.71l-3-1.8a2 2 0 0 0-2.06 0l-3 1.8Z"/><path d="M12 8 7.26 5.15"/><path d="m12 8 4.74-2.85"/><path d="M12 13.5V8"/>',
    clock: '<circle cx="12" cy="12" r="10"/><path d="M12 6v6l4 2"/>',
    settings: '<path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/>',
  };
  const icon = (n) => `<svg viewBox="0 0 24 24" aria-hidden="true">${ICONS[n] || ICONS.info}</svg>`;

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

  // ---- toasts: bottom-right stack, night surface, Lucide icon, cyan progress bar. Same call shape as bulma-toast
  //      (toast({ message, type })) so upstream's handleRequest is untouched; richer events use notify().
  const TOAST_ICON = { 'is-success': 'check', 'is-warning': 'alert', 'is-danger': 'xcircle', 'is-info': 'info' };
  let toasts = [];
  let toastSeq = 0;
  function notify({ title, message, type = 'is-info', icon: ic, ms, link }) {
    const id = ++toastSeq;
    ms = ms || (type === 'is-danger' ? 12000 : 7000);
    toasts = [...toasts.slice(-3), { id, title, message, type, icon: ic || TOAST_ICON[type] || 'info', ms, link }];
    setTimeout(() => dismiss(id), ms);
  }
  function toast({ message, type }) {
    const m = /Txhash:\s*(0x[0-9a-fA-F]{64})/.exec(message || '');
    if (type === 'is-success' && m) {
      paid = true;
      step = 3;
      return notify({ title: `${faucetInfo.payout} ${faucetInfo.symbol} sent`, message: 'Transaction ' + m[1].slice(0, 10) + '… is on its way.', type, icon: 'droplet', ms: 12000, link: { href: `${CHAIN.explorer}/tx/${m[1]}`, text: 'View on explorer' } });
    }
    if (type === 'is-danger' && /rate limit|too many|wait|already|minutes|hour/i.test(message || '')) {
      return notify({ title: 'Rate limited', message, type: 'is-warning', icon: 'clock' });
    }
    notify({ title: type === 'is-danger' ? 'Request failed' : type === 'is-warning' ? 'Check your input' : 'Notice', message, type });
  }
  function dismiss(id) {
    toasts = toasts.filter((t) => t.id !== id);
  }

  // ---- live blocks from the public RPC (read-only): the last 24 blocks with their tx lists, so the chart can show
  //      txs per block and the faucet's own payouts in them; then one block per new head (eth_blockNumber every 6 s).
  const BLOCKS_KEEP = 24;
  let blockNumber = null;
  let blockBumped = false;
  let blockTimer = null;
  let blocks = []; // { n, ts, txs, payouts }
  let rpcId = 1;
  async function rpc(method, params) {
    const res = await fetch(CHAIN.rpc, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ jsonrpc: '2.0', id: rpcId++, method, params }),
    });
    const data = await res.json();
    if (data?.error) throw new Error(data.error.message || 'rpc error');
    return data?.result;
  }
  async function loadBlock(n) {
    const b = await rpc('eth_getBlockByNumber', ['0x' + n.toString(16), true]);
    if (!b) return null;
    const from = (faucetInfo.account || '').toLowerCase();
    const txs = b.transactions || [];
    return { n, ts: parseInt(b.timestamp, 16), txs: txs.length, payouts: txs.filter((t) => (t.from || '').toLowerCase() === from).length };
  }
  async function pollBlock() {
    try {
      const n = parseInt(await rpc('eth_blockNumber', []), 16);
      if (!Number.isFinite(n)) return;
      if (n !== blockNumber) {
        blockNumber = n;
        blockBumped = true;
        setTimeout(() => (blockBumped = false), 900);
      }
      const have = new Set(blocks.map((b) => b.n));
      const want = [];
      for (let i = Math.max(0, n - BLOCKS_KEEP + 1); i <= n; i++) if (!have.has(i)) want.push(i);
      if (want.length) {
        const fresh = (await Promise.all(want.slice(-BLOCKS_KEEP).map(loadBlock))).filter(Boolean);
        blocks = [...blocks, ...fresh].filter((b) => b.n > n - BLOCKS_KEEP).sort((a, b) => a.n - b.n);
      }
    } catch (error) {
      console.error('Block poll failed:', error);
    }
  }
  $: payouts24 = blocks.reduce((s, b) => s + b.payouts, 0);
  $: txs24 = blocks.reduce((s, b) => s + b.txs, 0);

  // ---- bar chart: hand SVG, bars tween to new heights (rAF, 600 ms ease-out); reduced motion applies at once
  const reduced = typeof window !== 'undefined' && window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  const CW = 560, CH = 170, ML = 30, MR = 6, MT = 10, MB = 22;
  let bars = []; // { n, x, w, y, h, pay, tip }
  let shown = {}; // n -> displayed { y, h }
  let raf = null;
  let tipHtml = '';
  let tipX = 0, tipY = 0, tipOn = false;
  $: layoutBars(blocks);
  function layoutBars(list) {
    const n = list.length;
    const max = Math.max(4, ...list.map((b) => b.txs));
    const span = CW - ML - MR, slot = span / Math.max(1, n), bw = Math.max(3, Math.min(22, slot - 3));
    const base = CH - MB;
    const target = list.map((b, i) => ({
      n: b.n, x: ML + slot * i + (slot - bw) / 2, w: bw, pay: b.payouts > 0, txs: b.txs, ts: b.ts, payouts: b.payouts,
      y: base - (b.txs ? Math.max(3, (CH - MT - MB) * b.txs / max) : 0), h: b.txs ? Math.max(3, (CH - MT - MB) * b.txs / max) : 0,
    }));
    gridMax = max;
    const from = target.map((t) => shown[t.n] || { y: base, h: 0 });
    if (raf) cancelAnimationFrame(raf);
    if (reduced || !n) { bars = target; shown = Object.fromEntries(target.map((t) => [t.n, { y: t.y, h: t.h }])); return; }
    const t0 = performance.now();
    const step = (now) => {
      const k = Math.min(1, (now - t0) / 600), e = 1 - Math.pow(1 - k, 3);
      bars = target.map((t, i) => ({ ...t, y: from[i].y + (t.y - from[i].y) * e, h: from[i].h + (t.h - from[i].h) * e }));
      if (k < 1) raf = requestAnimationFrame(step);
      else { raf = null; shown = Object.fromEntries(target.map((t) => [t.n, { y: t.y, h: t.h }])); }
    };
    raf = requestAnimationFrame(step);
  }
  let gridMax = 4;
  function showTip(e, b) {
    tipHtml = `<b>${b.txs} tx${b.txs === 1 ? '' : 's'}${b.payouts ? ` · ${b.payouts} faucet payout${b.payouts === 1 ? '' : 's'}` : ''}</b><br><span>block ${b.n.toLocaleString('en-US')} · ${new Date(b.ts * 1000).toLocaleTimeString()}</span>`;
    tipX = Math.min(e.clientX + 14, window.innerWidth - 240);
    tipY = e.clientY - 10;
    tipOn = true;
  }

  // ---- stepper: 1 add network, 2 request ETH, 3 open the app
  let step = 1;
  let networkAdded = false;
  let paid = false;

  // ---- add the chain to an injected wallet (EIP-3085)
  let addingChain = false;
  async function addToWallet() {
    const eth = typeof window !== 'undefined' ? window.ethereum : null;
    if (!eth || typeof eth.request !== 'function') {
      notify({ title: 'No wallet found', message: 'Install a browser wallet, or add the network by hand with the settings below.', type: 'is-warning', icon: 'wallet' });
      step = Math.max(step, 2);
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
      networkAdded = true;
      step = Math.max(step, 2);
      notify({ title: 'Network added', message: `${CHAIN.name} (chain ${CHAIN.id}) is in your wallet.`, type: 'is-success', icon: 'wallet' });
      try {
        const accounts = await eth.request({ method: 'eth_accounts' });
        if (accounts && accounts[0] && !input) input = accounts[0];
      } catch {}
    } catch (error) {
      if (error?.code !== 4001) {
        notify({ title: 'Wallet refused', message: error?.message || 'The wallet refused the network.', type: 'is-warning' });
      }
    } finally {
      addingChain = false;
    }
  }

  async function copyAccount() {
    try {
      await navigator.clipboard.writeText(faucetInfo.account);
      notify({ title: 'Address copied', message: faucetInfo.account, type: 'is-info', icon: 'copy', ms: 4000 });
    } catch {
      notify({ title: 'Could not copy', message: 'Select the address and copy it by hand.', type: 'is-warning' });
    }
  }

  onMount(async () => {
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
    pollBlock();
    blockTimer = setInterval(pollBlock, 6000);
  });
  onDestroy(() => {
    if (blockTimer) clearInterval(blockTimer);
    if (raf) cancelAnimationFrame(raf);
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

      notify({ title: 'Request sent', message: `Asking the faucet for ${faucetInfo.payout} ${faucetInfo.symbol}…`, type: 'is-info', icon: 'droplet', ms: 4000 });

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

  const fmtBlock = (n) => (n == null ? '—' : n.toLocaleString('en-US'));
  const yTicks = (max) => [0, 1, 2, 3, 4].map((i) => ({ v: Math.round(max * i / 4), y: MT + (CH - MT - MB) * (1 - i / 4) }));
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
        <a href={l.href} rel="noopener"><span class="ic">{@html icon(l.icon)}</span><span class="t">{l.name}</span></a>
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
        <p class="lead">One claim per address and per IP every 24 hours. Testnet ETH has no value.</p>

        <ol class="stepper" aria-label="Steps">
          <li class:done={step > 1} class:active={step === 1}><span class="n">{#if step > 1}{@html icon('check')}{:else}1{/if}</span><span>Add network</span></li>
          <li class:done={step > 2} class:active={step === 2}><span class="n">{#if step > 2}{@html icon('check')}{:else}2{/if}</span><span>Request ETH</span></li>
          <li class:active={step === 3}><span class="n">3</span><span>Open the app</span></li>
        </ol>

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
              onfocus={() => { if (step < 2) step = 2; }}
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
                <span class="ic">{@html icon('droplet')}</span>Request {faucetInfo.payout} {faucetInfo.symbol}<span class="arr">{@html icon('arrow')}</span>
              {/if}
            </button>
          </div>
          <p class="hint">
            Serving from
            <button type="button" class="addr num" onclick={copyAccount} title="Copy the faucet address"><span class="ic small">{@html icon('copy')}</span>{faucetInfo.account}</button>
            {#if faucetInfo.hcaptcha_sitekey}
              · protected by hCaptcha
            {/if}
          </p>
        </form>

        <div class="actions">
          <button type="button" class="btn secondary" onclick={addToWallet} disabled={addingChain}>
            <span class="ic">{@html icon(networkAdded ? 'check' : 'plus')}</span>{networkAdded ? 'Network added' : 'Add GullyChain Testnet to wallet'}<span class="arr">{@html icon('arrow')}</span>
          </button>
          <a class="btn ghost" class:glow={paid} href={LINKS[0].href} rel="noopener">
            <span class="ic">{@html icon('grid')}</span>Open the app<span class="arr">{@html icon('arrow')}</span>
          </a>
        </div>
      </div>
    </div>
  </section>

  <section class="wrap details">
    <figure class="card chart">
      <div class="chart-head">
        <p class="section-label"><span class="ic">{@html icon('boxes')}</span>Live · transactions per block</p>
        <span class="legend"><i class="sw"></i>txs <i class="sw pay"></i>with faucet payout</span>
      </div>
      <p class="muted small">Last {BLOCKS_KEEP} GullyChain blocks from the public RPC: <span class="num">{txs24}</span> txs, <span class="num">{payouts24}</span> faucet payout{payouts24 === 1 ? '' : 's'}. Hover a bar.</p>
      <svg viewBox="0 0 {CW} {CH}" role="img" aria-label="Transactions per block, last {BLOCKS_KEEP} blocks" onmouseleave={() => (tipOn = false)}>
        {#each yTicks(gridMax) as t}
          <line class="gl" x1={ML} x2={CW - MR} y1={t.y} y2={t.y} />
          <text class="at" x={ML - 6} y={t.y + 3} text-anchor="end">{t.v}</text>
        {/each}
        {#if bars.length}
          <text class="at" x={ML} y={CH - 6}>#{fmtBlock(bars[0].n)}</text>
          <text class="at" x={CW - MR} y={CH - 6} text-anchor="end">#{fmtBlock(bars[bars.length - 1].n)}</text>
        {:else}
          <text class="at empty" x={CW / 2} y={CH / 2} text-anchor="middle">Waiting for blocks…</text>
        {/if}
        {#each bars as b (b.n)}
          <rect class="bk" class:pay={b.pay} x={b.x} width={b.w} y={b.y} height={Math.max(0, b.h)} rx={Math.min(4, b.w / 2)} onmousemove={(e) => showTip(e, b)} />
        {/each}
      </svg>
    </figure>
    <div class="card network">
      <p class="section-label"><span class="ic">{@html icon('settings')}</span>Network settings</p>
      <dl>
        <div><dt>Network name</dt><dd>{CHAIN.name}</dd></div>
        <div><dt>Chain ID</dt><dd class="num">{CHAIN.id}</dd></div>
        <div><dt>RPC URL</dt><dd class="num">{CHAIN.rpc}</dd></div>
        <div><dt>Currency</dt><dd class="num">ETH</dd></div>
        <div><dt>Explorer</dt><dd class="num"><a href={CHAIN.explorer} rel="noopener">{CHAIN.explorer.replace('https://', '')}</a></dd></div>
        <div><dt>Latest block</dt><dd class="num">{fmtBlock(blockNumber)}</dd></div>
      </dl>
      <p class="muted small">Need more? Bridge Arbitrum Sepolia ETH through the <a href={LINKS[1].href} rel="noopener">testnet bridge</a>, or read the <a href={LINKS[2].href} rel="noopener">docs</a>.</p>
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

<div class="tip" class:on={tipOn} style="left:{tipX}px;top:{tipY}px" role="tooltip">{@html tipHtml}</div>

<div class="toasts" aria-live="polite" aria-atomic="false">
  {#each toasts as t (t.id)}
    <div class="toast {t.type}" role="status">
      <span class="ic">{@html icon(t.icon)}</span>
      <div class="body">
        <b>{t.title}</b>
        {#if t.message}<p>{t.message}{#if t.link} · <a href={t.link.href} rel="noopener">{t.link.text}</a>{/if}</p>{/if}
      </div>
      <button type="button" class="x" onclick={() => dismiss(t.id)} aria-label="Dismiss">{@html icon('x')}</button>
      <span class="bar" style="animation-duration:{t.ms}ms"></span>
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
  /* Lucide icons: 20px, stroke 1.75, muted by default, cyan on active/hover */
  .ic {
    display: inline-flex;
    width: 20px;
    height: 20px;
    flex: none;
    color: var(--gc-muted);
    vertical-align: middle;
  }
  .ic.small {
    width: 14px;
    height: 14px;
    margin-right: 4px;
  }
  .ic :global(svg) {
    width: 100%;
    height: 100%;
    fill: none;
    stroke: currentColor;
    stroke-width: 1.75;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
  .num {
    font-family: var(--f-num);
    font-weight: 500;
    font-variant-numeric: tabular-nums;
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
    gap: 2px;
  }
  .nav a {
    display: inline-flex;
    align-items: center;
    gap: 7px;
    color: var(--gc-muted);
    font-weight: 600;
    font-size: 15px;
    padding: 8px 12px;
    border-radius: var(--gc-radius-pill);
    transition: color 0.15s, background 0.15s;
  }
  .nav a:hover {
    color: var(--gc-text);
    background: var(--gc-night);
  }
  .nav a:hover .ic {
    color: var(--gc-cyan);
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
    padding: clamp(28px, 4.5vw, 52px) 0 clamp(24px, 3.5vw, 40px);
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
    margin: 0 0 10px;
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
    font-size: clamp(34px, 4.8vw, 56px);
    line-height: 1.02;
    margin: 0 0 12px;
    color: #ffffff;
  }
  h1 .hl {
    color: var(--gc-cyan);
    font-variant-numeric: tabular-nums;
  }
  .lead {
    color: var(--gc-muted);
    font-size: 16.5px;
    max-width: 540px;
    margin: 0 0 18px;
  }

  /* ---- stepper ---- */
  .stepper {
    list-style: none;
    margin: 0 0 16px;
    padding: 0;
    display: flex;
    flex-wrap: wrap;
    gap: 6px 0;
    max-width: 620px;
  }
  .stepper li {
    display: flex;
    align-items: center;
    gap: 8px;
    font-weight: 600;
    font-size: 14px;
    color: var(--gc-muted);
  }
  .stepper li + li::before {
    content: '';
    width: 28px;
    height: 1px;
    background: var(--gc-line);
    margin: 0 10px;
  }
  .stepper .n {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    border-radius: 50%;
    border: 1px solid var(--gc-line);
    font-family: var(--f-num);
    font-weight: 600;
    font-size: 13px;
    color: var(--gc-muted);
    transition: background 0.2s, color 0.2s, border-color 0.2s;
  }
  .stepper .n :global(svg) {
    width: 16px;
    height: 16px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
  .stepper li.active {
    color: #ffffff;
  }
  .stepper li.active .n {
    background: var(--gc-cyan);
    border-color: var(--gc-cyan);
    color: #000000;
  }
  .stepper li.done {
    color: #ffffff;
  }
  .stepper li.done .n {
    background: var(--gc-royal);
    border-color: var(--gc-royal);
    color: #ffffff;
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
  /* signature buttons: pill, left icon, arrow slides on hover, cyan shine sweep, scale .98 pressed */
  .btn {
    position: relative;
    overflow: hidden;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    min-height: 44px;
    padding: 0 20px;
    border-radius: var(--gc-radius-pill);
    border: 1px solid transparent;
    font-family: var(--f-ui);
    font-weight: 700;
    font-size: 15px;
    cursor: pointer;
    white-space: nowrap;
    transition: transform 0.12s, filter 0.15s, border-color 0.15s;
  }
  .btn .ic {
    color: currentColor;
  }
  .btn .arr {
    display: inline-flex;
    width: 18px;
    height: 18px;
    margin-left: 2px;
    transition: transform 0.2s;
  }
  .btn .arr :global(svg) {
    width: 18px;
    height: 18px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
  .btn:hover:not(:disabled) .arr {
    transform: translateX(4px);
  }
  .btn::after {
    content: '';
    position: absolute;
    top: 0;
    bottom: 0;
    left: -60%;
    width: 40%;
    background: linear-gradient(100deg, transparent 0%, rgba(1, 217, 250, 0.45) 50%, transparent 100%);
    transform: skewX(-20deg);
    opacity: 0;
    pointer-events: none;
  }
  .btn:hover:not(:disabled)::after {
    animation: shine 0.7s ease-out;
  }
  @keyframes shine {
    from {
      left: -60%;
      opacity: 1;
    }
    to {
      left: 120%;
      opacity: 0;
    }
  }
  .btn:active:not(:disabled) {
    transform: scale(0.98);
  }
  .btn:disabled {
    cursor: default;
    opacity: 0.7;
  }
  .btn.primary {
    background: var(--gc-cyan);
    color: #000000;
  }
  .btn.primary::after {
    background: linear-gradient(100deg, transparent 0%, rgba(255, 255, 255, 0.6) 50%, transparent 100%);
  }
  .btn.secondary {
    background: var(--gc-royal);
    color: #ffffff;
  }
  .btn.ghost {
    background: transparent;
    color: var(--gc-text);
    border-color: var(--gc-line);
  }
  .btn.ghost:hover {
    border-color: var(--gc-cyan);
  }
  .btn.ghost .ic {
    color: var(--gc-muted);
  }
  .btn.ghost:hover .ic,
  .btn.ghost.glow .ic {
    color: var(--gc-cyan);
  }
  .btn.ghost.glow {
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
    display: inline-flex;
    align-items: center;
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
  .addr:hover .ic {
    color: var(--gc-cyan);
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    margin-top: 16px;
  }

  /* ---- details: live chart + network settings ---- */
  .details {
    display: grid;
    grid-template-columns: 1.3fr 1fr;
    gap: 16px;
    padding-top: 4px;
    padding-bottom: 36px;
  }
  .chart {
    margin: 0;
  }
  .chart-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    flex-wrap: wrap;
  }
  .section-label {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 0 0 8px;
    font-family: var(--f-badge);
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-size: 14px;
    color: var(--gc-cyan);
  }
  .section-label .ic {
    color: var(--gc-cyan);
  }
  .legend {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 12.5px;
    color: var(--gc-muted);
    margin-bottom: 8px;
  }
  .legend .sw {
    width: 10px;
    height: 10px;
    border-radius: 3px;
    background: var(--gc-cyan);
    margin-left: 8px;
  }
  .legend .sw.pay {
    background: var(--gc-royal);
  }
  .chart svg {
    width: 100%;
    height: auto;
    overflow: visible;
    margin-top: 6px;
  }
  .chart .gl {
    stroke: var(--gc-line);
    stroke-width: 1;
  }
  .chart .at {
    fill: var(--gc-muted);
    font-family: var(--f-num);
    font-size: 10px;
  }
  .chart .at.empty {
    font-size: 13px;
    font-family: var(--f-ui);
  }
  .chart .bk {
    fill: var(--gc-cyan);
    transition: fill 0.15s;
  }
  .chart .bk.pay {
    fill: var(--gc-royal);
  }
  .chart .bk:hover {
    fill: #ffffff;
  }
  .tip {
    position: fixed;
    z-index: 60;
    pointer-events: none;
    background: var(--gc-night);
    border: 1px solid var(--gc-line);
    border-radius: 10px;
    padding: 7px 10px;
    font-size: 13px;
    color: var(--gc-text);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
    opacity: 0;
    transition: opacity 0.12s;
    white-space: nowrap;
  }
  .tip.on {
    opacity: 1;
  }
  .tip :global(b) {
    font-family: var(--f-num);
    font-weight: 600;
    font-variant-numeric: tabular-nums;
  }
  .tip :global(span) {
    color: var(--gc-muted);
  }
  dl {
    margin: 0;
    display: grid;
    gap: 0;
  }
  dl div {
    display: grid;
    grid-template-columns: 120px 1fr;
    gap: 12px;
    padding: 8px 0;
    border-top: 1px solid var(--gc-line);
    font-size: 14px;
  }
  dt {
    color: var(--gc-muted);
  }
  dd {
    margin: 0;
    color: var(--gc-text);
    overflow-wrap: anywhere;
  }
  .muted {
    color: var(--gc-muted);
    font-size: 14px;
    margin: 0;
  }
  .muted.small {
    font-size: 13px;
    margin: 0 0 4px;
  }
  .network .muted.small {
    margin: 12px 0 0;
  }

  /* ---- footer ---- */
  .foot {
    border-top: 1px solid var(--gc-line);
    padding: 20px 0 28px;
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

  /* ---- toasts: bottom-right stack, night surface, 22px radius, Lucide icon, cyan progress bar ---- */
  .toasts {
    position: fixed;
    right: 16px;
    bottom: 16px;
    z-index: 50;
    display: flex;
    flex-direction: column;
    gap: 10px;
    width: min(380px, calc(100vw - 32px));
  }
  .toast {
    position: relative;
    overflow: hidden;
    display: grid;
    grid-template-columns: auto 1fr auto;
    gap: 12px;
    align-items: start;
    padding: 14px 14px 16px 16px;
    border-radius: var(--gc-radius-card);
    border: 1px solid var(--gc-line);
    background: var(--gc-night);
    color: var(--gc-text);
    font-size: 14.5px;
    box-shadow: 0 12px 32px rgba(0, 0, 0, 0.5);
    animation: rise 0.25s ease-out;
  }
  .toast .ic {
    margin-top: 1px;
    color: var(--gc-cyan);
  }
  .toast.is-warning {
    border-color: var(--gc-royal);
  }
  .toast.is-warning .ic {
    color: #ffffff;
  }
  .toast.is-danger {
    border-color: var(--gc-down);
  }
  .toast.is-danger .ic {
    color: #ffffff;
  }
  .toast b {
    display: block;
    font-weight: 700;
    font-size: 14.5px;
  }
  .toast p {
    margin: 2px 0 0;
    color: var(--gc-muted);
    font-size: 13.5px;
    overflow-wrap: anywhere;
  }
  .toast .x {
    background: none;
    border: 0;
    color: var(--gc-muted);
    cursor: pointer;
    padding: 0;
    width: 20px;
    height: 20px;
  }
  .toast .x :global(svg) {
    width: 20px;
    height: 20px;
    fill: none;
    stroke: currentColor;
    stroke-width: 1.75;
    stroke-linecap: round;
  }
  .toast .x:hover {
    color: var(--gc-text);
  }
  .toast .bar {
    position: absolute;
    left: 0;
    bottom: 0;
    height: 3px;
    width: 100%;
    background: var(--gc-cyan);
    transform-origin: left;
    animation: toast-bar linear forwards;
  }
  @keyframes rise {
    from {
      opacity: 0;
      transform: translateY(8px);
    }
  }
  @keyframes toast-bar {
    from {
      transform: scaleX(1);
    }
    to {
      transform: scaleX(0);
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
    .nav a .t {
      display: none;
    }
    .nav a {
      padding: 8px 10px;
    }
    dl div {
      grid-template-columns: 1fr;
      gap: 2px;
    }
    .ticker {
      height: 32px;
      font-size: 13px;
    }
    .stepper li + li::before {
      width: 12px;
      margin: 0 6px;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .ticker .dot,
    .toast,
    .btn::after {
      animation: none !important;
    }
    .toast .bar {
      display: none;
    }
    .btn .arr,
    .btn {
      transition: none;
    }
  }
</style>
