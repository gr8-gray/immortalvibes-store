// web/src/routes/shop/[slug]/+page.ts
import type { PageLoad } from './$types';
import type { Product } from '../+page.js';

export interface PageData {
  product: Product;
}

import { MOCK_PRODUCTS } from '../+page.js';

export const load: PageLoad = async ({ fetch, params }): Promise<PageData> => {
  const apiBase = import.meta.env.VITE_API_BASE_URL ?? '';
  try {
    const res = await fetch(`${apiBase}/api/products/${params.slug}`);
    if (res.status === 404) throw new Error('not_found');
    if (!res.ok) throw new Error(`${res.status}`);
    const product: Product = await res.json();
    return { product };
  } catch (err) {
    const mock = MOCK_PRODUCTS.find((p) => p.slug === params.slug);
    if (!mock) throw new Error(`Product not found: ${params.slug}`);
    return { product: mock };
  }
};
