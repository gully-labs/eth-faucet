// Who does this claim pay? Pure helpers for the address field (owner report 2026-10-31: a claim paid the Gully ID main
// wallet while kre8 acts as the user's kre8 passkey account, so kre8 still showed "bal $0.00").
//
//  - parseTarget(search): `?to=<address>` (or the older `?address=`) from an app link, e.g. kre8's "Get test ETH":
//    faucet.gullychain.xyz/?to=0x…&from=kre8. The page fills the field with it and locks it, so signing in with Gully
//    ID no longer swaps it for the main wallet.
//  - targetCheck(...): when signed in with Gully ID, is the address in the field the Gully ID account kre8 uses (the
//    locked app account when one came in the link, else the Gully ID main wallet)? A mismatch is warned about before
//    paying, with a one-tap "Use my Gully account".
import { getAddress } from '@ethersproject/address';

const APPS = { kre8: 'kre8' };

/** checksummed address or null */
export function normAddress(v) {
  if (typeof v !== 'string') return null;
  const t = v.trim();
  if (!/^0x[0-9a-fA-F]{40}$/.test(t)) return null;
  try {
    return getAddress(t.toLowerCase());
  } catch {
    return null;
  }
}

export const sameAddress = (a, b) => {
  const x = normAddress(a);
  const y = normAddress(b);
  return !!x && x === y;
};

/** `?to=` / `?address=` + `?from=` from a query string. Unknown apps are ignored (from = null). */
export function parseTarget(search) {
  let p;
  try {
    p = new URLSearchParams(typeof search === 'string' ? search : '');
  } catch {
    return { to: null, from: null };
  }
  const to = normAddress(p.get('to')) || normAddress(p.get('address'));
  const f = (p.get('from') || '').trim().toLowerCase();
  return { to, from: to && APPS[f] ? APPS[f] : null };
}

export const appName = (from) => (from === 'kre8' ? 'kre8' : 'the app');

/**
 * @param {{ field: string|null, signedIn: boolean, gullyMain: string|null, gullyWallets?: {address:string}[]|null, lockedTo?: string|null, from?: string|null }} o
 * @returns {{ state: 'idle'|'ok'|'mismatch', expected: string|null, title?: string, detail?: string, own?: boolean }}
 *   idle = nothing to compare (not signed in, empty field, an ENS name, or no Gully account to compare with)
 */
export function targetCheck(o) {
  const expected = normAddress(o.lockedTo) || normAddress(o.gullyMain);
  const field = normAddress(o.field);
  if (!o.signedIn || !expected || !field) return { state: 'idle', expected };
  if (field === expected) return { state: 'ok', expected };
  const own = (o.gullyWallets || []).some((w) => sameAddress(w && w.address, field));
  const target = o.lockedTo ? `your ${appName(o.from)} account` : 'your Gully ID account (the one kre8 uses)';
  return {
    state: 'mismatch',
    expected,
    own,
    title: 'This is not your kre8 / Gully ID account',
    detail: own
      ? `This address is one of your Gully ID wallets, but kre8 does not use it, so the ETH will not show in kre8. It pays ${target} instead if you tap "Use my Gully account".`
      : `kre8 will not show this ETH: it is not ${target}. Paying another address is fine, just make sure you meant it.`,
  };
}
