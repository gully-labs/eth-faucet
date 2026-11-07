<script>
  import { onMount, onDestroy } from 'svelte';
  import { getAddress } from '@ethersproject/address';
  import GullySignIn from './GullySignIn.svelte';
  import { parseTarget, targetCheck, appName, sameAddress } from './target.js';
  import { fetchPayouts, bucketPayouts, latestPayouts, niceMax } from './drips.js';

  // GullyChain testnet faucet (gully-labs fork): brand kit v4 (Gully Chain Night 2.0). The faucet logic,
  // flags and API below are upstream's; only the page changed.
  const CHAIN = {
    id: 485590,
    hex: '0x768d6',
    name: 'GullyChain Testnet',
    rpc: 'https://testnet-rpc.gullychain.xyz',
    explorer: 'https://testnet-explorer.gullychain.xyz',
  };
  const LINKS = [
    // owner 2026-10-31: the one Gully Chain header action (gully-ops sites/_shared/gc-chrome.json), never the consumer app
    { name: 'Start building', href: 'https://dev.gullychain.xyz/quickstart/', icon: 'rocket' },
    { name: 'Bridge', href: 'https://testnet-bridge.gullychain.xyz/', icon: 'arrows' },
    { name: 'Docs', href: 'https://docs.gullychain.xyz/', icon: 'book' },
    { name: 'Explorer', href: CHAIN.explorer + '/', icon: 'search' },
  ];
  // Header nav (design-gap audit 2026-10-29): every key Gully Chain surface one click from the faucet, icon + label, same
  // tab. The <gc-switcher> (1.7.0, opens on hover) carries the full product grid; below 900 px the same items sit in the menu panel.
  const NAV = [
    { name: 'Developers', href: 'https://dev.gullychain.xyz/', icon: 'code' },
    { name: 'Docs', href: 'https://docs.gullychain.xyz/', icon: 'book' },
    { name: 'Bridge', href: 'https://testnet-bridge.gullychain.xyz/', icon: 'arrows' },
    { name: 'Explorer', href: CHAIN.explorer + '/', icon: 'search' },
    { name: 'Ecosystem', href: 'https://gullychain.xyz/ecosystem/', icon: 'grid' },
    { name: 'Status', href: 'https://status.gullylabs.xyz/', icon: 'activity' },
  ];
  // Footer: every GullyChain product, same tab (one product family). the consumer app is listed only on the
  // Ecosystem pages (owner 2026-10-30), so it is not here.
  const PRODUCTS = [
    { name: 'GullyChain', href: 'https://gullychain.xyz/', icon: 'globe' },
    { name: 'Bridge', href: 'https://testnet-bridge.gullychain.xyz/', icon: 'arrows' },
    { name: 'Explorer', href: CHAIN.explorer + '/', icon: 'search' },
    { name: 'Sentry', href: 'https://sentry.gullychain.xyz/', icon: 'shield' },
    { name: 'Docs', href: 'https://docs.gullychain.xyz/', icon: 'book' },
    { name: 'Developers', href: 'https://dev.gullychain.xyz/', icon: 'code' },
    { name: 'Ecosystem', href: 'https://gullychain.xyz/ecosystem/', icon: 'grid' },
    { name: 'Status', href: 'https://status.gullylabs.xyz/', icon: 'activity' },
  ];
  const SOURCE = 'https://github.com/gully-labs/eth-faucet';
  // Every other Gully product is listed once, on the docs ecosystem page (owner rule 2026-10-29: Gully Chain nav only).
  const ECOSYSTEM = 'https://docs.gullychain.xyz/ecosystem/';
  // Gully Chain's one Terms of Service and Privacy Policy (gullychain.xyz) cover the faucet; the Go server 301s
  // /terms and /privacy here to them (internal/server/legal.go). Another origin: new tab.
  const TERMS = 'https://gullychain.xyz/terms';
  const PRIVACY = 'https://gullychain.xyz/privacy';
  const LEGAL = [
    { name: 'Terms', href: TERMS, icon: 'file' },
    { name: 'Privacy', href: PRIVACY, icon: 'shield' },
  ];
  // Official GullyChain channels (every GullyChain footer carries them).
  const SOCIALS = [
    { name: 'Telegram', href: 'https://t.me/GullyChain', icon: 'send' },
    { name: 'Deploys', href: 'https://t.me/GullyChainDeploys', icon: 'rocket' },
    { name: 'Discord', href: 'https://discord.com/invite/tY2ErsNNBp', icon: 'chat' },
    { name: 'X', href: 'https://x.com/Gully_Labs', icon: 'at' },
  ];

  // Lucide icons (ISC, lucide.dev): 24-unit paths, drawn at 20px, stroke 1.75
  const ICONS = {
    file: '<path d="M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z"/><path d="M14 2v4a2 2 0 0 0 2 2h4"/><path d="M10 9H8"/><path d="M16 13H8"/><path d="M16 17H8"/>',
    grid: '<rect width="7" height="7" x="3" y="3" rx="1"/><rect width="7" height="7" x="14" y="3" rx="1"/><rect width="7" height="7" x="14" y="14" rx="1"/><rect width="7" height="7" x="3" y="14" rx="1"/>',
    arrows: '<path d="M8 3 4 7l4 4"/><path d="M4 7h16"/><path d="m16 21 4-4-4-4"/><path d="M20 17H4"/>',
    book: '<path d="M12 7v14"/><path d="M3 18a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1h5a4 4 0 0 1 4 4 4 4 0 0 1 4-4h5a1 1 0 0 1 1 1v13a1 1 0 0 1-1 1h-6a3 3 0 0 0-3 3 3 3 0 0 0-3-3z"/>',
    search: '<circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/>',
    code: '<path d="m16 18 6-6-6-6"/><path d="m8 6-6 6 6 6"/>',
    activity: '<path d="M22 12h-2.48a2 2 0 0 0-1.93 1.46l-2.35 8.36a.25.25 0 0 1-.48 0L9.24 2.18a.25.25 0 0 0-.48 0l-2.35 8.36A2 2 0 0 1 4.49 12H2"/>',
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
    // nav, mobile menu and footer (owner rule 2026-10-30: every nav item on every Gully Chain surface has an icon)
    menu: '<path d="M4 12h16"/><path d="M4 18h16"/><path d="M4 6h16"/>',
    globe: '<circle cx="12" cy="12" r="10"/><path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20"/><path d="M2 12h20"/>',
    appwin: '<rect x="2" y="4" width="20" height="16" rx="2"/><path d="M10 4v4"/><path d="M2 8h20"/><path d="M6 4v4"/>',
    shield: '<path d="M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z"/><path d="m9 12 2 2 4-4"/>',
    branch: '<line x1="6" x2="6" y1="3" y2="15"/><circle cx="18" cy="6" r="3"/><circle cx="6" cy="18" r="3"/><path d="M18 9a9 9 0 0 1-9 9"/>',
    layers: '<path d="M12.83 2.18a2 2 0 0 0-1.66 0L2.6 6.08a1 1 0 0 0 0 1.83l8.58 3.91a2 2 0 0 0 1.66 0l8.58-3.9a1 1 0 0 0 0-1.83z"/><path d="M2 12a1 1 0 0 0 .58.91l8.6 3.91a2 2 0 0 0 1.65 0l8.58-3.9A1 1 0 0 0 22 12"/><path d="M2 17a1 1 0 0 0 .58.91l8.6 3.91a2 2 0 0 0 1.65 0l8.58-3.9A1 1 0 0 0 22 17"/>',
    send: '<path d="M14.536 21.686a.5.5 0 0 0 .937-.024l6.5-19a.496.496 0 0 0-.635-.635l-19 6.5a.5.5 0 0 0-.024.937l7.93 3.18a2 2 0 0 1 1.112 1.11z"/><path d="m21.854 2.147-10.94 10.939"/>',
    rocket: '<path d="M4.5 16.5c-1.5 1.26-2 5-2 5s3.74-.5 5-2c.71-.84.7-2.13-.09-2.91a2.18 2.18 0 0 0-2.91-.09z"/><path d="m12 15-3-3a22 22 0 0 1 2-3.95A12.88 12.88 0 0 1 22 2c0 2.72-.78 7.5-6 11a22.35 22.35 0 0 1-4 2z"/><path d="M9 12H4s.55-3.03 2-4c1.62-1.08 5 0 5 0"/><path d="M12 15v5s3.03-.55 4-2c1.08-1.62 0-5 0-5"/>',
    chat: '<path d="M7.9 20A9 9 0 1 0 4 16.1L2 22Z"/>',
    at: '<circle cx="12" cy="12" r="4"/><path d="M16 8v5a3 3 0 0 0 6 0v-1a10 10 0 1 0-4 8"/>',
  };
  // Mobile menu: below 900 px the header nav folds into this panel (the gc-switcher still carries the product grid)
  let menuOpen = false;
  const icon = (n) => `<svg viewBox="0 0 24 24" aria-hidden="true">${ICONS[n] || ICONS.info}</svg>`;

  let input = null;
  let fieldError = ''; // inline validation message under the address field
  // Sign in with Gully (only when /api/info has gullyid_client): fill the field with the main wallet, but never
  // overwrite an address the visitor typed; clear it again on sign-out only if it is still the filled-in one.
  let gullyFilled = null;
  let gullyToken = null; // session access token, sent as Authorization: Bearer (required mode: the gate; optional mode: unlocks brand-new wallets)
  let freshRefused = false; // the last claim was refused because the address has no history
  let gullyMain = null; // the Gully ID main wallet (profile `wallet`)
  let gullyWallets = []; // every wallet of the Gully ID (profile `wallets`), to tell "yours, but not kre8's" apart
  // An app link (kre8 "Get test ETH": ?to=<account>&from=kre8) locks the field to the account the app uses, so the
  // Gully ID sign-in below can not swap it for the main wallet (owner report 2026-10-31).
  const target = typeof window !== 'undefined' ? parseTarget(window.location.search) : { to: null, from: null };
  let lockedTo = target.to;
  if (lockedTo) input = lockedTo;
  let confirmOther = false; // the person saw the "not your account" warning and pressed Send again
  function onGullyWallet(address, user, session) {
    gullyToken = session?.accessToken || null;
    gullyMain = address || null;
    // signed in after a "no history" refusal: the brand-new wallet can be funded now, so the refusal goes
    if (gullyToken && freshRefused) { freshRefused = false; fieldError = ''; }
    gullyWallets = Array.isArray(user?.wallets) ? user.wallets : [];
    if (lockedTo) {
      if (step < 2 && address) step = 2;
      return;
    }
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
  function unlockTarget() {
    lockedTo = null;
    setTimeout(() => document.getElementById('address')?.focus(), 0);
  }
  function useMyGullyAccount() {
    if (!check.expected) return;
    input = check.expected;
    fieldError = '';
  }
  $: check = targetCheck({ field: input, signedIn: !!gullyMain || !!gullyToken, gullyMain, gullyWallets, lockedTo: target.to, from: target.from });
  // a new address in the field needs a new confirmation
  $: input, (confirmOther = false);

  let faucetInfo = {
    account: '0x0000000000000000000000000000000000000000',
    network: 'testnet',
    payout: '',
    symbol: 'ETH',
    hcaptcha_sitekey: '',
  };

  let mounted = false;
  let infoError = false;
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

  // ---- live numbers. RPC (read-only): head block, faucet balance, faucet nonce = payouts ever sent. Explorer index: the
  //      payouts per day and the latest payouts. A source that fails shows "–" with the reason, never a made-up number.
  const DRIP_DAYS = 14;
  let blockNumber = null;
  let blockBumped = false;
  let blockTimer = null;
  let balance = null; // faucet account balance, ETH
  let nonce = null; // payouts ever sent by the faucet account
  let shownNonce = null; // the same number, counting up when it changes
  let rpcDown = false;
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
  async function pollBlock() {
    if (typeof document !== 'undefined' && document.visibilityState === 'hidden') return;
    try {
      const n = parseInt(await rpc('eth_blockNumber', []), 16);
      if (!Number.isFinite(n)) return;
      rpcDown = false;
      if (n !== blockNumber) {
        blockNumber = n;
        blockBumped = true;
        setTimeout(() => (blockBumped = false), 900);
      }
      if (mounted) {
        rpc('eth_getBalance', [faucetInfo.account, 'latest']).then((h) => (balance = parseInt(h, 16) / 1e18)).catch(() => {});
        rpc('eth_getTransactionCount', [faucetInfo.account, 'latest']).then((h) => {
          const v = parseInt(h, 16);
          if (!Number.isFinite(v) || v === nonce) return;
          const first = nonce == null;
          nonce = v;
          countTo(v);
          if (!first) loadDrips(); // a new payout: pull it into the chart
        }).catch(() => {});
      }
      nowTs = Date.now();
    } catch (error) {
      rpcDown = true;
      console.error('Block poll failed:', error);
    }
  }
  const reduced = typeof window !== 'undefined' && window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  let countRaf = null;
  function countTo(v) {
    const from = shownNonce == null ? Math.max(0, v - Math.min(v, 40)) : shownNonce;
    if (countRaf) cancelAnimationFrame(countRaf);
    if (reduced || from === v || typeof requestAnimationFrame === 'undefined') { shownNonce = v; return; }
    const t0 = performance.now();
    const f = (now) => {
      const k = Math.min(1, (now - t0) / 700), e = 1 - Math.pow(1 - k, 3);
      shownNonce = Math.round(from + (v - from) * e);
      countRaf = k < 1 ? requestAnimationFrame(f) : null;
    };
    countRaf = requestAnimationFrame(f);
  }

  // ---- drips per day: hand SVG, bars tween to new heights (rAF, 600 ms ease-out); reduced motion applies at once
  let dripDays = []; // { day, n (null = not covered), eth }
  let recent = []; // the latest payouts
  let dripsError = '';
  let dripsLoading = true;
  let dripsTimer = null;
  let nowTs = Date.now();
  let dripsBusy = false;
  async function loadDrips() {
    if (dripsBusy || !mounted) return;
    dripsBusy = true;
    try {
      const { items, covered } = await fetchPayouts(fetch, CHAIN.explorer, faucetInfo.account, DRIP_DAYS);
      dripDays = bucketPayouts(items, faucetInfo.account, DRIP_DAYS, covered);
      recent = latestPayouts(items, faucetInfo.account);
      dripsError = '';
    } catch (error) {
      console.error('Payout history failed:', error);
      dripsError = 'The explorer index is not answering. Retrying every minute.';
    } finally {
      dripsBusy = false;
      dripsLoading = false;
    }
  }
  $: dripsToday = dripDays.length ? dripDays[dripDays.length - 1].n : null;
  $: dripsWin = dripDays.reduce((s, d) => s + (d.n || 0), 0);
  $: ethWin = dripDays.reduce((s, d) => s + (d.eth || 0), 0);
  $: claimsLeft = balance != null && +faucetInfo.payout > 0 ? Math.floor(balance / +faucetInfo.payout + 1e-9) : null;

  const CH = 200, ML = 34, MR = 10, MT = 12, MB = 26;
  let plotW = 0;
  $: CW = plotW || 640;
  $: gridMax = niceMax(Math.max(0, ...dripDays.map((b) => b.n || 0)));
  let bars = []; // { day, n, eth, x, w, y, h }
  let shown = {}; // day -> displayed { y, h }
  let raf = null;
  let tipHtml = '';
  let tipX = 0, tipY = 0, tipOn = false;
  $: layoutBars(dripDays, CW, gridMax);
  function layoutBars(list, CW, max) {
    const n = list.length;
    const span = CW - ML - MR, slot = span / Math.max(1, n), bw = Math.max(3, Math.min(34, slot - 6));
    const base = CH - MB;
    const target = list.map((b, i) => {
      const h = b.n ? Math.max(3, (CH - MT - MB) * b.n / max) : 0;
      return { ...b, today: i === n - 1, x: ML + slot * i + (slot - bw) / 2, w: bw, y: base - h, h };
    });
    const from = target.map((t) => shown[t.day] || { y: base, h: 0 });
    if (raf) cancelAnimationFrame(raf);
    if (reduced || !n || typeof requestAnimationFrame === 'undefined') { bars = target; shown = Object.fromEntries(target.map((t) => [t.day, { y: t.y, h: t.h }])); return; }
    const t0 = performance.now();
    const stepFn = (now) => {
      const k = Math.min(1, (now - t0) / 600), e = 1 - Math.pow(1 - k, 3);
      bars = target.map((t, i) => ({ ...t, y: from[i].y + (t.y - from[i].y) * e, h: from[i].h + (t.h - from[i].h) * e }));
      if (k < 1) raf = requestAnimationFrame(stepFn);
      else { raf = null; shown = Object.fromEntries(target.map((t) => [t.day, { y: t.y, h: t.h }])); }
    };
    raf = requestAnimationFrame(stepFn);
  }
  const dayLabel = (d) => new Date(d + 'T00:00:00Z').toLocaleDateString('en-US', { month: 'short', day: 'numeric', timeZone: 'UTC' });
  function showTip(e, b) {
    const what = b.n === null ? 'not in the pages read' : `${b.n} payout${b.n === 1 ? '' : 's'}${b.n ? ` · ${fmtEth(b.eth)} ETH` : ''}`;
    tipHtml = `<b>${what}</b><br><span>${dayLabel(b.day)} UTC${b.today ? ', so far today' : ''}</span>`;
    tipX = Math.min(e.clientX + 14, window.innerWidth - 240);
    tipY = e.clientY - 10;
    tipOn = true;
  }
  const fmtEth = (v) => (v >= 1 ? v.toFixed(2) : v >= 0.01 ? v.toFixed(3) : v.toFixed(4)).replace(/\.?0+$/, '');

  // the mascot breathes only while it is on screen (transform only, off under reduced motion)
  let mascotEl = null;
  let mascotOn = false;
  let mascotIO = null;
  $: if (mascotEl && !mascotIO && typeof IntersectionObserver !== 'undefined') {
    mascotIO = new IntersectionObserver((e) => (mascotOn = e[0].isIntersecting));
    mascotIO.observe(mascotEl);
  }

  // ---- sign-in mode, from /api/info signin_mode (owner 2026-11-05: any wallet, Gully ID optional):
  //      optional (the live config): the address field is first and usable at once: 1 enter address, 2 request ETH,
  //        3 done. "Sign in with Gully ID" is a secondary action under the field: it fills the address and unlocks
  //        brand-new wallets (the server skips its no-history check for a verified account).
  //      required: 1 sign in with Gully ID, 2 claim, 3 done; the claim stays locked until the session is there.
  //      off: no Gully ID at all, same steps as optional.
  //      Nothing here opens Gully ID by itself: only a press on the sign-in button does.
  let step = 1;
  let networkAdded = false;
  let paid = false;
  // the last successful claim: ETH tx + the token drip (USDC.e) result, shown under the form
  let claimResult = null; // { ethTx, token: { symbol, amount, sent, tx, msg } | null }
  $: mode = faucetInfo.signin_mode || (faucetInfo.gullyid_client ? (faucetInfo.signin_required ? 'required' : 'optional') : 'off');
  $: needSignIn = mounted && mode === 'required';
  $: signInOptional = mounted && mode === 'optional';
  $: hasAddress = !!(input || '').trim();
  $: payoutLabel = faucetInfo.payout ? `${faucetInfo.payout} ${faucetInfo.symbol}` : '–';
  $: heroAmount = faucetInfo.payout || infoError ? payoutLabel : '0.000 ETH'; // invisible until /api/info answers: it only holds the line
  $: steps = needSignIn ? ['Sign in', 'Claim', 'Done'] : ['Enter address', 'Request ETH', 'Done'];
  $: cur = needSignIn ? (paid ? 4 : !gullyToken ? 1 : 2) : paid ? 4 : hasAddress ? 2 : 1;
  $: claimBlocked = needSignIn && !gullyToken;

  // ---- Connect wallet: the shared Gully Chain dialog (public/gc-connect.js, canonical gully-ops sites/_shared/gc-connect,
  //      byte for byte): Gully Wallet (passkey) first, WalletConnect, detected EIP-6963 wallets, then install links.
  //      It only fills the address field; in required mode the claim still needs the Gully ID sign-in above.
  let connecting = false;
  let walletName = '';
  // Wallet code is not part of the first paint: the dialog loads when someone asks to connect.
  function loadConnect() {
    if (window.GullyConnect) return Promise.resolve(window.GullyConnect);
    return new Promise((done) => {
      const s = document.createElement('script');
      s.src = '/gc-connect.js?v=1.1.0';
      s.onload = () => done(window.GullyConnect || null);
      s.onerror = () => done(null);
      document.head.append(s);
    });
  }
  async function connectWallet() {
    const gc = typeof window !== 'undefined' ? await loadConnect() : null;
    if (!gc || typeof gc.open !== 'function') {
      notify({ title: 'Wallet picker unavailable', message: 'It did not load. Paste your address instead, or reload the page.', type: 'is-warning', icon: 'wallet' });
      return;
    }
    connecting = true;
    try {
      const r = await gc.open({ appName: 'Gully Faucet' });
      const a = r?.accounts?.[0];
      if (a) {
        input = a;
        walletName = r.name || 'wallet';
        fieldError = '';
        if (step < 2) step = 2;
        notify({ title: walletName + ' connected', message: 'Address ' + short(a) + ' filled in.', type: 'is-success', icon: 'wallet', ms: 4000 });
      }
    } catch (e) {
      if (e?.code !== 4001) notify({ title: 'Wallet not connected', message: e?.message || 'The wallet did not connect.', type: 'is-warning', icon: 'wallet' });
    } finally {
      connecting = false;
    }
  }

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

  // Last-known payout from the last visit, shown at once. Display only: the sign-in mode, the captcha and every claim
  // decision wait for the live /api/info.
  const LAST_KNOWN = 'gully-faucet-last-known-v1';
  function saveLastKnown(i) {
    try {
      localStorage.setItem(LAST_KNOWN, JSON.stringify({ payout: i.payout, symbol: i.symbol }));
    } catch {
      /* storage blocked: no last-known value next time */
    }
  }
  function restoreLastKnown() {
    try {
      const c = JSON.parse(localStorage.getItem(LAST_KNOWN) || 'null');
      if (c?.payout && !faucetInfo.payout) faucetInfo = { ...faucetInfo, payout: c.payout, symbol: c.symbol };
    } catch {
      /* no cache */
    }
  }
  async function loadInfo() {
    infoError = false;
    try {
      const res = await fetch('/api/info');
      if (!res.ok) {
        throw new Error('Failed to fetch faucet info');
      }
      faucetInfo = await res.json();
      saveLastKnown(faucetInfo);
      mounted = true;
      loadDrips();
      pollBlock();
    } catch (error) {
      console.error('Failed to load faucet info:', error);
      infoError = true;
    }
  }
  onMount(async () => {
    restoreLastKnown();
    await loadInfo();
    pollBlock();
    blockTimer = setInterval(pollBlock, 6000);
    dripsTimer = setInterval(() => { if (document.visibilityState !== 'hidden') loadDrips(); }, 60000);
  });
  onDestroy(() => {
    if (blockTimer) clearInterval(blockTimer);
    if (dripsTimer) clearInterval(dripsTimer);
    if (raf) cancelAnimationFrame(raf);
    if (countRaf) cancelAnimationFrame(countRaf);
    mascotIO?.disconnect();
  });

  // Register hCaptcha callback function
  if (typeof window !== 'undefined') {
    window[HCAPTCHA_CALLBACK_NAME] = () => {
      hcaptchaLoaded = true;
    };
  }

  $: if (faucetInfo.payout) document.title = `GullyChain Testnet Faucet · ${faucetInfo.payout} ${faucetInfo.symbol} a day`;

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

      // Signed in with Gully ID and paying another address than the Gully account kre8 uses: warn once, pay on the
      // second press (the warning above the field says why and offers "Use my Gully account").
      const c = targetCheck({ field: address, signedIn: !!gullyMain || !!gullyToken, gullyMain, gullyWallets, lockedTo: target.to, from: target.from });
      // Only in required mode: with sign-in optional any wallet is a normal target, so the note stays but one press pays.
      if (needSignIn && c.state === 'mismatch' && !confirmOther) {
        confirmOther = true;
        notify({ title: c.title, message: 'Tap "Use my Gully account", or press Send again to pay this address anyway.', type: 'is-warning', icon: 'wallet', ms: 9000 });
        return;
      }

      // Prepare request headers
      const headers = {
        'Content-Type': 'application/json',
      };
      if (needSignIn && !gullyToken) {
        fieldError = 'Sign in with Gully ID first (step 1): one claim per Gully ID a day.';
        toast({ message: 'Sign in with Gully ID to claim testnet ETH.', type: 'is-warning' });
        return;
      }
      // required: the session is the gate. optional: it is sent when there is one (any address, one claim per account);
      // without it the claim goes out as it is.
      if (gullyToken && mode !== 'off') headers['Authorization'] = 'Bearer ' + gullyToken;

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
        freshRefused = false;
      } else {
        const errorMessage = data?.msg || 'Request failed';
        fieldError = errorMessage;
        freshRefused = /no history/i.test(errorMessage);
        if (freshRefused && signInOptional) {
          // a brand-new wallet: the sign-in under the field is what funds it
          if (gullyToken) fieldError = 'This address has no history yet and your Gully ID session was not accepted (it may have expired). Sign out and sign in again to fund a brand-new wallet, or use a wallet that has made at least one transaction on Gully Chain or Arbitrum Sepolia.';
          notify({ title: 'Brand-new wallet', message: gullyToken ? 'Sign in with Gully ID again to fund it.' : 'Sign in with Gully ID (below the address) to fund it, or use a wallet that has made a transaction.', type: 'is-warning', icon: 'wallet', ms: 10000 });
        } else {
          toast({ message: errorMessage, type: 'is-danger' });
        }
        // an app account that is not on chain yet (kre8 passkey accounts go on chain with their first action)
        if (target.from && sameAddress(address, target.to) && freshRefused) fieldError = signInOptional && !gullyToken ? `Your ${appName(target.from)} account is not on chain yet. Sign in with Gully ID below to fund it now, or in ${appName(target.from)} tap "Get test ETH" again: it puts the account on chain first (free), then sends you here.` : `Your ${appName(target.from)} account is not on chain yet. In ${appName(target.from)}, tap "Get test ETH" again: it puts the account on chain first (free), then sends you here.`;
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

  const fmtBlock = (n) => (n == null ? '–' : n.toLocaleString('en-US'));
  const short = (a) => (a ? a.slice(0, 6) + '…' + a.slice(-4) : '');
  const ago = (ts, now = Date.now()) => { const s = Math.max(0, now / 1000 - ts); return s < 60 ? Math.round(s) + 's' : s < 3600 ? Math.round(s / 60) + 'm' : s < 86400 ? (s / 3600).toFixed(1) + 'h' : Math.round(s / 86400) + 'd'; };
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
      <!-- brand kit v4 logo/gc-icon.svg (byte for byte) on night; kit v4 has no wordmark file, so the icon stands alone.
           The switcher hides its own logo (logo="off") so the mark shows once. gc-switcher: public/gc-switcher.js, canonical gully-ops sites/_shared/gc-switcher -->
      <a class="logo" href="https://gullychain.xyz/" aria-label="Gully Chain home"><img src="/brand/logo/gc-icon.svg" width="32" height="32" alt="Gully Chain" /></a>
      <gc-switcher current="faucet" theme="dark" logo="off"></gc-switcher>
    </div>
    <nav class="topnav" aria-label="Gully Chain"><ul>{#each NAV as l}<li><a href={l.href} title={l.name}><span class="ic">{@html icon(l.icon)}</span><span class="tl">{l.name}</span></a></li>{/each}</ul></nav>
    <div class="navr">
      <span class="pill" class:live={blockNumber != null} class:flash={blockBumped} role="status" aria-live="polite" title="Latest block from {CHAIN.rpc}"><span class="dot" aria-hidden="true"></span><span>{blockNumber == null ? 'Connecting' : 'Live · #' + fmtBlock(blockNumber)}</span></span>
      <button type="button" class="btn ter sm hide-s" onclick={addToWallet} disabled={addingChain}><span class="ic">{@html icon(networkAdded ? 'check' : 'wallet')}</span>Add to wallet</button>
      <button type="button" class="menubtn" aria-label="Menu" aria-expanded={menuOpen} aria-controls="mnav" onclick={() => (menuOpen = !menuOpen)}><span class="ic">{@html icon(menuOpen ? 'x' : 'menu')}</span></button>
    </div>
  </div>
  {#if menuOpen}<nav class="mnav" id="mnav" aria-label="Gully Chain menu"><ul>{#each NAV as l}<li><a href={l.href} onclick={() => (menuOpen = false)}><span class="ic">{@html icon(l.icon)}</span><span>{l.name}</span></a></li>{/each}</ul></nav>{/if}
</header>
<svelte:window onkeydown={(e) => { if (e.key === 'Escape') menuOpen = false; }} />

<main>
  <section class="hero" aria-labelledby="h1">
    <div class="wrap">
      <div class="copy">
        <span class="badge">Testnet faucet · chain {CHAIN.id}</span>
        <h1 id="h1">Testnet ETH, <em class:hold={!faucetInfo.payout && !infoError}>{heroAmount.replace(' ', '\u00a0')}</em> a day</h1>
        <p class="promise" data-testid="promise">{#if needSignIn}Sign in with Gully ID, then one claim per account, address and IP every 24 hours.{:else if signInOptional}Paste any wallet address: one claim per address and per IP every 24 hours. Signing in with Gully ID is optional and also funds brand-new wallets.{:else}Paste any wallet address: one claim per address and per IP every 24 hours.{/if} {#if faucetInfo.token}Each claim also sends <b>{faucetInfo.token.amount} {faucetInfo.token.symbol}</b> (test stablecoin) while the daily {faucetInfo.token.symbol} budget lasts, for Gully Pay checkout and swaps. {/if}<b>Testnet ETH has no value</b>; use it to try the app, deploy contracts and test the bridge.</p>

        <ol class="stepper" aria-label="Steps" data-testid="stepper">
          {#each steps as label, i}
            <li class:done={cur > i + 1} class:active={cur === i + 1}><span class="n">{#if cur > i + 1}{@html icon('check')}{:else}{i + 1}{/if}</span><span>{label}</span></li>
          {/each}
        </ol>

        <div id="hcaptcha" data-size="invisible"></div>
        <form id="claim" class="claim" onsubmit={(e) => { e.preventDefault(); if (!isLoading) handleRequest(); }}>
          {#if infoError}
            <div class="inforr" role="alert" data-testid="info-error"><p>The faucet did not answer, so its settings are not loaded.</p><button type="button" class="btn ter sm" onclick={loadInfo}>Try again</button></div>
          {:else if !mounted}
            <div class="sk" aria-hidden="true"></div>
          {/if}
          {#if needSignIn}
            <GullySignIn clientId={faucetInfo.gullyid_client} issuer={faucetInfo.gullyid_issuer || undefined} required={true} onwallet={onGullyWallet} lockedNote={lockedTo ? `the claim pays your ${appName(target.from)} account ${short(lockedTo)}` : ''} />
          {/if}
          {#if claimBlocked}<p id="claim-why" class="hint" data-testid="claim-why">The claim unlocks after you sign in (step 1): one claim per account a day.{#if lockedTo} It pays your {appName(target.from)} account <span class="mono">{short(lockedTo)}</span>.{/if}</p>{/if}
          {#if mounted && !claimBlocked && !paid}
          <div class="addr-h">
            <label class="lab" for="address">Your address or ENS name</label>
            <button type="button" class="btn ter sm" data-testid="connect-wallet" onclick={connectWallet} disabled={connecting} aria-busy={connecting}><span class="ic">{@html icon('wallet')}</span>{walletName ? walletName + ' connected' : 'Connect wallet'}</button>
          </div>
          {#if !(input || '').trim()}<p class="hint addr-hint">Connect a wallet, or paste any address.</p>{/if}
          <div class="field">
            <input
              id="address"
              bind:value={input}
              class="input mono"
              type="text"
              autocomplete="off"
              spellcheck="false"
              placeholder="0x… or name.eth"
              readonly={!!lockedTo}
              data-testid="address"
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
            <button type="submit" class="btn pri" disabled={isLoading || claimBlocked} aria-busy={isLoading} aria-describedby={claimBlocked ? 'claim-why' : undefined} data-testid="claim">
              {#if isLoading}
                <span class="spinner" aria-hidden="true"></span>Sending
              {:else if confirmOther && check.state === 'mismatch'}
                <span class="ic">{@html icon('droplet')}</span>Send to this address anyway<span class="ic ar">{@html icon('arrow')}</span>
              {:else}
                <span class="ic">{@html icon('droplet')}</span>Request {faucetInfo.payout} {faucetInfo.symbol}<span class="ic ar">{@html icon('arrow')}</span>
              {/if}
            </button>
          </div>
          <p class="hint consent" data-testid="consent">By requesting testnet ETH you agree to the Gully Chain <a href={TERMS} target="_blank" rel="noopener noreferrer">Terms<span class="gn-sr"> (opens in a new tab)</span></a> and <a href={PRIVACY} target="_blank" rel="noopener noreferrer">Privacy Policy<span class="gn-sr"> (opens in a new tab)</span></a>. Testnet ETH has no value.</p>
          {#if lockedTo}
            <p class="lockline" data-testid="locked-to"><span class="ic">{@html icon('check')}</span><span>Pays your {appName(target.from)} account <span class="mono">{short(lockedTo)}</span>, the one {appName(target.from)} shows your balance for.</span> <button type="button" class="linkbtn" onclick={unlockTarget}>Use another address</button></p>
          {/if}
          {#if check.state === 'mismatch'}
            <div class="twarn" class:soft={!needSignIn} role={needSignIn ? 'alert' : 'note'} data-testid="not-your-account">
              <p class="twarn-h"><span class="ic">{@html icon('info')}</span>{needSignIn ? check.title : 'Paying another wallet than your Gully account'}</p>
              <p class="twarn-s">{needSignIn ? check.detail : 'That is fine: any address works. kre8 shows only the balance of your Gully account, so this ETH will not appear there.'}</p>
              <button type="button" class="btn sec sm" onclick={useMyGullyAccount} data-testid="use-my-gully-account"><span class="ic">{@html icon('wallet')}</span>Use my Gully account <span class="mono">{short(check.expected)}</span></button>
            </div>
          {/if}
          {#if fieldError}<p id="address-err" class="ferr">{fieldError}</p>{/if}
{#if faucetInfo.eligibility_check}<p class="hint" data-testid="who-is-paid">{#if needSignIn}Pays your Gully wallet, or any address with a transaction on Gully Chain or Arbitrum Sepolia. Empty addresses are refused.{:else if signInOptional && gullyToken}Signed in with Gully ID: any address works, brand-new wallets included.{:else if signInOptional}Works for any wallet that has made a transaction on Gully Chain or Arbitrum Sepolia. Brand-new wallet? Sign in with Gully ID below.{:else}Works for any wallet that has made a transaction on Gully Chain or Arbitrum Sepolia.{/if}</p>{/if}
          {/if}
          {#if signInOptional}
            <!-- optional, secondary, under the address and the claim button: it stays mounted after a claim (hidden) so the session is kept -->
            <div class="gidopt" data-testid="gid-optional" hidden={paid}>
              <GullySignIn clientId={faucetInfo.gullyid_client} issuer={faucetInfo.gullyid_issuer || undefined} required={false} highlight={freshRefused && !gullyToken} onwallet={onGullyWallet} lockedNote={lockedTo ? `the claim pays your ${appName(target.from)} account ${short(lockedTo)}` : 'any address works, brand-new wallets included'} />
            </div>
          {/if}
          {#if paid}
            <div class="done" data-testid="done">
          {#if claimResult}
            <div class="claim-result" role="status" data-testid="claim-result">
              <p><span class="ic ok">{@html icon('check')}</span>{faucetInfo.payout} {faucetInfo.symbol} sent{#if claimResult.ethTx}{' · '}<a class="mono" href={`${CHAIN.explorer}/tx/${claimResult.ethTx}`} target="_blank" rel="noopener noreferrer">{short(claimResult.ethTx)}</a>{/if}</p>
              {#if claimResult.token}
                {#if claimResult.token.sent}
                  <p data-testid="token-sent"><span class="ic ok">{@html icon('check')}</span>{claimResult.token.amount} {claimResult.token.symbol} sent · <a class="mono" href={`${CHAIN.explorer}/tx/${claimResult.token.tx}`} target="_blank" rel="noopener noreferrer">{short(claimResult.token.tx)}</a></p>
                {:else}
                  <p data-testid="token-skipped" class="muted"><span class="ic">{@html icon('info')}</span>{claimResult.token.msg}</p>
                {/if}
              {/if}
            </div>
          {/if}
                        <p class="hint">Next claim in {faucetInfo.window_hours || 24} hours. What to do with it:</p>
              <div class="ctas">
                <button type="button" class="btn ter sm" onclick={addToWallet} disabled={addingChain}><span class="ic">{@html icon(networkAdded ? 'check' : 'plus')}</span>{networkAdded ? 'Network added' : 'Add GullyChain Testnet'}</button>
                <a class="btn ter sm glow" href={LINKS[0].href} target="_blank" rel="noopener noreferrer"><span class="ic">{@html icon('rocket')}</span>Start building<span class="ic ar">{@html icon('arrow')}</span></a>
              </div>
            </div>
          {/if}
        </form>

      </div>

      <div class="term wall" aria-labelledby="h-wall" data-testid="drips">
        <div class="th"><span class="lab cy" id="h-wall"><span class="ic sm">{@html icon('droplet')}</span>Faucet drips · last {DRIP_DAYS} days</span><div class="r">UTC</div></div>
        <div class="kpis">
          <div class="kpi" title="Transactions the faucet account has ever sent (its on-chain nonce)"><b class="num" data-testid="drips-total">{shownNonce == null ? '–' : shownNonce.toLocaleString('en-US')}</b><span>{#if shownNonce == null && rpcDown}RPC not answering{:else}payouts all time{/if}</span></div>
          <div class="kpi"><b class="num">{dripsToday == null ? '–' : dripsToday}</b><span>today{#if dripDays.length}{' · '}{dripsWin} in {DRIP_DAYS} days{/if}</span></div>
          <div class="kpi" class:warn={claimsLeft != null && claimsLeft < 20} title="Faucet balance divided by one claim"><b class="num">{claimsLeft == null ? '–' : claimsLeft.toLocaleString('en-US')}</b><span>{balance == null ? (rpcDown ? 'RPC not answering' : 'claims left') : `claims left · ${fmtEth(balance)} ETH`}</span></div>
        </div>
        <div class="plot" bind:clientWidth={plotW}>
          <svg class="main" viewBox="0 0 {CW} {CH}" height={CH} role="img" aria-label={dripDays.length ? `Faucet payouts per day, last ${DRIP_DAYS} days: ${dripsWin} payouts, ${fmtEth(ethWin)} ETH` : 'Faucet payouts per day: no data yet'} onmouseleave={() => (tipOn = false)}>
            {#each [0, 1, 2, 3, 4] as i}
              <line class="gl" x1={ML} x2={CW - MR} y1={MT + (CH - MT - MB) * (1 - i / 4)} y2={MT + (CH - MT - MB) * (1 - i / 4)} />
              <text class="at" data-testid="ytick" x={ML - 6} y={MT + (CH - MT - MB) * (1 - i / 4) + 4} text-anchor="end">{Math.round(gridMax * i / 4)}</text>
            {/each}
            {#if bars.length}
              <text class="at" x={ML} y={CH - 8}>{dayLabel(bars[0].day)}</text>
              <text class="at" x={CW - MR} y={CH - 8} text-anchor="end">Today</text>
            {:else}
              <text class="at empty" x={CW / 2} y={CH / 2} text-anchor="middle">{dripsError ? 'No chart: the explorer index is not answering' : 'Loading payouts…'}</text>
            {/if}
            {#each bars as b (b.day)}
              <rect class="hit" x={b.x - 3} width={b.w + 6} y={MT} height={CH - MT - MB} onmousemove={(e) => showTip(e, b)} />
              <rect class="bk" class:today={b.today} class:nodata={b.n === null} x={b.x} width={b.w} y={b.n ? b.y : CH - MB - 2} height={b.n ? Math.max(0, b.h) : 2} rx={Math.min(3, b.w / 2)} />
            {/each}
          </svg>
        </div>
        {#if dripsError}<p class="note" role="status" data-testid="drips-error">{dripsError}</p>{:else if !dripsLoading && !dripsWin}<p class="note">No payouts in the last {DRIP_DAYS} days. Yours can be the first.</p>{/if}
        {#if dripsLoading}<div class="recent-sk" aria-hidden="true"></div>{/if}
        {#if recent.length}
          <ul class="recent" aria-label="Latest faucet payouts">
            {#each recent as e (e.hash)}
              <li><span class="ic">{@html icon('droplet')}</span><b class="num">{fmtEth(e.eth)} ETH</b><span>to <a class="mono" href="{CHAIN.explorer}/address/{e.to}" target="_blank" rel="noopener noreferrer">{short(e.to)}</a></span><a class="when" href="{CHAIN.explorer}/tx/{e.hash}" target="_blank" rel="noopener noreferrer">{ago(e.ts, nowTs)} ago</a></li>
            {/each}
          </ul>
        {/if}
        <p class="hint src">Serving from <button type="button" class="addr mono" onclick={() => copyText(faucetInfo.account, 'Faucet address')} title="Copy the faucet address {faucetInfo.account}" aria-label="Copy the faucet address">{short(faucetInfo.account)}</button>{#if faucetInfo.hcaptcha_sitekey}{' · '}protected by hCaptcha{/if}</p>
      </div>
    </div>
  </section>

  <section class="wrap more">
    <div class="group settings">
      <p class="lab cy"><span class="ic">{@html icon('settings')}</span>Network settings</p>
      <dl>
        <div><dt>Network name</dt><dd>{CHAIN.name}</dd></div>
        <div><dt>Chain ID</dt><dd class="num">{CHAIN.id}</dd></div>
        <div><dt>RPC URL</dt><dd class="rpcdd"><span class="mono">{CHAIN.rpc}</span><button class="btn ter xs icon" type="button" aria-label="Copy the RPC URL" title="Copy the RPC URL" onclick={() => copyText(CHAIN.rpc, 'RPC URL')}><span class="ic">{@html icon('copy')}</span></button></dd></div>
        <div><dt>Currency</dt><dd>ETH</dd></div>
        <div><dt>Explorer</dt><dd class="mono"><a href={CHAIN.explorer} target="_blank" rel="noopener noreferrer">{CHAIN.explorer.replace('https://', '')}</a></dd></div>
      </dl>
      <button type="button" class="btn ter sm netadd" onclick={addToWallet} disabled={addingChain}><span class="ic">{@html icon(networkAdded ? 'check' : 'plus')}</span>{networkAdded ? 'Network added' : 'Add GullyChain Testnet'}</button>
    </div>
    <div class="group band">
      <div class="band-copy">
        <p class="lab cy"><span class="ic">{@html icon('arrows')}</span>Need more than {faucetInfo.payout} {faucetInfo.symbol}?</p>
        <h2>Bridge from Arbitrum Sepolia</h2>
        <p class="muted">Deposit Arbitrum Sepolia ETH through the testnet bridge.</p>
        {#if faucetInfo.dev_tier}<p class="muted" data-testid="dev-tier">Building on GullyChain? Verified developers get <b>{faucetInfo.dev_tier.payout} {faucetInfo.symbol}</b> every {faucetInfo.dev_tier.window_hours} h with the same Gully ID sign-in. Email <a href="mailto:team@gullylabs.xyz?subject=Faucet%20developer%20tier">team@gullylabs.xyz</a> with your Gully ID and what you are building.</p>{/if}
        <div class="ctas">
          <a class="btn ter sm" href={LINKS[1].href} target="_blank" rel="noopener noreferrer"><span class="ic">{@html icon('arrows')}</span>Bridge<span class="ic ar">{@html icon('arrow')}</span></a>
          <a class="btn ter sm" href={LINKS[0].href} target="_blank" rel="noopener noreferrer"><span class="ic">{@html icon('rocket')}</span>Start building<span class="ic ar">{@html icon('arrow')}</span></a>
        </div>
      </div>
      <div class="mascot" class:on={mascotOn} bind:this={mascotEl} aria-hidden="true">
        <!-- kit v4 mascots/mascot-builder.png (630x502, byte for byte). The file itself is cut straight at its right edge, so it overhangs this frame and the cut stays hidden -->
        <img src="/brand/mascots/mascot-builder.png" width="630" height="502" alt="" loading="lazy" decoding="async" />
      </div>
    </div>
  </section>
</main>

<footer>
  <div class="wrap">
    <div><a class="fwm" href="https://gullychain.xyz" aria-label="Gully Chain home"><img src="/brand/logo/gc-icon.svg" width="32" height="32" alt="Gully Chain" loading="lazy" /></a><span class="tag">Testnet faucet · chain {CHAIN.id}</span></div>
    <nav class="fnav" aria-label="GullyChain products"><ul>{#each PRODUCTS as l}<li><a href={l.href}><span class="ic">{@html icon(l.icon)}</span>{l.name}</a></li>{/each}<li><a href={SOURCE} target="_blank" rel="noopener noreferrer"><span class="ic">{@html icon('branch')}</span>Source</a></li><li><a href={ECOSYSTEM} target="_blank" rel="noopener noreferrer"><span class="ic">{@html icon('layers')}</span>Gully ecosystem →</a></li></ul></nav>
    <ul class="fnav" aria-label="Legal" data-testid="legal">{#each LEGAL as l}<li><a href={l.href} target="_blank" rel="noopener noreferrer"><span class="ic">{@html icon(l.icon)}</span>{l.name}<span class="gn-sr"> (opens in a new tab)</span></a></li>{/each}</ul>
    <ul class="fnav" aria-label="Official GullyChain channels">{#each SOCIALS as l}<li><a href={l.href} target="_blank" rel="noopener noreferrer"><span class="ic">{@html icon(l.icon)}</span>{l.name}</a></li>{/each}</ul>
    <div class="fbot">
      <p class="copy2">Testnet only. Tokens have no value. MIT, a fork of chainflag/eth-faucet.</p>
      <span class="fbadges"><a class="gcb" href="https://gullychain.xyz" aria-label="Powered by GullyChain"><img src="/brand/badges/badge-dark.png" width="180" height="48" alt="Powered by GullyChain" loading="lazy" /></a>
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
        {#if t.message}<p>{t.message}{#if t.link}{' · '}<a href={t.link.href} target="_blank" rel="noopener noreferrer">{t.link.text}</a>{/if}</p>{/if}
      </div>
      <button type="button" class="x" onclick={() => dismiss(t.id)} aria-label="Dismiss">{@html icon('x')}</button>
      <span class="pb" style="animation-duration:{t.ms}ms"></span>
    </div>
  {/each}
</div>

<style>
  /* Gully Chain Night (GN, gully-ops sites/_shared/gullychain-night): black ground, ink cards with a muted 18% line,
     Inter (GN 2.0 = brand kit v4), emerald accent, mint for live and payouts, signature pill buttons, live animated charts (bars grow,
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
  .logo img { width: 32px; height: 32px; max-width: none; }
  .topnav { flex: 1 1 auto; min-width: 0; display: flex; justify-content: center; }
  .topnav ul { list-style: none; margin: 0; padding: 0; display: flex; gap: 2px; }
  .topnav a { display: inline-flex; align-items: center; gap: 8px; min-height: 40px; padding: 0 12px; border-radius: var(--gn-r-pill); color: var(--gn-text-2); font-size: 14px; text-decoration: none; white-space: nowrap; }
  /* nav icons (owner rule 2026-10-30): Lucide 20 px, stroke 1.75, muted by default, mint when active or hovered */
  .topnav a .ic, .mnav a .ic, footer .fnav a .ic { width: 20px; height: 20px; color: var(--gn-text-3); }
  .topnav a[aria-current="page"] .ic, .mnav a[aria-current="page"] .ic, footer .fnav a[aria-current="page"] .ic { color: var(--gn-link); }
  .topnav a:hover { color: var(--gn-text); background: var(--gn-surface-2); }
  .topnav a:hover .ic, .mnav a:hover .ic, footer .fnav a:hover .ic { color: var(--gn-link); }
  .menubtn { display: none; flex: none; align-items: center; justify-content: center; width: 40px; height: 40px; padding: 0; border: 1px solid var(--border); border-radius: var(--gn-r-pill); background: var(--gn-surface); color: var(--gn-text-2); cursor: pointer; }
  .menubtn .ic { width: 20px; height: 20px; color: currentColor; }
  .menubtn:hover { color: var(--gn-text); }
  .mnav { border-top: 1px solid var(--border); background: var(--gn-black); }
  .mnav ul { list-style: none; margin: 0 auto; padding: 8px var(--gn-gutter) 12px; max-width: var(--gn-maxw); display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 4px; }
  .mnav a { display: flex; align-items: center; gap: 10px; min-height: 44px; padding: 0 12px; border-radius: var(--gn-r-sm); color: var(--gn-text-2); font-size: 15px; }
  .mnav a:hover { color: var(--gn-text); background: var(--gn-surface-2); }
  @media (min-width: 900px) { .mnav { display: none; } }
  @media (max-width: 899px) { .menubtn { display: inline-flex; } }
  footer .fnav a { display: inline-flex; align-items: center; gap: 6px; }
  .topnav a:focus-visible { outline: none; box-shadow: var(--gn-focus); }
  @media (max-width: 1239px) { .topnav .tl { display: none; } .topnav a { padding: 0 10px; } }
  @media (max-width: 899px) { .topnav { display: none; } }
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
  .copy h1 em.hold { visibility: hidden; }
  .copy h1 em { font-style: normal; font-variant-numeric: tabular-nums; background: var(--gn-gradient); -webkit-background-clip: text; background-clip: text; color: transparent; }
  .copy .promise { color: var(--text-2); font-size: 15px; max-width: 540px; margin: 0; line-height: var(--gn-lh); }
  .copy .promise b { color: var(--text); }
  .ctas { display: flex; flex-wrap: wrap; gap: 10px; }

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
  .addr-h { display: flex; align-items: center; justify-content: space-between; gap: 8px; flex-wrap: wrap; margin-top: 4px; }
  .addr-h .sn { width: 20px; height: 20px; border-radius: 50%; display: inline-grid; place-items: center; background: var(--gn-surface-2); color: var(--gn-text); font-size: 11px; letter-spacing: 0; }
  .addr-hint { margin-top: -4px; }
  .input { flex: 1 1 auto; min-width: 0; min-height: var(--gn-btn-h); padding: 0 18px; border-radius: var(--gn-r-pill); border: 1px solid var(--border-2); background: var(--surface-2); color: var(--text); font-size: 14px; outline: none; transition: border-color var(--gn-dur), box-shadow var(--gn-dur); }
  .input:not(:focus) { text-overflow: ellipsis; }
  .input::placeholder { color: var(--muted-c); opacity: 1; font-family: var(--gn-font); }
  .input:focus { border-color: var(--gn-cyan); box-shadow: 0 0 0 3px var(--gn-info-bg); }
  .input[aria-invalid='true'] { border-color: var(--danger); }
  .ferr { margin: 0; font-size: 13px; color: var(--danger); }
  /* app link (?to=) lock + "not your kre8 / Gully ID account" warning (Gully Chain Night warn tokens) */
  .lockline { margin: 0; display: block; font-size: 13px; line-height: 1.6; color: var(--text-2); }
  .lockline .ic { display: inline-flex; width: 16px; height: 16px; margin-right: 6px; vertical-align: -3px; color: var(--ok); }
  .lockline .linkbtn { margin-left: 6px; }
  .lockline .mono { color: var(--gn-text); }
  .linkbtn { all: unset; cursor: pointer; color: var(--link); font-weight: var(--gn-w-bold); text-decoration: underline; text-underline-offset: 2px; }
  .linkbtn:focus-visible { box-shadow: var(--gn-focus); border-radius: 4px; }
  .twarn { display: grid; gap: 8px; justify-items: start; padding: 12px 14px; border-radius: var(--gn-r); border: 1px solid color-mix(in srgb, var(--gn-warn) 45%, transparent); background: var(--gn-warn-bg); }
  .twarn-h { margin: 0; display: flex; align-items: center; gap: 8px; font-weight: var(--gn-w-bold); color: var(--gn-text); }
  .twarn-h .ic { width: 18px; height: 18px; color: var(--gn-warn); flex: none; }
  .twarn-s { margin: 0; font-size: 13px; color: var(--text-2); }
  .twarn .mono { font-family: var(--f-mono); }
  /* sign-in optional: paying another wallet is normal, so the note is neutral (no warn colours) */
  .twarn.soft { border-color: var(--border); background: var(--surface-2); }
  .twarn.soft .twarn-h .ic { color: var(--muted-c); }
  /* optional "Sign in with Gully ID": a secondary block under the address and the claim button */
  .gidopt { border-top: 1px solid var(--border); padding-top: 14px; margin-top: 4px; }
  .gidopt[hidden] { display: none; }
  .hint { margin: 0; font-size: 13px; color: var(--muted-c); overflow-wrap: anywhere; }
  .hint.consent a { color: var(--text-2); text-decoration: underline; text-underline-offset: 2px; }
  .hint.consent a:hover { color: var(--text); }
  .addr { background: none; border: 0; padding: 0; color: var(--text-2); font-size: 12.5px; cursor: copy; text-decoration: underline dashed var(--border-2); text-underline-offset: 3px; overflow-wrap: anywhere; word-break: break-all; text-align: left; }
  .addr:hover { text-decoration-color: var(--gn-cyan); color: var(--link); }

  /* chart card */
  .term { border: 1px solid var(--border); border-radius: var(--gn-r-lg); background: var(--surface); position: relative; min-width: 0; display: flex; flex-direction: column; overflow: hidden; }
  .term .th { display: flex; align-items: center; justify-content: space-between; gap: 10px; padding: 14px 16px 0; min-height: 34px; }
  .term .th .r { display: flex; align-items: center; gap: 8px; font-size: 12px; color: var(--muted-c); font-variant-numeric: tabular-nums; }
  .wall .plot { margin: 8px 16px 6px; position: relative; min-height: 200px; }
  .wall svg.main { width: 100%; height: 200px; display: block; overflow: visible; }
  .gl { stroke: var(--gn-grid); stroke-width: 1; }
  .at { fill: var(--muted-c); font-family: var(--gn-font); font-weight: var(--gn-w-regular); font-size: 11px; }
  .at.empty { font-size: 13px; }
  .hit { fill: transparent; }
  .bk { fill: var(--gn-c1); pointer-events: none; transition: fill var(--gn-dur); }
  .bk.today { fill: var(--gn-c2); }
  .bk.nodata { fill: var(--gn-grid); }
  .hit:hover + .bk { fill: var(--gn-link-hover); }
  .tip { position: fixed; z-index: 60; pointer-events: none; background: var(--gn-surface-2); border: 1px solid var(--border-2); border-radius: var(--gn-r); padding: 8px 12px; font-size: 13px; color: var(--text); box-shadow: var(--gn-shadow-pop); opacity: 0; transition: opacity 0.12s; white-space: nowrap; }
  .tip.on { opacity: 1; }
  .tip :global(b) { font-weight: var(--gn-w-bold); font-variant-numeric: tabular-nums; }
  .tip :global(span) { color: var(--muted-c); }
  /* drips: three live numbers, the chart, the latest payouts */
  .kpis { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 12px; padding: 10px 16px 0; }
  .kpi { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
  .kpi b { font-size: var(--gn-fs-stat); letter-spacing: -0.02em; line-height: 1.1; color: var(--text); }
  .kpi span { font-size: 12px; color: var(--muted-c); }
  .kpi.warn b { color: var(--warn); }
  .recent-sk { height: 145px; border-top: 1px solid var(--border); }
  .note { margin: 0; padding: 0 16px 8px; font-size: 13px; color: var(--text-2); }
  .recent { list-style: none; margin: 0; padding: 0 16px; border-top: 1px solid var(--border); }
  .recent li { display: flex; align-items: center; gap: 8px; min-height: 36px; font-size: 13px; color: var(--text-2); border-top: 1px solid var(--border); font-variant-numeric: tabular-nums; }
  .recent li:first-child { border-top: 0; }
  .recent b { color: var(--text); }
  .recent .ic { width: 14px; height: 14px; color: var(--gn-green); }
  .recent .when { margin-left: auto; color: var(--muted-c); white-space: nowrap; }
  .recent .when:hover { color: var(--link); }
  .hint.src { padding: 10px 16px 14px; border-top: 1px solid var(--border); font-size: 12px; }

  /* settings + bridge band; the builder mascot on night */
  .more { display: grid; grid-template-columns: 1fr; gap: 14px; margin: 14px auto 40px; }
  .group { background: var(--surface); border: 1px solid var(--border); border-radius: var(--gn-r-lg); padding: 18px; min-width: 0; }
  .settings .lab { margin-bottom: 8px; }
  dl { margin: 0; }
  dl div { display: grid; grid-template-columns: 100px minmax(0, 1fr); gap: 12px; padding: 8px 0; border-top: 1px solid var(--border); font-size: 14px; }
  dt { color: var(--muted-c); }
  dd { margin: 0; color: var(--text); overflow-wrap: anywhere; }
  .band { display: grid; grid-template-columns: minmax(0, 1fr) 260px; gap: 16px; align-items: end; }
  .band-copy { align-self: start; }
  .band-copy { display: flex; flex-direction: column; gap: 10px; min-width: 0; }
  .band h2 { font-size: 24px; }
  .band .muted { font-size: 14px; }
  .band .ctas { margin-top: 4px; }
  .mascot { position: relative; aspect-ratio: 630 / 502; border-radius: var(--gn-r); background: var(--gn-night); overflow: hidden; pointer-events: none; }
  .mascot img { position: absolute; right: -8px; bottom: -6px; width: calc(100% + 8px); height: calc(100% + 2px); max-width: none; object-fit: contain; object-position: right bottom; transition: transform 0.3s ease; }
  .mascot.on img { animation: mfloat 6s ease-in-out infinite; }
  @keyframes mfloat { 50% { transform: translate3d(0, -3px, 0); } }

  /* footer (true black band) */
  footer { border-top: 1px solid var(--border); padding: 24px 0 20px; background: var(--gn-black); color: var(--muted-c); }
  footer .wrap { display: flex; flex-wrap: wrap; justify-content: space-between; gap: 14px; align-items: center; }
  footer ul { list-style: none; margin: 0; padding: 0; display: flex; flex-wrap: wrap; gap: 16px; font-size: 14px; }
  footer ul a { color: var(--text-2); }
  footer ul a:hover { color: var(--text); }
  footer .tag { font-size: 12px; color: var(--muted-c); margin-top: 8px; display: block; }
  footer .copy2 { font-size: 12px; color: var(--muted-c); width: 100%; margin: 0; }
  /* brand kit v4 logo/gc-icon.svg (byte for byte); kit v4 has no wordmark file */
  footer .fwm { display: inline-flex; align-items: center; line-height: 0; border-radius: var(--gn-r-sm); }
  footer .fwm img { display: block; flex: none; max-width: none; width: 32px; height: 32px; }
  .fbadges { display: inline-flex; align-items: center; gap: 12px; flex-wrap: wrap; }
  .gcb { display: inline-flex; line-height: 0; flex: none; border-radius: var(--gn-r-sm); }
  .gcb img { width: 180px; height: 48px; }
  footer .fbot { width: 100%; display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 10px 16px; padding-top: 14px; border-top: 1px solid var(--border); }
  footer .fbot .copy2 { width: auto; }
  .glb { display: inline-flex; line-height: 0; flex: none; border-radius: var(--gn-r-sm); }
  .glb img { height: 28px; width: auto; }
  @media (max-width: 640px) { .glb img { height: 26px; } }

  /* claim result (ETH + token drip) */
  .claim-result { padding: 10px 14px; border: 1px solid color-mix(in srgb, var(--gn-green) 28%, transparent); border-radius: var(--gn-r); background: var(--ok-bg); font-size: 14px; display: grid; gap: 6px; }
  .claim-result p { margin: 0; display: flex; align-items: center; gap: 8px; flex-wrap: wrap; word-break: break-word; }
  .claim-result .ic.ok { color: var(--ok); }
  .claim-result a { color: var(--link); }
  .claim-result .muted { color: var(--text-2); }

  /* done step */
  .done { display: grid; gap: 10px; }
  .done .claim-result p:first-child { font-size: 16px; font-weight: var(--gn-w-bold); color: var(--text); }
  /* before /api/info answers: the sign-in slot keeps its place, so nothing jumps */
  .sk { height: 118px; border-radius: var(--gn-r); background: var(--surface-2); }
  .inforr { display: grid; gap: 10px; justify-items: start; }
  .inforr p { margin: 0; color: var(--text-2); font-size: 14px; }
  .netadd { margin-top: 10px; }
  .rpcdd { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
  .rpcdd .mono { min-width: 0; overflow-wrap: anywhere; }
  .btn.icon { flex: none; width: 32px; min-height: 32px; padding: 0; }
  .btn.icon .ic { width: 16px; height: 16px; }

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
  @media (min-width: 900px) { .hero { padding: 48px 0 40px; } .hero .wrap { grid-template-columns: minmax(0, 2fr) minmax(0, 3fr); gap: 32px; align-items: start; } }
  @media (max-width: 899px) { .band { grid-template-columns: minmax(0, 1fr) 200px; } }
  @media (max-width: 640px) { .claim-result a { display: inline-flex; align-items: center; min-height: 44px; } footer ul { row-gap: 0; } .rpcdd .mono { font-size: 12px; } .toast .x { width: 44px; height: 44px; margin: -10px -10px 0 0; } .glb, .gcb { min-height: 44px; align-items: center; } footer ul a { display: inline-flex; align-items: center; min-height: 44px; min-width: 44px; } .btn.sm, .btn.xs { min-height: 44px; } .btn.icon { width: 44px; } .addr { display: inline-flex; align-items: center; min-height: 44px; padding: 0; } .logo, .fwm { min-width: 44px; min-height: 44px; align-items: center; justify-content: center; } .menubtn { width: 44px; height: 44px; } .recent li { min-height: 44px; } .recent .when { display: inline-flex; align-items: center; min-height: 44px; min-width: 44px; justify-content: flex-end; } .band { grid-template-columns: 1fr; } .mascot { width: 100%; max-width: 300px; justify-self: end; } .hide-s { display: none; } .field { flex-direction: column; } .btn.pri { width: 100%; } .ctas .btn { flex: 1 1 auto; } .tape .trk { animation: none; padding-left: 90px; overflow-x: auto; } dl div { grid-template-columns: 1fr; gap: 2px; } .stepper li + li::before { width: 12px; margin: 0 6px; } .band { grid-template-columns: minmax(0, 1fr); } .mascot { min-height: 170px; } }
  /* below 400 px the header holds the kit logo and the switcher only; the Latest block tile carries the live number */
  @media (max-width: 399px) { .navr .pill { display: none; } }
  @media (prefers-reduced-motion: reduce) { .pill .dot, .toast, .btn::after, .spinner, .mascot.on img { animation: none !important; } .toast .pb { display: none; } .btn, .btn .ar, .mascot img { transition: none; } .btn:hover:not(:disabled) .ar, .group.band:hover .mascot img { transform: none; } }
</style>
