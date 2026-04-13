/**
 * Returns the Go API base URL.
 * Empty string = relative URLs routed through the SvelteKit /api/[...path] proxy.
 */
export function getApiBase(): string {
  return '';
}
