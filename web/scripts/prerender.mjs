#!/usr/bin/env node
// Puts the faucet page into dist/index.html, with the stylesheets inline, so the first response is the first screen.
//   node scripts/prerender.mjs   (after vite build and the --ssr build into .ssr)
import { readFileSync, readdirSync, writeFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { pathToFileURL } from 'node:url';

const ssr = resolve('.ssr', readdirSync('.ssr').find((f) => /^entry-server.*\.m?js$/.test(f)));
const { renderApp } = await import(pathToFileURL(ssr).href);
const file = 'dist/index.html';
let html = readFileSync(file, 'utf8');
if (!html.includes('<!--gf-app-->')) throw new Error('index.html has no <!--gf-app--> marker');
html = html.replace(/<link rel="stylesheet"[^>]*href="(\/assets\/[^"]+\.css)"[^>]*>/g, (_, href) => `<style>${readFileSync(`dist${href}`, 'utf8')}</style>`);
const { body, head } = renderApp();
html = html.replace('<!--gf-app-->', () => body).replace('</head>', () => `${head}</head>`);
writeFileSync(file, html);
console.log(`prerendered index.html: ${(html.length / 1024).toFixed(1)} KB`);
