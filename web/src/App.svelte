<script>
  import Faucet from './Faucet.svelte';
</script>

<Faucet />

<style>
  /* GullyChain brand kit v3 tokens (brand-kit/tokens.json + CLAUDE_PROMPT.md). Values are exact; do not invent new ones.
     --st-warn / --st-down are the two state colours the kit does not define (warning, down). */
  :global(:root) {
    /* Studio Light tokens (gully-ops sites/_shared/studio-light); the kit v3 names alias them */
    --bg: #ffffff; --surface: #f6f7f9; --surface-2: #f4f5f7; --hover: #f1f2f4; --border: #e6e8ec; --border-2: #d5d9e0; --divider: #f0f1f4;
    --text: #0b0d12; --text-2: #3d4350; --muted-c: #5b6270; --accent: #0a86c2; --link: #075f8c; --link-h: #04456a;
    --ok: #0b6640; --ok-bg: #e5f6ee; --info: #075f8c; --info-bg: #e6f4fb; --warn: #7a4a00; --warn-bg: #fff4de; --danger: #b42318;
    --c-green: #16a34a; --c-indigo: #6366f1; --c-amber: #f59e0b;
    --gc-night: var(--bg);
    --gc-ink: #ffffff;
    --gc-blue: var(--c-indigo);
    --gc-cyan: var(--accent);
    --gc-green: var(--ok);
    --gc-white: var(--text);
    --gc-muted: var(--muted-c);
    --gc-line: var(--border);
    --gc-chain-gradient: linear-gradient(90deg, #0a86c2, #16a34a);
    --gc-font: Inter, system-ui, -apple-system, 'Segoe UI', sans-serif;
    --f-mono: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
    --gc-radius-card: 14px;
    --gc-radius-pill: 8px;
    --st-warn: var(--warn);
    --st-down: var(--danger);
    color-scheme: light;
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
    font-weight: 400;
    font-size: 14px;
    line-height: 1.5;
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
