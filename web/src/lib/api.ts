import { getApiBase } from './env';

// ─── Domain Types ────────────────────────────────────────

export interface Price {
  amount: number;   // integer cents
  currency: string; // ISO 4217, e.g. "USD"
}

export interface ProductVariant {
  id: string;
  title: string;
  price: Price;
  available: boolean;
  sku: string;
}

export interface Product {
  id: string;
  title: string;
  description: string;
  handle: string;
  images: string[];       // array of image URLs
  variants: ProductVariant[];
  tags: string[];
  available: boolean;
  createdAt: string;      // ISO 8601
}

export interface LineItem {
  variantId: string;
  quantity: number;
}

export interface Cart {
  id: string;
  items: LineItem[];
  subtotal: Price;
  createdAt: string;
  updatedAt: string;
}

export interface CheckoutSession {
  id: string;
  cartId: string;
  currency: string;
  url: string;          // redirect URL for payment provider
  expiresAt: string;    // ISO 8601
}

export interface Order {
  id: string;
  status: 'pending' | 'paid' | 'fulfilled' | 'cancelled';
  items: LineItem[];
  total: Price;
  createdAt: string;
  updatedAt: string;
}

// ─── Error Type ──────────────────────────────────────────

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

// ─── Internal fetch helper ────────────────────────────────

async function apiFetch<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const base = getApiBase();
  const url = `${base}${path}`;

  const response = await fetch(url, {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers
    },
    ...options
  });

  if (!response.ok) {
    let code = 'UNKNOWN_ERROR';
    let message = `HTTP ${response.status}`;
    try {
      const body = await response.json();
      code = body.code ?? code;
      message = body.message ?? message;
    } catch {
      // response body is not JSON — use defaults
    }
    throw new ApiError(response.status, code, message);
  }

  return response.json() as Promise<T>;
}

// ─── Product Endpoints ────────────────────────────────────

/** Fetch all products */
export function getProducts(): Promise<Product[]> {
  return apiFetch<Product[]>('/api/v1/products');
}

/** Fetch a single product by ID */
export function getProduct(id: string): Promise<Product> {
  return apiFetch<Product>(`/api/v1/products/${id}`);
}

// ─── Cart Endpoints ───────────────────────────────────────

/** Create a new empty cart */
export function createCart(): Promise<Cart> {
  return apiFetch<Cart>('/api/v1/carts', { method: 'POST' });
}

/** Fetch a cart by ID */
export function getCart(id: string): Promise<Cart> {
  return apiFetch<Cart>(`/api/v1/carts/${id}`);
}

/**
 * Replace the line items in a cart.
 * Pass the full desired item list — server replaces, not merges.
 */
export function updateCart(id: string, items: LineItem[]): Promise<Cart> {
  return apiFetch<Cart>(`/api/v1/carts/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ items })
  });
}

// ─── Checkout Endpoints ───────────────────────────────────

/** Create a checkout session for the given cart */
export function createCheckout(cartId: string, currency: string): Promise<CheckoutSession> {
  return apiFetch<CheckoutSession>('/api/v1/checkout', {
    method: 'POST',
    body: JSON.stringify({ cartId, currency })
  });
}

// ─── Order Endpoints ──────────────────────────────────────

/** Fetch a completed order by ID */
export function getOrder(id: string): Promise<Order> {
  return apiFetch<Order>(`/api/v1/orders/${id}`);
}
