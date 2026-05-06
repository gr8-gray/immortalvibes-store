// Shared shop types and mock data — imported by +page.ts and components

const R2 = 'https://pub-75a66fca0ddd4d93b3bb53bda5d6a29c.r2.dev';

export interface ProductVariant {
  colorName: string;    // e.g. "Blue", "Green"
  hex: string;          // swatch color
  productImage: string; // standalone transparent PNG (front view)
  backImage?: string;   // optional back view — pairs with productImage for auto-cycle cards
  gallery: string[];    // model shots for this colorway
  imageScale?: number;   // CSS scale for standalone shot (default 1.0)
  planetScale?: number;   // override productScale on the planet select screen
  planetOffsetY?: number; // vertical shift on planet (Three.js units, + = up)
  spriteBlending?: 'normal' | 'additive'; // additive: black→transparent, white glows
}

export interface Product {
  id: string;
  slug: string;
  name: string;
  description: string;
  price_usd: number;    // cents — 0 = TBD
  price_gbp: number;    // cents — 0 = TBD
  price_id: string;     // Stripe Price ID (USD) — populated from API
  currency: string;     // 'usd' | 'gbp'
  status: 'available' | 'sold_out' | 'coming_soon';
  sizes: string[];
  image_url: string;          // legacy fallback
  variants?: ProductVariant[]; // color variants with standalone + model shots
  mission_number: '001' | '002' | '003' | '004';
  setId?: 'warped-reality'; // groups products into a drop set
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
    price_usd: 2000,
    price_gbp: 1600,
    price_id: '',
    currency: 'usd',
    status: 'available',
    sizes: ['OS'],
    image_url: '/photos/_drop/beanie-blue.png',
    mission_number: '001',
    setId: 'warped-reality',
    variants: [
      {
        colorName: 'Earth Blue',
        hex: '#1E8FB8',
        productImage: '/photos/_drop/beanie-blue.png',
        gallery: ['/photos/blue-beanie-model.jpeg'],
      },
      {
        colorName: 'Green',
        hex: '#3D6B50',
        productImage: '/photos/product-green-beanie.png',
        gallery: ['/photos/green-beanie-model.jpeg'],
      },
    ],
  },
  {
    id: 'mock_002',
    slug: 'vanguard-trucker-hat',
    name: 'Vanguard Trucker Hat',
    description: 'Engineered for the lunar surface. Structured front, breathable mesh, mission-ready.',
    price_usd: 2000,
    price_gbp: 1600,
    price_id: '',
    currency: 'usd',
    status: 'available',
    sizes: ['OS'],
    image_url: `${R2}/hat/model-dramatic-trucker-tank-lighter.jpg`,
    mission_number: '002',
    variants: [
      {
        colorName: 'Olive',
        hex: '#6B6B4A',
        productImage: '/photos/product-hat.png',
        imageScale: 1.8,
        planetScale: 0.9,
        planetOffsetY: 0.18,
        gallery: [
          `${R2}/hat/model-restaurant-olive.jpg`,
          `${R2}/hat/owner-night-audi-standing.jpg`,
        ],
      },
    ],
  },
  {
    id: 'mock_003',
    slug: 'racerback-tanktop',
    name: 'Racerback Tanktop',
    description: 'Born in the stellar nursery. Lightweight, orbital-grade, built to move.',
    price_usd: 2200,
    price_gbp: 1750,
    price_id: '',
    currency: 'usd',
    status: 'available',
    sizes: ['XS', 'S', 'M', 'L', 'XL'],
    image_url: `${R2}/tanktop/owner-gym-back.jpg`,
    mission_number: '003',
    variants: [
      {
        colorName: 'Black',
        hex: '#1A1A1A',
        productImage: '/photos/product-tank.png',
        spriteBlending: 'additive',
        gallery: [
          `${R2}/tanktop/owner-gym-back.jpg`,
        ],
      },
    ],
  },
  {
    id: 'mock_004',
    slug: 'warped-reality-tee',
    name: 'Warped Reality Tee',
    description: 'Heavyweight oversized cut. Sky-blue collar print on the front, ornate spires graphic on the back. Black only.',
    price_usd: 3000,
    price_gbp: 2400,
    price_id: 'price_1TTl9AHy1AKk8SUWeLENS5Bu',
    currency: 'usd',
    status: 'available',
    sizes: ['S', 'M', 'L', 'XL', '2XL', '3XL'],
    image_url: '/photos/_drop/tee-front.png',
    mission_number: '001',
    setId: 'warped-reality',
    variants: [
      {
        colorName: 'Black',
        hex: '#0E0E12',
        productImage: '/photos/_drop/tee-front.png',
        backImage: '/photos/_drop/tee-back.png',
        spriteBlending: 'additive',
        gallery: [],
      },
    ],
  },
  {
    id: 'mock_005',
    slug: 'warped-reality-sweatpants',
    name: 'Warped Reality Sweatpants',
    description: 'Wide-leg, midweight fleece. Arched IMMORTAL VIBES graphic at the waist. Black only.',
    price_usd: 3000,
    price_gbp: 2400,
    price_id: 'price_1TTl9YHy1AKk8SUWGPXI3icl',
    currency: 'usd',
    status: 'available',
    sizes: ['S', 'M', 'L', 'XL', '2XL', '3XL'],
    image_url: '/photos/_drop/sweatpants-front.png',
    mission_number: '001',
    setId: 'warped-reality',
    variants: [
      {
        colorName: 'Black',
        hex: '#0E0E12',
        productImage: '/photos/_drop/sweatpants-front.png',
        spriteBlending: 'additive',
        gallery: [],
      },
    ],
  },
];
