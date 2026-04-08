import { writable } from 'svelte/store';

export const isCartOpen = writable<boolean>(false);

export function openCart() {
  isCartOpen.set(true);
}

export function closeCart() {
  isCartOpen.set(false);
}
