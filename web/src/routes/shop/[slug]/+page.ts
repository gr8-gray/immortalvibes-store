// web/src/routes/shop/[slug]/+page.ts
import type { PageLoad } from './$types';
import { MOCK_PRODUCTS } from '$lib/types/shop';
import type { Product } from '$lib/types/shop';

export interface PageData {
  product: Product;
}

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

export const load: PageLoad = async ({ fetch, params }): Promise<PageData> => {
  const mock = MOCK_PRODUCTS.find((p) => p.slug === params.slug);
  if (!mock) throw new Error(`Product not found: ${params.slug}`);

  const apiBase = import.meta.env.VITE_API_BASE_URL ?? '';
  try {
    const res = await fetch(`${apiBase}/api/products`);
    if (!res.ok) throw new Error(`${res.status}`);
    const apiProducts: ApiProduct[] = await res.json();

    const live = apiProducts.find((p) => p.name === mock.name);
    if (!live) return { product: mock };

    const usdPrice = live.prices.find((p) => p.currency === 'usd');
    return {
      product: {
        ...mock,
        id: live.id,
        price_id: usdPrice?.price_id ?? '',
        price_usd: usdPrice?.amount ?? mock.price_usd,
        status: live.stock_count > 0 ? 'available' : 'sold_out',
      },
    };
  } catch {
    return { product: mock };
  }
};
