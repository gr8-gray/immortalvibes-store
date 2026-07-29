// web/src/lib/live-products.ts
//
// The bridge between the static catalog ($lib/types/shop) and the live Go
// API. Both shop loaders — the /shop planet row and the /shop/[slug] product
// page — used to carry their own private copy of these wire types and of the
// merge below. Two copies of a merge is how one page ends up selling at a
// stale price, so the whole dance now lives here, once.
//
// Division of truth, as ever: the catalog owns visuals and identity (slugs,
// variants, galleries, scenes); the API owns money and stock (Stripe IDs,
// current prices, stock counts). The merge stitches the two along `name`.

import type { Product, StockVariant } from '$lib/types/shop';

// ─── Wire types for GET /api/products ──────────────────────────────────────
export interface ApiPrice {
  price_id: string;
  currency: string;
  amount: number;
}

export interface ApiStockVariant {
  variant: string;
  stock_count: number;
}

export interface ApiProduct {
  id: string;
  name: string;
  stock_count: number;
  prices: ApiPrice[];
  variants: ApiStockVariant[];
}

/**
 * Fetch the live product list. Callers pass SvelteKit's load `fetch` so the
 * request works identically during SSR and in the browser. Throws on any
 * non-2xx status — loaders catch and fall back to the static catalog, which
 * is what keeps the storefront browsable when the API is unreachable.
 */
export async function fetchLiveProducts(fetchFn: typeof fetch): Promise<ApiProduct[]> {
  const apiBase = import.meta.env.VITE_API_BASE_URL ?? '';
  const res = await fetchFn(`${apiBase}/api/products`);
  if (!res.ok) throw new Error(`${res.status}`);
  return res.json();
}

/**
 * Merge one live API row into one catalog entry. If the API doesn't know the
 * product (matched by name), the catalog entry passes through untouched —
 * it will render with its baseline price and an empty Stripe price_id.
 */
export function mergeLiveProduct(mock: Product, apiProducts: ApiProduct[]): Product {
  const live = apiProducts.find((p) => p.name === mock.name);
  if (!live) return mock;

  const usdPrice = live.prices.find((p) => p.currency === 'usd');
  const stockVariants: StockVariant[] = (live.variants ?? []).map((v) => ({
    variant: v.variant,
    stock_count: v.stock_count,
  }));

  return {
    ...mock,
    id: live.id,
    price_id: usdPrice?.price_id ?? '',
    price_usd: usdPrice?.amount ?? mock.price_usd,
    status: live.stock_count > 0 ? ('available' as const) : ('sold_out' as const),
    stockVariants,
  };
}
