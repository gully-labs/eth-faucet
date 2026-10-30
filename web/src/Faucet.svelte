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
    { name: 'App', href: 'https://kre8.fun/', icon: 'grid' },
    { name: 'Bridge', href: 'https://testnet-bridge.gullychain.xyz/', icon: 'arrows' },
    { name: 'Docs', href: 'https://docs.gullychain.xyz/', icon: 'book' },
    { name: 'Explorer', href: CHAIN.explorer + '/', icon: 'search' },
  ];
  // Footer: every GullyChain product, same tab (one product family). Header nav is the <gc-switcher> element.
  const PRODUCTS = [
    { name: 'GullyChain', href: 'https://gullychain.xyz/' },
    { name: 'App', href: 'https://kre8.fun/' },
    { name: 'Bridge', href: 'https://testnet-bridge.gullychain.xyz/' },
    { name: 'Explorer', href: CHAIN.explorer + '/' },
    { name: 'Sentry', href: 'https://sentry.gullychain.xyz/' },
    { name: 'Docs', href: 'https://docs.gullychain.xyz/' },
    { name: 'Developers', href: 'https://dev.gullychain.xyz/' },
    { name: 'Status', href: 'https://status.gullylabs.xyz/' },
  ];
  const SOURCE = 'https://github.com/gully-labs/eth-faucet';
  // Every other Gully product is listed once, on the docs ecosystem page (owner rule 2026-10-29: Gully Chain nav only).
  const ECOSYSTEM = 'https://docs.gullychain.xyz/ecosystem/';
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
  function toast({ message, type, amount, tier }) {
    const m = /Txhash:\s*(0x[0-9a-fA-F]{64})/.exec(message || '');
    if (type === 'is-success' && m) {
      paid = true;
      step = 3;
      return notify({ title: `${amount || faucetInfo.payout} ${faucetInfo.symbol} sent${tier === 'developer' ? ' (developer tier)' : ''}`, message: 'Transaction ' + m[1].slice(0, 10) + '… is on its way.', type, icon: 'droplet', ms: 12000, link: { href: `${CHAIN.explorer}/tx/${m[1]}`, text: 'View on explorer' } });
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
  // the last successful claim: ETH tx + the token drip (glUSD) result, shown under the form
  let claimResult = null; // { ethTx, token: { symbol, amount, sent, tx, msg } | null }

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
        toast({ message, type: 'is-success', amount: data?.amount, tier: data?.tier });
        // Gully Chain analytics (self-hosted Umami, cookieless): a coarse counter, no address
        try { window.umami?.track('faucet-claim', data?.tier ? { tier: String(data.tier) } : undefined); } catch {}
        const ethTx = (/Txhash:\s*(0x[0-9a-fA-F]{64})/.exec(message) || [])[1] || null;
        claimResult = { ethTx, token: data?.token || null };
        if (data?.token?.sent) {
          notify({ title: `${data.token.amount} ${data.token.symbol} sent`, message: 'Test stablecoin ' + data.token.tx.slice(0, 10) + '… is on its way.', type: 'is-success', icon: 'droplet', ms: 12000, link: { href: `${CHAIN.explorer}/tx/${data.token.tx}`, text: 'View on explorer' } });
        } else if (data?.token?.msg) {
          notify({ title: `No ${data.token.symbol} this time`, message: data.token.msg, type: 'is-warning' });
        }
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
    <div class="brand">
      <!-- brand kit v3 logo/gullychain-logo-transparent.png (1315x305, byte for byte) on night; the switcher hides its own
           logo (logo="off") so the mark shows once. gc-switcher: public/gc-switcher.js, canonical gully-ops sites/_shared/gc-switcher -->
      <a class="logo" href="https://gullychain.xyz/" aria-label="Gully Chain home"><img src="/brand/logo/gullychain-logo-transparent.png" width="155" height="36" alt="Gully Chain" /></a>
      <gc-switcher current="faucet" theme="dark" logo="off"></gc-switcher>
    </div>
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
        <p class="promise">{#if faucetInfo.signin_required}Sign in with Gully, then one claim per account, address and IP every 24 hours.{:else}One claim per address and per IP every 24 hours.{/if} {#if faucetInfo.token}Each claim also sends <b>{faucetInfo.token.amount} {faucetInfo.token.symbol}</b> (test stablecoin) while the daily {faucetInfo.token.symbol} budget lasts, for Gully Pay checkout and swaps. {/if}<b>Testnet ETH has no value</b>; use it to try the app, deploy contracts and test the bridge.</p>

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
          {#if claimResult}
            <div class="claim-result" role="status" data-testid="claim-result">
              <p><span class="ic ok">{@html icon('check')}</span>{faucetInfo.payout} {faucetInfo.symbol} sent{#if claimResult.ethTx} · <a class="mono" href={`${CHAIN.explorer}/tx/${claimResult.ethTx}`} target="_blank" rel="noopener noreferrer">{short(claimResult.ethTx)}</a>{/if}</p>
              {#if claimResult.token}
                {#if claimResult.token.sent}
                  <p data-testid="token-sent"><span class="ic ok">{@html icon('check')}</span>{claimResult.token.amount} {claimResult.token.symbol} sent · <a class="mono" href={`${CHAIN.explorer}/tx/${claimResult.token.tx}`} target="_blank" rel="noopener noreferrer">{short(claimResult.token.tx)}</a></p>
                {:else}
                  <p data-testid="token-skipped" class="muted"><span class="ic">{@html icon('info')}</span>{claimResult.token.msg}</p>
                {/if}
              {/if}
            </div>
          {/if}
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
        {#if faucetInfo.dev_tier}<p class="muted" data-testid="dev-tier">Building on GullyChain? Verified developers get <b>{faucetInfo.dev_tier.payout} {faucetInfo.symbol}</b> every {faucetInfo.dev_tier.window_hours} h with the same Gully ID sign-in. Email <a href="mailto:team@gullylabs.xyz?subject=Faucet%20developer%20tier">team@gullylabs.xyz</a> with your Gully ID and what you are building.</p>{/if}
        <div class="ctas">
          <a class="btn ter sm" href={LINKS[1].href} target="_blank" rel="noopener noreferrer"><span class="ic">{@html icon('arrows')}</span>Bridge<span class="ic ar">{@html icon('arrow')}</span></a>
          <a class="btn ter sm" href={LINKS[2].href} target="_blank" rel="noopener noreferrer"><span class="ic">{@html icon('book')}</span>Docs<span class="ic ar">{@html icon('arrow')}</span></a>
        </div>
      </div>
      <div class="mascot" aria-hidden="true">
        <!-- kit mascot (byte for byte) on the night ground -->
        <img src="/brand/mascots/gullychain-mascot-builder.png" width="370" height="385" alt="" loading="lazy" />
        <div class="fade"></div>
      </div>
    </div>
  </section>
</main>

<footer>
  <div class="wrap">
    <div><a class="fwm" href="https://gullychain.xyz"><img class="fwm-t" src="/brand/logo/gullychain-wordmark-for-dark-bg.svg" width="141" height="24" alt="Gully Chain" loading="lazy" /></a><span class="tag">Testnet faucet · chain {CHAIN.id}</span></div>
    <nav aria-label="GullyChain products"><ul>{#each PRODUCTS as l}<li><a href={l.href}>{l.name}</a></li>{/each}<li><a href={SOURCE} target="_blank" rel="noopener noreferrer">Source</a></li><li><a href={ECOSYSTEM} target="_blank" rel="noopener noreferrer">Gully ecosystem →</a></li></ul></nav>
    <ul aria-label="Official GullyChain channels">{#each SOCIALS as l}<li><a href={l.href} target="_blank" rel="noopener noreferrer">{l.name}</a></li>{/each}</ul>
    <div class="fbot">
      <p class="copy2">Testnet only. Tokens have no value. MIT, a fork of chainflag/eth-faucet.</p>
      <span class="fbadges"><a class="gcb" href="https://gullychain.xyz" aria-label="Powered by GullyChain"><img src="/brand/badges/gullychain-powered-by-dark.svg" srcset="/brand/badges/gullychain-powered-by-dark.png 1x, /brand/badges/gullychain-powered-by-dark-2x.png 2x, /brand/badges/gullychain-powered-by-dark-3x.png 3x" width="142" height="42" alt="Powered by GullyChain" loading="lazy" /></a>
      <a class="glb" href="https://gullylabs.xyz" target="_blank" rel="noopener noreferrer" aria-label="Powered by Gully Labs (opens in a new tab)"><img src="/brand/gully-labs/powered-by-gully-labs-dark.png" srcset="/brand/gully-labs/powered-by-gully-labs-dark.png 1x, /brand/gully-labs/powered-by-gully-labs-dark-2x.png 2x, /brand/gully-labs/powered-by-gully-labs-dark-3x.png 3x" width="127" height="32" alt="Powered by Gully Labs" loading="lazy" /></a></span>
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
  /* Gully Chain Night (GN, gully-ops sites/_shared/gullychain-night): black ground, ink cards with a muted 18% line,
     Exo 2 600/800, cyan accent, green for live and payouts, signature pill buttons, live animated charts (bars grow,
     sparklines grow) from the public RPC only, the builder mascot on night. Every colour is a --gn-* token (aliased
     in App.svelte); motion is off under prefers-reduced-motion. */
  .wrap { width: 100%; max-width: var(--gn-maxw); margin: 0 auto; padding: 0 var(--gn-gutter); }
  .skip { position: absolute; left: -999px; top: 8px; background: var(--gn-cyan); color: var(--gn-on-cyan); padding: 10px 16px; border-radius: var(--gn-r-pill); font-weight: var(--gn-w-bold); z-index: 100; }
  .skip:focus { left: 16px; }
  .lab { font-weight: var(--gn-w-bold); font-size: var(--gn-fs-xs); color: var(--muted-c); display: inline-flex; align-items: center; gap: 6px; margin: 0; text-transform: uppercase; letter-spacing: var(--gn-track-label); }
  .lab.cy { color: var(--gn-cyan); }
  .lab.cy .ic { color: var(--gn-cyan); }
  .num { font-variant-numeric: tabular-nums; }
  .mono { font-family: var(--gn-font-mono); font-weight: 400; font-size: 0.95em; letter-spacing: 0; }
  .muted { color: var(--text-2); margin: 0; }
  h1, h2 { font-weight: var(--gn-w-bold); letter-spacing: var(--gn-track-tight); margin: 0; line-height: var(--gn-lh-tight); color: var(--text); }
  b { font-weight: var(--gn-w-bold); }
  /* Lucide icons */
  .ic { display: inline-flex; width: 18px; height: 18px; flex: none; color: var(--muted-c); vertical-align: middle; }
  .ic :global(svg) { width: 100%; height: 100%; fill: none; stroke: currentColor; stroke-width: 1.75; stroke-linecap: round; stroke-linejoin: round; }
  .ic.sm { width: 14px; height: 14px; }
  .lab .ic { width: 16px; height: 16px; color: currentColor; }

  /* header: kit logo + switcher, live block pill */
  .nav { position: sticky; top: 0; z-index: 20; background: color-mix(in srgb, var(--gn-black) 80%, transparent); backdrop-filter: saturate(140%) blur(14px); -webkit-backdrop-filter: saturate(140%) blur(14px); border-bottom: 1px solid var(--border); }
  .nav .wrap { display: flex; align-items: center; justify-content: space-between; height: var(--gn-header-h); gap: 14px; }
  .brand { display: flex; align-items: center; gap: 10px; min-width: 0; }
  .logo { display: inline-flex; flex: none; line-height: 0; border-radius: var(--gn-r-sm); }
  .logo img { height: 36px; width: auto; max-width: none; }
  .navr { display: flex; align-items: center; gap: 10px; flex: none; }
  .pill { display: inline-flex; align-items: center; gap: 8px; height: 32px; padding: 0 12px; border-radius: var(--gn-r-pill); border: 1px solid var(--border); background: var(--gn-surface); font-size: 13px; color: var(--muted-c); white-space: nowrap; font-variant-numeric: tabular-nums; }
  .pill .dot { width: 8px; height: 8px; border-radius: 50%; background: var(--muted-c); flex: none; }
  .pill.live { color: var(--text); }
  .pill.live .dot { background: var(--gn-live); box-shadow: 0 0 0 0 color-mix(in srgb, var(--gn-live) 55%, transparent); animation: breathe 2.4s ease-in-out infinite; }
  .pill.flash .dot { animation: ping 0.9s ease-out 1; }
  @keyframes breathe { 50% { box-shadow: 0 0 0 6px color-mix(in srgb, var(--gn-live) 0%, transparent); } }
  @keyframes ping { from { box-shadow: 0 0 0 0 color-mix(in srgb, var(--gn-live) 60%, transparent); } to { box-shadow: 0 0 0 10px color-mix(in srgb, var(--gn-live) 0%, transparent); } }

  /* signature pill buttons (GN button tokens): pri cyan / night text, sec blue / white, ter = ghost line */
  .btn { position: relative; overflow: hidden; isolation: isolate; display: inline-flex; align-items: center; justify-content: center; gap: 8px; min-height: var(--gn-btn-h); padding: 0 var(--gn-btn-px) 0 18px; border-radius: var(--gn-btn-r); font-family: var(--gn-font); font-weight: var(--gn-w-bold); font-size: 15px; text-decoration: none; border: 1px solid transparent; cursor: pointer; white-space: nowrap; transition: transform var(--gn-dur) var(--gn-ease), background var(--gn-dur), border-color var(--gn-dur); }
  .btn .ic { width: 18px; height: 18px; color: currentColor; }
  .btn .ar { width: 16px; height: 16px; transition: transform var(--gn-dur) var(--gn-ease); }
  .btn:hover:not(:disabled) .ar { transform: translateX(4px); }
  .btn:active:not(:disabled) { transform: scale(0.98); }
  .btn:focus-visible { outline: none; box-shadow: var(--gn-focus); }
  .btn::after { content: ''; position: absolute; inset: 0; z-index: -1; background: linear-gradient(110deg, transparent 30%, color-mix(in srgb, var(--gn-white) 35%, transparent) 50%, transparent 70%); transform: translateX(-120%); pointer-events: none; }
  .btn:hover:not(:disabled)::after { transform: translateX(120%); transition: transform var(--gn-dur-slow) var(--gn-ease); }
  .btn:disabled { cursor: default; opacity: 0.6; }
  .btn.pri { background: var(--gn-btn-pri-bg); color: var(--gn-btn-pri-fg); }
  .btn.pri:hover:not(:disabled) { background: var(--gn-btn-pri-hover); }
  .btn.sec { background: var(--gn-btn-sec-bg); color: var(--gn-btn-sec-fg); }
  .btn.sec:hover:not(:disabled) { background: var(--gn-btn-sec-hover); }
  .btn.ter { background: transparent; color: var(--gn-btn-ghost-fg); border-color: var(--gn-btn-ghost-bd); }
  .btn.ter::after { background: linear-gradient(110deg, transparent 30%, var(--gn-info-bg) 50%, transparent 70%); }
  .btn.ter:hover:not(:disabled), .btn.ter.glow { border-color: var(--gn-muted); background: var(--gn-btn-ghost-hover); }
  .btn.ter.glow { box-shadow: var(--gn-glow); }
  .btn.ter .ic { color: var(--muted-c); }
  .btn.ter:hover .ic, .btn.ter.glow .ic { color: var(--gn-cyan); }
  .btn.sm { min-height: var(--gn-btn-h-sm); padding: 0 16px 0 14px; font-size: 14px; }
  .btn.xs { min-height: 30px; padding: 0 12px; font-size: 13px; gap: 6px; color: var(--text-2); }
  .spinner { width: 16px; height: 16px; border-radius: 50%; border: 2px solid color-mix(in srgb, var(--gn-on-cyan) 30%, transparent); border-top-color: var(--gn-on-cyan); animation: spin 0.8s linear infinite; }
  @keyframes spin { to { transform: rotate(360deg); } }

  /* hero band (true black): copy + chart card */
  .hero { padding: 32px 0 24px; position: relative; background: var(--gn-black); border-bottom: 1px solid var(--border); }
  .hero .wrap { display: grid; grid-template-columns: 1fr; gap: 20px; position: relative; }
  .copy { display: flex; flex-direction: column; gap: 14px; min-width: 0; }
  .badge { display: inline-flex; align-items: center; gap: 8px; height: 28px; padding: 0 12px; border-radius: var(--gn-r-pill); background: var(--info-bg); border: 1px solid var(--gn-line-accent); color: var(--info); font-weight: var(--gn-w-bold); font-size: var(--gn-fs-xs); text-transform: uppercase; letter-spacing: 0.1em; width: max-content; max-width: 100%; }
  .copy h1 { font-size: clamp(32px, 3.6vw, 52px); }
  .copy h1 em { font-style: normal; font-variant-numeric: tabular-nums; background: var(--gn-gradient); -webkit-background-clip: text; background-clip: text; color: transparent; }
  .copy .promise { color: var(--text-2); font-size: 15px; max-width: 540px; margin: 0; line-height: var(--gn-lh); }
  .copy .promise b { color: var(--text); }
  .ctas { display: flex; flex-wrap: wrap; gap: 10px; }
  .rpcrow { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; font-size: 13px; color: var(--muted-c); }
  .rpcrow code { font-family: var(--gn-font-mono); color: var(--text); background: var(--surface-2); border: 1px solid var(--border); padding: 4px 10px; border-radius: var(--gn-r-sm); font-size: 12.5px; overflow-wrap: anywhere; }

  /* stepper */
  .stepper { list-style: none; margin: 0; padding: 0; display: flex; flex-wrap: wrap; gap: 6px 0; }
  .stepper li { display: flex; align-items: center; gap: 8px; font-size: 14px; color: var(--muted-c); }
  .stepper li + li::before { content: ''; width: 24px; height: 1px; background: var(--border-2); margin: 0 10px; }
  .stepper .n { display: inline-flex; align-items: center; justify-content: center; width: 28px; height: 28px; border-radius: 50%; border: 1px solid var(--border-2); background: var(--surface); font-weight: var(--gn-w-bold); font-size: 13px; color: var(--muted-c); font-variant-numeric: tabular-nums; transition: background var(--gn-dur), color var(--gn-dur), border-color var(--gn-dur); }
  .stepper .n :global(svg) { width: 15px; height: 15px; fill: none; stroke: currentColor; stroke-width: 2.2; stroke-linecap: round; stroke-linejoin: round; }
  .stepper li.active { color: var(--text); font-weight: var(--gn-w-bold); }
  .stepper li.active .n { background: var(--gn-cyan); border-color: var(--gn-cyan); color: var(--gn-on-cyan); }
  .stepper li.done { color: var(--text); }
  .stepper li.done .n { background: var(--ok-bg); border-color: transparent; color: var(--ok); }

  /* claim form card */
  .claim { margin: 0; background: var(--surface); border: 1px solid var(--border); border-radius: var(--gn-r-lg); padding: 18px; display: flex; flex-direction: column; gap: 10px; }
  .claim .lab { font-size: var(--gn-fs-xs); }
  .field { display: flex; gap: 8px; }
  .input { flex: 1 1 auto; min-width: 0; min-height: var(--gn-btn-h); padding: 0 18px; border-radius: var(--gn-r-pill); border: 1px solid var(--border-2); background: var(--surface-2); color: var(--text); font-size: 14px; outline: none; transition: border-color var(--gn-dur), box-shadow var(--gn-dur); }
  .input::placeholder { color: var(--muted-c); opacity: 1; font-family: var(--gn-font); }
  .input:focus { border-color: var(--gn-cyan); box-shadow: 0 0 0 3px var(--gn-info-bg); }
  .input[aria-invalid='true'] { border-color: var(--danger); }
  .ferr { margin: 0; font-size: 13px; color: var(--danger); }
  .hint { margin: 0; font-size: 13px; color: var(--muted-c); overflow-wrap: anywhere; }
  .addr { background: none; border: 0; padding: 0; color: var(--text-2); font-size: 12.5px; cursor: copy; border-bottom: 1px dashed var(--border-2); overflow-wrap: anywhere; word-break: break-all; text-align: left; }
  .addr:hover { border-bottom-color: var(--gn-cyan); color: var(--link); }

  /* chart card */
  .term { border: 1px solid var(--border); border-radius: var(--gn-r-lg); background: var(--surface); position: relative; min-width: 0; display: flex; flex-direction: column; overflow: hidden; }
  .term .th { display: flex; align-items: center; justify-content: space-between; gap: 10px; padding: 14px 16px 0; min-height: 34px; }
  .term .th .r { display: flex; align-items: center; gap: 8px; font-size: 12px; color: var(--muted-c); font-variant-numeric: tabular-nums; }
  .wall .hn { display: flex; align-items: baseline; gap: 10px; flex-wrap: wrap; padding: 8px 16px 0; }
  .wall .hn b { font-size: var(--gn-fs-stat); letter-spacing: -0.02em; line-height: 1; }
  .wall .hn span { color: var(--muted-c); font-size: 13px; }
  .wall .plot { padding: 8px 16px 6px; position: relative; flex: 1; }
  .wall svg.main { width: 100%; height: auto; display: block; overflow: visible; }
  .gl { stroke: var(--gn-grid); stroke-width: 1; }
  .at { fill: var(--muted-c); font-family: var(--gn-font); font-weight: var(--gn-w-regular); font-size: 11px; }
  .at.empty { font-size: 13px; }
  .bk { fill: var(--gn-c1); transition: fill var(--gn-dur); transform-box: fill-box; transform-origin: bottom; animation: grow 0.8s var(--gn-ease) both; }
  .bk.pay { fill: var(--gn-c2); }
  .bk:hover { fill: var(--gn-link-hover); }
  @keyframes grow { from { transform: scaleY(0); } to { transform: scaleY(1); } }
  .tip { position: fixed; z-index: 60; pointer-events: none; background: var(--gn-surface-2); border: 1px solid var(--border-2); border-radius: var(--gn-r); padding: 8px 12px; font-size: 13px; color: var(--text); box-shadow: var(--gn-shadow-pop); opacity: 0; transition: opacity 0.12s; white-space: nowrap; }
  .tip.on { opacity: 1; }
  .tip :global(b) { font-weight: var(--gn-w-bold); font-variant-numeric: tabular-nums; }
  .tip :global(span) { color: var(--muted-c); }
  /* tape */
  .tape { border-top: 1px solid var(--border); height: 40px; display: flex; align-items: center; overflow: hidden; position: relative; background: var(--gn-night); }
  .tape::before, .tape::after { content: ''; position: absolute; top: 0; bottom: 0; width: 40px; z-index: 2; pointer-events: none; }
  .tape::before { left: 0; background: linear-gradient(90deg, var(--gn-night), transparent); }
  .tape::after { right: 0; background: linear-gradient(270deg, var(--gn-night), transparent); }
  .tape .lb { position: absolute; left: 0; top: 0; bottom: 0; z-index: 3; display: flex; align-items: center; padding: 0 12px; background: var(--gn-night); border-right: 1px solid var(--border); color: var(--text); }
  .tape .trk { display: flex; white-space: nowrap; animation: tape 60s linear infinite; padding-left: 100%; will-change: transform; }
  .tape .trk.static { animation: none; padding-left: 90px; }
  .tape:hover .trk, .tape:focus-within .trk { animation-play-state: paused; }
  .tape .it { display: inline-flex; align-items: center; gap: 8px; padding: 0 20px; font-size: 13px; color: var(--text-2); font-variant-numeric: tabular-nums; border-right: 1px solid var(--border); }
  .tape .it b { color: var(--text); }
  .tape .it .ic { width: 14px; height: 14px; color: var(--gn-green); }
  @keyframes tape { to { transform: translateX(-100%); } }

  /* stat tiles */
  .stats { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; margin-top: 24px; }
  .tile { padding: 16px 16px 12px; background: var(--surface); border: 1px solid var(--border); border-radius: var(--gn-r-lg); display: flex; flex-direction: column; gap: 6px; min-width: 0; position: relative; overflow: hidden; transition: border-color var(--gn-dur), box-shadow var(--gn-dur); }
  .tile:hover { border-color: var(--gn-line-accent); box-shadow: var(--gn-glow); }
  .tile .v { font-weight: var(--gn-w-bold); font-size: 26px; letter-spacing: -0.02em; line-height: 1.1; font-variant-numeric: tabular-nums; color: var(--text); }
  .tile .v small { font-weight: var(--gn-w-regular); font-size: 12px; color: var(--muted-c); letter-spacing: 0; margin-left: 6px; }
  .tile .sub { font-size: 12px; color: var(--muted-c); font-variant-numeric: tabular-nums; }
  .tile svg.spark { width: 100%; height: 32px; display: block; overflow: visible; }
  .tile.warn { background: var(--warn-bg); border-color: color-mix(in srgb, var(--gn-warn) 30%, transparent); }
  .tile.warn .v { color: var(--warn); }
  .spark rect { fill: var(--gn-c1); opacity: 0.85; transform-box: fill-box; transform-origin: bottom; animation: grow 0.8s var(--gn-ease) both; }
  .spark rect.last { fill: var(--gn-c2); opacity: 1; }

  /* settings + bridge band; the builder mascot on night */
  .more { display: grid; grid-template-columns: 1fr; gap: 14px; margin: 14px auto 40px; }
  .group { background: var(--surface); border: 1px solid var(--border); border-radius: var(--gn-r-lg); padding: 18px; min-width: 0; }
  .settings .lab { margin-bottom: 8px; }
  dl { margin: 0; }
  dl div { display: grid; grid-template-columns: 120px 1fr; gap: 12px; padding: 8px 0; border-top: 1px solid var(--border); font-size: 14px; }
  dt { color: var(--muted-c); }
  dd { margin: 0; color: var(--text); overflow-wrap: anywhere; }
  .band { display: grid; grid-template-columns: minmax(0, 1fr) 220px; gap: 16px; align-items: stretch; }
  .band-copy { display: flex; flex-direction: column; gap: 10px; min-width: 0; }
  .band h2 { font-size: 24px; }
  .band .muted { font-size: 14px; }
  .band .ctas { margin-top: 4px; }
  .mascot { position: relative; min-height: 200px; border-radius: var(--gn-r); background: var(--gn-night); overflow: hidden; pointer-events: none; }
  .mascot img { position: absolute; right: 0; bottom: 0; height: 100%; max-height: 230px; width: auto; max-width: none; transition: transform 0.3s ease; }
  .group.band:hover .mascot img { transform: scale(1.04) rotate(-2deg); }
  .mascot .fade { display: none; }

  /* footer (true black band) */
  footer { border-top: 1px solid var(--border); padding: 32px 0 24px; background: var(--gn-black); color: var(--muted-c); }
  footer .wrap { display: flex; flex-wrap: wrap; justify-content: space-between; gap: 14px; align-items: center; }
  footer ul { list-style: none; margin: 0; padding: 0; display: flex; flex-wrap: wrap; gap: 16px; font-size: 14px; }
  footer ul a { color: var(--text-2); }
  footer ul a:hover { color: var(--text); }
  footer .tag { font-size: 12px; color: var(--muted-c); margin-top: 8px; display: block; }
  footer .copy2 { font-size: 12px; color: var(--muted-c); width: 100%; margin: 0; }
  /* brand kit v3 logo/gullychain-wordmark-for-dark-bg.svg (byte for byte) */
  footer .fwm { display: inline-flex; align-items: center; line-height: 0; border-radius: var(--gn-r-sm); }
  footer .fwm img { display: block; flex: none; max-width: none; }
  footer .fwm .fwm-t { width: auto; height: 24px; }
  .fbadges { display: inline-flex; align-items: center; gap: 12px; flex-wrap: wrap; }
  .gcb { display: inline-flex; line-height: 0; flex: none; border-radius: var(--gn-r-sm); }
  .gcb img { width: 142px; height: 42px; }
  footer .fbot { width: 100%; display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 10px 16px; padding-top: 14px; border-top: 1px solid var(--border); }
  footer .fbot .copy2 { width: auto; }
  .glb { display: inline-flex; line-height: 0; flex: none; border-radius: var(--gn-r-sm); }
  .glb img { height: 28px; width: auto; }
  @media (max-width: 640px) { .glb img { height: 26px; } }

  /* claim result (ETH + token drip) */
  .claim-result { margin-top: 8px; padding: 10px 14px; border: 1px solid color-mix(in srgb, var(--gn-green) 28%, transparent); border-radius: var(--gn-r); background: var(--ok-bg); font-size: 14px; display: grid; gap: 6px; }
  .claim-result p { margin: 0; display: flex; align-items: center; gap: 8px; flex-wrap: wrap; word-break: break-word; }
  .claim-result .ic.ok { color: var(--ok); }
  .claim-result a { color: var(--link); }
  .claim-result .muted { color: var(--text-2); }

  /* toasts */
  .toasts { position: fixed; right: 16px; bottom: 16px; z-index: 60; display: flex; flex-direction: column; gap: 8px; width: min(360px, calc(100vw - 32px)); pointer-events: none; }
  .toast { pointer-events: auto; position: relative; overflow: hidden; display: grid; grid-template-columns: auto 1fr auto; gap: 10px; align-items: start; background: var(--gn-surface-2); border: 1px solid var(--border-2); border-radius: var(--gn-r); padding: 12px 12px 14px 14px; box-shadow: var(--gn-shadow-pop); animation: rise 0.25s ease-out; }
  .toast .ic { color: var(--gn-cyan); margin-top: 1px; }
  .toast.is-success .ic { color: var(--ok); }
  .toast.is-warning { border-color: color-mix(in srgb, var(--gn-warn) 40%, transparent); }
  .toast.is-warning .ic { color: var(--warn); }
  .toast.is-danger { border-color: color-mix(in srgb, var(--gn-danger) 40%, transparent); }
  .toast.is-danger .ic { color: var(--danger); }
  .toast b { display: block; font-size: 14px; color: var(--text); }
  .toast p { display: block; color: var(--text-2); font-size: 13px; margin: 2px 0 0; word-break: break-word; }
  .toast .x { all: unset; cursor: pointer; color: var(--muted-c); width: 24px; height: 24px; display: grid; place-items: center; border-radius: 6px; }
  .toast .x :global(svg) { width: 16px; height: 16px; fill: none; stroke: currentColor; stroke-width: 1.75; stroke-linecap: round; }
  .toast .x:hover { color: var(--text); background: var(--gn-surface-3); }
  .toast .x:focus-visible { box-shadow: var(--gn-focus); }
  .toast .pb { position: absolute; left: 14px; right: 14px; bottom: 0; height: 2px; border-radius: 2px; background: var(--gn-gradient); transform-origin: left; animation: pb linear forwards; }
  @keyframes rise { from { opacity: 0; transform: translateY(10px); } }
  @keyframes pb { from { transform: scaleX(1); } to { transform: scaleX(0); } }

  @media (min-width: 700px) { .more { grid-template-columns: minmax(0, 2fr) minmax(0, 3fr); } }
  @media (min-width: 900px) { .hero { padding: 48px 0 40px; } .hero .wrap { grid-template-columns: minmax(0, 2fr) minmax(0, 3fr); gap: 32px; align-items: stretch; } .stats { grid-template-columns: repeat(5, minmax(0, 1fr)); } }
  @media (max-width: 899px) { .band { grid-template-columns: minmax(0, 1fr) 180px; } .stats > .tile:last-child:nth-child(odd) { grid-column: span 2; } }
  @media (max-width: 640px) { footer ul a { display: inline-flex; align-items: center; min-height: 40px; } .btn.xs { min-height: 40px; } .addr { padding: 8px 0; } .hide-s { display: none; } .field { flex-direction: column; } .btn.pri { width: 100%; } .ctas .btn { flex: 1 1 auto; } .tape .trk { animation: none; padding-left: 90px; overflow-x: auto; } dl div { grid-template-columns: 1fr; gap: 2px; } .stepper li + li::before { width: 12px; margin: 0 6px; } .band { grid-template-columns: minmax(0, 1fr); } .mascot { min-height: 170px; } }
  /* below 400 px the header holds the kit logo and the switcher only; the Latest block tile carries the live number */
  @media (max-width: 399px) { .navr .pill { display: none; } }
  @media (prefers-reduced-motion: reduce) { .pill .dot, .toast, .btn::after, .spinner, .bk, .spark rect { animation: none !important; } .toast .pb { display: none; } .tape .trk { animation: none; padding-left: 90px; overflow-x: auto; } .btn, .btn .ar, .mascot img { transition: none; } .btn:hover:not(:disabled) .ar, .group.band:hover .mascot img { transform: none; } }
</style>
