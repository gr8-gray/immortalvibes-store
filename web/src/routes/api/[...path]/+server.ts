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

  try {
    const response = await fetch(targetUrl, {
      method: request.method,
      headers,
      body: request.method !== 'GET' && request.method !== 'HEAD'
        ? request.body
        : undefined,
    });

    const responseHeaders = new Headers(response.headers);
    responseHeaders.set('content-type', 'application/json');

    const body = await response.text();
    let json: string;
    try {
      JSON.parse(body);
      json = body;
    } catch {
      json = JSON.stringify({ error: body || 'upstream error', status: response.status });
    }

    return new Response(json, {
      status: response.status,
      headers: responseHeaders,
    });
  } catch (err) {
    return new Response(JSON.stringify({ error: 'API unreachable', detail: String(err) }), {
      status: 502,
      headers: { 'content-type': 'application/json' },
    });
  }
};

export const GET    = proxy;
export const POST   = proxy;
export const PUT    = proxy;
export const DELETE = proxy;
export const PATCH  = proxy;
