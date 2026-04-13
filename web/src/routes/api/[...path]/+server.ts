// web/src/routes/api/[...path]/+server.ts
// Proxy all /api/* requests to the Go API with the proxy secret injected.
// Compiled into _worker.js by the CF adapter — no external Worker needed.

import type { RequestHandler } from './$types';
import { env } from '$env/dynamic/private';

const GO_API = 'https://immortalvibes-api.fly.dev';

const proxy: RequestHandler = async ({ request, params, url }) => {
  const targetUrl = `${GO_API}/api/${params.path}${url.search}`;

  const headers = new Headers(request.headers);
  headers.set('X-Proxy-Secret', env.PROXY_SECRET);
  headers.delete('host');

  const response = await fetch(targetUrl, {
    method: request.method,
    headers,
    body: request.method !== 'GET' && request.method !== 'HEAD'
      ? request.body
      : undefined,
  });

  return new Response(response.body, {
    status: response.status,
    headers: response.headers,
  });
};

export const GET    = proxy;
export const POST   = proxy;
export const PUT    = proxy;
export const DELETE = proxy;
export const PATCH  = proxy;
