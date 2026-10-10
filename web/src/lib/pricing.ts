// web/src/lib/pricing.ts
//
// Client-side mirror of the API's discount arithmetic (api/handlers/discount.go,
// Discount.Amount). The server is authoritative and the checkout page shows
// the server's total once the payment intent exists; this keeps the estimate
// shown before that point identical to what will be charged.
import type { PromoDiscount } from '$lib/api';

/** Smallest amount a discount may leave on an order (Stripe USD minimum). */
export const MIN_CHARGE_CENTS = 50;

export interface PricedLine {
  productId: string;
  unitPrice: number; // cents
  quantity: number;
}

export type DiscountResult =
  | { ok: true; cents: number }
  | { ok: false; reason: 'minimum' | 'not_applicable' };

/**
 * Discount in cents for the given lines and shipping. Percentage discounts
 * apply to items plus shipping unless restricted to specific products, in
 * which case they apply to those items only.
 */
export function discountCents(
  lines: PricedLine[],
  shippingCents: number,
  discount: PromoDiscount
): DiscountResult {
  const restricted = (discount.product_ids?.length ?? 0) > 0;
  let subtotal = 0;
  let eligible = 0;
  for (const l of lines) {
    const line = l.unitPrice * l.quantity;
    subtotal += line;
    if (!restricted || discount.product_ids!.includes(l.productId)) eligible += line;
  }
  if (!restricted) eligible += shippingCents;

  if (discount.minimum_amount && subtotal < discount.minimum_amount) {
    return { ok: false, reason: 'minimum' };
  }
  if (eligible <= 0) return { ok: false, reason: 'not_applicable' };

  const off =
    discount.type === 'percent_off'
      ? Math.floor((eligible * discount.value) / 100)
      : Math.min(discount.value, eligible);
  const total = subtotal + shippingCents;
  return { ok: true, cents: Math.max(0, Math.min(off, total - MIN_CHARGE_CENTS)) };
}
