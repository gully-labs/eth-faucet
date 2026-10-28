<script>
  import { onMount, onDestroy } from 'svelte';
  import { getAddress } from '@ethersproject/address';
  import GullySignIn from './GullySignIn.svelte';

  // GullyChain testnet faucet (gully-labs fork): brand kit v3 in the landing's "Terminal" language. The faucet logic,
  // flags and API below are upstream's; only the page changed.
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
  // Footer: every GullyChain product, same tab (one product family). Header nav is the <gc-switcher> element.
  const PRODUCTS = [
    { name: 'GullyChain', href: 'https://gullychain.xyz/' },
    { name: 'App', href: 'https://app.gullychain.xyz/' },
    { name: 'Bridge', href: 'https://testnet-bridge.gullychain.xyz/' },
    { name: 'Explorer', href: CHAIN.explorer + '/' },
    { name: 'Docs', href: 'https://docs.gullychain.xyz/' },
    { name: 'Developers', href: 'https://dev.gullychain.xyz/' },
    { name: 'Status', href: 'https://status.gullylabs.xyz/' },
  ];
  const SOURCE = 'https://github.com/gully-labs/eth-faucet';
  // Official GullyChain channels (every GullyChain footer carries them).
  const SOCIALS = [
    { name: 'Telegram', href: 'https://t.me/GullyChain' },
    { name: 'Deploys', href: 'https://t.me/GullyChainDeploys' },
    { name: 'Discord', href: 'https://discord.com/invite/tY2ErsNNBp' },
    { name: 'X', href: 'https://x.com/Gully_Labs' },
  ];

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
    radio: '<path d="M4.9 19.1C1 15.2 1 8.8 4.9 4.9"/><path d="M7.8 16.2c-2.3-2.3-2.3-6.1 0-8.5"/><circle cx="12" cy="12" r="2"/><path d="M16.2 7.8c2.3 2.3 2.3 6.1 0 8.5"/><path d="M19.1 4.9C23 8.8 23 15.1 19.1 19"/>',
    clock: '<circle cx="12" cy="12" r="10"/><path d="M12 6v6l4 2"/>',
    settings: '<path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/>',
  };
  const icon = (n) => `<svg viewBox="0 0 24 24" aria-hidden="true">${ICONS[n] || ICONS.info}</svg>`;

  let input = null;
  let fieldError = ''; // inline validation message under the address field
  // Sign in with Gully (only when /api/info has gullyid_client): fill the field with the main wallet, but never
  // overwrite an address the visitor typed; clear it again on sign-out only if it is still the filled-in one.
  let gullyFilled = null;
  let gullyToken = null; // session access token, sent as Authorization: Bearer when the faucet requires sign-in
  function onGullyWallet(address, user, session) {
    gullyToken = session?.accessToken || null;
    const current = (input || '').trim();
    if (address) {
      if (!current || current === gullyFilled) input = address;
      gullyFilled = address;
      if (step < 2) step = 2;
    } else {
      if (gullyFilled && current === gullyFilled) input = '';
      gullyFilled = null;
    }
  }
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

  // ---- toasts: bottom-right stack, ink surface, Lucide icon, cyan progress bar. Same call shape as bulma-toast
  //      (toast({ message, type })) so upstream's handleRequest stays as it is; richer events use notify().
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
  let blocks = []; // { n, ts, txs, payouts, baseFee, pays: [{to, value, hash}] }
  let balance = null; // faucet account balance, ETH
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
    const pays = txs.filter((t) => (t.from || '').toLowerCase() === from).map((t) => ({ to: t.to, value: parseInt(t.value, 16) / 1e18, hash: t.hash }));
    return { n, ts: parseInt(b.timestamp, 16), txs: txs.length, payouts: pays.length, pays, baseFee: parseInt(b.baseFeePerGas || '0x0', 16) / 1e9 };
  }
  async function pollBlock() {
    try {
      const n = parseInt(await rpc('eth_blockNumber', []), 16);
      if (!Number.isFinite(n)) return;
      if (n !== blockNumber) {
        blockNumber = n;
        blockBumped = true;
        setTimeout(() => (blockBumped = false), 900);
        if (mounted) rpc('eth_getBalance', [faucetInfo.account, 'latest']).then((h) => (balance = parseInt(h, 16) / 1e18)).catch(() => {});
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
  $: payoutsWin = blocks.reduce((s, b) => s + b.payouts, 0);
  $: txsWin = blocks.reduce((s, b) => s + b.txs, 0);
  $: lastBlock = blocks[blocks.length - 1];
  $: claimsLeft = balance != null && +faucetInfo.payout > 0 ? Math.floor(balance / +faucetInfo.payout + 1e-9) : null;
  $: events = blocks
    .flatMap((b) => b.pays.map((p) => ({ ...p, n: b.n, ts: b.ts })))
    .reverse()
    .slice(0, 12);

  // ---- bar chart: hand SVG, bars tween to new heights (rAF, 600 ms ease-out); reduced motion applies at once
  const reduced = typeof window !== 'undefined' && window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  const CW = 760, CH = 230, ML = 34, MR = 10, MT = 12, MB = 26;
  let bars = []; // { n, x, w, y, h, pay, txs, ts, payouts }
  let shown = {}; // n -> displayed { y, h }
  let raf = null;
  let gridMax = 4;
  let tipHtml = '';
  let tipX = 0, tipY = 0, tipOn = false;
  $: layoutBars(blocks);
  function layoutBars(list) {
    const n = list.length;
    const max = Math.max(4, ...list.map((b) => b.txs));
    const span = CW - ML - MR, slot = span / Math.max(1, n), bw = Math.max(3, Math.min(26, slot - 4));
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
    const stepFn = (now) => {
      const k = Math.min(1, (now - t0) / 600), e = 1 - Math.pow(1 - k, 3);
      bars = target.map((t, i) => ({ ...t, y: from[i].y + (t.y - from[i].y) * e, h: from[i].h + (t.h - from[i].h) * e }));
      if (k < 1) raf = requestAnimationFrame(stepFn);
      else { raf = null; shown = Object.fromEntries(target.map((t) => [t.n, { y: t.y, h: t.h }])); }
    };
    raf = requestAnimationFrame(stepFn);
  }
  function showTip(e, b) {
    tipHtml = `<b>${b.txs} tx${b.txs === 1 ? '' : 's'}${b.payouts ? ` · ${b.payouts} faucet payout${b.payouts === 1 ? '' : 's'}` : ''}</b><br><span>block ${b.n.toLocaleString('en-US')} · ${new Date(b.ts * 1000).toLocaleTimeString()}</span>`;
    tipX = Math.min(e.clientX + 14, window.innerWidth - 240);
    tipY = e.clientY - 10;
    tipOn = true;
  }
  // sparklines (200x40): blue, last mark green, like the landing's stat strip
  const sparkBars = (vals) => {
    const n = vals.length; if (!n) return [];
    const max = Math.max(1e-9, ...vals), slot = 200 / n, bw = Math.max(2, slot - 2);
    return vals.map((v, i) => ({ x: i * slot, w: bw, h: v > 0 ? Math.max(2, 34 * v / max) : 1, last: i === n - 1 }));
  };
  $: spTxs = sparkBars(blocks.slice(-20).map((b) => b.txs));
  $: spPay = sparkBars(blocks.map((b) => b.payouts));

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

  async function copyText(text, what) {
    try {
      await navigator.clipboard.writeText(text);
      notify({ title: what + ' copied', message: text, type: 'is-info', icon: 'copy', ms: 4000 });
    } catch {
      notify({ title: 'Could not copy', message: 'Select it and copy by hand.', type: 'is-warning' });
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
      rpc('eth_getBalance', [faucetInfo.account, 'latest']).then((h) => (balance = parseInt(h, 16) / 1e18)).catch(() => {});
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
    fieldError = '';
    if (!address) {
      fieldError = 'Enter an address (0x…) or an ENS name.';
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
            fieldError = 'That ENS name has no address.';
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
        fieldError = 'That is not a valid address: 42 characters, starting with 0x.';
        toast({ message: 'Invalid Ethereum address', type: 'is-warning' });
        return;
      }

      // Prepare request headers
      const headers = {
        'Content-Type': 'application/json',
      };
      if (faucetInfo.signin_required) {
        if (!gullyToken) {
          fieldError = 'Sign in with Gully first (button above): one claim per Gully ID a day.';
          toast({ message: 'Sign in with Gully to claim testnet ETH.', type: 'is-warning' });
          return;
        }
        headers['Authorization'] = 'Bearer ' + gullyToken;
      }

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
  const short = (a) => (a ? a.slice(0, 6) + '…' + a.slice(-4) : '');
  const ago = (ts) => { const s = Math.max(0, Date.now() / 1000 - ts); return s < 60 ? Math.round(s) + 's' : s < 3600 ? Math.round(s / 60) + 'm' : (s / 3600).toFixed(1) + 'h'; };
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

<header class="nav">
  <div class="wrap">
    <!-- GullyChain product switcher (public/gc-switcher.js, canonical copy in gully-ops sites/_shared/gc-switcher) -->
    <gc-switcher current="faucet"></gc-switcher>
    <div class="navr">
      <span class="pill" class:live={blockNumber != null} class:flash={blockBumped} role="status" aria-live="polite" title="Latest block from {CHAIN.rpc}"><span class="dot" aria-hidden="true"></span><span>{blockNumber == null ? 'Connecting' : 'Live · #' + fmtBlock(blockNumber)}</span></span>
      <button type="button" class="btn ter sm hide-s" onclick={addToWallet} disabled={addingChain}><span class="ic">{@html icon(networkAdded ? 'check' : 'wallet')}</span>Add to wallet</button>
    </div>
  </div>
</header>

<main>
  <section class="hero" aria-labelledby="h1">
    <div class="wrap">
      <div class="copy">
        <span class="badge">Testnet faucet · chain {CHAIN.id}</span>
        <h1 id="h1">Testnet ETH, <em>{faucetInfo.payout}&nbsp;{faucetInfo.symbol}</em> a day</h1>
        <p class="promise">{#if faucetInfo.signin_required}Sign in with Gully, then one claim per account, address and IP every 24 hours.{:else}One claim per address and per IP every 24 hours.{/if} <b>Testnet ETH has no value</b>; use it to try the app, deploy contracts and test the bridge.</p>

        <ol class="stepper" aria-label="Steps">
          <li class:done={step > 1} class:active={step === 1}><span class="n">{#if step > 1}{@html icon('check')}{:else}1{/if}</span><span>Add network</span></li>
          <li class:done={step > 2} class:active={step === 2}><span class="n">{#if step > 2}{@html icon('check')}{:else}2{/if}</span><span>Request ETH</span></li>
          <li class:active={step === 3}><span class="n">3</span><span>Open the app</span></li>
        </ol>

        <div id="hcaptcha" data-size="invisible"></div>
        <form id="claim" class="claim" onsubmit={(e) => { e.preventDefault(); if (!isLoading) handleRequest(); }}>
          {#if faucetInfo.gullyid_client}
            <GullySignIn clientId={faucetInfo.gullyid_client} issuer={faucetInfo.gullyid_issuer || undefined} required={!!faucetInfo.signin_required} onwallet={onGullyWallet} />
          {/if}
          <label class="lab" for="address">Your address or ENS name</label>
          <div class="field">
            <input
              id="address"
              bind:value={input}
              class="input mono"
              type="text"
              autocomplete="off"
              spellcheck="false"
              placeholder="0x… or name.eth"
              aria-invalid={fieldError ? 'true' : undefined}
              aria-describedby={fieldError ? 'address-err' : undefined}
              oninput={() => (fieldError = '')}
              onfocus={() => { if (step < 2) step = 2; }}
              onkeydown={(e) => {
                if (e.key === 'Enter' && !isLoading) {
                  e.preventDefault();
                  handleRequest();
                }
              }}
            />
            <button type="submit" class="btn pri" disabled={isLoading} aria-busy={isLoading}>
              {#if isLoading}
                <span class="spinner" aria-hidden="true"></span>Sending
              {:else}
                <span class="ic">{@html icon('droplet')}</span>Request {faucetInfo.payout} {faucetInfo.symbol}<span class="ic ar">{@html icon('arrow')}</span>
              {/if}
            </button>
          </div>
          {#if fieldError}<p id="address-err" class="ferr">{fieldError}</p>{/if}
          <p class="hint">Serving from <button type="button" class="addr mono" onclick={() => copyText(faucetInfo.account, 'Faucet address')} title="Copy the faucet address">{faucetInfo.account}</button>{#if faucetInfo.hcaptcha_sitekey} · protected by hCaptcha{/if}</p>
          {#if faucetInfo.eligibility_check}<p class="hint">Send it to your Gully wallet, or to an address with at least one transaction on Gully Chain or Arbitrum Sepolia. Brand-new empty addresses are refused (bots farm those).</p>{/if}
        </form>

        <div class="ctas">
          <button type="button" class="btn sec" onclick={addToWallet} disabled={addingChain}>
            <span class="ic">{@html icon(networkAdded ? 'check' : 'plus')}</span>{networkAdded ? 'Network added' : 'Add GullyChain Testnet'}<span class="ic ar">{@html icon('arrow')}</span>
          </button>
          <a class="btn ter" class:glow={paid} href={LINKS[0].href} target="_blank" rel="noopener noreferrer">
            <span class="ic">{@html icon('grid')}</span>Open the app<span class="ic ar">{@html icon('arrow')}</span>
          </a>
        </div>
        <div class="rpcrow"><span class="lab">RPC</span><code>{CHAIN.rpc}</code><button class="btn ter xs" type="button" onclick={() => copyText(CHAIN.rpc, 'RPC URL')}><span class="ic">{@html icon('copy')}</span>Copy</button></div>
      </div>

      <div class="term wall" aria-labelledby="h-wall">
        <div class="th"><span class="lab cy" id="h-wall"><span class="ic sm">{@html icon('radio')}</span>Transactions per block · last {BLOCKS_KEEP} blocks</span><div class="r">{#if blocks.length}<span>#{fmtBlock(blocks[0].n)} → #{fmtBlock(lastBlock.n)}</span>{/if}</div></div>
        <div class="hn"><b class="num">{txsWin}</b><span>transactions in the window</span><span class="num">{payoutsWin} faucet payout{payoutsWin === 1 ? '' : 's'}</span>{#if lastBlock}<span class="num">base fee {lastBlock.baseFee.toFixed(2)} gwei</span>{/if}</div>
        <div class="plot">
          <svg class="main" viewBox="0 0 {CW} {CH}" role="img" aria-label="Transactions per block, last {BLOCKS_KEEP} blocks" onmouseleave={() => (tipOn = false)}>
            {#each yTicks(gridMax) as t}
              <line class="gl" x1={ML} x2={CW - MR} y1={t.y} y2={t.y} />
              <text class="at" x={ML - 6} y={t.y + 4} text-anchor="end">{t.v}</text>
            {/each}
            {#if bars.length}
              <text class="at" x={ML} y={CH - 8}>#{fmtBlock(bars[0].n)}</text>
              <text class="at" x={CW - MR} y={CH - 8} text-anchor="end">#{fmtBlock(bars[bars.length - 1].n)}</text>
            {:else}
              <text class="at empty" x={CW / 2} y={CH / 2} text-anchor="middle">Waiting for blocks…</text>
            {/if}
            {#each bars as b (b.n)}
              <rect class="bk" class:pay={b.pay} x={b.x} width={b.w} y={b.y} height={Math.max(0, b.h)} rx={Math.min(3, b.w / 2)} onmousemove={(e) => showTip(e, b)} />
            {/each}
          </svg>
        </div>
        <div class="tape" aria-label="Latest faucet payouts"><span class="lb lab">Faucet</span><div class="trk" class:static={!events.length}>
          {#if events.length}
            {#each events as e (e.hash)}
              <span class="it"><span class="ic">{@html icon('droplet')}</span><b>{e.value} ETH</b> to <a class="mono" href="{CHAIN.explorer}/address/{e.to}" target="_blank" rel="noopener noreferrer">{short(e.to)}</a> · #{fmtBlock(e.n)} · {ago(e.ts)} ago</span>
            {/each}
          {:else}
            <span class="it">No payouts in the last {BLOCKS_KEEP} blocks{#if lastBlock} · block #{fmtBlock(lastBlock.n)} · {ago(lastBlock.ts)} ago{/if}</span>
          {/if}
        </div></div>
      </div>
    </div>
  </section>

  <div class="wrap" aria-label="Live numbers">
    <div class="stats">
      <div class="tile"><span class="lab">Latest block</span><div class="v num">{fmtBlock(blockNumber)}</div><svg class="spark" viewBox="0 0 200 40" preserveAspectRatio="none">{#each spTxs as b}<rect x={b.x} y={40 - b.h} width={b.w} height={b.h} class:last={b.last} />{/each}</svg><span class="sub">txs per block, last 20</span></div>
      <div class="tile"><span class="lab">Transactions</span><div class="v num">{txsWin}<small>{BLOCKS_KEEP} blocks</small></div><span class="sub">{lastBlock ? lastBlock.txs + ' in the latest block' : 'waiting for blocks'}</span></div>
      <div class="tile"><span class="lab">Faucet payouts</span><div class="v num">{payoutsWin}<small>{BLOCKS_KEEP} blocks</small></div><svg class="spark" viewBox="0 0 200 40" preserveAspectRatio="none">{#each spPay as b}<rect x={b.x} y={40 - b.h} width={b.w} height={b.h} class:last={b.last} />{/each}</svg><span class="sub">sent from <span class="mono">{short(faucetInfo.account)}</span></span></div>
      <div class="tile" class:warn={claimsLeft != null && claimsLeft < 20}><span class="lab">Claims left</span><div class="v num">{claimsLeft == null ? '—' : claimsLeft.toLocaleString('en-US')}</div><span class="sub">{balance == null ? 'reading balance…' : balance.toFixed(4) + ' ETH / ' + faucetInfo.payout + ' ETH per claim'}</span></div>
      <div class="tile"><span class="lab">Base fee</span><div class="v num">{lastBlock ? lastBlock.baseFee.toFixed(2) : '—'}<small>gwei</small></div><span class="sub">latest block</span></div>
    </div>
  </div>

  <section class="wrap more">
    <div class="group settings">
      <p class="lab cy"><span class="ic">{@html icon('settings')}</span>Network settings</p>
      <dl>
        <div><dt>Network name</dt><dd>{CHAIN.name}</dd></div>
        <div><dt>Chain ID</dt><dd class="num">{CHAIN.id}</dd></div>
        <div><dt>RPC URL</dt><dd class="mono">{CHAIN.rpc}</dd></div>
        <div><dt>Currency</dt><dd>ETH</dd></div>
        <div><dt>Explorer</dt><dd class="mono"><a href={CHAIN.explorer} target="_blank" rel="noopener noreferrer">{CHAIN.explorer.replace('https://', '')}</a></dd></div>
      </dl>
    </div>
    <div class="group band">
      <div class="band-copy">
        <p class="lab cy"><span class="ic">{@html icon('arrows')}</span>Need more than {faucetInfo.payout} {faucetInfo.symbol}?</p>
        <h2>Bridge from Arbitrum Sepolia</h2>
        <p class="muted">Deposit Arbitrum Sepolia ETH through the testnet bridge, or read how the testnet works in the docs.</p>
        <div class="ctas">
          <a class="btn ter sm" href={LINKS[1].href} target="_blank" rel="noopener noreferrer"><span class="ic">{@html icon('arrows')}</span>Bridge<span class="ic ar">{@html icon('arrow')}</span></a>
          <a class="btn ter sm" href={LINKS[2].href} target="_blank" rel="noopener noreferrer"><span class="ic">{@html icon('book')}</span>Docs<span class="ic ar">{@html icon('arrow')}</span></a>
        </div>
      </div>
      <div class="mascot" aria-hidden="true">
        <img src="/brand/mascots/gullychain-mascot-builder.png" width="370" height="385" alt="" loading="lazy" />
        <div class="fade"></div>
      </div>
    </div>
  </section>
</main>

<footer>
  <div class="wrap">
    <div><a class="fwm" href="https://gullychain.xyz" target="_blank" rel="noopener noreferrer" aria-label="GullyChain home (opens in a new tab)"><img src="/brand/logo/gullychain-logo-transparent.png" width="656" height="152" alt="GullyChain" loading="lazy" /></a><span class="tag">Testnet faucet · chain {CHAIN.id}</span></div>
    <nav aria-label="GullyChain products"><ul>{#each PRODUCTS as l}<li><a href={l.href}>{l.name}</a></li>{/each}<li><a href={SOURCE} target="_blank" rel="noopener noreferrer">Source</a></li></ul></nav>
    <ul aria-label="Official GullyChain channels">{#each SOCIALS as l}<li><a href={l.href} target="_blank" rel="noopener noreferrer">{l.name}</a></li>{/each}</ul>
    <div class="fbot">
      <p class="copy2">Testnet only. Tokens have no value. MIT, a fork of chainflag/eth-faucet.</p>
      <a class="glb" href="https://gullylabs.xyz" target="_blank" rel="noopener noreferrer" aria-label="Powered by Gully Labs (opens in a new tab)"><img src="/brand/gully-labs/powered-by-gully-labs-dark.png" srcset="/brand/gully-labs/powered-by-gully-labs-dark.png 1x, /brand/gully-labs/powered-by-gully-labs-dark-2x.png 2x, /brand/gully-labs/powered-by-gully-labs-dark-3x.png 3x" width="127" height="32" alt="Powered by Gully Labs" loading="lazy" /></a>
    </div>
  </div>
</footer>

<div class="tip" class:on={tipOn} style="left:{tipX}px;top:{tipY}px" role="tooltip">{@html tipHtml}</div>

<div class="toasts" role="region" aria-label="Notifications" aria-live="polite">
  {#each toasts as t (t.id)}
    <div class="toast {t.type}" role="status">
      <span class="ic">{@html icon(t.icon)}</span>
      <div class="body">
        <b>{t.title}</b>
        {#if t.message}<p>{t.message}{#if t.link} · <a href={t.link.href} target="_blank" rel="noopener noreferrer">{t.link.text}</a>{/if}</p>{/if}
      </div>
      <button type="button" class="x" onclick={() => dismiss(t.id)} aria-label="Dismiss">{@html icon('x')}</button>
      <span class="pb" style="animation-duration:{t.ms}ms"></span>
    </div>
  {/each}
</div>

<style>
  /* Brand kit v3 in the landing's "Terminal" language: night page, ink cards with 1px lines, square terminal frame for
     the chart, stat strip with sparklines, event tape, 22px groups, cyan uppercase labels, Exo 2 600/800, tabular
     numbers, hashes in system monospace, the landing's pill buttons and Lucide nav icons. */
  .wrap { width: 100%; max-width: 1360px; margin: 0 auto; padding: 0 16px; }
  .skip { position: absolute; left: -999px; top: 8px; background: var(--gc-cyan); color: var(--gc-night); padding: 8px 14px; border-radius: var(--gc-radius-pill); font-weight: 800; z-index: 100; }
  .skip:focus { left: 16px; }
  .lab { font-weight: 800; text-transform: uppercase; letter-spacing: 0.05em; font-size: 12.5px; color: var(--gc-muted); display: inline-flex; align-items: center; gap: 6px; margin: 0; }
  .lab.cy { color: var(--gc-cyan); }
  .num { font-variant-numeric: tabular-nums; }
  .mono { font-family: var(--f-mono); font-size: 0.95em; }
  .muted { color: var(--gc-muted); margin: 0; }
  h1, h2 { font-weight: 800; letter-spacing: -0.01em; margin: 0; line-height: 1.05; }
  /* Lucide icons */
  .ic { display: inline-flex; width: 20px; height: 20px; flex: none; color: var(--gc-muted); vertical-align: middle; }
  .ic :global(svg) { width: 100%; height: 100%; fill: none; stroke: currentColor; stroke-width: 1.75; stroke-linecap: round; stroke-linejoin: round; }
  .ic.sm { width: 14px; height: 14px; }
  .lab .ic { width: 16px; height: 16px; color: currentColor; }

  /* nav: kit logo at 40px (logo/gullychain-logo-transparent.png, width auto) */
  .nav { position: sticky; top: 0; z-index: 20; background: rgba(5, 8, 13, 0.88); backdrop-filter: blur(12px); -webkit-backdrop-filter: blur(12px); border-bottom: 1px solid var(--gc-line); }
  .nav .wrap { display: flex; align-items: center; justify-content: space-between; height: 64px; gap: 14px; }
  .navr { display: flex; align-items: center; gap: 10px; flex: none; }
  .pill { display: inline-flex; align-items: center; gap: 8px; height: 32px; padding: 0 12px; border-radius: var(--gc-radius-pill); border: 1px solid var(--gc-line); background: var(--gc-ink); font-size: 13px; font-weight: 600; color: var(--gc-muted); white-space: nowrap; font-variant-numeric: tabular-nums; }
  .pill .dot { width: 8px; height: 8px; border-radius: 50%; background: var(--gc-muted); flex: none; }
  .pill.live { color: var(--gc-white); }
  .pill.live .dot { background: var(--gc-green); box-shadow: 0 0 0 0 rgba(70, 251, 4, 0.5); animation: breathe 2.4s ease-in-out infinite; }
  .pill.flash .dot { animation: ping 0.9s ease-out 1; }
  @keyframes breathe { 50% { box-shadow: 0 0 0 6px rgba(70, 251, 4, 0); } }
  @keyframes ping { from { box-shadow: 0 0 0 0 rgba(70, 251, 4, 0.6); } to { box-shadow: 0 0 0 10px rgba(70, 251, 4, 0); } }

  /* signature buttons */
  .btn { position: relative; overflow: hidden; isolation: isolate; display: inline-flex; align-items: center; justify-content: center; gap: 9px; min-height: 44px; padding: 0 20px 0 18px; border-radius: var(--gc-radius-pill); font-family: var(--gc-font); font-weight: 800; font-size: 15px; text-decoration: none; border: 1px solid transparent; cursor: pointer; white-space: nowrap; transition: transform 0.12s ease, box-shadow 0.15s ease, border-color 0.15s ease; }
  .btn .ic { width: 18px; height: 18px; color: currentColor; }
  .btn .ar { width: 16px; height: 16px; transition: transform 0.2s ease; }
  .btn:hover:not(:disabled) .ar { transform: translateX(4px); }
  .btn:active:not(:disabled) { transform: scale(0.98); }
  .btn::after { content: ''; position: absolute; inset: 0; z-index: -1; background: linear-gradient(115deg, transparent 35%, rgba(255, 255, 255, 0.45) 50%, transparent 65%); transform: translateX(-120%); }
  .btn:hover:not(:disabled)::after { transform: translateX(120%); transition: transform 0.7s ease; }
  .btn:disabled { cursor: default; opacity: 0.7; }
  .btn.pri { background: var(--gc-cyan); color: var(--gc-night); }
  .btn.pri:hover:not(:disabled) { box-shadow: 0 0 0 4px rgba(0, 178, 250, 0.22); }
  .btn.sec { background: var(--gc-blue); color: var(--gc-white); }
  .btn.sec::after { background: linear-gradient(115deg, transparent 35%, rgba(0, 178, 250, 0.5) 50%, transparent 65%); }
  .btn.sec:hover:not(:disabled) { box-shadow: 0 0 0 4px rgba(37, 99, 255, 0.25); }
  .btn.ter { background: transparent; color: var(--gc-white); border-color: var(--gc-line); }
  .btn.ter::after { background: linear-gradient(115deg, transparent 35%, rgba(0, 178, 250, 0.25) 50%, transparent 65%); }
  .btn.ter:hover:not(:disabled), .btn.ter.glow { border-color: var(--gc-cyan); }
  .btn.ter .ic { color: var(--gc-muted); }
  .btn.ter:hover .ic, .btn.ter.glow .ic { color: var(--gc-cyan); }
  .btn.sm { min-height: 36px; padding: 0 14px 0 12px; font-size: 14px; }
  .btn.xs { min-height: 30px; padding: 0 10px; font-size: 13px; gap: 6px; color: var(--gc-muted); }
  .spinner { width: 16px; height: 16px; border-radius: 50%; border: 2px solid rgba(5, 8, 13, 0.25); border-top-color: var(--gc-night); animation: spin 0.8s linear infinite; }
  @keyframes spin { to { transform: rotate(360deg); } }

  /* hero: copy + chart wall */
  .hero { padding: 26px 0 18px; position: relative; }
  .hero::before { content: ''; position: absolute; inset: 0; background: radial-gradient(900px 400px at 85% 0%, rgba(37, 99, 255, 0.16), transparent 70%); pointer-events: none; }
  .hero .wrap { display: grid; grid-template-columns: 1fr; gap: 18px; position: relative; }
  .copy { display: flex; flex-direction: column; gap: 14px; min-width: 0; }
  .badge { display: inline-flex; align-items: center; gap: 8px; height: 30px; padding: 0 12px; border-radius: var(--gc-radius-pill); background: var(--gc-blue); color: var(--gc-white); font-weight: 800; text-transform: uppercase; letter-spacing: 0.05em; font-size: 13px; width: max-content; }
  .copy h1 { font-size: clamp(36px, 4.2vw, 56px); }
  .copy h1 em { font-style: normal; background: var(--gc-chain-gradient); -webkit-background-clip: text; background-clip: text; color: transparent; font-variant-numeric: tabular-nums; }
  .copy .promise { color: var(--gc-muted); font-size: clamp(15px, 1.15vw, 17px); max-width: 480px; margin: 0; }
  .copy .promise b { color: var(--gc-white); font-weight: 800; }
  .ctas { display: flex; flex-wrap: wrap; gap: 10px; }
  .rpcrow { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; font-size: 13px; color: var(--gc-muted); }
  .rpcrow code { font-family: var(--f-mono); color: var(--gc-white); background: var(--gc-ink); border: 1px solid var(--gc-line); padding: 4px 8px; border-radius: 6px; font-size: 12.5px; overflow-wrap: anywhere; }

  /* stepper */
  .stepper { list-style: none; margin: 0; padding: 0; display: flex; flex-wrap: wrap; gap: 6px 0; }
  .stepper li { display: flex; align-items: center; gap: 8px; font-weight: 600; font-size: 14px; color: var(--gc-muted); }
  .stepper li + li::before { content: ''; width: 24px; height: 1px; background: var(--gc-line); margin: 0 10px; }
  .stepper .n { display: inline-flex; align-items: center; justify-content: center; width: 28px; height: 28px; border-radius: 50%; border: 1px solid var(--gc-line); font-weight: 800; font-size: 13px; color: var(--gc-muted); font-variant-numeric: tabular-nums; transition: background 0.2s, color 0.2s, border-color 0.2s; }
  .stepper .n :global(svg) { width: 16px; height: 16px; fill: none; stroke: currentColor; stroke-width: 2; stroke-linecap: round; stroke-linejoin: round; }
  .stepper li.active { color: var(--gc-white); }
  .stepper li.active .n { background: var(--gc-cyan); border-color: var(--gc-cyan); color: var(--gc-night); }
  .stepper li.done { color: var(--gc-white); }
  .stepper li.done .n { background: var(--gc-green); border-color: var(--gc-green); color: var(--gc-night); }

  /* claim form (ink 22px card) */
  .claim { margin: 0; background: var(--gc-ink); border: 1px solid var(--gc-line); border-radius: var(--gc-radius-card); padding: 18px 20px; display: flex; flex-direction: column; gap: 10px; }
  .claim .lab { font-size: 12px; }
  .field { display: flex; gap: 10px; }
  .input { flex: 1 1 auto; min-width: 0; min-height: 44px; padding: 0 16px; border-radius: var(--gc-radius-pill); border: 1px solid var(--gc-line); background: var(--gc-night); color: var(--gc-white); font-size: 14px; outline: none; transition: border-color 0.15s; }
  .input::placeholder { color: var(--gc-muted); opacity: 0.8; font-family: var(--gc-font); }
  .input:focus { border-color: var(--gc-cyan); }
  .input:focus-visible { box-shadow: 0 0 0 2px var(--gc-cyan); }
  .input[aria-invalid='true'] { border-color: #ff6b6b; }
  .ferr { margin: 0; font-size: 13px; color: #ff8a8a; }
  .hint { margin: 0; font-size: 13px; color: var(--gc-muted); overflow-wrap: anywhere; }
  .addr { background: none; border: 0; padding: 0; color: var(--gc-white); font-size: 12.5px; cursor: copy; border-bottom: 1px dashed var(--gc-line); overflow-wrap: anywhere; word-break: break-all; text-align: left; }
  .addr:hover { border-bottom-color: var(--gc-cyan); }

  /* terminal frame + chart */
  .term { border: 1px solid var(--gc-line); background: var(--gc-ink); position: relative; min-width: 0; display: flex; flex-direction: column; }
  .term .th { display: flex; align-items: center; justify-content: space-between; gap: 10px; padding: 8px 12px; border-bottom: 1px solid var(--gc-line); min-height: 38px; }
  .term .th .lab { font-size: 12.5px; }
  .term .th .r { display: flex; align-items: center; gap: 8px; font-size: 12px; color: var(--gc-muted); font-variant-numeric: tabular-nums; }
  .wall .hn { display: flex; align-items: baseline; gap: 10px; flex-wrap: wrap; padding: 12px 12px 0; }
  .wall .hn b { font-weight: 800; font-size: clamp(30px, 3vw, 44px); letter-spacing: -0.01em; line-height: 1; }
  .wall .hn span { color: var(--gc-muted); font-size: 13px; }
  .wall .plot { padding: 10px 12px 6px; position: relative; flex: 1; }
  .wall svg.main { width: 100%; height: auto; display: block; overflow: visible; }
  .gl { stroke: var(--gc-line); stroke-width: 1; }
  .at { fill: var(--gc-muted); font-family: var(--gc-font); font-weight: 600; font-size: 11px; }
  .at.empty { font-size: 13px; }
  .bk { fill: var(--gc-cyan); transition: fill 0.15s; }
  .bk.pay { fill: var(--gc-green); }
  .bk:hover { fill: var(--gc-white); }
  .tip { position: fixed; z-index: 60; pointer-events: none; background: var(--gc-ink); border: 1px solid var(--gc-line); border-radius: 10px; padding: 7px 10px; font-size: 13px; color: var(--gc-white); box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5); opacity: 0; transition: opacity 0.12s; white-space: nowrap; }
  .tip.on { opacity: 1; }
  .tip :global(b) { font-weight: 800; font-variant-numeric: tabular-nums; }
  .tip :global(span) { color: var(--gc-muted); }
  /* tape */
  .tape { border-top: 1px solid var(--gc-line); height: 38px; display: flex; align-items: center; overflow: hidden; position: relative; background: var(--gc-night); }
  .tape::before, .tape::after { content: ''; position: absolute; top: 0; bottom: 0; width: 40px; z-index: 2; pointer-events: none; }
  .tape::before { left: 0; background: linear-gradient(90deg, var(--gc-night), transparent); }
  .tape::after { right: 0; background: linear-gradient(270deg, var(--gc-night), transparent); }
  .tape .lb { position: absolute; left: 0; top: 0; bottom: 0; z-index: 3; display: flex; align-items: center; padding: 0 12px; background: var(--gc-night); border-right: 1px solid var(--gc-line); color: var(--gc-cyan); }
  .tape .trk { display: flex; white-space: nowrap; animation: tape 60s linear infinite; padding-left: 100%; will-change: transform; }
  .tape .trk.static { animation: none; padding-left: 90px; }
  .tape:hover .trk, .tape:focus-within .trk { animation-play-state: paused; }
  .tape .it { display: inline-flex; align-items: center; gap: 8px; padding: 0 22px; font-size: 13px; color: var(--gc-muted); font-variant-numeric: tabular-nums; border-right: 1px solid var(--gc-line); }
  .tape .it b { color: var(--gc-white); font-weight: 800; }
  .tape .it .ic { width: 14px; height: 14px; color: var(--gc-green); }
  @keyframes tape { to { transform: translateX(-100%); } }

  /* stat strip */
  .stats { display: grid; grid-template-columns: repeat(2, 1fr); gap: 1px; background: var(--gc-line); border: 1px solid var(--gc-line); border-top: 0; }
  .tile { padding: 12px 14px 10px; background: var(--gc-ink); display: flex; flex-direction: column; gap: 6px; min-width: 0; position: relative; overflow: hidden; }
  .tile .lab { font-size: 12px; }
  .tile .v { font-weight: 800; font-size: clamp(22px, 2vw, 30px); letter-spacing: -0.01em; line-height: 1; font-variant-numeric: tabular-nums; }
  .tile .v small { font-weight: 600; font-size: 12px; color: var(--gc-muted); letter-spacing: 0; margin-left: 6px; }
  .tile .sub { font-size: 12px; color: var(--gc-muted); font-variant-numeric: tabular-nums; }
  .tile svg.spark { width: 100%; height: 40px; display: block; overflow: visible; }
  .tile.warn { background: rgba(245, 166, 35, 0.08); }
  .tile.warn .v { color: var(--st-warn); }
  .spark rect { fill: var(--gc-blue); }
  .spark rect.last { fill: var(--gc-green); }

  /* settings + bridge band (ink 22px groups); the kit mascot at its native 370x385, uncropped, on night */
  .more { display: grid; grid-template-columns: 1fr; gap: 16px; padding-top: 16px; padding-bottom: 36px; }
  .group { background: var(--gc-ink); border: 1px solid var(--gc-line); border-radius: var(--gc-radius-card); padding: 20px; min-width: 0; }
  .settings .lab { margin-bottom: 8px; }
  dl { margin: 0; }
  dl div { display: grid; grid-template-columns: 120px 1fr; gap: 12px; padding: 8px 0; border-top: 1px solid var(--gc-line); font-size: 14px; }
  dt { color: var(--gc-muted); }
  dd { margin: 0; color: var(--gc-white); overflow-wrap: anywhere; }
  .band { position: relative; overflow: hidden; display: flex; align-items: center; min-height: 425px; background: var(--gc-night); } /* 385px mascot + padding: never cropped */
  .band-copy { position: relative; z-index: 2; display: flex; flex-direction: column; gap: 8px; max-width: 420px; }
  .band h2 { font-size: clamp(22px, 2.4vw, 30px); }
  .band .muted { font-size: 14px; }
  .band .ctas { margin-top: 6px; }
  .mascot { position: absolute; inset: 0 0 0 auto; width: 370px; max-width: 60%; pointer-events: none; display: flex; align-items: center; justify-content: flex-end; }
  .mascot img { width: 370px; height: 385px; max-width: none; }
  .mascot .fade { position: absolute; inset: 0; background: linear-gradient(90deg, var(--gc-night) 0%, rgba(5, 8, 13, 0.6) 25%, transparent 55%); }

  /* footer */
  footer { border-top: 1px solid var(--gc-line); padding: 22px 0; background: var(--gc-night); }
  footer .wrap { display: flex; flex-wrap: wrap; justify-content: space-between; gap: 12px; align-items: center; }
  footer ul { list-style: none; margin: 0; padding: 0; display: flex; flex-wrap: wrap; gap: 14px; font-size: 13.5px; }
  footer ul a { color: var(--gc-muted); }
  footer ul a:hover { color: var(--gc-white); }
  footer .tag { font-weight: 600; letter-spacing: 0.05em; font-size: 12px; color: var(--gc-muted); margin-top: 6px; display: block; text-transform: uppercase; }
  footer .copy2 { font-size: 12.5px; color: var(--gc-muted); width: 100%; margin: 0; }
  footer .fwm { display: inline-flex; line-height: 0; }
  footer .fwm img { height: 36px; width: auto; }
  footer .fbot { width: 100%; display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 10px 16px; padding-top: 12px; border-top: 1px solid var(--gc-line); }
  footer .fbot .copy2 { width: auto; }
  .glb { display: inline-flex; line-height: 0; flex: none; border-radius: 8px; }
  .glb img { height: 32px; width: auto; }
  .glb:focus-visible { outline: 2px solid var(--gc-cyan); outline-offset: 3px; }
  @media (max-width: 640px) { .glb img { height: 26px; } }

  /* toasts */
  .toasts { position: fixed; right: 16px; bottom: 16px; z-index: 60; display: flex; flex-direction: column; gap: 10px; width: min(360px, calc(100vw - 32px)); pointer-events: none; }
  .toast { pointer-events: auto; position: relative; overflow: hidden; display: grid; grid-template-columns: auto 1fr auto; gap: 12px; align-items: start; background: var(--gc-ink); border: 1px solid var(--gc-line); border-radius: 22px; padding: 14px 14px 16px 16px; box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5); animation: rise 0.25s ease-out; }
  .toast .ic { color: var(--gc-cyan); margin-top: 1px; }
  .toast.is-success .ic { color: var(--gc-green); }
  .toast.is-warning { border-color: var(--st-warn); }
  .toast.is-warning .ic { color: var(--st-warn); }
  .toast.is-danger { border-color: var(--st-down); }
  .toast.is-danger .ic { color: var(--gc-white); }
  .toast b { display: block; font-weight: 800; font-size: 15px; }
  .toast p { display: block; color: var(--gc-muted); font-size: 13px; margin: 2px 0 0; word-break: break-word; }
  .toast .x { all: unset; cursor: pointer; color: var(--gc-muted); width: 24px; height: 24px; display: grid; place-items: center; border-radius: 50%; }
  .toast .x :global(svg) { width: 18px; height: 18px; fill: none; stroke: currentColor; stroke-width: 1.75; stroke-linecap: round; }
  .toast .x:hover { color: var(--gc-white); }
  .toast .pb { position: absolute; left: 16px; right: 16px; bottom: 0; height: 3px; border-radius: 2px; background: var(--gc-cyan); transform-origin: left; animation: pb linear forwards; }
  @keyframes rise { from { opacity: 0; transform: translateY(10px); } }
  @keyframes pb { from { transform: scaleX(1); } to { transform: scaleX(0); } }

  @media (min-width: 700px) { .wrap { padding: 0 28px; } .more { grid-template-columns: minmax(0, 2fr) minmax(0, 3fr); } }
  @media (min-width: 900px) { .hero .wrap { grid-template-columns: minmax(0, 2fr) minmax(0, 3fr); gap: 28px; align-items: stretch; } .stats { grid-template-columns: repeat(5, 1fr); } }
  @media (max-width: 899px) { .mascot { opacity: 0.45; } }
  @media (max-width: 640px) { footer ul a { display: inline-flex; align-items: center; min-height: 40px; } .btn.xs { min-height: 40px; } .addr { padding: 8px 0; } .hide-s { display: none; } .field { flex-direction: column; } .btn.pri { width: 100%; } .ctas .btn { flex: 1 1 auto; } .tape .trk { animation: none; padding-left: 90px; overflow-x: auto; } dl div { grid-template-columns: 1fr; gap: 2px; } .stepper li + li::before { width: 12px; margin: 0 6px; } }
  @media (prefers-reduced-motion: reduce) { .pill .dot, .toast, .btn::after, .spinner { animation: none !important; } .toast .pb { display: none; } .tape .trk { animation: none; padding-left: 90px; overflow-x: auto; } .btn, .btn .ar { transition: none; } }
</style>
