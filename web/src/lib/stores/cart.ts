import { writable, derived, get } from 'svelte/store';
import { updateCartItem, type GoCart } from '$lib/api';

export interface CartItem {
  variantId: string;
  productId: string;
  title: string;
  quantity: number;
  unitPrice: number; // in cents
  currency: string;
}

function mapGoCart(goCart: GoCart): { id: string; items: CartItem[] } {
  return {
    id: goCart.token,
    items: goCart.line_items.map((li) => ({
      variantId: li.price_id,
      productId: li.product_id,
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

      try {
        const goCart = await updateCartItem(token, variantId, clamped);
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
