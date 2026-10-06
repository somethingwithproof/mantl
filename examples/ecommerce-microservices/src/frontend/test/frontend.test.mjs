// SPDX-License-Identifier: Apache-2.0
import assert from 'node:assert/strict';
import { once } from 'node:events';
import { createServer as createHttpServer } from 'node:http';
import test from 'node:test';
import { createServer } from 'vite';
import configure from '../vite.config.mjs';

test('preserves the public API URL and production output directory', () => {
  const previous = process.env.REACT_APP_API_URL;
  try {
    delete process.env.REACT_APP_API_URL;
    const defaults = configure({ mode: 'production' });
    assert.equal(defaults.define['process.env.REACT_APP_API_URL'], '"/api"');
    assert.equal(defaults.build.outDir, 'build');
    process.env.REACT_APP_API_URL = 'https://api.example.test/api';
    const customized = configure({ mode: 'production' });
    assert.equal(customized.define['process.env.REACT_APP_API_URL'], '"https://api.example.test/api"');
  } finally {
    if (previous === undefined) delete process.env.REACT_APP_API_URL;
    else process.env.REACT_APP_API_URL = previous;
  }
});

test('serves the React entry point and proxies API requests', { timeout: 15000 }, async (t) => {
  const backend = createHttpServer((request, response) => {
    response.setHeader('content-type', 'application/json');
    response.end(JSON.stringify({ path: request.url }));
  });
  backend.listen(0, '127.0.0.1');
  await once(backend, 'listening');
  t.after(() => new Promise((resolve) => backend.close(resolve)));

  const previous = process.env.API_PROXY_TARGET;
  process.env.API_PROXY_TARGET = `http://127.0.0.1:${backend.address().port}`;
  t.after(() => {
    if (previous === undefined) delete process.env.API_PROXY_TARGET;
    else process.env.API_PROXY_TARGET = previous;
  });
  const frontend = await createServer({
    server: { host: '127.0.0.1', port: 0 },
    logLevel: 'silent',
  });
  t.after(() => frontend.close());
  await frontend.listen();
  const base = `http://127.0.0.1:${frontend.httpServer.address().port}`;

  const index = await fetch(base);
  assert.equal(index.status, 200);
  assert.match(await index.text(), /src\/index\.tsx/);
  const entry = await fetch(`${base}/src/index.tsx`);
  assert.equal(entry.status, 200);
  assert.match(await entry.text(), /App/);
  const api = await fetch(`${base}/api/products?limit=2`);
  assert.equal(api.status, 200);
  assert.deepEqual(await api.json(), { path: '/api/products?limit=2' });
});
