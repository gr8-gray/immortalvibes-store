<!-- web/src/lib/components/HeroScene.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import gsap from 'gsap';


  let canvas: HTMLCanvasElement;
  let ctx: CanvasRenderingContext2D;
  let rafId: number;
  let w = 0;
  let h = 0;

  // Camera — smoothed via lerp
  let camX = 0;    // -1 (left) to +1 (right)
  let camY = 0;    // -1 (down/ground) to +1 (up/sky)
  let targetX = 0;
  let targetY = 0;

  // Stars in virtual world space
  interface Star {
    wx: number;         // 0..1 across virtual world (3x viewport)
    wy: number;         // 0..1 across sky height
    r: number;
    a: number;
    twinkleSpeed: number;
    twinklePhase: number;
    tint: 'warm' | 'cool' | 'white';
  }

  let stars: Star[] = [];
  const STAR_COUNT = 2600;
  const WORLD_SCALE = 3; // world is 3x viewport wide

  // Milky Way — pre-rendered offscreen canvas
  let mwCanvas: HTMLCanvasElement | null = null;

  // Time progression — golden hour → night over 90 seconds
  let startTime = 0;

  // Silhouette figures — Stitch-generated PNGs with per-pose animation
  interface FigureConfig {
    key: string;
    cxFrac: number;              // horizontal center as fraction of w
    hFrac: number;               // height as fraction of h
    flip: boolean;
    depth: number;               // 0=front (no pan), 1=far (pans with world)
    anim: 'strain' | 'claw' | 'slump' | 'sway-weighted' | 'sway' | 'bob' | 'bounce' | 'awe-snap';
    phase: number;               // time phase offset for variety
    freq: number;                // animation speed
    footFrac: number;            // 1.0=screen bottom, 0.0=horizonY, 0.5=midpoint
  }

  // JJK-quality animation engine — anticipation, overshoot, eased settle
  function easeOut3(t: number): number { return 1 - Math.pow(1 - t, 3); }
  function easeIn2(t: number): number  { return t * t; }

  function animFrame(anim: FigureConfig['anim'], t: number, freq: number, figH: number): { dy: number; rot: number } {
    const p = ((t * freq) % 1 + 1) % 1; // 0-1 cycle phase
    switch (anim) {
      case 'strain': {
        // Anticipate back → snap lunge up → overshoot → fall → rest
        let v: number;
        if      (p < 0.08) v = -(p / 0.08) * 0.2;
        else if (p < 0.30) v = easeIn2((p - 0.08) / 0.22);
        else if (p < 0.42) v = 1.0 + Math.sin((p - 0.30) / 0.12 * Math.PI) * 0.18;
        else if (p < 0.62) v = easeOut3(1 - (p - 0.42) / 0.20);
        else               v = Math.max(0, 0.15 * (1 - (p - 0.62) / 0.38));
        return { dy: -v * figH * 0.09, rot: -v * 0.045 };
      }
      case 'claw': {
        // Faster, no rest — continuous urgent clawing with snap overshoot
        let v: number;
        if      (p < 0.12) v = -(p / 0.12) * 0.15;
        else if (p < 0.38) v = easeIn2((p - 0.12) / 0.26);
        else if (p < 0.50) v = 1.0 + Math.sin((p - 0.38) / 0.12 * Math.PI) * 0.22;
        else               v = easeOut3(1 - (p - 0.50) / 0.50);
        return { dy: -v * figH * 0.08, rot: -v * 0.04 };
      }
      case 'slump': {
        // Defeated slow cycle — lean forward, half-hearted reach, slump back
        const lean = (Math.sin(t * freq) + 1) * 0.5;
        return { dy: lean * figH * 0.015, rot: lean * 0.07 - 0.02 };
      }
      case 'sway-weighted': {
        return { dy: 0, rot: Math.sin(t * freq) * 0.04 };
      }
      case 'sway':      return { dy: 0,                                           rot: Math.sin(t * freq) * 0.033 };
      case 'bob':       return { dy: -Math.abs(Math.sin(t * freq)) * figH * 0.045, rot: 0 };
      case 'bounce':    return { dy: -Math.abs(Math.sin(t * freq)) * figH * 0.08,  rot: 0 };
      case 'awe-snap': {
        const snap = Math.max(0, Math.sin(t * freq * 3.1)) * 0.018;
        return { dy: 0, rot: Math.sin(t * freq) * 0.038 - snap };
      }
    }
  }

  const FIGURES: FigureConfig[] = [
    // All feet at horizonY (footFrac=0.0) — depth communicated by size only
    // ── CLOSE — tallest, heads well above horizon ──
    { key: 'p', cxFrac: 0.07,  hFrac: 0.42, flip: false, depth: 0,   anim: 'strain',        phase: 0.0, freq: 0.8,  footFrac: 0.0 },
    { key: 'r', cxFrac: 0.21,  hFrac: 0.35, flip: true,  depth: 0,   anim: 'slump',         phase: 1.8, freq: 0.4,  footFrac: 0.0 },
    { key: 's', cxFrac: 0.50,  hFrac: 0.40, flip: false, depth: 0,   anim: 'strain',        phase: 3.1, freq: 0.9,  footFrac: 0.0 },
    { key: 'w', cxFrac: 0.73,  hFrac: 0.37, flip: true,  depth: 0,   anim: 'sway-weighted', phase: 2.4, freq: 0.5,  footFrac: 0.0 },
    { key: 't', cxFrac: 0.91,  hFrac: 0.40, flip: false, depth: 0,   anim: 'awe-snap',      phase: 0.7, freq: 0.7,  footFrac: 0.0 },
    // ── MID-CLOSE ──
    { key: 'n', cxFrac: 0.13,  hFrac: 0.30, flip: false, depth: 0.2, anim: 'strain',        phase: 1.1, freq: 1.4,  footFrac: 0.0 },
    { key: 'f', cxFrac: 0.30,  hFrac: 0.28, flip: false, depth: 0.2, anim: 'claw',          phase: 2.2, freq: 1.8,  footFrac: 0.0 },
    { key: 'q', cxFrac: 0.46,  hFrac: 0.27, flip: true,  depth: 0.2, anim: 'claw',          phase: 0.9, freq: 1.6,  footFrac: 0.0 },
    { key: 'u', cxFrac: 0.64,  hFrac: 0.29, flip: false, depth: 0.2, anim: 'claw',          phase: 0.4, freq: 1.5,  footFrac: 0.0 },
    { key: 'v', cxFrac: 0.82,  hFrac: 0.31, flip: false, depth: 0.2, anim: 'strain',        phase: 1.6, freq: 1.1,  footFrac: 0.0 },
    // ── MID ──
    { key: 'm', cxFrac: 0.05,  hFrac: 0.20, flip: true,  depth: 0.4, anim: 'strain',        phase: 2.9, freq: 1.0,  footFrac: 0.0 },
    { key: 'e', cxFrac: 0.22,  hFrac: 0.16, flip: false, depth: 0.4, anim: 'slump',         phase: 0.6, freq: 0.35, footFrac: 0.0 },
    { key: 'o', cxFrac: 0.40,  hFrac: 0.19, flip: false, depth: 0.4, anim: 'sway-weighted', phase: 1.4, freq: 0.55, footFrac: 0.0 },
    { key: 'i', cxFrac: 0.60,  hFrac: 0.18, flip: true,  depth: 0.4, anim: 'strain',        phase: 3.5, freq: 0.85, footFrac: 0.0 },
    { key: 'b', cxFrac: 0.78,  hFrac: 0.20, flip: true,  depth: 0.4, anim: 'claw',          phase: 2.0, freq: 1.6,  footFrac: 0.0 },
    { key: 'c', cxFrac: 0.94,  hFrac: 0.16, flip: false, depth: 0.4, anim: 'claw',          phase: 0.3, freq: 1.9,  footFrac: 0.0 },
    // ── FAR — small bumps at the horizon ──
    { key: 'k', cxFrac: 0.09,  hFrac: 0.12, flip: false, depth: 0.7, anim: 'sway-weighted', phase: 1.2, freq: 0.8,  footFrac: 0.0 },
    { key: 'l', cxFrac: 0.26,  hFrac: 0.11, flip: true,  depth: 0.7, anim: 'strain',        phase: 2.7, freq: 1.1,  footFrac: 0.0 },
    { key: 'h', cxFrac: 0.44,  hFrac: 0.14, flip: false, depth: 0.7, anim: 'claw',          phase: 0.9, freq: 1.3,  footFrac: 0.0 },
    { key: 'd', cxFrac: 0.62,  hFrac: 0.13, flip: true,  depth: 0.7, anim: 'claw',          phase: 1.5, freq: 1.2,  footFrac: 0.0 },
    { key: 'f', cxFrac: 0.80,  hFrac: 0.12, flip: true,  depth: 0.7, anim: 'strain',        phase: 3.8, freq: 0.9,  footFrac: 0.0 },
  ];

  const figureImgs = new Map<string, HTMLImageElement>();

  function loadFigures(): Promise<void[]> {
    return Promise.all(['e','f','h','d','i','c','b','k','l','m','n','o','p','q','r','s','t','u','v','w'].map(key =>
      new Promise<void>(resolve => {
        const img = new Image();
        img.onload  = () => { figureImgs.set(key, img); resolve(); };
        img.onerror = () => resolve();
        img.src = `/fig-${key}.png`;
      })
    ));
  }

  // Drift particles — upward soul-light streams
  interface Particle {
    x: number; y: number; speed: number; alpha: number; r: number;
    type: 'ember' | 'wisp' | 'streak';
    phase: number; wobble: number;
  }
  const PARTICLES: Particle[] = Array.from({ length: 180 }, (_, i) => ({
    x: Math.random(),
    y: Math.random(),
    speed: i < 60  ? 0.08 + Math.random() * 0.18   // slow drifters
          : i < 120 ? 0.22 + Math.random() * 0.35   // mid-speed
                    : 0.45 + Math.random() * 0.70,   // fast streaks
    alpha: 0.12 + Math.random() * 0.65,
    r: i < 60  ? 1.2 + Math.random() * 2.0          // large wisps
       : i < 120 ? 0.6 + Math.random() * 1.2         // medium embers
                 : 0.25 + Math.random() * 0.6,        // tiny sparks
    type: i < 60 ? 'wisp' : i < 140 ? 'ember' : 'streak',
    phase: Math.random() * Math.PI * 2,
    wobble: 0.002 + Math.random() * 0.006,
  }));

  function lerpN(a: number, b: number, t: number): number { return a + (b - a) * t; }
  function lerpC(r1:number,g1:number,b1:number, r2:number,g2:number,b2:number, t:number): string {
    return `rgba(${Math.round(lerpN(r1,r2,t))},${Math.round(lerpN(g1,g2,t))},${Math.round(lerpN(b1,b2,t))},1)`;
  }

  function rand(a: number, b: number): number {
    return a + Math.random() * (b - a);
  }

  function buildStars(): void {
    stars = Array.from({ length: STAR_COUNT }, () => {
      // Bias 35% of stars toward the Milky Way band region
      const inBand = Math.random() < 0.35;
      const wy = inBand ? rand(0.15, 0.65) : Math.random();
      const r = Math.random() < 0.02 ? rand(1.8, 2.6) :
                Math.random() < 0.1  ? rand(1.0, 1.7) :
                Math.random() < 0.35 ? rand(0.5, 1.0) : rand(0.2, 0.5);
      const tint: 'warm' | 'cool' | 'white' =
        Math.random() < 0.12 ? 'warm' :
        Math.random() < 0.25 ? 'cool' : 'white';
      return {
        wx: Math.random(),
        wy,
        r,
        a: rand(0.06, 1.0),
        twinkleSpeed: rand(0.003, 0.016),
        twinklePhase: rand(0, Math.PI * 2),
        tint
      };
    });
  }

  function buildMilkyWay(vw: number, vh: number): void {
    mwCanvas = document.createElement('canvas');
    mwCanvas.width = vw;
    mwCanvas.height = vh;
    const mc = mwCanvas.getContext('2d')!;

    // Milky Way band: diagonal from lower-left to upper-right, ~33deg
    const angle = -0.58; // radians
    const cx = vw * 0.5;
    const cy = vh * 0.48;

    // Multiple blurred passes — wide outer glow to tight warm core
    const passes: Array<{ halfW: number; opacity: number; r: number; g: number; b: number }> = [
      { halfW: 320, opacity: 0.055, r: 180, g: 195, b: 255 }, // wide cool outer
      { halfW: 200, opacity: 0.080, r: 210, g: 218, b: 255 }, // mid blue
      { halfW: 110, opacity: 0.110, r: 238, g: 235, b: 230 }, // core off-white
      { halfW:  55, opacity: 0.090, r: 255, g: 245, b: 210 }, // warm inner
      { halfW:  22, opacity: 0.060, r: 255, g: 250, b: 190 }, // hot center
    ];

    passes.forEach(({ halfW, opacity, r, g, b }) => {
      mc.save();
      mc.filter = `blur(${Math.round(halfW * 0.55)}px)`;
      mc.translate(cx, cy);
      mc.rotate(angle);

      const grad = mc.createLinearGradient(0, -halfW, 0, halfW);
      grad.addColorStop(0,    'rgba(0,0,0,0)');
      grad.addColorStop(0.3,  `rgba(${r},${g},${b},${opacity})`);
      grad.addColorStop(0.5,  `rgba(${r},${g},${b},${opacity * 1.6})`);
      grad.addColorStop(0.7,  `rgba(${r},${g},${b},${opacity})`);
      grad.addColorStop(1,    'rgba(0,0,0,0)');

      mc.fillStyle = grad;
      mc.fillRect(-vw * 1.5, -halfW, vw * 3, halfW * 2);
      mc.restore();
    });
  }

  function drawSky(horizonY: number, panPx: number, progress: number): void {
    // Sky gradient — interpolates from golden hour (0) to full night (1)
    const skyGrd = ctx.createLinearGradient(0, 0, 0, horizonY);
    skyGrd.addColorStop(0,    lerpC(18,14,48,  1,1,7,    progress)); // top
    skyGrd.addColorStop(0.45, lerpC(65,40,100, 1,2,12,   progress)); // upper-mid purple→void
    skyGrd.addColorStop(0.78, lerpC(140,75,30, 2,7,21,   progress)); // amber→deep blue
    skyGrd.addColorStop(1,    lerpC(200,100,20,3,10,30,  progress)); // horizon glow→navy
    ctx.fillStyle = skyGrd;
    ctx.fillRect(0, 0, w, horizonY);

    // Setting sun — warm radial glow at right horizon, fades after progress 0.35
    const sunAlpha = Math.max(0, 1 - progress * 2.8);
    if (sunAlpha > 0.01) {
      const sunX = w * (0.75 - camX * 0.1);
      const sunGrd = ctx.createRadialGradient(sunX, horizonY, 0, sunX, horizonY, w * 0.5);
      sunGrd.addColorStop(0,   `rgba(255,200,80,${sunAlpha * 0.9})`);
      sunGrd.addColorStop(0.15,`rgba(255,140,30,${sunAlpha * 0.6})`);
      sunGrd.addColorStop(0.4, `rgba(200,80,10,${sunAlpha * 0.25})`);
      sunGrd.addColorStop(1,   'rgba(0,0,0,0)');
      ctx.fillStyle = sunGrd;
      ctx.fillRect(0, 0, w, horizonY);
    }

    // Milky Way canvas — only visible at night
    if (mwCanvas && progress > 0.2) {
      const mwAlpha = Math.min(1, (progress - 0.2) / 0.5) * Math.min(1, horizonY / h * 1.6);
      const mwOffX = -(w * WORLD_SCALE - w) * 0.5 + panPx * 0.18;
      const mwOffY = -h * 0.08 + camY * h * 0.08;
      ctx.globalAlpha = mwAlpha;
      ctx.drawImage(mwCanvas, mwOffX, mwOffY, w * WORLD_SCALE, h * 1.2);
      ctx.globalAlpha = 1;
    }

    // Stars — emerge as progress increases
    const starVisibility = Math.max(0, (progress - 0.05) / 0.5);
    if (starVisibility > 0.01) {
      const worldW = w * WORLD_SCALE;
      stars.forEach(star => {
        const sx = star.wx * worldW - (worldW - w) * 0.5 - panPx;
        if (sx < -4 || sx > w + 4) return;
        const sy = star.wy * horizonY;
        star.twinklePhase += star.twinkleSpeed;
        const twink = 0.6 + 0.4 * Math.sin(star.twinklePhase);
        const a = star.a * twink * starVisibility;
        ctx.beginPath();
        ctx.arc(sx, sy, star.r, 0, Math.PI * 2);
        ctx.fillStyle = star.tint === 'warm' ? `rgba(255,242,210,${a})` :
                        star.tint === 'cool' ? `rgba(200,220,255,${a})` :
                                               `rgba(240,237,230,${a})`;
        ctx.fill();
      });
    }

    // Terminator glow — faint warm amber far left, only at golden hour
    if (progress < 0.6) {
      const tAlpha = (1 - progress / 0.6) * 0.06;
      const termX = w * (-0.2 - panPx / w * 0.4);
      const tGrd = ctx.createRadialGradient(termX, horizonY, 0, termX, horizonY, w * 0.75);
      tGrd.addColorStop(0, `rgba(190,95,18,${tAlpha})`);
      tGrd.addColorStop(1, 'rgba(0,0,0,0)');
      ctx.fillStyle = tGrd;
      ctx.fillRect(0, horizonY * 0.4, w, horizonY * 0.6);
    }
  }

  function drawGround(horizonY: number, panPx: number): void {
    // Ground fill — dark earth
    const gGrd = ctx.createLinearGradient(0, horizonY, 0, h);
    gGrd.addColorStop(0,   '#04070e');
    gGrd.addColorStop(0.18,'#02050a');
    gGrd.addColorStop(1,   '#010203');
    ctx.fillStyle = gGrd;
    ctx.fillRect(0, horizonY, w, h - horizonY);

    // Horizon atmosphere glow — thin band
    const hGrd = ctx.createRadialGradient(w * 0.5, horizonY, 0, w * 0.5, horizonY, w * 0.65);
    hGrd.addColorStop(0,   'rgba(79,195,247,0.10)');
    hGrd.addColorStop(0.25,'rgba(40,105,210,0.055)');
    hGrd.addColorStop(1,   'rgba(0,0,0,0)');
    ctx.fillStyle = hGrd;
    ctx.fillRect(0, horizonY - h * 0.04, w, h * 0.09);

    // Horizon line
    const lGrd = ctx.createLinearGradient(0, 0, w, 0);
    lGrd.addColorStop(0,   'rgba(79,195,247,0)');
    lGrd.addColorStop(0.15,'rgba(79,195,247,0.22)');
    lGrd.addColorStop(0.5, 'rgba(130,218,255,0.40)');
    lGrd.addColorStop(0.85,'rgba(79,195,247,0.22)');
    lGrd.addColorStop(1,   'rgba(79,195,247,0)');
    ctx.beginPath();
    ctx.moveTo(0, horizonY);
    ctx.lineTo(w, horizonY);
    ctx.strokeStyle = lGrd;
    ctx.lineWidth = 1;
    ctx.stroke();

  }


  function drawParticles(horizonY: number, progress: number, t: number): void {
    const intensity = Math.max(0, (progress - 0.05) * 1.2);
    if (intensity < 0.01) return;

    PARTICLES.forEach(p => {
      // Rise upward — faster particles rise faster
      p.y -= p.speed * 0.0004;
      // Gentle horizontal wobble
      p.phase += p.wobble;
      const wobbleX = Math.sin(p.phase) * 0.003;

      // Reset when off the top — respawn near horizon
      if (p.y < 0) {
        p.y = (horizonY / h) * (0.85 + Math.random() * 0.15);
        p.x = Math.random();
        p.phase = Math.random() * Math.PI * 2;
      }

      const px = (p.x + wobbleX) * w;
      const py = p.y * h;
      if (py > horizonY || py < 0) return;

      // Fade out near top of sky, fade in from horizon
      const skyFrac = 1 - (py / horizonY);
      const edgeFade = Math.min(skyFrac * 4, 1) * Math.min((1 - skyFrac) * 6, 1);
      const a = p.alpha * intensity * edgeFade;
      if (a < 0.01) return;

      if (p.type === 'streak') {
        // Vertical streak — thin rising line
        const len = p.r * 8 + p.speed * 18;
        const grad = ctx.createLinearGradient(px, py, px, py + len);
        grad.addColorStop(0, `rgba(255,235,180,${a})`);
        grad.addColorStop(1, `rgba(255,200,100,0)`);
        ctx.beginPath();
        ctx.moveTo(px, py);
        ctx.lineTo(px + Math.sin(p.phase) * 1.5, py + len);
        ctx.strokeStyle = grad;
        ctx.lineWidth = p.r * 0.8;
        ctx.stroke();
      } else if (p.type === 'wisp') {
        // Large soft glow orb
        const grd = ctx.createRadialGradient(px, py, 0, px, py, p.r * 3.5);
        grd.addColorStop(0,   `rgba(255,240,200,${a * 0.9})`);
        grd.addColorStop(0.4, `rgba(255,210,130,${a * 0.5})`);
        grd.addColorStop(1,   `rgba(255,180,80,0)`);
        ctx.beginPath();
        ctx.arc(px, py, p.r * 3.5, 0, Math.PI * 2);
        ctx.fillStyle = grd;
        ctx.fill();
      } else {
        // Crisp ember dot
        ctx.beginPath();
        ctx.arc(px, py, p.r, 0, Math.PI * 2);
        const warm = Math.random() < 0.15;
        ctx.fillStyle = warm
          ? `rgba(200,230,255,${a})`
          : `rgba(255,215,130,${a})`;
        ctx.fill();
      }
    });
  }

  function drawSilhouettes(horizonY: number, panPx: number, t: number): void {
    // Crowd fades as camY rises toward RISE BEYOND — sync with text fade-in at 0.35
    const crowdAlpha = Math.max(0, Math.min(1, 1 - (camY - 0.32) / 0.26));
    if (crowdAlpha <= 0) return;

    // Painter's algorithm: draw smallest (far) first, largest (close) on top
    const sorted = [...FIGURES].sort((a, b) => a.hFrac - b.hFrac);

    for (const fig of sorted) {
      const img = figureImgs.get(fig.key);
      if (!img) continue;

      const figH = h * fig.hFrac;
      const figW = (img.naturalWidth / img.naturalHeight) * figH;
      const footY = horizonY + (h - horizonY) * fig.footFrac;
      const cx = fig.cxFrac * w - panPx * fig.depth * 0.6;

      ctx.save();
      ctx.globalAlpha = crowdAlpha;

      // Warm ground-fire backlight — embers glow up from below
      ctx.shadowBlur  = 22 + fig.hFrac * 20;
      ctx.shadowColor = `rgba(210, 120, 30, ${0.55 * crowdAlpha})`;

      ctx.translate(cx, footY);
      if (fig.flip) ctx.scale(-1, 1);
      ctx.drawImage(img, -figW / 2, -figH, figW, figH);

      // Second pass — cool blue atmosphere at shoulders/top (offset upward)
      ctx.shadowBlur  = 14;
      ctx.shadowColor = `rgba(79, 195, 247, ${0.22 * crowdAlpha})`;
      ctx.globalAlpha = crowdAlpha * 0.35;
      ctx.drawImage(img, -figW / 2, -figH - figH * 0.08, figW, figH);

      ctx.restore();
    }
    ctx.globalAlpha = 1;
    ctx.shadowBlur  = 0;
    ctx.shadowColor = 'transparent';
  }

  // Mission planet orbs — appear in sky when tilted up
  function drawMissionOrbits(horizonY: number, panPx: number, t: number): void {
    const orbAlpha = Math.max(0, Math.min(1, (camY - 0.28) / 0.30));
    if (orbAlpha < 0.01) return;

    const orbs = [
      { xFrac: 0.22, yFrac: 0.18, r: 5,  color: [79,  195, 247], pan: -0.15 }, // LEO — blue
      { xFrac: 0.50, yFrac: 0.10, r: 4,  color: [200, 184, 154], pan:  0.00 }, // Lunar — warm gray
      { xFrac: 0.78, yFrac: 0.16, r: 5,  color: [185,  60, 255], pan:  0.15 }, // Nebula — bright violet
    ];

    orbs.forEach(({ xFrac, yFrac, r, color, pan }) => {
      const ox = xFrac * w - panPx * pan;
      const oy = horizonY * yFrac;
      const pulse = 0.75 + 0.25 * Math.sin(t * 0.8 + xFrac * 6);
      const a = orbAlpha * pulse;
      const [cr, cg, cb] = color;

      // Dark backdrop — punches through the Milky Way so orb reads clearly
      const dark = ctx.createRadialGradient(ox, oy, 0, ox, oy, r * 12);
      dark.addColorStop(0,   `rgba(0,0,4,${a * 0.72})`);
      dark.addColorStop(0.4, `rgba(0,0,4,${a * 0.40})`);
      dark.addColorStop(1,   'rgba(0,0,4,0)');
      ctx.beginPath();
      ctx.arc(ox, oy, r * 12, 0, Math.PI * 2);
      ctx.fillStyle = dark;
      ctx.fill();

      // Outer diffuse color glow
      const grd = ctx.createRadialGradient(ox, oy, 0, ox, oy, r * 8);
      grd.addColorStop(0,   `rgba(${cr},${cg},${cb},${a * 0.55})`);
      grd.addColorStop(0.35,`rgba(${cr},${cg},${cb},${a * 0.22})`);
      grd.addColorStop(1,   `rgba(${cr},${cg},${cb},0)`);
      ctx.beginPath();
      ctx.arc(ox, oy, r * 8, 0, Math.PI * 2);
      ctx.fillStyle = grd;
      ctx.fill();

      // Orbit ring — thin circle, unmistakably a destination not a star
      ctx.beginPath();
      ctx.arc(ox, oy, r * 3.2, 0, Math.PI * 2);
      ctx.strokeStyle = `rgba(${cr},${cg},${cb},${a * 0.55})`;
      ctx.lineWidth = 0.8;
      ctx.stroke();

      // Inner glow + hard core
      const mid = ctx.createRadialGradient(ox, oy, 0, ox, oy, r * 2);
      mid.addColorStop(0,   `rgba(${cr},${cg},${cb},${a * 0.95})`);
      mid.addColorStop(1,   `rgba(${cr},${cg},${cb},0)`);
      ctx.beginPath();
      ctx.arc(ox, oy, r * 2, 0, Math.PI * 2);
      ctx.fillStyle = mid;
      ctx.fill();

      ctx.beginPath();
      ctx.arc(ox, oy, r * 0.8, 0, Math.PI * 2);
      ctx.fillStyle = `rgba(255,255,255,${a * 0.95})`;
      ctx.fill();
    });
  }

  let mwPhotoEl: HTMLElement;
  let mortalEl: HTMLElement;
  let riseEl: HTMLElement;
  let ctaEl: HTMLAnchorElement;
  let logoEl: HTMLImageElement;

  function draw(): void {
    if (!ctx || !canvas) return;

    // Neck-tilt feel — fast enough to feel physical
    camX += (targetX - camX) * 0.20;
    camY += (targetY - camY) * 0.20;

    // Time progression: 0 = golden hour, 1 = full night (over 90 seconds)
    const progress = Math.min(1, (Date.now() - startTime) / 90000);

    // Neutral shows 65% sky. Ground exits at camY≈0.49 — exactly at button threshold.
    const horizonY = h * (0.68 + camY * 0.72);
    const panPx    = camX * w * 0.40;

    ctx.clearRect(0, 0, w, h);
    const t = (Date.now() - startTime) / 1000;
    drawSky(horizonY, panPx, progress);
    drawMissionOrbits(horizonY, panPx, t);
    drawGround(horizonY, panPx);
    drawSilhouettes(horizonY, panPx, t);
    drawParticles(horizonY, progress, t);

    // IV logo — visible at neutral, fades as you look up
    if (logoEl) {
      logoEl.style.opacity = String(Math.max(0, 1 - camY * 3.5));
    }

    // "THE MORTAL PLANE" — anchored just above the horizon, fades early in tilt
    if (mortalEl) {
      mortalEl.style.top = `${horizonY - 28}px`;
      mortalEl.style.opacity = String(Math.max(0, 1 - camY * 2.8));
    }

    // "RISE BEYOND" — emerges as you enter sky, fully visible before button
    if (riseEl) {
      riseEl.style.opacity = String(Math.min(1, Math.max(0, (camY - 0.35) * 4)));
    }

    // CTA button — only reachable when fully in sky realm (ground ~6% strip, mortal plane gone)
    if (ctaEl) {
      const ctaOpacity = Math.max(0, (camY - 0.50) * 5);
      ctaEl.style.opacity = String(ctaOpacity);
      ctaEl.style.pointerEvents = ctaOpacity < 0.1 ? 'none' : 'all';
    }

    // MW photo: faint base "leak" always visible, swells to full at tilt
    if (mwPhotoEl) {
      const mwOpacity = Math.min(0.88, 0.04 + Math.max(0, camY) * 0.86);
      mwPhotoEl.style.opacity = String(mwOpacity);
      // Scroll image upward as camY increases — shows zenith at full tilt
      mwPhotoEl.style.backgroundPosition = `center ${Math.max(0, 25 - camY * 30)}%`;
      // Dynamic mask — fills more screen as you look up
      const maskFade = Math.round(72 + camY * 28); // 72% at neutral → 100% at full tilt
      const maskStr = `linear-gradient(to bottom, black 30%, transparent ${maskFade}%)`;
      mwPhotoEl.style.maskImage = maskStr;
      mwPhotoEl.style.webkitMaskImage = maskStr;
    }
    rafId = requestAnimationFrame(draw);
  }

  function resize(): void {
    if (!canvas) return;
    w = window.innerWidth;
    h = window.innerHeight;
    canvas.width  = w;
    canvas.height = h;
    buildMilkyWay(w * WORLD_SCALE, h * 1.25);
  }

  onMount(() => {
    if (!browser) return;

    ctx = canvas.getContext('2d')!;
    w = window.innerWidth;
    h = window.innerHeight;
    canvas.width  = w;
    canvas.height = h;

    buildStars();
    buildMilkyWay(w * WORLD_SCALE, h * 1.25);
    loadFigures();

    window.addEventListener('resize', resize);

    function onMouseMove(e: MouseEvent) {
      targetX = Math.max(-1, Math.min(1,  (e.clientX / w - 0.5) * 2.5));
      targetY = Math.max(-0.5, Math.min(1, -((e.clientY / h - 0.5) * 2.5)));
    }

    function onMouseLeave() { targetX = 0; targetY = 0; }

    // ── Touch drag ────────────────────────────────────────────────────────
    let touchStartY = 0;
    let touchStartX = 0;
    let touchBaseY  = 0;
    let touchBaseX  = 0;

    function onTouchStart(e: TouchEvent) {
      touchStartY = e.touches[0].clientY;
      touchStartX = e.touches[0].clientX;
      touchBaseY  = targetY;
      touchBaseX  = targetX;
    }

    function onTouchMove(e: TouchEvent) {
      const dy = (e.touches[0].clientY - touchStartY) / h;
      const dx = (e.touches[0].clientX - touchStartX) / w;
      targetY = Math.max(-0.5, Math.min(1, touchBaseY - dy * 2.2));
      targetX = Math.max(-1,   Math.min(1, touchBaseX + dx * 2.0));
    }

    function onTouchEnd() { /* keep position — don't snap back */ }

    // ── Device orientation (gyroscope) ───────────────────────────────────
    let gyroActive = false;

    function onDeviceOrientation(e: DeviceOrientationEvent) {
      if (e.beta === null || e.gamma === null) return;
      // beta: 0=flat, 90=upright portrait. Tilt back = look up.
      const tiltY = -((e.beta - 75) / 40);  // 0 at neutral upright, +1 tilted back
      const tiltX = (e.gamma ?? 0) / 35;
      targetY = Math.max(-0.5, Math.min(1, tiltY));
      targetX = Math.max(-1,   Math.min(1, tiltX));
    }

    async function tryGyro() {
      if (typeof DeviceOrientationEvent === 'undefined') return;
      const DOE = DeviceOrientationEvent as unknown as { requestPermission?: () => Promise<string> };
      if (typeof DOE.requestPermission === 'function') {
        // iOS 13+ — needs user gesture; we attach to first touch
        const handler = async () => {
          const perm = await DOE.requestPermission!();
          if (perm === 'granted') {
            window.addEventListener('deviceorientation', onDeviceOrientation);
            gyroActive = true;
          }
          window.removeEventListener('touchstart', handler);
        };
        window.addEventListener('touchstart', handler, { once: true });
      } else {
        // Android / non-gated
        window.addEventListener('deviceorientation', onDeviceOrientation);
        gyroActive = true;
      }
    }

    const isMobile = browser && /Android|iPhone|iPad|iPod/i.test(navigator.userAgent);

    window.addEventListener('mousemove', onMouseMove);
    window.addEventListener('mouseleave', onMouseLeave);

    if (isMobile) {
      window.addEventListener('touchstart', onTouchStart, { passive: true });
      window.addEventListener('touchmove',  onTouchMove,  { passive: true });
      window.addEventListener('touchend',   onTouchEnd);
      tryGyro();
    }

    // Entry reveal — fade canvas in
    gsap.fromTo(canvas,
      { opacity: 0 },
      { opacity: 1, duration: 2.2, ease: 'power2.inOut', delay: 0.4 }
    );


    startTime = Date.now();
    rafId = requestAnimationFrame(draw);

    return () => {
      cancelAnimationFrame(rafId);
      window.removeEventListener('mousemove', onMouseMove);
      window.removeEventListener('mouseleave', onMouseLeave);
      window.removeEventListener('resize', resize);
      if (isMobile) {
        window.removeEventListener('touchstart', onTouchStart);
        window.removeEventListener('touchmove',  onTouchMove);
        window.removeEventListener('touchend',   onTouchEnd);
        if (gyroActive) window.removeEventListener('deviceorientation', onDeviceOrientation);
      }
    };
  });
