<script>
  import Faucet from './Faucet.svelte';
</script>

<Faucet />

<style>
  /* GullyChain brand kit v2 tokens (brand-kit-v2/CLAUDE_PROMPT.md). Values are exact; do not invent new ones. */
  :global(:root) {
    --gc-black: #000000;
    --gc-night: #0a0f1a;
    --gc-line: #16213a;
    --gc-royal: #1567f9;
    --gc-electric: #0f89f7;
    --gc-cyan: #01d9fa;
    --gc-indigo: #2e36f7;
    --gc-text: #ffffff;
    --gc-muted: #9fb3cc;
    --gc-up: #2e7d32;
    --gc-down: #c0392b;
    --gc-chain-gradient: linear-gradient(90deg, #1567f9 0%, #0f89f7 35%, #01d9fa 100%);
    --gc-icon-gradient: linear-gradient(180deg, #01d9f9 0%, #2e36f7 100%);
    --gc-radius-card: 22px;
    --gc-radius-pill: 9999px;
    --f-wordmark: 'Poppins', system-ui, sans-serif;
    --f-tagline: 'Montserrat', system-ui, sans-serif;
    --f-badge: 'Barlow Condensed', system-ui, sans-serif;
    --f-ui: 'Hanken Grotesk', system-ui, -apple-system, 'Segoe UI', sans-serif;
    --f-num: 'IBM Plex Sans', system-ui, sans-serif;
  }
  :global(*),
  :global(*::before),
  :global(*::after) {
    box-sizing: border-box;
  }
  :global(html) {
    background: var(--gc-black);
  }
  :global(body) {
    margin: 0;
    min-height: 100vh;
    background: var(--gc-black);
    color: var(--gc-text);
    font-family: var(--f-ui);
    font-size: 16px;
    line-height: 1.5;
    -webkit-font-smoothing: antialiased;
    overflow-x: hidden;
  }
  :global(a) {
    color: var(--gc-cyan);
    text-decoration: none;
  }
  :global(:focus-visible) {
    outline: 2px solid var(--gc-cyan);
    outline-offset: 3px;
    border-radius: 4px;
  }
  :global(img) {
    display: block;
    max-width: 100%;
  }
</style>
