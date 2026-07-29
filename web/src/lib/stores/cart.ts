// web/src/lib/stores/cart.ts
//
// The client-side view of the cart. The Go API (via CF KV) is the real
// keeper of cart state — this store mirrors it so the drawer and checkout
// can render synchronously, and falls back to local-only math whenever the
// API can't be reached mid-session.
import { writable, derived, get } from 'svelte/store';
import { updateCartItem, type GoCart } from '$lib/api';

export interface CartItem {
  variantId: string;  // dedup key: `${price_id}:${size}` (size may be empty for OS)
  productId: string;
  priceId:   string;  // Stripe price ID
  size:      string;  // variant label or '' for one-size products
  title: string;
  quantity: number;
  unitPrice: number; // in cents
  currency: string;
}

/**
 * The one place the cart line-item dedup key is built. Same product+size on
 * two add-to-cart clicks must collapse into one line, so everything that
 * constructs a variantId — this store and the product page's add-to-cart
 * mapping — has to agree on the shape. They agree by calling this.
 */
export function makeVariantId(price_id: string, size: string): string {
  return size ? `${price_id}:${size}` : price_id;
}

function mapGoCart(goCart: GoCart): { id: string; items: CartItem[] } {
  return {
    id: goCart.token,
    items: goCart.line_items.map((li) => ({
      variantId: makeVariantId(li.price_id, li.size ?? ''),
      productId: li.product_id,
      priceId:   li.price_id,
      size:      li.size ?? '',
      title:     li.name,
      quantity:  li.quantity,
      unitPrice: li.amount,
      currency:  li.currency,
    })),
  };
}

export interface CartState {
  id: string | null;
  items: CartItem[];
}

const initialState: CartState = {
  id: null,
  items: []
};

function createCartStore() {
  const { subscribe, set, update } = writable<CartState>(initialState);

  return {
    subscribe,
    setCart(id: string, items: CartItem[]) {
      set({ id, items });
    },
    /**
     * Replace local state with a server cart response wholesale. Every
     * surface that receives a GoCart back from the API (layout hydration,
     * add-to-cart, quantity updates) funnels through the same mapGoCart
     * translation, so line items always carry identical dedup keys.
     */
    setFromGoCart(goCart: GoCart) {
      set(mapGoCart(goCart));
    },
    addItem(item: CartItem) {
      update((state) => {
        const existing = state.items.find((i) => i.variantId === item.variantId);
        if (existing) {
          return {
            ...state,
            items: state.items.map((i) =>
              i.variantId === item.variantId
                ? { ...i, quantity: i.quantity + item.quantity }
                : i
            )
          };
        }
        return { ...state, items: [...state.items, item] };
      });
    },
    removeItem(variantId: string) {
      update((state) => ({
        ...state,
        items: state.items.filter((i) => i.variantId !== variantId)
      }));
    },
    /**
     * Set the quantity of a specific line item, syncing to the server cart.
     * Passing quantity <= 0 removes the line. Falls back to a local-only
     * update if no server token is present or the API call fails so the UI
     * stays responsive.
     */
    async setItemQuantity(variantId: string, quantity: number) {
      const state = get({ subscribe });
      const token = state.id;
      const clamped = Math.max(0, Math.floor(quantity));

      const localFallback = () => {
        update((s) =>
          clamped === 0
            ? { ...s, items: s.items.filter((i) => i.variantId !== variantId) }
            : {
                ...s,
                items: s.items.map((i) =>
                  i.variantId === variantId ? { ...i, quantity: clamped } : i
                ),
              }
        );
      };

      if (!token) {
        localFallback();
        return;
      }

      // Find the item to get its priceId and size for the server call.
      const item = state.items.find((i) => i.variantId === variantId);
      if (!item) {
        localFallback();
        return;
      }

      try {
        const goCart = await updateCartItem(token, item.priceId, item.size, clamped);
        set(mapGoCart(goCart));
      } catch {
        localFallback();
      }
    },
    clear() {
      set(initialState);
    }
  };
}

export const cart = createCartStore();

/** Derived: total item count for badge display */
export const cartCount = derived(cart, ($cart) =>
  $cart.items.reduce((sum, item) => sum + item.quantity, 0)
);
