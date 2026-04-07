import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

/** HTML 导航请求回退到 index.html，避免 BrowserRouter 下刷新 /messages 等路径 404 */
function spaFallback() {
  const handler = (req, res, next) => {
    if (req.method !== 'GET' && req.method !== 'HEAD') return next();
    const accept = req.headers.accept || '';
    if (!accept.includes('text/html')) return next();
    const raw = req.url || '';
    const pathOnly = raw.split('?')[0] || '';
    if (pathOnly.startsWith('/@') || pathOnly.startsWith('/src') || pathOnly.startsWith('/node_modules')) {
      return next();
    }
    if (/\.[a-zA-Z0-9]+$/.test(pathOnly) && !pathOnly.endsWith('.html')) {
      return next();
    }
    if (pathOnly !== '/' && pathOnly !== '/index.html') {
      req.url = '/index.html';
    }
    next();
  };
  return {
    name: 'spa-html-fallback',
    configureServer(server) {
      server.middlewares.use(handler);
    },
    configurePreviewServer(server) {
      server.middlewares.use(handler);
    },
  };
}

export default defineConfig({
  plugins: [react(), spaFallback()],
  appType: 'spa',
  server: { host: true, port: 5173 },
  preview: { host: true, port: 4173 },
});
