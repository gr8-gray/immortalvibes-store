// web/src/lib/animations/system-reformation.ts
// Always-on intro for /shop during the Warped Reality drop window.
// Three acts: trucker decommission → WR ascension → tank rebalance.
//
// Total duration ~3.9s. Returns a GSAP timeline so callers can attach
// onComplete, scrub via skip button, or kill on unmount.

import { gsap } from 'gsap';

export interface ReformationRefs {
  truckerSlot:     HTMLElement | null;  // mission slot to decommission + unmount
  truckerHalo:     HTMLElement | null;  // diffuse glow ring
  truckerPlanet:   HTMLElement | null;  // the planet-wrap to compress + implode + burst
  wrSlot:          HTMLElement | null;  // Warped Reality (LEO) slot to grow
  wrHalo:          HTMLElement | null;
  wrCluster:       HTMLElement | null;
  tankSlot:        HTMLElement | null;
  // Implosion phase
  collapseRing1:   HTMLElement | null;  // outer ring that crashes inward
  collapseRing2:   HTMLElement | null;
  collapseRing3:   HTMLElement | null;  // innermost ring
  singularity:     HTMLElement | null;  // white pinpoint at implosion peak
  // Supernova phase
  flash:           HTMLElement | null;  // bright radial bloom at burst moment
  shockwave:       HTMLElement | null;  // primary expanding ring
  shockwave2:      HTMLElement | null;  // secondary concentric ring
  shockwave3:      HTMLElement | null;  // tertiary trailing ring
  streaks:         HTMLElement | null;  // 24 radial debris streaks
  particles:       HTMLElement | null;  // ember/debris burst container
}

export interface ReformationOptions {
  onComplete?: () => void;
  onTruckerGone?: () => void; // fires when the trucker has vaporized — caller starts the reflow
  reduceMotion?: boolean;     // collapse to ~0.6s if true
}