</script>

<!-- IV logo — top center, fades as you tilt up -->
<img
  bind:this={logoEl}
  src="/logo-bare.png"
  alt="Immortal Vibes"
  class="hero-logo"
  aria-hidden="true"
/>

<!-- THE MORTAL PLANE text — anchored to the horizon -->
<div
  bind:this={mortalEl}
  aria-hidden="true"
  class="text-mortal"
>
  THE MORTAL PLANE
</div>

<!-- RISE BEYOND — lives in the stars, revealed at deep tilt -->
<div
  bind:this={riseEl}
  aria-hidden="true"
  class="text-rise"
>
  RISE BEYOND
</div>

<a
  bind:this={ctaEl}
  href="/shop"
  class="text-cta"
>
  ENTER THE MISSIONS
</a>

<!-- Milky Way photo — screen blend adds nebula color on top of canvas sky -->
<!-- Replace URL with /milky-way.jpg once you save the photo to web/static/ -->
<div
  bind:this={mwPhotoEl}
  aria-hidden="true"
  class="mw-photo"
></div>

<canvas
  bind:this={canvas}
  aria-hidden="true"
  style="position: fixed; inset: 0; width: 100vw; height: 100vh; pointer-events: none; z-index: 2;"
></canvas>

<style>
  .hero-logo {
    position: fixed;
    top: 1.4rem;
    left: 50%;
    transform: translateX(-50%);
    width: 80px;
    height: 80px;
    object-fit: contain;
    pointer-events: none;
    z-index: 5;
    opacity: 1;
    will-change: opacity;
  }

  .text-mortal {
    position: fixed;
    left: 50%;
    transform: translateX(-50%);
    font-family: 'Cormorant Garamond', serif;
    font-size: 0.75rem;
    font-weight: 300;
    letter-spacing: 0.4em;
    color: rgba(240, 237, 230, 0.55);
    white-space: nowrap;
    pointer-events: none;
    z-index: 5;
    opacity: 0;
    text-shadow: 0 0 20px rgba(0, 0, 0, 0.9);
  }

  .text-rise {
    position: fixed;
    top: 14%;
    left: 50%;
    transform: translateX(-50%);
    font-family: 'Cormorant Garamond', serif;
    font-size: 1.6rem;
    font-weight: 300;
    letter-spacing: 0.45em;
    color: rgba(240, 237, 230, 0.88);
    white-space: nowrap;
    pointer-events: none;
    z-index: 5;
    opacity: 0;
    text-shadow: 0 0 60px rgba(240, 237, 230, 0.15);
  }

  .mw-photo {
    position: fixed;
    inset: 0;
    background-image: url('/milky-way.webp');
    background-size: cover;
    background-position: center 25%; /* overridden dynamically in draw() */
    opacity: 0;
    mix-blend-mode: screen;
    pointer-events: none;
    z-index: 3;
    will-change: opacity, background-position;
  }

  .text-cta {
    position: fixed;
    top: 22%;
    left: 50%;
    transform: translateX(-50%);
    z-index: 5;
    display: inline-block;
    border: 1px solid rgba(240, 237, 230, 0.3);
    border-bottom-color: rgba(200, 146, 42, 0.5);
    color: rgba(240, 237, 230, 0.9);
    font-family: 'Inter', sans-serif;
    font-size: 0.60rem;
    letter-spacing: 0.22em;
    padding: 1rem 2.5rem;
    text-decoration: none;
    background: rgba(0, 0, 0, 0.4);
    backdrop-filter: blur(8px);
    box-shadow: 0 0 30px rgba(240, 237, 230, 0.08), 0 0 60px rgba(200, 146, 42, 0.06);
    animation: ctaPulse 2.8s ease-in-out infinite;
    opacity: 0;
    pointer-events: none;
    white-space: nowrap;
    cursor: pointer;
    transition: border-color 0.2s, box-shadow 0.2s, color 0.2s;
  }

  .text-cta:hover {
    border-color: rgba(240, 237, 230, 0.8);
    border-bottom-color: rgba(200, 146, 42, 1);
    color: #F0EDE6;
    box-shadow: 0 0 50px rgba(240, 237, 230, 0.18), 0 0 100px rgba(200, 146, 42, 0.14);
  }

  @keyframes ctaPulse {
    0%, 100% { border-bottom-color: rgba(200, 146, 42, 0.15); }
    50%       { border-bottom-color: rgba(200, 146, 42, 0.65); }
  }
</style>
