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

export const MOCK_PRODUCTS: Product[] = [
  {
    id: 'mock_001',
    slug: 'warped-reality-beanie',
    name: 'Warped Reality Beanie',
    description: 'Knit for those who drift between dimensions. One size, infinite orbits.',
    price_usd: 3500,
    price_gbp: 2800,
    currency: 'usd',
    status: 'available',
    sizes: ['OS'],
    image_url: '/photos/blue-beanie.jpeg',
    mission_number: '001',
  },
  {
    id: 'mock_002',
    slug: 'vanguard-trucker-hat',
    name: 'Vanguard Trucker Hat',
    description: 'Engineered for the lunar surface. Structured front, breathable mesh, mission-ready.',
    price_usd: 3800,
    price_gbp: 3000,
    currency: 'usd',
    status: 'available',
    sizes: ['OS'],
    image_url: '',
    mission_number: '002',
  },
  {
    id: 'mock_003',
    slug: 'racerback-tanktop',
    name: 'Racerback Tanktop',
    description: 'Born in the stellar nursery. Lightweight, orbital-grade, built to move.',
    price_usd: 3200,
    price_gbp: 2500,
    currency: 'usd',
    status: 'available',
    sizes: ['XS', 'S', 'M', 'L', 'XL'],
    image_url: '/photos/tank-front.png',
    mission_number: '003',
  },
];

export const load: PageLoad = async ({ fetch }): Promise<PageData> => {
  const apiBase = import.meta.env.VITE_API_BASE_URL ?? '';
  try {
    const res = await fetch(`${apiBase}/api/products`);
    if (!res.ok) throw new Error(`${res.status}`);
    const products: Product[] = await res.json();
    return { products };
  } catch {
    return { products: MOCK_PRODUCTS };
  }
};
