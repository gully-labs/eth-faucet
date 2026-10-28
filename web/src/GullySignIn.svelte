<script>
  // "Sign in with Gully" for the faucet. The signed-in user's main wallet pre-fills the address field (through
  // `onwallet`) and the .gully name / avatar is shown. `onwallet` also gets the session: when the faucet requires
  // sign-in (/api/info signin_required) its access token goes to /api/claim as `Authorization: Bearer`, and the server
  // checks it with Gully ID (one claim per account per day).
  // Shown only when /api/info has `gullyid_client` (Go flag -gullyid.client / GULLY_ID_CLIENT_ID). Fails soft: if the
  // SDK cannot load, nothing is rendered.
  import { onMount, onDestroy } from 'svelte';
  import { loadGullyID, shortAddress, DEFAULT_GULLY_ID_ISSUER } from './gullyid.js';

  let { clientId, issuer = DEFAULT_GULLY_ID_ISSUER, required = false, onwallet = () => {} } = $props();

  let host = $state(null);
  let ready = $state(false);
  let user = $state(null);
  const cleanups = [];
  let destroyed = false;

  const safeAvatar = (u) => (typeof u === 'string' && /^https:\/\//i.test(u) ? u : null);

  function apply(session) {
    user = session?.user || null;
    const address = user?.wallet?.address || null;
    onwallet(address, user, session || null);
  }

  onMount(async () => {
    try {
      const GullyID = await loadGullyID(issuer);
      if (destroyed) return;
      const opts = { clientId };
      if (issuer && issuer !== DEFAULT_GULLY_ID_ISSUER) opts.issuer = issuer;
      const gully = GullyID.init(opts);
      try {
        await gully.handleRedirect();
      } catch (e) {
        console.warn('Gully ID: redirect sign-in did not finish', e);
      }
      if (destroyed) return;
      const off = gully.onChange?.(apply);
      if (typeof off === 'function') cleanups.push(off);
      apply(gully.session());
      ready = true;
      const unmount = gully.renderButton(host, {
        onSignIn: apply,
        onError: (e) => console.warn('Gully ID sign-in failed', e),
        passkey: true,
        theme: 'dark',
      });
      if (typeof unmount === 'function') cleanups.push(unmount);
    } catch (e) {
      ready = false;
      console.warn('Sign in with Gully is unavailable', e);
    }
  });

  onDestroy(() => {
    destroyed = true;
    for (const fn of cleanups.splice(0)) {
      try { fn(); } catch {}
    }
  });
</script>

<div class="gid" class:on={ready} aria-live="polite">
  <div class="gid-btn" bind:this={host}></div>
  {#if ready}
    {#if user}
      <p class="gid-who">
        {#if safeAvatar(user.avatar)}
          <img class="gid-av" src={safeAvatar(user.avatar)} alt="" width="22" height="22" referrerpolicy="no-referrer" />
        {/if}
        <span>Signed in as <b>{user.gullyName || user.name}</b>{#if user.wallet?.address} · address filled with your main wallet <span class="mono">{shortAddress(user.wallet.address)}</span>{:else} · no wallet linked yet{/if}</span>
      </p>
    {:else}
      {#if required}
        <p class="gid-who"><span><b>Sign in with Gully to claim.</b> One claim per Gully ID a day keeps the faucet from being drained by bots.</span></p>
      {:else}
        <p class="gid-who">Sign in to fill in your Gully wallet. Optional: any address works.</p>
      {/if}
    {/if}
  {/if}
</div>

<style>
  .gid { display: none; }
  .gid.on { display: flex; flex-wrap: wrap; align-items: center; gap: 8px 12px; padding-bottom: 10px; border-bottom: 1px solid var(--gc-line); }
  .gid-btn:empty { display: none; }
  .gid-who { margin: 0; flex: 1 1 220px; min-width: 0; display: inline-flex; align-items: center; gap: 8px; font-size: 13px; color: var(--gc-muted); overflow-wrap: anywhere; }
  .gid-who b { color: var(--gc-white); }
  .gid-who .mono { font-family: var(--f-mono); color: var(--gc-white); }
  .gid-av { width: 22px; height: 22px; border-radius: 50%; object-fit: cover; flex: none; border: 1px solid var(--gc-line); }
</style>
