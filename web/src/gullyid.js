// Lazy loader for the Gully ID browser SDK (`<issuer>/sdk.js`, exposes window.GullyID).
// Contract: gully-id/docs/sign-in-with-gully.md. Only called when /api/info carries `gullyid_client`, so the faucet
// never loads a third script while the flag is off. Any failure (blocked, offline, SDK not deployed yet) rejects and
// the caller hides the button; the faucet keeps working as before.

export const DEFAULT_GULLY_ID_ISSUER = 'https://id.gullychain.xyz';

let pending = null;

export function loadGullyID(issuer = DEFAULT_GULLY_ID_ISSUER, timeoutMs = 10000) {
  if (typeof window === 'undefined') return Promise.reject(new Error('no window'));
  if (window.GullyID) return Promise.resolve(window.GullyID);
  if (pending) return pending;

  const origin = String(issuer || DEFAULT_GULLY_ID_ISSUER).replace(/\/+$/, '');
  pending = new Promise((resolve, reject) => {
    const s = document.createElement('script');
    s.src = origin + '/sdk.js';
    s.async = true;
    s.referrerPolicy = 'strict-origin';
    const timer = setTimeout(() => done(new Error('Gully ID SDK timed out')), timeoutMs);
    function done(err) {
      clearTimeout(timer);
      s.onload = s.onerror = null;
      if (!err && window.GullyID && typeof window.GullyID.init === 'function') return resolve(window.GullyID);
      pending = null;
      s.remove();
      reject(err || new Error('Gully ID SDK loaded without window.GullyID'));
    }
    s.onload = () => done();
    s.onerror = () => done(new Error('Gully ID SDK failed to load'));
    document.head.appendChild(s);
  });
  return pending;
}

export function shortAddress(a) {
  return a && a.length > 12 ? a.slice(0, 6) + '…' + a.slice(-4) : a || '';
}
