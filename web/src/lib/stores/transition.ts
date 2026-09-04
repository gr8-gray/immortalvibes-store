// web/src/lib/stores/transition.ts
//
// Page-to-page transition state. Four cinematic transitions cover the four
// kinds of navigation the store knows about (see resolveTransition below);
// this store is how a page hands the overlay its type, origin point, and
// accent color before the route actually changes.
import { writable } from 'svelte/store';
import { SLUG } from '$lib/types/shop';

export type TransitionType = 'T1' | 'T2' | 'T3' | 'T4';

export interface TransitionState {
  active: boolean;
  type: TransitionType | null;
  clickX: number;  // for T2 — streak origin (normalised 0..1)
  clickY: number;
  missionAccent: string;  // for T2 color tint
}

// Per-slug accent colors matching each mission environment. Slugs come from
// the catalog's SLUG map so a renamed product cannot orphan its accent.
export const MISSION_ACCENT: Record<string, string> = {
  [SLUG.beanie]: '#4FC3F7',
  [SLUG.truckerHat]: 'rgba(200,190,180,0.9)',
  [SLUG.tank]:    'rgba(255,130,50,0.9)',
  [SLUG.phantomShorts]: 'rgba(46,111,224,0.9)',
};

export const transitionStore = writable<TransitionState>({
  active: false,
  type: null,
  clickX: 0.5,
  clickY: 0.5,
  missionAccent: '#4FC3F7',
});

// Ordered mission slugs for T3 prev/next
export const MISSION_ORDER: string[] = [
  SLUG.beanie,
  SLUG.truckerHat,
  SLUG.tank,
  SLUG.phantomShorts,
];

// Resolve which transition to use given from/to pathnames
export function resolveTransition(from: string, to: string): TransitionType | null {
  if (to === '/') return 'T4';
  if (to === '/shop') return 'T1';
  if (to.startsWith('/shop/') && from === '/shop') return 'T2';
  if (to.startsWith('/shop/') && from.startsWith('/shop/')) return 'T3';
  return null;
}

// Extract slug from pathname e.g. '/shop/warped-reality-beanie' → 'warped-reality-beanie'
export function slugFromPath(path: string): string {
  return path.split('/').pop() ?? '';
}
