// Shared shop types and mock data — imported by +page.ts and components

const R2 = 'https://pub-75a66fca0ddd4d93b3bb53bda5d6a29c.r2.dev';

export interface ProductVariant {
  colorName: string;    // e.g. "Blue", "Green"
  hex: string;          // swatch color
  productImage: string; // standalone transparent PNG
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
  price_usd: number;    // cents
  price_gbp: number;    // cents
  currency: string;     // 'usd' | 'gbp'
  status: 'available' | 'sold_out' | 'coming_soon';
  sizes: string[];
  image_url: string;          // legacy fallback
  variants?: ProductVariant[]; // color variants with standalone + model shots
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
    image_url: `${R2}/beanie/model-waterfront-blue-portrait.jpg`,
    mission_number: '001',
    variants: [
      {
        colorName: 'Blue',
        hex: '#3A6EA5',
        productImage: '/photos/product-beanie.png',
        gallery: [
          `${R2}/beanie/model-waterfront-blue-portrait.jpg`,
          `${R2}/beanie/model-closeup-blue-pull.jpg`,
          `${R2}/beanie/model-sky-blue-lowangle.jpg`,
          `${R2}/beanie/model-waterfront-blue-back-pull.jpg`,
          `${R2}/beanie/owner-waterfront-blue-seated-laugh.jpg`,
          `${R2}/beanie/owner-waterfront-blue-seated-smile.jpg`,
        ],
      },
      {
        colorName: 'Green',
        hex: '#3D6B50',
        productImage: '/photos/product-green-beanie.png',
        gallery: [
          `${R2}/beanie/model-waterfront-green-crossed.jpg`,
          `${R2}/beanie/owner-night-green-audi.jpg`,
          `${R2}/beanie/owner-night-green-truck.jpg`,
        ],
      },
    ],
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
          `${R2}/hat/model-dramatic-trucker-tank-lighter.jpg`,
        ],
      },
    ],
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
];
