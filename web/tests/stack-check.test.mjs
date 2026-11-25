// Gully stack check (gully-ops stack/, owner rule 2026-11-25 "Whole-stack sync"): this repo must match the stack registry
// snapshot in .stack/. Never edit .stack/snapshot.json or .stack/stack-check.mjs here: change the registry in gully-ops and run
// stack/tools/propagate.mjs. .stack/checks.json is this repo's own list of what it promises.
import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..');

test('stack check: this repo matches the Gully stack registry', () => {
  let out;
  try {
    out = execFileSync(process.execPath, [path.join(root, '.stack/stack-check.mjs'), '--root', root], { encoding: 'utf8' });
  } catch (e) {
    throw new Error(String(e.stdout || e));
  }
  assert.match(out, /: PASS/);
});
