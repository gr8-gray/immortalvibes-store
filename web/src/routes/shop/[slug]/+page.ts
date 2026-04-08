// web/src/routes/shop/[slug]/+page.ts
import type { PageLoad } from './$types';
import type { Product } from '../+page.js';

export interface PageData {
  product: Product;
}

export const load: PageLoad = async ({ fetch, params }): Promise<PageData> => {
  const apiBase = import.meta.env.VITE_API_BASE_URL ?? '';
  const res = await fetch(`${apiBase}/api/products/${params.slug}`);

  if (res.status === 404) {
    throw new Error(`Product not found: ${params.slug}`);
  }

  if (!res.ok) {
    throw new Error(`Failed to fetch product: ${res.status}`);
  }

  const product: Product = await res.json();
  return { product };
};
