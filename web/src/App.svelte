<script>
  import Faucet from './Faucet.svelte';
</script>

<Faucet />

<style>
  /* GullyChain brand kit v3 tokens (brand-kit/tokens.json + CLAUDE_PROMPT.md). Values are exact; do not invent new ones.
     --st-warn / --st-down are the two state colours the kit does not define (warning, down). */
  :global(:root) {
    --gc-night: #05080d;
    --gc-ink: #0a1018;
    --gc-blue: #2563ff;
    --gc-cyan: #00b2fa;
    --gc-green: #46fb04;
    --gc-white: #f2f5f7;
    --gc-muted: #8a97a8;
    --gc-line: rgba(138, 151, 168, 0.18);
    --gc-chain-gradient: linear-gradient(90deg, #00b2fa, #46fb04);
    --gc-font: 'Exo 2', system-ui, sans-serif;
    --f-mono: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
    --gc-radius-card: 22px;
    --gc-radius-pill: 9999px;
    --st-warn: #f5a623;
    --st-down: #c0392b;
  }
  :global(*),
  :global(*::before),
  :global(*::after) {
    box-sizing: border-box;
  }
  :global(html) {
    background: var(--gc-night);
  }
  :global(body) {
    margin: 0;
    min-height: 100vh;
    background: var(--gc-night);
    color: var(--gc-white);
    font-family: var(--gc-font);
    font-weight: 600;
    font-size: 15px;
    line-height: 1.45;
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
