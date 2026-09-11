// web/src/lib/transitions/t5-hyperspace.ts
//
// Hyperspace engine — a true 3D projected starfield. Stars streak outward from a
// vanishing point (the "origin"), neon-blue at the edges ramping to white-blue at
// the throat, additive compositing so the throat blooms. Timeline: charge ->
// violent accel -> brief lightspeed -> decelerate (coast to a stop) -> settle. The
// route swaps at the midpoint under a gentle blue-white cover veil; no harsh flash.
//
// One engine drives three transitions, differing only in origin + character:
//   T5 — interstellar jump, center origin. About / Contact / enter-store / to-Earth.
//   T2 — dive INTO the clicked planet: origin set to the planet's screen position.
//   T3 — quick jump between planets, center origin, snappier.
//
// Tuned live with Eric (T5 2026-09-05; T2/T3 origin-aware 2026-09-11). Character
// lives in HYPER_PROFILES; the shared frame math lives in playHyperspace.
import gsap from 'gsap';

export interface T5Elements {
  overlay: HTMLElement;
  canvas: HTMLCanvasElement;
}

export interface HyperProfile {
  dur: number;      // total seconds
  density: number;  // star count
  streak: number;   // motion-blur (streak) length multiplier
  warp: number;     // peak acceleration intensity
  swirl: number;    // rotational twist as stars near the camera (0 = pure radial)
  cover: number;    // blue-white mask alpha at the swap point (0 = none, 1 = full)
  mid: number;      // progress at which the route swaps (== peakEnd)
  originX: number;  // vanishing point, normalised 0..1 (0.5 = screen center)
  originY: number;
}

const BASE: HyperProfile = {
  dur: 2.0, density: 1600, streak: 2.0, warp: 1.25,
  swirl: 0.5, cover: 0.35, mid: 0.58, originX: 0.5, originY: 0.5,
};

// Per-transition character. Origin for T2 is overridden per-click at call time.
export const HYPER_PROFILES = {
  T5: { ...BASE },
  T2: { ...BASE, dur: 1.5, density: 1400, warp: 1.4, swirl: 0.62, cover: 0.4, mid: 0.55 },
  T3: { ...BASE, dur: 1.2, density: 1200, warp: 1.5, swirl: 0.72, cover: 0.42, mid: 0.5 },
} satisfies Record<string, HyperProfile>;

const smooth = (t: number) => (t <= 0 ? 0 : t >= 1 ? 1 : t * t * (3 - 2 * t));

interface Star { x: number; y: number; z: number; }

