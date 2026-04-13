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
  client_secret: string;
  order_id:      string;
  currency:      string;
  total_amount:  number;  // cents
}

export interface Order {
  id:               string;
  payment_intent_id: string;
  cart_token:       string;
  email:            string;
  currency:         string;
  total_amount:     number; // cents
  status:           string; // "pending" | "complete"
  created_at:       string;
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

// ─── Go Cart Types ────────────────────────────────────────

export interface GoLineItem {
  price_id:   string;
  product_id: string;
  name:       string;
  image_url:  string;
  currency:   string;
  amount:     number;  // cents
  quantity:   number;
}

export interface GoCart {
  token:      string;
  line_items: GoLineItem[];
}

export interface AddItemPayload {
  price_id:   string;
  product_id: string;
  name:       string;
  image_url:  string;
  currency:   string;
  amount:     number;
  quantity:   number;
}

// ─── Cart Endpoints ───────────────────────────────────────

/**
 * Add an item to the cart (or increment quantity if price_id already exists).
 * The Go API auto-creates the cart on first call and sets a cart_token cookie.
 * Returns the full cart including the cart token.
 */
export function addItemToCart(item: AddItemPayload): Promise<GoCart> {
  return apiFetch<GoCart>('/api/cart', {
    method: 'POST',
    body: JSON.stringify(item),
  });
}

/** Fetch a cart by token */
export function getCart(token: string): Promise<GoCart> {
  return apiFetch<GoCart>(`/api/cart/${token}`);
}

/**
 * Set the quantity of a specific line item. Pass quantity=0 to remove.
 * Requires the cart_token cookie to match the token in the URL.
 */
export function updateCartItem(token: string, price_id: string, quantity: number): Promise<GoCart> {
  return apiFetch<GoCart>(`/api/cart/${token}`, {
    method: 'PUT',
    body: JSON.stringify({ price_id, quantity }),
  });
}

// ─── Checkout Endpoints ───────────────────────────────────

export interface ShippingAddress {
  shipping_name: string;
  line1:         string;
  line2?:        string;
  city:          string;
  state:         string;
  postal_code:   string;
  country:       string;
}

/** Create a Stripe PaymentIntent for the current cart */
export function createCheckout(cartToken: string, email: string, address: ShippingAddress): Promise<CheckoutSession> {
  return apiFetch<CheckoutSession>('/api/checkout', {
    method: 'POST',
    body: JSON.stringify({ cart_token: cartToken, email, ...address }),
  });
}

// ─── Order Endpoints ──────────────────────────────────────

/** Fetch a completed order by ID */
export function getOrder(id: string): Promise<Order> {
  return apiFetch<Order>(`/api/order/${id}`);
}
