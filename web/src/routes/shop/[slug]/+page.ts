// web/src/routes/shop/[slug]/+page.ts
//
// Loader for a single product page. Resolves the slug against the catalog
// first — an unknown slug is a hard error — then layers live price/stock on
// top. `allProducts` ships the raw catalog because the mission nav only
// needs slugs and setIds, never live prices.
import type { PageLoad } from './$types';
import { MOCK_PRODUCTS } from '$lib/types/shop';
import type { Product } from '$lib/types/shop';
import { fetchLiveProducts, mergeLiveProduct } from '$lib/live-products';

export interface PageData {
  product: Product;
  allProducts: Product[];
}

export const load: PageLoad = async ({ fetch, params }): Promise<PageData> => {
  const mock = MOCK_PRODUCTS.find((p) => p.slug === params.slug);
  if (!mock) throw new Error(`Product not found: ${params.slug}`);

  try {
    const apiProducts = await fetchLiveProducts(fetch);
    return { product: mergeLiveProduct(mock, apiProducts), allProducts: MOCK_PRODUCTS };
  } catch {
    return { product: mock, allProducts: MOCK_PRODUCTS };
  }
};