// The shared engine. Origin + character come from `p` (a HyperProfile).
export function playHyperspace(
  els: T5Elements,
  prof: HyperProfile,
  onMidpoint: () => void,
  onComplete: () => void
): gsap.core.Timeline {
  const { overlay, canvas } = els;
  const CFG = prof;
  // Phase boundaries as fractions of dur, anchored to the swap point (peakEnd == mid).
  const T = { hold: 0.08, peakStart: CFG.mid - 0.14, peakEnd: CFG.mid, decelEnd: CFG.mid + 0.34 };
  const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

  // Reduced motion: skip the warp — quick calm cross-fade with the swap under cover.
  if (reduced) {
    const tl = gsap.timeline({ onComplete });
    gsap.set(overlay, { display: 'block', opacity: 0 });
    gsap.set(canvas, { opacity: 0 });
    tl.to(overlay, { opacity: 1, duration: 0.18, ease: 'power2.out' }, 0)
      .call(onMidpoint, [], 0.2)
      .to(overlay, { opacity: 0, duration: 0.24, ease: 'power2.inOut' }, 0.24)
      .call(() => gsap.set(overlay, { display: 'none', opacity: 1 }));
    return tl;
  }

  const ctx = canvas.getContext('2d')!;
  let W = 0, H = 0, cx = 0, cy = 0, scale = 1, maxDim = 1;

  function resize(): void {
    const dpr = Math.min(window.devicePixelRatio || 1, 2);
    W = window.innerWidth; H = window.innerHeight;
    canvas.width = Math.round(W * dpr);
    canvas.height = Math.round(H * dpr);
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    cx = W * CFG.originX; cy = H * CFG.originY;   // origin-aware vanishing point
    maxDim = Math.hypot(W, H);
    scale = Math.min(W, H) * 0.62;
  }

  const rnd = (a: number, b: number) => a + Math.random() * (b - a);
  function respawn(s: Star, z?: number): void {
    s.x = rnd(-1.5, 1.5); s.y = rnd(-1.5, 1.5);
    s.z = z == null ? rnd(0.05, 1) : z;
  }
  let stars: Star[] = [];
  function buildStars(): void {
    stars = new Array(CFG.density);
    for (let i = 0; i < CFG.density; i++) {
      const s: Star = { x: 0, y: 0, z: 0 };
      respawn(s);
      stars[i] = s;
    }
  }

  const maxSp = 3.1 * CFG.warp;
  function speedAt(p: number): number {
    const base = 0.05;
    if (p < T.hold) return base;
    if (p < T.peakStart) {
      const a = (p - T.hold) / (T.peakStart - T.hold);
      return base + Math.pow(a, 2.3) * maxSp;         // violent acceleration
    }
    if (p < T.peakEnd) return base + maxSp;           // brief lightspeed
    if (p < T.decelEnd) {
      const a = (p - T.peakEnd) / (T.decelEnd - T.peakEnd);
      return base + maxSp * (1 - smooth(a));          // ease down — coast to a stop
    }
    return base;                                      // settled drift
  }

  function step(dt: number, p: number): number {
    const sp = speedAt(p);
    for (const s of stars) {
      s.z -= sp * dt;
      if (s.z < 0.03) respawn(s, 1);
    }
    return sp;
  }

  function drawFrame(p: number, sp: number): void {
    ctx.globalCompositeOperation = 'source-over';
    ctx.fillStyle = '#02030a';
    ctx.fillRect(0, 0, W, H);

    const speedN = Math.min(sp / maxSp, 1);
    const twist = p < T.hold ? 0 : CFG.swirl;

    ctx.globalCompositeOperation = 'lighter';
    ctx.lineCap = 'round';

    for (const s of stars) {
      const zPrev = Math.min(1.2, s.z + sp * (1 / 60) * CFG.streak);
      const aCur = twist * (1 - s.z), aPrev = twist * (1 - zPrev);
      const cc = Math.cos(aCur), sc = Math.sin(aCur), cp = Math.cos(aPrev), sp2 = Math.sin(aPrev);
      const xC = s.x * cc - s.y * sc, yC = s.x * sc + s.y * cc;
      const xP = s.x * cp - s.y * sp2, yP = s.x * sp2 + s.y * cp;

      const sx = cx + (xC / s.z) * scale, sy = cy + (yC / s.z) * scale;
      const px = cx + (xP / zPrev) * scale, py = cy + (yP / zPrev) * scale;
      if ((sx < -40 && px < -40) || (sx > W + 40 && px > W + 40) ||
          (sy < -40 && py < -40) || (sy > H + 40 && py > H + 40)) continue;

      const near = 1 - s.z;                             // 0 far .. 1 close
      const t = Math.min(near * 1.15 * speedN + 0.12, 1); // neon-blue -> white-blue
      const r = (58 + (234 - 58) * t) | 0, g = (160 + (244 - 160) * t) | 0;
      let a = Math.min(near * 1.05 + 0.14, 1);
      if (p < T.hold) a *= 0.55;                        // dim the static charge field
      const w = 0.4 + near * 3.0 * (0.6 + 0.4 * speedN);

      ctx.strokeStyle = `rgba(${r},${g},255,${a.toFixed(3)})`;
      ctx.lineWidth = w;
      ctx.beginPath(); ctx.moveTo(px, py); ctx.lineTo(sx, sy); ctx.stroke();
    }

    // soft blue-white throat bloom — tracks speed, swells at lightspeed, calmly fades.
    const glow = speedN * speedN;
    if (glow > 0.02) {
      const rad = maxDim * (0.14 + glow * 0.55);
      const gr = ctx.createRadialGradient(cx, cy, 0, cx, cy, rad);
      gr.addColorStop(0, `rgba(206,230,255,${(0.40 * glow).toFixed(3)})`);
      gr.addColorStop(0.4, `rgba(92,178,255,${(0.28 * glow).toFixed(3)})`);
      gr.addColorStop(1, 'rgba(58,160,255,0)');
      ctx.globalCompositeOperation = 'lighter';
      ctx.fillStyle = gr; ctx.fillRect(0, 0, W, H);
    }

    // cover veil — a brief blue-white mask centered on the swap point so the DOM
    // route-swap is hidden. cov=0 -> pure calm, cov=1 -> full cover.
    if (CFG.cover > 0.001) {
      const half = 0.20;
      let cw = 1 - Math.min(Math.abs(p - T.peakEnd) / half, 1);
      cw = smooth(cw);
      const a = CFG.cover * cw;
      if (a > 0.002) {
        ctx.globalCompositeOperation = 'source-over';
        ctx.fillStyle = `rgba(206,228,255,${a.toFixed(3)})`;
        ctx.fillRect(0, 0, W, H);
      }
    }
  }

  // ── run ──
  resize();
  buildStars();
  gsap.set(overlay, { display: 'block', opacity: 1 });
  gsap.set(canvas, { opacity: 1 });

  const onResize = () => resize();
  window.addEventListener('resize', onResize);

  const tl = gsap.timeline({ onComplete });
  let rafId = 0;
  let last: number | null = null;

  function loop(ts: number): void {
    const p = Math.min(tl.time() / CFG.dur, 1);
    if (last == null) last = ts;
    let dt = (ts - last) / 1000; last = ts;
    if (dt > 0.05) dt = 0.05;                          // clamp tab-switch spikes
    const sp = step(dt, p);
    drawFrame(p, sp);
    if (p < 1) rafId = requestAnimationFrame(loop);
  }

  tl.call(onMidpoint, [], CFG.dur * CFG.mid)            // route swap under cover veil
    .to(overlay, { opacity: 0, duration: 0.34, ease: 'power2.inOut' }, CFG.dur * 0.86)
    .call(() => {
      cancelAnimationFrame(rafId);
      window.removeEventListener('resize', onResize);
      ctx.setTransform(1, 0, 0, 1, 0, 0);
      ctx.clearRect(0, 0, canvas.width, canvas.height);
      gsap.set(overlay, { display: 'none', opacity: 1 });
      gsap.set(canvas, { opacity: 1 });
    }, [], CFG.dur * 0.86 + 0.34);

  rafId = requestAnimationFrame(loop);
  return tl;
}

// T5 — interstellar jump, center origin. Kept as the named entry the overlay uses
// for About / Contact / enter-store / return-to-Earth.
export function playT5(
  els: T5Elements,
  onMidpoint: () => void,
  onComplete: () => void
): gsap.core.Timeline {
  return playHyperspace(els, HYPER_PROFILES.T5, onMidpoint, onComplete);
}

// T2 — dive into the clicked planet. origin = the planet's screen position (0..1).
export function playT2(
  els: T5Elements,
  originX: number,
  originY: number,
  onMidpoint: () => void,
  onComplete: () => void
): gsap.core.Timeline {
  return playHyperspace(els, { ...HYPER_PROFILES.T2, originX, originY }, onMidpoint, onComplete);
}

// T3 — quick jump between planets, center origin.
export function playT3(
  els: T5Elements,
  onMidpoint: () => void,
  onComplete: () => void
): gsap.core.Timeline {
  return playHyperspace(els, HYPER_PROFILES.T3, onMidpoint, onComplete);
}