export function playSystemReformation(refs: ReformationRefs, opts: ReformationOptions = {}) {
  const { reduceMotion = false } = opts;

  // Reduced-motion fast path: snap to end state, skip animation.
  if (reduceMotion) {
    if (refs.wrSlot)      gsap.set(refs.wrSlot,   { scale: 1.0, opacity: 1 });
    if (refs.wrCluster)   gsap.set(refs.wrCluster,{ opacity: 1 });
    if (refs.tankSlot)    gsap.set(refs.tankSlot, { scale: 1.0, opacity: 1 });
    opts.onTruckerGone?.();
    opts.onComplete?.();
    return gsap.timeline();
  }

  const tl = gsap.timeline({
    defaults: { ease: 'power2.out' },
    onComplete: () => {
      // Permanent end state: trucker removed from layout flow.
      if (refs.truckerSlot) refs.truckerSlot.style.display = 'none';
      opts.onComplete?.();
    },
  });

  // Initial state — set BEFORE animation starts.
  // WR begins at slightly muted scale + cluster hidden, tank at neutral, trucker visible.
  if (refs.wrSlot)    gsap.set(refs.wrSlot,    { scale: 1.0, transformOrigin: '50% 50%' });
  if (refs.wrCluster) gsap.set(refs.wrCluster, { opacity: 0, scale: 0.7 });
  if (refs.tankSlot)  gsap.set(refs.tankSlot,  { scale: 1.0, transformOrigin: '50% 50%', x: 0 });
  if (refs.collapseRing1) gsap.set(refs.collapseRing1, { scale: 1.6, opacity: 0 });
  if (refs.collapseRing2) gsap.set(refs.collapseRing2, { scale: 1.6, opacity: 0 });
  if (refs.collapseRing3) gsap.set(refs.collapseRing3, { scale: 1.6, opacity: 0 });
  if (refs.singularity)   gsap.set(refs.singularity,   { scale: 0,   opacity: 0 });
  if (refs.shockwave)     gsap.set(refs.shockwave,     { scale: 0.1, opacity: 0 });
  if (refs.shockwave2)    gsap.set(refs.shockwave2,    { scale: 0.1, opacity: 0 });
  if (refs.shockwave3)    gsap.set(refs.shockwave3,    { scale: 0.1, opacity: 0 });
  if (refs.flash)         gsap.set(refs.flash,         { scale: 0.2, opacity: 0 });
  if (refs.streaks)       gsap.set(refs.streaks,       { opacity: 0 });
  if (refs.particles)     gsap.set(refs.particles,     { opacity: 0 });

  // ── Act 1 — SUPERNOVA-IMPLOSION (2.4s) ───────────────────────────────────
  // Beats:
  //   0.00–0.40  compression — planet shrinks (anticipation)
  //   0.20–0.85  collapse rings — three rings crash inward to the centre
  //   0.40–0.90  implosion — planet shrinks to a point + brightens
  //   0.85–1.10  singularity hold — white pinpoint pulses (held tension)
  //   1.10–1.16  SUPERNOVA FLASH — sudden bone-white bloom
  //   1.10–1.40  burst — planet explodes outward through the flash
  //   1.10–1.85  triple shockwave rings expand
  //   1.16–1.95  radial debris streaks fly outward
  //   1.20–1.95  ember field bursts
  //   1.40–1.95  planet vaporizes (scale-to-0 + rotation + blur)
  //   1.20–1.70  slot screen-shake
  //   1.95–2.40  cleanup fadeout
  tl.addLabel('act1', 0);

  // (1) Anticipation — planet compresses, slot trembles softly
  if (refs.truckerPlanet) {
    tl.to(refs.truckerPlanet, {
      scale: 0.78, filter: 'brightness(0.7)',
      duration: 0.4, ease: 'power2.in',
    }, 'act1');
  }
  if (refs.truckerSlot) {
    tl.to(refs.truckerSlot, {
      x: '+=2', duration: 0.05, repeat: 6, yoyo: true, ease: 'none',
    }, 'act1+=0.10');
  }

  // (2) Collapse rings — outer to inner, crash inward toward singularity
  const collapseRings = [refs.collapseRing1, refs.collapseRing2, refs.collapseRing3];
  collapseRings.forEach((ring, i) => {
    if (!ring) return;
    tl.fromTo(ring,
      { scale: 1.8, opacity: 0 },
      { scale: 1.8, opacity: 0.9, duration: 0.12 },
      `act1+=${0.20 + i * 0.05}`
    ).to(ring, {
      scale: 0.05, opacity: 1.0,
      duration: 0.55, ease: 'power3.in',
    }, `act1+=${0.32 + i * 0.05}`)
     .to(ring, {
       opacity: 0, duration: 0.08, ease: 'none',
     }, `act1+=${0.87 + i * 0.05}`);
  });

  // (3) Implosion — planet shrinks to point and brightens (becomes the singularity)
  if (refs.truckerPlanet) {
    tl.to(refs.truckerPlanet, {
      scale: 0.04, filter: 'brightness(4) blur(1px)',
      duration: 0.45, ease: 'power3.in',
    }, 'act1+=0.40');
  }

  // (4) Singularity hold — white pinpoint pulses before the burst
  if (refs.singularity) {
    tl.to(refs.singularity, {
      scale: 1.0, opacity: 1, duration: 0.06, ease: 'power3.out',
    }, 'act1+=0.85')
      .to(refs.singularity, {
        scale: 1.6, duration: 0.18, ease: 'power1.inOut',
      }, 'act1+=0.91')
      .to(refs.singularity, {
        scale: 0, opacity: 0, duration: 0.05, ease: 'power3.out',
      }, 'act1+=1.10');
  }

  // (5) SUPERNOVA FLASH — sudden bloom at the burst moment
  if (refs.flash) {
    tl.to(refs.flash, {
      scale: 5.5, opacity: 1.0, duration: 0.06, ease: 'power4.out',
    }, 'act1+=1.10')
      .to(refs.flash, {
        opacity: 0, duration: 0.65, ease: 'power3.in',
      }, 'act1+=1.16');
  }

  // (6) Burst — planet erupts back outward as the explosion fireball
  if (refs.truckerPlanet) {
    tl.to(refs.truckerPlanet, {
      scale: 1.55, filter: 'brightness(2.5) blur(4px)',
      duration: 0.30, ease: 'power3.out',
    }, 'act1+=1.10');
  }

  // (7) Triple shockwaves — primary, secondary, tertiary
  if (refs.shockwave) {
    tl.to(refs.shockwave, {
      scale: 4.2, opacity: 0.95, duration: 0.55, ease: 'power3.out',
    }, 'act1+=1.10')
      .to(refs.shockwave, { opacity: 0, duration: 0.35, ease: 'power1.in' }, 'act1+=1.50');
  }
  if (refs.shockwave2) {
    tl.to(refs.shockwave2, {
      scale: 5.6, opacity: 0.7, duration: 0.70, ease: 'power3.out',
    }, 'act1+=1.18')
      .to(refs.shockwave2, { opacity: 0, duration: 0.4, ease: 'power1.in' }, 'act1+=1.65');
  }
  if (refs.shockwave3) {
    tl.to(refs.shockwave3, {
      scale: 6.8, opacity: 0.5, duration: 0.85, ease: 'power3.out',
    }, 'act1+=1.26')
      .to(refs.shockwave3, { opacity: 0, duration: 0.4, ease: 'power1.in' }, 'act1+=1.85');
  }

  // (8) Halo blasts then dies
  if (refs.truckerHalo) {
    tl.to(refs.truckerHalo, {
      opacity: 1.0, scale: 2.4, duration: 0.35, ease: 'power3.out',
    }, 'act1+=1.10')
      .to(refs.truckerHalo, {
        opacity: 0, duration: 0.55, ease: 'power2.in',
      }, 'act1+=1.45');
  }

  // (9) Radial streaks — debris fanning outward
  if (refs.streaks) {
    tl.to(refs.streaks, {
      opacity: 1, duration: 0.06, ease: 'none',
    }, 'act1+=1.16')
      .to(refs.streaks, {
        opacity: 0, duration: 0.55, ease: 'power2.in',
      }, 'act1+=1.40');
  }

  // (10) Ember/debris burst
  if (refs.particles) {
    tl.to(refs.particles, {
      opacity: 1, duration: 0.08, ease: 'power1.out',
    }, 'act1+=1.20')
      .to(refs.particles, {
        opacity: 0, duration: 0.55, ease: 'power2.in',
      }, 'act1+=1.50');
  }

  // (11) Planet vaporizes — scale to 0 with rotation + blur
  if (refs.truckerPlanet) {
    tl.to(refs.truckerPlanet, {
      scale: 0.0, rotation: 32, filter: 'brightness(0.4) blur(16px)',
      duration: 0.55, ease: 'power3.in',
    }, 'act1+=1.40');
  }

  // (12) Slot screen-shake during the burst
  if (refs.truckerSlot) {
    tl.to(refs.truckerSlot, {
      x: '+=6', y: '-=4', duration: 0.035, repeat: 9, yoyo: true, ease: 'none',
    }, 'act1+=1.10')
      .set(refs.truckerSlot, { x: 0, y: 0 }, 'act1+=1.50');
  }

  // (13) Slot fades out completely after the planet is gone
  if (refs.truckerSlot) {
    tl.to(refs.truckerSlot, {
      opacity: 0, duration: 0.35, ease: 'power1.in',
    }, 'act1+=1.95');
  }

  // (14) Caller hook — start the layout reflow.
  // Fires when the planet is vaporized; runs in parallel with the slot fadeout
  // so the row collapses smoothly under the WR ascension.
  if (opts.onTruckerGone) {
    tl.call(opts.onTruckerGone, [], 'act1+=1.95');
  }

  // ── Act 2 — Warped Reality ascension (1.4s, overlaps act 1 fadeout) ─────
  tl.addLabel('act2', 'act1+=2.05');

  if (refs.wrHalo) {
    tl.to(refs.wrHalo, {
      opacity: 1.0, scale: 1.35, duration: 0.6, ease: 'power2.out',
    }, 'act2');
  }

  if (refs.wrSlot) {
    tl.to(refs.wrSlot, {
      scale: 1.45, duration: 0.55, ease: 'power2.out',
    }, 'act2')
      .to(refs.wrSlot, {
        // Settle to 1.0 so WR matches the tank's CSS size at rest.
        // Drama lives in the 1.45 peak, not the resting state.
        scale: 1.0, duration: 0.55, ease: 'power2.inOut',
      }, 'act2+=0.55');
  }

  if (refs.wrCluster) {
    tl.to(refs.wrCluster, {
      opacity: 1, scale: 1.0,
      duration: 0.85, ease: 'power2.out',
    }, 'act2+=0.4');
  }

  // ── Act 3 — Tank acknowledgement (0.7s) ──────────────────────────────────
  // Tank is the surviving sibling — pulse to acknowledge, then settle at full scale.
  // (No rebalance shrink — diminishing the tank made it feel sidelined.)
  tl.addLabel('act3', 'act2+=1.0');

  if (refs.tankSlot) {
    tl.to(refs.tankSlot, {
      scale: 1.08, duration: 0.35, ease: 'power2.out',
    }, 'act3')
      .to(refs.tankSlot, {
        scale: 1.0, duration: 0.35, ease: 'power1.inOut',
      }, 'act3+=0.35');
  }

  return tl;
}

// Total duration: ~4.4s (act1 ~2.4s + act2 starts at 2.05 + 1.4 = 3.45 + act3 0.7 starting at act2+1.0 = 3.75 + buffer)
export const REFORMATION_DURATION_MS = 4400;
