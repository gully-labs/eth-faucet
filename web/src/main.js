// Gully Chain Night: the shared black design (gully-ops sites/_shared/gullychain-night, vendored by
// scripts/sync-gullychain-night.mjs; never edit src/night/ by hand)
import './night/tokens.css';
import './night/components.css';
import { hydrate } from 'svelte';
import App from './App.svelte';

const target = document.getElementById('app');
if (!target) {
  throw new Error('App container element not found');
}

// The page is already in the HTML (scripts/prerender.mjs): take it over in place.
const app = hydrate(App, {
  target,
});

// Repeat visits paint from cache (load/build-sw.mjs). Registered after load so it never competes with first paint.
if (import.meta.env.PROD && 'serviceWorker' in navigator) {
  addEventListener('load', () => void navigator.serviceWorker.register('/sw.js').catch(() => undefined));
}

export default app;
