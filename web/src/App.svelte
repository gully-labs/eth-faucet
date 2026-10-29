<script>
  import Faucet from './Faucet.svelte';
</script>

<Faucet />

<style>
  /* Gully Chain Night (GN): every value comes from the shared tokens (src/night/tokens.css, vendored byte for byte from
     gully-ops sites/_shared/gullychain-night). The faucet's older names only alias --gn-* so each rule maps 1:1. */
  :global(:root) {
    --bg: var(--gn-bg); --surface: var(--gn-surface); --surface-2: var(--gn-surface-2); --hover: var(--gn-surface-2);
    --border: var(--gn-line); --border-2: var(--gn-line-strong); --divider: var(--gn-grid);
    --text: var(--gn-text); --text-2: var(--gn-text-2); --muted-c: var(--gn-text-3);
    --accent: var(--gn-cyan); --link: var(--gn-link); --link-h: var(--gn-link-hover);
    --ok: var(--gn-ok); --ok-bg: var(--gn-ok-bg); --info: var(--gn-info); --info-bg: var(--gn-info-bg);
    --warn: var(--gn-warn); --warn-bg: var(--gn-warn-bg); --danger: var(--gn-danger); --danger-bg: var(--gn-danger-bg);
    --c-green: var(--gn-green); --c-blue: var(--gn-c3);
    --gc-night: var(--gn-night); --gc-ink: var(--gn-ink); --gc-cyan: var(--gn-cyan); --gc-green: var(--gn-green);
    --gc-white: var(--gn-text); --gc-muted: var(--gn-text-3); --gc-line: var(--gn-line);
    --gc-chain-gradient: var(--gn-gradient);
    --gc-font: var(--gn-font);
    --f-mono: var(--gn-font-mono);
    --gc-radius-card: var(--gn-r-lg);
    --gc-radius-pill: var(--gn-r-pill);
    --st-warn: var(--warn);
    --st-down: var(--danger);
    color-scheme: dark;
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
    font-family: var(--gn-font);
    font-weight: var(--gn-w-regular);
    font-size: var(--gn-fs-sm);
    line-height: var(--gn-lh);
    -webkit-font-smoothing: antialiased;
    overflow-x: hidden;
  }
  :global(a) {
    color: var(--link);
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
