import { writable, derived } from 'svelte/store';

export interface CartItem {
  variantId: string;
  productId: string;
  title: string;
  quantity: number;
  unitPrice: number; // in cents
  currency: string;
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
