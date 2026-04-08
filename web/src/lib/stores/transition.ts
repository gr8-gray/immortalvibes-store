// web/src/lib/stores/transition.ts
import { writable } from 'svelte/store';

export type TransitionType = 'T1' | 'T2' | 'T3' | 'T4';

export interface TransitionState {
  active: boolean;
  type: TransitionType | null;
  clickX: number;  // for T2 — streak origin (normalised 0..1)
  clickY: number;
  missionAccent: string;  // for T2 color tint
}

export const transitionStore = writable<TransitionState>({
  active: false,
  type: null,
  clickX: 0.5,
  clickY: 0.5,
  missionAccent: '#4FC3F7',
});

// Per-slug accent colors matching each mission environment
export const MISSION_ACCENT: Record<string, string> = {
  'warped-reality-beanie': '#4FC3F7',
  'vanguard-trucker-hat': 'rgba(200,190,180,0.9)',
  'racerback-tanktop':    'rgba(255,130,50,0.9)',
};

// Ordered mission slugs for T3 prev/next
export const MISSION_ORDER = [
  'warped-reality-beanie',
  'vanguard-trucker-hat',
  'racerback-tanktop',
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
