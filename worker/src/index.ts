export interface Env {
  GO_API_URL: string;
  PROXY_SECRET: string;
  RATE_LIMIT_KV: KVNamespace;
  CARTS: KVNamespace;
  SESSIONS: KVNamespace;
}

const RATE_LIMIT_WINDOW_MS = 60_000; // 1 minute
const RATE_LIMIT_MAX = 60;           // 60 requests per minute per IP

async function isRateLimited(kv: KVNamespace, ip: string): Promise<boolean> {
  const key = `ratelimit:${ip}`;
  const raw = await kv.get(key);
  const count = raw ? parseInt(raw, 10) : 0;

  if (count >= RATE_LIMIT_MAX) return true;

  // Increment — TTL resets the window
  await kv.put(key, String(count + 1), { expirationTtl: 60 });
  return false;
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);

    // Health check on the worker itself — no forwarding needed
    if (url.pathname === "/worker-health") {
      return new Response(JSON.stringify({ status: "ok" }), {
        headers: { "Content-Type": "application/json" },
      });
    }

    // Rate limiting — skip for health checks
    if (url.pathname !== "/health" && env.RATE_LIMIT_KV) {
      const ip = request.headers.get("CF-Connecting-IP") ?? "unknown";
      const limited = await isRateLimited(env.RATE_LIMIT_KV, ip);
      if (limited) {
        return new Response("Too Many Requests", { status: 429 });
      }
    }

    // Forward to Go API with proxy secret injected
    const targetUrl = `${env.GO_API_URL}${url.pathname}${url.search}`;
    const headers = new Headers(request.headers);
    headers.set("X-Proxy-Secret", env.PROXY_SECRET);

    const response = await fetch(targetUrl, {
      method: request.method,
      headers,
      body: request.method !== "GET" && request.method !== "HEAD"
        ? request.body
        : undefined,
    });

    return response;
  },
};
