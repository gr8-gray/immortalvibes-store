// Shared shop types and catalog data — imported by +page.ts and components.
//
// This file is the storefront's single source of truth for WHAT is sold:
// slugs, names, sizes, baseline prices, and per-colorway imagery. The API
// layered on top (see +page.ts merge) is the source of truth for LIVE state:
// Stripe IDs, current prices, and stock counts. Everything else in web/
// should derive product facts from here rather than restating them.

// Public R2 bucket that serves the heavier model/gallery photography.
// Product cut-out PNGs ship from /static instead so they survive R2 outages.
const R2 = 'https://pub-75a66fca0ddd4d93b3bb53bda5d6a29c.r2.dev';

// ─── Canonical slugs ───────────────────────────────────────────────────────
// Every route, accent map, and mission-order list keys off these strings.
// They were once retyped by hand in five different files; now a slug rename
// is a one-line change here. Keys are short internal names, values are the
// public URL segments under /shop/.
export const SLUG = {
  beanie: 'warped-reality-beanie',
  truckerHat: 'vanguard-trucker-hat',
  tank: 'racerback-tanktop',
  tee: 'warped-reality-tee',
  sweatpants: 'warped-reality-sweatpants',
  lightSweatpants: 'immortal-light-sweatpants',
  set: 'warped-reality-set',
} as const;

// The Warped Reality drop groups three products into one "set" landing page.
// The set itself is not a product — it has its own route and planet on /shop.
export const WARPED_REALITY_SET_ID = 'warped-reality';
export const WR_SET = {
  slug: SLUG.set,
  name: 'Warped Reality',
  missionNumber: '001',
} as const;

// ─── Mission environments ──────────────────────────────────────────────────
// Each mission number is themed as a place in space. The product page hero
// and the /shop planet row both read from this map.
export const MISSION_LABELS: Record<string, string> = {
  '001': 'Low Earth Orbit',
  '002': 'Lunar Surface',
  '003': 'Stellar Nursery',
  '004': 'Deep Space',
};

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

/** Per-variant (size/colorway) stock data from the API. */
export interface StockVariant {
  variant: string;     // matches a size label or ProductVariant.colorName exactly
  stock_count: number;
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
  stockVariants?: StockVariant[]; // per-variant stock from API; empty/absent = all sizes enabled
  mission_number: '001' | '002' | '003' | '004';
  setId?: typeof WARPED_REALITY_SET_ID; // groups products into a drop set
}

export interface PageData {
  products: Product[];
}

export const MOCK_PRODUCTS: Product[] = [
  {
    id: 'mock_001',
    slug: SLUG.beanie,
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
    setId: WARPED_REALITY_SET_ID,
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
    slug: SLUG.truckerHat,
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
    slug: SLUG.tank,
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
    slug: SLUG.tee,
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
    setId: WARPED_REALITY_SET_ID,
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
    slug: SLUG.sweatpants,
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
    setId: WARPED_REALITY_SET_ID,
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
  {
    id: 'mock_006',
    slug: SLUG.lightSweatpants,
    name: 'Immortal Light Sweatpants',
    description: 'Forged in deep space. All-over rocky serpentine-scale terrain print with a gold IMMORTAL wordmark running the left leg. Drawstring waist, unisex cut.',
    price_usd: 3000,
    price_gbp: 2400,
    price_id: '',
    currency: 'usd',
    status: 'available',
    sizes: ['S', 'M', 'L', 'XL', '2XL', '3XL'],
    image_url: `${R2}/sweatpants/model-1-hero.jpg`,
    mission_number: '004',
    variants: [
      {
        colorName: 'Mars Rust',
        hex: '#6B4A38',
        productImage: '/photos/_drop/immortal-light-sweatpants-front.png',
        gallery: [
          `${R2}/sweatpants/model-1-hero.jpg`,
          `${R2}/sweatpants/model-porch.jpg`,
          `${R2}/sweatpants/model-willow.jpg`,
          `${R2}/sweatpants/model-lake.jpg`,
        ],
      },
    ],
  },
];

/**
 * Look up a catalog entry by its public slug. Pages that need one product's
 * facts (name, mission number, sizes) should reach for this instead of
 * restating those facts in place — that restating is exactly how the /shop
 * planet row and the set landing page drifted from the catalog in the past.
 */
export function productBySlug(slug: string): Product | undefined {
  return MOCK_PRODUCTS.find((p) => p.slug === slug);
}
