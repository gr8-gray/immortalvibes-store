// web/src/routes/shop/+page.ts
import type { PageLoad } from './$types';

export interface Product {
  id: string;
  slug: string;
  name: string;
  description: string;
  price_usd: number;    // cents
  price_gbp: number;    // cents
  currency: string;     // 'usd' | 'gbp' — determined by Go API from CF geo header
  status: 'available' | 'sold_out' | 'coming_soon';
  sizes: string[];
  image_url: string;
  mission_number: '001' | '002' | '003' | '004';
}

export interface PageData {
  products: Product[];
}

export const load: PageLoad = async ({ fetch }): Promise<PageData> => {
  const apiBase = import.meta.env.VITE_API_BASE_URL ?? '';
  const res = await fetch(`${apiBase}/api/products`);

  if (!res.ok) {
    throw new Error(`Failed to fetch products: ${res.status}`);
  }

  const products: Product[] = await res.json();
  return { products };
};
