// web/src/routes/shop/+page.ts
import type { PageLoad } from './$types';
import { MOCK_PRODUCTS } from '$lib/types/shop';
import type { PageData } from '$lib/types/shop';

export const load: PageLoad = async ({ fetch }): Promise<PageData> => {
  const apiBase = import.meta.env.VITE_API_BASE_URL ?? '';
  try {
    const res = await fetch(`${apiBase}/api/products`);
    if (!res.ok) throw new Error(`${res.status}`);
    const products = await res.json();
    return { products };
  } catch {
    return { products: MOCK_PRODUCTS };
  }
};
