import { browser } from '$app/environment';
import { PUBLIC_API_URL } from '$env/static/public';

/**
 * Returns the Go API base URL.
 * In dev: http://localhost:8080
 * In prod: value of PUBLIC_API_URL env var (the CF Worker URL)
 */
export function getApiBase(): string {
  if (!browser && typeof PUBLIC_API_URL === 'undefined') {
    // SSR build-time fallback — adapter-cloudflare will use the env var at runtime
    return 'http://localhost:8080';
  }
  return PUBLIC_API_URL ?? 'http://localhost:8080';
}
