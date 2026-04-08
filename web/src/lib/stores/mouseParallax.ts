// web/src/lib/stores/mouseParallax.ts
import { writable } from 'svelte/store';

export const mouseParallax = writable<{ x: number; y: number }>({ x: 0, y: 0 });
