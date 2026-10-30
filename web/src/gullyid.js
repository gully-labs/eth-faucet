// Lazy loader for the Gully ID browser SDK (exposes window.GullyID): `<issuer>/sdk/v1.js` (the 1.x channel), falling
// back to the legacy `<issuer>/sdk.js` while v1.js is not deployed. Contract: gully-id docs/gully-id-current-spec.
// Only called when /api/info carries `gullyid_client`, so the faucet never loads a third script while the flag is
// off. Any failure (blocked, offline, SDK not deployed yet) rejects and the caller hides the button; the faucet keeps
// working as before.

export const DEFAULT_GULLY_ID_ISSUER = 'https://gullyid.xyz';

/** SDK URLs, best first. */
export function sdkUrls(issuer = DEFAULT_GULLY_ID_ISSUER) {
  const origin = String(issuer || DEFAULT_GULLY_ID_ISSUER).replace(/\/+$/, '');
  return [origin + '/sdk/v1.js', origin + '/sdk.js'];
}

let pending = null;

export function loadGullyID(issuer = DEFAULT_GULLY_ID_ISSUER, timeoutMs = 10000) {
  if (typeof window === 'undefined') return Promise.reject(new Error('no window'));
  if (window.GullyID) return Promise.resolve(window.GullyID);
  if (pending) return pending;

  pending = new Promise((resolve, reject) => {
    const timer = setTimeout(() => fail(new Error('Gully ID SDK timed out')), timeoutMs);
    let current = null;
    function fail(err) {
      clearTimeout(timer);
      if (current) { current.onload = current.onerror = null; current.remove(); }
      pending = null;
      reject(err);
    }
    function tryNext(urls, lastErr) {
      if (!urls.length) return fail(lastErr || new Error('Gully ID SDK failed to load'));
      const s = document.createElement('script');
      current = s;
      s.src = urls[0];
      s.async = true;
      s.referrerPolicy = 'strict-origin';
      s.onload = () => {
        s.onload = s.onerror = null;
        if (window.GullyID && typeof window.GullyID.init === 'function') { clearTimeout(timer); return resolve(window.GullyID); }
        s.remove();
        tryNext(urls.slice(1), new Error('Gully ID SDK loaded without window.GullyID'));
      };
      s.onerror = () => {
        s.onload = s.onerror = null;
        s.remove();
        tryNext(urls.slice(1), new Error('Gully ID SDK failed to load'));
      };
      document.head.appendChild(s);
    }
    tryNext(sdkUrls(issuer));
  });
  return pending;
}

export function shortAddress(a) {
  return a && a.length > 12 ? a.slice(0, 6) + '…' + a.slice(-4) : a || '';
}
