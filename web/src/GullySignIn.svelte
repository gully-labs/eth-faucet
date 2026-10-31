<script>
  // Step 1 of the claim, "Sign in with Gully ID" (owner 2026-10-30: the sign-in must be visible and prominent).
  // The button is OUR kit v4 primary button and is in the markup from the first paint: it never waits for the Gully ID
  // SDK and is never hidden. Before 2026-10-30 the row stayed display:none until the SDK had loaded and rendered its own
  // button, so a slow, blocked or failed SDK load left the page saying "Sign in with Gully" with no button at all.
  // The SDK (gullyid.js: <issuer>/sdk/v1.js, fallback /sdk.js) is preloaded on mount; a click signs in with the popup
  // (the window opens inside the click), or, if the SDK is not in yet, loads it and uses the redirect flow. A failure
  // shows an inline error with a retry. The signed-in user's main wallet pre-fills the address (through `onwallet`,
  // which also gets the session: with /api/info signin_required its access token goes to /api/claim as
  // Authorization: Bearer, and the server checks it with Gully ID: one claim per account a day).
  import { onMount, onDestroy } from 'svelte';
  import { loadGullyID, shortAddress, DEFAULT_GULLY_ID_ISSUER } from './gullyid.js';

  let { clientId, issuer = DEFAULT_GULLY_ID_ISSUER, required = false, onwallet = () => {} } = $props();

  let gully = $state(null); // GullyID.init(...) once the SDK is in
  let user = $state(null);
  let busy = $state(false);
  let error = $state('');
  const cleanups = [];
  let destroyed = false;

  const safeAvatar = (u) => (typeof u === 'string' && /^https:\/\//i.test(u) ? u : null);

  function apply(session) {
    user = session?.user || null;
    onwallet(user?.wallet?.address || null, user, session || null);
  }

  async function start() {
    const GullyID = await loadGullyID(issuer);
    if (destroyed) return null;
    if (gully) return gully;
    const opts = { clientId };
    if (issuer && issuer !== DEFAULT_GULLY_ID_ISSUER) opts.issuer = issuer;
    const g = GullyID.init(opts);
    try {
      await g.handleRedirect();
    } catch (e) {
      console.warn('Gully ID: redirect sign-in did not finish', e);
    }
    const off = g.onChange?.(apply);
    if (typeof off === 'function') cleanups.push(off);
    gully = g;
    apply(g.session?.() || null);
    return g;
  }

  onMount(() => {
    start().catch((e) => console.warn('Gully ID SDK did not load yet (the button retries on click)', e));
  });

  onDestroy(() => {
    destroyed = true;
    for (const fn of cleanups.splice(0)) {
      try { fn(); } catch {}
    }
  });

  function fail(e) {
    busy = false;
    if (e?.code === 'closed' || e?.code === 4001 || /closed|cancel/i.test(e?.message || '')) { error = ''; return; }
    error = 'Sign-in did not finish: ' + (e?.message || 'Gully ID is unreachable') + '.';
  }

  function signIn() {
    if (busy) return;
    error = '';
    busy = true;
    if (gully) {
      // SDK ready: the popup opens synchronously inside this click
      let p;
      try { p = gully.signIn({ mode: 'popup' }); } catch (e) { return fail(e); }
      Promise.resolve(p).then((s) => { busy = false; if (s) apply(s); }, fail);
      return;
    }
    // SDK not in yet: load it, then the redirect flow (needs no user gesture, so a late load cannot be popup-blocked)
    start().then((g) => {
      if (!g) return fail(new Error('Gully ID is unreachable'));
      return Promise.resolve(g.signIn({ mode: 'redirect' })).then(() => { busy = false; });
    }).catch(fail);
  }

  function signOut() {
    error = '';
    Promise.resolve(gully?.signOut?.()).catch(() => {}).finally(() => apply(null));
  }
</script>

<div class="gid" data-testid="gid-step" aria-live="polite">
  <div class="gid-h">
    <span class="gid-n" class:ok={!!user} aria-hidden="true">{#if user}<svg viewBox="0 0 24 24"><path d="m5 12 5 5L20 7"/></svg>{:else if required}1{:else}<svg viewBox="0 0 24 24"><circle cx="10" cy="7" r="4"/><path d="M10.3 15H7a4 4 0 0 0-4 4v2"/></svg>{/if}</span>
    <p class="gid-t"><b>Sign in with Gully ID</b>{#if required}<span class="gid-req">Required</span>{:else}<span class="gid-opt">Optional</span>{/if}</p>
  </div>
  {#if user}
    <p class="gid-who" data-testid="gid-who">
      {#if safeAvatar(user.avatar)}
        <img class="gid-av" src={safeAvatar(user.avatar)} alt="" width="22" height="22" referrerpolicy="no-referrer" />
      {/if}
      <span>Signed in as <b>{user.gullyName || user.name}</b>{#if user.wallet?.address} · address filled with your main wallet <span class="mono">{shortAddress(user.wallet.address)}</span>{:else} · no wallet linked yet{/if}</span>
      <button type="button" class="gid-link" onclick={signOut}>Sign out</button>
    </p>
  {:else}
    <button type="button" class="gid-btn" data-testid="gid-signin" onclick={signIn} disabled={busy} aria-busy={busy}>
      <svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="10" cy="7" r="4"/><path d="M10.3 15H7a4 4 0 0 0-4 4v2"/><circle cx="17" cy="16" r="3"/><path d="m21 20-1.9-1.9"/></svg>
      <span>{busy ? 'Opening Gully ID…' : 'Sign in with Gully ID'}</span>
      <svg class="ar" viewBox="0 0 24 24" aria-hidden="true"><path d="M5 12h14"/><path d="m12 5 7 7-7 7"/></svg>
    </button>
    <p class="gid-sub">{#if required}Passkey, no seed phrase. One claim per Gully ID a day keeps bots from draining the faucet, and your Gully wallet address is filled in for you.{:else}Fills in your Gully wallet address. Any address works without it.{/if}</p>
  {/if}
  {#if error}<p class="gid-err" role="alert" data-testid="gid-error">{error} <button type="button" class="gid-link" onclick={signIn}>Retry</button></p>{/if}
</div>

<style>
  .gid { display: flex; flex-direction: column; gap: 10px; padding: 14px; border: 1px solid var(--gc-line); border-radius: var(--gn-r); background: var(--gn-ink); }
  .gid-h { display: flex; align-items: center; gap: 10px; }
  .gid-n { width: 24px; height: 24px; flex: none; border-radius: 50%; display: grid; place-items: center; font: 700 12px/1 var(--gn-font); background: var(--gn-emerald); color: var(--gn-night); }
  .gid-n svg { width: 14px; height: 14px; fill: none; stroke: currentColor; stroke-width: 2.5; stroke-linecap: round; stroke-linejoin: round; }
  .gid-t { margin: 0; display: flex; align-items: center; gap: 8px; font-size: 15px; color: var(--gn-text); }
  .gid-req, .gid-opt { font: 600 11px/1 var(--gn-font); padding: 4px 8px; border-radius: var(--gn-r-pill); background: var(--gn-deep-green); color: var(--gn-mint); }
  .gid-opt { background: transparent; border: 1px solid var(--gc-line); color: var(--gn-text-3); }
  .gid-btn { all: unset; box-sizing: border-box; cursor: pointer; display: inline-flex; align-items: center; justify-content: center; gap: 10px; min-height: 48px; padding: 0 22px; border-radius: var(--gn-r-pill); width: 100%;
    background: var(--gn-emerald); color: var(--gn-night); font: 700 15px/1 var(--gn-font); transition: transform var(--gn-dur-fast), filter var(--gn-dur-fast); }
  .gid-btn:hover { filter: brightness(1.08); }
  .gid-btn:active { transform: scale(0.98); }
  .gid-btn:focus-visible, .gid-link:focus-visible { outline: 2px solid var(--gn-mint); outline-offset: 2px; }
  .gid-btn[disabled] { cursor: progress; opacity: 0.8; }
  .gid-btn svg { width: 20px; height: 20px; fill: none; stroke: currentColor; stroke-width: 1.75; stroke-linecap: round; stroke-linejoin: round; flex: none; }
  .gid-btn .ar { transition: transform var(--gn-dur-fast); }
  .gid-btn:hover .ar { transform: translateX(3px); }
  .gid-sub { margin: 0; font-size: 13px; color: var(--gn-text-3); }
  .gid-who { margin: 0; display: flex; flex-wrap: wrap; align-items: center; gap: 8px; font-size: 13px; color: var(--gn-text-3); overflow-wrap: anywhere; }
  .gid-who b { color: var(--gn-text); }
  .gid-who .mono { font-family: var(--f-mono); color: var(--gn-text); }
  .gid-av { width: 22px; height: 22px; border-radius: 50%; object-fit: cover; flex: none; border: 1px solid var(--gc-line); }
  .gid-link { all: unset; cursor: pointer; color: var(--gn-mint); font-weight: 600; text-decoration: underline; text-underline-offset: 3px; }
  .gid-err { margin: 0; font-size: 13px; color: var(--gn-error); }
  @media (prefers-reduced-motion: reduce) { .gid-btn, .gid-btn .ar { transition: none; } .gid-btn:active { transform: none; } }
</style>
