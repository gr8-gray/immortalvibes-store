// web/src/routes/shop/+page.ts
//
// Loader for the mission-select planet row. Ships the full catalog merged
// with live prices/stock; if the API is down or slow to answer, the static
// catalog renders alone and the page still works.
import type { PageLoad } from './$types';
import { MOCK_PRODUCTS } from '$lib/types/shop';
import type { PageData } from '$lib/types/shop';
import { fetchLiveProducts, mergeLiveProduct } from '$lib/live-products';

export const load: PageLoad = async ({ fetch }): Promise<PageData> => {
  try {
    const apiProducts = await fetchLiveProducts(fetch);
    return { products: MOCK_PRODUCTS.map((mock) => mergeLiveProduct(mock, apiProducts)) };
  } catch {
    return { products: MOCK_PRODUCTS };
  }
};
