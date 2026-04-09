<!-- web/src/lib/components/HeroScene.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import gsap from 'gsap';

  // Callback: parent receives camY so it can show/hide CTA
  export let onCameraUpdate: (camY: number) => void = () => {};

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

  // Drift particles — warm embers floating upward
  interface Particle { x: number; y: number; speed: number; alpha: number; r: number; }
  const PARTICLES: Particle[] = Array.from({ length: 55 }, () => ({
    x: Math.random(),
    y: Math.random(),
    speed: 0.15 + Math.random() * 0.45,
    alpha: 0.15 + Math.random() * 0.5,
    r: 0.5 + Math.random() * 1.2
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
    const starVisibility = Math.max(0, (progress - 0.15) / 0.6);
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

    // Landscape silhouettes
    drawLandscape(horizonY, panPx);
  }

  function drawLandscape(horizonY: number, panPx: number): void {
    const dark = 'rgba(2,4,9,0.97)';
    ctx.fillStyle = dark;

    // ── LEFT: Desert mesa/ridge — slides in as you look left ──
    const mesaX = w * 0.08 - panPx * 1.5;
    ctx.beginPath();
    ctx.moveTo(mesaX - 120, h);
    ctx.lineTo(mesaX - 120, horizonY + 6);
    ctx.lineTo(mesaX - 50,  horizonY + 1);
    ctx.lineTo(mesaX,       horizonY - 24);
    ctx.lineTo(mesaX + 55,  horizonY - 30);
    ctx.lineTo(mesaX + 110, horizonY - 20);
    ctx.lineTo(mesaX + 185, horizonY - 8);
    ctx.lineTo(mesaX + 260, horizonY - 3);
    ctx.lineTo(mesaX + 340, horizonY + 2);
    ctx.lineTo(mesaX + 420, h);
    ctx.closePath();
    ctx.fill();

    // Smaller rock outcrop — further left
    const rock2X = mesaX - 280;
    ctx.beginPath();
    ctx.moveTo(rock2X,        h);
    ctx.lineTo(rock2X,        horizonY + 3);
    ctx.lineTo(rock2X + 30,   horizonY - 10);
    ctx.lineTo(rock2X + 70,   horizonY - 14);
    ctx.lineTo(rock2X + 110,  horizonY - 6);
    ctx.lineTo(rock2X + 150,  h);
    ctx.closePath();
    ctx.fill();

    // ── RIGHT: Launch tower — slides in as you look right ──
    const towerX = w * 0.92 - panPx * 1.7;
    ctx.fillStyle = dark;

    // Tower body
    ctx.fillRect(towerX - 7,  horizonY - 200, 14, 200 + (h - horizonY));

    // Upper truss
    ctx.fillRect(towerX - 5,  horizonY - 170, 10, 170);

    // Horizontal cross-arm
    ctx.fillRect(towerX - 55, horizonY - 150, 55, 7);
    // Cross-arm support
    ctx.beginPath();
    ctx.moveTo(towerX - 55, horizonY - 143);
    ctx.lineTo(towerX - 7,  horizonY - 150);
    ctx.lineTo(towerX - 7,  horizonY - 135);
    ctx.closePath();
    ctx.fill();

    // Upper arm
    ctx.fillRect(towerX - 38, horizonY - 110, 38, 5);

    // Antenna spike
    ctx.fillRect(towerX - 1,  horizonY - 248, 2, 48);

    // Red warning light glow
    const lightGrd = ctx.createRadialGradient(
      towerX, horizonY - 248, 0,
      towerX, horizonY - 248, 10
    );
    lightGrd.addColorStop(0, 'rgba(255,50,30,0.8)');
    lightGrd.addColorStop(0.5, 'rgba(255,50,30,0.2)');
    lightGrd.addColorStop(1, 'rgba(255,50,30,0)');
    ctx.fillStyle = lightGrd;
    ctx.beginPath();
    ctx.arc(towerX, horizonY - 248, 10, 0, Math.PI * 2);
    ctx.fill();

    // Subtle ground lines — flat arid terrain lines
    ctx.strokeStyle = 'rgba(8,18,38,0.5)';
    ctx.lineWidth = 1;
    for (let i = 0; i < 3; i++) {
      const ly = horizonY + (h - horizonY) * (0.12 + i * 0.22);
      ctx.beginPath();
      ctx.moveTo(0, ly);
      ctx.lineTo(w, ly);
      ctx.stroke();
    }
  }

  function drawParticles(horizonY: number, progress: number): void {
    const intensity = Math.max(0, (progress - 0.1) * 1.1);
    if (intensity < 0.02) return;

    PARTICLES.forEach(p => {
      p.y -= p.speed / h;
      if (p.y < 0) { p.y = (horizonY / h) * 0.98; p.x = Math.random(); }
      const py = p.y * h;
      if (py > horizonY) return;
      ctx.beginPath();
      ctx.arc(p.x * w, py, p.r, 0, Math.PI * 2);
      ctx.fillStyle = `rgba(255,210,120,${p.alpha * intensity * 0.7})`;
      ctx.fill();
    });
  }

  function drawFigures(horizonY: number, panPx: number): void {
    const dark = 'rgba(3,5,10,0.94)';

    function figure(x: number, sz: number): void {
      ctx.fillStyle = dark;
      // Head
      ctx.beginPath();
      ctx.arc(x, horizonY - sz * 0.92, sz * 0.1, 0, Math.PI * 2);
      ctx.fill();
      // Torso
      ctx.beginPath();
      ctx.moveTo(x - sz*0.13, horizonY - sz*0.78);
      ctx.lineTo(x - sz*0.17, horizonY - sz*0.38);
      ctx.lineTo(x + sz*0.17, horizonY - sz*0.38);
      ctx.lineTo(x + sz*0.13, horizonY - sz*0.78);
      ctx.closePath();
      ctx.fill();
      // Legs
      ctx.beginPath();
      ctx.moveTo(x - sz*0.13, horizonY - sz*0.38);
      ctx.lineTo(x - sz*0.15, horizonY);
      ctx.lineTo(x - sz*0.03, horizonY);
      ctx.lineTo(x,           horizonY - sz*0.32);
      ctx.lineTo(x + sz*0.03, horizonY);
      ctx.lineTo(x + sz*0.15, horizonY);
      ctx.lineTo(x + sz*0.13, horizonY - sz*0.38);
      ctx.closePath();
      ctx.fill();
    }

    // Congregation near launch tower (right side, pans with scene)
    const towerX = w * 0.92 - panPx * 1.7;
    figure(towerX - 80, h * 0.078);
    figure(towerX - 55, h * 0.068);
    figure(towerX - 35, h * 0.082);
    figure(towerX - 18, h * 0.072);
    figure(towerX + 40, h * 0.065);

    // Scattered — left side
    const mesaX = w * 0.08 - panPx * 1.5;
    figure(mesaX + 160, h * 0.060);
    figure(mesaX + 320, h * 0.055);
  }

  let mwPhotoEl: HTMLElement;
  let mortalEl: HTMLElement;
  let riseEl: HTMLElement;
  let ctaEl: HTMLAnchorElement;
  let frameCount = 0;
  let milkyWayOffsetY = 0;

  function draw(): void {
    if (!ctx || !canvas) return;

    // Neck-tilt feel — fast enough to feel physical
    camX += (targetX - camX) * 0.20;
    camY += (targetY - camY) * 0.20;
    frameCount++;

    // Time progression: 0 = golden hour, 1 = full night (over 90 seconds)
    const progress = Math.min(1, (Date.now() - startTime) / 90000);

    // Neutral shows 65% sky. Ground exits at camY≈0.49 — exactly at button threshold.
    const horizonY = h * (0.65 + camY * 0.72);
    const panPx    = camX * w * 0.40;

    ctx.clearRect(0, 0, w, h);
    drawSky(horizonY, panPx, progress);
    drawGround(horizonY, panPx);
    drawFigures(horizonY, panPx);
    drawParticles(horizonY, progress);

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
    milkyWayOffsetY = 0;

    // Notify parent every 4 frames
    if (frameCount % 4 === 0) onCameraUpdate(camY);

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

    window.addEventListener('resize', resize);

    function onMouseMove(e: MouseEvent) {
      // 2.5x amplification — mouse only needs to reach ~30% from top for full reveal
      targetX = Math.max(-1, Math.min(1,  (e.clientX / w - 0.5) * 2.5));
      targetY = Math.max(-0.5, Math.min(1, -((e.clientY / h - 0.5) * 2.5)));
    }

    function onMouseLeave() {
      targetX = 0;
      targetY = 0;
    }

    window.addEventListener('mousemove', onMouseMove);
    window.addEventListener('mouseleave', onMouseLeave);

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
    };
  });
</script>

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
  style="transform: translateY({milkyWayOffsetY}px);"
></div>

<canvas
  bind:this={canvas}
  aria-hidden="true"
  style="position: fixed; inset: 0; width: 100vw; height: 100vh; pointer-events: none; z-index: 2;"
></canvas>

<style>
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
    will-change: transform;
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
    font-size: 0.65rem;
    letter-spacing: 0.25em;
    padding: 1rem 2.5rem;
    text-decoration: none;
    background: rgba(0, 0, 0, 0.4);
    backdrop-filter: blur(8px);
    box-shadow: 0 0 30px rgba(240, 237, 230, 0.08), 0 0 60px rgba(200, 146, 42, 0.06);
    animation: ctaPulse 2.8s ease-in-out infinite;
    opacity: 0;
    pointer-events: none;
    white-space: nowrap;
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
