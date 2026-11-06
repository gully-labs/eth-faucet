// The prerender: the faucet page as real HTML at build time (scripts/prerender.mjs). onMount does not run here, so nothing
// waits for /api/info, a wallet or a Gully ID session.
import { render } from 'svelte/server';
import App from './App.svelte';

export const renderApp = () => render(App);
