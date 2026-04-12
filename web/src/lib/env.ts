/**
 * Returns the Go API base URL.
 * In dev: http://localhost:8080
 * In prod: PUBLIC_API_URL env var (CF Worker URL) set in CF Pages environment variables
 */
export function getApiBase(): string {
  // Use dynamic env so the build doesn't fail if the var isn't set at build time
  try {
    // @ts-ignore — dynamic import resolves at runtime on CF Pages
    const url = import.meta.env.PUBLIC_API_URL;
    if (url) return url;
  } catch { /* noop */ }
  return 'http://localhost:8080';
}
