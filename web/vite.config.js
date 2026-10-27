import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// https://vitejs.dev/config/
// sitemap.xml is generated at build time so lastmod is the build date (the faucet is one indexable page)
const sitemap = () => ({
  name: 'gully-sitemap',
  apply: 'build',
  generateBundle() {
    const lastmod = new Date().toISOString().slice(0, 10)
    this.emitFile({
      type: 'asset',
      fileName: 'sitemap.xml',
      source: `<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n  <url><loc>https://faucet.gullychain.xyz/</loc><lastmod>${lastmod}</lastmod><changefreq>daily</changefreq><priority>1.0</priority></url>\n</urlset>\n`,
    })
  },
})

export default defineConfig({
  plugins: [svelte(), sitemap()],
})
