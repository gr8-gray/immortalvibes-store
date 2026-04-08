export interface Env {
  GO_API_URL: string;
  PROXY_SECRET: string;
  RATE_LIMIT_KV: KVNamespace;
  CARTS: KVNamespace;
  SESSIONS: KVNamespace;
}

const RATE_LIMIT_MAX = 30;          // 30 requests per minute per IP (tightened)
const DAILY_GLOBAL_CAP = 80_000;    // hard stop at 80k/day — CF free tier is 100k

async function isRateLimited(kv: KVNamespace, ip: string): Promise<boolean> {
  const key = `ratelimit:${ip}`;
  const raw = await kv.get(key);
  const count = raw ? parseInt(raw, 10) : 0;

  if (count >= RATE_LIMIT_MAX) return true;

  await kv.put(key, String(count + 1), { expirationTtl: 60 });
  return false;
}

async function isGlobalCapExceeded(kv: KVNamespace): Promise<boolean> {
  const today = new Date().toISOString().slice(0, 10); // "2026-04-07"
  const key = `global:${today}`;
  const raw = await kv.get(key);
  const count = raw ? parseInt(raw, 10) : 0;

  if (count >= DAILY_GLOBAL_CAP) return true;

  // TTL of 86400s ensures the key auto-expires after 24h
  await kv.put(key, String(count + 1), { expirationTtl: 86400 });
  return false;
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);

    // Worker self-health — no forwarding, no rate limit
    if (url.pathname === "/worker-health") {
      return new Response(JSON.stringify({ status: "ok" }), {
        headers: { "Content-Type": "application/json" },
      });
    }

    if (url.pathname !== "/health" && env.RATE_LIMIT_KV) {
      // Global daily cap — checked first (cheapest rejection)
      if (await isGlobalCapExceeded(env.RATE_LIMIT_KV)) {
        return new Response("Service temporarily unavailable", { status: 503 });
      }

      // Per-IP rate limit
      const ip = request.headers.get("CF-Connecting-IP") ?? "unknown";
      if (await isRateLimited(env.RATE_LIMIT_KV, ip)) {
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
