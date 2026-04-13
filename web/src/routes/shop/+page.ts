// web/src/routes/shop/+page.ts
import type { PageLoad } from './$types';
import { MOCK_PRODUCTS } from '$lib/types/shop';
import type { PageData } from '$lib/types/shop';

interface ApiPrice {
  price_id: string;
  currency: string;
  amount: number;
}

interface ApiProduct {
  id: string;
  name: string;
  stock_count: number;
  prices: ApiPrice[];
}

export const load: PageLoad = async ({ fetch }): Promise<PageData> => {
  const apiBase = import.meta.env.VITE_API_BASE_URL ?? '';
  try {
    const res = await fetch(`${apiBase}/api/products`);
    if (!res.ok) throw new Error(`${res.status}`);
    const apiProducts: ApiProduct[] = await res.json();

    // Merge live data (Stripe ID, price_id, stock) into static config (variants, gallery, scene).
    // Static config is the source of truth for visuals; API is the source of truth for prices/stock.
    const merged = MOCK_PRODUCTS.map((mock) => {
      const live = apiProducts.find((p) => p.name === mock.name);
      if (!live) return mock;
      const usdPrice = live.prices.find((p) => p.currency === 'usd');
      return {
        ...mock,
        id: live.id,
        price_id: usdPrice?.price_id ?? '',
        price_usd: usdPrice?.amount ?? mock.price_usd,
        status: live.stock_count > 0 ? ('available' as const) : ('sold_out' as const),
      };
    });

    return { products: merged };
  } catch {
    return { products: MOCK_PRODUCTS };
  }
};
