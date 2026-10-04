import react from '@vitejs/plugin-react';
import { defineConfig, loadEnv } from 'vite';

export default defineConfig(({ mode }) => {
  const environment = loadEnv(mode, process.cwd(), '');
  return {
    plugins: [react()],
    // Preserve the demo's existing public build-time API URL contract.
    define: {
      'process.env.REACT_APP_API_URL': JSON.stringify(environment.REACT_APP_API_URL || '/api'),
    },
    build: {
      outDir: 'build',
    },
    server: {
      host: '0.0.0.0',
      port: 3000,
      strictPort: true,
      proxy: {
        '/api': {
          target: environment.API_PROXY_TARGET || 'http://localhost:8000',
          changeOrigin: true,
        },
      },
    },
  };
});
