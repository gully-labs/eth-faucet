// Gully Chain Night: the shared black design (gully-ops sites/_shared/gullychain-night, vendored by
// scripts/sync-gullychain-night.mjs; never edit src/night/ by hand)
import './night/tokens.css';
import './night/components.css';
import { mount } from 'svelte';
import App from './App.svelte';

const target = document.getElementById('app');
if (!target) {
  throw new Error('App container element not found');
}

const app = mount(App, {
  target,
});

export default app;
